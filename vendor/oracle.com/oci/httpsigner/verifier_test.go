// Copyright (c) 2017 Oracle and/or its affiliates. All rights reserved.

package httpsigner

import (
	"crypto/rsa"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractKeyValue(t *testing.T) {
	testIO := []struct {
		tc       string
		input    string
		expKey   string
		expValue string
	}{
		{tc: `should return empty strings for empty string`,
			input: ``, expKey: ``, expValue: ``},
		{tc: `should return empty value for input with no =`,
			input: `key`, expKey: `key`, expValue: ``},
		{tc: `should return empty key for input with leading =`,
			input: `=value`, expKey: ``, expValue: `value`},
		{tc: `should return key, value with quotes stripped for standard field`,
			input: `key="value"`, expKey: `key`, expValue: `value`},
		{tc: `should return key, value for unquoted value`,
			input: `key=value`, expKey: `key`, expValue: `value`},
		{tc: `should return key, value for embedded equals and spaces`,
			input: `key="value with ="`, expKey: `key`, expValue: `value with =`},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			key, val := extractKeyValue(test.input)
			assert.Equal(t, test.expKey, key)
			assert.Equal(t, test.expValue, val)
		})
	}
}

// cannot take the address of a string literal
func strptr(s string) *string {
	return &s
}

func TestExtractSignatureFields(t *testing.T) {
	testIO := []struct {
		tc         string
		authzHdr   *string // nil for no header
		expHeaders []string
		expKeyID   string
		expSig     string
		expAlg     string
		expErr     error
	}{
		{
			tc:         `should return all values for valid header`,
			authzHdr:   strptr(`Signature version="1", headers="date (request-target)", keyid="key", algorithm="alg", signature="sig"`),
			expHeaders: []string{`date`, `(request-target)`},
			expKeyID:   `key`,
			expSig:     `sig`,
			expAlg:     `alg`,
			expErr:     nil,
		},
		{
			tc:         `should ErrMissingAuthzHeader when header not present`,
			authzHdr:   nil,
			expHeaders: nil,
			expKeyID:   ``,
			expSig:     ``,
			expAlg:     ``,
			expErr:     ErrMissingAuthzHeader,
		},
		{
			tc:         `should ErrMissingAuthzHeader for empty header`,
			authzHdr:   strptr(``),
			expHeaders: nil,
			expKeyID:   ``,
			expSig:     ``,
			expAlg:     ``,
			expErr:     ErrMissingAuthzHeader,
		},
		{
			tc:         `should ErrUnsupportedScheme for non-signature header`,
			authzHdr:   strptr(`Basic dXNlcm5hbWU6cGFzc3dvcmQ=`),
			expHeaders: nil,
			expKeyID:   ``,
			expSig:     ``,
			expAlg:     ``,
			expErr:     ErrUnsupportedScheme,
		},
		{
			tc:         `should return empty values for no signature fields`,
			authzHdr:   strptr(`Signature`),
			expHeaders: nil,
			expKeyID:   ``,
			expSig:     ``,
			expAlg:     ``,
			expErr:     nil,
		},
		{
			tc:         `should return partial fields when some fields missing`,
			authzHdr:   strptr(`Signature version="1", keyid="key"`),
			expHeaders: nil,
			expKeyID:   `key`,
			expSig:     ``,
			expAlg:     ``,
			expErr:     nil,
		},
		{
			tc:         `should ignore unknown fields`,
			authzHdr:   strptr(`Signature version="1",headers="date (request-target)",keyid="key",algorithm="alg",signature="sig",extra="value"`),
			expHeaders: []string{`date`, `(request-target)`},
			expKeyID:   `key`,
			expSig:     `sig`,
			expAlg:     `alg`,
			expErr:     nil,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			// setup
			request := httptest.NewRequest("GET", "http://localhost", nil)
			if test.authzHdr != nil {
				request.Header.Set(HdrAuthorization, *test.authzHdr)
			}

			// go
			headers, keyID, sig, algName, err := ExtractSignatureFields(request)

			// validate
			if test.expErr == nil {
				assert.Nil(t, err)
				assert.Equal(t, test.expHeaders, headers)
				assert.Equal(t, test.expKeyID, keyID)
				assert.Equal(t, test.expSig, sig)
				assert.Equal(t, test.expAlg, algName)
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, test.expErr, err)
			}
		})
	}
}

