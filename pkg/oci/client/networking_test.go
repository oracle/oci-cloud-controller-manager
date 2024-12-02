package client

import (
	"reflect"
	"testing"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"k8s.io/client-go/tools/cache"
)

var (
	subnets = map[string]*core.Subnet{
		"IPv4-subnet": {
			Id:                 common.String("IPv4-subnet"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: nil,
			CidrBlock:          common.String("10.0.10.0/24"),
		},
		"IPv6-subnet": {
			Id:                 common.String("IPv6-subnet"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: nil,
			Ipv6CidrBlock:      common.String("2603:c020:f:d222::/64"),
			Ipv6CidrBlocks:     []string{"2603:c020:f:d222::/64"},
		},
		"IPv4-IPv6-subnet": {
			Id:                 common.String("IPv4-IPv6-subnet"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: nil,
			CidrBlock:          common.String("10.0.11.0/24"),
			Ipv6CidrBlocks:     []string{},
			Ipv6CidrBlock:      common.String("2603:c020:f:d277::/64"),
		},
		"IPv6-subnet-ULA": {
			Id:                 common.String("IPv6-subnet-ULA"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: nil,
			Ipv6CidrBlock:      nil,
			Ipv6CidrBlocks:     []string{"fd12:3456:789a:1::/64"},
		},
	}
)

func Test_client_GetSubnetFromCacheByIP(t *testing.T) {
	type fields struct {
		subnetCache cache.Store
	}
	type args struct {
		ip IpAddresses
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *core.Subnet
		wantErr bool
	}{
		{
			name: "IPv4 SingleStack Subnet",
			fields: fields{
				subnetCache: cache.NewTTLStore(subnetCacheKeyFn, time.Duration(24)*time.Hour),
			},
			args: args{
				ip: IpAddresses{
					V4: "10.0.10.1",
				},
			},
			want:    subnets["IPv4-subnet"],
			wantErr: false,
		},
		{
			name: "IPv6 SingleStack Subnet",
			fields: fields{
				subnetCache: cache.NewTTLStore(subnetCacheKeyFn, time.Duration(24)*time.Hour),
			},
			args: args{
				ip: IpAddresses{
					V6: "2603:C020:000F:D222:0000:0000:0000:0000",
				},
			},
			want:    subnets["IPv6-subnet"],
			wantErr: false,
		},
		{
			name: "DualStack  Subnet",
			fields: fields{
				subnetCache: cache.NewTTLStore(subnetCacheKeyFn, time.Duration(24)*time.Hour),
			},
			args: args{
				ip: IpAddresses{
					V4: "10.0.11.1",
					V6: "2603:C020:000F:D277:0000:0000:0000:0000",
				},
			},
			want:    subnets["IPv4-IPv6-subnet"],
			wantErr: false,
		},
		{
			name: "IPv6 ULA Subnet",
			fields: fields{
				subnetCache: cache.NewTTLStore(subnetCacheKeyFn, time.Duration(24)*time.Hour),
			},
			args: args{
				ip: IpAddresses{
					V6: "fd12:3456:789a:0001:0000:0000:0000:0000",
				},
			},
			want:    subnets["IPv6-subnet-ULA"],
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &client{
				subnetCache: cache.NewTTLStore(subnetCacheKeyFn, time.Duration(24)*time.Hour),
			}
			err := c.subnetCache.Add(subnets["IPv4-subnet"])
			err = c.subnetCache.Add(subnets["IPv6-subnet"])
			err = c.subnetCache.Add(subnets["IPv4-IPv6-subnet"])
			err = c.subnetCache.Add(subnets["IPv6-subnet-ULA"])
			if err != nil {
				return
			}

			got, err := c.GetSubnetFromCacheByIP(tt.args.ip)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetSubnetFromCacheByIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSubnetFromCacheByIP() got = %v, want %v", got, tt.want)
			}
		})
	}
}
