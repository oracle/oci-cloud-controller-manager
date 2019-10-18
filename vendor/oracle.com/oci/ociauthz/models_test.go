// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http/httptest"
	"strconv"
	"testing"

	"oracle.com/oci/httpsigner"

	"github.com/stretchr/testify/assert"
)

var (
	testIssuer          = "issuer"
	testSubject         = "test-subject"
	testSubjectClaim    = Claim{testIssuer, ClaimSubject, testSubject}
	testPrincipalClaims = Claims{
		ClaimIssuer:  []Claim{{testIssuer, ClaimIssuer, testIssuer}},
		ClaimSubject: []Claim{{testIssuer, ClaimSubject, testSubject}},
		ClaimExpires: []Claim{{testIssuer, ClaimExpires, "10"}},
	}
	testMultipleClaimTypes = Claims{
		ClaimOrgUnit: []Claim{
			{testIssuer, ClaimOrgUnit, "first-value"},
			{testIssuer, ClaimOrgUnit, "second-value"},
		},
	}
	testAuthzHdr = `Signature version="1", headers="date (request-target)", keyid="ST$key", algorithm="alg", signature="sig"`
)

// TestSimpleSubject tests the principal model's ID() and tenantID() methods
func TestSimpleSubject(t *testing.T) {
	testIO := []struct {
		tc          string
		principalID string
		expectedID  string
		tenantID    string
	}{
		{tc: `should create a principal from empty string with the corresponding undetected-subject ID`,
			principalID: "", tenantID: "", expectedID: subjectUndetected},
		{tc: `should create a principal from valid inputs`,
			principalID: "id", tenantID: "tid", expectedID: "id"},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			p := NewPrincipal(test.principalID, test.tenantID)
			assert.Equal(t, p.ID(), test.expectedID)
			assert.Equal(t, p.TenantID(), test.tenantID)
			assert.Nil(t, p.Claims())
		})
	}
}

func TestNewPrincipalFromToken(t *testing.T) {
	testIO := []struct {
		tc            string
		token         *Token
		delegate      *Principal
		expectedID    string
		expectedError error
	}{
		{
			tc:            `should create a principal from a valid jwt token with empty Claims with the corresponding undetected-subject ID`,
			token:         &Token{},
			expectedError: nil,
			expectedID:    subjectUndetected,
			delegate:      nil,
		},
		{tc: `should create a principal from a valid jwt token`,
			token: &Token{Claims: Claims{
				ClaimSubject:     []Claim{{testIssuer, ClaimSubject, "subject-id"}},
				ClaimTenant:      []Claim{{testIssuer, ClaimTenant, "tenant-id"}},
				ClaimServiceName: []Claim{{testIssuer, ClaimServiceName, "svc-name"}},
			}},
			expectedError: nil,
			expectedID:    "subject-id",
			delegate:      nil,
		},
		{
			tc: `should create a principal from a valid jwt token with empty Claims except s2s ttype with the corresponding undetected-subject ID`,
			token: &Token{Claims: Claims{
				ClaimTokenType: []Claim{{testIssuer, ClaimTokenType, "s2s"}},
			}},
			expectedError: nil,
			expectedID:    subjectUndetected,
			delegate:      nil,
		},
		{
			tc: `should create a principal from a valid jwt token with ttype obo with delegate`,
			token: &Token{Claims: Claims{
				ClaimSubject:   []Claim{{testIssuer, ClaimSubject, "subject-id"}},
				ClaimTenant:    []Claim{{testIssuer, ClaimTenant, "tenant-id"}},
				ClaimTokenType: []Claim{{testIssuer, ClaimTokenType, "obo"}},
			}},
			expectedError: nil,
			expectedID:    "subject-id",
			delegate:      &Principal{},
		},
		{
			tc:            `should return a ErrInvalidToken error for nil token`,
			token:         nil,
			expectedError: ErrInvalidToken,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			p, e := NewPrincipalFromToken(test.token, test.delegate)
			if test.expectedError == nil {
				assert.Nil(t, e)
				assert.Equal(t, p.ID(), test.expectedID)
				assert.Equal(t, p.TenantID(), test.token.Claims.GetString(ClaimTenant))
				assert.Equal(t, test.delegate, p.Delegate())
				assert.Equal(t, p.Claims(), test.token.Claims)
			} else {
				assert.Nil(t, p)
				assert.Equal(t, e, test.expectedError)
			}
		})
	}
}

