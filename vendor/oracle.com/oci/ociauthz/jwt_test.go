// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"fmt"
	"strconv"

	"oracle.com/oci/httpsigner"
)

// from https://bitbucket.oci.oraclecorp.com/projects/COMMONS/repos/jwt/browse/src/test/java/com/oracle/pic/commons/jwt/TokenConstants.java
const (
	publicKeyPEM = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCfudfKxPMm7J5dnqZLeqrfcph0
39/JeHbPE7jjffUeA0yq/ETnAmI1L5L3RJ9XrFtAPReAli+vZOi8OxpgIBmKgSya
Ktji2kyfbDoDlrPVI9xpzKiWACtH2sCuIjLGS7G21COUZdMophPaDt+woBQdhezm
AT0GNqbjVCeY2Aa+oQIDAQAB
-----END PUBLIC KEY-----`
)

var (
	validSignedToken = GenerateTokenForTest(time.Now().Add(time.Hour * 1))

	testHeaderStr          = base64.StdEncoding.EncodeToString([]byte(`{"kid":"foo","alg":"RS256"}`))
	testHeader             = &Header{KeyID: "foo", Algorithm: "RS256"}
	testHeaderInvalidAlg   = &Header{KeyID: "foo", Algorithm: "bar"}
	testPayload            = []byte(`{"sub":"test"}`)
	testPayloadEncoded     = base64.StdEncoding.EncodeToString(testPayload)
	testSignature          = []byte("AAAA")
	testSignatureEncoded   = base64.StdEncoding.EncodeToString(testSignature)
	testTokenStr           = testHeaderStr + "." + testPayloadEncoded + "." + testSignatureEncoded
	testTokenStrBadPayload = testHeaderStr + ".?.AAAA"
	testTokenStrBadSig     = testHeaderStr + "." + testPayloadEncoded + ".?"
	testBadClaimsPayload   = base64.StdEncoding.EncodeToString([]byte(`?`))
	testTokenStrBadClaims  = testHeaderStr + "." + testBadClaimsPayload + ".AAAA"
	errTest                = errors.New("test")

	testClaims = Claims{
		ClaimSubject: []Claim{{"", ClaimSubject, `test`}},
	}
	testToken = &Token{Header: *testHeader, Claims: testClaims}

	errClaimJSON = json.Unmarshal([]byte(`?`), &Claims{})
)

// TestNewToken tests that the NewToken constructor can return a valid token struct
func TestNewToken(t *testing.T) {
	token, _ := NewToken(validSignedToken, testKeyService)

	var tokenType *Token
	assert.IsType(t, tokenType, token)
}

// TestNewTokenEmptyString tests that the NewToken constructor won't accept an empty string as
// a valid token
func TestNewTokenEmptyString(t *testing.T) {
	token, err := NewToken("", testKeyService)
	assert.Nil(t, token)
	assert.NotNil(t, err)
	assert.Equal(t, ErrInvalidArg, err)
}

func TestParseToken(t *testing.T) {
	ks := &mockKeySupplier{key: testPublicKey}
	as := httpsigner.Algorithms{"RS256": &mockAlgorithm{}}
	testIO := []struct {
		tc       string
		token    string
		expToken *Token
		expErr   error
	}{
		{tc: `should return *Token for valid token`,
			token: testTokenStr, expToken: testToken, expErr: nil},
		{tc: `should return verifyToken errors`,
			token: "one.two", expToken: nil, expErr: ErrJWTMalformed},
		{tc: `should return json unmarshal errors for bad claims json`,
			token: testTokenStrBadClaims, expToken: nil, expErr: errClaimJSON},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			token, err := ParseToken(test.token, ks, as)
			assert.Equal(t, test.expErr, err)
			assert.Equal(t, test.expToken, token)
		})
	}
}

func TestExtractParts(t *testing.T) {
	testIO := []struct {
		tc           string
		token        string
		expHeader    Header
		expBody      []byte
		expSignature []byte
		expError     error
	}{
		{
			tc:           "should return parts with no error",
			token:        testTokenStr,
			expHeader:    *testHeader,
			expBody:      testPayload,
			expSignature: []byte(testSignature),
			expError:     nil,
		},
		{
			tc:           "should return error on bad token",
			token:        "some random data.......",
			expHeader:    Header{},
			expBody:      nil,
			expSignature: nil,
			expError:     ErrJWTMalformed,
		},
		{
			tc:           "should return on bad token less than 3 components",
			token:        "a.b",
			expHeader:    Header{},
			expBody:      nil,
			expSignature: nil,
			expError:     ErrJWTMalformed,
		},
		{
			tc:           "should return header decode error",
			token:        fmt.Sprintf("%s.%s.%s", `"random"`, testPayloadEncoded, testSignatureEncoded),
			expHeader:    Header{},
			expBody:      nil,
			expSignature: nil,
			expError:     base64.CorruptInputError(0),
		},
		{
			tc:           "should return error on bad body in token",
			token:        fmt.Sprintf("%s.%s.%s", testHeaderStr, `"random"`, testSignatureEncoded),
			expHeader:    *testHeader,
			expBody:      []byte{},
			expSignature: nil,
			expError:     base64.CorruptInputError(0),
		},
		{
			tc:           "should return error on bad  signature",
			token:        fmt.Sprintf("%s.%s.%s", testHeaderStr, testPayloadEncoded, "======"),
			expHeader:    *testHeader,
			expBody:      testPayload,
			expSignature: []byte{},
			expError:     base64.CorruptInputError(0),
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			header, body, signature, err := extractParts(test.token)
			assert.Equal(t, test.expError, err)
			assert.Equal(t, test.expHeader, header)
			assert.Equal(t, test.expBody, body)
			assert.Equal(t, test.expSignature, signature)
		})
	}
}
func TestVerifyToken(t *testing.T) {
	validHeader, _, validSignature, _ := extractParts(validSignedToken)
	testIO := []struct {
		tc        string
		token     string
		ks        httpsigner.KeySupplier
		as        httpsigner.AlgorithmSupplier
		header    Header
		signature []byte
		expErr    error
	}{
		{
			tc:        `should return no error for valid token`,
			token:     validSignedToken,
			ks:        &mockKeySupplier{key: "asw"},
			as:        httpsigner.Algorithms{"RS256": &mockAlgorithm{}},
			header:    validHeader,
			signature: validSignature,
			expErr:    nil,
		},
		{
			tc:        `should return ErrJWTMalformed for invalid token`,
			token:     `onetwo`,
			ks:        &mockKeySupplier{key: "asw"},
			as:        httpsigner.Algorithms{"RS256": &mockAlgorithm{}},
			header:    Header{},
			signature: validSignature,
			expErr:    ErrJWTMalformed,
		},
		{
			tc:        `should return KeySupplier errors`,
			token:     testTokenStr,
			ks:        &mockKeySupplier{err: errTest},
			as:        httpsigner.Algorithms{"RS256": &mockAlgorithm{}},
			signature: validSignature,
			header:    Header{},
			expErr:    errTest,
		},
		{
			tc:        `should return AlgorithmSupplier errors`,
			token:     testTokenStr,
			ks:        &mockKeySupplier{key: ""},
			as:        httpsigner.Algorithms{},
			signature: validSignature,
			header:    Header{},
			expErr:    httpsigner.ErrUnsupportedAlgorithm,
		},
		{
			tc:        `should return Algorithm errors`,
			token:     testTokenStr,
			ks:        &mockKeySupplier{key: ""},
			as:        httpsigner.Algorithms{"RS256": &mockAlgorithm{err: errTest}},
			signature: validSignature,
			header:    Header{Algorithm: "RS256"},
			expErr:    errTest,
		},
		{
			tc:        `should return invalid args when key supplier is nil`,
			token:     testTokenStr,
			ks:        nil,
			as:        httpsigner.Algorithms{"RS256": &mockAlgorithm{err: errTest}},
			signature: validSignature,
			header:    Header{Algorithm: "RS256"},
			expErr:    ErrInvalidArg,
		},
		{
			tc:        `should return invalid args when algorithm supplier is nil`,
			token:     testTokenStr,
			ks:        &mockKeySupplier{key: ""},
			as:        nil,
			signature: validSignature,
			header:    Header{},
			expErr:    ErrInvalidArg,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			err := verifyToken(test.header, test.signature, test.token, test.ks, test.as)
			assert.Equal(t, test.expErr, err)
		})
	}
}

func TestTokenValidFor(t *testing.T) {
	testTokenNotBefore := &Token{Claims: Claims{
		ClaimNotBefore: []Claim{{testIssuer, ClaimNotBefore, "10"}},
		ClaimExpires:   []Claim{{testIssuer, ClaimExpires, "20"}},
	}}
	testTokenExpires := &Token{Claims: Claims{
		ClaimExpires: []Claim{{testIssuer, ClaimExpires, "20"}},
	}}
	testTokenEmptyClaims := &Token{Claims: Claims{}}
	testTokenInvalidNotBefore := &Token{Claims: Claims{
		ClaimNotBefore: []Claim{{testIssuer, ClaimNotBefore, "ten"}},
	}}
	testTokenInvalidExpires := &Token{Claims: Claims{
		ClaimExpires: []Claim{{testIssuer, ClaimExpires, "twenty"}},
	}}

	testIO := []struct {
		tc     string
		token  *Token
		clock  int64
		expErr error
	}{
		{tc: `should return nil for token in validity period with NotBefore set`,
			token: testTokenNotBefore, clock: 15, expErr: nil},
		{tc: `should return ErrTokenExpired for expired token with NotBefore set`,
			token: testTokenNotBefore, clock: 30, expErr: ErrTokenExpired},
		{tc: `should return ErrTokenNotValidYet for token with NotBefore set with clock before NotBefore`,
			token: testTokenNotBefore, clock: 5, expErr: ErrTokenNotValidYet},
		{tc: `should return nil for token with no NotBefore set and clock before NotBefore`,
			token: testTokenExpires, clock: 5, expErr: nil},
		{tc: `should return nil for token with no NotBefore in validity period`,
			token: testTokenExpires, clock: 15, expErr: nil},
		{tc: `should return ErrTokenExpired for expired token without NotBefore set`,
			token: testTokenExpires, clock: 30, expErr: ErrTokenExpired},
		{tc: `should return ErrTokenExpired if claims is empty`,
			token: testTokenEmptyClaims, clock: 10, expErr: ErrTokenExpired},
		{tc: `should return NumError if nbf claim is not an int`,
			token:  testTokenInvalidNotBefore,
			clock:  10,
			expErr: &strconv.NumError{Func: "ParseInt", Num: "ten", Err: strconv.ErrSyntax}},
		{tc: `should return NumError if exp claim is not an int`,
			token:  testTokenInvalidExpires,
			clock:  10,
			expErr: &strconv.NumError{Func: "ParseInt", Num: "twenty", Err: strconv.ErrSyntax}},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			clock := time.Unix(test.clock, 0)
			err := test.token.ValidFor(clock)
			assert.Equal(t, test.expErr, err)
		})
	}
}

func TestSignJWT(t *testing.T) {

	keyID := "somekeyid"
	jwk := NewJWKFromPublicKey(keyID, testPublicKey)
	nonHTTP200Response := &http.Response{StatusCode: http.StatusTeapot}

	// Setup httptest server to return the key
	ts := SetupHTTPTestServerToReturnJWK(jwk)
	defer ts.Close()

	testIO := []struct {
		tc            string
		token         *STSToken
		previousToken *STSToken
		keyid         string
		expected      *rsa.PrivateKey
		expectedError error
		client        httpsigner.Client
	}{
		{
			tc:            `should return ServiceResponseError if the response from STS is a non-2xx response`,
			token:         expiredToken,
			previousToken: nil,
			keyid:         expiredToken.String(),
			client:        &MockSigningClient{doResponse: nonHTTP200Response},
			expectedError: &ServiceResponseError{nonHTTP200Response},
		},
		{
			tc:            `should return KeyRotationError if KeyIDForceRotate is supplied as the keyid`,
			token:         goodToken,
			previousToken: nil,
			keyid:         KeyIDForceRotate,
			client:        &MockSigningClient{doResponse: newValidServerResponse()},
			expectedError: &httpsigner.KeyRotationError{},
		},
		{
			tc:            `should return ErrInvalidKey if the private key is invalid`,
			token:         goodToken,
			previousToken: nil,
			keyid:         goodToken.String(),
			client:        &MockSigningClient{doResponse: newValidServerResponse()},
			expectedError: httpsigner.ErrInvalidKey,
		},
		{
			tc:            `should return the signed RPT and no error for valid token`,
			token:         goodToken,
			previousToken: nil,
			keyid:         goodToken.String(),
			client:        &MockSigningClient{doResponse: newValidServerResponse()},
			expectedError: nil,
		},
	}

	for _, test := range testIO {
		t.Run(fmt.Sprintf(test.tc), func(t *testing.T) {
			// should not be able to sign if the private key is invalid
			if test.expectedError == httpsigner.ErrInvalidKey {
				iks := newMockInvalidKeySupplier()
				_, err := signJWT("abc.def", test.keyid, "RS256", iks, OCIJWTSigningAlgorithms)
				assert.Equal(t, test.expectedError, err)

			} else {
				sts, _ := NewSTSKeySupplier(testX509CertificateSupplier, testX509Client, ts.URL)
				sts.client = test.client
				sts.token = test.token
				sts.previousToken = test.previousToken

				sJWT, err := signJWT("abc.def", test.keyid, "RS256", sts, OCIJWTSigningAlgorithms)

				if test.expectedError != nil {
					assert.Equal(t, "", sJWT)

				} else {
					assert.NotNil(t, sJWT)
					assert.Nil(t, err)
				}

				if _, ok := test.expectedError.(*httpsigner.KeyRotationError); !ok {
					assert.Equal(t, test.expectedError, err)
				}
			}

		})
	}
}

func TestGenerateJWT(t *testing.T) {
	signError := errors.New("signing error")
	testIO := []struct {
		tc     string
		keyID  string
		ks     httpsigner.KeySupplier
		as     httpsigner.AlgorithmSupplier
		alg    string
		claims string
		expErr error
	}{
		{
			tc:     `should return signed JWT and no error`,
			ks:     &mockKeySupplier{key: "asw"},
			as:     httpsigner.Algorithms{"RS256": &mockAlgorithm{}},
			alg:    "RS256",
			claims: "testclaimsstring",
			expErr: nil,
		},
		{
			tc:     `should return error from the sign method returned by the mock algorithm`,
			ks:     &mockKeySupplier{key: "asw"},
			as:     httpsigner.Algorithms{"RS256": &mockAlgorithm{err: signError}},
			alg:    "RS256",
			claims: "testclaimsstring",
			expErr: signError,
		},
	}
	for _, test := range testIO {
		t.Run(fmt.Sprintf(test.tc), func(t *testing.T) {
			jwt, err := generateJWT(test.keyID, test.alg, test.claims, test.ks, test.as)
			if test.expErr == nil {
				assert.NotEmpty(t, jwt)
			}
			assert.Equal(t, test.expErr, err)

		})
	}
}
