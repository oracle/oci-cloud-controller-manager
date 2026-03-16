// Copyright 2026 Oracle and/or its affiliates. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package flexcidr

import (
	"context"
	"encoding/json"
	"fmt"
	"math/bits"
	"net"
	"sync"
	"time"

	ociclient "github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/util/retry"
)

const (
	podCIDRTag             = "pod-cidr"
	primaryVnicMetadataKey = "flexcidr-primary-vnic"
	createTimeout          = 55 * time.Second
	listTimeout            = 25 * time.Second
	defaultNamespace       = "default"
)

type IpFamily struct {
	IPv4 string
	IPv6 string
}

type PrimaryVnicConfig struct {
	IPCount    *int     `json:"ip-count"`
	CIDRBlocks []string `json:"cidr-blocks"`
}

func ParsePrimaryVnicConfig(instance *core.Instance) (PrimaryVnicConfig, bool) {
	v, ok := instance.Metadata[primaryVnicMetadataKey]
	if !ok {
		return PrimaryVnicConfig{}, false
	}

	var config PrimaryVnicConfig
	if err := json.Unmarshal([]byte(v), &config); err != nil {
		return PrimaryVnicConfig{}, false
	}
	return config, true
}

func GetClusterIpFamily(ctx context.Context, serviceLister corelisters.ServiceLister) (IpFamily, error) {
	svc, err := serviceLister.Services(defaultNamespace).Get("kubernetes")
	if err != nil {
		return IpFamily{}, err
	}
	var family IpFamily
	ipFamilies := svc.Spec.IPFamilies
	if len(ipFamilies) == 0 || len(ipFamilies) > 2 {
		return family, fmt.Errorf("IPFamily unset/invalid")
	}
	for _, ipFamily := range ipFamilies {
		if ipFamily == corev1.IPv4Protocol {
			family.IPv4 = "IPv4"
		}
		if ipFamily == corev1.IPv6Protocol {
			family.IPv6 = "IPv6"
		}
	}
	return family, nil
}

func PatchNodePodCIDRs(ctx context.Context, kubeClient kubernetes.Interface, nodeName string, podCIDRs []string, logger *zap.SugaredLogger) error {
	if len(podCIDRs) == 0 {
		return fmt.Errorf("no PodCIDRs computed")
	}

	patchBytes, err := json.Marshal(map[string]any{
		"spec": map[string]any{
			"podCIDR":  podCIDRs[0],
			"podCIDRs": append([]string(nil), podCIDRs...),
		},
	})
	if err != nil {
		return err
	}

	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		_, patchErr := kubeClient.CoreV1().Nodes().Patch(ctx, nodeName, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
		return patchErr
	}); err != nil {
		if logger != nil {
			logger.Errorf("failed to patch node %s podCIDRs: %v", nodeName, err)
		}
		return err
	}

	if logger != nil {
		logger.Infof("successfully patched node %s podCIDRs to %v", nodeName, podCIDRs)
	}
	return nil
}

func StringSlicesEqualIgnoreOrder(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]int, len(a))
	for _, s := range a {
		m[s]++
	}
	for _, s := range b {
		m[s]--
		if m[s] < 0 {
			return false
		}
	}
	for _, v := range m {
		if v != 0 {
			return false
		}
	}
	return true
}

func Ipv4PrefixFromCount(ipCount int) (int, error) {
	if ipCount <= 0 || (ipCount&(ipCount-1)) != 0 {
		return 0, fmt.Errorf("ipCount must be a power of 2, got %d", ipCount)
	}
	k := bits.Len(uint(ipCount)) - 1
	pfx := 32 - k
	if pfx < 18 {
		return 0, fmt.Errorf("ipCount=%d yields /%d; OCI requires cidrPrefixLength >= 18 (max ipCount 16384)", ipCount, pfx)
	}
	if pfx > 30 {
		return 0, fmt.Errorf("ipCount=%d yields /%d; must be <= /30 (min ipCount 4)", ipCount, pfx)
	}
	return pfx, nil
}

func Ipv6PrefixFromCount(ipCount int) (int, error) {
	if ipCount <= 0 || (ipCount&(ipCount-1)) != 0 {
		return 0, fmt.Errorf("ipCount must be a power of 2, got %d", ipCount)
	}
	k := bits.Len(uint(ipCount)) - 1
	pfx := 128 - k
	validPfx := pfx / 4 * 4
	if pfx < 80 {
		return 0, fmt.Errorf("ipCount=%d yields /%d; must be >= /80 (max ipCount 2^48)", ipCount, pfx)
	}
	return validPfx, nil
}

func getCIDRsByFamily(cidrBlocks []string) ([]string, []string, error) {
	var ipv4CidrBlocks, ipv6CidrBlocks []string

	for _, cidr := range cidrBlocks {
		ip, _, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid CIDR block in %s.cidrBlocks: %q: %w", primaryVnicMetadataKey, cidr, err)
		}

		if ip.To4() != nil {
			ipv4CidrBlocks = append(ipv4CidrBlocks, cidr)
		} else {
			ipv6CidrBlocks = append(ipv6CidrBlocks, cidr)
		}
	}

	return ipv4CidrBlocks, ipv6CidrBlocks, nil
}

