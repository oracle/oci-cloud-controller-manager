// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package httpsigner

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"io"
)

// Included Signing Algorithms
var (
	// AlgorithmRSASHA256 is the 'rsa-sha256' signing algorithm
	AlgorithmRSASHA256 = algorithmRSASHA256{}

	// AlgorithmRSAPSSSHA256 is the 'rsa-pss-sha256' signing algorithm (using a 256-bit salt)
	AlgorithmRSAPSSSHA256 = algorithmRSAPSSSHA256{
		randReader: rand.Reader,
		pssOptions: rsa.PSSOptions{SaltLength: 32}}
)

//  Standard algorithm names
const (
	AlgRSASHA256    string = "rsa-sha256"
	AlgRSAPSSSHA256 string = "rsa-pss-sha256"
)

// StdAlgorithms defines the HTTP signature algorithms provided by this package
var StdAlgorithms = Algorithms{
	AlgRSASHA256:    AlgorithmRSASHA256,
	AlgRSAPSSSHA256: AlgorithmRSAPSSSHA256,
}

// Algorithm defines a common interface for interacting with http request signing algorithms
type Algorithm interface {
	// Name returns the name of the algorithm
	Name() string

	// Sign applies the implemented algorithm to the provided message using the given key and returns the resulting
	// signature.  If the key is not of the expected type then it should return ErrInvalidKey.
	Sign(message []byte, key interface{}) (signature []byte, err error)

	// Verify should verify that the signature applies to the given message using the provided key.  If the signature is
	// valid, it should return a nil error.  If the signature is invalid it should return an appropriate error.
	// Verify may also return invalid argument errors.
	Verify(message, signature []byte, key interface{}) error
}

type algorithmRSASHA256 struct {
}

// Name returns the name of the algorithm
func (a algorithmRSASHA256) Name() string {
	return AlgRSASHA256
}

// Sign applies the 'rsa-sha256' algorithm to the supplied message with the given key and returns the resulting
// signature.  It uses the crypto primitives from the golang std lib.  The supplied key must be an *rsa.PrivateKey or
// else ErrInvalidKey will be returned.
func (a algorithmRSASHA256) Sign(message []byte, key interface{}) (sig []byte, err error) {

	// key type
	priv, ok := key.(*rsa.PrivateKey)
	if !ok {
		err = ErrInvalidKey
		return
	}

	// digest message
	hashed, err := digest(message, crypto.SHA256)
	if err != nil {
		return
	}

	// delegate to stdlib for signature
	return rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hashed)
}

// Verify verifies that the given signature applies to the given message using the given key using rsa-sha256.  The
// underlying crypto operations are performed by the Go stdlib.  Will return nil when the signature is valid otherwise
// returns an appropriate error.  The key must be of type *rsa.PublicKey otherwise ErrInvalidKey will be returned.
func (a algorithmRSASHA256) Verify(message, signature []byte, key interface{}) (err error) {

	// key type
	pub, ok := key.(*rsa.PublicKey)
	if !ok {
		return ErrInvalidKey
	}

	// digest message
	hashed, err := digest(message, crypto.SHA256)
	if err != nil {
		return
	}

	// delegate signature check to stdlib
	return rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed, signature)
}

type algorithmRSAPSSSHA256 struct {
	// randRdr is the random source for PSS signature of this algorithm
	// It is set to rand.Reader for consumers of this library.
	randReader io.Reader
	pssOptions rsa.PSSOptions
}

// Name returns the name of the algorithm
func (a algorithmRSAPSSSHA256) Name() string {
	return AlgRSAPSSSHA256
}

// Sign applies the 'rsa-pss-sha256' algorithm to the supplied message with the given key and returns the resulting
// signature.  It uses the crypto primitives from the golang std lib.  The supplied key must be an *rsa.PrivateKey or
// else ErrInvalidKey will be returned.
func (a algorithmRSAPSSSHA256) Sign(message []byte, key interface{}) (sig []byte, err error) {

	// key type
	priv, ok := key.(*rsa.PrivateKey)
	if !ok {
		err = ErrInvalidKey
		return
	}

	// digest message
	hashed, err := digest(message, crypto.SHA256)
	if err != nil {
		return
	}

	// delegate to stdlib for signature
	return rsa.SignPSS(a.randReader, priv, crypto.SHA256, hashed, &a.pssOptions)
}

// Verify verifies that the given signature applies to the given message using the given key using rsa-pss-sha256.  The
// underlying crypto operations are performed by the Go stdlib.  Will return nil when the signature is valid otherwise
// returns an appropriate error.  The key must be of type *rsa.PublicKey otherwise ErrInvalidKey will be returned.
func (a algorithmRSAPSSSHA256) Verify(message, signature []byte, key interface{}) (err error) {

	// key type
	pub, ok := key.(*rsa.PublicKey)
	if !ok {
		return ErrInvalidKey
	}

	// digest message
	hashed, err := digest(message, crypto.SHA256)
	if err != nil {
		return
	}

	// delegate signature check to stdlib
	return rsa.VerifyPSS(pub, crypto.SHA256, hashed, signature, &a.pssOptions)
}

// digest calculates and returns the digest of the given message using the given Hash. It will return ErrUnsupportedHash
// if hash is not available.
func digest(message []byte, hash crypto.Hash) ([]byte, error) {
	if !hash.Available() {
		return nil, ErrUnsupportedHash
	}
	hasher := hash.New()
	hasher.Write(message)
	return hasher.Sum(nil), nil
}

// AlgorithmSupplier is a standard interface for an object that can return an Algorithm implementation given a name.
type AlgorithmSupplier interface {

	// Algorithm should return an appropriate Algorithm implementation given the provided name.  If no Algorithm by that
	// name is available, it should return ErrUnsupportedAlgorithm.
	Algorithm(name string) (Algorithm, error)
}

// Algorithms defines a table of Algorithm objects keyed by name
type Algorithms map[string]Algorithm

// Algorithm allows Algorithms to be used as an AlgorithmSupplier
func (a Algorithms) Algorithm(name string) (Algorithm, error) {
	alg, ok := a[name]
	if !ok {
		return nil, ErrUnsupportedAlgorithm
	}
	return alg, nil
}
