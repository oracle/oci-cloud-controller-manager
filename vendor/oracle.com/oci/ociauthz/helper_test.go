// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import "net/http"

const (
	testKeyID = "ocid1.tenancy.oc1..aaaaaaaaba3pv6wkcr4jqae5f15p2b2m2yt2j6rx32uzr4h25vqstifsfdsq/ocid1.user.oc1..aaaaaaaat5nvwcna5j6aqzjcaty5eqbb6qt2jvpkanghtgdaqedqw3rynjq/20:3b:97:13:55:1c:5b:0d:d3:37:d8:50:4e:c5:3a:34"
)

// MockRequestSigner doesn't actually sign requests, but keeps track of when SignRequest is called.
type MockRequestSigner struct {
	signingError error

	// mock state
	SignRequestCalled bool
	ProfferedRequest  *http.Request
	ProfferedKey      string
	ProfferedHeaders  []string
}

// SignRequest will return the proffered http.Request unless signingError is not nil in which case it will return the
// value of signingError.
func (mrs *MockRequestSigner) SignRequest(r *http.Request, k string, h []string) (*http.Request, error) {

	// save state of call
	mrs.SignRequestCalled = true
	mrs.ProfferedRequest = r
	mrs.ProfferedKey = k
	mrs.ProfferedHeaders = h

	// mock response
	if mrs.signingError != nil {
		return nil, mrs.signingError
	}
	return r, nil
}

// Reset clears state so the mock can be reused
func (mrs *MockRequestSigner) Reset() {
	mrs.SignRequestCalled = false
	mrs.ProfferedRequest = nil
	mrs.ProfferedKey = ""
	mrs.ProfferedHeaders = nil
}

// The following keys were generated using
// openssl dsaparam 1024 < /dev/random > dsaparam.pem
// openssl gendsa dsaparam.pem -out dsa_priv.pem
// openssl dsa -in dsa_priv.pem -pubout -out dsa_pub.pem
var testStaticDSAPublicKey = `
-----BEGIN PUBLIC KEY-----
MIIBtjCCASsGByqGSM44BAEwggEeAoGBAO2FFZcTm4RgF/q83xe/BrY3aQonNRk+
ySfv7/3WvHBUg+TrXEa3NECH8KxyxnZH5CD29lLjmw3u+s7RLXHR6cgGfQ2hSA1V
3Gl0H+cC1PasMrQk2015F1NiEU5l9weyinsgrRw3fpLnBZsq3yyApIivircb6KG+
AOhkeNwdpnmjAhUA8wN+qySstq/nu09MChMHh6AYcQkCgYBwlyXsuIkDNTU7U33H
EDVFWMFVsI7OgyQm52k2gM/b/FmiqWFHDD1JvFOzp0tiUw3s8/ceZr4NfFeuhO1G
DGEcMCuOu+UfCO5FvIH5M4oSSNUNwnoMtQzho5McQEV76QdcGpO6wda/ociGfpvq
Mcr48/7T505cZHnaQhEp7eGU5AOBhAACgYBmc1ehOvjN0AFob7+xxkPG+5BcQMbZ
+CXhxJ98qEAlt1ASQ8WUgaehH0pSTSRaMLBPOsrjt5OoL3sDPF+L1bvQu5+YMXLC
W+PdAq9JaXGZB0br+HKa+BGQmwFPv6rsCp6XAHjNXGZFTYwH/j4ll2GOEmtajLJu
MgoKuP+RxoIjfg==
-----END PUBLIC KEY-----
`

var testStaticDSAPrivateKey = `
-----BEGIN DSA PRIVATE KEY-----
MIIBugIBAAKBgQDthRWXE5uEYBf6vN8Xvwa2N2kKJzUZPskn7+/91rxwVIPk61xG
tzRAh/CscsZ2R+Qg9vZS45sN7vrO0S1x0enIBn0NoUgNVdxpdB/nAtT2rDK0JNtN
eRdTYhFOZfcHsop7IK0cN36S5wWbKt8sgKSIr4q3G+ihvgDoZHjcHaZ5owIVAPMD
fqskrLav57tPTAoTB4egGHEJAoGAcJcl7LiJAzU1O1N9xxA1RVjBVbCOzoMkJudp
NoDP2/xZoqlhRww9SbxTs6dLYlMN7PP3Hma+DXxXroTtRgxhHDArjrvlHwjuRbyB
+TOKEkjVDcJ6DLUM4aOTHEBFe+kHXBqTusHWv6HIhn6b6jHK+PP+0+dOXGR52kIR
Ke3hlOQCgYBmc1ehOvjN0AFob7+xxkPG+5BcQMbZ+CXhxJ98qEAlt1ASQ8WUgaeh
H0pSTSRaMLBPOsrjt5OoL3sDPF+L1bvQu5+YMXLCW+PdAq9JaXGZB0br+HKa+BGQ
mwFPv6rsCp6XAHjNXGZFTYwH/j4ll2GOEmtajLJuMgoKuP+RxoIjfgIUPbMiNpwP
idjmgpGl4gp6tpfRyiM=
-----END DSA PRIVATE KEY-----
`
