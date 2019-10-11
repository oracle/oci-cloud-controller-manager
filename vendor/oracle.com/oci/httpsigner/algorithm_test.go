// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package httpsigner

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestSignInvalidKeyRSASHA256 tests AlgorithmRSASHA256.Sign returns
// ErrInvalidKey when the key is not of type *rsa.PrivateKey
func TestSignInvalidKeyRSASHA256(t *testing.T) {
	message := []byte("msg")
	key := "invalidkey"
	sig, err := AlgorithmRSASHA256.Sign(message, key)
	assert.Nil(t, sig)
	assert.Equal(t, ErrInvalidKey, err)
}

// TestSignPKCS1v15ErrRSASHA256 tests AlgorithmRSASHA256.Sign returns
// an error from rsa.SignPKCS1v15 by using a mockReader for rand.Reader.
// This returns an error when rsa.SignPKCS1v15 calls Read() on that reader
func TestSignPKCS1v15ErrRSASHA256(t *testing.T) {
	message := []byte("msg")
	key, _ := NewPKCS1RSAPrivateKeyFromPEM(testPrivateKey)

	origRand := rand.Reader
	defer func() {
		rand.Reader = origRand
	}()
	rand.Reader = mockReader

	sig, err := AlgorithmRSASHA256.Sign(message, key)
	assert.Nil(t, sig)
	assert.Equal(t, errMock, err)
}

// TestSignInvalidKeyRSASHA256 tests AlgorithmRSAPSSSHA256.Sign returns
// ErrInvalidKey when the key is not of type *rsa.PrivateKey
func TestSignInvalidKeyRSAPSSSHA256(t *testing.T) {
	message := []byte("msg")
	key := "invalidkey"
	sig, err := AlgorithmRSAPSSSHA256.Sign(message, key)
	assert.Nil(t, sig)
	assert.Equal(t, ErrInvalidKey, err)
}

// TestSignPKCS1v15ErrRSAPSSSHA256 tests AlgorithmRSAPSSSHA256.Sign returns
// an error from rsa.SignPKCS1v15 by using a mockReader for rand.Reader.
// This returns an error when rsa.SignPSS calls Read() on that reader
func TestSignPKCS1v15ErrRSAPSSSHA256(t *testing.T) {
	message := []byte("msg")
	key, _ := NewPKCS1RSAPrivateKeyFromPEM(testPrivateKey)

	origRand := AlgorithmRSAPSSSHA256.randReader
	defer func() {
		AlgorithmRSAPSSSHA256.randReader = origRand
	}()
	AlgorithmRSAPSSSHA256.randReader = mockReader

	sig, err := AlgorithmRSAPSSSHA256.Sign(message, key)
	assert.Nil(t, sig)
	assert.Equal(t, errMock, err)
}

func TestDigest(t *testing.T) {
	testMsg := []byte(`doublesecret`)
	testMsgDigest := sha256.Sum256(testMsg)
	emptyMsg := []byte{}
	emptyMsgDigest := sha256.Sum256(emptyMsg)

	testIO := []struct {
		tc             string
		msg            []byte
		hash           crypto.Hash
		expectedDigest []byte
		expectedError  error
	}{
		{tc: `should return expected hash value for valid input`,
			msg: testMsg, hash: crypto.SHA256, expectedDigest: testMsgDigest[:], expectedError: nil},
		{tc: `should return ErrUnsupportedHash for unavailable hash`,
			msg: testMsg, hash: 1337, expectedDigest: nil, expectedError: ErrUnsupportedHash},
		{tc: `should return valid digest for empty message`,
			msg: emptyMsg, hash: crypto.SHA256, expectedDigest: emptyMsgDigest[:], expectedError: nil},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			dval, err := digest(test.msg, test.hash)
			if test.expectedError == nil {
				assert.Equal(t, test.expectedDigest, dval)
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, test.expectedError, err)
			}
		})
	}
}

// test data for "RSAAlgorithms" tests
var (
	message    = []byte(`doublesecret`)
	privKey, _ = rsa.GenerateKey(rand.Reader, 2048)

	rsaAlgorithms = []Algorithm{
		AlgorithmRSAPSSSHA256,
		AlgorithmRSASHA256,
	}
)

func TestSignVerifyRSAAlgorithms(t *testing.T) {
	for _, alg := range rsaAlgorithms {
		t.Run(alg.Name(), func(t *testing.T) {
			// test signature
			signed, err := alg.Sign(message, privKey)
			assert.NotNil(t, signed)
			assert.Nil(t, err)

			// invalid key
			err = alg.Verify(message, signed, privKey)
			assert.NotNil(t, err)
			assert.Equal(t, ErrInvalidKey, err)

			// test verify fails
			err = alg.Verify([]byte{}, signed, &privKey.PublicKey)
			assert.NotNil(t, err)
			assert.Equal(t, rsa.ErrVerification, err)

			// test verify
			err = alg.Verify(message, signed, &privKey.PublicKey)
			assert.Nil(t, err)
		})
	}
}

func TestRSAAlgorithmsNoSHA256(t *testing.T) {
	// temporarily disable SHA256
	defer crypto.RegisterHash(crypto.SHA256, sha256.New)
	crypto.RegisterHash(crypto.SHA256, nil)

	for _, alg := range rsaAlgorithms {
		t.Run(alg.Name(), func(t *testing.T) {

			// Sign
			sig, err := alg.Sign([]byte{}, privKey)
			assert.Nil(t, sig)
			assert.NotNil(t, err)
			assert.Equal(t, ErrUnsupportedHash, err)

			// Verify
			err = alg.Verify([]byte{}, []byte{}, &privKey.PublicKey)
			assert.NotNil(t, err)
			assert.Equal(t, ErrUnsupportedHash, err)
		})
	}
}

func TestAlgorithmsAlgorithm(t *testing.T) {
	testIO := []struct {
		tc      string
		algs    Algorithms
		algName string
		expAlg  Algorithm
		expErr  error
	}{
		{tc: `should return algorithm if available`,
			algs: StdAlgorithms, algName: AlgRSAPSSSHA256, expAlg: AlgorithmRSAPSSSHA256, expErr: nil},
		{tc: `should return ErrUnsupportedAlgorithm when requested algorithm not available`,
			algs: StdAlgorithms, algName: algEcho, expAlg: nil, expErr: ErrUnsupportedAlgorithm},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			alg, err := test.algs.Algorithm(test.algName)
			if test.expErr == nil {
				assert.Nil(t, err)
				assert.Equal(t, test.expAlg, alg)
			} else {
				assert.Nil(t, alg)
				assert.NotNil(t, err)
				assert.Equal(t, test.expErr, err)
			}
		})
	}
}