func TestVerifyRequest(t *testing.T) {
	testIO := []struct {
		tc              string
		requestHeaders  map[string]string
		keySupplier     *MockKeySupplier
		algSupplier     AlgorithmSupplier
		expKeySupCalled bool
		expAlgCalled    bool
		expErr          error
	}{
		{
			// happy path
			tc: `should return nil for when algorithm returns nil`,
			requestHeaders: map[string]string{
				HdrDate:          `1984`,
				HdrAuthorization: `Signature version="1",headers="Date",keyid="key",algorithm="test",signature=""`,
			},
			keySupplier:     &MockKeySupplier{},
			algSupplier:     Algorithms{`test`: &mockAlg{}},
			expKeySupCalled: true,
			expAlgCalled:    true,
			expErr:          nil,
		},
		{
			tc:              `should return ExtractSignatureFields errors`,
			requestHeaders:  map[string]string{},
			keySupplier:     &MockKeySupplier{},
			algSupplier:     Algorithms{`test`: &mockAlg{}},
			expKeySupCalled: false,
			expAlgCalled:    false,
			expErr:          ErrMissingAuthzHeader,
		},
		{
			tc: `should return KeySupplier errors`,
			requestHeaders: map[string]string{
				HdrDate:          `1984`,
				HdrAuthorization: `Signature version="1",headers="Date",keyid="key",algorithm="test",signature=""`,
			},
			keySupplier:     &MockKeySupplier{err: ErrKeyNotFound},
			algSupplier:     Algorithms{`test`: &mockAlg{}},
			expKeySupCalled: true,
			expAlgCalled:    false,
			expErr:          ErrKeyNotFound,
		},
		{
			tc: `should return AlgorithmSupplier errors`,
			requestHeaders: map[string]string{
				HdrDate:          `1984`,
				HdrAuthorization: `Signature version="1",headers="Date",keyid="key",algorithm="test",signature=""`,
			},
			keySupplier:     &MockKeySupplier{},
			algSupplier:     Algorithms{},
			expKeySupCalled: true,
			expAlgCalled:    false,
			expErr:          ErrUnsupportedAlgorithm,
		},
		{
			tc: `should return base64 decode errors`,
			requestHeaders: map[string]string{
				HdrDate:          `1984`,
				HdrAuthorization: `Signature version="1",headers="Date",keyid="key",algorithm="test",signature="?"`,
			},
			keySupplier:     &MockKeySupplier{},
			algSupplier:     Algorithms{`test`: &mockAlg{}},
			expKeySupCalled: true,
			expAlgCalled:    false,
			expErr:          base64.CorruptInputError(0),
		},
		{
			tc: `should return Algorithm.Verify errors`,
			requestHeaders: map[string]string{
				HdrDate:          `1984`,
				HdrAuthorization: `Signature version="1",headers="Date",keyid="key",algorithm="test",signature=""`,
			},
			keySupplier:     &MockKeySupplier{},
			algSupplier:     Algorithms{`test`: &mockAlg{verifyErr: rsa.ErrVerification}},
			expKeySupCalled: true,
			expAlgCalled:    true,
			expErr:          rsa.ErrVerification,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			// setup
			request := httptest.NewRequest("GET", "http://localhost", nil)
			for hdr, val := range test.requestHeaders {
				request.Header.Set(hdr, val)
			}

			// go
			err := VerifyRequest(request, test.keySupplier, test.algSupplier)

			// validation
			if test.expErr == nil {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, test.expErr, err)
			}

			// check mocks
			assert.Equal(t, test.expKeySupCalled, test.keySupplier.KeyCalled)
			alg, ok := test.algSupplier.(Algorithms)[`test`].(*mockAlg)
			if ok {
				assert.Equal(t, test.expAlgCalled, alg.VerifyCalled)
			}
		})
	}
}

func TestVerifyRequestBadArgs(t *testing.T) {
	testRequest := httptest.NewRequest("GET", "http://localhost", nil)
	testIO := []struct {
		tc          string
		req         *http.Request
		keySupplier KeySupplier
		algSupplier AlgorithmSupplier
		expErr      error
	}{
		{tc: `should return ErrInvalidRequest for nil request`,
			req: nil, keySupplier: nil, algSupplier: nil, expErr: ErrInvalidRequest},
		{tc: `should return ErrInvalidKeySupplier for nil key supplier`,
			req: testRequest, keySupplier: nil, algSupplier: nil, expErr: ErrInvalidKeySupplier},
		{tc: `should return ErrInvalidAlgorithmSupplier for nil algorithm supplier`,
			req: testRequest, keySupplier: &MockKeySupplier{}, algSupplier: nil, expErr: ErrInvalidAlgorithmSupplier},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			err := VerifyRequest(test.req, test.keySupplier, test.algSupplier)
			assert.Equal(t, test.expErr, err)
		})
	}
}

//
// RequestVerifier tests
//

func TestNewRequestVerifierHappyPath(t *testing.T) {
	t.Run(`should return valid RequestVerifier for non-nil key and algorithm suppliers`,
		func(t *testing.T) {
			var verifier = NewRequestVerifier(testSupplier, StdAlgorithms)
			var instance = verifier.(*requestVerifier)
			assert.Equal(t, testSupplier, instance.keySupplier)
			assert.Equal(t, StdAlgorithms, instance.algSupplier)
		})
}

func TestNewRequestVerifierBadArgs(t *testing.T) {
	testIO := []struct {
		tc string
		ks KeySupplier
		as AlgorithmSupplier
	}{
		{tc: `should panic when KeySupplier is nil`,
			ks: nil, as: StdAlgorithms},
		{tc: `should panic when AlgorithmSupplier is nil`,
			ks: testSupplier, as: nil},
		{tc: `should panic when both KeySupplier and AlgorithmSuplier are nil`,
			ks: nil, as: nil},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			assert.Panics(t, func() { NewRequestVerifier(test.ks, test.as) })
		})
	}
}

func TestRequestVerifierVerifyRequest(t *testing.T) {
	t.Run(
		`should call httpsigner.VerifyRequest with member keySupplier and algSupplier`,
		func(t *testing.T) {
			ks := &MockKeySupplier{}
			alg := &mockAlg{}
			as := Algorithms{"mock": alg}
			verifier := NewRequestVerifier(ks, as)

			req := httptest.NewRequest("GET", "http://example.com", nil)
			req.Header.Set(
				HdrAuthorization,
				`Signature version="1",headers="(request-target)",keyid="test",algorithm="mock",signature="1337"`,
			)
			err := verifier.VerifyRequest(req)
			assert.Nil(t, err)
			assert.True(t, ks.KeyCalled, `KeySupplier not called`)
			assert.True(t, alg.VerifyCalled, `AlgorithmSupplier not called`)
		})
}
