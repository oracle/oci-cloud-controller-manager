package oci

import (
	"errors"
	"reflect"
	"testing"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"

	api "k8s.io/api/core/v1"
)

func TestExtractNodeAddressesFromVNIC(t *testing.T) {
	testCases := []struct {
		name string
		in   *core.Vnic
		out  []api.NodeAddress
		err  error
	}{
		{
			name: "basic-complete",
			in: &core.Vnic{
				PrivateIp: common.String("10.0.0.1"),
				PublicIp:  common.String("0.0.0.1"),
			},
			out: []api.NodeAddress{
				api.NodeAddress{Type: api.NodeInternalIP, Address: "10.0.0.1"},
				api.NodeAddress{Type: api.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
		{
			name: "no-external-ip",
			in: &core.Vnic{
				PrivateIp: common.String("10.0.0.1"),
			},
			out: []api.NodeAddress{
				api.NodeAddress{Type: api.NodeInternalIP, Address: "10.0.0.1"},
			},
			err: nil,
		},
		{
			name: "no-internal-ip",
			in: &core.Vnic{
				PublicIp: common.String("0.0.0.1"),
			},
			out: []api.NodeAddress{
				api.NodeAddress{Type: api.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
		{
			name: "invalid-external-ip",
			in: &core.Vnic{
				PublicIp: common.String("0.0.0."),
			},
			out: nil,
			err: errors.New(`instance has invalid public address: "0.0.0."`),
		},
		{
			name: "invalid-external-ip",
			in: &core.Vnic{
				PrivateIp: common.String("10.0.0."),
			},
			out: nil,
			err: errors.New(`instance has invalid private address: "10.0.0."`),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractNodeAddressesFromVNIC(tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("extractNodeAddressesFromVNIC(%+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("extractNodeAddressesFromVNIC(%+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}
