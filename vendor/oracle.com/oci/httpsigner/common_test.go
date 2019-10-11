// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package httpsigner

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/http"
)

const (
	// testPublicKey is the public key from https://bitbucket.oci.oraclecorp.com/projects/SDK/repos/signing-examples/browse/raw/example
	testPublicKey string = `
	-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDCFENGw33yGihy92pDjZQhl0C3
6rPJj+CvfSC8+q28hxA161QFNUd13wuCTUcq0Qd2qsBe/2hFyc2DCJJg0h1L78+6
Z4UMR7EOcpfdUE9Hf3m/hs+FUR45uBJeDK1HSFHD8bHKD6kv8FPGfJTotc+2xjJw
oYi+1hqp1fIekaxsyQIDAQAB
-----END PUBLIC KEY-----
`
	// testPrivateKey is the private key from https://bitbucket.oci.oraclecorp.com/projects/SDK/repos/signing-examples/browse/raw/example
	testPrivateKey string = `
-----BEGIN RSA PRIVATE KEY-----
MIICXgIBAAKBgQDCFENGw33yGihy92pDjZQhl0C36rPJj+CvfSC8+q28hxA161QF
NUd13wuCTUcq0Qd2qsBe/2hFyc2DCJJg0h1L78+6Z4UMR7EOcpfdUE9Hf3m/hs+F
UR45uBJeDK1HSFHD8bHKD6kv8FPGfJTotc+2xjJwoYi+1hqp1fIekaxsyQIDAQAB
AoGBAJR8ZkCUvx5kzv+utdl7T5MnordT1TvoXXJGXK7ZZ+UuvMNUCdN2QPc4sBiA
QWvLw1cSKt5DsKZ8UETpYPy8pPYnnDEz2dDYiaew9+xEpubyeW2oH4Zx71wqBtOK
kqwrXa/pzdpiucRRjk6vE6YY7EBBs/g7uanVpGibOVAEsqH1AkEA7DkjVH28WDUg
f1nqvfn2Kj6CT7nIcE3jGJsZZ7zlZmBmHFDONMLUrXR/Zm3pR5m0tCmBqa5RK95u
412jt1dPIwJBANJT3v8pnkth48bQo/fKel6uEYyboRtA5/uHuHkZ6FQF7OUkGogc
mSJluOdc5t6hI1VsLn0QZEjQZMEOWr+wKSMCQQCC4kXJEsHAve77oP6HtG/IiEn7
kpyUXRNvFsDE0czpJJBvL/aRFUJxuRK91jhjC68sA7NsKMGg5OXb5I5Jj36xAkEA
gIT7aFOYBFwGgQAQkWNKLvySgKbAZRTeLBacpHMuQdl1DfdntvAyqpAZ0lY0RKmW
G6aFKaqQfOXKCyWoUiVknQJAXrlgySFci/2ueKlIE1QqIiLSZ8V8OlpFLRnb1pzI
7U1yQXnTAEFYM560yJlzUpOb1V4cScGd365tiSMvxLOvTA==
-----END RSA PRIVATE KEY-----
`
	// testKeyID is the keyid from https://bitbucket.oci.oraclecorp.com/projects/SDK/repos/signing-examples/browse/raw/example
	testKeyID = "ocid1.tenancy.oc1..aaaaaaaaba3pv6wkcr4jqae5f15p2b2m2yt2j6rx32uzr4h25vqstifsfdsq/ocid1.user.oc1..aaaaaaaat5nvwcna5j6aqzjcaty5eqbb6qt2jvpkanghtgdaqedqw3rynjq/20:3b:97:13:55:1c:5b:0d:d3:37:d8:50:4e:c5:3a:34"
	// testBody is the body from from https://bitbucket.oci.oraclecorp.com/projects/SDK/repos/signing-examples/browse/raw/example
	testBody string = `{
    "compartmentId": "ocid1.compartment.oc1..aaaaaaaam3we6vgnherjq5q2idnccdflvjsnog7mlr6rtdb25gilchfeyjxa",
    "instanceId": "ocid1.instance.oc1.phx.abuw4ljrlsfiqw6vzzxb43vyypt4pkodawglp3wqxjqofakrwvou52gb6s5a",
    "volumeId": "ocid1.volume.oc1.phx.abyhqljrgvttnlx73nmrwfaux7kcvzfs3s66izvxf2h4lgvyndsdsnoiwr5q"
}`
	// algEcho is the name of the 'echo-algorithm' used for testing
	algEcho string = "echo-algorithm"
)

var (
	// mockAlgorithm implements the Algorithm interface and returns errMock on Sign()
	mockAlgorithm = &mockAlg{}
	// errMock is a mock error which is returned by mockReader.Read()
	errMock = errors.New("Mock Read Error")
	// mockReader implements the io.Reader interface and returns errMock on Read()
	mockReader = &mockRdr{}
)

type mockAlg struct {
	SignCalled   bool
	VerifyCalled bool
	verifyErr    error
}

