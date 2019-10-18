// Copyright (c) 2017-2019, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"crypto/rand"
	"net/http/httptest"

	"oracle.com/oci/httpsigner"
)

// TestNewSTSToken tests that the NewSTSToken constructor can return a valid token struct
func TestNewSTSToken(t *testing.T) {
	token, _ := NewSTSToken(validSignedToken, testKeyService)

	var tokenType *STSToken
	assert.IsType(t, tokenType, token)
}

// TestNewTokenEmptyString tests that the NewSTSToken constructor won't accept an empty string as
// a valid token
func TestNewSTSTokenEmptyString(t *testing.T) {
	token, err := NewSTSToken("", testKeyService)
	assert.Nil(t, token)
	assert.NotNil(t, err)
	assert.Equal(t, ErrInvalidArg, err)
}

// TestTokenAsString tests that the String() function of token returns the token string
func TestSTSTokenAsString(t *testing.T) {
	token, err := NewSTSToken(validSignedToken, testKeyService)

	assert.Nil(t, err)
	assert.Equal(t, "ST$"+validSignedToken, token.String())
}

func TestNewCustomSTSKeySupplier(t *testing.T) {
	tm := []struct {
		tc        string
		transport *http.Transport
	}{
		{
			tc:        "custom transport should be found in the generated suppliers client options",
			transport: &http.Transport{},
		},
	}
	var supplier CertificateSupplier
	for _, test := range tm {
		t.Run(fmt.Sprintf(test.tc), func(t *testing.T) {
			sks, err := NewCustomSTSKeySupplier(
				supplier,
				testX509Client,
				endpoint,
				&ClientOptions{
					HTTPClient: &http.Client{
						Transport: test.transport,
					},
				},
			)
			assert.Nil(t, err)
			assert.Equal(t, sks.validationClientOptions.HTTPClient.Transport, test.transport)

		})
	}
}

// Test that we can create an STSKeySupplier
func TestNewSTSKeySupplier(t *testing.T) {
	testIO := []struct {
		tc            string
		supplier      CertificateSupplier
		client        httpsigner.Client
		endpoint      string
		expectedError error
	}{
		{tc: "should return invalid x509Supplier error with nil supplier",
			supplier: nil, client: testX509Client, endpoint: endpoint, expectedError: ErrInvalidX509Supplier},
		{tc: "should return invalid client error with nil client",
			supplier: testX509CertificateSupplier, client: nil, endpoint: endpoint, expectedError: ErrInvalidClient},
		{tc: "should return invalid endpoint error with empty endpoint",
			supplier: testX509CertificateSupplier, client: testX509Client, endpoint: "", expectedError: ErrInvalidEndpoint},
		{tc: "should create a key supplier with valid supplier and client",
			supplier: testX509CertificateSupplier, client: testX509Client, endpoint: endpoint, expectedError: nil},
	}

	for _, test := range testIO {
		t.Run(fmt.Sprintf(test.tc), func(t *testing.T) {
			s, e := NewSTSKeySupplier(test.supplier, test.client, test.endpoint)
			if test.expectedError == nil {
				assert.Nil(t, e)
				assert.NotNil(t, s)
				assert.Equal(t, s.tokenPurpose, "")
			} else {
				assert.Equal(t, e, test.expectedError)
				assert.Nil(t, s)
			}
		})
	}
}

var (
	goodToken, _    = NewSTSToken(GenerateTokenForTest(time.Now().Add(time.Hour*time.Duration(12))), testKeyService)
	expiredToken, _ = NewSTSToken(GenerateTokenForTest(time.Now()), testKeyService)
)

func newValidServerResponse() *http.Response {
	return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewBufferString(fmt.Sprintf(`{"token": "%s"}`, goodToken.rawToken)))}
}

func newServerResponse(output string) *http.Response {
	return &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(fmt.Sprintf(output)))}
}

