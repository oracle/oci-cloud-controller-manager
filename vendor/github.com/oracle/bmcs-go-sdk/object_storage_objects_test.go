// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *ObjectStorageTestSuite) TestListObject() {
	var reqs interface{}
	namespace := Namespace("namespace")
	bucket := "bucket"
	opts := &ListObjectsOptions{}
	details := &requestDetails{
		ids: urlParts{

			namespace,
			resourceBuckets,
			bucket,
			resourceObjects,
		},
		optional: opts,
		required: reqs,
	}

	expected := &ListObjects{
		Objects: []ObjectSummary{
			{
				Name:        "name",
				Size:        150,
				MD5:         "md5",
				TimeCreated: time.Now(),
			},
		},
		Prefixes:      []string{""},
		NextStartWith: "namea",
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(expected),
	}
	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, err := s.requestor.ListObjects(namespace, bucket, opts)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(expected.Objects[0].Name, actual.Objects[0].Name)
}

func (s *ObjectStorageTestSuite) TestGetObject() {
	namespaceID := Namespace("namespace1")
	bucketID := "bucket1"
	objectID := "object1"
	opts := &GetObjectOptions{}
	var reqs interface{}
	body := []byte("testBody")

	object := &Object{
		Body: body,
	}

	details := &requestDetails{
		ids:      urlParts{namespaceID, resourceBuckets, bucketID, resourceObjects, objectID},
		optional: opts,
		required: reqs,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	headers.Set(headerContentLength, "")
	resp := &response{
		body:   object.Body,
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetObject(namespaceID, bucketID, objectID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(object.Body, actual.Body)
	s.Equal("ETAG", actual.ETag)
}

func (s *ObjectStorageTestSuite) TestDeleteObject() {
	namespaceID := Namespace("namespace1")
	bucketID := "bucket1"
	objectID := "object1"
	opts := &DeleteObjectOptions{}
	var reqs interface{}

	object := &DeleteObject{}

	details := &requestDetails{
		ids:      urlParts{namespaceID, resourceBuckets, bucketID, resourceObjects, objectID},
		optional: opts,
		required: reqs,
	}

	resp := &response{
		body: marshalObjectForTest(object),
	}

	s.requestor.On("request", http.MethodDelete, details).Return(resp, nil)

	actual, e := s.requestor.DeleteObject(namespaceID, bucketID, objectID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
}

func (s *ObjectStorageTestSuite) TestHeadObject() {
	namespaceID := Namespace("namespace1")
	bucketID := "bucket1"
	objectID := "object1"
	opts := &HeadObjectOptions{}
	var reqs interface{}

	object := &HeadBucket{}
	object.ETag = "ETAG"
	object.ClientRequestID = "reqID"

	details := &requestDetails{
		ids:      urlParts{namespaceID, resourceBuckets, bucketID, resourceObjects, objectID},
		optional: opts,
		required: reqs,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	headers.Set(headerContentLength, "0")
	resp := &response{
		body:   marshalObjectForTest(object),
		header: headers,
	}

	s.requestor.On("request", http.MethodHead, details).Return(resp, nil)

	actual, e := s.requestor.HeadObject(namespaceID, bucketID, objectID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(object.ETag, actual.ETag)
}

// &baremetal.requestDetails{ids:baremetal.urlParts{"n", "namespace1", "b", "bucket1", "o", "object1"}, name:"", optional:(*baremetal.PutObjectOptions)(0xc82035a870), required:struct { baremetal.bodyRequirement; ContentLength uint64 "header:\"Content-Lengt
// h\" json:\"-\" url:\"-\"" }{bodyRequirement:baremetal.bodyRequirement{Body:[]uint8{0x74, 0x65, 0x73, 0x74, 0x42, 0x6f, 0x64, 0x79}}, ContentLength:0x8}}

func (s *ObjectStorageTestSuite) TestPutObject() {
	namespaceID := Namespace("namespace1")
	bucketID := "bucket1"
	objectID := "object1"

	content := []byte("testBody")

	required := struct {
		bodyRequirement
		ContentLength uint64 `header:"Content-Length" json:"-" url:"-"`
	}{
		ContentLength: uint64(len(content)),
	}
	required.Body = content

	opts := &PutObjectOptions{}

	details := &requestDetails{
		ids:      urlParts{namespaceID, resourceBuckets, bucketID, resourceObjects, objectID},
		optional: opts,
		required: required,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	headers.Set(headerContentLength, "8")
	resp := &response{
		body:   required.Body,
		header: headers,
	}

	s.requestor.On("request", http.MethodPut, details).Return(resp, nil)

	actual, e := s.requestor.PutObject(namespaceID, bucketID, objectID, content, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(required.Body, actual.Body)
	s.Equal("ETAG", actual.ETag)
}
