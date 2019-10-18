// Copyright (c) 2018, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"oracle.com/oci/httpsigner"
	"time"
)

var testRequestPrincipal = &Principal{claims: Claims{
	ClaimSubject:         []Claim{{testIssuer, ClaimSubject, "subject-id"}},
	ClaimTenant:          []Claim{{testIssuer, ClaimTenant, "tenant-id"}},
	"h_authorization":    []Claim{{HdrClaimIssuer, "h_authorization", testAuthzHdr}},
	"h_date":             []Claim{{HdrClaimIssuer, "h_date", "Thu, 21 June 1582"}},
	"h_(request-target)": []Claim{{HdrClaimIssuer, "h_(request-target)", "get /"}},
	"h_host":             []Claim{{HdrClaimIssuer, "h_host", "example.test"}},
}}
var testRequestPrincipalHeaders = map[string][]string{
	"authorization":    {testAuthzHdr},
	"date":             {"Thu, 21 June 1582"},
	"(request-target)": {"get /"},
	"host":             {"example.test"}}

func TestGetRequestHeaders(t *testing.T) {

	testIO := []struct {
		tc              string
		principal       *Principal
		expectedHeaders map[string][]string
	}{
		{tc: `should return empty headers for nil principal`,
			principal:       nil,
			expectedHeaders: make(map[string][]string),
		},
		{tc: `should return headers from principal header claims`,
			principal:       testRequestPrincipal,
			expectedHeaders: testRequestPrincipalHeaders,
		},
	}
	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			r := getRequestHeaders(test.principal)
			assert.Equal(t, test.expectedHeaders, r)
		})
	}
}

func TestGetOBOToken(t *testing.T) {

	validClient := &MockSigningClient{}
	validEndpoint := "http://localhost/v1"
	validRequestPrincipal := testRequestPrincipal
	validRequestID := "arequestid"
	validExpiration := DefaultTokenExpiration
	validTargetServiceNames := []string{"SVC"}
	testTokenVal := "test-token"
	validJSON, _ := json.Marshal(&STSTokenResponse{Token: testTokenVal})
	invalidJSON := []byte(`{test: "test"`)
	invalidJSONErr := json.Unmarshal(invalidJSON, &STSTokenResponse{})
	validOboURI := "http://localhost/v1/obo"

	testIO := []struct {
		tc                 string
		client             httpsigner.Client
		endpoint           string
		requestPrincipal   *Principal
		delegatePrincipal  *Principal
		requestID          string
		expiration         int
		targetServiceNames []string
		expectedHeaders    map[string][]string
		expectedToken      string
		expectedError      error
	}{
		{tc: `should return ErrInvalidClient if client is nil`,
			client:             nil,
			endpoint:           validEndpoint,
			requestPrincipal:   validRequestPrincipal,
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: validTargetServiceNames,
			expectedToken:      "",
			expectedError:      ErrInvalidClient,
		},
		{tc: `should return ErrInvalidEndpoint if endpoint is empty`,
			client:             validClient,
			endpoint:           "",
			requestPrincipal:   validRequestPrincipal,
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: validTargetServiceNames,
			expectedToken:      "",
			expectedError:      ErrInvalidEndpoint,
		},
		{tc: `should return ErrInvalidRequestPrincipal if requestPrincipal is nil`,
			client:             validClient,
			endpoint:           validEndpoint,
			requestPrincipal:   nil,
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: validTargetServiceNames,
			expectedToken:      "",
			expectedError:      ErrInvalidRequestPrincipal,
		},
		{tc: `should return ErrNoTargetServiceNames if targetServiceNames empty`,
			client:             validClient,
			endpoint:           validEndpoint,
			requestPrincipal:   validRequestPrincipal,
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: make([]string, 0),
			expectedToken:      "",
			expectedError:      ErrNoTargetServiceNames,
		},
		{tc: `should return SyntaxError if identity response has malformed json`,
			client:             &MockSigningClient{doResponse: OKResponse(invalidJSON)},
			endpoint:           validEndpoint,
			requestPrincipal:   validRequestPrincipal,
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: validTargetServiceNames,
			expectedToken:      "",
			expectedError:      invalidJSONErr,
		},
		{tc: `should return ServiceError if identity response is not 200 status`,
			client:             &MockSigningClient{doResponse: non200Response},
			endpoint:           validEndpoint,
			requestPrincipal:   validRequestPrincipal,
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: validTargetServiceNames,
			expectedToken:      "",
			expectedError:      &ServiceResponseError{non200Response},
		},
		{tc: `should return an obo token after making a valid obo request`,
			client:             &MockSigningClient{doResponse: OKResponse(validJSON)},
			endpoint:           validEndpoint,
			requestPrincipal:   validRequestPrincipal,
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: validTargetServiceNames,
			expectedHeaders:    testRequestPrincipalHeaders,
			expectedToken:      testTokenVal,
			expectedError:      nil,
		},
	}
	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			token, err := GetOboToken(test.client,
				test.endpoint,
				test.requestPrincipal,
				test.delegatePrincipal,
				test.requestID,
				test.expiration,
				test.targetServiceNames)
			assert.Equal(t, test.expectedToken, token)
			assert.Equal(t, test.expectedError, err)
			if test.expectedError == nil {
				var oboReq onBehalfOfRequest
				c := test.client.(*MockSigningClient)
				req := c.DoRequest

				assert.Equal(t, test.requestID, req.Header.Get(requestIDHeader))
				assert.Equal(t, validOboURI, req.URL.String())

				buf := new(bytes.Buffer)
				buf.ReadFrom(req.Body)
				body := buf.Bytes()
				err := json.Unmarshal(body, &oboReq)

				assert.Nil(t, err)
				assert.Equal(t, oboReq.RequestHeaders, test.expectedHeaders)
				assert.Equal(t, oboReq.Expiration, test.expiration)
				assert.Equal(t, oboReq.TargetServiceNames, test.targetServiceNames)
				assert.Equal(t, oboReq.DelegateGroups, make([]string, 0))
				assert.Equal(t, oboReq.OboToken, "")
				assert.Equal(t, oboReq.RequestType, requestTypeOBO)
			}
		})
	}
}