// Name returns the name of the mock algorithm
func (a mockAlg) Name() string {
	return "mock-algorithm"
}

// Sign returns errMock and record that Sign() was called for later verification.
func (a *mockAlg) Sign(message []byte, key interface{}) (sig []byte, err error) {
	a.SignCalled = true
	err = errMock
	return
}

func (a *mockAlg) Verify(message, signature []byte, key interface{}) error {
	a.VerifyCalled = true
	return a.verifyErr
}

type mockRdr struct {
}

func (r *mockRdr) Read(p []byte) (n int, err error) {
	err = errMock
	return
}

// MockKeySupplier always returns the key "doublesecret", but also records that Key() was called so that test code can
// validate that it was used (or not) when expected.
type MockKeySupplier struct {
	KeyCalled bool
	key       interface{}
	err       error
}

func (mks *MockKeySupplier) Key(keyID string) (interface{}, error) {
	mks.KeyCalled = true
	return mks.key, mks.err
}

func (mks *MockKeySupplier) Reset() {
	mks.KeyCalled = false
}

// MockRequestSigner doesn't actually sign requests, but keeps track of when SignRequest is called.
type MockRequestSigner struct {
	signingError error

	// mock state
	SignRequestCalled bool
	ProfferedRequest  *http.Request
	ProfferedKey      string
	ProfferedHeaders  []string
}

// SignRequest will return the proffered http.Request unless signingError is not nil in which case it will return the
// value of signingError.
func (mrs *MockRequestSigner) SignRequest(r *http.Request, k string, h []string) (*http.Request, error) {

	// save state of call
	mrs.SignRequestCalled = true
	mrs.ProfferedRequest = r
	mrs.ProfferedKey = k
	mrs.ProfferedHeaders = h

	// mock response
	if mrs.signingError != nil {
		return nil, mrs.signingError
	}
	return r, nil
}

// Reset clears state so the mock can be reused
func (mrs *MockRequestSigner) Reset() {
	mrs.SignRequestCalled = false
	mrs.ProfferedRequest = nil
	mrs.ProfferedKey = ""
	mrs.ProfferedHeaders = nil
}

// MockHTTPClient overrides Do so that no action happens, but records the call and its arguments
type MockHTTPClient struct {
	response  *http.Response
	respError error

	DoCalled         bool
	ProfferedRequest *http.Request
	*http.Client
}

// Do saves the values passed in and response with configured values
func (mhc *MockHTTPClient) Do(r *http.Request) (*http.Response, error) {
	mhc.DoCalled = true
	mhc.ProfferedRequest = r
	return mhc.response, mhc.respError
}

// Reset clears state so the mock can be reused
func (mhc *MockHTTPClient) Reset() {
	mhc.DoCalled = false
	mhc.ProfferedRequest = nil
}

// NewtestAlgorithmRSAPSSSHA256 returns a new instance of the 'rsa-pss-sha256' signing algorithm
// which uses a non-random source for its salt for consistent test results
func NewtestAlgorithmRSAPSSSHA256() Algorithm {
	nonRandomReader := bytes.NewReader(make([]byte, 160))
	return algorithmRSAPSSSHA256{
		randReader: nonRandomReader,
		pssOptions: rsa.PSSOptions{SaltLength: 32}}
}

// NewPKCS1RSAPrivateKeyFromPEM creates a *rsa.PrivateKey
// from a PKCS1 unencrypted private key PEM string
// It returns an ErrInvalidKey error if the key cannot be parsed
func NewPKCS1RSAPrivateKeyFromPEM(k string) (*rsa.PrivateKey, error) {
	block, err := newPrivateKeyBlockFromPEM(k)
	if err != nil {
		return nil, err
	}
	privKey, _ := x509.ParsePKCS1PrivateKey(block.Bytes)
	if privKey == nil {
		return nil, ErrInvalidKey
	}
	return privKey, nil
}

// NewPKCS8RSAPrivateKeyFromPEM creates a *rsa.PrivateKey
// from a PKCS8 unencrypted private key PEM string
// It returns an ErrInvalidKey error if the key cannot be parsed
func NewPKCS8RSAPrivateKeyFromPEM(k string) (*rsa.PrivateKey, error) {
	block, err := newPrivateKeyBlockFromPEM(k)
	if err != nil {
		return nil, err
	}
	privKey, _ := x509.ParsePKCS8PrivateKey(block.Bytes)
	if privKey == nil {
		return nil, ErrInvalidKey
	}
	return privKey.(*rsa.PrivateKey), nil
}

// newPrivateKeyBlockFromPEM creates a private key Block
// from an unecrpyted PEM string
// It returns an ErrInvalidKey error if the key cannot be parsed
func newPrivateKeyBlockFromPEM(k string) (*pem.Block, error) {
	if k == "" {
		return nil, ErrInvalidKey
	}
	block, _ := pem.Decode([]byte(k))
	if block == nil {
		return nil, ErrInvalidKey
	}
	return block, nil
}
