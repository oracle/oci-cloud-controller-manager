package client

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestValidateConfig(t *testing.T) {
	testCases := []struct {
		name string
		in   *Config
		errs field.ErrorList
	}{
		{
			name: "valid",
			in: &Config{
				Auth: AuthConfig{
					Region:          "us-phoenix-1",
					TenancyOCID:     "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					CompartmentOCID: "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq",
					UserOCID:        "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					PrivateKey:      "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:     "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaahuxrgvs65iwdz7ekwgg3l5gyah7ww5klkwjcso74u3e4i64hvtvq",
				},
			},
			errs: field.ErrorList{},
		}, {
			name: "missing_region",
			in: &Config{
				Auth: AuthConfig{
					TenancyOCID:     "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					CompartmentOCID: "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq",
					UserOCID:        "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					PrivateKey:      "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:     "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaahuxrgvs65iwdz7ekwgg3l5gyah7ww5klkwjcso74u3e4i64hvtvq",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.region", BadValue: ""},
			},
		}, {
			name: "missing_tenancy",
			in: &Config{
				Auth: AuthConfig{
					Region:          "us-phoenix-1",
					CompartmentOCID: "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq",
					UserOCID:        "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					PrivateKey:      "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:     "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaahuxrgvs65iwdz7ekwgg3l5gyah7ww5klkwjcso74u3e4i64hvtvq",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.tenancy", BadValue: ""},
			},
		}, {
			name: "missing_compartment",
			in: &Config{
				Auth: AuthConfig{
					Region:      "us-phoenix-1",
					TenancyOCID: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					UserOCID:    "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					PrivateKey:  "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint: "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaahuxrgvs65iwdz7ekwgg3l5gyah7ww5klkwjcso74u3e4i64hvtvq",
				},
			},
			errs: field.ErrorList{},
		}, {
			name: "missing_user",
			in: &Config{
				Auth: AuthConfig{
					Region:          "us-phoenix-1",
					TenancyOCID:     "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					CompartmentOCID: "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq",
					PrivateKey:      "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:     "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaahuxrgvs65iwdz7ekwgg3l5gyah7ww5klkwjcso74u3e4i64hvtvq",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.user", BadValue: ""},
			},
		}, {
			name: "missing_key",
			in: &Config{
				Auth: AuthConfig{
					Region:          "us-phoenix-1",
					TenancyOCID:     "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					CompartmentOCID: "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq",
					UserOCID:        "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					Fingerprint:     "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaahuxrgvs65iwdz7ekwgg3l5gyah7ww5klkwjcso74u3e4i64hvtvq",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.key", BadValue: ""},
			},
		}, {
			name: "missing_figerprint",
			in: &Config{
				Auth: AuthConfig{
					Region:          "us-phoenix-1",
					TenancyOCID:     "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					CompartmentOCID: "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq",
					UserOCID:        "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					PrivateKey:      "-----BEGIN RSA PRIVATE KEY----- (etc)",
				},
				LoadBalancer: LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaahuxrgvs65iwdz7ekwgg3l5gyah7ww5klkwjcso74u3e4i64hvtvq",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.fingerprint", BadValue: ""},
			},
		}, {
			name: "missing_subnet1",
			in: &Config{
				Auth: AuthConfig{
					Region:          "us-phoenix-1",
					TenancyOCID:     "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					CompartmentOCID: "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq",
					UserOCID:        "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					PrivateKey:      "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:     "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: LoadBalancerConfig{
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaahuxrgvs65iwdz7ekwgg3l5gyah7ww5klkwjcso74u3e4i64hvtvq",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "loadBalancer.subnet1", BadValue: ""},
			},
		}, {
			name: "missing_subnet2",
			in: &Config{
				Auth: AuthConfig{
					Region:          "us-phoenix-1",
					TenancyOCID:     "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					CompartmentOCID: "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq",
					UserOCID:        "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					PrivateKey:      "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:     "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "loadBalancer.subnet2", BadValue: ""},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateConfig(tt.in)
			if !reflect.DeepEqual(result, tt.errs) {
				t.Errorf("ValidateConfig(%#v)\n=> %#v\nExpected: %#v", tt.in, result, tt.errs)
			}
		})
	}
}
