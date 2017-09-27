// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package acceptance

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oracle/bmcs-go-sdk/acceptance-test/helpers"
)

func TestBucketCrud(t *testing.T) {
	client := helpers.GetClient("fixtures/bucket")
	defer client.Stop()
	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err, "Setup Compartment")

	// Create Bucket
	bucket, err := client.CreateBucket(compartmentID, "bucketname", "mynamespace", nil)
	assert.NoError(t, err, "Create Bucket")
	assert.NotNil(t, bucket, "Create Bucket")

	helpers.Sleep(10 * time.Second)

	// TODO: Get Bucket
	// TODO: Update Bucket
	// TODO: Head Bucket

	// Put Object
	obj, err := client.PutObject("mynamespace", "bucketname", "objectName", []byte("{\"content\": \"SomeContent\"}"), nil)
	assert.NoError(t, err, "Put Object")
	assert.NotNil(t, obj, "Put Object")

	// TODO: List Objects
	// TODO: Get Object
	// TODO: Head Object

	// Delete Object
	_, err = client.DeleteObject("mynamespace", "bucketname", "objectName", nil)
	assert.NoError(t, err, "Delete Object")

	// Delete Bucket
	err = client.DeleteBucket("bucketname", "mynamespace", nil)
	assert.NoError(t, err, "Delete Bucket")
}
