// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"net/http"
	"testing"

	"strings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type builderFnMatcher func(urlBuilderFn) bool

func matchBuilderFns(urlBuilder urlBuilderFn) builderFnMatcher {
	return func(ub urlBuilderFn) bool {
		return fmt.Sprint(urlBuilder) == fmt.Sprint(ub)
	}
}

type mockRequestOptions struct {
	mock.Mock
}

func (mr *mockRequestOptions) marshalURL(urlTemplate string, region string, b urlBuilderFn) (string, error) {
	args := mr.Called(b)
	return args.Get(0).(string), nil
}

func (mr *mockRequestOptions) marshalHeader() http.Header {
	args := mr.Called()
	return args.Get(0).(http.Header)
}

func (mr *mockRequestOptions) marshalBody() (body []byte, e error) {
	args := mr.Called()
	return args.Get(0).([]byte), nil
}

func getTestClient() (c *Client, e error) {
	return NewClient(
		"userOCID",
		"tenancyOCID",
		"fingerprint",
		PrivateKeyFilePath(getTestDataPEMPath()),
		PrivateKeyPassword("password"),
	)
}

func TestAddRequestHeaders(t *testing.T) {
	request, _ := http.NewRequest(http.MethodGet, "http://www.goo.com", nil)

	addRequiredRequestHeaders(request, "user-agent", []byte{})
	assert.Equal(t, request.Header.Get("content-type"), "application/json")
	assert.Equal(t, request.Header.Get("user-agent"), "user-agent")

	request, _ = http.NewRequest(http.MethodGet, "http://www.goo.com", nil)

	request.Header.Add("content-type", "something")

	addRequiredRequestHeaders(request, "", []byte{})

	assert.Equal(t, request.Header.Get("content-type"), "something")

	buffer := []byte("12345")
	body := bytes.NewBuffer(buffer)
	request, _ = http.NewRequest(http.MethodPost, "http://www.postme.com", body)

	addRequiredRequestHeaders(request, "", buffer)

	assert.Equal(t, request.Header.Get("content-length"), "5")

}

func TestGetSigningString(t *testing.T) {
	request, _ := http.NewRequest(
		http.MethodGet,
		"https://core.us-az-phoenix-1.oracleiaas.com/v1/instances?availabilityDomain=Pjwf%3A%20PHX-AD-1",
		nil,
	)

	request.Header.Add("date", "Thu, 05 Jan 2014 21:31:40 GMT")
	addRequiredRequestHeaders(request, "", []byte{})
	actual := getSigningString(request)
	expected := `date: Thu, 05 Jan 2014 21:31:40 GMT
(request-target): get /v1/instances?availabilityDomain=Pjwf%3A%20PHX-AD-1
host: core.us-az-phoenix-1.oracleiaas.com`

	if !assert.Equal(t, actual, expected) {
		t.Log("Actual   ", actual)
		t.Log("Expected ", expected)
	}

	buffer := []byte("{'foo':'bar'}")
	body := bytes.NewBuffer(buffer)
	request, _ = http.NewRequest(
		http.MethodPost,
		"https://core.us-az-phoenix-1.oracleiaas.com/v1/instances?availabilityDomain=Pjwf%3A%20PHX-AD-1",
		body,
	)

	request.Header.Add("date", "Thu, 05 Jan 2014 21:31:40 GMT")
	addRequiredRequestHeaders(request, "", buffer)
	actual = getSigningString(request)
	expected = "date: Thu, 05 Jan 2014 21:31:40 GMT\n" +
		"(request-target): post /v1/instances?availabilityDomain=Pjwf%3A%20PHX-AD-1\n" +
		"host: core.us-az-phoenix-1.oracleiaas.com\n" +
		"content-length: 13\n" +
		"content-type: application/json\n" +
		"x-content-sha256: " + getBodyHash([]byte("{'foo':'bar'}"))

	if !assert.Equal(t, actual, expected) {
		t.Log("Actual   ", actual)
		t.Log("Expected ", expected)
	}

}

