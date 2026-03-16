package flexcidr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/bits"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	ociclient "github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/util/retry"
)

var retryRNG = rand.New(rand.NewSource(time.Now().UnixNano()))

const (
	podCIDRTag             = "pod-cidr"
	primaryVnicMetadataKey = "flexcidr-primary-vnic"
	createTimeout          = 55 * time.Second
	listTimeout            = 25 * time.Second
	defaultNamespace       = "default"
)

var initRetryOnce sync.Once

func initOCIRetry() {
	initRetryOnce.Do(func() {
		p := common.DefaultRetryPolicy()
		common.GlobalRetry = &p
	})
}

func init() { initOCIRetry() }

func runWithRateLimitRetry(ctx context.Context, logger *zap.SugaredLogger, operation string, fn func(context.Context) error) error {
	policy := common.NewRetryPolicyWithOptions(
		common.WithMaximumNumberAttempts(6),
		common.WithShouldRetryOperation(func(r common.OCIOperationResponse) bool {
			return isRateLimitError(r.Error)
		}),
		common.WithNextDuration(func(r common.OCIOperationResponse) time.Duration {
			attempt := float64(r.AttemptNumber - 1)
			base := math.Pow(2, attempt)
			jitter := 1 + (retryRNG.Float64()-0.5)*0.2
			return time.Duration(base * jitter * float64(time.Second))
		}),
	)

	maxAttempts := policy.MaximumNumberAttempts
	var lastErr error

	for attempt := uint(1); maxAttempts == 0 || attempt <= maxAttempts; attempt++ {
		opErr := fn(ctx)
		lastErr = opErr
		operationResponse := common.OCIOperationResponse{Error: opErr, AttemptNumber: attempt}

		if !policy.ShouldRetryOperation(operationResponse) {
			return opErr
		}

		backoff := policy.NextDuration(operationResponse)
		if logger != nil {
			logger.Warnf("%s hit rate limit on attempt %d, retrying in %s", operation, attempt, backoff)
		}
		if deadline, ok := ctx.Deadline(); ok && time.Now().Add(backoff).After(deadline) {
			return fmt.Errorf("%s retry exceeded context deadline: %w", operation, context.DeadlineExceeded)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("operation %s reached retry limit", operation)
	}
	return fmt.Errorf("%s retry exceeded maximum attempts: %w", operation, lastErr)
}

func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	var serviceErr common.ServiceError
	if errors.As(err, &serviceErr) {
		return serviceErr.GetHTTPStatusCode() == http.StatusTooManyRequests
	}
	return false
}

type IpFamily struct {
	IPv4 string
	IPv6 string
}

type PrimaryVnicConfig struct {
	IPCount    *int     `json:"ip-count"`
	CIDRBlocks []string `json:"cidr-blocks"`
}

func ParsePrimaryVnicConfig(instance *core.Instance) (PrimaryVnicConfig, bool) {
	for k, v := range instance.Metadata {
		if k == primaryVnicMetadataKey {
			var config PrimaryVnicConfig
			if err := json.Unmarshal([]byte(v), &config); err != nil {
				return PrimaryVnicConfig{}, false
			}
			return config, true
		}
	}
	return PrimaryVnicConfig{}, false
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

	node, err := kubeClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	nodeDry := node.DeepCopy()
	nodeDry.Spec.PodCIDR = podCIDRs[0]
	nodeDry.Spec.PodCIDRs = append([]string(nil), podCIDRs...)
	if _, err := kubeClient.CoreV1().Nodes().Update(ctx, nodeDry, metav1.UpdateOptions{DryRun: []string{metav1.DryRunAll}}); err != nil {
		if logger != nil {
			logger.Errorf("dry-run update failed for node %s: %v", nodeName, err)
		}
		return err
	}

	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		currentNode, getErr := kubeClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
		if getErr != nil {
			return getErr
		}
		currentNode.Spec.PodCIDR = podCIDRs[0]
		currentNode.Spec.PodCIDRs = append([]string(nil), podCIDRs...)
		_, updateErr := kubeClient.CoreV1().Nodes().Update(ctx, currentNode, metav1.UpdateOptions{})
		return updateErr
	}); err != nil {
		if logger != nil {
			logger.Errorf("failed to update node %s: %v", nodeName, err)
		}
		return err
	}

	updatedNode, err := kubeClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if updatedNode.Spec.PodCIDR != podCIDRs[0] {
		return fmt.Errorf("post-update check: spec.podCIDR=%q (expected %q)", updatedNode.Spec.PodCIDR, podCIDRs[0])
	}
	if !StringSlicesEqualIgnoreOrder(updatedNode.Spec.PodCIDRs, podCIDRs) {
		return fmt.Errorf("post-update check: spec.podCIDRs=%v (expected %v)", updatedNode.Spec.PodCIDRs, podCIDRs)
	}

	if logger != nil {
		logger.Infof("successfully updated node %s podCIDRs to %v", nodeName, podCIDRs)
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

	if f.ClusterIpFamily.IPv4 != "" {
		ipv4FlexCIDR, err := f.CreateFlexCidr(primaryVnicID, true, false)
		if err != nil {
			return nil, err
		}
		flexCidrs = append(flexCidrs, ipv4FlexCIDR)
	}

	if f.ClusterIpFamily.IPv6 != "" {
		ipv6FlexCIDR, err := f.CreateFlexCidr(primaryVnicID, false, true)
		if err != nil {
			return nil, err
		}
		flexCidrs = append(flexCidrs, ipv6FlexCIDR)
	}

	if len(flexCidrs) == 0 {
		return nil, apierrors.NewBadRequest("no flex CIDRs created")
	}

	return flexCidrs, nil
}