func TestNewPrincipalFromTokenAndRequest(t *testing.T) {
	testIO := []struct {
		tc             string
		token          *Token
		headers        map[string]string
		expectedID     string
		expectedError  error
		expectedClaims Claims
	}{
		{tc: `should return error propagated from NewPrincipalFromToken`,
			token:         nil,
			headers:       map[string]string{},
			expectedError: ErrInvalidToken,
		},
		{tc: `should return error if no authorization header`,
			token: &Token{Claims: Claims{
				ClaimSubject: []Claim{{testIssuer, ClaimSubject, "subject-id"}},
				ClaimTenant:  []Claim{{testIssuer, ClaimTenant, "tenant-id"}},
			}},
			headers:       map[string]string{},
			expectedError: httpsigner.ErrMissingAuthzHeader,
		},
		{tc: `should return principal with header claims given a request with headers`,
			token: &Token{Claims: Claims{
				ClaimSubject: []Claim{{testIssuer, ClaimSubject, "subject-id"}},
				ClaimTenant:  []Claim{{testIssuer, ClaimTenant, "tenant-id"}},
			}},
			headers: map[string]string{
				"authorization": testAuthzHdr,
				"date":          "Thu, 21 June 1582",
				"request-id":    "aaaa.BBBB.cccc",
			},
			expectedError: nil,
			expectedID:    "subject-id",
			expectedClaims: Claims{
				ClaimSubject:         []Claim{{testIssuer, ClaimSubject, "subject-id"}},
				ClaimTenant:          []Claim{{testIssuer, ClaimTenant, "tenant-id"}},
				"h_authorization":    []Claim{{HdrClaimIssuer, "h_authorization", testAuthzHdr}},
				"h_date":             []Claim{{HdrClaimIssuer, "h_date", "Thu, 21 June 1582"}},
				"h_(request-target)": []Claim{{HdrClaimIssuer, "h_(request-target)", "get /"}},
			},
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			request := httptest.NewRequest("GET", "https://localhost", nil)
			for key, value := range test.headers {
				request.Header.Add(key, value)
			}
			p, e := NewPrincipalFromTokenAndRequest(test.token, nil, request)
			if test.expectedError == nil {
				assert.Nil(t, e)
				assert.Equal(t, test.expectedID, p.ID())
				assert.Equal(t, test.token.Claims.GetString(ClaimTenant), p.TenantID())
				assert.Equal(t, test.expectedClaims, p.Claims())
			} else {
				assert.Nil(t, p)
				assert.Equal(t, e, test.expectedError)
			}
		})
	}
}

func TestPrincipalAddClaim(t *testing.T) {
	principalNoClaims := NewPrincipal(testSubjectID, testTenantID)
	principalClaims, _ := NewPrincipalFromToken(testToken, nil)

	testIO := []struct {
		tc        string
		principal *Principal
		claim     Claim
		expected  Claims
	}{
		{tc: `should add claim to principal with no claims`,
			principal: principalNoClaims,
			claim:     Claim{testIssuer, "claim-key", "value"},
			expected: Claims{
				"claim-key": []Claim{{testIssuer, "claim-key", "value"}},
			},
		},
		{tc: `should add claim to principal with claims`,
			principal: principalClaims,
			claim:     Claim{testIssuer, "claim-key", "value"},
			expected: Claims{
				"claim-key":  []Claim{{testIssuer, "claim-key", "value"}},
				ClaimSubject: []Claim{{"", ClaimSubject, `test`}},
			},
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			test.principal.AddClaim(test.claim)
			assert.Equal(t, test.expected, test.principal.Claims())
		})
	}
}

func TestPrincipalType(t *testing.T) {
	testIO := []struct {
		tc        string
		principal *Principal
		expected  string
	}{
		{tc: `should return PrincipalTypeUser if Claims is nil`,
			principal: &Principal{},
			expected:  PrincipalTypeUser,
		},
		{tc: `should return PrincipalTypeUser if Claims.PrincipalType is empty`,
			principal: &Principal{claims: Claims{}},
			expected:  PrincipalTypeUser,
		},
		{tc: `should return PrincipalTypeInstance set in Claims.PrincipalType`,
			principal: &Principal{claims: Claims{
				ClaimPrincipalType: []Claim{{testIssuer, ClaimPrincipalType, PrincipalTypeInstance}},
			}},
			expected: PrincipalTypeInstance,
		},
		{tc: `should return PrincipalTypeService set in Claims.PrincipalType`,
			principal: &Principal{claims: Claims{
				ClaimPrincipalType: []Claim{{testIssuer, ClaimPrincipalType, PrincipalTypeService}},
			}},
			expected: PrincipalTypeService,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			ptype := test.principal.Type()
			assert.Equal(t, test.expected, ptype)
		})
	}
}

