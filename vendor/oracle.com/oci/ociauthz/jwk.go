// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"crypto/rsa"
	"encoding/binary"
	"math/big"
)

// JWK represents a JSON Web Key
// See: https://tools.ietf.org/html/rfc7517
type JWK struct {
	// rawJwk string
	KeyID     string `json:"kid,omitempty"`
	KeyType   string `json:"kty,omitempty"`
	Algorithm string `json:"alg,omitempty"`
	Use       string `json:"use,omitempty"`

	// Key fields. One or more of these will be set depending on the key type
	N string `json:"n,omitempty"`
	E string `json:"e,omitempty"`
}

// PublicKey extracts the modulus (N) and exponent (E) from the JWK and returns a valid
// *rsa.PublicKey structure. The method supports only JWKs of type RSA and will return
// ErrUnsupportedJWKType for other types. Errors extracting the key material will result in
// ErrInvalidJWK being returned.
func (k *JWK) PublicKey() (*rsa.PublicKey, error) {
	if k.KeyType != "RSA" {
		return nil, ErrUnsupportedJWKType
	}

	if k.N == "" || k.E == "" {
		return nil, ErrInvalidJWK
	}

	// Decode modulus
	modBytes, err := decodeTokenPart(k.N)
	if err != nil {
		return nil, ErrInvalidJWK
	}

	// Decode exponent
	expBytes, err := decodeTokenPart(k.E)
	if err != nil {
		return nil, ErrInvalidJWK
	}

	// cannot accept exponents greater than 64bits
	if len(expBytes) > 8 {
		return nil, ErrUnsupportedExponentSize
	}

	// Pad out the exponent value to the size of a native int
	if len(expBytes) < 8 {
		expN := make([]byte, 8)
		copy(expN[8-len(expBytes):], expBytes)
		expBytes = expN
	}

	// Use these to form the public key
	publicKey := &rsa.PublicKey{
		N: new(big.Int).SetBytes(modBytes),
		E: int(binary.BigEndian.Uint64(expBytes[:])),
	}

	return publicKey, nil
}
