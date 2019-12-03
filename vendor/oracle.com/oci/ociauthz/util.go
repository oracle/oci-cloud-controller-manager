// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"
)

// base64EncodeBlock base64 encodes a block by first converting it to PEM format, and then stripping
// the headers and newlines from the string
func base64EncodeBlock(block *pem.Block) string {
	pem := pem.EncodeToMemory(block)

	s := string(pem)
	s = strings.Replace(s, "-----BEGIN CERTIFICATE-----", "", -1)
	s = strings.Replace(s, "-----END CERTIFICATE-----", "", -1)
	s = strings.Replace(s, "\n", "", -1)

	return s
}

// base64EncodeCertificate returns the base64 encoded version of the certificate, after first
// converting it to PEM format
func base64EncodeCertificate(cert *x509.Certificate) string {
	certBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		},
	)
	certBlock, _ := pem.Decode(certBytes)
	return base64EncodeBlock(certBlock)
}

// GenerateRsaKeyPair creates a new RSA key pair using the provided key size
func GenerateRsaKeyPair(keySize int) (*rsa.PrivateKey, *rsa.PublicKey) {
	privkey, _ := rsa.GenerateKey(rand.Reader, keySize)
	return privkey, &privkey.PublicKey
}

// PEMToPublicKey converts a PEM string to an rsa.PublicKey
func PEMToPublicKey(key string) (publicKey *rsa.PublicKey, err error) {
	// Parse PEM
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, ErrParsePEM
	}

	// Parse key
	rawKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	publicKey, ok := rawKey.(*rsa.PublicKey)
	if !ok {
		return nil, ErrExpectedRSAPublicKey
	}

	return publicKey, nil
}
