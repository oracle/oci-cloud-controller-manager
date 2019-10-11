// Copyright (c) 2018, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"bytes"
	"crypto/rsa"
	"encoding/asn1"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"oracle.com/oci/httpsigner"

	"github.com/stretchr/testify/assert"
)

var (
	testCachePeriod = time.Minute * keyServiceDefaultCachePeriodMinutes

	pkixPublicKey        = &rsa.PublicKey{N: big.NewInt(3722417341), E: 65537}
	encodedPKIXPublicKey = []byte("-----BEGIN PUBLIC KEY-----\n" +
		"MCAwDQYJKoZIhvcNAQEBBQADDwAwDAIFAN3flL0CAwEAAQ==\n" +
		"-----END PUBLIC KEY-----")
	encodedBadPublicKey = []byte("-----BEGIN PUBLIC KEY-----\n" +
		"-----END PUBLIC KEY-----")
	encodedSecret = []byte("-----BEGIN SECRET-----" +
		"U2hoaGhoaGhoIQ==" +
		"-----END SECRET-----")
)

// TestNewKeyServiceKeySupplier tests that NewKeyServiceKeySupplier returns a new instance of
// the key supplier
func TestNewKeyServiceKeySupplier(t *testing.T) {
	k := NewKeyServiceKeySupplier(testX509Client, endpoint)

	var keyServiceType *KeyServiceKeySupplier
	assert.IsType(t, keyServiceType, k)
	assert.Equal(t, fmt.Sprintf("%s/keys", endpoint), k.uri)
	assert.Equal(t, time.Minute*180, k.cachePeriod)
}

// TestNewAPIKeyServiceKeySupplier tests the NewAPIKeyServiceKeySupplier returns a new instanceof the key supplier with
// the correct URI and timeout
func TestNewAPIKeyServiceKeySupplier(t *testing.T) {
	k := NewAPIKeyServiceKeySupplier(testX509Client, endpoint)

	var keyServiceType *KeyServiceKeySupplier
	assert.IsType(t, keyServiceType, k)
	assert.Equal(t, fmt.Sprintf("%s/SR/keys", endpoint), k.uri)
	assert.Equal(t, time.Minute*2, k.cachePeriod)
}

// TestKeyHappyPath tests a successful key request
func TestKeyHappyPath(t *testing.T) {
	keyID := "aswtestkey"
	jwk := NewJWKFromPublicKey(keyID, testPublicKey)

	// Setup httptest server to return the key
	ts := SetupHTTPTestServerToReturnJWK(jwk)
	defer ts.Close()

	ks := NewKeyServiceKeySupplier(testX509Client, ts.URL)

	key, err := ks.Key(keyID)

	assert.Nil(t, err)
	assert.Equal(t, testPublicKey, key)
}

// TestKeyHappyPathPKIX tests a successful key request for a customer Key
func TestKeyHappyPathPKIX(t *testing.T) {
	keyID := "pkixkey"

	// Setup httptest server to return the key
	ts := SetupHTTPTestServer(string(encodedPKIXPublicKey))
	defer ts.Close()

	ks := NewAPIKeyServiceKeySupplier(testX509Client, ts.URL)

	key, err := ks.Key(keyID)

	assert.Nil(t, err)
	assert.Equal(t, pkixPublicKey, key)
}

// TestKeyCached tests that an already cached key is returned by the Key Service without making
// a request
func TestKeyCached(t *testing.T) {
	ks := NewKeyServiceKeySupplier(testX509Client, endpoint)

	keyID := "aswxyz"
	cacheKeyID := "SYSTEMKEYaswxyz"

	// Seed the cache with the public key
	ks.keyCache.Store(cacheKeyID, testPublicKey, testCachePeriod)

	// Try and retrieve the key
	key, err := ks.Key(keyID)

	assert.Nil(t, err)
	assert.Equal(t, testPublicKey, key)
}

// TestKeyNotFound tests that the key service handles a 404 for non-existent keys
func TestKeyNotFound(t *testing.T) {
	// Setup handler to respond to the http request
	echoHandler := func(w http.ResponseWriter, r *http.Request) {
		// Set response status code
		w.WriteHeader(http.StatusNotFound)

		// Set response body
		fmt.Fprint(w, "{\"code\": \"NotFound\", \"message\": \"Not Found\"}")
	}

	// Setup httptest server
	ts := httptest.NewServer(http.HandlerFunc(echoHandler))
	defer ts.Close()

	ks := NewKeyServiceKeySupplier(testX509Client, ts.URL)

	keyID := "aswidontexist"

	// Try and retrieve the key
	key, err := ks.Key(keyID)

	assert.Nil(t, key)
	assert.Equal(t, ErrKeyNotFound, err)
}

