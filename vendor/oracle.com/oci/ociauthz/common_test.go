// Copyright (c) 2018, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	"oracle.com/oci/httpsigner"
)

var (
	testPrivateKey, testPublicKey = GenerateRsaKeyPair(2048)

	serialNumberLimit = new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _   = rand.Int(rand.Reader, serialNumberLimit)

	testCertTemplate = x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{Organization: []string{"Oracle"}},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour), // valid for an hour
		BasicConstraintsValid: true,
	}

	testCertDER, _ = x509.CreateCertificate(rand.Reader, &testCertTemplate, &testCertTemplate, testPublicKey, testPrivateKey)
	testCert, _    = x509.ParseCertificate(testCertDER)

	endpoint = "http://localhost/v1"
	tenantID = "xyz"

	testX509CertificateSupplier, _ = NewX509CertificateSupplier(tenantID, testCert, []*x509.Certificate{testCert}, testPrivateKey)
	testX509Signer                 = httpsigner.NewRequestSigner(testX509CertificateSupplier, httpsigner.AlgorithmRSAPSSSHA256)
	testX509Client                 = NewSigningClient(testX509Signer, testX509CertificateSupplier.KeyID())

	// Setup a KeySupplier that returns testPublicKey, to replace the KeyServiceKeySupplier in tests
	testKeyService, _ = httpsigner.NewStaticRSAPubKeySupplier(testPublicKey, "asw")

	non200Response = &http.Response{StatusCode: http.StatusTeapot}
)

// Claim, subject values for tests
var (
	testAudience            = "audience"
	testJwtID               = "jwt-id"
	testExp                 = "3234567890"
	testNbf                 = "2234567890"
	testIat                 = "1234567890"
	testFPrint              = "fingerprint"
	testPType               = "user"
	testTTypeLogin          = "login"
	testTTypeSAML           = "saml"
	testPSType              = "fed"
	testFederatedUserGroups = "QXZlbmdlcnMK;R3VhcmRpYW5zIG9mIHRoZSBHYWxheHkK"
)

type MockSigningClient struct {
	signRequestResponse *http.Request
	signRequestError    error

	// DoRequest stores the last request parameter sent to the Do func
	// for verifying in tests
	DoRequest  *http.Request
	doResponse *http.Response
	doError    error
}

func (m *MockSigningClient) SignRequest(request *http.Request) (*http.Request, error) {
	return m.signRequestResponse, m.signRequestError
}

func (m *MockSigningClient) Do(request *http.Request) (*http.Response, error) {
	m.DoRequest = request
	return m.doResponse, m.doError
}

// OKResponse returns a 200 OK Response with body from given bytes
func OKResponse(b []byte) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewBuffer(b)),
	}
}

// GenerateTokenForTest generates a new token with the specified expiry time for use in unit tests
func GenerateTokenForTest(expires time.Time) string {
	header := Header{
		KeyID:     "asw",
		Algorithm: "RS256",
	}

	claims := []Claim{
		{testIssuer, ClaimExpires, strconv.FormatInt(expires.Unix(), 10)},
	}

	return GenerateToken(header, claims, testPrivateKey)
}

// SetupHTTPTestServerToReturnJWK sets up a httptest server to return a JWK for use in unit tests
func SetupHTTPTestServerToReturnJWK(jwk *JWK) *httptest.Server {
	// Convert the JWK to a JSON string
	jwkJSON, err := json.Marshal(jwk)
	if err != nil {
		panic(err)
	}

	return SetupHTTPTestServer(string(jwkJSON))
}

// SetupHTTPTestServer sets up an httptest server to echo back a static string
func SetupHTTPTestServer(response string) *httptest.Server {

	// Setup handler to respond to the http request
	echoHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, response)
	}

	// Setup httptest server
	return httptest.NewServer(http.HandlerFunc(echoHandler))
}

// SetupHTTPNotFoundTestServer constructs a test server that always responds with 404
func SetupHTTPNotFoundTestServer() *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}))
}

// EncodeTokenPart returns the base64 URL encoded version of the supplied token part
func EncodeTokenPart(part []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(part), "=")
}

// GenerateToken generates a JWT token using the supplied header, claims and private key
func GenerateToken(header Header, claims []Claim, privateKey *rsa.PrivateKey) string {
	headerJSON, _ := json.Marshal(header)
	claimsMap := map[string]string{}
	for _, c := range claims {
		claimsMap[c.Key] = c.Value
	}
	claimsJSON, _ := json.Marshal(claimsMap)

	stringToSign := strings.Join([]string{EncodeTokenPart(headerJSON), EncodeTokenPart(claimsJSON)}, ".")

	h := sha256.New()
	h.Write([]byte(stringToSign))

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, h.Sum(nil))
	if err != nil {
		panic(err)
	}

	return strings.Join([]string{stringToSign, EncodeTokenPart(signature)}, ".")
}

type mockAlgorithm struct {
	VerifyCalled bool
	err          error
}

func (ma *mockAlgorithm) Name() string {
	return "mock"
}

func (ma *mockAlgorithm) Sign(message []byte, key interface{}) (signature []byte, err error) {
	return nil, ma.err
}

func (ma *mockAlgorithm) Verify(message, signature []byte, key interface{}) error {
	ma.VerifyCalled = true
	return ma.err
}

type mockKeySupplier struct {
	KeyCalled bool
	key       interface{}
	err       error
}

func (mks *mockKeySupplier) Key(kid string) (interface{}, error) {
	mks.KeyCalled = true
	return mks.key, mks.err
}

type mockInvalidKeySupplier struct{}

func (iks *mockInvalidKeySupplier) Key(kid string) (interface{}, error) {
	return rsa.PrivateKey{}, nil
}

func newMockInvalidKeySupplier() *mockInvalidKeySupplier {
	return &mockInvalidKeySupplier{}
}
