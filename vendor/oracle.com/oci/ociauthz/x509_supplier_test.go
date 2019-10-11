// Copyright (c) 2017-2019, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"testing"

	"oracle.com/oci/httpsigner"

	"bytes"
	"crypto/rand"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/stretchr/testify/assert"
)

var (
	emptyIntermediateCerts       = []*x509.Certificate{}
	testTenantID                 = "test-tenantid"
	testCertFromTestCertBytes, _ = parseCertificateByteArray(testCertBytes)
)

// Test that we can create an STSKeySupplier
func TestNewX509CertificateSupplier(t *testing.T) {
	testIO := []struct {
		tc            string
		tenantID      string
		certificate   *x509.Certificate
		intermediate  []*x509.Certificate
		private       *rsa.PrivateKey
		expectedError error
	}{
		{tc: "should return error with nil certificate",
			tenantID: testTenantID, certificate: nil, intermediate: emptyIntermediateCerts,
			private: testPrivateKey, expectedError: ErrInvalidCertificate},
		{tc: "should not return error with nil intermediate certificate",
			tenantID: testTenantID, certificate: testCert, intermediate: nil,
			private: testPrivateKey, expectedError: nil},
		{tc: "should return error with nil private key",
			tenantID: testTenantID, certificate: testCert, intermediate: emptyIntermediateCerts,
			private: nil, expectedError: ErrInvalidPrivateKey},
		{tc: "should return error with empty tenant-id string",
			tenantID: "", certificate: testCert, intermediate: emptyIntermediateCerts,
			private: testPrivateKey, expectedError: ErrInvalidTenantID},
		{tc: "should return x509 certificate supplier with valid input",
			tenantID: testTenantID, certificate: testCert, intermediate: emptyIntermediateCerts,
			private: testPrivateKey, expectedError: nil},
	}

	for _, test := range testIO {
		t.Run(fmt.Sprintf(test.tc), func(t *testing.T) {
			s, e := NewX509CertificateSupplier(test.tenantID, test.certificate, test.intermediate, test.private)
			if test.expectedError == nil {
				assert.Nil(t, e)
				assert.NotNil(t, s)
				c, ce := s.CertificateOrError()
				assert.NoError(t, ce)
				assert.Equal(t, s.Certificate(), test.certificate)
				assert.Equal(t, s.Certificate(), c)
				i, ie := s.IntermediateOrError()
				assert.NoError(t, ie)
				assert.Equal(t, s.Intermediate(), test.intermediate)
				assert.Equal(t, s.Intermediate(), i)
				p, pe := s.PrivateKeyOrError()
				assert.NoError(t, pe)
				assert.Equal(t, s.PrivateKey(), test.private)
				assert.Equal(t, s.PrivateKey(), p)
				private, err := s.Key(s.KeyID())
				assert.Nil(t, err)
				assert.Equal(t, private.(*rsa.PrivateKey), test.private)
			} else {
				assert.Equal(t, e, test.expectedError)
				assert.Nil(t, s)
			}
		})
	}
}

func TestKey(t *testing.T) {
	testIO := []struct {
		tc        string
		fakeKeyID bool
	}{
		{tc: `should return the key if the given keyID matches the one inside the supplier`,
			fakeKeyID: false},
		{tc: `should return ErrKeyNotFound if the given KeyID does not match the one inside the supplier`,
			fakeKeyID: true},
	}
	for _, test := range testIO {
		t.Run(fmt.Sprintf(test.tc), func(t *testing.T) {
			s, e := NewX509CertificateSupplier(testTenantID, testCert, emptyIntermediateCerts, testPrivateKey)
			assert.Nil(t, e)
			if test.fakeKeyID {
				k, e := s.Key("super-unique-key-id")
				assert.Nil(t, k)
				assert.Equal(t, e, httpsigner.ErrKeyNotFound)
			} else {
				k, e := s.Key(s.KeyID())
				assert.Nil(t, e)
				assert.Equal(t, k.(*rsa.PrivateKey), testPrivateKey)
			}
		})
	}
}