// SEC-2456: This test verifies that System and APIKey key suppliers do not return each others keys even though they
// currently share a cache
//
// KeyServiceKeySupplier.Key() should handle keyID collisions between system and api keys
func TestKeyIDCollision(t *testing.T) {

	// identical key id
	keyID := "aswcollidingKeyID"

	// init system key and supplier
	jwk := NewJWKFromPublicKey(keyID, testPublicKey)
	tsSystemKey := SetupHTTPTestServerToReturnJWK(jwk)
	defer tsSystemKey.Close()
	systemKeySupplier := NewKeyServiceKeySupplier(testX509Client, tsSystemKey.URL)

	// init api key and supplier
	tsAPIKey := SetupHTTPTestServer(string(encodedPKIXPublicKey))
	defer tsAPIKey.Close()
	apiKeySupplier := NewAPIKeyServiceKeySupplier(testX509Client, tsAPIKey.URL)

	// fetch key same key ID from each supplier
	syskey, err := systemKeySupplier.Key(keyID)
	assert.Nil(t, err)
	assert.Equal(t, testPublicKey, syskey)

	apikey, err := apiKeySupplier.Key(keyID)
	assert.Nil(t, err)
	assert.Equal(t, pkixPublicKey, apikey)

	// ensure that they are not the same key
	assert.NotEqual(t, syskey, apikey)
}

// SEC-2456:
// KeyServiceKeySupplier.Key() should return an error for an invalid system key even if a key with the requested id is
// already in the cache as an API Key.
func TestKeyCacheMissOnDifferentPrefix(t *testing.T) {

	// identical key id
	keyID := "tenancy/user/fingerprint"

	// init system key and supplier
	tsSystemKey := SetupHTTPNotFoundTestServer()
	defer tsSystemKey.Close()
	systemKeySupplier := NewKeyServiceKeySupplier(testX509Client, tsSystemKey.URL)

	// init api key and supplier
	tsAPIKey := SetupHTTPTestServer(string(encodedPKIXPublicKey))
	defer tsAPIKey.Close()
	apiKeySupplier := NewAPIKeyServiceKeySupplier(testX509Client, tsAPIKey.URL)

	// fetch api key
	apikey, err := apiKeySupplier.Key(keyID)
	assert.Nil(t, err)
	assert.Equal(t, pkixPublicKey, apikey)

	// attempt to fetch system key by the same id
	syskey, err := systemKeySupplier.Key(keyID)
	assert.Nil(t, syskey)
	assert.NotNil(t, err)
	assert.Equal(t, ErrUnsupportedKeyFormat, err)
}

// TestInvalidSystemKeyRelativePath verifies that keyID values with relative path components are rejected
func TestInvalidSystemKeyRelativePath(t *testing.T) {
	// relative path keyID
	keyID := "asw/../../SR/keys/tenancy/user/fingerprint"

	// init system key and supplier
	tsSystemKey := SetupHTTPNotFoundTestServer()
	defer tsSystemKey.Close()
	systemKeySupplier := NewKeyServiceKeySupplier(testX509Client, tsSystemKey.URL)

	// attempt to fetch system key by the same id
	syskey, err := systemKeySupplier.Key(keyID)
	assert.Nil(t, syskey)
	assert.NotNil(t, err)
	assert.Equal(t, ErrUnsupportedKeyFormat, err)
}

// Test error cases that Key function is expected to handle and return
func TestKeyErrors(t *testing.T) {
	testErr := errors.New("test-error")
	invalidJSON := []byte(`{test: test`)
	validJSON := []byte(`{"title": "ociauthz"}`)

	testIO := []struct {
		tc       string
		endpoint string
		keyid    string
		client   httpsigner.SigningClient
	}{
		// We use a bad URI to trigger error on http.NewRequest.  This is an invalid URI - since the URI is
		// Go does not allow %en which encodes to %25 (the encoded value corresponds to ampersand) in
		// the host section during URI parsing.
		{tc: `should handle NewRequest errors`,
			endpoint: "http://%en/", keyid: "test-key", client: nil},
		{tc: `should handle errors raised by SignngClient.Do`,
			endpoint: endpoint, keyid: "test-key", client: &MockSigningClient{doError: testErr}},
		{tc: `should handle unknown status code returned by the sts service`,
			endpoint: endpoint, keyid: "test-key", client: &MockSigningClient{doResponse: &http.Response{StatusCode: 418}}},
		{tc: `should handle ioutil ReadAll error`,
			endpoint: endpoint, keyid: "test-key", client: &MockSigningClient{
				doResponse: &http.Response{StatusCode: 200, Body: mockReadCloser{readError: testErr}}}},
		{tc: `should handle JWK unmarshall error`,
			endpoint: endpoint, keyid: "test-key", client: &MockSigningClient{
				doResponse: &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(invalidJSON))}}},
		{tc: `should handle JWK public key parse error`,
			endpoint: endpoint, keyid: "test-key", client: &MockSigningClient{
				doResponse: &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(validJSON))}}},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			// setup
			ks := NewKeyServiceKeySupplier(testX509Client, test.endpoint)
			if test.client != nil {
				ks.client = test.client
			}

			// Go
			r, e := ks.Key(test.keyid)
			assert.Nil(t, r)
			assert.NotNil(t, e)
		})
	}
}

