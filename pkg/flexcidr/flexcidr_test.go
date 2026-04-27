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
	"testing"
	"time"

	ociclient "github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

type testOciCoreClient struct {
	ociclient.NetworkingInterface

	listPrivateIpsResp []core.PrivateIp
	listPrivateIpsErr  error
	listIpv6sResp      []core.Ipv6
	listIpv6sErr       error

	createPrivateIpResp core.PrivateIp
	createPrivateIpErr  error
	lastCreatePrivate   *core.CreatePrivateIpRequest
	privateStarted      chan struct{}
	privateRelease      <-chan struct{}

	createIpv6Resp core.Ipv6
	createIpv6Err  error
	lastCreateIpv6 *core.CreateIpv6Request
	ipv6Started    chan struct{}
	ipv6Release    <-chan struct{}
}

func (m *testOciCoreClient) ListPrivateIps(_ context.Context, _ string) ([]core.PrivateIp, error) {
	return m.listPrivateIpsResp, m.listPrivateIpsErr
}

func (m *testOciCoreClient) ListIpv6s(_ context.Context, _ string) ([]core.Ipv6, error) {
	return m.listIpv6sResp, m.listIpv6sErr
}

func (m *testOciCoreClient) CreatePrivateIpWithRequest(_ context.Context, req core.CreatePrivateIpRequest) (core.PrivateIp, error) {
	m.lastCreatePrivate = &req
	if m.privateStarted != nil {
		close(m.privateStarted)
	}
	if m.privateRelease != nil {
		<-m.privateRelease
	}
	return m.createPrivateIpResp, m.createPrivateIpErr
}

func (m *testOciCoreClient) CreateIpv6WithRequest(_ context.Context, req core.CreateIpv6Request) (core.Ipv6, error) {
	m.lastCreateIpv6 = &req
	if m.ipv6Started != nil {
		close(m.ipv6Started)
	}
	if m.ipv6Release != nil {
		<-m.ipv6Release
	}
	return m.createIpv6Resp, m.createIpv6Err
}

func testLogger() *zap.SugaredLogger {
	return zap.NewNop().Sugar()
}

func TestParsePrimaryVnicConfig(t *testing.T) {
	instance := &core.Instance{Metadata: map[string]string{primaryVnicMetadataKey: `{"ip-count":16,"cidr-blocks":["10.0.0.0/24"]}`}}
	cfg, ok := ParsePrimaryVnicConfig(instance)
	assert.True(t, ok)
	if assert.NotNil(t, cfg.IPCount) {
		assert.Equal(t, 16, *cfg.IPCount)
	}
	assert.Equal(t, []string{"10.0.0.0/24"}, cfg.CIDRBlocks)

	_, ok = ParsePrimaryVnicConfig(&core.Instance{Metadata: map[string]string{primaryVnicMetadataKey: `{"ip-count":`}})
	assert.False(t, ok)

	_, ok = ParsePrimaryVnicConfig(&core.Instance{Metadata: map[string]string{"other": "x"}})
	assert.False(t, ok)
}

func TestGetClusterIpFamily(t *testing.T) {
	kubeClient := fake.NewSimpleClientset(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "kubernetes", Namespace: defaultNamespace},
		Spec:       corev1.ServiceSpec{IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol, corev1.IPv6Protocol}},
	})
	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	serviceInformer := factory.Core().V1().Services()
	assert.NoError(t, serviceInformer.Informer().GetStore().Add(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "kubernetes", Namespace: defaultNamespace},
		Spec:       corev1.ServiceSpec{IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol, corev1.IPv6Protocol}},
	}))

	family, err := GetClusterIpFamily(context.Background(), serviceInformer.Lister())
	assert.NoError(t, err)
	assert.Equal(t, IpFamily{IPv4: "IPv4", IPv6: "IPv6"}, family)
}

func TestPatchNodePodCIDRs(t *testing.T) {
	kubeClient := fake.NewSimpleClientset(&corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "node-1",
			Labels: map[string]string{"existing": "label"},
		},
		Spec: corev1.NodeSpec{
			ProviderID: "oci://instance",
		},
	})

	err := PatchNodePodCIDRs(context.Background(), kubeClient, "node-1", []string{"10.0.0.0/24", "2001:db8::/80"}, testLogger())
	assert.NoError(t, err)

	updated, err := kubeClient.CoreV1().Nodes().Get(context.Background(), "node-1", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.0/24", updated.Spec.PodCIDR)
	assert.ElementsMatch(t, []string{"10.0.0.0/24", "2001:db8::/80"}, updated.Spec.PodCIDRs)
	assert.Equal(t, "label", updated.Labels["existing"])
	assert.Equal(t, "oci://instance", updated.Spec.ProviderID)
}

