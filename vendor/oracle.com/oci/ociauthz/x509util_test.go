// Copyright (c) 2017-2019, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"

	"errors"
	"github.com/stretchr/testify/assert"
)

var (
	// testCert -> PEM -> byte array
	testCertBytes = pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: testCert.Raw,
		},
	)

	emptyIntermediateCertsByteArray = [][]byte{}
	intermediateCertsByteArray      = [][]byte{testCertBytes}

	testPrivateKeyPw    = []byte("supersecret")
	testPrivateKeyBytes = x509.MarshalPKCS1PrivateKey(testPrivateKey)
	testPrivateKeyBlock = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: testPrivateKeyBytes,
	}

	// create a password protected version of testPrivateKey
	testPrivateKeyWithPWBlock, _ = x509.EncryptPEMBlock(rand.Reader, testPrivateKeyBlock.Type, testPrivateKeyBlock.Bytes, testPrivateKeyPw, x509.PEMCipherAES256)

	testPrivateKeyWithPWBytes    = pem.EncodeToMemory(testPrivateKeyWithPWBlock)
	testPrivateKeyWithoutPWBytes = pem.EncodeToMemory(testPrivateKeyBlock)
)

func TestNewX509CertificateSupplierFromByteArrays(t *testing.T) {

	testIO := []struct {
		tc            string
		tenantID      string
		certificate   []byte
		intermediates [][]byte
		private       []byte
		password      []byte
		errorExpected bool
		expectedError error
	}{
		{tc: "should return error with empty tenant-id string",
			tenantID: "", certificate: testCertBytes, intermediates: emptyIntermediateCertsByteArray,
			private: testPrivateKeyWithPWBytes, password: testPrivateKeyPw, errorExpected: true, expectedError: ErrInvalidTenantID},
		{tc: "should return error with bad certificate data",
			tenantID: testTenantID, certificate: nil, intermediates: emptyIntermediateCertsByteArray,
			private: testPrivateKeyWithPWBytes, password: testPrivateKeyPw, errorExpected: true, expectedError: ErrInvalidCertPEM},
		{tc: "should return error with bad intermediate certificate data",
			tenantID: testTenantID, certificate: testCertBytes, intermediates: nil,
			private: testPrivateKeyWithPWBytes, password: testPrivateKeyPw, errorExpected: true, expectedError: ErrInvalidIntermediateCertPEM},
		{tc: "should return error with bad private key data",
			tenantID: testTenantID, certificate: testCertBytes, intermediates: emptyIntermediateCertsByteArray,
			private: nil, password: testPrivateKeyPw, errorExpected: true, expectedError: ErrInvalidPrivateKeyPEM},
		{tc: "should return error when using a password protected private key without a password",
			tenantID: testTenantID, certificate: testCertBytes, intermediates: emptyIntermediateCertsByteArray,
			private: testPrivateKeyWithPWBytes, password: nil, errorExpected: true},
		{tc: "should return x509 certificate supplier with valid input (pw-protected private key)",
			tenantID: testTenantID, certificate: testCertBytes, intermediates: intermediateCertsByteArray,
			private: testPrivateKeyWithPWBytes, password: testPrivateKeyPw},
		{tc: "should return x509 certificate supplier with valid input (non password protected private key)",
			tenantID: testTenantID, certificate: testCertBytes, intermediates: intermediateCertsByteArray,
			private: testPrivateKeyWithoutPWBytes, password: nil},
	}

	for _, test := range testIO {
		t.Run(fmt.Sprintf(test.tc), func(t *testing.T) {
			s, e := NewX509CertificateSupplierFromByteArrays(test.tenantID, test.certificate, test.intermediates, test.private, test.password)
			if test.errorExpected {
				assert.Nil(t, s)
				if test.expectedError != nil {
					assert.Equal(t, test.expectedError, e)
				}
			} else {
				assert.Nil(t, e)
				assert.NotNil(t, s)
				assert.Equal(t, testCert, s.Certificate())
				assert.Equal(t, []*x509.Certificate{testCert}, s.Intermediate())
				assert.Equal(t, testPrivateKey, s.PrivateKey())
				private, err := s.Key(s.KeyID())
				assert.Nil(t, err)
				assert.Equal(t, testPrivateKey, private.(*rsa.PrivateKey))
			}
		})
	}
}

func TestParsePrivateKeyFromBytes(t *testing.T) {
	randBytes := make([]byte, 100)
	rand.Read(randBytes)
	testIO := []struct {
		tc           string
		expectsError error
		inByte       []byte
		inPass       []byte
		outKey       *rsa.PrivateKey
	}{
		{
			tc:     "should return a valid key from bytes",
			inByte: testPrivateKeyWithPWBytes,
			inPass: testPrivateKeyPw,
			outKey: testPrivateKey,
		},
		{
			tc:     "should return a valid unecrypted key from bytes",
			inByte: testPrivateKeyWithoutPWBytes,
			inPass: nil,
			outKey: testPrivateKey,
		},
		{
			tc:           "should return an error with invalid password",
			expectsError: errors.New("private Key PEM data is invalid"),
			inByte:       testPrivateKeyBytes,
			inPass:       []byte("randomPassword"),
			outKey:       testPrivateKey,
		},
		{
			tc:           "should return an error with nil password",
			expectsError: errors.New("private Key PEM data is invalid"),
			inByte:       testPrivateKeyBytes,
			inPass:       nil,
			outKey:       testPrivateKey,
		},
		{
			tc:           "should return an error from invalid key bytes",
			expectsError: ErrInvalidPrivateKeyPEM,
			inByte:       randBytes,
			inPass:       testPrivateKeyPw,
			outKey:       testPrivateKey,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			k, e := parsePrivateKeyFromBytes(test.inByte, test.inPass)
			if test.expectsError != nil {
				assert.Nil(t, k)
				assert.Error(t, e)
				assert.Equal(t, test.expectsError, e)
			} else {
				assert.NoError(t, e)
				assert.Equal(t, test.outKey, k)
			}
		})
	}
}