func TestCreateAuthorizationHeader(t *testing.T) {

	testGetURI := "https://core.us-az-phoenix-1.oracleiaas.com/v1/instances" +
		"?availabilityDomain=Pjwf%3A%20PHX-AD-1" +
		"&compartmentId=ocid1.compartment.oc1..aaaaaaaayzim47sto5wqh5d4vugrsx566gjqmflvhlifte3p5ez3miy6e4lq" +
		"&displayName=TeamXInstances" +
		"&volumeId=ocid1.volume.oc1.phx.abyhqljrav2k323acohquoxszz2zyh5vj5v2gbvntg7ifd4ndusyvr332whq"

	expected := "Signature version=\"1\",headers=\"date (request-target) host\",keyId=\"ocid1.tenancy.oc1..aaaaaaaaq3hulfjvrouw3e6qx2ncxtp256aq7etiabqqtzunnhxjslzkfyxq/ocid1.user.oc1..aaaaaaaaflxvsdpjs5ztahmsf7vjxy5kdqnuzyqpvwnncbkfhavexwd4w5ra/b4:8a:7d:54:e6:81:04:b2:99:8e:b3:ed:10:e2:12:2b\",algorithm=\"rsa-sha256\",signature=\"A8PvWR1EshLFd6q6WuJZDKm7LUINhFdMrTxN1puVgZ58umMHn5Aja9ZAoeo7x6uMvRIaPVwGWY2OwF/BAJWnCoUn6hk/fhIRqY+gm3jU9Xmf15WUjgYwF+PbcueHw8WzLTa1zlsC4kkv4Bz/cuY8QvagWquwrQQ/9ZoxH/RMLMDJcSJHb8uqhj4n1seoUJ9ja8LddbQIkkgdtogs5nnDbk7ynwkVnRHwLVysVfoTip4Q7hNJe++zb1SEdRsz9wSuUDHgox9lTO0ZzSjRWiTNZ72BDW28Eln7WpLx6YvdvWI3YDA718AgJYWiwNbCcdcIiVHI56MX1pztve0Ru4JYqw==\""

	var privateKey *rsa.PrivateKey
	var e error
	pass := "password"

	if privateKey, e = PrivateKeyFromBytes(testPrivateKey, &pass); e != nil {
		t.Error("Couldn't create private key", e)
	}

	ai := &authenticationInfo{
		privateRSAKey:  privateKey,
		tenancyOCID:    testTenancyOCID,
		userOCID:       testUserOCID,
		keyFingerPrint: testKeyFingerPrint,
	}

	request, _ := http.NewRequest(
		http.MethodGet,
		testGetURI,
		nil,
	)
	// Set date to be the same date we used to generate test auth header, otherwise
	// sig will be different every time
	request.Header.Add("date", "Thu, 05 Jan 2014 21:31:40 GMT")

	e = createAuthorizationHeader(request, ai, "", []byte{})

	assert.Nil(t, e)

	authHeader := request.Header.Get("Authorization")

	if !assert.Equal(t, authHeader, expected) {
		t.Log("Actual   ", authHeader)
		t.Log("Expected ", expected)
	}
}

