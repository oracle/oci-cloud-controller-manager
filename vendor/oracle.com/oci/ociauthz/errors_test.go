// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceResponseError(t *testing.T) {
	urlString := "http://localhost/woot"
	url, _ := url.Parse(urlString)
	goodHeaders := http.Header{}
	goodHeaders.Add("opc-request-id", "123456")

	testIO := []struct {
		tc                string
		response          *http.Response
		expectedStatus    string
		expectedTarget    string
		expectedRequestID string
	}{
		{tc: `should use default values when the response is nil`,
			response:          nil,
			expectedStatus:    "unknown",
			expectedTarget:    "unknown",
			expectedRequestID: "unknown"},
		{tc: `should use default values for an empty response`,
			response:          &http.Response{},
			expectedStatus:    "unknown",
			expectedTarget:    "unknown",
			expectedRequestID: "unknown"},
		{tc: `should use the default target for a request with no information about the original request`,
			response:          &http.Response{Status: "200 OK", Header: goodHeaders},
			expectedStatus:    "200 OK",
			expectedTarget:    "unknown",
			expectedRequestID: goodHeaders.Get("opc-request-id")},
		{tc: `should use the default target for a request with no information about the original request URL`,
			response:          &http.Response{Status: "200 OK", Header: goodHeaders, Request: &http.Request{}},
			expectedStatus:    "200 OK",
			expectedTarget:    "unknown",
			expectedRequestID: goodHeaders.Get("opc-request-id")},
		{tc: `should have error message with default target for a request with a nil Request URL`,
			response:          &http.Response{Status: "200 OK", Header: goodHeaders, Request: &http.Request{URL: nil}},
			expectedStatus:    "200 OK",
			expectedTarget:    "unknown",
			expectedRequestID: goodHeaders.Get("opc-request-id")},
		{tc: `should have error message with default status code for a response which does not contain the status code for some reason`,
			response:          &http.Response{Request: &http.Request{URL: url}, Header: goodHeaders},
			expectedStatus:    "unknown",
			expectedTarget:    urlString,
			expectedRequestID: goodHeaders.Get("opc-request-id")},
		{tc: `should use the default header for an empty header`,
			response:          &http.Response{Status: "404 Not Found", Request: &http.Request{URL: url}, Header: http.Header{}},
			expectedStatus:    "404 Not Found",
			expectedTarget:    urlString,
			expectedRequestID: "unknown"},
		{tc: `should have error message with expected message for a proper Response object`,
			response:          &http.Response{Status: "404 Not Found", Request: &http.Request{URL: url}, Header: goodHeaders},
			expectedStatus:    "404 Not Found",
			expectedTarget:    urlString,
			expectedRequestID: goodHeaders.Get("opc-request-id")},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			err := &ServiceResponseError{test.response}
			msg := err.Error()
			expected := fmt.Sprintf(serviceResponseTemplate, test.expectedStatus, test.expectedTarget, test.expectedRequestID)
			assert.Equal(t, expected, msg)
		})
	}
}