func TestSTSSupplierKey(t *testing.T) {
	serverError := errors.New("Some error")
	keyID := "somekeyid"
	jwk := NewJWKFromPublicKey(keyID, testPublicKey)
	non200Response := &http.Response{StatusCode: http.StatusTeapot}

	// Setup httptest server to return the key
	ts := SetupHTTPTestServerToReturnJWK(jwk)
	defer ts.Close()

	testIO := []struct {
		tc            string
		token         *STSToken
		previousToken *STSToken
		keyid         string
		expected      *rsa.PrivateKey
		expectedError error
		client        httpsigner.Client
	}{
		{
			tc:            `should return nil and error if given keyid doesn't match current or previous token`,
			token:         goodToken,
			previousToken: nil,
			keyid:         "invalid-keyid-input",
			client:        testX509Client,
			expectedError: httpsigner.ErrKeyNotFound,
		},
		{
			tc:            `should return nil and error if there is an error fetching a new token`,
			token:         expiredToken,
			previousToken: nil,
			keyid:         expiredToken.String(),
			client:        &MockSigningClient{doError: serverError},
			expectedError: serverError,
		},
		{
			tc:            `should return ServiceResponseError if the response from STS is a non-2xx response`,
			token:         expiredToken,
			previousToken: nil,
			keyid:         expiredToken.String(),
			client:        &MockSigningClient{doResponse: non200Response},
			expectedError: &ServiceResponseError{non200Response},
		},
		{
			tc:            `should return nil and key expired error which contains a new token when key is expired`,
			token:         expiredToken,
			previousToken: nil,
			keyid:         expiredToken.String(),
			client:        &MockSigningClient{doResponse: newValidServerResponse()},
			expectedError: &httpsigner.KeyRotationError{},
		},
		{
			tc:            `should return KeyRotationError with current keyid for previous keyid without calling the key service`,
			token:         goodToken,
			previousToken: expiredToken,
			keyid:         expiredToken.String(),
			client:        &MockSigningClient{doError: serverError},
			expectedError: &httpsigner.KeyRotationError{},
		},
		{
			tc:            `should return KeyRotationError if KeyIDForceRotate is supplied as the keyid`,
			token:         goodToken,
			previousToken: nil,
			keyid:         KeyIDForceRotate,
			client:        &MockSigningClient{doResponse: newValidServerResponse()},
			expectedError: &httpsigner.KeyRotationError{},
		},
		{
			tc:            `should return the temporary private key and no error for valid token`,
			token:         goodToken,
			previousToken: nil,
			keyid:         goodToken.String(),
			client:        &MockSigningClient{doResponse: newValidServerResponse()},
			expectedError: nil,
		},
	}

	for _, test := range testIO {
		t.Run(fmt.Sprintf(test.tc), func(t *testing.T) {
			sts, _ := NewSTSKeySupplier(testX509CertificateSupplier, testX509Client, ts.URL)
			sts.client = test.client
			sts.token = test.token
			sts.previousToken = test.previousToken

			key, err := sts.Key(test.keyid)

			if test.expectedError != nil {
				assert.Nil(t, key)
				if _, ok := test.expectedError.(*httpsigner.KeyRotationError); ok {
					expired, ok := err.(*httpsigner.KeyRotationError)
					if ok {
						// Newly rotated
						if test.previousToken == nil {
							assert.Equal(t, expired.OldKeyID, test.token.String())
							assert.Equal(t, sts.previousToken.rawToken, test.token.rawToken)
						} else {
							// Saved previous token was returned
							assert.Equal(t, expired.OldKeyID, test.previousToken.String())
							assert.Equal(t, sts.previousToken.rawToken, test.previousToken.rawToken)
						}

						assert.Equal(t, expired.ReplacementKeyID, goodToken.String())
						assert.Equal(t, sts.token.rawToken, goodToken.rawToken)
					} else {
						assert.Equal(t, test.expectedError, err)
					}
				} else {
					assert.Equal(t, test.expectedError, err)
				}
			} else {
				assert.NotNil(t, key)
				assert.Nil(t, err)
			}

		})
	}
}