func TestAgainstSigningExample(t *testing.T) {
	// This test uses the example input and expected output from https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/signingrequests.htm#six.

	exampleKey := []byte(`-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`)

	privateKey, e := PrivateKeyFromBytes(exampleKey, nil)
	if e != nil {
		t.Error("Couldn't create private key", e)
	}

	ai := &authenticationInfo{
		privateRSAKey:  privateKey,
		tenancyOCID:    "ocid1.tenancy.oc1..aaaaaaaaba3pv6wkcr4jqae5f15p2b2m2yt2j6rx32uzr4h25vqstifsfdsq",
		userOCID:       "ocid1.user.oc1..aaaaaaaat5nvwcna5j6aqzjcaty5eqbb6qt2jvpkanghtgdaqedqw3rynjq",
		keyFingerPrint: "20:3b:97:13:55:1c:5b:0d:d3:37:d8:50:4e:c5:3a:34",
	}

	body := []byte(`{
    "compartmentId": "ocid1.compartment.oc1..aaaaaaaam3we6vgnherjq5q2idnccdflvjsnog7mlr6rtdb25gilchfeyjxa",
    "instanceId": "ocid1.instance.oc1.phx.abuw4ljrlsfiqw6vzzxb43vyypt4pkodawglp3wqxjqofakrwvou52gb6s5a",
    "volumeId": "ocid1.volume.oc1.phx.abyhqljrgvttnlx73nmrwfaux7kcvzfs3s66izvxf2h4lgvyndsdsnoiwr5q"
}`)

	request, _ := http.NewRequest(
		http.MethodPost,
		"https://iaas.us-phoenix-1.oraclecloud.com/20160918/volumeAttachments",
		bytes.NewBuffer(body),
	)
	request.Header.Add("Date", "Thu, 05 Jan 2014 21:31:40 GMT")
	addRequiredRequestHeaders(request, "", body)

	e = createAuthorizationHeader(request, ai, "", []byte{})

	expectedAuthHeader := `Signature version="1",headers="date (request-target) host content-length content-type x-content-sha256",keyId="ocid1.tenancy.oc1..aaaaaaaaba3pv6wkcr4jqae5f15p2b2m2yt2j6rx32uzr4h25vqstifsfdsq/ocid1.user.oc1..aaaaaaaat5nvwcna5j6aqzjcaty5eqbb6qt2jvpkanghtgdaqedqw3rynjq/20:3b:97:13:55:1c:5b:0d:d3:37:d8:50:4e:c5:3a:34",algorithm="rsa-sha256",signature="Mje8vIDPlwIHmD/cTDwRxE7HaAvBg16JnVcsuqaNRim23fFPgQfLoOOxae6WqKb1uPjYEl0qIdazWaBy/Ml8DRhqlocMwoSXv0fbukP8J5N80LCmzT/FFBvIvTB91XuXI3hYfP9Zt1l7S6ieVadHUfqBedWH0itrtPJBgKmrWso="`
	assert.NoError(t, e)
	assert.Equal(t, request.Header.Get("Content-Length"), "316")
	assert.Equal(t, expectedAuthHeader, request.Header.Get("Authorization"))

	request, _ = http.NewRequest(
		http.MethodGet,
		"https://iaas.us-phoenix-1.oraclecloud.com/20160918/instances"+
			"?availabilityDomain=Pjwf%3A%20PHX-AD-1&"+
			"compartmentId=ocid1.compartment.oc1..aaaaaaaam3we6vgnherjq5q2idnccdflvjsnog7mlr6rtdb25gilchfeyjxa"+
			"&displayName=TeamXInstances&volumeId=ocid1.volume.oc1.phx.abyhqljrgvttnlx73nmrwfaux7kcvzfs3s66izvxf2h4lgvyndsdsnoiwr5q",
		nil,
	)
	request.Header.Add("Date", "Thu, 05 Jan 2014 21:31:40 GMT")
	addRequiredRequestHeaders(request, "", nil)

	e = createAuthorizationHeader(request, ai, "", []byte{})

	expectedAuthHeader = strings.Replace(`Signature version="1",headers="date (request-target) host",keyId="ocid1.t
enancy.oc1..aaaaaaaaba3pv6wkcr4jqae5f15p2b2m2yt2j6rx32uzr4h25vqstifsfdsq/ocid1.user.oc1..aaaaaaaat5nvwcna5j6aqzjcaty5eqbb6qt2jvpkanghtgdaqedqw3ryn
jq/20:3b:97:13:55:1c:5b:0d:d3:37:d8:50:4e:c5:3a:34",algorithm="rsa-sha256",signature="GBas7grhyrhSKHP6AVIj/h5/Vp8bd/peM79H9Wv8kjoaCivujVXlpbKLjMPe
DUhxkFIWtTtLBj3sUzaFj34XE6YZAHc9r2DmE4pMwOAy/kiITcZxa1oHPOeRheC0jP2dqbTll
8fmTZVwKZOKHYPtrLJIJQHJjNvxFWeHQjMaR7M="`, "\n", "", -1)

	assert.NoError(t, e)
	assert.Equal(t, expectedAuthHeader, request.Header.Get("Authorization"))
}

func TestConcatenateHeaders(t *testing.T) {
	headers := []string{
		"foo",
		"bar",
		"baz",
	}
	expected := "foo bar baz"
	actual := concatenateHeaders(headers)

	assert.Equal(t, actual, expected)

	headers = []string{"foo"}

	expected = "foo"
	actual = concatenateHeaders(headers)

	assert.Equal(t, actual, expected)

	headers = []string{}
	expected = ""
	actual = concatenateHeaders(headers)

	assert.Equal(t, actual, expected)

}