func TestX509KeyID(t *testing.T) {
	s, _ := NewX509CertificateSupplier(testTenantID, testCert, emptyIntermediateCerts, testPrivateKey)
	keyID := s.KeyID()
	parts := strings.Split(keyID, "/")

	assert.Len(t, parts, 3)
	assert.Equal(t, parts[0], testTenantID)
	assert.Equal(t, parts[1], "fed-x509")
	assert.NotEqual(t, parts[2], "")
}

func TestNewURLCertificateSupplierFromURLs(t *testing.T) {
	_, badURLErr := url.Parse("")

	testIO := []struct {
		tc               string
		expectError      error
		tenantID         string
		certificateURL   string
		privateKeyURL    string
		keypassphrase    []byte
		intermediateURLS []string
		client           httpsigner.Client
	}{
		{
			tc:               "should create supplier with proper data",
			expectError:      nil,
			tenantID:         "somevalidTenant",
			certificateURL:   "/cert/url",
			privateKeyURL:    "/key/url",
			keypassphrase:    nil,
			intermediateURLS: []string{"/inter/1", "/inter/2"},
			client:           http.DefaultClient,
		},
		{
			tc:               "should fail supplier due to bad url",
			expectError:      badURLErr,
			tenantID:         "somevalidTenant",
			certificateURL:   "/cert/url",
			privateKeyURL:    "/key/url",
			keypassphrase:    nil,
			intermediateURLS: []string{"/inter/1", ""},
			client:           http.DefaultClient,
		},
		{
			tc:               "should fail due to empty tenant ID",
			expectError:      ErrInvalidTenantID,
			tenantID:         "",
			certificateURL:   "/cert/url",
			privateKeyURL:    "/key/url",
			keypassphrase:    nil,
			intermediateURLS: []string{"/inter/1", "/inter/2"},
			client:           http.DefaultClient,
		},
		{
			tc:               "should fail due to nil http client",
			expectError:      ErrInvalidClient,
			tenantID:         "tenant",
			certificateURL:   "/cert/url",
			privateKeyURL:    "/key/url",
			keypassphrase:    nil,
			intermediateURLS: []string{"/inter/1", "/inter/2"},
			client:           nil,
		},
	}
	for _, test := range testIO {
		t.Run(fmt.Sprintf(test.tc), func(t *testing.T) {
			s, err := NewX509CertificateSupplierFromURLs(test.client, test.tenantID, test.certificateURL, test.privateKeyURL,
				test.keypassphrase, test.intermediateURLS...)
			if test.expectError != nil {
				assert.Nil(t, s)
				assert.Equal(t, test.expectError, err)
			} else {
				assert.NotNil(t, s)
			}
		})
	}
}

type mockedResponses struct {
	URL      string
	Err      error
	Response []byte
}

type mockClient struct {
	Responses []mockedResponses
}

func (m mockClient) Do(r *http.Request) (response *http.Response, err error) {
	for _, req := range m.Responses {
		if r.URL.Path == req.URL {
			body := bytes.NewBuffer(req.Response)
			statusCode := http.StatusOK
			if req.Err != nil {
				err = req.Err
				statusCode = http.StatusBadRequest
			}
			response = &http.Response{
				Header:     http.Header{},
				StatusCode: statusCode,
				Body:       ioutil.NopCloser(body),
			}
			return
		}
	}
	return nil, fmt.Errorf("not a valid request")
}

type expectedResponse struct {
	Value interface{}
	Err   error
}