func TestGetOBOOrDelegationToken(t *testing.T) {

	validClient := &MockSigningClient{}
	validOboURI := "http://localhost/v1/obo"
	validRequestPrincipal := testRequestPrincipal
	validRequestID := "arequestid"
	validExpiration := DefaultTokenExpiration * time.Second
	validTargetServiceNames := []string{"SVC"}
	testTokenVal := "test-token"
	validJSON, _ := json.Marshal(&STSTokenResponse{Token: testTokenVal})
	invalidJSON := []byte(`{test: "test"`)
	invalidJSONErr := json.Unmarshal(invalidJSON, &STSTokenResponse{})
	validDelegateGroups := make([]string, 0)
	invalidRequestType := "AN-INVALID-TYPE"

	testIO := []struct {
		tc                 string
		client             httpsigner.Client
		oboURI             string
		requestPrincipal   *Principal
		delegatePrincipal  *Principal
		OboToken           string
		requestID          string
		expiration         time.Duration
		targetServiceNames []string
		delegateGroups     []string
		requestType        string
		expectedHeaders    map[string][]string
		expectedToken      string
		expectedError      error
	}{
		{tc: `should return ErrInvalidClient if client is nil`,
			client:             nil,
			oboURI:             validOboURI,
			requestPrincipal:   validRequestPrincipal,
			delegatePrincipal:  nil,
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: validTargetServiceNames,
			delegateGroups:     validDelegateGroups,
			requestType:        requestTypeOBO,
			expectedToken:      "",
			expectedError:      ErrInvalidClient,
		},
		{tc: `should return ErrInvalidEndpoint if oboURI is empty`,
			client:             validClient,
			oboURI:             "",
			requestPrincipal:   validRequestPrincipal,
			delegatePrincipal:  nil,
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: validTargetServiceNames,
			delegateGroups:     validDelegateGroups,
			requestType:        requestTypeOBO,
			expectedToken:      "",
			expectedError:      ErrInvalidEndpoint,
		},
		{tc: `should return ErrInvalidRequestPrincipal if requestPrincipal is nil`,
			client:             validClient,
			oboURI:             validOboURI,
			requestPrincipal:   nil,
			delegatePrincipal:  nil,
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: validTargetServiceNames,
			delegateGroups:     validDelegateGroups,
			requestType:        requestTypeOBO,
			expectedToken:      "",
			expectedError:      ErrInvalidRequestPrincipal,
		},
		{tc: `should return ErrNoTargetServiceNames if targetServiceNames empty`,
			client:             validClient,
			oboURI:             validOboURI,
			requestPrincipal:   validRequestPrincipal,
			delegatePrincipal:  nil,
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: make([]string, 0),
			delegateGroups:     validDelegateGroups,
			requestType:        requestTypeOBO,
			expectedToken:      "",
			expectedError:      ErrNoTargetServiceNames,
		},
		{tc: `should return ErrInvalidRequestType if requestType not OBO or DELEGATE`,
			client:             validClient,
			oboURI:             validOboURI,
			requestPrincipal:   validRequestPrincipal,
			delegatePrincipal:  nil,
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: validTargetServiceNames,
			delegateGroups:     validDelegateGroups,
			requestType:        invalidRequestType,
			expectedToken:      "",
			expectedError:      ErrInvalidRequestType,
		},
		{tc: `should return SyntaxError if identity response has malformed json`,
			client:             &MockSigningClient{doResponse: OKResponse(invalidJSON)},
			oboURI:             validOboURI,
			requestPrincipal:   validRequestPrincipal,
			delegatePrincipal:  nil,
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: validTargetServiceNames,
			delegateGroups:     validDelegateGroups,
			requestType:        requestTypeOBO,
			expectedToken:      "",
			expectedError:      invalidJSONErr,
		},
		{tc: `should return ServiceError if identity response is not 200 status`,
			client:             &MockSigningClient{doResponse: non200Response},
			oboURI:             validOboURI,
			requestPrincipal:   validRequestPrincipal,
			delegatePrincipal:  nil,
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: validTargetServiceNames,
			delegateGroups:     validDelegateGroups,
			requestType:        requestTypeOBO,
			expectedToken:      "",
			expectedError:      &ServiceResponseError{non200Response},
		},
		{tc: `should return an obo token after making a valid obo request`,
			client:             &MockSigningClient{doResponse: OKResponse(validJSON)},
			oboURI:             validOboURI,
			requestPrincipal:   validRequestPrincipal,
			delegatePrincipal:  nil,
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: validTargetServiceNames,
			delegateGroups:     validDelegateGroups,
			requestType:        requestTypeOBO,
			expectedHeaders:    testRequestPrincipalHeaders,
			expectedToken:      testTokenVal,
			expectedError:      nil,
		},
		{tc: `should return ErrInvalidRequestPrincipal when request principal does not have an obo token claim`,
			client:           validClient,
			oboURI:           validOboURI,
			requestPrincipal: testRequestPrincipal,
			delegatePrincipal: &Principal{
				claims: Claims{
					ClaimPrincipalType:   []Claim{{Value: PrincipalTypeService}},
					"h_authorization":    []Claim{{HdrClaimIssuer, "h_authorization", testAuthzHdr}},
					"h_date":             []Claim{{HdrClaimIssuer, "h_date", "Thu, 21 June 1582"}},
					"h_(request-target)": []Claim{{HdrClaimIssuer, "h_(request-target)", "get /"}},
					"h_host":             []Claim{{HdrClaimIssuer, "h_host", "example.test"}},
				},
			},
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: validTargetServiceNames,
			delegateGroups:     validDelegateGroups,
			requestType:        requestTypeDelegation,
			expectedHeaders:    testRequestPrincipalHeaders,
			expectedError:      ErrInvalidRequestPrincipal,
		},
		{tc: `should return an obo token after making a valid obo request with a delegate principal`,
			client: &MockSigningClient{doResponse: OKResponse(validJSON)},
			oboURI: validOboURI,
			delegatePrincipal: &Principal{
				claims: Claims{
					ClaimPrincipalType:   []Claim{{Value: PrincipalTypeInstance}},
					"h_authorization":    []Claim{{HdrClaimIssuer, "h_authorization", testAuthzHdr}},
					"h_date":             []Claim{{HdrClaimIssuer, "h_date", "Thu, 21 June 1582"}},
					"h_(request-target)": []Claim{{HdrClaimIssuer, "h_(request-target)", "get /"}},
					"h_host":             []Claim{{HdrClaimIssuer, "h_host", "example.test"}},
				},
			},
			OboToken: "test-obo-token",
			requestPrincipal: &Principal{
				claims: Claims{
					ClaimOBOToken: []Claim{{Value: "test-obo-token"}},
				},
			},
			requestID:          "delegationRequestID",
			expiration:         validExpiration,
			targetServiceNames: validTargetServiceNames,
			delegateGroups:     validDelegateGroups,
			requestType:        requestTypeDelegation,
			expectedHeaders:    testRequestPrincipalHeaders,
			expectedToken:      testTokenVal,
			expectedError:      nil,
		},
	}
	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			token, err := getOboOrDelegationToken(test.client,
				test.oboURI,
				test.requestPrincipal,
				test.delegatePrincipal,
				test.requestID,
				test.expiration,
				test.targetServiceNames,
				test.delegateGroups,
				test.requestType)
			assert.Equal(t, test.expectedToken, token)
			assert.Equal(t, test.expectedError, err)
			if test.expectedError == nil {
				var oboReq onBehalfOfRequest
				c := test.client.(*MockSigningClient)
				req := c.DoRequest

				assert.Equal(t, test.requestID, req.Header.Get(requestIDHeader))
				assert.Equal(t, test.oboURI, req.URL.String())

				buf := new(bytes.Buffer)
				buf.ReadFrom(req.Body)
				body := buf.Bytes()
				err := json.Unmarshal(body, &oboReq)

				assert.Nil(t, err)
				assert.Equal(t, oboReq.RequestHeaders, test.expectedHeaders)
				assert.Equal(t, time.Duration(oboReq.Expiration)*time.Second, test.expiration)
				assert.Equal(t, oboReq.TargetServiceNames, test.targetServiceNames)
				assert.Equal(t, oboReq.DelegateGroups, test.delegateGroups)
				assert.Equal(t, oboReq.OboToken, test.OboToken)
				assert.Equal(t, oboReq.RequestType, test.requestType)
			}
		})
	}
}
func TestGetDelegationToken(t *testing.T) {

	validClient := &MockSigningClient{}
	validOboURI := "http://localhost/v1/obo"
	validRequestPrincipal := testRequestPrincipal
	validRequestID := "arequestid"
	validExpiration := DefaultTokenExpirationDuration
	validTargetServiceNames := []string{"SVC"}
	testTokenVal := "test-token"
	validDelegateGroups := make([]string, 0)

	testIO := []struct {
		tc                 string
		client             httpsigner.Client
		oboURI             string
		requestPrincipal   *Principal
		delegatePrincipal  *Principal
		requestID          string
		expiration         time.Duration
		targetServiceNames []string
		delegateGroups     []string
		expectedHeaders    map[string][]string
		expectedToken      string
		expectedError      error
	}{
		{tc: `should return ErrInvalidClient if client is nil`,
			client:             nil,
			oboURI:             validOboURI,
			requestPrincipal:   validRequestPrincipal,
			delegatePrincipal:  nil,
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: validTargetServiceNames,
			delegateGroups:     validDelegateGroups,
			expectedToken:      "",
			expectedError:      ErrInvalidClient,
		},
		{tc: `should return ErrInvalidEndpoint if oboURI is empty`,
			client:             validClient,
			oboURI:             "",
			requestPrincipal:   validRequestPrincipal,
			delegatePrincipal:  nil,
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: validTargetServiceNames,
			delegateGroups:     validDelegateGroups,
			expectedToken:      "",
			expectedError:      ErrInvalidEndpoint,
		},
		{tc: `should return ErrInvalidRequestPrincipal if requestPrincipal is nil`,
			client:             validClient,
			oboURI:             validOboURI,
			requestPrincipal:   nil,
			delegatePrincipal:  nil,
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: validTargetServiceNames,
			delegateGroups:     validDelegateGroups,
			expectedToken:      "",
			expectedError:      ErrInvalidRequestPrincipal,
		},
		{tc: `should return ErrNoTargetServiceNames if targetServiceNames empty`,
			client:             validClient,
			oboURI:             validOboURI,
			requestPrincipal:   validRequestPrincipal,
			delegatePrincipal:  nil,
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: make([]string, 0),
			delegateGroups:     validDelegateGroups,
			expectedError:      ErrNoTargetServiceNames,
		},
		{tc: `should return ErrInvalidDelegateGroups if delegateGroups is empty`,
			client:             validClient,
			oboURI:             validOboURI,
			requestPrincipal:   validRequestPrincipal,
			delegatePrincipal:  nil,
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: validTargetServiceNames,
			delegateGroups:     []string{},
			expectedError:      ErrInvalidDelegateGroups,
		},
		{tc: `should return a error due to request principal does not contain obo`,
			client: mockClient{
				Responses: []mockedResponses{
					{URL: "/v1/obo/obo", Err: nil, Response: []byte(`{"token": "test-token"}`)},
				},
			},
			oboURI:           validOboURI,
			requestPrincipal: validRequestPrincipal,
			delegatePrincipal: &Principal{
				claims: Claims{
					ClaimPrincipalType:   []Claim{{Value: PrincipalTypeService}},
					"h_authorization":    []Claim{{HdrClaimIssuer, "h_authorization", testAuthzHdr}},
					"h_date":             []Claim{{HdrClaimIssuer, "h_date", "Thu, 21 June 1582"}},
					"h_(request-target)": []Claim{{HdrClaimIssuer, "h_(request-target)", "get /"}},
					"h_host":             []Claim{{HdrClaimIssuer, "h_host", "example.test"}},
				},
			},
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: validTargetServiceNames,
			delegateGroups:     []string{"g1"},
			expectedToken:      "",
			expectedError:      ErrInvalidRequestPrincipal,
		},
		{tc: `should return a valid obo token with a valid response`,
			client: mockClient{
				Responses: []mockedResponses{
					{URL: "/v1/obo/obo", Err: nil, Response: []byte(`{"token": "test-token"}`)},
				},
			},
			oboURI: validOboURI,
			requestPrincipal: &Principal{
				claims: Claims{
					ClaimOBOToken: []Claim{{Value: "test-obo-token"}},
				},
			},
			delegatePrincipal: &Principal{
				claims: Claims{
					ClaimPrincipalType:   []Claim{{Value: PrincipalTypeService}},
					"h_authorization":    []Claim{{HdrClaimIssuer, "h_authorization", testAuthzHdr}},
					"h_date":             []Claim{{HdrClaimIssuer, "h_date", "Thu, 21 June 1582"}},
					"h_(request-target)": []Claim{{HdrClaimIssuer, "h_(request-target)", "get /"}},
					"h_host":             []Claim{{HdrClaimIssuer, "h_host", "example.test"}},
				},
			},
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: validTargetServiceNames,
			delegateGroups:     []string{"g1"},
			expectedToken:      testTokenVal,
			expectedError:      nil,
		},
		{tc: `should return a valid token`,
			client: mockClient{
				Responses: []mockedResponses{
					{URL: "/v1/obo/obo", Err: nil, Response: []byte(`{"token": "test-token"}`)},
				},
			},
			oboURI:             validOboURI,
			requestPrincipal:   validRequestPrincipal,
			delegatePrincipal:  nil,
			requestID:          validRequestID,
			expiration:         validExpiration,
			targetServiceNames: validTargetServiceNames,
			delegateGroups:     []string{"g1"},
			expectedToken:      testTokenVal,
			expectedError:      nil,
		},
	}
	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			token, err := GetDelegationToken(test.client,
				test.oboURI,
				test.requestPrincipal,
				test.delegatePrincipal,
				test.requestID,
				test.expiration,
				test.targetServiceNames,
				test.delegateGroups)
			assert.Equal(t, test.expectedToken, token)
			assert.Equal(t, test.expectedError, err)
		})
	}
}
