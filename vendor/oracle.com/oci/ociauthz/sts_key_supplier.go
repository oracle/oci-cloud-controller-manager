// Copyright (c) 2017-2019, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"oracle.com/oci/httpsigner"
)

// expirationTimePadding is extra time to subtract from the real expiration time to consider token expired
const expirationTimePadding = time.Duration(10) * time.Minute

// mutexKeyRotation used for locking read/write operation of the STSKeySupplier token.  The mutexKeyRotation is used from both
// Key() and KeyID() methods since they both have the ability to read/update the current token.
var mutexKeyRotation = &sync.Mutex{}

// STSToken extends Token and represents a JWT used for STS
type STSToken struct {
	*Token
	rawToken string
}

// NewSTSToken returns a new STS token instance
func NewSTSToken(rawToken string, ks httpsigner.KeySupplier) (*STSToken, error) {
	token, err := NewToken(rawToken, ks)
	if err != nil {
		return nil, err
	}

	return &STSToken{token, rawToken}, nil
}

// NewSTSTokenWithVerificationFn returns a new STS token instance validation, the verification function decides if the token
// needs to be validated or not
func newSTSTokenWithConditionalVerification(rawToken string, ks httpsigner.KeySupplier, shouldVerify tokenVerificationPolicy) (*STSToken, error) {
	token, err := newTokenWithTokenVerificationPolicy(rawToken, ks, shouldVerify)
	if err != nil {
		return nil, err
	}

	return &STSToken{token, rawToken}, nil
}

// String returns the raw token string with the "ST$" prefix
func (t *STSToken) String() string {
	if t.rawToken == "" {
		return ""
	}

	return formatSTSToken(t.rawToken)
}

// STSKeySupplier is an implementation of httpsigner.KeySupplier that will fetch STS token using x509 KeySupplier
type STSKeySupplier struct {
	sessionKeySupplier SessionKeySupplier
	token              *STSToken
	previousToken      *STSToken
	shouldVerifyToken  tokenVerificationPolicy
	endpoint           string
	tokenPurpose       string

	// remove once https://jira.oci.oraclecorp.com/browse/ID-1362 is resolved
	validationClientOptions *ClientOptions

	client       httpsigner.Client
	x509Supplier CertificateSupplier
}

// S2SRequest represents the data used in the body of the S2S request
type S2SRequest struct {
	Certificate              string   `json:"certificate"`
	IntermediateCertificates []string `json:"intermediateCertificates,omitempty"`
	PublicKey                string   `json:"publicKey"`
	Purpose                  string   `json:"purpose,omitempty"`
}

// STSTokenResponse represents the data returned from the security token service
type STSTokenResponse struct {
	Token string `json:"token"`
}

// S2SResponse represents a service to service (S2S) response from the security token service
type S2SResponse STSTokenResponse

// NewSTSKeySupplier will create a new key supplier which can be used to register internally managed private key
// with the STS service.
func NewSTSKeySupplier(x509Supplier CertificateSupplier, client httpsigner.Client, endpoint string) (*STSKeySupplier, error) {
	if x509Supplier == nil {
		return nil, ErrInvalidX509Supplier
	}
	if client == nil {
		return nil, ErrInvalidClient
	}
	if endpoint == "" {
		return nil, ErrInvalidEndpoint
	}

	return NewCustomSTSKeySupplier(x509Supplier, client, endpoint, DefaultClientOptions())
}

// NewCustomSTSKeySupplier will create a new key supplier which can be used to register internally managed private key
// with the STS service. Pass validationKeyOptions to override the http.Client used in validationKeySupplier.
func NewCustomSTSKeySupplier(x509Supplier CertificateSupplier, client httpsigner.Client, endpoint string, validationKeyOptions *ClientOptions) (*STSKeySupplier, error) {
	sessionKeySupplier := NewMemorySessionKeySupplier()

	return &STSKeySupplier{
		sessionKeySupplier:      sessionKeySupplier,
		x509Supplier:            x509Supplier,
		client:                  client,
		endpoint:                endpoint,
		validationClientOptions: validationKeyOptions,
	}, nil
}

