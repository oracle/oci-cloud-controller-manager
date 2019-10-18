// Copyright (c) 2018, Oracle and/or its affiliates. All rights reserved.

package httpsigner

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetRequestBody(t *testing.T) {
	testIO := []struct {
		tc         string
		getbody    func() (io.ReadCloser, error)
		body       []byte
		getbodySet bool
	}{
		{
			tc:         `should not set GetBody if GetBody is not already set but Body is nil`,
			getbody:    nil,
			body:       nil,
			getbodySet: false,
		},
		{
			tc:         `should set GetBody if GetBody is not already set and empty Body exists`,
			getbody:    nil,
			body:       []byte{},
			getbodySet: true,
		},
		{
			tc:         `should set GetBody if GetBody is not already set and Body is not nil`,
			getbody:    nil,
			body:       []byte("check out my cool byte string"),
			getbodySet: true,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			var body io.ReadCloser
			if test.body != nil {
				body = ioutil.NopCloser(bytes.NewBuffer(test.body))
			}
			req := httptest.NewRequest(http.MethodPost, "/zones", nil)
			req.GetBody = test.getbody
			req.Body = body

			err := SetRequestGetBody(req)
			assert.Nil(t, err)

			if !test.getbodySet {
				assert.Nil(t, req.GetBody)
			} else {
				assert.NotNil(t, req.GetBody)

				result, e := ioutil.ReadAll(req.Body)
				assert.Nil(t, e)
				assert.Equal(t, test.body, result)

				result, e = ioutil.ReadAll(req.Body)
				assert.Nil(t, e)
				assert.Equal(t, []byte{}, result) // second time calling Request.Body will be empty

				// GetBody should work multiple times
				for i := 0; i < 2; i++ {
					b, e := req.GetBody()
					assert.Nil(t, e)

					result, e = ioutil.ReadAll(b)
					assert.Nil(t, e)
					assert.Equal(t, test.body, result)
				}
			}
		})
	}
}
