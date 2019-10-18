// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

func parseCertificateByteArray(rawCert []byte) (*x509.Certificate, error) {
	// Decode cert PEM
	certBlock, _ := pem.Decode(rawCert)
	if certBlock == nil {
		return nil, ErrInvalidCertPEM
	}

	// Convert the PEM block into an X509 Certificate
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// parsePrivateKeyFromBytes is a helper function that will produce a RSA private
// key from bytes.
func parsePrivateKeyFromBytes(pemData, password []byte) (key *rsa.PrivateKey, err error) {
	var pemBlock *pem.Block
	if pemBlock, _ = pem.Decode(pemData); pemBlock == nil {
		err = ErrInvalidPrivateKeyPEM
		return
	}

	decrypted := pemBlock.Bytes
	if x509.IsEncryptedPEMBlock(pemBlock) {
		if decrypted, err = x509.DecryptPEMBlock(pemBlock, password); err != nil {
			return
		}
	}

	key, err = x509.ParsePKCS1PrivateKey(decrypted)
	return
}

// NewX509CertificateSupplierFromByteArrays creates a new instance of the X509CertificateSupplier using byte array data for the arguments.
// For password protected private keys use 'privateKeyPw' to specify the password. For non password-protected private keys this argument
// should be set to nil instead.
func NewX509CertificateSupplierFromByteArrays(tenantID string, rawCert []byte, rawIntermediates [][]byte, rawPrivateKey []byte, privateKeyPw []byte) (*X509CertificateSupplier, error) {

	// Check arguments
	if rawCert == nil {
		return nil, ErrInvalidCertPEM
	}

	if rawIntermediates == nil {
		return nil, ErrInvalidIntermediateCertPEM
	}

	// Parse certificate data
	cert, err := parseCertificateByteArray(rawCert)
	if err != nil {
		return nil, err
	}

	// Parse intermediate certificate data
	var intermediates = make([]*x509.Certificate, len(rawIntermediates))
	for i, rawIntermediate := range rawIntermediates {
		intermediate, ierr := parseCertificateByteArray(rawIntermediate)
		if ierr != nil {
			return nil, ierr
		}
		intermediates[i] = intermediate
	}

	privateKey, err := parsePrivateKeyFromBytes(rawPrivateKey, privateKeyPw)
	if err != nil {
		return nil, err
	}

	return NewX509CertificateSupplier(tenantID, cert, intermediates, privateKey)
}
