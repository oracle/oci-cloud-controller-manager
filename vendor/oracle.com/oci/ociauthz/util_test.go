// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base32"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	"testing"

	"github.com/stretchr/testify/assert"
)

// PublicKeyToPEM converts a publicKey to PEM format. This is currently only used in unit
// tests
func PublicKeyToPEM(publicKey *rsa.PublicKey) ([]byte, error) {
	publicKeyDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, errors.New("Failed to marshal public key")
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyDER,
	}), nil
}

// TestBase64EncodeBlock tests that Base64EncodeBlock can convert a cert block to a base64
// encoded string
func TestBase64EncodeBlock(t *testing.T) {
	block := &pem.Block{Type: "CERTIFICATE", Bytes: testCertDER}
	result := base64EncodeBlock(block)

	var stringType string
	assert.IsType(t, stringType, result)
}

// TestBase64EncodeCertificate tests that Base64EncodeCertificate can convert a certificate
// into a base64 encoded string
func TestBase64EncodeCertificate(t *testing.T) {
	result := base64EncodeCertificate(testCert)

	var stringType string
	assert.IsType(t, stringType, result)
}

// TestGenerateRsaKeyPair tests that GenerateRsaKeyPair can return a valid key pair
func TestGenerateRsaKeyPair(t *testing.T) {
	privateKey, publicKey := GenerateRsaKeyPair(2048)

	var privateKeyType *rsa.PrivateKey
	assert.IsType(t, privateKeyType, privateKey)

	var publicKeyType *rsa.PublicKey
	assert.IsType(t, publicKeyType, publicKey)
}

// TestPEMToPublicKey tests that PEMToPublicKey can convert a PEM string into type *rsa.PublicKey
func TestPEMToPublicKey(t *testing.T) {
	publicKey, _ := PEMToPublicKey(publicKeyPEM)

	var publicKeyType *rsa.PublicKey
	assert.IsType(t, publicKeyType, publicKey)
}

// TestPublicKeyToPEM tests that PublicKeyToPEM can convert a public key into PEM format
func TestPublicKeyToPEM(t *testing.T) {
	_, publicKey := GenerateRsaKeyPair(32)
	publicKeyPEM, _ := PublicKeyToPEM(publicKey)

	expectedPublicKey, _ := PEMToPublicKey(string(publicKeyPEM))

	assert.Equal(t, expectedPublicKey, publicKey)
}

func TestPEMToPublicKeyError(t *testing.T) {
	privateKeyPEM := new(bytes.Buffer)
	pemBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(testPrivateKey)}
	pem.Encode(privateKeyPEM, pemBlock)
	testIO := []struct {
		tc            string
		key           string
		expectedError error
	}{
		{tc: `should handle error from decoding PEM`,
			key: `invalid-pem`},
		{tc: `should handle error when attempting to parse PublicKey PEM`,
			key: privateKeyPEM.String()},
		{tc: `should handle error when the decoded PEM is not RSA`,
			key: testStaticDSAPublicKey},
	}
	for _, test := range testIO {
		t.Run(fmt.Sprintf(test.tc), func(t *testing.T) {
			p, e := PEMToPublicKey(test.key)
			assert.Nil(t, p)
			assert.NotNil(t, e)
		})
	}
}

func TestTokenID(t *testing.T) {
	testVals := []int{10, 20, 30}
	for _, n := range testVals {
		tid, err := tokenID(n)
		assert.Equal(t, err, nil)
		l, _ := base32.StdEncoding.DecodeString(strings.ToUpper(tid))
		assert.Equal(t, len(l), n)
	}
}
