// Copyright (c) 2019, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"oracle.com/oci/httpsigner"
)

func TestNewRPSTRequest(t *testing.T) {
	rr := newRPSTRequest(testTokenStr, testKeyID, "testpk")
	var reqtype *resourcePrincipalSessionTokenRequest
	assert.IsType(t, reqtype, rr)
}

func TestNewResourcePrincipalSessionTokenProvider(t *testing.T) {
	s, _ := NewSTSKeySupplier(testX509CertificateSupplier, testX509Client, "http://localhost/v1")
	// should return a ResourcePrincipalSessionTokenProvider with the arguments in the appropriate members
	rp := NewResourcePrincipalSessionTokenProvider("http://localhost/v1", s, OCIJWTSigningAlgorithms, testX509Client, "RS256")
	var provtype *ResourcePrincipalSessionTokenProvider
	assert.IsType(t, provtype, rp)
	assert.Equal(t, rp.resourcePrincipalSessionTokenEndpoint, "http://localhost/v1")
	assert.Equal(t, rp.stsKeySupplier, s)
	assert.Equal(t, rp.stsClient, testX509Client)
	assert.Equal(t, rp.signingAlgorithm, "RS256")
}
func TestGetRPSTWithEmptySigner(t *testing.T) {
	s, _ := NewSTSKeySupplier(testX509CertificateSupplier, testX509Client, "http://localhost/v1")
	rp := NewResourcePrincipalSessionTokenProvider("http://localhost/v1", s, OCIJWTSigningAlgorithms, testX509Client, "RS256")
	rp.stsKeySupplier.client = &MockSigningClient{}
	rp.stsKeySupplier.token = goodToken
	// should return a url Error with empty signing client
	rpst, err := rp.GetRPST(testRPTClaims)
	assert.IsType(t, &url.Error{}, err)
	assert.Empty(t, rpst)
}
func TestGetRPSTNon200Response(t *testing.T) {
	s, _ := NewSTSKeySupplier(testX509CertificateSupplier, testX509Client, "http://localhost/v1")
	rp := NewResourcePrincipalSessionTokenProvider("http://localhost/v1", s, OCIJWTSigningAlgorithms, testX509Client, "RS256")
	rp.stsClient = &MockSigningClient{doResponse: &http.Response{StatusCode: 404}, doError: nil}
	rp.stsKeySupplier.token = goodToken
	// should return 404 if 404 response received
	rpst, err := rp.GetRPST(testRPTClaims)
	assert.IsType(t, err, &ServiceResponseError{})
	assert.Empty(t, rpst)
}
func TestGetRPSTWithInvalidKey(t *testing.T) {
	s, _ := NewSTSKeySupplier(testX509CertificateSupplier, testX509Client, "http://localhost/v1")
	rp := NewResourcePrincipalSessionTokenProvider("http://localhost/v1", s, OCIJWTSigningAlgorithms, testX509Client, "RS256")
	stsKeyIDError := errors.New("some-error")
	rp.stsKeySupplier.client = &MockSigningClient{doError: stsKeyIDError}
	rp.stsKeySupplier.token = expiredToken
	// should return InvalidKey error on using expired token (using a mock client to return the error)
	rpst, err := rp.GetRPST(testRPTClaims)
	assert.Equal(t, err, stsKeyIDError)
	assert.Empty(t, rpst)
}
func TestRPSTFailToGenerateSignedRPT(t *testing.T) {
	s, _ := NewSTSKeySupplier(testX509CertificateSupplier, testX509Client, "http://localhost/v1")
	rp := NewResourcePrincipalSessionTokenProvider("http://localhost/v1", s, OCIJWTSigningAlgorithms, testX509Client, "RS256")
	rp.stsKeySupplier.client = &MockSigningClient{}
	rp.stsKeySupplier.token = goodToken
	rp.signingAlgorithm = "foo"
	// should return ErrInvalidSigningAlgorithm for unsupported alg name
	rpst, err := rp.GetRPST(testRPTClaims)
	assert.IsType(t, httpsigner.ErrUnsupportedAlgorithm, err)
	assert.Empty(t, rpst)
}
func TestGetRPSTInvalidJSONresponse(t *testing.T) {
	invalidJSON := []byte(`{test: test`)
	s, _ := NewSTSKeySupplier(testX509CertificateSupplier, testX509Client, "http://localhost/v1")
	rp := NewResourcePrincipalSessionTokenProvider("http://localhost/v1", s, OCIJWTSigningAlgorithms, testX509Client, "RS256")
	rp.stsClient = &MockSigningClient{doResponse: &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(invalidJSON))}, doError: nil}
	rp.stsKeySupplier.token = goodToken
	// should return json error on getting an invalid json response
	rpst, err := rp.GetRPST(testRPTClaims)
	assert.IsType(t, err, &json.SyntaxError{})
	assert.Empty(t, rpst)
}
func TestGetRPST200Response(t *testing.T) {
	s, _ := NewSTSKeySupplier(testX509CertificateSupplier, testX509Client, "http://localhost/v1")
	rp := NewResourcePrincipalSessionTokenProvider("http://localhost/v1", s, OCIJWTSigningAlgorithms, testX509Client, "RS256")
	rp.stsClient = &MockSigningClient{doResponse: &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewBufferString(fmt.Sprintf(`{"token": "%s"}`, goodToken.rawToken)))}, doError: nil}
	rp.stsKeySupplier.token = goodToken
	// should return valid rspt
	rpst, err := rp.GetRPST(testRPTClaims)
	assert.Nil(t, err)
	assert.NotEmpty(t, rpst)
}