func formatIpCidr(ip string, mask *int) string {
	if mask == nil {
		return ip
	}
	return fmt.Sprintf("%s/%d", ip, *mask)
}

type FlexCIDR struct {
	Logger            *zap.SugaredLogger
	PrimaryVnicConfig PrimaryVnicConfig
	ClusterIpFamily   IpFamily
	OciCoreClient     ociclient.NetworkingInterface
}

func (f *FlexCIDR) validateCidrBlocks(ipv4Blocks []string, ipv6Blocks []string) (string, string, error) {
	if f.ClusterIpFamily.IPv4 == "" && len(ipv4Blocks) > 0 {
		return "", "", fmt.Errorf("IPv4 CIDR is not allowed for this cluster but provided: %v", ipv4Blocks)
	}
	if f.ClusterIpFamily.IPv6 == "" && len(ipv6Blocks) > 0 {
		return "", "", fmt.Errorf("IPv6 CIDR is not allowed for this cluster but provided: %v", ipv6Blocks)
	}
	if len(ipv4Blocks) > 1 || len(ipv6Blocks) > 1 {
		return "", "", fmt.Errorf("only one IPv4 CIDR block and one IPv6 CIDR block are allowed for DualStack cluster; found %d IPv4 and %d IPv6 CIDR blocks", len(ipv4Blocks), len(ipv6Blocks))
	}
	var ipv4, ipv6 string
	if len(ipv4Blocks) == 1 {
		ipv4 = ipv4Blocks[0]
	}
	if len(ipv6Blocks) == 1 {
		ipv6 = ipv6Blocks[0]
	}
	return ipv4, ipv6, nil
}

func (f *FlexCIDR) getCidrBlocks() (string, string, error) {
	if f.Logger != nil {
		f.Logger.Infof("PrimaryVnicConfig CIDR blocks: %v", f.PrimaryVnicConfig.CIDRBlocks)
	}
	ipv4Blocks, ipv6Blocks, err := getCIDRsByFamily(f.PrimaryVnicConfig.CIDRBlocks)
	if err != nil {
		return "", "", err
	}
	return f.validateCidrBlocks(ipv4Blocks, ipv6Blocks)
}

func (f *FlexCIDR) ValidateFlexCidrList(flexCidrs []string) bool {
	if len(flexCidrs) == 0 {
		if f.Logger != nil {
			f.Logger.Errorf("flexCidrs is empty")
		}
		return false
	}
	for _, flexCidr := range flexCidrs {
		if flexCidr == "" {
			if f.Logger != nil {
				f.Logger.Errorf("flexCidrs contains empty string")
			}
			return false
		}
	}
	if f.ClusterIpFamily.IPv4 != "" && f.ClusterIpFamily.IPv6 != "" && len(flexCidrs) != 2 {
		if f.Logger != nil {
			f.Logger.Errorf("existing flexCidrs %v for dual stack should contain both IPv4 and IPv6 CIDRs", flexCidrs)
		}
		return false
	}
	if (f.ClusterIpFamily.IPv4 != "" && f.ClusterIpFamily.IPv6 == "") || (f.ClusterIpFamily.IPv4 == "" && f.ClusterIpFamily.IPv6 != "") {
		if len(flexCidrs) != 1 {
			if f.Logger != nil {
				f.Logger.Errorf("flexCidrs %v is not valid for single stack cluster", flexCidrs)
			}
			return false
		}
	}
	return true
}

func (f *FlexCIDR) GetFlexCidrList(primaryVnicID string) ([]string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), listTimeout)
	defer cancel()

	var flexCidrs []string

	if f.ClusterIpFamily.IPv4 != "" {
		privateIPs, err := f.OciCoreClient.ListPrivateIps(ctx, primaryVnicID)
		if err != nil {
			if f.Logger != nil {
				f.Logger.Errorf("failed to list private IPs for VNIC %s: %v", primaryVnicID, err)
			}
			return flexCidrs, false
		}

		for _, privateIP := range privateIPs {
			if _, ok := privateIP.FreeformTags[podCIDRTag]; ok {
				flexCidrs = append(flexCidrs, formatIpCidr(*privateIP.IpAddress, privateIP.CidrPrefixLength))
			}
		}
	}

	if f.ClusterIpFamily.IPv6 != "" {
		ipv6s, err := f.OciCoreClient.ListIpv6s(ctx, primaryVnicID)
		if err != nil {
			if f.Logger != nil {
				f.Logger.Errorf("failed to list IPv6s for VNIC %s: %v", primaryVnicID, err)
			}
			return flexCidrs, false
		}

		for _, ipv6 := range ipv6s {
			if _, ok := ipv6.FreeformTags[podCIDRTag]; ok {
				flexCidrs = append(flexCidrs, formatIpCidr(*ipv6.IpAddress, ipv6.CidrPrefixLength))
			}
		}
	}

	return flexCidrs, len(flexCidrs) > 0
}