func TestSTSSupplierKeyForceRotate(t *testing.T) {
	testIO := []struct {
		tc            string
		token         *STSToken
		previousToken *STSToken
	}{
		{
			tc:            `should return KeyRotationError if KeyIDForceRotate is supplied as the keyid and sts supplier still has a valid token`,
			token:         goodToken,
			previousToken: nil,
		},
		{
			tc:            `should return KeyRotationError if KeyIDForceRotate is supplied as the keyid and sts supplier has an expired tokens`,
			token:         expiredToken,
			previousToken: nil,
		},
		{
			tc:            `should return KeyRotationError if KeyIDForceRotate is supplied as the keyid and sts supplier has no tokens`,
			token:         nil,
			previousToken: nil,
		},
	}

	for _, test := range testIO {
		t.Run(fmt.Sprintf(test.tc), func(t *testing.T) {
			sts, _ := NewSTSKeySupplier(testX509CertificateSupplier, testX509Client, endpoint)
			sts.client = &MockSigningClient{doResponse: newValidServerResponse()}
			sts.token = test.token
			sts.previousToken = test.previousToken

			key, err := sts.Key(KeyIDForceRotate)
			assert.Nil(t, key)

			expired, ok := err.(*httpsigner.KeyRotationError)
			assert.True(t, ok)

			assert.Equal(t, goodToken.String(), expired.ReplacementKeyID)
			if test.token == nil {
				assert.Equal(t, "", expired.OldKeyID)
			} else {
				assert.Equal(t, test.token.String(), expired.OldKeyID)
			}
		})
	}
}

func TestSTSKeyID(t *testing.T) {
	stsKeyIDError := errors.New("some-error")
	testIO := []struct {
		tc            string
		token         *STSToken
		client        httpsigner.Client
		expectedError error
	}{
		{tc: `should return token prepended by ST$ for valid token`,
			token: goodToken, client: &MockSigningClient{}, expectedError: nil},
		{tc: `should return KEY_ID_FORCE_ROTATE and an error if error is returned from the key service`,
			token: expiredToken, client: &MockSigningClient{doError: stsKeyIDError}, expectedError: stsKeyIDError},
	}

	for _, test := range testIO {
		t.Run(fmt.Sprintf(test.tc), func(t *testing.T) {
			sts, _ := NewSTSKeySupplier(testX509CertificateSupplier, testX509Client, endpoint)
			sts.client = test.client
			sts.token = test.token

			s2s, err := sts.KeyID()
			if test.expectedError != nil {
				assert.Equal(t, s2s, KeyIDForceRotate)
				assert.Equal(t, err, test.expectedError)
			} else {
				assert.Equal(t, s2s, "ST$"+test.token.rawToken)
				assert.Nil(t, err)
			}
		})

	}
}

func TestSecurityTokenFromServerError(t *testing.T) {
	testErr := errors.New("test-error")
	invalidJSON := `{test: test`
	testIO := []struct {
		tc       string
		client   httpsigner.Client
		endpoint string
	}{
		{tc: `should handle error from http.NewRequest error`,
			client: &MockSigningClient{doResponse: newValidServerResponse()}, endpoint: "http://%en/"},
		{tc: `should handle error from ioutil.ReadAll`,
			client:   &MockSigningClient{doResponse: &http.Response{Body: mockReadCloser{readError: testErr}}},
			endpoint: endpoint},
		{tc: `should handle JSON unmarshal error`,
			client:   &MockSigningClient{doResponse: newServerResponse(invalidJSON)},
			endpoint: endpoint},
	}

	for _, test := range testIO {
		t.Run(fmt.Sprintf(test.tc), func(t *testing.T) {
			sts, _ := NewSTSKeySupplier(testX509CertificateSupplier, testX509Client, test.endpoint)
			sts.client = test.client
			token, e := sts.SecurityTokenFromServer()
			assert.Nil(t, token)
			assert.NotNil(t, e)
		})
	}

}
func TestSecurityToken(t *testing.T) {
	keyID := "somekeyid"
	jwk := NewJWKFromPublicKey(keyID, testPublicKey)

	// Setup httptest server to return the key
	ts := SetupHTTPTestServerToReturnJWK(jwk)
	defer ts.Close()

	testIO := []struct {
		tc              string
		cachedToken     *STSToken
		client          httpsigner.Client
		errorExpected   bool
		refreshExpected bool
	}{
		{tc: `should return blank string and an error if error is returned from the key service`,
			cachedToken: expiredToken, client: &MockSigningClient{doError: errors.New("some error")},
			errorExpected: true, refreshExpected: false},
		{tc: `should return the same token if it's still valid`,
			cachedToken: goodToken, client: &MockSigningClient{}, errorExpected: false, refreshExpected: false},
		{tc: `should return new token if it's successfully able to retrieve a new one from the key service`,
			cachedToken: expiredToken, client: &MockSigningClient{doResponse: newValidServerResponse()}, errorExpected: false, refreshExpected: true},
		{tc: `should cache the token if it's valid`,
			cachedToken: nil, client: &MockSigningClient{doResponse: newValidServerResponse()}, errorExpected: false, refreshExpected: true},
	}

	for _, test := range testIO {
		t.Run(fmt.Sprintf(test.tc), func(t *testing.T) {
			sts, _ := NewSTSKeySupplier(testX509CertificateSupplier, testX509Client, ts.URL)
			sts.token = test.cachedToken
			sts.client = test.client
			oldKey := sts.sessionKeySupplier.PrivateKey()

			token, err := sts.updateSecurityToken(false)
			if test.errorExpected {
				assert.NotNil(t, err)
				assert.Nil(t, token)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, token)
				assert.NotNil(t, sts.token)
				if test.refreshExpected {
					assert.NotEqual(t, oldKey, sts.sessionKeySupplier.PrivateKey())
				} else {
					assert.Equal(t, oldKey, sts.sessionKeySupplier.PrivateKey())
				}
			}
		})
	}
}

