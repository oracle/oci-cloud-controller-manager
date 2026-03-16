package flexcidr

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"testing"

	ociclient "github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

type requireAssertions struct{}

var require requireAssertions

func (requireAssertions) True(t *testing.T, value bool, msgAndArgs ...interface{}) {
	t.Helper()
	if !value {
		t.Fatal(msgAndArgs...)
	}
}

func (requireAssertions) False(t *testing.T, value bool, msgAndArgs ...interface{}) {
	t.Helper()
	if value {
		t.Fatal(msgAndArgs...)
	}
}

func (requireAssertions) NoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err != nil {
		t.Fatal(append([]interface{}{err}, msgAndArgs...)...)
	}
}

func (requireAssertions) Error(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err == nil {
		t.Fatal(msgAndArgs...)
	}
}

func (requireAssertions) Equal(t *testing.T, expected interface{}, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Fatal(append([]interface{}{"expected", expected, "actual", actual}, msgAndArgs...)...)
	}
}

func (requireAssertions) ElementsMatch(t *testing.T, expected []string, actual []string, msgAndArgs ...interface{}) {
	t.Helper()
	if !StringSlicesEqualIgnoreOrder(expected, actual) {
		t.Fatal(append([]interface{}{"expected", expected, "actual", actual}, msgAndArgs...)...)
	}
}

func (requireAssertions) NotNil(t *testing.T, value interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if value == nil {
		t.Fatal(msgAndArgs...)
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		if rv.IsNil() {
			t.Fatal(msgAndArgs...)
		}
	}
}

func (requireAssertions) Nil(t *testing.T, value interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if value == nil {
		return
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		if rv.IsNil() {
			return
		}
	}
	if value != nil {
		t.Fatal(msgAndArgs...)
	}
}

func (requireAssertions) Contains(t *testing.T, s string, contains string, msgAndArgs ...interface{}) {
	t.Helper()
	if !strings.Contains(s, contains) {
		t.Fatal(append([]interface{}{"expected", s, "to contain", contains}, msgAndArgs...)...)
	}
}

type fakeServiceError struct {
	status int
}

func (e fakeServiceError) Error() string           { return "service error" }
func (e fakeServiceError) GetCode() string         { return "TooManyRequests" }
func (e fakeServiceError) GetMessage() string      { return "rate limited" }
func (e fakeServiceError) GetHTTPStatusCode() int  { return e.status }
func (e fakeServiceError) GetOpcRequestID() string { return "opc-request-id" }

type testOciCoreClient struct {
	ociclient.NetworkingInterface

	listPrivateIpsResp []core.PrivateIp
	listPrivateIpsErr  error
	listIpv6sResp      []core.Ipv6
	listIpv6sErr       error

	createPrivateIpResp core.PrivateIp
	createPrivateIpErr  error
	lastCreatePrivate   *core.CreatePrivateIpRequest

	createIpv6Resp core.Ipv6
	createIpv6Err  error
	lastCreateIpv6 *core.CreateIpv6Request
}

func (m *testOciCoreClient) ListPrivateIps(_ context.Context, _ string) ([]core.PrivateIp, error) {
	return m.listPrivateIpsResp, m.listPrivateIpsErr
}

func (m *testOciCoreClient) ListIpv6s(_ context.Context, _ string) ([]core.Ipv6, error) {
	return m.listIpv6sResp, m.listIpv6sErr
}

func (m *testOciCoreClient) CreatePrivateIpWithRequest(_ context.Context, req core.CreatePrivateIpRequest) (core.PrivateIp, error) {
	m.lastCreatePrivate = &req
	return m.createPrivateIpResp, m.createPrivateIpErr
}

func (m *testOciCoreClient) CreateIpv6WithRequest(_ context.Context, req core.CreateIpv6Request) (core.Ipv6, error) {
	m.lastCreateIpv6 = &req
	return m.createIpv6Resp, m.createIpv6Err
}

func testLogger() *zap.SugaredLogger {
	return zap.NewNop().Sugar()
}