func TestClaimIsEmpty(t *testing.T) {
	testIO := []struct {
		tc       string
		claim    Claim
		expected bool
	}{
		{tc: `should return true if claim is empty`,
			claim:    Claim{},
			expected: true,
		},
		{tc: `should return false if claim is NOT empty`,
			claim:    testSubjectClaim,
			expected: false,
		},
	}
	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			result := test.claim.IsEmpty()
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestClaimsGetSingleClaim(t *testing.T) {
	testIO := []struct {
		tc       string
		key      string
		claims   Claims
		expected Claim
	}{
		{tc: `should return empty Claim for empty claims`,
			key:      ``,
			claims:   Claims{},
			expected: Claim{},
		},
		{tc: `should return expected string value given expires claim key`,
			key:      ClaimSubject,
			claims:   testPrincipalClaims,
			expected: testSubjectClaim,
		},
		{tc: `should returns the first element in claims if multiple values are present for a claim type `,
			key:      ClaimOrgUnit,
			claims:   testMultipleClaimTypes,
			expected: Claim{testIssuer, ClaimOrgUnit, "first-value"},
		},
		{tc: `should return an empty string for an unexpected key`,
			key:      "invalid",
			claims:   testPrincipalClaims,
			expected: Claim{},
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			value := test.claims.GetSingleClaim(test.key)
			assert.Equal(t, test.expected, value)
		})
	}
}

func TestClaimsGetString(t *testing.T) {
	testIO := []struct {
		tc       string
		key      string
		claims   Claims
		expected string
	}{
		{tc: `should return empty string for empty claims`,
			key:      ``,
			claims:   Claims{},
			expected: "",
		},
		{tc: `should return expected string value given expires claim key`,
			key:      ClaimExpires,
			claims:   testPrincipalClaims,
			expected: "10",
		},
		{tc: `should returns the first element in claims if multiple values are present for a claim type `,
			key:      ClaimOrgUnit,
			claims:   testMultipleClaimTypes,
			expected: "first-value",
		},
		{tc: `should return an empty string for an unexpected key`,
			key:      "invalid",
			claims:   testPrincipalClaims,
			expected: "",
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			value := test.claims.GetString(test.key)
			assert.Equal(t, test.expected, value)
		})
	}
}

