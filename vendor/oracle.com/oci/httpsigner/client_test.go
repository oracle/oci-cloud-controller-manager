// Copyright (c) 2017 Oracle and/or its affiliates. All rights reserved.

package httpsigner

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errTestSigning = errors.New("Test Signing Error")
var testSigningFunc = func(*http.Request) (*http.Request, []string, error) {
	return nil, []string{}, errTestSigning
}

func TestNewClientValidSigner(t *testing.T) {
	var testSigner RequestSigner = &requestSigner{}
	t.Run(
		`should return a client with non-nil signer`,
		func(t *testing.T) {
			testClient := &http.Client{Timeout: 10}
			c := NewClient(testSigner, testKeyID, testClient, testSigningFunc)
			assert.NotNil(t, c)

			// Access fields
			_, signedErr := c.(*DefaultSigningClient).SignRequest(nil)
			assert.Equal(t, testClient, c.(*DefaultSigningClient).httpClient)
			assert.Equal(t, errTestSigning, signedErr)
			assert.Equal(t, testSigner, c.(*DefaultSigningClient).signer)
			assert.Equal(t, testKeyID, c.(*DefaultSigningClient).keyID)
			assert.Equal(t, testSigner, c.(*DefaultSigningClient).Signer())
			assert.Equal(t, testKeyID, c.(*DefaultSigningClient).KeyID())

			// SetKeyID
			newKeyID := "new-key-id"
			c.(*DefaultSigningClient).SetKeyID(newKeyID)
			assert.Equal(t, newKeyID, c.(*DefaultSigningClient).keyID)
			assert.Equal(t, newKeyID, c.(*DefaultSigningClient).KeyID())
		})
}

func TestNewClientNilSigner(t *testing.T) {
	t.Run(
		`should panic when passed a nil signer`,
		func(t *testing.T) {
			assert.Panics(t, func() {
				NewClient(nil, "", &http.Client{}, testSigningFunc)
			})
		})
}

func TestNewClientNilClient(t *testing.T) {
	var testSigner RequestSigner = &requestSigner{}
	t.Run(
		`should panic when passed a nil client`,
		func(t *testing.T) {
			assert.Panics(t, func() {
				NewClient(testSigner, "", nil, testSigningFunc)
			})
		})
}

func TestNewClientNilSignerFunc(t *testing.T) {
	var testSigner RequestSigner = &requestSigner{}
	t.Run(
		`should panic when passed a nil signing func`,
		func(t *testing.T) {
			assert.Panics(t, func() {
				NewClient(testSigner, "", &http.Client{}, nil)
			})
		})
}

func TestNewSimpleClientValidSigner(t *testing.T) {
	var testSigner RequestSigner = &requestSigner{}
	t.Run(
		`should return a client with non-nil signer`,
		func(t *testing.T) {
			c := NewSimpleClient(testSigner, testKeyID)
			assert.NotNil(t, c)

			// Access fields
			assert.Equal(t, testSigner, c.(*DefaultSigningClient).signer)
			assert.Equal(t, testKeyID, c.(*DefaultSigningClient).keyID)
			assert.Equal(t, testSigner, c.(*DefaultSigningClient).Signer())
			assert.Equal(t, testKeyID, c.(*DefaultSigningClient).KeyID())

			// SetKeyID
			newKeyID := "new-key-id"
			c.(*DefaultSigningClient).SetKeyID(newKeyID)
			assert.Equal(t, newKeyID, c.(*DefaultSigningClient).keyID)
			assert.Equal(t, newKeyID, c.(*DefaultSigningClient).KeyID())
		})
}

func TestNewSimpleClientNilSigner(t *testing.T) {
	t.Run(
		`should panic when passed a nil signer`,
		func(t *testing.T) {
			assert.Panics(t, func() {
				NewSimpleClient(nil, "")
			})
		})
}

func TestClientSignRequest(t *testing.T) {
	replacementKey := "replacement-key"

	testIO := []struct {
		tc            string
		signingError  error
		expectedError error
	}{
		{tc: `should pass args through to signer and return non-error result`,
			signingError: nil, expectedError: nil},
		{tc: `should pass args through to signer and return error`,
			signingError: errors.New("test error"), expectedError: errors.New("test error")},
		{tc: `should handle key expiration and set the rotated keyid`,
			signingError:  NewKeyRotationError(replacementKey, testKeyID),
			expectedError: NewKeyRotationError(replacementKey, testKeyID),
		},
		{tc: `should return correct error when rotated keyid is empty`,
			signingError:  NewKeyRotationError("", testKeyID),
			expectedError: ErrReplacementKeyIDEmpty,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			// setup
			ms := &MockRequestSigner{signingError: test.signingError}
			client := NewSimpleClient(ms, testKeyID).(SigningClient)
			req := httptest.NewRequest("GET", "http://localhost", nil)

			// go
			sreq, err := client.SignRequest(req)

			// verify results
			if test.signingError == nil {
				assert.Equal(t, req, sreq)
				assert.Nil(t, err)
				assert.Equal(t, testKeyID, ms.ProfferedKey)
			} else {
				assert.Equal(t, test.expectedError, err)
				if expired, ok := err.(*KeyRotationError); ok {
					assert.Equal(t, expired.ReplacementKeyID, client.(*DefaultSigningClient).keyID)
					assert.Equal(t, expired.OldKeyID, testKeyID)
				} else {
					assert.Nil(t, sreq)
					assert.Equal(t, testKeyID, ms.ProfferedKey)
				}
			}

			// verify mock
			assert.True(t, ms.SignRequestCalled)
			assert.Equal(t, req, ms.ProfferedRequest)
			assert.Equal(t, DefaultHeadersToSign, ms.ProfferedHeaders)
		})
	}
}

func TestClientDo(t *testing.T) {
	testError := errors.New("test error")
	testIO := []struct {
		tc            string
		signer        *MockRequestSigner
		expectedError error
	}{
		{
			tc:            `should pass signed request to underlying client on valid request`,
			signer:        &MockRequestSigner{},
			expectedError: nil,
		},
		{
			tc:            `should return signing error without calling underlying client`,
			signer:        &MockRequestSigner{signingError: testError},
			expectedError: testError,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			// setup
			client := NewSimpleClient(test.signer, testKeyID)
			req := httptest.NewRequest("GET", "http://localhost", nil)

			// Go
			_, err := client.Do(req)

			// verify
			assert.True(t, test.signer.SignRequestCalled)
			if test.expectedError != nil {
				assert.Equal(t, test.expectedError, err)
			}
		})
	}
}
