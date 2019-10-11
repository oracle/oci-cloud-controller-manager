// Copyright (c) 2019, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"oracle.com/oci/httpsigner"
)

var (
	testRSAKey         = generateKey()
	testRSAPubKey      = &testRSAKey.PublicKey
	testSupplier, _    = httpsigner.NewStaticRSAKeySupplier(testRSAKey, testKeyID)
	testPubSupplier, _ = httpsigner.NewStaticRSAPubKeySupplier(testRSAPubKey, testKeyID)
	testRPTClaims      = &ResourcePrincipalTokenClaimValues{
		Issuer:        "testissuer",
		TenantID:      "testtenantID",
		ResourceID:    "testresourceID",
		CompartmentID: "testcompatmentID",
		TokenType:     "testVersion",
		ResourceType:  "testResourceType",
		ResourceTag:   "testRestag",
		PublicKey:     "testpk",
	}
)

func generateKey() *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	return key
}

func TestNewResourcePrincipalTokenProvider(t *testing.T) {
	// should return a ResourcePrincipalTokenProvider with the arguments in the appropriate members
	testrptp := NewResourcePrincipalTokenProvider(testSupplier, OCIJWTSigningAlgorithms)
	var ptype *ResourcePrincipalTokenProvider
	assert.IsType(t, ptype, testrptp)
	assert.Equal(t, testSupplier, testrptp.keySupplier)
	assert.IsType(t, OCIJWTSigningAlgorithms, testrptp.algorithmSupplier)
}

func TestBuildRPTPubKeySupplier(t *testing.T) {
	// should build the RPTPublicKeySupplier
	pks, err := buildRPTPublicKeySupplier(RPTFixedKeyID, testKeyID, testPubSupplier)
	assert.Nil(t, err)

	// should return the key corresponding to the supplier with which the RPTPublicKeySupplier was built on passing the fixedKeyID
	pk1, _ := pks.Key(RPTFixedKeyID)
	pk2, _ := testPubSupplier.Key(testKeyID)
	assert.Equal(t, pk1, pk2)

	// should return ErrKeyNotFound on passing an invalid key for the supplier
	_, err = buildRPTPublicKeySupplier(RPTFixedKeyID, "testKeyID", testSupplier)
	assert.Equal(t, httpsigner.ErrKeyNotFound, err)
}

func TestBuildRPTSigningKeySupplier(t *testing.T) {
	// should build the RPTSigningKeySupplier
	pks, err := buildRPTSigningKeySupplier(RPTFixedKeyID, testKeyID, testSupplier)
	assert.Nil(t, err)

	// should return the key corresponding to the supplier with which the RPTSigningKeySupplier was built on passing the fixedKeyID
	pk1, _ := pks.Key(RPTFixedKeyID)
	pk2, _ := testSupplier.Key(testKeyID)
	assert.Equal(t, pk1, pk2)

	// should return ErrKeyNotFound on passing an invalid key for the supplier
	_, err = buildRPTSigningKeySupplier(RPTFixedKeyID, "testKeyID", testSupplier)
	assert.Equal(t, httpsigner.ErrKeyNotFound, err)

	//setup server and key supplier to return required key
	keyID := "somekeyid"
	jwk := NewJWKFromPublicKey(keyID, testPublicKey)
	ts := SetupHTTPTestServerToReturnJWK(jwk)
	defer ts.Close()
	sts, _ := NewSTSKeySupplier(testX509CertificateSupplier, testX509Client, ts.URL)
	sts.client = &MockSigningClient{doResponse: newValidServerResponse()}
	sts.token = goodToken
	sts.previousToken = nil

	// should return KeyRotationError when KeyIDForceRotate is used
	_, err = buildRPTSigningKeySupplier(RPTFixedKeyID, KeyIDForceRotate, sts)
	assert.IsType(t, &httpsigner.KeyRotationError{}, err)
}

func TestParseRPT(t *testing.T) {
	testrptp := NewResourcePrincipalTokenProvider(testSupplier, OCIJWTSigningAlgorithms)

	// generate token and test there was no error
	rawToken, err := testrptp.GenerateRPT(testKeyID, "RS256", testRPTClaims)
	assert.NotEmpty(t, rawToken)
	assert.Equal(t, nil, err)
	var tokenType *Token

	//test Parse RPT
	testIO := []struct {
		tc       string
		token    string
		expToken *Token
		expErr   error
		keyid    string
	}{
		{tc: `should return ErrJWTMalformed on passing a malformed JWT`,
			token: "one.two", keyid: testKeyID, expToken: nil, expErr: ErrJWTMalformed},
		{tc: `should return ErrKeyNotFound error on passing empty key`,
			token: testTokenStr, keyid: "", expToken: nil, expErr: httpsigner.ErrKeyNotFound},
		{tc: `should parse token successfully`,
			token: rawToken, keyid: testKeyID, expToken: nil, expErr: nil},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			token, err := parseRPT(test.token, test.keyid, OCIJWTSigningAlgorithms, testPubSupplier)
			if test.expErr == nil {
				assert.IsType(t, tokenType, token)
				assert.Equal(t, test.expErr, err)
			} else {
				assert.Equal(t, test.expErr, err)
				assert.Equal(t, test.expToken, token)
			}

		})
	}

}

