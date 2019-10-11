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
	actual := buildIdentityURL(urlTemplate, region, resourcePolicies, nil, "/")
	assert.Equal(t, expected, actual)

	expected = baseUrl + "/" + SDKVersion + "/policies?foo=bar%2Fbaz"
	actual = buildIdentityURL(urlTemplate, region, resourcePolicies, url.Values{"foo": []string{"bar/baz"}})
	assert.Equal(t, expected, actual)

	expected = baseUrl + "/" + SDKVersion + "/policies/one/two?foo=bar%2Fbaz"
	actual = buildIdentityURL(urlTemplate, region, resourcePolicies, url.Values{"foo": []string{"bar/baz"}}, "one", "two")
	assert.Equal(t, expected, actual)

	expected = baseUrl + "/" + SDKVersion + "/policies/one/two"
	actual = buildIdentityURL(urlTemplate, region, resourcePolicies, nil, "one", "two")
	assert.Equal(t, expected, actual)

	expected = baseUrl + "/" + SDKVersion + "/policies/one/two/"
	actual = buildIdentityURL(urlTemplate, region, resourcePolicies, nil, "one", "two", "/")
	assert.Equal(t, expected, actual)
}

func TestBuildObjectStorageURL(t *testing.T) {
	urlTemplate := "https://%s.%s.oraclecloud.com"
	region := "fake-region"
	baseUrl := baseUrlHelper(urlTemplate, objectStorageServiceAPI, region)

	expected := baseUrl + "/n/example_namespace/b"
	actual := buildObjectStorageURL(urlTemplate, region, resourceBuckets, nil, "example_namespace", "b")
	assert.Equal(t, expected, actual)

	expected = baseUrl + "/n/example_namespace/b/foo"
	actual = buildObjectStorageURL(urlTemplate, region, resourceBuckets, nil, "example_namespace", "b", "foo")
	assert.Equal(t, expected, actual)
}

func TestBuildCoreURL(t *testing.T) {
	urlTemplate := "https://%s.%s.oraclecloud.com"
	region := "fake-region"
	baseUrl := baseUrlHelper(urlTemplate, coreServiceAPI, region)

	expected := baseUrl + "/" + SDKVersion + "/cpes"
	actual := buildCoreURL(urlTemplate, region, resourceCustomerPremiseEquipment, nil)
	assert.Equal(t, expected, actual)

	expected = baseUrl + "/" + SDKVersion + "/instanceConsoleHistories/12/data"
	actual = buildCoreURL(urlTemplate, region, resourceInstanceConsoleHistories, nil, "12", "data")
	assert.Equal(t, expected, actual)
}
