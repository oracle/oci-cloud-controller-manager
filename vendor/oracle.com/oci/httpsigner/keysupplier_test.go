// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package httpsigner

import (
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
)

// test constants
var (
	emptyKey           = &rsa.PrivateKey{}
	emptyPubKey        = &emptyKey.PublicKey
	testSupplier, _    = NewStaticRSAKeySupplier(emptyKey, testKeyID)
	testPubSupplier, _ = NewStaticRSAPubKeySupplier(emptyPubKey, testKeyID)
)

func TestNewStaticRSAKeySupplier(t *testing.T) {
	testIO := []struct {
		tc            string
		key           *rsa.PrivateKey
		keyID         string
		expectedError error
	}{
		{tc: `Should store valid key and id`,
			key: emptyKey, keyID: testKeyID, expectedError: nil},
		{tc: `should raise error on nil key`,
			key: nil, keyID: testKeyID, expectedError: ErrInvalidKeyArg},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			ks, err := NewStaticRSAKeySupplier(test.key, test.keyID)
			assert.Equal(t, test.expectedError, err)
			if err == nil {
				assert.Equal(t, test.key, ks.key)
				assert.Equal(t, test.keyID, ks.keyID)
			}
		})
	}
}

func TestStaticRSAKeySupplierKey(t *testing.T) {
	testIO := []struct {
		tc            string
		keyID         string
		expectedKey   *rsa.PrivateKey
		expectedError error
	}{
		{tc: `should return key and nil error with matching id`,
			keyID: testKeyID, expectedKey: emptyKey, expectedError: nil},
		{tc: `should return nil and ErrKeyNotFound for different id`,
			keyID: `ni`, expectedKey: nil, expectedError: ErrKeyNotFound},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			key, err := testSupplier.Key(test.keyID)
			assert.Equal(t, test.expectedError, err)
			if test.expectedError == nil {
				assert.Equal(t, test.expectedKey, key.(*rsa.PrivateKey))
				assert.Nil(t, err)
			} else {
				assert.Nil(t, test.expectedKey)
				assert.NotNil(t, err)
			}
		})
	}

}

func TestNewStaticRSAPubKeySupplier(t *testing.T) {
	testIO := []struct {
		tc            string
		key           *rsa.PublicKey
		keyID         string
		expectedError error
	}{
		{tc: `Should store valid key and id`,
			key: emptyPubKey, keyID: testKeyID, expectedError: nil},
		{tc: `should raise error on nil key`,
			key: nil, keyID: testKeyID, expectedError: ErrInvalidKeyArg},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			ks, err := NewStaticRSAPubKeySupplier(test.key, test.keyID)
			assert.Equal(t, test.expectedError, err)
			if err == nil {
				assert.Equal(t, test.key, ks.key)
				assert.Equal(t, test.keyID, ks.keyID)
			}
		})
	}
}

func TestStaticRSAPubKeySupplierKey(t *testing.T) {
	testIO := []struct {
		tc            string
		keyID         string
		expectedKey   *rsa.PublicKey
		expectedError error
	}{
		{tc: `should return key and nil error with matching id`,
			keyID: testKeyID, expectedKey: emptyPubKey, expectedError: nil},
		{tc: `should return nil and ErrKeyNotFound for different id`,
			keyID: `ni`, expectedKey: nil, expectedError: ErrKeyNotFound},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			key, err := testPubSupplier.Key(test.keyID)
			assert.Equal(t, test.expectedError, err)
			if test.expectedError == nil {
				assert.Equal(t, test.expectedKey, key.(*rsa.PublicKey))
				assert.Nil(t, err)
			} else {
				assert.Nil(t, test.expectedKey)
				assert.NotNil(t, err)
			}
		})
	}
}

func TestKeySupplierMux(t *testing.T) {
	re0 := "^a"
	re1 := "^b"
	ks0 := &MockKeySupplier{key: emptyKey}
	ks1 := &MockKeySupplier{key: nil, err: ErrKeyNotFound}
	singleMux := NewKeySupplierMux(map[string]KeySupplier{re0: ks0})
	multiMux := NewKeySupplierMux(map[string]KeySupplier{re0: ks1, re1: ks1})

	testIO := []struct {
		tc                string
		keyID             string
		supplier          KeySupplier
		expectedKs0Called bool
		expectedKs1Called bool
		expectedKey       interface{}
		expectedError     error
	}{
		{
			tc:                `Should return ErrKeyNotFound for empty mux`,
			keyID:             `dne`,
			supplier:          &KeySupplierMux{},
			expectedKs0Called: false,
			expectedKs1Called: false,
			expectedKey:       nil,
			expectedError:     ErrKeyNotFound,
		},
		{
			tc:                `Should return ErrKeyNotFound for non-matching keyid`,
			keyID:             `dne`,
			supplier:          singleMux,
			expectedKs0Called: false,
			expectedKs1Called: false,
			expectedKey:       nil,
			expectedError:     ErrKeyNotFound,
		},
		{
			tc:                `Should call matching KeySupplier and return that key`,
			keyID:             `akey`,
			supplier:          singleMux,
			expectedKs0Called: true,
			expectedKs1Called: false,
			expectedKey:       emptyKey,
			expectedError:     nil,
		},
		{
			tc:                `Should propagate errors from matching KeySupplier`,
			keyID:             `bkey`,
			supplier:          multiMux,
			expectedKs0Called: false,
			expectedKs1Called: true,
			expectedKey:       nil,
			expectedError:     ErrKeyNotFound,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			// setup
			ks0.Reset()
			ks1.Reset()

			// go
			key, err := test.supplier.Key(test.keyID)

			// verify
			assert.Equal(t, test.expectedKs0Called, ks0.KeyCalled)
			assert.Equal(t, test.expectedKs1Called, ks1.KeyCalled)
			if test.expectedError == nil {
				assert.Nil(t, err)
				assert.Equal(t, test.expectedKey, key)
			} else {
				assert.Nil(t, key)
				assert.Equal(t, test.expectedError, err)
			}
		})
	}
}
