package client

import (
	"context"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/flowcontrol"
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

type retryingVirtualNetworkClient struct {
	mockVirtualNetworkClient

	createPrivateIpCalls int
	createIpv6Calls      int
}

func (c *retryingVirtualNetworkClient) CreatePrivateIp(ctx context.Context, request core.CreatePrivateIpRequest) (core.CreatePrivateIpResponse, error) {
	c.createPrivateIpCalls++
	if c.createPrivateIpCalls == 1 {
		return core.CreatePrivateIpResponse{}, mockServiceError{
			StatusCode: http.StatusTooManyRequests,
			Code:       HTTP429TooManyRequestsCode,
			Message:    "rate limited",
		}
	}

	return core.CreatePrivateIpResponse{
		PrivateIp: core.PrivateIp{
			Id: common.String("private-ip-id"),
		},
	}, nil
}

func (c *retryingVirtualNetworkClient) CreateIpv6(ctx context.Context, request core.CreateIpv6Request) (core.CreateIpv6Response, error) {
	c.createIpv6Calls++
	if c.createIpv6Calls == 1 {
		return core.CreateIpv6Response{}, mockServiceError{
			StatusCode: http.StatusTooManyRequests,
			Code:       HTTP429TooManyRequestsCode,
			Message:    "rate limited",
		}
	}

	return core.CreateIpv6Response{
		Ipv6: core.Ipv6{
			Id: common.String("ipv6-id"),
		},
	}, nil
}

func TestCreatePrivateIpWithRequestRetriesRateLimit(t *testing.T) {
	originalMaxAttempts := rateLimitRetryMaxAttempts
	originalNextDuration := rateLimitRetryNextDuration
	rateLimitRetryMaxAttempts = 2
	rateLimitRetryNextDuration = func(attempt uint) time.Duration { return 0 }
	defer func() {
		rateLimitRetryMaxAttempts = originalMaxAttempts
		rateLimitRetryNextDuration = originalNextDuration
	}()

	network := &retryingVirtualNetworkClient{}
	c := &client{
		network: network,
		logger:  zap.NewNop().Sugar(),
		rateLimiter: RateLimiter{
			Reader: flowcontrol.NewFakeAlwaysRateLimiter(),
			Writer: flowcontrol.NewFakeAlwaysRateLimiter(),
		},
	}

	privateIP, err := c.CreatePrivateIpWithRequest(context.Background(), core.CreatePrivateIpRequest{
		CreatePrivateIpDetails: core.CreatePrivateIpDetails{
			VnicId: common.String("vnic-id"),
		},
	})
	if err != nil {
		t.Fatalf("CreatePrivateIpWithRequest() error = %v", err)
	}
	if network.createPrivateIpCalls != 2 {
		t.Fatalf("CreatePrivateIpWithRequest() calls = %d, want 2", network.createPrivateIpCalls)
	}
	if privateIP.Id == nil || *privateIP.Id != "private-ip-id" {
		got := "<nil>"
		if privateIP.Id != nil {
			got = *privateIP.Id
		}
		t.Fatalf("CreatePrivateIpWithRequest() private IP id = %q, want %q", got, "private-ip-id")
	}
}

func TestCreateIpv6WithRequestRetriesRateLimit(t *testing.T) {
	originalMaxAttempts := rateLimitRetryMaxAttempts
	originalNextDuration := rateLimitRetryNextDuration
	rateLimitRetryMaxAttempts = 2
	rateLimitRetryNextDuration = func(attempt uint) time.Duration { return 0 }
	defer func() {
		rateLimitRetryMaxAttempts = originalMaxAttempts
		rateLimitRetryNextDuration = originalNextDuration
	}()

	network := &retryingVirtualNetworkClient{}
	c := &client{
		network: network,
		logger:  zap.NewNop().Sugar(),
		rateLimiter: RateLimiter{
			Reader: flowcontrol.NewFakeAlwaysRateLimiter(),
			Writer: flowcontrol.NewFakeAlwaysRateLimiter(),
		},
	}

	ipv6, err := c.CreateIpv6WithRequest(context.Background(), core.CreateIpv6Request{
		CreateIpv6Details: core.CreateIpv6Details{
			VnicId: common.String("vnic-id"),
		},
	})
	if err != nil {
		t.Fatalf("CreateIpv6WithRequest() error = %v", err)
	}
	if network.createIpv6Calls != 2 {
		t.Fatalf("CreateIpv6WithRequest() calls = %d, want 2", network.createIpv6Calls)
	}
	if ipv6.Id == nil || *ipv6.Id != "ipv6-id" {
		got := "<nil>"
		if ipv6.Id != nil {
			got = *ipv6.Id
		}
		t.Fatalf("CreateIpv6WithRequest() IPv6 id = %q, want %q", got, "ipv6-id")
	}
}
