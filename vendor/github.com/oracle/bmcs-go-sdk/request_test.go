// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"net/http"
	"testing"

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
(request-target): get /v1/instances?availabilityDomain=Pjwf%3A%20PHX-AD-1`

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

	expected := "Signature headers=\"date (request-target)\",keyId=\"ocid1.tenancy.oc1..aaaaaaaaq3hulfjvrouw3e6qx2ncxtp256aq7etiabqqtzunnhxjslzkfyxq/ocid1.user.oc1..aaaaaaaaflxvsdpjs5ztahmsf7vjxy5kdqnuzyqpvwnncbkfhavexwd4w5ra/b4:8a:7d:54:e6:81:04:b2:99:8e:b3:ed:10:e2:12:2b\",algorithm=\"rsa-sha256\",signature=\"R5I2tf3iU5ExvtywRi2fj4YxjDBuhJT7TQwiK1XOU5Wf/hLgq25iLdX8YbRJWvOaLHOuDShhZeisODl/ksVSJISDArLe+cLailYmYPWB7T3987U7IgtbhgucHw4bY09MGoRn3rHEfWYTj16C4O2y7zMRmdUwt3f2ioAe1EFrn8bixEM+AavCU/ydLFCcxXr13pDSP+NAPvJ0dsyRyBzkYbuYPRulncBYEmFqVxFRARHzIAO7z0OBv8lkoGQTJhKI/5ZZxnYmYfwgvM6djK57QdoBSXyrcwi2BdeiBjdhLRphjbWmB5l0OlWeQo6sEEFVcGOzuxazO0XTRwbaiJYfng==\""

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