func TestStringSlicesEqualIgnoreOrder(t *testing.T) {
	assert.True(t, StringSlicesEqualIgnoreOrder([]string{"a", "b"}, []string{"b", "a"}))
	assert.True(t, StringSlicesEqualIgnoreOrder([]string{"a", "a", "b"}, []string{"a", "b", "a"}))
	assert.False(t, StringSlicesEqualIgnoreOrder([]string{"a"}, []string{"a", "b"}))
	assert.False(t, StringSlicesEqualIgnoreOrder([]string{"a", "b"}, []string{"a", "c"}))
}

func TestIpv4PrefixFromCount(t *testing.T) {
	tests := []struct {
		name    string
		count   int
		prefix  int
		errText string
	}{
		{name: "valid minimum", count: 4, prefix: 30},
		{name: "valid maximum", count: 16384, prefix: 18},
		{name: "not power of two", count: 3, errText: "power of 2"},
		{name: "too small prefix", count: 32768, errText: "requires cidrPrefixLength >= 18"},
		{name: "too large prefix", count: 2, errText: "must be <= /30"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, err := Ipv4PrefixFromCount(tt.count)
			if tt.errText != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errText)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.prefix, prefix)
		})
	}
}

func TestIpv6PrefixFromCount(t *testing.T) {
	tests := []struct {
		name    string
		count   int
		prefix  int
		errText string
	}{
		{name: "valid rounds to nibble", count: 1024, prefix: 116},
		{name: "valid exact nibble", count: 65536, prefix: 112},
		{name: "not power of two", count: 7, errText: "power of 2"},
		{name: "too small prefix", count: 1 << 49, errText: "must be >= /80"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, err := Ipv6PrefixFromCount(tt.count)
			if tt.errText != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errText)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.prefix, prefix)
		})
	}
}

func TestGetCIDRsByFamily(t *testing.T) {
	ipv4, ipv6, err := getCIDRsByFamily([]string{"10.0.0.0/24", "2001:db8::/64"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"10.0.0.0/24"}, ipv4)
	assert.Equal(t, []string{"2001:db8::/64"}, ipv6)

	_, _, err = getCIDRsByFamily([]string{"invalid"})
	assert.Error(t, err)
}

func TestValidateCidrBlocks(t *testing.T) {
	f := &FlexCIDR{Logger: testLogger(), ClusterIpFamily: IpFamily{IPv4: "IPv4", IPv6: "IPv6"}}
	ipv4, ipv6, err := f.validateCidrBlocks([]string{"10.0.0.0/24"}, []string{"2001:db8::/64"})
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.0/24", ipv4)
	assert.Equal(t, "2001:db8::/64", ipv6)

	f = &FlexCIDR{Logger: testLogger(), ClusterIpFamily: IpFamily{IPv4: "IPv4"}}
	_, _, err = f.validateCidrBlocks(nil, []string{"2001:db8::/64"})
	assert.Error(t, err)

	f = &FlexCIDR{Logger: testLogger(), ClusterIpFamily: IpFamily{IPv4: "IPv4", IPv6: "IPv6"}}
	_, _, err = f.validateCidrBlocks([]string{"10.0.0.0/24", "10.1.0.0/24"}, nil)
	assert.Error(t, err)
}

func TestValidateFlexCidrList(t *testing.T) {
	fDual := &FlexCIDR{Logger: testLogger(), ClusterIpFamily: IpFamily{IPv4: "IPv4", IPv6: "IPv6"}}
	assert.True(t, fDual.ValidateFlexCidrList([]string{"10.0.0.1/30", "2001:db8::1/116"}))
	assert.False(t, fDual.ValidateFlexCidrList([]string{"10.0.0.1/30"}))

	fSingle := &FlexCIDR{Logger: testLogger(), ClusterIpFamily: IpFamily{IPv4: "IPv4"}}
	assert.True(t, fSingle.ValidateFlexCidrList([]string{"10.0.0.1/30"}))
	assert.False(t, fSingle.ValidateFlexCidrList([]string{"10.0.0.1/30", "10.0.0.2/30"}))
	assert.False(t, fSingle.ValidateFlexCidrList([]string{""}))
}