func TestGenerateRPT(t *testing.T) {
	testIO := []struct {
		tc       string
		rptcl    *ResourcePrincipalTokenClaimValues
		keyid    string
		supplier httpsigner.KeySupplier
		alg      string
		expToken string
		expError error
	}{
		{
			tc: `should return ErrInvalidSigningAlgorithm for unsupported alg name`,
			rptcl: &ResourcePrincipalTokenClaimValues{Issuer: "testissuer",
				TenantID:      "testtenantID",
				ResourceID:    "testresourceID",
				CompartmentID: "testcompatmentID",
				TokenType:     "testVersion",
				ResourceType:  "testResourceType",
				ResourceTag:   "testRestag",
				PublicKey:     "testpk"},
			keyid:    testKeyID,
			supplier: testSupplier,
			alg:      "",
			expToken: "",
			expError: httpsigner.ErrUnsupportedAlgorithm,
		},
		{
			tc: `should return ErrInvalidClaimPublicKey on empty public key`,
			rptcl: &ResourcePrincipalTokenClaimValues{Issuer: "testissuer",
				TenantID:      "testtenantID",
				ResourceID:    "testresourceID",
				CompartmentID: "testcompatmentID",
				TokenType:     "testVersion",
				ResourceType:  "testResourceType",
				ResourceTag:   "testRestag",
				PublicKey:     ""},
			keyid:    testKeyID,
			supplier: testSupplier,
			alg:      "RS256",
			expToken: "",
			expError: ErrInvalidClaimPublicKey,
		},
		{
			tc:       `should return ErrKeyNotFound with an invalid Key ID`,
			rptcl:    testRPTClaims,
			keyid:    "testKeyID",
			supplier: testSupplier,
			alg:      "RS256",
			expError: httpsigner.ErrKeyNotFound,
		},
		{
			tc:       `should generate RPT string successfully`,
			rptcl:    testRPTClaims,
			keyid:    testKeyID,
			supplier: testSupplier,
			alg:      "RS256",
			expError: nil,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			rptp := NewResourcePrincipalTokenProvider(test.supplier, OCIJWTSigningAlgorithms)
			rawToken, err := rptp.GenerateRPT(test.keyid, test.alg, test.rptcl)
			if test.expError == nil {
				assert.Equal(t, err, test.expError)
			} else {
				assert.Equal(t, rawToken, test.expToken)
				assert.Equal(t, err, test.expError)
			}
		})
	}
}

