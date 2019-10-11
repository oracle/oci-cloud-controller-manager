// Copyright (c) 2017-2019, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"fmt"
	"strings"

	"io/ioutil"
	"net/http"
	"net/url"

	"oracle.com/oci/httpsigner"
)

// CertificateSupplier represents all suppliers of x509 certificates
type CertificateSupplier interface {
	// Certificate returns X509 certificates
	CertificateOrError() (*x509.Certificate, error)
	// Certificate returns X509 certificates or nil if there is an error
	// Deprecated: Use CertificateOrError instead
	Certificate() *x509.Certificate
	// Intermediate returns intermediate X509 certificates
	IntermediateOrError() ([]*x509.Certificate, error)
	// Intermediate returns intermediate X509 certificates or nil if there is an error
	// Deprecated: Use IntermediateOrError instead
	Intermediate() []*x509.Certificate
	// PrivateKey will always return the private key associated with the x509 certificate
	PrivateKeyOrError() (*rsa.PrivateKey, error)
	// PrivateKey will always return the private key associated with the x509 certificate or nil if there is an error
	// Deprecated: Use PrivateKeyOrError instead
	PrivateKey() *rsa.PrivateKey
}

// X509CertificateSupplier is an implementation of httpsigner.KeySupplier that stores X509 certificates and the private key.
type X509CertificateSupplier struct {
	certificate  *x509.Certificate
	intermediate []*x509.Certificate
	private      *rsa.PrivateKey
	tenantID     string
	keyID        string
}

// KeyID returns the keyID generated from the x509 certificate
func (x *X509CertificateSupplier) KeyID() string {
	return x.keyID
}

// Key will return the private key associated with x509 certificate given the correct keyID
func (x *X509CertificateSupplier) Key(keyID string) (interface{}, error) {
	if keyID != x.KeyID() {
		return nil, httpsigner.ErrKeyNotFound
	}
	return x.PrivateKey(), nil
}

// PrivateKey will always return the private key associated with the x509 certificate
// Deprecated: Use PrivateKeyOrError instead
func (x *X509CertificateSupplier) PrivateKey() *rsa.PrivateKey {
	return x.private
}

// PrivateKeyOrError will always return the private key associated with the x509 certificate
func (x *X509CertificateSupplier) PrivateKeyOrError() (*rsa.PrivateKey, error) {
	return x.private, nil
}

// Certificate returns X509 certificates
// Deprecated: Use CertificateOrError instead
func (x *X509CertificateSupplier) Certificate() *x509.Certificate {
	return x.certificate
}

// CertificateOrError returns X509 certificates
func (x *X509CertificateSupplier) CertificateOrError() (*x509.Certificate, error) {
	return x.certificate, nil
}

// Intermediate returns intermediate X509 certificates
// Deprecated: Use CertificateOrError instead
func (x *X509CertificateSupplier) Intermediate() []*x509.Certificate {
	return x.intermediate
}

// IntermediateOrError returns intermediate X509 certificates
func (x *X509CertificateSupplier) IntermediateOrError() ([]*x509.Certificate, error) {
	return x.intermediate, nil
}

// GenerateKeyID will generate a keyID string from the given x509 certificate and tenantID
func GenerateKeyID(certificate *x509.Certificate, tenantID string) string {
	keyParts := []string{
		tenantID,
		"fed-x509",
		strings.Replace(fmt.Sprintf("% X", sha1.Sum(certificate.Raw)), " ", ":", -1),
	}

	return strings.Join(keyParts, "/")
}

// NewX509CertificateSupplier returns a new instance of X509CertificateSupplier
func NewX509CertificateSupplier(tenantID string, certificate *x509.Certificate, intermediate []*x509.Certificate, private *rsa.PrivateKey) (*X509CertificateSupplier, error) {
	if certificate == nil {
		return nil, ErrInvalidCertificate
	}
	if private == nil {
		return nil, ErrInvalidPrivateKey
	}
	if tenantID == "" {
		return nil, ErrInvalidTenantID
	}

	keyID := GenerateKeyID(certificate, tenantID)

	return &X509CertificateSupplier{
		tenantID:     tenantID,
		certificate:  certificate,
		intermediate: intermediate,
		private:      private,
		keyID:        keyID,
	}, nil
}

// URLX509CertificateSupplier provides certificates from a URLs
type URLX509CertificateSupplier struct {
	client                                  httpsigner.Client
	tenantID, certificateURL, privateKeyURL string
	privateKeyPassphrase                    []byte
	intermediateURLs                        []string
}

