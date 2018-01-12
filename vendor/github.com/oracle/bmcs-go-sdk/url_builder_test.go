// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildIdentityURL(t *testing.T) {
	urlTemplate := "https://%s.%s.oraclecloud.com"
	region := "fake-region"
	baseUrl := baseUrlHelper(urlTemplate, identityServiceAPI, region)
	expected := baseUrl + "/" + SDKVersion + "/policies/"
	actual, _ := buildIdentityURL(urlTemplate, region, resourcePolicies, nil, "/")
	assert.Equal(t, expected, actual)

	expected = baseUrl + "/" + SDKVersion + "/policies?foo=bar%2Fbaz"
	actual, _ = buildIdentityURL(urlTemplate, region, resourcePolicies, url.Values{"foo": []string{"bar/baz"}})
	assert.Equal(t, expected, actual)

	expected = baseUrl + "/" + SDKVersion + "/policies/one/two?foo=bar%2Fbaz"
	actual, _ = buildIdentityURL(urlTemplate, region, resourcePolicies, url.Values{"foo": []string{"bar/baz"}}, "one", "two")
	assert.Equal(t, expected, actual)

	expected = baseUrl + "/" + SDKVersion + "/policies/one/two"
	actual, _ = buildIdentityURL(urlTemplate, region, resourcePolicies, nil, "one", "two")
	assert.Equal(t, expected, actual)

	expected = baseUrl + "/" + SDKVersion + "/policies/one/two/"
	actual, _ = buildIdentityURL(urlTemplate, region, resourcePolicies, nil, "one", "two", "/")
	assert.Equal(t, expected, actual)
}

func TestBuildObjectStorageURL(t *testing.T) {
	urlTemplate := "https://%s.%s.oraclecloud.com"
	region := "fake-region"
	baseUrl := baseUrlHelper(urlTemplate, objectStorageServiceAPI, region)

	expected := baseUrl + "/n/example_namespace/b"
	actual, _ := buildObjectStorageURL(urlTemplate, region, resourceBuckets, nil, "example_namespace", "b")
	assert.Equal(t, expected, actual)

	expected = baseUrl + "/n/example_namespace/b/foo"
	actual, _ = buildObjectStorageURL(urlTemplate, region, resourceBuckets, nil, "example_namespace", "b", "foo")
	assert.Equal(t, expected, actual)
}

func TestBuildCoreURL(t *testing.T) {
	urlTemplate := "https://%s.%s.oraclecloud.com"
	region := "fake-region"
	baseUrl := baseUrlHelper(urlTemplate, coreServiceAPI, region)

	expected := baseUrl + "/" + SDKVersion + "/cpes"
	actual, _ := buildCoreURL(urlTemplate, region, resourceCustomerPremiseEquipment, nil)
	assert.Equal(t, expected, actual)

	expected = baseUrl + "/" + SDKVersion + "/instanceConsoleHistories/12/data"
	actual, _ = buildCoreURL(urlTemplate, region, resourceInstanceConsoleHistories, nil, "12", "data")
	assert.Equal(t, expected, actual)
}

func TestUnparsableURL(t *testing.T) {
	urlTemplate := "%!@%s.%s"
	region := "fake-region"

	_, err := buildCoreURL(urlTemplate, region, resourceInstanceConsoleHistories, nil, "12", "data")
	assert.NotNil(t, err)
	urlError := err.(*url.Error)
	assert.Equal(t, "parse", urlError.Op)
}

func TestBadType(t *testing.T) {
	urlTemplate := "https://%s.%s.oraclecloud.com"
	region := "fake-region"

	// Floats are not a supported type in the "ids" field.
	_, err := buildCoreURL(urlTemplate, region, resourceInstanceConsoleHistories, nil, 1.23)
	assert.Error(t, err, "Unsupported type")
}