func (f *FlexCIDR) CreateFlexCidr(primaryVnicID string, isIPv4 bool, isIPv6 bool) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), createTimeout)
	defer cancel()

	flexCidr := ""

	if isIPv4 == isIPv6 {
		return "", fmt.Errorf("exactly one IP family must be requested")
	}

	ipv4CidrBlock, ipv6CidrBlock, err := f.getCidrBlocks()
	if err != nil {
		return flexCidr, err
	}

	if f.PrimaryVnicConfig.IPCount == nil {
		return "", fmt.Errorf("primaryVNIC.ipCount is nil")
	}
	ipCount := *f.PrimaryVnicConfig.IPCount

	if isIPv4 {
		cidrPrefixLength, err := Ipv4PrefixFromCount(ipCount)
		if err != nil {
			return "", err
		}
		createPrivateIPDetails := core.CreatePrivateIpDetails{
			VnicId:           common.String(primaryVnicID),
			CidrPrefixLength: common.Int(cidrPrefixLength),
			FreeformTags:     map[string]string{podCIDRTag: "true"},
		}
		if ipv4CidrBlock != "" {
			createPrivateIPDetails.Ipv4SubnetCidrAtCreation = common.String(ipv4CidrBlock)
		}
		privateIP, err := f.OciCoreClient.CreatePrivateIpWithRequest(ctx, core.CreatePrivateIpRequest{
			CreatePrivateIpDetails: createPrivateIPDetails,
		})
		if err != nil {
			return flexCidr, fmt.Errorf("failed to assign flex CIDR to IPv4 VNIC: %w", err)
		}
		ipv4Address := *privateIP.IpAddress
		parsedIP := net.ParseIP(ipv4Address)
		if parsedIP == nil || parsedIP.To4() == nil {
			return flexCidr, fmt.Errorf("flex CIDR address (%s) returned by VCN is not a valid IPv4 address", ipv4Address)
		}
		flexCidr = formatIpCidr(ipv4Address, privateIP.CidrPrefixLength)
	}

	if isIPv6 {
		cidrPrefixLength, err := Ipv6PrefixFromCount(ipCount)
		if err != nil {
			return "", err
		}
		createIPv6Details := core.CreateIpv6Details{
			VnicId:           common.String(primaryVnicID),
			CidrPrefixLength: common.Int(cidrPrefixLength),
			FreeformTags:     map[string]string{podCIDRTag: "true"},
		}
		if ipv6CidrBlock != "" {
			createIPv6Details.Ipv6SubnetCidr = common.String(ipv6CidrBlock)
		}
		ipv6, err := f.OciCoreClient.CreateIpv6WithRequest(ctx, core.CreateIpv6Request{
			CreateIpv6Details: createIPv6Details,
		})
		if err != nil {
			return flexCidr, fmt.Errorf("failed to assign flex CIDR to IPv6 VNIC: %w", err)
		}
		ipv6Address := *ipv6.IpAddress
		parsedIP := net.ParseIP(ipv6Address)
		if parsedIP == nil || parsedIP.To4() != nil {
			return flexCidr, fmt.Errorf("flex CIDR address (%s) returned by VCN is not a valid IPv6 address", ipv6Address)
		}
		flexCidr = formatIpCidr(ipv6Address, ipv6.CidrPrefixLength)
	}

	return flexCidr, nil
}

func (f *FlexCIDR) GetOrCreateFlexCidrList(primaryVnicID string) ([]string, error) {
	var flexCidrs []string

	existingFlexCIDRs, ok := f.GetFlexCidrList(primaryVnicID)
	if ok {
		if f.Logger != nil {
			f.Logger.Infof("flexCidrs %v already exist on primary VNIC %s", existingFlexCIDRs, primaryVnicID)
		}
		if f.ValidateFlexCidrList(existingFlexCIDRs) {
			return existingFlexCIDRs, nil
		}
		return nil, fmt.Errorf("flexCidrs %v is invalid", existingFlexCIDRs)
	}

	type createResult struct {
		cidr string
		err  error
	}

	var (
		wg         sync.WaitGroup
		ipv4Result createResult
		ipv6Result createResult
	)

	if f.ClusterIpFamily.IPv4 != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ipv4Result.cidr, ipv4Result.err = f.CreateFlexCidr(primaryVnicID, true, false)
		}()
	}

	if f.ClusterIpFamily.IPv6 != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ipv6Result.cidr, ipv6Result.err = f.CreateFlexCidr(primaryVnicID, false, true)
		}()
	}

	wg.Wait()

	if ipv4Result.err != nil {
		return nil, ipv4Result.err
	}
	if ipv6Result.err != nil {
		return nil, ipv6Result.err
	}
	if f.ClusterIpFamily.IPv4 != "" {
		flexCidrs = append(flexCidrs, ipv4Result.cidr)
	}
	if f.ClusterIpFamily.IPv6 != "" {
		flexCidrs = append(flexCidrs, ipv6Result.cidr)
	}

	if len(flexCidrs) == 0 {
		return nil, apierrors.NewBadRequest("no flex CIDRs created")
	}

	return flexCidrs, nil
}
