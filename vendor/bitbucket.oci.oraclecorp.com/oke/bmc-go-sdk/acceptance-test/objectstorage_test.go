// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,objectstorage recording,all !recording

package acceptance

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"time"

	bm "github.com/MustWin/baremetal-sdk-go"
)

type ObjectstorageTestSuite struct {
	suite.Suite
	compartmentID string
}

func (s *ObjectstorageTestSuite) SetupSuite() {
	client := getClient("fixtures/load_balancer/setup")
	defer client.Stop()
	// get a compartment, any compartment
	var listOpts bm.ListOptions
	listOpts.Limit = 1
	list, err := client.ListCompartments(&listOpts)
	s.Require().NoError(err)
	if len(list.Compartments) == 1 {
		s.compartmentID = list.Compartments[0].ID
	} else {
		id, err := resourceApply(createCompartment(client))
		s.Require().NoError(err)
		s.compartmentID = id
	}

}

func (s *ObjectstorageTestSuite) TestBucket() {
	client := getClient("fixtures/objectstorage/bucket")
	defer client.Stop()
	bucket, err := client.CreateBucket(s.compartmentID, "bucketname", "mustwin", nil)
	s.Require().NoError(err)
	defer func() {
		err := client.DeleteBucket("bucketname", "mustwin", nil)
		s.NoError(err)
	}()
	s.NotNil(bucket)

	time.Sleep(10 * time.Second)
	obj, err := client.PutObject("mustwin", "bucketname", "objectName", []byte("{\"content\": \"SomeContent\"}"), nil)
	s.Require().NoError(err)
	defer func() {
		_, err := client.DeleteObject("mustwin", "bucketname", "objectName", nil)
		s.NoError(err)
	}()
	s.NotNil(obj)

}

func TestObjectStorageTestSuite(t *testing.T) {
	suite.Run(t, new(ObjectstorageTestSuite))
}
