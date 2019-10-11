// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"encoding/json"
	"strings"

	"oracle.com/oci/httpsigner"
)

// STSPubKeySupplier implements httpsigner.KeySupplier to provide the public key from an STS token
type STSPubKeySupplier struct {
	tokenParser TokenParser
}

// NewSTSPubKeySupplier creates a new instance of the STSPubKeySupplier. Will panic on nil TokenParser.
func NewSTSPubKeySupplier(tokenParser TokenParser) *STSPubKeySupplier {
	if tokenParser == nil {
		panic(`Programmer Error: must provide a TokenParser`)
	}
	return &STSPubKeySupplier{tokenParser: tokenParser}
}

// Key takes an STS token and after verifying it, extracts and returns the embedded public key. KeyIDs other than STS
// tokens will result in httpsigner.ErrInvalidKey.  If there is an issue extracting the key, an appropriate error is
// returned.
func (s *STSPubKeySupplier) Key(keyID string) (interface{}, error) {
	// Check the token prefix
	if !strings.HasPrefix(keyID, "ST$") {
		return nil, httpsigner.ErrInvalidKey
	}

	// Parse the token
	token, err := s.tokenParser.Parse(keyID[3:])
	if err != nil {
		return nil, err
	}

	// Make sure there is a JWK
	rawJWK := token.Claims.GetString(ClaimJWK)
	if rawJWK == "" {
		return nil, ErrNoJWK
	}

	// Unmarshal JWK
	jwk := JWK{}
	err = json.Unmarshal([]byte(rawJWK), &jwk)
	if err != nil {
		return nil, err
	}

	// Extract the public key from the JWK
	publicKey, err := jwk.PublicKey()
	if err != nil {
		return nil, err
	}

	return publicKey, nil
}