func TestURLX509CertificateSupplier(t *testing.T) {
	testErr := &ServiceResponseError{Response: nil}
	randBytes := make([]byte, 100)
	rand.Read(randBytes)
	var nilCert *x509.Certificate
	var nilKey *rsa.PrivateKey
	var nilIntemediates []*x509.Certificate

	testIO := []struct {
		tc                           string
		tenantID                     string
		expectError                  bool
		CertURL                      string
		PrivateKeyURL                string
		IntermediateURLs             []string
		expectedCertResponse         expectedResponse
		expectedPrivateKeyResponse   expectedResponse
		expectedIntermediateResponse expectedResponse
		customClient                 httpsigner.Client
	}{
		{
			tc:                           "should return all data with no errors",
			tenantID:                     "some tenant",
			expectError:                  false,
			CertURL:                      "/certs",
			PrivateKeyURL:                "/private",
			IntermediateURLs:             []string{"/inter"},
			expectedCertResponse:         expectedResponse{Value: testCertFromTestCertBytes, Err: nil},
			expectedPrivateKeyResponse:   expectedResponse{Value: testPrivateKey, Err: nil},
			expectedIntermediateResponse: expectedResponse{Value: []*x509.Certificate{testCertFromTestCertBytes}, Err: nil},
			customClient: mockClient{
				Responses: []mockedResponses{
					{
						URL:      "/certs",
						Err:      nil,
						Response: testCertBytes,
					},
					{
						URL:      "/private",
						Err:      nil,
						Response: testPrivateKeyWithoutPWBytes,
					},
					{
						URL:      "/inter",
						Err:      nil,
						Response: testCertBytes,
					},
				},
			},
		},
		{
			tc:                           "should pass with multiple intermediates urls",
			tenantID:                     "some tenant",
			expectError:                  false,
			CertURL:                      "/certs",
			PrivateKeyURL:                "/private",
			IntermediateURLs:             []string{"/inter1", "/inter2", "/inter3"},
			expectedCertResponse:         expectedResponse{Value: testCertFromTestCertBytes, Err: nil},
			expectedPrivateKeyResponse:   expectedResponse{Value: testPrivateKey, Err: nil},
			expectedIntermediateResponse: expectedResponse{Value: []*x509.Certificate{testCertFromTestCertBytes, testCertFromTestCertBytes, testCertFromTestCertBytes}, Err: nil},
			customClient: mockClient{
				Responses: []mockedResponses{
					{
						URL:      "/certs",
						Err:      nil,
						Response: testCertBytes,
					},
					{
						URL:      "/private",
						Err:      nil,
						Response: testPrivateKeyWithoutPWBytes,
					},
					{
						URL:      "/inter1",
						Err:      nil,
						Response: testCertBytes,
					},
					{
						URL:      "/inter2",
						Err:      nil,
						Response: testCertBytes,
					},
					{
						URL:      "/inter3",
						Err:      nil,
						Response: testCertBytes,
					},
				},
			},
		},
		{
			tc:                           "should pass but intermediates should fail due to error",
			tenantID:                     "some tenant",
			expectError:                  false,
			CertURL:                      "/certs",
			PrivateKeyURL:                "/private",
			IntermediateURLs:             []string{"/inter1", "/inter2", "/inter3"},
			expectedCertResponse:         expectedResponse{Value: testCertFromTestCertBytes, Err: nil},
			expectedPrivateKeyResponse:   expectedResponse{Value: testPrivateKey, Err: nil},
			expectedIntermediateResponse: expectedResponse{Value: nilIntemediates, Err: testErr},
			customClient: mockClient{
				Responses: []mockedResponses{
					{
						URL:      "/certs",
						Err:      nil,
						Response: testCertBytes,
					},
					{
						URL:      "/private",
						Err:      nil,
						Response: testPrivateKeyWithoutPWBytes,
					},
					{
						URL:      "/inter1",
						Err:      nil,
						Response: testCertBytes,
					},
					{
						URL:      "/inter2",
						Err:      nil,
						Response: testCertBytes,
					},
					{
						URL:      "/inter3",
						Err:      testErr,
						Response: nil,
					},
				},
			},
		},
		{
			tc:                           "should fail on due to random data as response",
			tenantID:                     "some tenant",
			expectError:                  true,
			CertURL:                      "/certs",
			PrivateKeyURL:                "/private",
			IntermediateURLs:             []string{"/inter"},
			expectedCertResponse:         expectedResponse{Value: nilCert, Err: ErrInvalidCertPEM},
			expectedPrivateKeyResponse:   expectedResponse{Value: nilKey, Err: ErrInvalidPrivateKeyPEM},
			expectedIntermediateResponse: expectedResponse{Value: nilIntemediates, Err: ErrInvalidCertPEM},
			customClient: mockClient{
				Responses: []mockedResponses{
					{
						URL:      "/certs",
						Err:      nil,
						Response: randBytes,
					},
					{
						URL:      "/private",
						Err:      nil,
						Response: randBytes,
					},
					{
						URL:      "/inter",
						Err:      nil,
						Response: randBytes,
					},
				},
			},
		},
		{
			tc:                           "should fail on  all data with errors",
			tenantID:                     "some tenant",
			expectError:                  true,
			CertURL:                      "/certs",
			PrivateKeyURL:                "/private",
			IntermediateURLs:             []string{"/inter"},
			expectedCertResponse:         expectedResponse{Value: nilCert, Err: testErr},
			expectedPrivateKeyResponse:   expectedResponse{Value: nilKey, Err: testErr},
			expectedIntermediateResponse: expectedResponse{Value: nilIntemediates, Err: testErr},
			customClient: mockClient{
				Responses: []mockedResponses{
					{
						URL:      "/certs",
						Err:      testErr,
						Response: nil,
					},
					{
						URL:      "/private",
						Err:      testErr,
						Response: nil,
					},
					{
						URL:      "/inter",
						Err:      testErr,
						Response: nil,
					},
				},
			},
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			s := URLX509CertificateSupplier{
				client:           test.customClient,
				tenantID:         test.tenantID,
				certificateURL:   test.CertURL,
				intermediateURLs: test.IntermediateURLs,
				privateKeyURL:    test.PrivateKeyURL,
			}
			p, pe := s.PrivateKeyOrError()
			assert.Equal(t, test.expectedPrivateKeyResponse.Value, p)
			assert.Equal(t, test.expectedPrivateKeyResponse.Err, pe)
			assert.Equal(t, p, s.PrivateKey())

			c, ce := s.CertificateOrError()
			assert.Equal(t, test.expectedCertResponse.Value, c)
			assert.Equal(t, test.expectedCertResponse.Err, ce)
			assert.Equal(t, c, s.Certificate())
			if test.expectedCertResponse.Err != nil {
				kid, errID := s.KeyID()
				key, err := s.Key(kid)
				assert.NotNil(t, errID)
				assert.NotNil(t, err)
				assert.Nil(t, key)
			}

			i, ie := s.IntermediateOrError()
			assert.Equal(t, test.expectedIntermediateResponse.Value, i)
			assert.Equal(t, test.expectedIntermediateResponse.Err, ie)
			assert.Equal(t, i, s.Intermediate())

			if test.expectedCertResponse.Err == nil && test.expectedIntermediateResponse.Err == nil &&
				test.expectedPrivateKeyResponse.Err == nil {
				kid, _ := s.KeyID()
				key, _ := s.Key(kid)
				assert.NotNil(t, key)
			}

		})
	}
}