func TestIsRateLimitError(t *testing.T) {
	require.False(t, isRateLimitError(nil))
	require.False(t, isRateLimitError(context.DeadlineExceeded))
	require.False(t, isRateLimitError(fakeServiceError{status: http.StatusBadRequest}))
	require.True(t, isRateLimitError(fakeServiceError{status: http.StatusTooManyRequests}))
}

func TestParsePrimaryVnicConfig(t *testing.T) {
	instance := &core.Instance{Metadata: map[string]string{primaryVnicMetadataKey: `{"ip-count":16,"cidr-blocks":["10.0.0.0/24"]}`}}
	cfg, ok := ParsePrimaryVnicConfig(instance)
	require.True(t, ok)
	require.NotNil(t, cfg.IPCount)
	require.Equal(t, 16, *cfg.IPCount)
	require.Equal(t, []string{"10.0.0.0/24"}, cfg.CIDRBlocks)

	_, ok = ParsePrimaryVnicConfig(&core.Instance{Metadata: map[string]string{primaryVnicMetadataKey: `{"ip-count":`}})
	require.False(t, ok)

	_, ok = ParsePrimaryVnicConfig(&core.Instance{Metadata: map[string]string{"other": "x"}})
	require.False(t, ok)
}

func TestGetClusterIpFamily(t *testing.T) {
	kubeClient := fake.NewSimpleClientset(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "kubernetes", Namespace: defaultNamespace},
		Spec:       corev1.ServiceSpec{IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol, corev1.IPv6Protocol}},
	})
	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	serviceInformer := factory.Core().V1().Services()
	require.NoError(t, serviceInformer.Informer().GetStore().Add(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "kubernetes", Namespace: defaultNamespace},
		Spec:       corev1.ServiceSpec{IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol, corev1.IPv6Protocol}},
	}))

	family, err := GetClusterIpFamily(context.Background(), serviceInformer.Lister())
	require.NoError(t, err)
	require.Equal(t, IpFamily{IPv4: "IPv4", IPv6: "IPv6"}, family)
}

func TestPatchNodePodCIDRs(t *testing.T) {
	kubeClient := fake.NewSimpleClientset(&corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
	})

	err := PatchNodePodCIDRs(context.Background(), kubeClient, "node-1", []string{"10.0.0.0/24", "2001:db8::/80"}, testLogger())
	require.NoError(t, err)

	updated, err := kubeClient.CoreV1().Nodes().Get(context.Background(), "node-1", metav1.GetOptions{})
	require.NoError(t, err)
	require.Equal(t, "10.0.0.0/24", updated.Spec.PodCIDR)
	require.ElementsMatch(t, []string{"10.0.0.0/24", "2001:db8::/80"}, updated.Spec.PodCIDRs)
}

func TestStringSlicesEqualIgnoreOrder(t *testing.T) {
	require.True(t, StringSlicesEqualIgnoreOrder([]string{"a", "b"}, []string{"b", "a"}))
	require.True(t, StringSlicesEqualIgnoreOrder([]string{"a", "a", "b"}, []string{"a", "b", "a"}))
	require.False(t, StringSlicesEqualIgnoreOrder([]string{"a"}, []string{"a", "b"}))
	require.False(t, StringSlicesEqualIgnoreOrder([]string{"a", "b"}, []string{"a", "c"}))
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
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errText)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.prefix, prefix)
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
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errText)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.prefix, prefix)
		})
	}
}

func TestGetCIDRsByFamily(t *testing.T) {
	ipv4, ipv6, err := getCIDRsByFamily([]string{"10.0.0.0/24", "2001:db8::/64"})
	require.NoError(t, err)
	require.Equal(t, []string{"10.0.0.0/24"}, ipv4)
	require.Equal(t, []string{"2001:db8::/64"}, ipv6)

	_, _, err = getCIDRsByFamily([]string{"invalid"})
	require.Error(t, err)
}