func TestIsSecurityTokenValid(t *testing.T) {
	// This token expires in now + expirationTimePadding
	expiredPaddedToken, _ := NewSTSToken(GenerateTokenForTest(time.Now().Add(expirationTimePadding)), testKeyService)

	testIO := []struct {
		tc       string
		token    *STSToken
		expected bool
	}{
		{tc: `should return true if the current token is still valid`,
			token: goodToken, expected: true},
		{tc: `should return false if the current token is expired`,
			token: expiredToken, expected: false},
		{tc: `should return false if the current token is valid, but considered expired due to the time padding`,
			token: expiredPaddedToken, expected: false},
		{tc: `should return false if the current token has no field`,
			token: &STSToken{}, expected: false},
		{tc: `should return false if the current token is nil`,
			token: nil, expected: false},
	}
	for _, test := range testIO {
		t.Run(fmt.Sprintf(test.tc), func(t *testing.T) {
			sts, _ := NewSTSKeySupplier(testX509CertificateSupplier, testX509Client, endpoint)
			sts.token = test.token
			result := sts.IsSecurityTokenValid()
			assert.Equal(t, test.expected, result)
		})
	}
}

type TimedMockSigningClient struct {
	signRequestResponse *http.Request
	signRequestError    error

	doResponse *http.Response
	doError    error

	waitDuration time.Duration
}

func (m *TimedMockSigningClient) SignRequest(request *http.Request) (*http.Request, error) {
	time.Sleep(m.waitDuration)
	return m.signRequestResponse, m.signRequestError
}

func (m *TimedMockSigningClient) Do(request *http.Request) (*http.Response, error) {
	time.Sleep(m.waitDuration)
	return m.doResponse, m.doError
}

func TestMutexKeyRotation(t *testing.T) {
	// setup
	waitDuration := time.Duration(1) * time.Second
	sts, err := NewSTSKeySupplier(testX509CertificateSupplier, testX509Client, endpoint)
	sts.client = &TimedMockSigningClient{waitDuration: waitDuration, doResponse: newValidServerResponse()}
	sts.token = expiredToken
	assert.Nil(t, err)

	// Attempt to fetch the Key - this will lock the keyRotationMutex
	go sts.Key(expiredToken.rawToken)

	// This second goroutine should not be executed until the mutex is unlocked and the key has been rotated
	// Note that this test will fail in an unlikely scenario this go routine takes more than the value of waitDuration to start
	go func() { keyid, _ := sts.KeyID(); assert.Equal(t, keyid, "ST$"+goodToken.rawToken) }()
}

func TestValidationKeySupplier(t *testing.T) {
	testIO := []struct {
		tc string
	}{
		{tc: `should return a new key service`},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			sts, err := NewSTSKeySupplier(testX509CertificateSupplier, testX509Client, endpoint)
			assert.Nil(t, err)
			keyService, err := sts.validationKeySupplier("token")

			assert.Nil(t, err)
			assert.NotNil(t, keyService)
		})
	}
}