func TestGetFlexCidrList(t *testing.T) {
	pfx4 := 30
	pfx6 := 116
	fakeClient := &testOciCoreClient{
		listPrivateIpsResp: []core.PrivateIp{
			{IpAddress: common.String("10.0.0.10"), CidrPrefixLength: &pfx4, FreeformTags: map[string]string{podCIDRTag: "true"}},
			{IpAddress: common.String("10.0.0.11"), CidrPrefixLength: &pfx4, FreeformTags: map[string]string{"other": "x"}},
		},
		listIpv6sResp: []core.Ipv6{
			{IpAddress: common.String("2001:db8::10"), CidrPrefixLength: &pfx6, FreeformTags: map[string]string{podCIDRTag: "true"}},
		},
	}
	f := &FlexCIDR{Logger: testLogger(), ClusterIpFamily: IpFamily{IPv4: "IPv4", IPv6: "IPv6"}, OciCoreClient: fakeClient}

	cidrs, ok := f.GetFlexCidrList("vnic")
	assert.True(t, ok)
	assert.ElementsMatch(t, []string{"10.0.0.10/30", "2001:db8::10/116"}, cidrs)
}

func TestCreateFlexCidrIPv4ConfiguresSubnetCidrWhenSet(t *testing.T) {
	pfx := 22
	fakeClient := &testOciCoreClient{createPrivateIpResp: core.PrivateIp{IpAddress: common.String("10.0.1.5"), CidrPrefixLength: &pfx}}
	ipCount := 1024
	f := &FlexCIDR{
		Logger:            testLogger(),
		PrimaryVnicConfig: PrimaryVnicConfig{IPCount: &ipCount, CIDRBlocks: []string{"10.0.0.0/24"}},
		ClusterIpFamily:   IpFamily{IPv4: "IPv4"},
		OciCoreClient:     fakeClient,
	}

	cidr, err := f.CreateFlexCidr("vnic", true, false)
	assert.NoError(t, err)
	assert.Equal(t, "10.0.1.5/22", cidr)
	if assert.NotNil(t, fakeClient.lastCreatePrivate) && assert.NotNil(t, fakeClient.lastCreatePrivate.CreatePrivateIpDetails.Ipv4SubnetCidrAtCreation) {
		assert.Equal(t, "10.0.0.0/24", *fakeClient.lastCreatePrivate.CreatePrivateIpDetails.Ipv4SubnetCidrAtCreation)
	}
}

func TestCreateFlexCidrIPv6ConfiguresSubnetCidrWhenSet(t *testing.T) {
	pfx := 116
	fakeClient := &testOciCoreClient{createIpv6Resp: core.Ipv6{IpAddress: common.String("2001:db8::5"), CidrPrefixLength: &pfx}}
	ipCount := 1024
	f := &FlexCIDR{
		Logger:            testLogger(),
		PrimaryVnicConfig: PrimaryVnicConfig{IPCount: &ipCount, CIDRBlocks: []string{"2001:db8:1::/64"}},
		ClusterIpFamily:   IpFamily{IPv6: "IPv6"},
		OciCoreClient:     fakeClient,
	}

	cidr, err := f.CreateFlexCidr("vnic", false, true)
	assert.NoError(t, err)
	assert.Equal(t, "2001:db8::5/116", cidr)
	if assert.NotNil(t, fakeClient.lastCreateIpv6) && assert.NotNil(t, fakeClient.lastCreateIpv6.CreateIpv6Details.Ipv6SubnetCidr) {
		assert.Equal(t, "2001:db8:1::/64", *fakeClient.lastCreateIpv6.CreateIpv6Details.Ipv6SubnetCidr)
	}
}

func TestCreateFlexCidrIPv6DoesNotConfigureSubnetCidrWhenUnset(t *testing.T) {
	pfx := 116
	fakeClient := &testOciCoreClient{createIpv6Resp: core.Ipv6{IpAddress: common.String("2001:db8::6"), CidrPrefixLength: &pfx}}
	ipCount := 1024
	f := &FlexCIDR{
		Logger:            testLogger(),
		PrimaryVnicConfig: PrimaryVnicConfig{IPCount: &ipCount},
		ClusterIpFamily:   IpFamily{IPv6: "IPv6"},
		OciCoreClient:     fakeClient,
	}

	cidr, err := f.CreateFlexCidr("vnic", false, true)
	assert.NoError(t, err)
	assert.Equal(t, "2001:db8::6/116", cidr)
	if assert.NotNil(t, fakeClient.lastCreateIpv6) {
		assert.Nil(t, fakeClient.lastCreateIpv6.CreateIpv6Details.Ipv6SubnetCidr)
	}
}