// Key returns the private key associated with this supplier
func (s *STSKeySupplier) Key(keyid string) (interface{}, error) {
	// Lock immediately, unlock when function exits
	mutexKeyRotation.Lock()
	defer mutexKeyRotation.Unlock()

	// Force rotate the key
	if keyid == KeyIDForceRotate {
		// reset existing token
		return nil, s.rotateKey(true)
	}

	// If we're given the previous keyid, give them the updated one
	if s.previousToken != nil && keyid == s.previousToken.String() {
		return nil, httpsigner.NewKeyRotationError(s.token.String(), s.previousToken.String())
	}

	// Compare given keyid against the current token
	if s.token != nil && keyid != s.token.String() {
		return nil, httpsigner.ErrKeyNotFound
	}

	// Invalid token, refresh
	if !s.IsSecurityTokenValid() {
		return nil, s.rotateKey(false)
	}

	return s.sessionKeySupplier.PrivateKey(), nil
}

// KeyID returns the currently registered S2S token
func (s *STSKeySupplier) KeyID() (string, error) {
	// Lock immediately, unlock when function exits
	mutexKeyRotation.Lock()
	defer mutexKeyRotation.Unlock()

	token, err := s.updateSecurityToken(false)
	if token != nil {
		return token.String(), nil
	}

	return KeyIDForceRotate, err
}

// rotateKey rotates the currently registered token
func (s *STSKeySupplier) rotateKey(forceUpdate bool) error {
	// Update current and previous tokens
	token, err := s.updateSecurityToken(forceUpdate)
	if err != nil {
		return err
	}

	// Set rotated token and old token inside the error
	expiration := httpsigner.NewKeyRotationError(token.String(), s.previousToken.String())

	return expiration
}

// updateSecurityToken updates STSKeySupplier's token from the security token service and saves the previous token.
// Also returns the new STS token.
func (s *STSKeySupplier) updateSecurityToken(forceUpdate bool) (*STSToken, error) {
	if s.IsSecurityTokenValid() && !forceUpdate {
		return s.token, nil
	}

	// If the token was not valid we need to regenerate the session key pair
	s.sessionKeySupplier.RefreshKeys()

	token, err := s.SecurityTokenFromServer()

	// Store the token for next time
	if err == nil {
		if s.token != nil {
			s.previousToken = s.token
		} else {
			// We set a fake STSToken here to prevent a nil pointer when NewKeyRotationError tries to access previousToken.String().
			// The fake STSToken will simply return a blank string.
			s.previousToken = &STSToken{}
		}
		s.token = token
	}

	return token, err
}

// IsSecurityTokenValid returns whether or not the current security token is valid. It will
// check the token expiry as well as the public key in the token
func (s *STSKeySupplier) IsSecurityTokenValid() bool {
	if s.token == nil || s.token.rawToken == "" {
		return false
	}

	// Check token exipry
	err := s.token.ValidFor(time.Now().Add(expirationTimePadding))

	// Check token public key TODO

	return err == nil
}