// NewX509CertificateSupplierFromURLs builds a CertificateSupplier from the URL arguments
func NewX509CertificateSupplierFromURLs(client httpsigner.Client, tenantID, certificateURL, privateKeyURL string, keypassphrase []byte, intermediateURL ...string) (*URLX509CertificateSupplier, error) {

	if tenantID == "" {
		return nil, ErrInvalidTenantID
	}

	if client == nil {
		return nil, ErrInvalidClient
	}
	urls := append([]string{certificateURL, privateKeyURL}, intermediateURL...)

	for _, u := range urls {
		if _, e := url.Parse(u); e != nil {
			return nil, e
		}
	}

	return &URLX509CertificateSupplier{
		client:               client,
		tenantID:             tenantID,
		certificateURL:       certificateURL,
		intermediateURLs:     intermediateURL,
		privateKeyURL:        privateKeyURL,
		privateKeyPassphrase: keypassphrase,
	}, nil
}

// KeyID returns the keyID generated from the x509 certificate
func (x URLX509CertificateSupplier) KeyID() (string, error) {
	cert, err := x.CertificateOrError()
	if err != nil {
		return "", err
	}
	return GenerateKeyID(cert, x.tenantID), nil
}

// Key will return the private key associated with x509 certificate given the correct keyID, or it will
// return a KeyRotationError with the current keyID if the provided keyID does not correspond to the
// one calculated from the current certificate. This is intended to detect a cert rotation, but can also
// be the result of an invalid ID passed as a parameter.
func (x URLX509CertificateSupplier) Key(keyID string) (interface{}, error) {
	calculatedKeyID, err := x.KeyID()
	if err != nil {
		return nil, err
	}
	if keyID != calculatedKeyID {
		return nil, httpsigner.NewKeyRotationError(calculatedKeyID, keyID)
	}
	return x.PrivateKeyOrError()
}

// PrivateKey will always return the private key associated with the x509 certificate or nil if there is an error
// Deprecated: Use PrivateKeyOrError
func (x URLX509CertificateSupplier) PrivateKey() *rsa.PrivateKey {
	key, err := x.PrivateKeyOrError()
	if err != nil {
		return nil
	}
	return key
}

// PrivateKeyOrError will always return the private key associated with the x509 certificate
func (x URLX509CertificateSupplier) PrivateKeyOrError() (*rsa.PrivateKey, error) {
	content, err := getURLContentOrError(x.privateKeyURL, x.client)
	if err != nil {
		return nil, err
	}
	return parsePrivateKeyFromBytes(content, x.privateKeyPassphrase)
}

// Certificate returns X509 certificates or nil if there is an error
// Deprecated: Use CertificateOrError
func (x URLX509CertificateSupplier) Certificate() *x509.Certificate {
	certificate, err := x.CertificateOrError()
	if err != nil {
		return nil
	}
	return certificate
}

// CertificateOrError returns X509 certificates
func (x URLX509CertificateSupplier) CertificateOrError() (*x509.Certificate, error) {
	content, err := getURLContentOrError(x.certificateURL, x.client)
	if err != nil {
		return nil, err
	}
	return parseCertificateByteArray(content)
}

// Intermediate returns intermediate X509 certificates. It returns if it fails to
// read a certificate from any of the provided URLs. Further it stops reading
// certificates after the first error
// Deprecated: Use IntermediateOrError
func (x URLX509CertificateSupplier) Intermediate() []*x509.Certificate {
	certificates, err := x.IntermediateOrError()
	if err != nil {
		return nil
	}
	return certificates

}

// IntermediateOrError returns intermediate X509 certificates. It returns if it fails to
// read a certificate from any of the provided URLs. Further it stops reading
// certificates after the first error
func (x URLX509CertificateSupplier) IntermediateOrError() ([]*x509.Certificate, error) {
	certificates := make([]*x509.Certificate, len(x.intermediateURLs))
	for i, u := range x.intermediateURLs {
		content, err := getURLContentOrError(u, x.client)
		if err != nil {
			return nil, err
		}

		cert, err := parseCertificateByteArray(content)
		if err != nil {
			return nil, err
		}
		certificates[i] = cert

	}
	return certificates, nil
}

func getURLContentOrError(url string, client httpsigner.Client) (content []byte, err error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}

	response, err := client.Do(request)
	if err != nil {
		return
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		err = &ServiceResponseError{Response: response}
		return
	}

	content, err = ioutil.ReadAll(response.Body)
	return
}
