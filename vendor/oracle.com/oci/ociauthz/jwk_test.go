// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"crypto/rsa"
	"math/big"

	"github.com/stretchr/testify/assert"
	"testing"
)

// from JWKsTest.java in the Java SDK
var (
	validJWKKeyID = "abcd"

	validJWKModulus = "AKoYq6Q7UN7vOFmPr4fSq2NORXHBMKm8p7h4JnQU-quLRxvYll9cn8OBhIXq9SnCYkbzBV" +
		"BkqN4ZyMM4vlSWy66wWdwLNYFDtEo1RJ6yZBExIaRVvX_eP6yRnpS1b7m7T2Uc2yPq1DnWzVI-sIGR51s1_ROnQZswkPJHh71PThln"

	validJWKExponent    = "AQAB"
	fourByteExponent    = "AAEAAQ"
	eightByteExponent   = "AAAAAAABAAE"
	sixteenByteExponent = "AAAAAAAAAAAAAAAAAABAAE"

	// 500000000001
	bigExponent      = "AAAAdGpSiAE"
	bigExponentBytes = []byte{0x00, 0x00, 0x00, 0x74, 0x6A, 0x52, 0x88, 0x01}
	bigExponentInt   = 500000000001

	validJWK       = NewJWKFromPublicKey(validJWKKeyID, testPublicKey)
	bigExponentKey = &rsa.PublicKey{
		N: new(big.Int).SetBytes(bigExponentBytes),
		E: bigExponentInt,
	}
)

// NewJWK creates a new JWK instance for use in tests
func NewJWK(keyID string, n string, e string) *JWK {
	return &JWK{
		KeyID:     keyID,
		KeyType:   "RSA",
		Algorithm: "RS256",
		Use:       "sig",
		N:         n,
		E:         e,
	}
}

// NewJWKFromPublicKey creates a new JSON Web Key from an existing RSA public key
func NewJWKFromPublicKey(keyID string, publicKey *rsa.PublicKey) *JWK {
	n := EncodeTokenPart(publicKey.N.Bytes())
	e := EncodeTokenPart(big.NewInt(int64(publicKey.E)).Bytes())

	return NewJWK(keyID, n, e)
}

// TestJWKPublicKey tests that PublicKey() returns the correct public key
func TestJWKPublicKey(t *testing.T) {
	publicKey, err := validJWK.PublicKey()

	assert.Nil(t, err)
	assert.Equal(t, testPublicKey, publicKey)
}

// TestJWKOnlyRSASupported tests that attempting to extract a public key from a non RSA-JWK results
// in an error
func TestJWKOnlyRSASupported(t *testing.T) {
	jwk := NewJWK(validJWKKeyID, validJWKModulus, validJWKExponent)
	jwk.KeyType = "EC"

	key, err := jwk.PublicKey()

	assert.Nil(t, key)
	assert.Equal(t, ErrUnsupportedJWKType, err)
}

func TestJWTPublicKey(t *testing.T) {
	testIO := []struct {
		tc     string
		jwk    *JWK
		expKey *rsa.PublicKey
		expErr error
	}{
		{tc: `should return valid key for valid jwt`,
			jwk: validJWK, expKey: testPublicKey, expErr: nil},
		{tc: `should return valid key for 4 byte padded exponent`,
			jwk: NewJWK(validJWKKeyID, validJWK.N, fourByteExponent), expKey: testPublicKey, expErr: nil},
		{tc: `should return valid key for 4 byte exponent padded to 8 bytes`,
			jwk: NewJWK(validJWKKeyID, validJWK.N, eightByteExponent), expKey: testPublicKey, expErr: nil},
		{tc: `should return valid key for 8 byte exponent`,
			jwk: NewJWK(validJWKKeyID, bigExponent, bigExponent), expKey: bigExponentKey, expErr: nil},
		{tc: `should return ErrUnsupportedExponentSize for 16 byte exponent`,
			jwk: NewJWK(validJWKKeyID, bigExponent, sixteenByteExponent), expKey: nil, expErr: ErrUnsupportedExponentSize},
		{tc: `should return ErrInvalidJWK for empty modulus value`,
			jwk: NewJWK(validJWKKeyID, ``, validJWKExponent), expKey: nil, expErr: ErrInvalidJWK},
		{tc: `should return ErrInvalidJWK for empty exponent value`,
			jwk: NewJWK(validJWKKeyID, validJWKModulus, ``), expKey: nil, expErr: ErrInvalidJWK},
		{tc: `should return ErrInvalidJWK for bad modulus value`,
			jwk: NewJWK(validJWKKeyID, `««BOOM»»`, validJWKExponent), expKey: nil, expErr: ErrInvalidJWK},
		{tc: `should return ErrInvalidJWK for bad exponent value`,
			jwk: NewJWK(validJWKKeyID, validJWKModulus, `««BOOM»»`), expKey: nil, expErr: ErrInvalidJWK},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			key, err := test.jwk.PublicKey()
			assert.Equal(t, test.expErr, err)
			assert.Equal(t, test.expKey, key)
		})
	}
}
