// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package httpsigner

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSignRequestInvalidArgs tests SignRequest returns errors for nil/empty args
func TestSignRequestInvalidArgs(t *testing.T) {
	validRequest := httptest.NewRequest("GET", "http://example.org", nil)
	validKeyID := testKeyID
	validPrivateKey, e1 := NewPKCS1RSAPrivateKeyFromPEM(testPrivateKey)
	assert.Nil(t, e1)
	validKeySupplier, e2 := NewStaticRSAKeySupplier(validPrivateKey, testKeyID)
	assert.Nil(t, e2)
	validHeadersToSign := []string{"Date", "(request-target)", "Host"}
	validAlgorithm := AlgorithmRSASHA256

	testIO := []struct {
		tc                    string
		request               *http.Request
		keyID                 string
		keySupplier           KeySupplier
		headersToSign         []string
		algorithm             Algorithm
		expectedSignedRequest *http.Request
		expectedError         error
	}{
		{
			tc:                    `should return error for nil request arg`,
			request:               nil,
			keyID:                 validKeyID,
			keySupplier:           validKeySupplier,
			headersToSign:         validHeadersToSign,
			algorithm:             validAlgorithm,
			expectedSignedRequest: nil,
			expectedError:         newErrorInvalidArg("request"),
		},
		{
			tc:                    `should return error for empty keyID arg`,
			request:               validRequest,
			keyID:                 "",
			keySupplier:           validKeySupplier,
			headersToSign:         validHeadersToSign,
			algorithm:             validAlgorithm,
			expectedSignedRequest: nil,
			expectedError:         newErrorInvalidArg("keyID"),
		},
		{
			tc:                    `should return error for nil keySupplier arg`,
			request:               validRequest,
			keyID:                 validKeyID,
			keySupplier:           nil,
			headersToSign:         validHeadersToSign,
			algorithm:             validAlgorithm,
			expectedSignedRequest: nil,
			expectedError:         newErrorInvalidArg("keySupplier"),
		},
		{
			tc:                    `should return error for nil algorithm arg`,
			request:               validRequest,
			keyID:                 validKeyID,
			keySupplier:           validKeySupplier,
			headersToSign:         validHeadersToSign,
			algorithm:             nil,
			expectedSignedRequest: nil,
			expectedError:         newErrorInvalidArg("algorithm"),
		},
		{
			tc:                    `should return key not found for unknown key`,
			request:               validRequest,
			keyID:                 `unknown`,
			keySupplier:           validKeySupplier,
			headersToSign:         validHeadersToSign,
			algorithm:             validAlgorithm,
			expectedSignedRequest: nil,
			expectedError:         ErrKeyNotFound,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			signedRequest, err := SignRequest(
				test.request,
				test.keyID,
				test.keySupplier,
				test.headersToSign,
				test.algorithm)
			assert.Equal(t, test.expectedSignedRequest, signedRequest)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

// TestSignRequestNilHeadersToSign tests SignRequest defaults nil headersToSign to "Date" header
func TestSignRequestNilHeadersToSign(t *testing.T) {
	request := httptest.NewRequest("GET", "http://example.org", nil)
	key, _ := NewPKCS1RSAPrivateKeyFromPEM(testPrivateKey)
	supplier, _ := NewStaticRSAKeySupplier(key, testKeyID)
	var headersToSign []string // nil headersToSign
	alg := AlgorithmRSASHA256
	signedRequest, err := SignRequest(request, testKeyID, supplier, headersToSign, alg)
	assert.Nil(t, err)
	// default to signing date header
	expectedAuthHdrVal := `Signature version="1",headers="date",keyId="ocid1.tenancy.oc1..aaaaaaaaba3pv6wkcr4jqae5f15p2b2m2yt2j6rx32uzr4h25vqstifsfdsq/ocid1.user.oc1..aaaaaaaat5nvwcna5j6aqzjcaty5eqbb6qt2jvpkanghtgdaqedqw3rynjq/20:3b:97:13:55:1c:5b:0d:d3:37:d8:50:4e:c5:3a:34",algorithm="rsa-sha256",signature="Fp7Zpew1qWbpyqVeOyTznBOpcG112PPcLy4jeYmxf6exYNlsY8Tn7ao0qDiAbOiC/6hE7/0O6qbub3mG3jgW3uv5V+g8/TPTB2gdnz7QPeXW4YIAOb2/DPlrq2M2DaH4cxhuMJZP8cmOxmojfPHZdwrQSv+WGuddZTH27Tqib6g="`
	authHdrVal := signedRequest.Header.Get("Authorization")
	assert.Equal(t, expectedAuthHdrVal, authHdrVal)
}

// TestSignRequestNilHeadersToSign tests SignRequest defaults empty headersToSign to "Date" header
func TestSignRequestEmptyHeadersToSign(t *testing.T) {
	request := httptest.NewRequest("GET", "http://example.org", nil)
	key, _ := NewPKCS1RSAPrivateKeyFromPEM(testPrivateKey)
	supplier, _ := NewStaticRSAKeySupplier(key, testKeyID)
	var headersToSign = []string{} // empty headersToSign
	alg := AlgorithmRSASHA256
	signedRequest, err := SignRequest(request, testKeyID, supplier, headersToSign, alg)
	assert.Nil(t, err)
	// default to signing date header
	expectedAuthHdrVal := `Signature version="1",headers="date",keyId="ocid1.tenancy.oc1..aaaaaaaaba3pv6wkcr4jqae5f15p2b2m2yt2j6rx32uzr4h25vqstifsfdsq/ocid1.user.oc1..aaaaaaaat5nvwcna5j6aqzjcaty5eqbb6qt2jvpkanghtgdaqedqw3rynjq/20:3b:97:13:55:1c:5b:0d:d3:37:d8:50:4e:c5:3a:34",algorithm="rsa-sha256",signature="Fp7Zpew1qWbpyqVeOyTznBOpcG112PPcLy4jeYmxf6exYNlsY8Tn7ao0qDiAbOiC/6hE7/0O6qbub3mG3jgW3uv5V+g8/TPTB2gdnz7QPeXW4YIAOb2/DPlrq2M2DaH4cxhuMJZP8cmOxmojfPHZdwrQSv+WGuddZTH27Tqib6g="`
	authHdrVal := signedRequest.Header.Get("Authorization")
	assert.Equal(t, expectedAuthHdrVal, authHdrVal)
}

// TestSignRequestAlgSignErr tests SignRequest returns errMock for mockAlgorithm.Sign() err
func TestSignRequestAlgSignErr(t *testing.T) {
	request := httptest.NewRequest("GET", "http://example.org", nil)
	key, _ := NewPKCS1RSAPrivateKeyFromPEM(testPrivateKey)
	supplier, _ := NewStaticRSAKeySupplier(key, testKeyID)
	headersToSign := []string{"Date", "(request-target)", "Host"}
	alg := mockAlgorithm
	signedRequest, err := SignRequest(request, testKeyID, supplier, headersToSign, alg)
	assert.Nil(t, signedRequest)
	assert.Equal(t, errMock, err)
}

// TestSignRequestRSASHA256ValidGet tests signing a GET request using the 'rsa-sha256' signing algorithm
// It uses values from https://bitbucket.oci.oraclecorp.com/projects/SDK/repos/signing-examples/browse/raw/example
func TestSignRequestRSASHA256ValidGet(t *testing.T) {
	target := "https://iaas.us-phoenix-1.oraclecloud.com/20160918/instances?availabilityDomain=Pjwf%3A%20PHX-AD-1&compartmentId=ocid1.compartment.oc1..aaaaaaaam3we6vgnherjq5q2idnccdflvjsnog7mlr6rtdb25gilchfeyjxa&displayName=TeamXInstances&volumeId=ocid1.volume.oc1.phx.abyhqljrgvttnlx73nmrwfaux7kcvzfs3s66izvxf2h4lgvyndsdsnoiwr5q"
	key, _ := NewPKCS1RSAPrivateKeyFromPEM(testPrivateKey)
	supplier, _ := NewStaticRSAKeySupplier(key, testKeyID)
	request := httptest.NewRequest("GET", target, nil)
	headersToSign := []string{"Date", "(request-target)", "Host"}
	request.Header.Set("Date", "Thu, 05 Jan 2014 21:31:40 GMT")
	singedRequest, err := SignRequest(request, testKeyID, supplier, headersToSign, AlgorithmRSASHA256)
	assert.Nil(t, err)
	authHdrVal := singedRequest.Header.Get("Authorization")
	expectedAuthHdrVal := `Signature version="1",headers="date (request-target) host",keyId="ocid1.tenancy.oc1..aaaaaaaaba3pv6wkcr4jqae5f15p2b2m2yt2j6rx32uzr4h25vqstifsfdsq/ocid1.user.oc1..aaaaaaaat5nvwcna5j6aqzjcaty5eqbb6qt2jvpkanghtgdaqedqw3rynjq/20:3b:97:13:55:1c:5b:0d:d3:37:d8:50:4e:c5:3a:34",algorithm="rsa-sha256",signature="GBas7grhyrhSKHP6AVIj/h5/Vp8bd/peM79H9Wv8kjoaCivujVXlpbKLjMPeDUhxkFIWtTtLBj3sUzaFj34XE6YZAHc9r2DmE4pMwOAy/kiITcZxa1oHPOeRheC0jP2dqbTll8fmTZVwKZOKHYPtrLJIJQHJjNvxFWeHQjMaR7M="`
	assert.Equal(t, expectedAuthHdrVal, authHdrVal)
}

// TestSignRequestRSASHA256ValidPost tests signing a POST request using the 'rsa-sha256' signing algorithm
// It uses values from https://bitbucket.oci.oraclecorp.com/projects/SDK/repos/signing-examples/browse/raw/example
func TestSignRequestRSASHA256ValidPost(t *testing.T) {
	target := "https://iaas.us-phoenix-1.oraclecloud.com/20160918/volumeAttachments"
	key, _ := NewPKCS1RSAPrivateKeyFromPEM(testPrivateKey)
	supplier, _ := NewStaticRSAKeySupplier(key, testKeyID)
	request := httptest.NewRequest("POST", target, strings.NewReader(testBody))
	headersToSign := []string{"Date", "(request-target)", "Host", "Content-Length", "Content-Type", "x-content-sha256"}
	request.Header.Set("Date", "Thu, 05 Jan 2014 21:31:40 GMT")
	request.Header.Set("Content-Length", "316")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-content-sha256", "V9Z20UJTvkvpJ50flBzKE32+6m2zJjweHpDMX/U4Uy0=")
	singedRequest, err := SignRequest(request, testKeyID, supplier, headersToSign, AlgorithmRSASHA256)
	assert.Nil(t, err)
	authHdrVal := singedRequest.Header.Get("Authorization")
	expectedAuthHdrVal := `Signature version="1",headers="date (request-target) host content-length content-type x-content-sha256",keyId="ocid1.tenancy.oc1..aaaaaaaaba3pv6wkcr4jqae5f15p2b2m2yt2j6rx32uzr4h25vqstifsfdsq/ocid1.user.oc1..aaaaaaaat5nvwcna5j6aqzjcaty5eqbb6qt2jvpkanghtgdaqedqw3rynjq/20:3b:97:13:55:1c:5b:0d:d3:37:d8:50:4e:c5:3a:34",algorithm="rsa-sha256",signature="Mje8vIDPlwIHmD/cTDwRxE7HaAvBg16JnVcsuqaNRim23fFPgQfLoOOxae6WqKb1uPjYEl0qIdazWaBy/Ml8DRhqlocMwoSXv0fbukP8J5N80LCmzT/FFBvIvTB91XuXI3hYfP9Zt1l7S6ieVadHUfqBedWH0itrtPJBgKmrWso="`
	assert.Equal(t, expectedAuthHdrVal, authHdrVal)
}

// TestSignRequestRSAPSSSHA256ValidGet tests signing a GET request using the 'rsa-pss-sha256' signing algorithm
// It uses values from https://bitbucket.oci.oraclecorp.com/projects/SDK/repos/signing-examples/browse/raw/example
func TestSignRequestRSASPSSHA256ValidGet(t *testing.T) {
	target := "https://iaas.us-phoenix-1.oraclecloud.com/20160918/instances?availabilityDomain=Pjwf%3A%20PHX-AD-1&compartmentId=ocid1.compartment.oc1..aaaaaaaam3we6vgnherjq5q2idnccdflvjsnog7mlr6rtdb25gilchfeyjxa&displayName=TeamXInstances&volumeId=ocid1.volume.oc1.phx.abyhqljrgvttnlx73nmrwfaux7kcvzfs3s66izvxf2h4lgvyndsdsnoiwr5q"
	key, _ := NewPKCS1RSAPrivateKeyFromPEM(testPrivateKey)
	supplier, _ := NewStaticRSAKeySupplier(key, testKeyID)
	request := httptest.NewRequest("GET", target, nil)
	headersToSign := []string{"Date", "(request-target)", "Host"}
	request.Header.Set("Date", "Thu, 05 Jan 2014 21:31:40 GMT")
	singedRequest, err := SignRequest(request, testKeyID, supplier, headersToSign, NewtestAlgorithmRSAPSSSHA256())
	assert.Nil(t, err)
	authHdrVal := singedRequest.Header.Get("Authorization")
	expectedAuthHdrVal := `Signature version="1",headers="date (request-target) host",keyId="ocid1.tenancy.oc1..aaaaaaaaba3pv6wkcr4jqae5f15p2b2m2yt2j6rx32uzr4h25vqstifsfdsq/ocid1.user.oc1..aaaaaaaat5nvwcna5j6aqzjcaty5eqbb6qt2jvpkanghtgdaqedqw3rynjq/20:3b:97:13:55:1c:5b:0d:d3:37:d8:50:4e:c5:3a:34",algorithm="rsa-pss-sha256",signature="blb25qN/kzlRMmgywt4nTwvuk5FMI84n+1v0YqWTnMcvJ8PVr10K5Y0yYS5YNzIGqCMMCnqGOk8A3ICrA0ngODsklJpmINPW89nDihqy1A4QSv2vqUWBDx2vEhmMRdxKouLMn0lwDJPeAjFckixxT3zZK77AddBTTZYb1HW+p4s="`
	assert.Equal(t, expectedAuthHdrVal, authHdrVal)
}

// TestSignRequestRSAPSSSHA256ValidPost tests signing a POST request using the 'rsa-pss-sha256' signing algorithm
// It uses values from https://bitbucket.oci.oraclecorp.com/projects/SDK/repos/signing-examples/browse/raw/example
func TestSignRequestRSAPSSSHA256ValidPost(t *testing.T) {
	target := "https://iaas.us-phoenix-1.oraclecloud.com/20160918/volumeAttachments"
	key, _ := NewPKCS1RSAPrivateKeyFromPEM(testPrivateKey)
	supplier, _ := NewStaticRSAKeySupplier(key, testKeyID)
	request := httptest.NewRequest("POST", target, strings.NewReader(testBody))
	headersToSign := []string{"Date", "(request-target)", "Host", "Content-Length", "Content-Type", "x-content-sha256"}
	request.Header.Set("Date", "Thu, 05 Jan 2014 21:31:40 GMT")
	request.Header.Set("Content-Length", "316")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-content-sha256", "V9Z20UJTvkvpJ50flBzKE32+6m2zJjweHpDMX/U4Uy0=")
	singedRequest, err := SignRequest(request, testKeyID, supplier, headersToSign, NewtestAlgorithmRSAPSSSHA256())
	assert.Nil(t, err)
	authHdrVal := singedRequest.Header.Get("Authorization")
	expectedAuthHdrVal := `Signature version="1",headers="date (request-target) host content-length content-type x-content-sha256",keyId="ocid1.tenancy.oc1..aaaaaaaaba3pv6wkcr4jqae5f15p2b2m2yt2j6rx32uzr4h25vqstifsfdsq/ocid1.user.oc1..aaaaaaaat5nvwcna5j6aqzjcaty5eqbb6qt2jvpkanghtgdaqedqw3rynjq/20:3b:97:13:55:1c:5b:0d:d3:37:d8:50:4e:c5:3a:34",algorithm="rsa-pss-sha256",signature="MLL5Kqtodor7j8Hm/6/lFQSiV+PaD2yrDBTd4UeSSqD/0JEhYhvo6e4UInO4d0RVYArvoUp+4DbXVfw64qxcwBcUjuwJtYuwsfp1i3GgwBdYdwLW30dpck6GCdNWnJkPjQ1Rv7PPtuSrb2CZpMZ+U7MmjiAvs6OkX5W8IYkxa7Q="`
	assert.Equal(t, expectedAuthHdrVal, authHdrVal)
}

//
// RequestSigner tests
//

var (
	testAlgorithm = AlgorithmRSAPSSSHA256
)

func TestNewRequestSignerHappyPath(t *testing.T) {
	t.Run(
		`should return valid RequestSigner for valid KeySupplier and Algorithm`,
		func(t *testing.T) {
			var rs = NewRequestSigner(testSupplier, testAlgorithm)
			var instance = rs.(*requestSigner)
			assert.Equal(t, testSupplier, instance.keySupplier)
			assert.Equal(t, testAlgorithm, instance.algorithm)
		})
}

func TestNewRequestSignerBadArgs(t *testing.T) {
	testIO := []struct {
		tc       string
		supplier KeySupplier
		alg      Algorithm
	}{
		{tc: `should panic when supplier is nil`,
			supplier: nil, alg: testAlgorithm},
		{tc: `should panic when algorithm is nil`,
			supplier: testSupplier, alg: nil},
		{tc: `should panic when both supplier and algorithm are nil`,
			supplier: nil, alg: nil},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			assert.Panics(t, func() { NewRequestSigner(test.supplier, test.alg) })
		})
	}
}

func TestRequestSignerSignRequest(t *testing.T) {
	t.Run(
		`should call httpsigner.SignRequest with member keySupplier and algorithm`,
		func(t *testing.T) {
			supplier := &MockKeySupplier{}
			alg := &mockAlg{}
			var signer = NewRequestSigner(supplier, alg)

			req := httptest.NewRequest("GET", "http://example.com", nil)
			headers := []string{HdrRequestTarget}
			sreq, err := signer.SignRequest(req, "key", headers)
			assert.Nil(t, sreq)
			assert.NotNil(t, err)
			assert.True(t, supplier.KeyCalled, `KeySupplier not called`)
			assert.True(t, alg.SignCalled, `Algorithm not called`)
		})
}