func TestClaimsGetInt(t *testing.T) {
	testIO := []struct {
		tc          string
		key         string
		claims      Claims
		expected    int64
		expectedErr error
	}{
		{tc: `should return 0 for empty claims`,
			key:      ``,
			claims:   Claims{},
			expected: 0,
		},
		{tc: `should return expected int value given expires claim key`,
			key:      ClaimExpires,
			claims:   testPrincipalClaims,
			expected: 10,
		},
		{tc: `should return 0 and a string parse error if the value is not an int`,
			key:         ClaimSubject,
			claims:      testPrincipalClaims,
			expected:    0,
			expectedErr: &strconv.NumError{Func: "ParseInt", Num: "test-subject", Err: strconv.ErrSyntax},
		},
		{tc: `should return 0 for an unexpected key`,
			key:      "invalid",
			claims:   testPrincipalClaims,
			expected: 0,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			value, err := test.claims.GetInt(test.key)
			assert.Equal(t, test.expected, value)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}

func TestAddClaim(t *testing.T) {
	testIO := []struct {
		tc       string
		inputs   []Claim
		expected Claims
	}{
		{tc: `should add single claim`,
			inputs:   []Claim{testSubjectClaim},
			expected: Claims{ClaimSubject: []Claim{testSubjectClaim}},
		},
		{tc: `should add multiple claims to same claim type`,
			inputs: []Claim{
				testSubjectClaim,
				{testIssuer, ClaimSubject, "subject-alt"},
			},
			expected: Claims{
				ClaimSubject: []Claim{testSubjectClaim, {testIssuer, ClaimSubject, "subject-alt"}},
			},
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			claims := Claims{}
			for _, input := range test.inputs {
				claims.Add(input)
			}
			assert.Equal(t, test.expected, claims)
		})
	}
}

func TestToSlice(t *testing.T) {
	testIO := []struct {
		tc       string
		claims   Claims
		expected []Claim
	}{
		{tc: `should return empty slice from empty Claims`,
			claims:   Claims{},
			expected: []Claim{},
		},
		{tc: `should return expected list of claims containing single claim type`,
			claims:   Claims{ClaimSubject: []Claim{testSubjectClaim}},
			expected: []Claim{testSubjectClaim},
		},
		{tc: `should return expected list of claims containing with multiple claims of same claim type`,
			claims: testMultipleClaimTypes,
			expected: []Claim{
				{testIssuer, ClaimOrgUnit, "first-value"},
				{testIssuer, ClaimOrgUnit, "second-value"},
			},
		},
	}
	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			result := test.claims.ToSlice()
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestUnmarshalClaims(t *testing.T) {
	errJSON := json.Unmarshal([]byte(`?`), &map[string]string{})
	testIO := []struct {
		tc            string
		input         []byte
		expected      Claims
		expectedError error
	}{
		{tc: `valid input should return expected claims`,
			input: []byte(`{"iss": "test-issuer", "sub": "test-subject"}`),
			expected: Claims{
				ClaimIssuer:  []Claim{{"test-issuer", ClaimIssuer, "test-issuer"}},
				ClaimSubject: []Claim{{"test-issuer", ClaimSubject, "test-subject"}},
			},
			expectedError: nil,
		},
		{tc: `invalid json should return JSON unmarshal error`,
			input:         []byte(`?`),
			expected:      nil,
			expectedError: errJSON,
		},
		{tc: `empty issuer claim should result in empty issuer in Claim`,
			input: []byte(`{"sub": "test-subject"}`),
			expected: Claims{
				ClaimSubject: []Claim{{"", ClaimSubject, "test-subject"}},
			},
			expectedError: nil,
		},
		{tc: `should handle whole number value`,
			input: []byte(`{"exp": 1999}`),
			expected: Claims{
				ClaimExpires: []Claim{{"", ClaimExpires, "1999"}},
			},
			expectedError: nil,
		},
		{tc: `should handle a large whole number value`,
			input: []byte(`{"exp": 1520008887}`),
			expected: Claims{
				ClaimExpires: []Claim{{"", ClaimExpires, "1520008887"}},
			},
			expectedError: nil,
		},
		{tc: `should handle max size int`,
			input: []byte(fmt.Sprintf(`{"exp": %d}`, math.MaxInt64)),
			expected: Claims{
				ClaimExpires: []Claim{{"", ClaimExpires, "9223372036854775807"}},
			},
			expectedError: nil,
		},
		{tc: `should handle over max size int`,
			input: []byte(`{"exp": 9223372036854775808}`),
			expected: Claims{
				ClaimExpires: []Claim{{"", ClaimExpires, "9223372036854775808"}},
			},
			expectedError: nil,
		},
		{tc: `should handle a float value`,
			input: []byte(`{"exp": 0.0000000001}`),
			expected: Claims{
				ClaimExpires: []Claim{{"", ClaimExpires, "0.0000000001"}},
			},
			expectedError: nil,
		},
		{tc: `should handle max size float`,
			input: []byte(fmt.Sprintf(`{"exp": %f}`, math.MaxFloat64)),
			expected: Claims{
				// aprox. 1.8 * 10^308 per IEEE 754
				// https://en.wikipedia.org/wiki/Double-precision_floating-point_format
				ClaimExpires: []Claim{{"", ClaimExpires, fmt.Sprintf("%f", math.MaxFloat64)}},
			},
			expectedError: nil,
		},
		{tc: `should handle boolean value`,
			input: []byte(`{"sub": true}`),
			expected: Claims{
				ClaimSubject: []Claim{{"", ClaimSubject, "true"}},
			},
			expectedError: nil,
		},
		{tc: `should use InvalidClaimType for an unsupported type of array`,
			input: []byte(`{"sub": ["a"]}`),
			expected: Claims{
				ClaimSubject: []Claim{{"", ClaimSubject, "<INVALID_CLAIM_TYPE>"}},
			},
			expectedError: nil,
		},
		{tc: `should use InvalidClaimType for an unsupported type of embedded JSON`,
			input: []byte(`{"sub": {}}`),
			expected: Claims{
				ClaimSubject: []Claim{{"", ClaimSubject, "<INVALID_CLAIM_TYPE>"}},
			},
			expectedError: nil,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			claims, err := UnmarshalClaims(test.input)
			assert.Equal(t, test.expected, claims)
			assert.Equal(t, test.expectedError, err)
		})
	}
}
