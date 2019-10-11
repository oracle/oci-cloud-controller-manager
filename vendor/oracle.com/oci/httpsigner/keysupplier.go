// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package httpsigner

import (
	"crypto/rsa"
	"regexp"
)

// KeySupplier defines an object that will supply a "key" identified by keyID.
type KeySupplier interface {

	// Key should return the key indicated by keyID if possible.  If the supplier is unable to supply the indicated key
	// it SHOULD respond with ErrKeyNotFound.  If the supplier recognises that the requested key has been rotated, it
	// MAY respond with a new KeyRotationError that includes the ID of the replacement key.
	Key(keyID string) (key interface{}, err error)
}

// StaticRSAKeySupplier holds and supplies a single *rsa.PrivateKey
type StaticRSAKeySupplier struct {
	key   interface{}
	keyID string
}

// Key returns the *rsa.PrivateKey
func (s *StaticRSAKeySupplier) Key(keyID string) (interface{}, error) {
	if keyID == s.keyID {
		return s.key, nil
	}
	return nil, ErrKeyNotFound
}

// NewStaticRSAKeySupplier creates a StaticRSAKeySupplier with the given *rsa.PrivateKey. It will respond with
// ErrInvalidKeyArg if the given key is nil.
func NewStaticRSAKeySupplier(key *rsa.PrivateKey, keyID string) (*StaticRSAKeySupplier, error) {
	if key == nil {
		return nil, ErrInvalidKeyArg
	}
	return &StaticRSAKeySupplier{key: key, keyID: keyID}, nil
}

// NewStaticRSAPubKeySupplier creates a StaticRSAKeySupplier with the given *rsa.PublicKey. It will respond with
// ErrInvalidKeyArg if the given key is nil.
func NewStaticRSAPubKeySupplier(key *rsa.PublicKey, keyID string) (*StaticRSAKeySupplier, error) {
	if key == nil {
		return nil, ErrInvalidKeyArg
	}
	return &StaticRSAKeySupplier{key: key, keyID: keyID}, nil
}

// KeySupplierMux is a composite key supplier that can route requests between multiple KeySupplier implementations based
// upon the result of a regexp match.
type KeySupplierMux struct {
	suppliers map[string]KeySupplier
	patterns  map[string]*regexp.Regexp
}

// NewKeySupplierMux builds a new KeySupplierMux given a map of string regexp patterns and KeySuppliers.  Any pattern
// failing to compile results in a panic.
func NewKeySupplierMux(suppliers map[string]KeySupplier) *KeySupplierMux {

	// pre-compile the patterns for performance
	patterns := make(map[string]*regexp.Regexp, len(suppliers))
	for p := range suppliers {
		patterns[p] = regexp.MustCompile(p)
	}

	return &KeySupplierMux{
		suppliers: suppliers,
		patterns:  patterns,
	}
}

// Key will match keyID with each regexp and call the associated KeySupplier. Note that the order of comparison is
// indeterminate so the regular expressions should not overlap.  If no pattern matches keyID, then ErrKeyNotFound is
// returned.
func (ksm *KeySupplierMux) Key(keyID string) (key interface{}, err error) {
	for pattern, re := range ksm.patterns {
		if re.MatchString(keyID) {
			return ksm.suppliers[pattern].Key(keyID)
		}
	}
	return nil, ErrKeyNotFound
}