func TestValidateCidrBlocks(t *testing.T) {
	f := &FlexCIDR{Logger: testLogger(), ClusterIpFamily: IpFamily{IPv4: "IPv4", IPv6: "IPv6"}}
	ipv4, ipv6, err := f.validateCidrBlocks([]string{"10.0.0.0/24"}, []string{"2001:db8::/64"})
	require.NoError(t, err)
	require.Equal(t, "10.0.0.0/24", ipv4)
	require.Equal(t, "2001:db8::/64", ipv6)

	f = &FlexCIDR{Logger: testLogger(), ClusterIpFamily: IpFamily{IPv4: "IPv4"}}
	_, _, err = f.validateCidrBlocks(nil, []string{"2001:db8::/64"})
	require.Error(t, err)

	f = &FlexCIDR{Logger: testLogger(), ClusterIpFamily: IpFamily{IPv4: "IPv4", IPv6: "IPv6"}}
	_, _, err = f.validateCidrBlocks([]string{"10.0.0.0/24", "10.1.0.0/24"}, nil)
	require.Error(t, err)
}

func TestValidateFlexCidrList(t *testing.T) {
	fDual := &FlexCIDR{Logger: testLogger(), ClusterIpFamily: IpFamily{IPv4: "IPv4", IPv6: "IPv6"}}
	require.True(t, fDual.ValidateFlexCidrList([]string{"10.0.0.1/30", "2001:db8::1/116"}))
	require.False(t, fDual.ValidateFlexCidrList([]string{"10.0.0.1/30"}))

	fSingle := &FlexCIDR{Logger: testLogger(), ClusterIpFamily: IpFamily{IPv4: "IPv4"}}
	require.True(t, fSingle.ValidateFlexCidrList([]string{"10.0.0.1/30"}))
	require.False(t, fSingle.ValidateFlexCidrList([]string{"10.0.0.1/30", "10.0.0.2/30"}))
	require.False(t, fSingle.ValidateFlexCidrList([]string{""}))
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
	require.True(t, ok)
	require.ElementsMatch(t, []string{"10.0.0.10/30", "2001:db8::10/116"}, cidrs)
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
	require.NoError(t, err)
	require.Equal(t, "10.0.1.5/22", cidr)
	require.NotNil(t, fakeClient.lastCreatePrivate)
	require.NotNil(t, fakeClient.lastCreatePrivate.CreatePrivateIpDetails.Ipv4SubnetCidrAtCreation)
	require.Equal(t, "10.0.0.0/24", *fakeClient.lastCreatePrivate.CreatePrivateIpDetails.Ipv4SubnetCidrAtCreation)
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
	require.NoError(t, err)
	require.Equal(t, "2001:db8::5/116", cidr)
	require.NotNil(t, fakeClient.lastCreateIpv6)
	require.NotNil(t, fakeClient.lastCreateIpv6.CreateIpv6Details.Ipv6SubnetCidr)
	require.Equal(t, "2001:db8:1::/64", *fakeClient.lastCreateIpv6.CreateIpv6Details.Ipv6SubnetCidr)
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
	require.NoError(t, err)
	require.Equal(t, "2001:db8::6/116", cidr)
	require.NotNil(t, fakeClient.lastCreateIpv6)
	require.Nil(t, fakeClient.lastCreateIpv6.CreateIpv6Details.Ipv6SubnetCidr)
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
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"10.0.1.5/22", "2001:db8::7/116"}, cidrs)
}

func TestFormatIpCidr(t *testing.T) {
	pfx := 24
	require.Equal(t, "10.0.0.1/24", formatIpCidr("10.0.0.1", &pfx))
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
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "invalid"))
}

func TestParsePrimaryVnicConfigCIDRBlocksOrderIndependent(t *testing.T) {
	instance := &core.Instance{Metadata: map[string]string{primaryVnicMetadataKey: `{"ip-count":16,"cidr-blocks":["2001:db8::/64","10.0.0.0/24"]}`}}
	cfg, ok := ParsePrimaryVnicConfig(instance)
	require.True(t, ok)
	require.NotNil(t, cfg.IPCount)
	require.Equal(t, 16, *cfg.IPCount)
	require.True(t, reflect.DeepEqual([]string{"2001:db8::/64", "10.0.0.0/24"}, cfg.CIDRBlocks))
}
