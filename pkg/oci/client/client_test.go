package client

import (
	"errors"
	"reflect"
	"testing"

	baremetal "github.com/oracle/bmcs-go-sdk"

	api "k8s.io/api/core/v1"
)

func TestExtractNodeAddressesFromVNIC(t *testing.T) {
	testCases := []struct {
		name string
		in   *baremetal.Vnic
		out  []api.NodeAddress
		err  error
	}{
		{
			name: "basic-complete",
			in: &baremetal.Vnic{
				PrivateIPAddress: "10.0.0.1",
				PublicIPAddress:  "0.0.0.1",
			},
			out: []api.NodeAddress{
				api.NodeAddress{Type: api.NodeInternalIP, Address: "10.0.0.1"},
				api.NodeAddress{Type: api.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
		{
			name: "no-external-ip",
			in: &baremetal.Vnic{
				PrivateIPAddress: "10.0.0.1",
			},
			out: []api.NodeAddress{
				api.NodeAddress{Type: api.NodeInternalIP, Address: "10.0.0.1"},
			},
			err: nil,
		},
		{
			name: "no-internal-ip",
			in: &baremetal.Vnic{
				PublicIPAddress: "0.0.0.1",
			},
			out: []api.NodeAddress{
				api.NodeAddress{Type: api.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
		{
			name: "invalid-external-ip",
			in: &baremetal.Vnic{
				PublicIPAddress: "0.0.0.",
			},
			out: nil,
			err: errors.New(`instance has invalid public address: "0.0.0."`),
		},
		{
			name: "invalid-external-ip",
			in: &baremetal.Vnic{
				PrivateIPAddress: "10.0.0.",
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