// SecurityTokenFromServer calls out to the S2S service to generate a new token
func (s *STSKeySupplier) SecurityTokenFromServer() (*STSToken, error) {

	// Build request body
	s2sRequest, err := s.buildS2SRequest()
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(s2sRequest)
	if err != nil {
		return nil, err
	}

	// Build request
	request, err := http.NewRequest("POST", fmt.Sprintf(x509URITemplate, s.endpoint), bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	// Call
	response, err := s.client.Do(request)
	if err != nil {
		return nil, err
	}

	// Return ServiceResponseError if we get a wacky status code back
	if response.StatusCode != 200 {
		err = &ServiceResponseError{response}
		return nil, err
	}

	// Parse response
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var result S2SResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	keyService, err := s.validationKeySupplier(result.Token)
	if err != nil {
		return nil, err
	}

	return newSTSTokenWithConditionalVerification(result.Token, keyService, s.shouldVerifyToken)
}

// validationKeySupplier will return the KeySupplier to use for verifying the STS token.
// The key supplier is an instance of KeyServiceKeySupplier which will fetch the public key
// associated with the STS token which we can use for verification.
func (s *STSKeySupplier) validationKeySupplier(token string) (httpsigner.KeySupplier, error) {

	keyid := formatSTSToken(token)
	supplier, err := httpsigner.NewStaticRSAKeySupplier(s.sessionKeySupplier.PrivateKey(), keyid)
	if err != nil {
		return nil, err
	}
	signer := httpsigner.NewRequestSigner(supplier, httpsigner.AlgorithmRSAPSSSHA256)
	client := NewCustomSigningClient(signer, keyid, s.validationClientOptions)
	keyService := NewKeyServiceKeySupplier(client, s.endpoint)

	return keyService, nil
}

// buildS2SRequest builds the S2SRequest struct using the certificates and the supplied
// public key
func (s *STSKeySupplier) buildS2SRequest() (*S2SRequest, error) {
	// Convert the certificate to PEM format and then extract the base64 encoded value
	certificate, err := s.x509Supplier.CertificateOrError()
	if err != nil {
		return nil, err
	}
	certEncoded := base64EncodeCertificate(certificate)

	// Convert the intermediate certs to PEM format and extract their base64 encoded values
	intermediateDecoded, err := s.x509Supplier.IntermediateOrError()
	if err != nil {
		return nil, err
	}
	intermediateEncoded := make([]string, len(intermediateDecoded))
	for ii, cert := range intermediateDecoded {
		intermediateEncoded[ii] = base64EncodeCertificate(cert)
	}

	publicKeyDER, err := x509.MarshalPKIXPublicKey(s.sessionKeySupplier.PublicKey())
	if err != nil {
		return nil, err
	}

	// Convert the public key to base64
	publicKeyEncoded := base64.StdEncoding.EncodeToString(publicKeyDER)

	// Build request
	request := S2SRequest{
		Certificate: certEncoded,
		PublicKey:   publicKeyEncoded,
		Purpose:     s.tokenPurpose,
	}

	// We do not want to include the slice in the S2S struct if it's empty so that it is omitted from the
	// marshalled JSON body.  The identity service does not like empty value for the intermediate
	// certs.
	if len(intermediateEncoded) > 0 {
		request.IntermediateCertificates = intermediateEncoded
	}

	return &request, nil
}

// formatSTSToken prepends ST$ to the given token
func formatSTSToken(token string) string {
	return STSTokenPrefix + token
}

const (
	//URL for leaf certificate for instance principals
	leafCertificateURL = `http://169.254.169.254/opc/v1/identity/cert.pem`
	//URL for key for instance principals
	leafCertificateKeyURL = `http://169.254.169.254/opc/v1/identity/key.pem`
	//URL for intermediate certificate for instance principals
	intermediateCertificateURL = `http://169.254.169.254/opc/v1/identity/intermediate.pem`
	//Purpose for requesting a service principal token from the instance principals
	servicePrincipalSTSPurpose = "SERVICE_PRINCIPAL"
)

// NewInstanceKeySupplier returns a key supplier that reads certificates from the instance
func NewInstanceKeySupplier(tenancyID, endpoint string, certificateSupplierClient httpsigner.Client) (*STSKeySupplier, error) {
	certificateSupplier, err := NewX509CertificateSupplierFromURLs(
		certificateSupplierClient, tenancyID, leafCertificateURL, leafCertificateKeyURL, nil, intermediateCertificateURL)
	if err != nil {
		return nil, err
	}

	keyID, err := certificateSupplier.KeyID()
	if err != nil {
		return nil, err
	}
	signer := httpsigner.NewRequestSigner(certificateSupplier, httpsigner.AlgorithmRSAPSSSHA256)
	client := NewSigningClient(signer, keyID)
	instanceKeySupplier, err := NewSTSKeySupplier(certificateSupplier, client, endpoint)
	if err != nil {
		return nil, err
	}

	//ID-1362 Instance principals can not be verified at this time due to endpoint not being addressable outside the service enclave
	//Do not verify token
	instanceKeySupplier.shouldVerifyToken = func(token string) bool { return false }

	return instanceKeySupplier, nil
}

// NewServiceInstanceKeySupplier returns a key supplier that reads certificates from the instance
// metadata service for an instance that must make Service-to-Service (S2S) requests.  NOTE: Only
// instances that are part of a service tenancy (or are otherwise authorized by the identity service)
// can use this functionality.
func NewServiceInstanceKeySupplier(tenancyID, endpoint string, certificateSupplierClient httpsigner.Client) (*STSKeySupplier, error) {
	instanceKeySupplier, err := NewInstanceKeySupplier(tenancyID, endpoint, certificateSupplierClient)
	if err != nil {
		return nil, err
	}

	// Instance principal based S2S requests must set SP purpose
	instanceKeySupplier.tokenPurpose = servicePrincipalSTSPurpose

	return instanceKeySupplier, nil
}