func TestDecodePKIXPublicKeyError(t *testing.T) {
	testIO := []struct {
		tc            string
		encoded       []byte
		expectedKey   interface{}
		expectedError error
	}{
		{
			tc:            `should return a key for valid PKIX key`,
			encoded:       encodedPKIXPublicKey,
			expectedKey:   pkixPublicKey,
			expectedError: nil,
		},
		{
			tc:            `should ErrPEMDecodeError for empty string`,
			encoded:       []byte{},
			expectedKey:   nil,
			expectedError: ErrPEMDecodeError,
		},
		{
			tc:            `should ErrPEMDecodeError for non-public key`,
			encoded:       encodedSecret,
			expectedKey:   nil,
			expectedError: ErrPEMDecodeError,
		},
		{
			tc:            `should ErrPEMDecodeError for malformed public key`,
			encoded:       encodedBadPublicKey,
			expectedKey:   nil,
			expectedError: asn1.SyntaxError{Msg: `sequence truncated`},
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			key, err := decodePKIXPublicKey(test.encoded)
			if test.expectedError != nil {
				assert.Nil(t, key)
				assert.Equal(t, test.expectedError, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expectedKey, key)
			}
		})
	}
}

// TestKeyCacheGetHappyPath tests that the key cache can return an item
func TestKeyCacheGetHappyPath(t *testing.T) {
	keyID := "abc"
	kc := KeyCache{}
	kc.Store(keyID, testPublicKey, testCachePeriod)

	cachedKey, ok := kc.Get(keyID)

	assert.True(t, ok)
	assert.Equal(t, testPublicKey, cachedKey)
}

// TestKeyCacheGetNotFound tests that the key cache gracefully handles a non existent item
func TestKeyCacheGetNotFound(t *testing.T) {
	// store an item in the cache
	keyID := "abc"
	kc := KeyCache{}
	kc.Store(keyID, testPublicKey, testCachePeriod)

	// try and fetch a totally different item
	cachedKey, ok := kc.Get("xyz")

	assert.False(t, ok)
	assert.Nil(t, cachedKey)
}

// TestKeyCacheGetMultipleItems tests that the key cache returns the correct key when it holds
// multiple
func TestKeyCacheGetMultipleItems(t *testing.T) {
	keyID := "abc"
	kc := KeyCache{}
	kc.Store(keyID, testPublicKey, testCachePeriod)

	// Generate a bunch of extra keys and store them
	for i := 0; i < 5; i++ {
		_, publicKey := GenerateRsaKeyPair(32)
		kc.Store(string(i), publicKey, testCachePeriod)
	}

	cachedKey, ok := kc.Get(keyID)

	assert.True(t, ok)
	assert.Equal(t, testPublicKey, cachedKey)
}

// TestKeyCacheExpiredIsNotReturned tests that an expired item is not returned
func TestKeyCacheExpiredIsNotReturned(t *testing.T) {
	keyID := "abc"
	kc := KeyCache{}

	// Seed the cache with an expired item
	expires := time.Now().Add(-(time.Minute * 5))
	kc[keyID] = &CachedKey{
		Key:     testPublicKey,
		Expires: expires,
	}

	cachedKey, ok := kc.Get(keyID)

	assert.False(t, ok)
	assert.Nil(t, cachedKey)
}

// TestKeyCacheClearExpired tests that calling ClearExpired() will remove any expired items
func TestKeyCacheClearExpired(t *testing.T) {
	keyID := "dontdeleteme"
	kc := KeyCache{}

	// Seed the cache with a valid item
	kc.Store(keyID, testPublicKey, testCachePeriod)

	// ...and two expired items
	for i := 0; i < 2; i++ {
		_, publicKey := GenerateRsaKeyPair(32)
		expires := time.Now().Add(-(time.Minute * 5))
		kc[string(i)] = &CachedKey{
			Key:     publicKey,
			Expires: expires,
		}
	}

	// The cache should now have three items in it
	assert.Equal(t, 3, len(kc))

	// Run the cleanup
	kc.ClearExpired()

	// The cache should now have just one item in it
	assert.Equal(t, 1, len(kc))

	// ...and it should be the right one
	cachedKey, ok := kc.Get(keyID)

	assert.True(t, ok)
	assert.Equal(t, testPublicKey, cachedKey)
}

// TestKeyCacheLocking tests the key cache locking
func TestKeyCacheLocking(t *testing.T) {
	kc := KeyCache{}

	// Seed the cache with an expired item
	keyID := "Ihaveexpired"
	expires := time.Now().Add(-(time.Minute * 5))
	kc[keyID] = &CachedKey{
		Key:     testPublicKey,
		Expires: expires,
	}

	// Run clear - this should lock the cache and ultimately remove the expired item
	go kc.ClearExpired()

	// Attempt to fetch the item - this should not succeed as the item should have been removed
	// by the previous routine
	cachedKey, ok := kc.Get(keyID)

	assert.False(t, ok)
	assert.Nil(t, cachedKey)
}
