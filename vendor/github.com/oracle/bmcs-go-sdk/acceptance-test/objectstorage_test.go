// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,objectstorage recording,all !recording

package acceptance

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MustWin/baremetal-sdk-go/acceptance-test/helpers"
)

func TestObjectStorageBucket(t *testing.T) {
	client := helpers.GetClient("fixtures/objectstorage/bucket")
	defer client.Stop()
	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err)
	bucket, err := client.CreateBucket(compartmentID, "bucketname", "mustwin", nil)
	require.NoError(t, err)
	defer func() {
		err := client.DeleteBucket("bucketname", "mustwin", nil)
		assert.NoError(t, err)
	}()
	assert.NotNil(t, bucket)

	helpers.Sleep(10 * time.Second)
	obj, err := client.PutObject("mustwin", "bucketname", "objectName", []byte("{\"content\": \"SomeContent\"}"), nil)
	require.NoError(t, err)
	defer func() {
		_, err := client.DeleteObject("mustwin", "bucketname", "objectName", nil)
		assert.NoError(t, err)
	}()
	assert.NotNil(t, obj)

}