// Used in rotation tests
var (
	testSecondPrivateKey, testSecondPublicKey = GenerateRsaKeyPair(2048)
	testSecondPrivateKeyBytes                 = x509.MarshalPKCS1PrivateKey(testSecondPrivateKey)
	testSecondPrivateKeyBlock                 = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: testSecondPrivateKeyBytes,
	}
	testSecondPrivateKeyWithoutPWBytes = pem.EncodeToMemory(testSecondPrivateKeyBlock)
	testSecondCertDER, _               = x509.CreateCertificate(rand.Reader, &testCertTemplate, &testCertTemplate, testSecondPublicKey, testSecondPrivateKey)
	testSecondCert, _                  = x509.ParseCertificate(testSecondCertDER)
	testSecondCertBytes                = pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: testSecondCert.Raw,
		},
	)
	testSecondCertFromTestCertBytes, _ = parseCertificateByteArray(testSecondCertBytes)
)

func TestURLX509CertificateSupplierKeyRotation(t *testing.T) {
	randBytes := make([]byte, 100)
	rand.Read(randBytes)

	testIO := []struct {
		tc               string
		tenantID         string
		CertURL          string
		PrivateKeyURL    string
		IntermediateURLs []string
		// mocked client for the first read of certs
		customClient httpsigner.Client
		// mocked client for the read of rotated certs
		customRotatedClient httpsigner.Client
		// whether to expect a rotation error
		expectRotationError bool
	}{
		{
			tc:               "should not error if no key rotation occurred",
			tenantID:         "some tenant",
			CertURL:          "/certs",
			PrivateKeyURL:    "/private",
			IntermediateURLs: []string{"/inter"},
			customClient: mockClient{
				Responses: []mockedResponses{
					{
						URL:      "/certs",
						Err:      nil,
						Response: testCertBytes,
					},
					{
						URL:      "/private",
						Err:      nil,
						Response: testPrivateKeyWithoutPWBytes,
					},
					{
						URL:      "/inter",
						Err:      nil,
						Response: testCertBytes,
					},
				},
			},
			customRotatedClient: mockClient{
				Responses: []mockedResponses{
					{
						URL:      "/certs",
						Err:      nil,
						Response: testCertBytes,
					},
					{
						URL:      "/private",
						Err:      nil,
						Response: testPrivateKeyWithoutPWBytes,
					},
					{
						URL:      "/inter",
						Err:      nil,
						Response: testCertBytes,
					},
				},
			},
			expectRotationError: false,
		},
		{
			tc:               "should produce the appropriate error if key rotation occurred",
			tenantID:         "some tenant",
			CertURL:          "/certs",
			PrivateKeyURL:    "/private",
			IntermediateURLs: []string{"/inter"},
			customClient: mockClient{
				Responses: []mockedResponses{
					{
						URL:      "/certs",
						Err:      nil,
						Response: testCertBytes,
					},
					{
						URL:      "/private",
						Err:      nil,
						Response: testPrivateKeyWithoutPWBytes,
					},
					{
						URL:      "/inter",
						Err:      nil,
						Response: testCertBytes,
					},
				},
			},
			customRotatedClient: mockClient{
				Responses: []mockedResponses{
					{
						URL:      "/certs",
						Err:      nil,
						Response: testSecondCertBytes,
					},
					{
						URL:      "/private",
						Err:      nil,
						Response: testSecondPrivateKeyWithoutPWBytes,
					},
					{
						URL:      "/inter",
						Err:      nil,
						Response: testSecondCertBytes,
					},
				},
			},
			expectRotationError: true,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			s := URLX509CertificateSupplier{
				client:           test.customClient,
				tenantID:         test.tenantID,
				certificateURL:   test.CertURL,
				intermediateURLs: test.IntermediateURLs,
				privateKeyURL:    test.PrivateKeyURL,
			}

			// We get the ID of the first key.
			firstKeyID, errID := s.KeyID()
			assert.Nil(t, errID)

			// Then we modify our supplier so that we use the rotated certs.
			s.client = test.customRotatedClient

			// And we check the properties of the keys returned and the rotation
			// error, if such an error is expected.
			secondKeyID, errID := s.KeyID()
			assert.Nil(t, errID)
			if test.expectRotationError {
				// The key has been rotated, we expect a different key and a
				// rotation error.
				assert.NotEqual(t, firstKeyID, secondKeyID)
				key, err := s.Key(firstKeyID)
				assert.Nil(t, key)
				assert.Equal(t, httpsigner.NewKeyRotationError(secondKeyID, firstKeyID), err)
			} else {
				// No rotation occurred, we expect the key to be the same and
				// no error.
				assert.Equal(t, firstKeyID, secondKeyID)
				key, err := s.Key(firstKeyID)
				assert.NotNil(t, key)
				assert.Nil(t, err)
			}
		})
	}

}