func TestCreateFlexCidrRejectsInvalidFamilySelection(t *testing.T) {
	ipCount := 1024
	f := &FlexCIDR{
		Logger:            testLogger(),
		PrimaryVnicConfig: PrimaryVnicConfig{IPCount: &ipCount},
		OciCoreClient:     &testOciCoreClient{},
	}

	_, err := f.CreateFlexCidr("vnic", false, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one IP family")

	_, err = f.CreateFlexCidr("vnic", true, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one IP family")
}

func TestGetOrCreateFlexCidrList(t *testing.T) {
	ipCount := 1024
	pfx4 := 22
	pfx6 := 116
	fakeClient := &testOciCoreClient{
		createPrivateIpResp: core.PrivateIp{IpAddress: common.String("10.0.1.5"), CidrPrefixLength: &pfx4},
		createIpv6Resp:      core.Ipv6{IpAddress: common.String("2001:db8::7"), CidrPrefixLength: &pfx6},
	}
	f := &FlexCIDR{
		Logger:            testLogger(),
		PrimaryVnicConfig: PrimaryVnicConfig{IPCount: &ipCount},
		ClusterIpFamily:   IpFamily{IPv4: "IPv4", IPv6: "IPv6"},
		OciCoreClient:     fakeClient,
	}

	cidrs, err := f.GetOrCreateFlexCidrList("vnic")
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"10.0.1.5/22", "2001:db8::7/116"}, cidrs)
}

func TestGetOrCreateFlexCidrListCreatesFamiliesInParallel(t *testing.T) {
	ipCount := 1024
	pfx4 := 22
	pfx6 := 116
	privateStarted := make(chan struct{})
	ipv6Started := make(chan struct{})
	privateRelease := make(chan struct{})
	ipv6Release := make(chan struct{})
	fakeClient := &testOciCoreClient{
		createPrivateIpResp: core.PrivateIp{IpAddress: common.String("10.0.1.5"), CidrPrefixLength: &pfx4},
		createIpv6Resp:      core.Ipv6{IpAddress: common.String("2001:db8::7"), CidrPrefixLength: &pfx6},
		privateStarted:      privateStarted,
		privateRelease:      privateRelease,
		ipv6Started:         ipv6Started,
		ipv6Release:         ipv6Release,
	}
	f := &FlexCIDR{
		Logger:            testLogger(),
		PrimaryVnicConfig: PrimaryVnicConfig{IPCount: &ipCount},
		ClusterIpFamily:   IpFamily{IPv4: "IPv4", IPv6: "IPv6"},
		OciCoreClient:     fakeClient,
	}

	done := make(chan struct{})
	var (
		cidrs []string
		err   error
	)
	go func() {
		cidrs, err = f.GetOrCreateFlexCidrList("vnic")
		close(done)
	}()

	select {
	case <-privateStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for IPv4 flex CIDR creation to start")
	}

	select {
	case <-ipv6Started:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for IPv6 flex CIDR creation to start")
	}

	close(privateRelease)
	close(ipv6Release)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for parallel flex CIDR creation to finish")
	}

	assert.NoError(t, err)
	assert.Equal(t, []string{"10.0.1.5/22", "2001:db8::7/116"}, cidrs)
}

func TestFormatIpCidr(t *testing.T) {
	pfx := 24
	assert.Equal(t, "10.0.0.1/24", formatIpCidr("10.0.0.1", &pfx))
	assert.Equal(t, "10.0.0.1", formatIpCidr("10.0.0.1", nil))
}

func TestGetOrCreateFlexCidrListRejectsInvalidExistingCIDRs(t *testing.T) {
	pfx4 := 30
	fakeClient := &testOciCoreClient{
		listPrivateIpsResp: []core.PrivateIp{
			{IpAddress: common.String("10.0.0.10"), CidrPrefixLength: &pfx4, FreeformTags: map[string]string{podCIDRTag: "true"}},
		},
	}
	f := &FlexCIDR{
		Logger:          testLogger(),
		ClusterIpFamily: IpFamily{IPv4: "IPv4", IPv6: "IPv6"},
		OciCoreClient:   fakeClient,
	}

	_, err := f.GetOrCreateFlexCidrList("vnic")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestParsePrimaryVnicConfigCIDRBlocksOrderIndependent(t *testing.T) {
	instance := &core.Instance{Metadata: map[string]string{primaryVnicMetadataKey: `{"ip-count":16,"cidr-blocks":["2001:db8::/64","10.0.0.0/24"]}`}}
	cfg, ok := ParsePrimaryVnicConfig(instance)
	assert.True(t, ok)
	if assert.NotNil(t, cfg.IPCount) {
		assert.Equal(t, 16, *cfg.IPCount)
	}
	assert.Equal(t, []string{"2001:db8::/64", "10.0.0.0/24"}, cfg.CIDRBlocks)
}
