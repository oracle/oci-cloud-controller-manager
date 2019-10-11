// Copyright (c) 2017 Oracle and/or its affiliates. All rights reserved.

package httpsigner

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	f        bool
	t        = true
	falsePtr = &f
	truePtr  = &t
)

// noopTransport is an http.RoundTripper that hijacks an outgoing http request
// and returns without having done anything except track that RoundTrip was
// called.
//
// Note that because RoundTrip does not accept a pointer receiver we must track
// roundTripCalled via a provided boolean pointer
type noopTransport struct {
	roundTripCalled *bool
}

// RoundTrip implements the http.RoundTripper interface
func (t noopTransport) RoundTrip(*http.Request) (*http.Response, error) {
	*t.roundTripCalled = true
	return &http.Response{}, nil
}

func safeNewRequest(method, u string, body io.Reader) *http.Request {
	r, _ := http.NewRequest(method, u, body)
	return r
}

func TestSignerTransport(t *testing.T) {
	testIO := []struct {
		tc              string
		inp             *http.Request
		roundTripCalled *bool
		signingError    error
		expectedError   error
	}{
		{
			tc:              "should sign read request without issue",
			inp:             safeNewRequest(http.MethodGet, "http://localhost", nil),
			roundTripCalled: truePtr,
			signingError:    nil,
			expectedError:   nil,
		},
		{
			tc:              "should sign write request without issue, ",
			inp:             safeNewRequest(http.MethodPut, "http://localhost", nil),
			roundTripCalled: truePtr,
			signingError:    nil,
			expectedError:   nil,
		},
		{
			tc:              "should pass args through to signer and return error",
			inp:             safeNewRequest(http.MethodGet, "http://localhost", nil),
			roundTripCalled: falsePtr,
			signingError:    errors.New("test error"),
			expectedError: &url.Error{
				Op:  "Get",
				URL: "http://localhost",
				Err: errors.New("test error"),
			},
		},
		{
			tc:              "should return correct error when rotated keyid is empty",
			inp:             safeNewRequest(http.MethodGet, "http://localhost", nil),
			roundTripCalled: falsePtr,
			signingError:    NewKeyRotationError("", testKeyID),
			expectedError: &url.Error{
				Op:  "Get",
				URL: "http://localhost",
				Err: ErrReplacementKeyIDEmpty,
			},
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			ms := &MockRequestSigner{signingError: test.signingError}
			c := NewSimpleClient(ms, testKeyID)

			embeddedTransport := noopTransport{falsePtr}
			client := &http.Client{
				Transport: NewTransport(c, embeddedTransport),
			}

			resp, err := client.Do(test.inp)

			assert.Equal(t, test.roundTripCalled, embeddedTransport.roundTripCalled)
			// verify results
			if test.signingError == nil {
				assert.Nil(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, testKeyID, ms.ProfferedKey)
			} else {
				assert.Nil(t, resp)
				assert.Equal(t, test.expectedError, err)
			}
		})
	}
}
