// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"oracle.com/oci/httpsigner"
)

var (
	testTokenParser    = NewTokenParser(testKeyService, httpsigner.JWTAlgorithms)
	testTokenStrBadJwk = GenerateToken(
		Header{KeyID: "asw", Algorithm: "RS256"},
		[]Claim{{"", ClaimJWK, `?`}},
		testPrivateKey,
	)
	errJwkJSON = json.Unmarshal([]byte(`?`), &JWK{})
)

// TestNewSTSPubKeySupplier tests that NewSTSPubKeySupplier returns a new STSPubKeySupplier
// instance
func TestNewSTSPubKeySupplier(t *testing.T) {
	s := NewSTSPubKeySupplier(testTokenParser)

	var stsPubKSType *STSPubKeySupplier
	assert.IsType(t, stsPubKSType, s)
	assert.Equal(t, testTokenParser, s.tokenParser)

	assert.Panics(t, func() { NewSTSPubKeySupplier(nil) })
}

// TestSTSPubKeySupplierKeyHappyPath tests that the STSPubKeySupplier can extract the public key
// from a given JWT
func TestSTSPubKeySupplierKeyHappyPath(t *testing.T) {
	jwk := NewJWKFromPublicKey(validJWKKeyID, testPublicKey)
	jwkJSON, err := json.Marshal(jwk)
	assert.Nil(t, err)

	header := Header{
		KeyID:     "asw",
		Algorithm: "RS256",
	}
	claims := []Claim{
		{testIssuer, ClaimExpires, string(time.Now().Add(time.Hour * 1).Unix())},
		{testIssuer, ClaimJWK, string(jwkJSON)},
	}

	jwtStr := GenerateToken(header, claims, testPrivateKey)

	s := NewSTSPubKeySupplier(testTokenParser)
	key, err := s.Key("ST$" + jwtStr)

	assert.Nil(t, err)
	assert.Equal(t, testPublicKey, key)
}

// Verify that Key will return errors from converting to *rsa.PublicKey
func TestSTSPubKeySupplierJWKError(t *testing.T) {
	jwk := NewJWKFromPublicKey(validJWKKeyID, testPublicKey)
	jwk.KeyType = "EC"
	jwkJSON, err := json.Marshal(jwk)
	assert.Nil(t, err)

	header := Header{
		KeyID:     "asw",
		Algorithm: "RS256",
	}
	claims := []Claim{
		{testIssuer, ClaimExpires, string(time.Now().Add(time.Hour * 1).Unix())},
		{testIssuer, ClaimJWK, string(jwkJSON)},
	}

	jwtStr := GenerateToken(header, claims, testPrivateKey)

	s := NewSTSPubKeySupplier(testTokenParser)
	key, err := s.Key("ST$" + jwtStr)

	assert.Nil(t, key)
	assert.NotNil(t, err)
	assert.Equal(t, ErrUnsupportedJWKType, err)
}

// TestSTSPubKeySupplierKeyNoJWK tests that the STSPubKeySupplier gracefully handles a token
// that does not contain a JWK
func TestSTSPubKeySupplierKeyNoJWK(t *testing.T) {
	s := NewSTSPubKeySupplier(testTokenParser)
	key, err := s.Key("ST$" + validSignedToken)

	assert.Nil(t, key)
	assert.Equal(t, err, ErrNoJWK)
}

// TestBadJWK should return json unmarshal error for bad JWK
func TestBadJWK(t *testing.T) {
	s := NewSTSPubKeySupplier(testTokenParser)
	key, err := s.Key("ST$" + testTokenStrBadJwk)

	assert.Nil(t, key)
	assert.Equal(t, err, errJwkJSON)
}

// TestSTSPubKeySupplierKeyInvalidKey tests that the STSPubKeySupplier rejects a key that cannot
// be decoded
func TestSTSPubKeySupplierKeyInvalidKey(t *testing.T) {
	s := NewSTSPubKeySupplier(testTokenParser)
	key, err := s.Key("ST$not.areal.key")

	assert.Nil(t, key)

	var expectedErrType *json.SyntaxError
	assert.IsType(t, expectedErrType, err)
}

// TestSTSPubKeySupplierKeyInvalidKeyPrefix tests that the key supplier rejects keys withou a "ST$"
// prefix
func TestSTSPubKeySupplierKeyInvalidKeyPrefix(t *testing.T) {
	s := NewSTSPubKeySupplier(testTokenParser)
	key, err := s.Key("XY$" + validSignedToken)

	assert.Nil(t, key)
	assert.Equal(t, httpsigner.ErrInvalidKey, err)
}

func TestSTSPubKeySupplierKeyLength(t *testing.T) {
	testIO := []struct {
		tc            string
		key           string
		expectedError error
	}{
		{tc: `should return ErrInvalidKey if the key is empty`,
			key: "", expectedError: httpsigner.ErrInvalidKey},
		{tc: `should return ErrInvalidKey if the key length is less than 3`,
			key: "ST", expectedError: httpsigner.ErrInvalidKey},
		{tc: `should return ErrJWTMalformed from the parser if the key length is exactly 3 with ST$ prefix`,
			key: "ST$", expectedError: ErrJWTMalformed},
	}
	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			s := NewSTSPubKeySupplier(testTokenParser)
			key, err := s.Key(test.key)

			assert.Nil(t, key)
			assert.Equal(t, test.expectedError, err)
		})
	}
}