func TestEncodeClaims(t *testing.T) {
	testExpiry := 1562950591
	testIO := []struct {
		tc        string
		rptcl     *ResourcePrincipalTokenClaimValues
		expClaims string
		expError  error
		testType  string
	}{
		{
			tc: `should return ErrInvalidClaimIssuer on empty claims`,
			rptcl: &ResourcePrincipalTokenClaimValues{Issuer: "",
				TenantID:      "testtenantID",
				ResourceID:    "testresourceID",
				CompartmentID: "testcompatmentID",
				TokenType:     "testVersion",
				ResourceType:  "testResourceType",
				ResourceTag:   "testRestag",
				PublicKey:     "testpk"},
			expClaims: "",
			expError:  ErrInvalidClaimIssuer,
			testType:  "claim",
		},
		{
			tc: `should return ErrInvalidClaimTokenType on empty token type`,
			rptcl: &ResourcePrincipalTokenClaimValues{Issuer: "testissuer",
				TenantID:      "testtenantID",
				ResourceID:    "testresourceID",
				CompartmentID: "testcompatmentID",
				TokenType:     "",
				ResourceType:  "testResourceType",
				ResourceTag:   "testRestag",
				PublicKey:     "testpk"},
			expClaims: "",
			expError:  ErrInvalidClaimTokenType,
			testType:  "claim",
		},
		{
			tc: `should return ErrInvalidClaimResourceType on empty resource type`,
			rptcl: &ResourcePrincipalTokenClaimValues{Issuer: "testissuer",
				TenantID:      "testtenantID",
				ResourceID:    "testresourceID",
				CompartmentID: "testcompatmentID",
				TokenType:     "testVersion",
				ResourceType:  "",
				ResourceTag:   "testRestag",
				PublicKey:     "testpk"},
			expClaims: "",
			expError:  ErrInvalidClaimResourceType,
			testType:  "claim",
		},
		{
			tc: `should return ErrInvalidClaimTenantID on empty tenant ID`,
			rptcl: &ResourcePrincipalTokenClaimValues{Issuer: "testissuer",
				TenantID:      "",
				ResourceID:    "testresourceID",
				CompartmentID: "testcompatmentID",
				TokenType:     "testVersion",
				ResourceType:  "testResourceType",
				ResourceTag:   "testRestag",
				PublicKey:     "testpk"},
			expClaims: "",
			expError:  ErrInvalidClaimTenantID,
			testType:  "claim",
		},
		{
			tc: `should return ErrInvalidClaimCompartmentID on empty compartment ID`,
			rptcl: &ResourcePrincipalTokenClaimValues{Issuer: "testissuer",
				TenantID:      "testtenantID",
				ResourceID:    "testresourceID",
				CompartmentID: "",
				TokenType:     "testVersion",
				ResourceType:  "testResourceType",
				ResourceTag:   "testRestag",
				PublicKey:     "testpk"},
			expClaims: "",
			expError:  ErrInvalidClaimCompartmentID,
			testType:  "claim",
		},
		{
			tc: `should return ErrInvalidClaimResourceID on empty resource ID`,
			rptcl: &ResourcePrincipalTokenClaimValues{Issuer: "testissuer",
				TenantID:      "testtenantID",
				ResourceID:    "",
				CompartmentID: "testcompatmentID",
				TokenType:     "testVersion",
				ResourceType:  "testResourceType",
				ResourceTag:   "testRestag",
				PublicKey:     "testpk"},
			expClaims: "",
			expError:  ErrInvalidClaimResourceID,
			testType:  "claim",
		},
		{
			tc: `should return ErrInvalidClaimResourceTag on empty resource tag`,
			rptcl: &ResourcePrincipalTokenClaimValues{Issuer: "testissuer",
				TenantID:      "testtenantID",
				ResourceID:    "testresourceID",
				CompartmentID: "testcompatmentID",
				TokenType:     "testVersion",
				ResourceType:  "testResourceType",
				ResourceTag:   "",
				PublicKey:     "testpk"},
			expClaims: "",
			expError:  ErrInvalidClaimResourceTag,
			testType:  "claim",
		},
		{
			tc: `should return ErrInvalidClaimPublicKey on empty public key`,
			rptcl: &ResourcePrincipalTokenClaimValues{Issuer: "testissuer",
				TenantID:      "testtenantID",
				ResourceID:    "testresourceID",
				CompartmentID: "testcompatmentID",
				TokenType:     "testVersion",
				ResourceType:  "testResourceType",
				ResourceTag:   "testRestag",
				PublicKey:     ""},
			expClaims: "",
			expError:  ErrInvalidClaimPublicKey,
			testType:  "claim",
		},
		{
			tc:       `should use default expiry if expiry claim not set`,
			rptcl:    testRPTClaims,
			expError: nil,
			testType: "expirydefault",
		},
		{
			tc: `should use set expiry if expiry claim set`,
			rptcl: &ResourcePrincipalTokenClaimValues{Issuer: "testissuer",
				TenantID:      "testtenantID",
				ResourceID:    "testresourceID",
				CompartmentID: "testcompatmentID",
				TokenType:     "testVersion",
				ResourceType:  "testResourceType",
				ResourceTag:   "testRestag",
				PublicKey:     "testPublicKey",
				Expiry:        testExpiry},
			expError: nil,
			testType: "expiryset",
		},
	}
	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {

			exp := int(time.Now().Add(tokenDefaultLifeTime).Unix())
			c, err := test.rptcl.EncodeClaims()

			if test.testType == "expirydefault" {
				var cl *ResourcePrincipalTokenClaimValues
				_ = json.Unmarshal([]byte(c), &cl)
				assert.WithinDuration(t, time.Unix(int64(exp), 0), time.Unix(int64(cl.Expiry), 0), 5*time.Second)
				assert.Nil(t, err)
			} else if test.testType == "expiryset" {
				var cl *ResourcePrincipalTokenClaimValues
				_ = json.Unmarshal([]byte(c), &cl)
				assert.Equal(t, testExpiry, cl.Expiry)
			} else {
				assert.Equal(t, err, test.expError)
				assert.Equal(t, test.expClaims, c)
			}
		})
	}
}