func TestBuildS2SRequestIntermediateCert(t *testing.T) {
	testX509CertificateSupplierEmptyIntermediate, _ := NewX509CertificateSupplier(tenantID, testCert, []*x509.Certificate{}, testPrivateKey)
	testX509CertificateSupplierNilIntermediate, _ := NewX509CertificateSupplier(tenantID, testCert, nil, testPrivateKey)
	testX509CertificateSupplierMultipleIntermediate, _ := NewX509CertificateSupplier(tenantID, testCert, []*x509.Certificate{testCert, testCert, testCert}, testPrivateKey)

	testIO := []struct {
		tc                       string
		x509Supplier             CertificateSupplier
		intermediateFieldPresent bool
	}{
		{tc: `should contain intermediate certificates field in the marshaled json`,
			x509Supplier: testX509CertificateSupplier, intermediateFieldPresent: true},
		{tc: `should contain multiple intermediate certificates field in the marshaled json`,
			x509Supplier: testX509CertificateSupplierMultipleIntermediate, intermediateFieldPresent: true},
		{tc: `should not contain intermediate certificates field in the marshaled json if it's empty`,
			x509Supplier: testX509CertificateSupplierEmptyIntermediate, intermediateFieldPresent: false},
		{tc: `should not contain intermediate certificates field in the marshaled json if it's nil`,
			x509Supplier: testX509CertificateSupplierNilIntermediate, intermediateFieldPresent: false},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			sts, err := NewSTSKeySupplier(test.x509Supplier, testX509Client, endpoint)
			assert.Nil(t, err)

			s2sRequest, err := sts.buildS2SRequest()
			assert.Nil(t, err)

			marshaled, err := json.Marshal(s2sRequest)
			assert.Nil(t, err)

			if test.intermediateFieldPresent {
				assert.NotNil(t, s2sRequest.IntermediateCertificates)
				assert.Equal(t, len(s2sRequest.IntermediateCertificates), len(test.x509Supplier.Intermediate()))
				assert.True(t, strings.Contains(string(marshaled), `intermediateCertificates`))
			} else {
				assert.Nil(t, s2sRequest.IntermediateCertificates)
				assert.False(t, strings.Contains(string(marshaled), `intermediateCertificates`))
			}
		})
	}
}

type serverResults struct {
	Results map[string][]byte
}

func newMetadataService(results **serverResults) *httptest.Server {
	var mockMetadaService = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if value, ok := (*results).Results[r.URL.Path]; ok {
			fmt.Fprint(w, string(value))
		}
	}))
	return mockMetadaService
}
func TestBuildS2SRequestWithURLCertSupplier(t *testing.T) {
	var res *serverResults
	service := newMetadataService(&res)
	service.Start()
	defer service.Close()

	randBytes := make([]byte, 100)
	rand.Read(randBytes)

	createWithoutError := func(t, cert, private string, pass []byte, inter ...string) *URLX509CertificateSupplier {
		urls := make([]string, len(inter))
		for i, u := range inter {
			urls[i] = service.URL + u
		}
		supplier, _ := NewX509CertificateSupplierFromURLs(http.DefaultClient,
			tenantID,
			service.URL+cert,
			service.URL+private,
			nil,
			urls...,
		)
		return supplier
	}

	testIO := []struct {
		tc                       string
		expectError              error
		serviceResponses         map[string][]byte
		x509Supplier             CertificateSupplier
		intermediateFieldPresent bool
	}{
		{
			tc:                       `should contain intermediate certificates field in the marshaled json with url certificate supplier`,
			serviceResponses:         map[string][]byte{"/cert.pem": testCertBytes, "/private.pem": testPrivateKeyWithoutPWBytes, "/intermediate.pem": testCertBytes},
			x509Supplier:             createWithoutError(tenantID, "/cert.pem", "/private.pem", nil, "/intermediate.pem"),
			intermediateFieldPresent: true,
		},
		{
			tc:                       `should contain multiple intermediate certificates field in the marshaled json with url certificate supplier`,
			serviceResponses:         map[string][]byte{"/cert.pem": testCertBytes, "/private.pem": testPrivateKeyWithoutPWBytes, "/i1.pem": testCertBytes, "/i2.pem": testCertBytes},
			x509Supplier:             createWithoutError(tenantID, "/cert.pem", "/private.pem", nil, "/i1.pem", "/i2.pem"),
			intermediateFieldPresent: true,
		},
		{
			tc:                       `should fail with PEM error contain intermediate certificates field in the marshaled json with url certificate supplier`,
			expectError:              errors.New("certificate PEM data is invalid"),
			serviceResponses:         map[string][]byte{"/cert.pem": testCertBytes, "/private.pem": testPrivateKeyWithoutPWBytes, "/intermediate.pem": randBytes},
			x509Supplier:             createWithoutError(tenantID, "/cert.pem", "/private.pem", nil, "/intermediate.pem"),
			intermediateFieldPresent: false,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			res = &serverResults{Results: test.serviceResponses}
			sts, err := NewSTSKeySupplier(test.x509Supplier, testX509Client, endpoint)
			assert.Nil(t, err)

			s2sRequest, err := sts.buildS2SRequest()
			assert.Equal(t, test.expectError, err)
			if err != nil {
				return
			}

			marshaled, err := json.Marshal(s2sRequest)
			assert.Nil(t, err)

			if test.intermediateFieldPresent {
				assert.NotNil(t, s2sRequest.IntermediateCertificates)
				assert.Equal(t, len(s2sRequest.IntermediateCertificates), len(test.x509Supplier.Intermediate()))
				assert.True(t, strings.Contains(string(marshaled), `intermediateCertificates`))
			} else {
				assert.Nil(t, s2sRequest.IntermediateCertificates)
				assert.False(t, strings.Contains(string(marshaled), `intermediateCertificates`))
			}
		})
	}
}

