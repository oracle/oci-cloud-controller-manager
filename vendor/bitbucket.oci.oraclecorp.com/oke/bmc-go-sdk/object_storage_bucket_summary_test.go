// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *ObjectStorageTestSuite) TestListBuckets() {
	bucket := BucketSummary{
		Namespace:     "namespace",
		Name:          "name",
		CompartmentID: "compartment",
		CreatedBy:     "someone",
		TimeCreated:   time.Now(),
		ETag:          "string",
	}
	buckets := &ListBuckets{
		BucketSummaries: []BucketSummary{
			bucket,
		},
	}

	opts := &ListBucketsOptions{}

	required := listOCIDRequirement{}
	required.CompartmentID = bucket.CompartmentID

	details := &requestDetails{
		ids:      urlParts{bucket.Namespace, resourceBuckets},
		optional: opts,
		required: required,
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(buckets.BucketSummaries),
	}
	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, err := s.requestor.ListBuckets(
		bucket.CompartmentID,
		bucket.Namespace,
		opts,
	)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(buckets.BucketSummaries[0].Name, actual.BucketSummaries[0].Name)
}
