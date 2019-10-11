// Copyright (c) 2018, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"oracle.com/oci/httpsigner"
)

// TokenParser represents an object that a caller can use to parse a keyid to retrieve a JWT object representation.
// The implementation of this interface should manage a KeySupplier so that the caller does not have to manage it
// separately.
type TokenParser interface {

	// Parse takes a keyid string and returns a JWT object representation.  If error is encountered, it should return
	// a nil object and a corresponding error.
	Parse(keyID string) (jwt *Token, err error)
}

// tokenParser is the default implementation of TokenParser that simply wraps the call to ParseToken with preset key and
// algorithm suppliers.
type tokenParser struct {
	keySupplier httpsigner.KeySupplier
	algSupplier httpsigner.AlgorithmSupplier
}

// ParseToken delegates to the package function ParseToken using the KeySupplier and AlgorithmSupplier from the
// tokenParser instance.
func (tp *tokenParser) Parse(token string) (jwt *Token, err error) {
	return ParseToken(token, tp.keySupplier, tp.algSupplier)
}

// NewTokenParser builds a new token parser with the given KeySupplier and AlgorithmSupplier.  A nil KeySupplier or
// AlgorithmSupplier will result in a panic.
// SECURITY NOTE: The provided key supplier should *only* be able to supply keys trusted for token signing.  For
// example, do not pass a composite key supplier capable of looking up both service and api keys.
func NewTokenParser(ks httpsigner.KeySupplier, as httpsigner.AlgorithmSupplier) TokenParser {
	if ks == nil || as == nil {
		panic("Programmer error: must supply a KeySupplier and AlgorithmSupplier")
	}
	return &tokenParser{keySupplier: ks, algSupplier: as}
}
