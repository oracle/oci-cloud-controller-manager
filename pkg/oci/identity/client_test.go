// Copyright 2019 Oracle and/or its affiliates. All rights reserved.
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
package identity

import (
	"bytes"
	"context"
	"fmt"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/identity"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
)

var (
	dataTpl = `[DEFAULT]
	tenancy=sometenancy
	user=someuser
	fingerprint=somefingerprint
	key_file=%s
	region=noregion
	`
	testPrivateKeyConf = `-----BEGIN RSA PRIVATE KEY-----
MIICXgIBAAKBgQDCFENGw33yGihy92pDjZQhl0C36rPJj+CvfSC8+q28hxA161QF
NUd13wuCTUcq0Qd2qsBe/2hFyc2DCJJg0h1L78+6Z4UMR7EOcpfdUE9Hf3m/hs+F
UR45uBJeDK1HSFHD8bHKD6kv8FPGfJTotc+2xjJwoYi+1hqp1fIekaxsyQIDAQAB
AoGBAJR8ZkCUvx5kzv+utdl7T5MnordT1TvoXXJGXK7ZZ+UuvMNUCdN2QPc4sBiA
QWvLw1cSKt5DsKZ8UETpYPy8pPYnnDEz2dDYiaew9+xEpubyeW2oH4Zx71wqBtOK
kqwrXa/pzdpiucRRjk6vE6YY7EBBs/g7uanVpGibOVAEsqH1AkEA7DkjVH28WDUg
f1nqvfn2Kj6CT7nIcE3jGJsZZ7zlZmBmHFDONMLUrXR/Zm3pR5m0tCmBqa5RK95u
412jt1dPIwJBANJT3v8pnkth48bQo/fKel6uEYyboRtA5/uHuHkZ6FQF7OUkGogc
mSJluOdc5t6hI1VsLn0QZEjQZMEOWr+wKSMCQQCC4kXJEsHAve77oP6HtG/IiEn7
kpyUXRNvFsDE0czpJJBvL/aRFUJxuRK91jhjC68sA7NsKMGg5OXb5I5Jj36xAkEA
gIT7aFOYBFwGgQAQkWNKLvySgKbAZRTeLBacpHMuQdl1DfdntvAyqpAZ0lY0RKmW
G6aFKaqQfOXKCyWoUiVknQJAXrlgySFci/2ueKlIE1QqIiLSZ8V8OlpFLRnb1pzI
7U1yQXnTAEFYM560yJlzUpOb1V4cScGd365tiSMvxLOvTA==
-----END RSA PRIVATE KEY-----`
)

// FakeCaller will mock metadataClient.HTTPClient Do call
type FakeCaller struct {
	FakeCall func(*http.Request) (*http.Response, error)
}

func (f FakeCaller) Do(req *http.Request) (*http.Response, error) {
	return f.FakeCall(req)
}

func writeTempFile(data string) (filename string) {
	f, _ := ioutil.TempFile("", "IdentityMetadataSvcTest")
	f.WriteString(data)
	filename = f.Name()
	return
}

func TestMetadataClient_GetTenantByCompartment(t *testing.T) {
	keyFile := writeTempFile(testPrivateKeyConf)
	data := fmt.Sprintf(dataTpl, keyFile)
	tmpConfFile := writeTempFile(data)
	defer os.Remove(tmpConfFile)
	defer os.Remove(keyFile)

	// ConfigurationProvider creation requires non-empty values of conf variables
	cp, err := common.ConfigurationProviderFromFile(tmpConfFile, "")
	if err != nil {
		t.Errorf("Error creating Configuration provider: %v", err)
	}
	metadataClient, err := NewMetadataClientWithConfigurationProvider(cp)
	if err != nil {
		t.Errorf("Error creating metadatClient: %v", err)
	}

	testcases := map[string]struct {
		compartment_id *string
		tenancy        identity.Tenancy
		json           string
	}{"Test1": {
		compartment_id: common.String("ocid.comp1"),
		tenancy: identity.Tenancy{
			Name: common.String("tenant1"),
			Id:   common.String("ocid.tenant1"),
		},
		json: `{"id" : "ocid.tenant1","name" : "tenant1"}`,
	}, "Test2": {
		compartment_id: common.String("ocid.comp2"),
		tenancy: identity.Tenancy{
			Name: common.String("tenant2"),
			Id:   common.String("ocid.tenant2"),
		},
		json: `{"id" : "ocid.tenant2","name" : "tenant2"}`,
	},
	}

	//overwrite metadataclient to use FakeCaller.
	metadataClient.HTTPClient = FakeCaller{
		FakeCall: func(request *http.Request) (*http.Response, error) {
			response := &http.Response{
				Header:     http.Header{},
				StatusCode: 200,
			}
			// path is of the form /v1/compartments/{compartment}/tenant
			if strings.Contains(request.URL.Path, *testcases["Test1"].compartment_id) {
				response.Body = ioutil.NopCloser(bytes.NewBufferString(testcases["Test1"].json))
			}
			if strings.Contains(request.URL.Path, *testcases["Test2"].compartment_id) {
				response.Body = ioutil.NopCloser(bytes.NewBufferString(testcases["Test2"].json))
			}
			return response, nil
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			result, err := metadataClient.GetTenantByCompartment(context.Background(), GetTenantByCompartmentRequest{CompartmentId: tc.compartment_id})
			if err != nil {
				t.Errorf("Error running test: %v", err)
			}
			if *result.Id != *tc.tenancy.Id {
				t.Errorf("Expected value to be %s but got %s", *tc.tenancy.Id, *result.Id)
			}
		})
	}
}