// Test that we can create an NewInstancePrincipalKeySupplier
func TestNewInstanceKeySupplier(t *testing.T) {
	testIO := []struct {
		tc              string
		tenancyID       string
		expectedError   error
		mockedResponses []mockedResponses
	}{
		{
			tc: "should create a new supplier with valid data",
			mockedResponses: []mockedResponses{
				{
					URL:      "/opc/v1/identity/cert.pem",
					Err:      nil,
					Response: testCertBytes,
				},
				{
					URL:      "/opc/v1/identity/key.pem",
					Err:      nil,
					Response: testPrivateKeyWithoutPWBytes,
				},
				{
					URL:      "/opc/v1/identity/intermediate.pem",
					Err:      nil,
					Response: testCertBytes,
				},
				{
					URL:      "/opc/v1/instance/region",
					Err:      nil,
					Response: []byte("someregion"),
				},
			},
			tenancyID: tenantID, expectedError: nil},
	}

	for _, test := range testIO {
		t.Run(fmt.Sprintf(test.tc), func(t *testing.T) {
			client := mockClient{Responses: test.mockedResponses}
			s, e := NewInstanceKeySupplier(test.tenancyID, "someendpoint", client)
			if test.expectedError == nil {
				assert.Nil(t, e)
				assert.NotNil(t, s)
				assert.Equal(t, s.tokenPurpose, "")
			} else {
				assert.Equal(t, e, test.expectedError)
				assert.Nil(t, s)
			}
		})
	}
}

// Test that we can create a NewServiceInstanceKeySupplier
func TestNewServiceInstanceKeySupplier(t *testing.T) {
	testIO := []struct {
		tc              string
		tenancyID       string
		expectedError   error
		mockedResponses []mockedResponses
	}{
		{
			tc: "should create a new supplier with valid data",
			mockedResponses: []mockedResponses{
				{
					URL:      "/opc/v1/identity/cert.pem",
					Err:      nil,
					Response: testCertBytes,
				},
				{
					URL:      "/opc/v1/identity/key.pem",
					Err:      nil,
					Response: testPrivateKeyWithoutPWBytes,
				},
				{
					URL:      "/opc/v1/identity/intermediate.pem",
					Err:      nil,
					Response: testCertBytes,
				},
				{
					URL:      "/opc/v1/instance/region",
					Err:      nil,
					Response: []byte("someregion"),
				},
			},
			tenancyID:     tenantID,
			expectedError: nil,
		},
	}

	for _, test := range testIO {
		t.Run(fmt.Sprintf(test.tc), func(t *testing.T) {
			client := mockClient{Responses: test.mockedResponses}
			s, e := NewServiceInstanceKeySupplier(test.tenancyID, "someendpoint", client)
			if test.expectedError == nil {
				assert.Nil(t, e)
				assert.NotNil(t, s)
				assert.Equal(t, s.tokenPurpose, servicePrincipalSTSPurpose)
			} else {
				assert.Equal(t, e, test.expectedError)
				assert.Nil(t, s)
			}
		})
	}
}
