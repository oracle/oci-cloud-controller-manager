// Copyright (c) 2017-2019, Oracle and/or its affiliates. All rights reserved.

package httpsigner

import (
	"errors"
	"fmt"
)

// Error Definitions
var (
	// ErrInvalidKey error means a key is invalid
	ErrInvalidKey = errors.New("invalid key")

	// ErrKeyNotFound is returned when a given operation cannot find a key associated with a given key ID.
	ErrKeyNotFound = errors.New("no key found with given ID")

	// ErrReplacementKeyIDEmpty indicates failure to automatically rotate key due to empty replacement keyID
	ErrReplacementKeyIDEmpty = errors.New("replacement KeyID is empty, key could not be rotated")

	// ErrUnsupportedHash denotes attempting to use an unsupported hash algorithm
	ErrUnsupportedHash = errors.New("unsupported Hash algorithm")

	// ErrUnsupportedAlgorithm indicates that an unsupported algorithm was requested
	ErrUnsupportedAlgorithm = errors.New("unsupported signature algorithm")

	// ErrMissingAuthzHeader indicates an `Authorization` header was expected, but not found
	ErrMissingAuthzHeader = errors.New("missing Authorization header")

	// ErrUnsupportedScheme indicates that the Auhtorization header had an unsupported scheme.  This is typically
	// `Signature` for this library.
	ErrUnsupportedScheme = errors.New("unsupported Authorization header scheme")

	// Invalid argument errors
	ErrInvalidKeyArg            = newErrorInvalidArg("key")
	ErrInvalidRequest           = newErrorInvalidArg("request")
	ErrInvalidKeyID             = newErrorInvalidArg("keyID")
	ErrInvalidKeySupplier       = newErrorInvalidArg("keySupplier")
	ErrInvalidAlgorithm         = newErrorInvalidArg("algorithm")
	ErrInvalidAlgorithmSupplier = newErrorInvalidArg("algorithmSupplier")
)

// ErrorInvalidArg is an error which stores the name of an invalid argument
type ErrorInvalidArg struct {
	argName string
}

// Error returns a string detailing the error
func (e *ErrorInvalidArg) Error() string {
	return fmt.Sprintf("httpsigner: Invalid argument value for '%s'", e.argName)
}

// newErrorInvalidArg returns a pointer to an ErrorInvalidArg struct with a given arg name
func newErrorInvalidArg(name string) error {
	return &ErrorInvalidArg{name}
}

// KeyRotationError is an error type that contains a replacement keyID so the caller may retry the call with
// up-to-date keyID.
type KeyRotationError struct {
	ReplacementKeyID string
	OldKeyID         string
}

// Error returns a string which describes the error.
func (e *KeyRotationError) Error() string {
	return "Requested key has expired and has been automatically rotated."
}

// NewKeyRotationError returns a new instance of KeyRotationError with the given replacement keyID
func NewKeyRotationError(replacement string, old string) *KeyRotationError {
	return &KeyRotationError{ReplacementKeyID: replacement, OldKeyID: old}
}
