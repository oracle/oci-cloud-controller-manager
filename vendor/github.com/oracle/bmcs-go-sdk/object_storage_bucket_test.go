// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *ObjectStorageTestSuite) TestCreateBucket() {
	metadata := map[string]string{
		"foo": "bar",
	}
	timeCreated := Time{
		Time: time.Now(),
	}

	bucket := &Bucket{
		Namespace:     "namespace",
		Name:          "name",
		CompartmentID: "compartmentID",
		Metadata:      metadata,
		CreatedBy:     "userOCID",
		TimeCreated:   timeCreated,
		AccessType:    ObjectRead,
	}

	opts := &CreateBucketOptions{
		Metadata:   metadata,
		AccessType: ObjectRead,
	}

	required := struct {
		ocidRequirement
		Name string `header:"-" json:"name" url:"-"`
	}{
		Name: bucket.Name,
	}
	required.CompartmentID = bucket.CompartmentID

	details := &requestDetails{
		ids:      urlParts{bucket.Namespace, resourceBuckets},
		optional: opts,
		required: required,
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(bucket),
	}
	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, err := s.requestor.CreateBucket(
		bucket.CompartmentID,
		bucket.Name,
		bucket.Namespace,
		opts,
	)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(bucket.CompartmentID, actual.CompartmentID)
	s.Equal(bucket.AccessType, actual.AccessType)
}

func (s *ObjectStorageTestSuite) TestGetBucket() {
	bucket := &Bucket{
		Namespace:     "namespace",
		Name:          "name",
		CompartmentID: "compartmentID",
		Metadata:      nil,
		CreatedBy:     "userOCID",
		TimeCreated:   Time{Time: time.Now()},
		AccessType:    NoPublicAccess,
	}

	details := &requestDetails{
		ids: urlParts{bucket.Namespace, resourceBuckets, bucket.Name},
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	resp := &response{
		body:   marshalObjectForTest(bucket),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetBucket(bucket.Name, bucket.Namespace)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(bucket.Namespace, actual.Namespace)
	s.Equal(bucket.AccessType, NoPublicAccess)
	s.Equal("ETAG", actual.ETag)
}

func (s *ObjectStorageTestSuite) TestUpdateBucket() {
	metadata := map[string]string{
		"foo": "bar",
	}

	timeCreated := Time{
		Time: time.Now(),
	}

	bucket := &Bucket{
		Namespace:     "namespace",
		Name:          "name",
		CompartmentID: "compartmentID",
		Metadata:      metadata,
		CreatedBy:     "userOCID",
		TimeCreated:   timeCreated,
		AccessType:    NoPublicAccess,
	}

	respHeaders := http.Header{}
	respHeaders.Set(headerETag, "ETAG!")

	new_metadata := map[string]string{
		"foo": "bar!",
	}
	opts := &UpdateBucketOptions{
		Metadata:   new_metadata,
		AccessType: ObjectRead,
	}

	required := struct {
		ocidRequirement
	}{}
	required.CompartmentID = bucket.CompartmentID

	details := &requestDetails{
		ids:      urlParts{bucket.Namespace, resourceBuckets, bucket.Name},
		optional: opts,
		required: required,
	}

	bucket_updated := &Bucket{
		Namespace:     "namespace",
		Name:          "name",
		CompartmentID: "compartmentID",
		Metadata:      new_metadata,
		CreatedBy:     "userOCID",
		TimeCreated:   timeCreated,
		AccessType:    ObjectRead,
	}

	resp := &response{
		header: respHeaders,
		body:   marshalObjectForTest(bucket_updated),
	}

	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, err := s.requestor.UpdateBucket(
		bucket.CompartmentID,
		bucket.Name,
		bucket.Namespace,
		opts,
	)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal("ETAG!", actual.ETag)
	s.Equal(new_metadata, actual.Metadata)
	s.Equal(bucket_updated.AccessType, actual.AccessType)
}

func (s *ObjectStorageTestSuite) TestDeleteBucket() {
	s.testDeleteResource(
		resourceBuckets, "name", "namespace", s.requestor.DeleteBucket,
	)
}
