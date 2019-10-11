// Copyright (c) 2018-2019, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"oracle.com/oci/httpsigner"
	"oracle.com/oci/tagging"
)

var (
	testRequestID         = "customer/requestID/response"
	testRequestIDOriginal = "customer/requestID/"
	testOperationID       = "operationID"
	testCompartmentID     = "compartmentID"
	testServiceName       = "testService"
	testSubjectID         = "test-subject-id"
	testRequestRegion     = "test-request-region"
	testPhysicalAD        = "test-physical-ad"
	testActionKind        = ActionKind("READ")
)

var validAuthorizationResponse = AuthorizationResponse{
	OutboundAuthorizationRequest: OutboundAuthorizationRequest{
		RequestID:     testRequestIDOriginal,
		ServiceName:   "test-service-name",
		UserPrincipal: AuthorizationRequestPrincipal{},
		Principal:     AuthorizationRequestPrincipal{},
		Context:       [][]ContextVariable{},
		PhysicalAD:    "test-ad",
		RequestRegion: "test-region",
	},
}

var validAssociationAuthorizationResponse = AssociationAuthorizationResponse{
	Responses: []AuthorizationResponse{{
		OutboundAuthorizationRequest: OutboundAuthorizationRequest{
			RequestID:     testRequestIDOriginal,
			ServiceName:   "test-service-name",
			UserPrincipal: AuthorizationRequestPrincipal{},
			Principal:     AuthorizationRequestPrincipal{},
			Context:       [][]ContextVariable{},
			PhysicalAD:    "test-ad",
			RequestRegion: "test-region",
		},
	}, {
		OutboundAuthorizationRequest: OutboundAuthorizationRequest{
			RequestID:     testRequestIDOriginal,
			ServiceName:   "test-service-name",
			UserPrincipal: AuthorizationRequestPrincipal{},
			Principal:     AuthorizationRequestPrincipal{},
			Context:       [][]ContextVariable{},
			PhysicalAD:    "test-ad",
			RequestRegion: "test-region",
		},
	},
	},
	AssocationResult: AssociationSuccess,
}

var allActionKinds = []ActionKind{
	ActionKindCreate,
	ActionKindRead,
	ActionKindUpdate,
	ActionKindDelete,
	ActionKindList,
	ActionKindAttach,
	ActionKindDetach,
	ActionKindOther,
	ActionKindNotDefined,
	ActionKindSearch,
	ActionKindUpdateRemoveOnly}

var (
	testFreeformTagSet = tagging.FreeformTagSet{
		"key1": "val1",
		"key2": "val2",
	}
	testDefinedTagSet = tagging.DefinedTagSet{
		"ns1": map[string]string{
			"key1": "val1",
			"key2": "val2",
		},
	}
	testTagSlug, _     = tagging.NewTagSlug(testFreeformTagSet, testDefinedTagSet)
	emptyTagSlug, _    = tagging.NewTagSlug(tagging.FreeformTagSet{}, tagging.DefinedTagSet{})
	freeformTagSlug, _ = tagging.NewTagSlug(testFreeformTagSet, tagging.DefinedTagSet{})
	definedTagSlug, _  = tagging.NewTagSlug(tagging.FreeformTagSet{}, testDefinedTagSet)
	fullTagSlug, _     = tagging.NewTagSlug(testFreeformTagSet, testDefinedTagSet)

	validAuthzRespWithNewTagSlug      = newAuthzResponseWithTags(validAuthorizationResponse, nil, testTagSlug, testTagSlug, "")
	validAuthzRespWithExistingTagSlug = newAuthzResponseWithTags(validAuthorizationResponse, testTagSlug, nil, testTagSlug, "")
)

func newAuthzResponseWithTags(a AuthorizationResponse, existingTagSlug, newTagSlug, mergedTagSlug *tagging.TagSlug, tagSlugError string) *AuthorizationResponse {
	a.OutboundAuthorizationRequest.TagSlugOriginal = existingTagSlug
	a.OutboundAuthorizationRequest.TagSlugChanges = newTagSlug
	a.OutboundAuthorizationRequest.TagSlugMerged = mergedTagSlug
	a.OutboundAuthorizationRequest.TagSlugError = tagSlugError
	return &a
}

func validAuthzResponseWithTags(a AuthorizationResponse) *AuthorizationResponse {
	var tagErr error
	if a.OutboundAuthorizationRequest.TagSlugError != "" {
		tagErr = errors.New(a.OutboundAuthorizationRequest.TagSlugError)
	}
	return authzResponseWithContext(a, AuthorizationContext{
		RequestID:     testRequestID,
		TagSlugError:  tagErr,
		TagSlugMerged: a.GetTagSlug(),
	})
}

func authzResponseWithContext(a AuthorizationResponse, c AuthorizationContext) *AuthorizationResponse {
	a.AuthorizationContext.RequestID = c.RequestID
	a.AuthorizationContext.Permissions = c.Permissions
	a.AuthorizationContext.TagSlugMerged = c.TagSlugMerged
	a.AuthorizationContext.TagSlugError = c.TagSlugError
	return &a
}

func authzResponseJSON(a *AuthorizationResponse) []byte {
	resp, err := json.Marshal(*a)
	if err != nil {
		panic(err)
	}
	return resp
}

func validAuthzResponseWithContext(c AuthorizationContext) *AuthorizationResponse {
	return authzResponseWithContext(validAuthorizationResponse, c)
}

var validContextVariable = [][]ContextVariable{{{P: `__COMMON__`}}, {{P: `permission-1`}}, {{P: `permission-2`}}}
var validEmptyContextVariable = [][]ContextVariable{{{P: `__COMMON__`}}}

var validAuthorizationResponseJSON, _ = json.Marshal(validAuthorizationResponse)

var validAssociationAuthorizationResponseJSON, _ = json.Marshal(validAssociationAuthorizationResponse)

// authz context
var (
	authzContext0Permission = AuthorizationContext{
		RequestID:   testRequestID,
		Permissions: []string{},
	}
	authzContext1Permission = AuthorizationContext{
		RequestID:   testRequestID,
		Permissions: []string{"permission-1"},
	}
	authzContext2Permissions = AuthorizationContext{
		RequestID:   testRequestID,
		Permissions: []string{"permission-1", "permission-2"},
	}
)

// outbound authz request
var (
	testEmptyAuthzPrincipal   = AuthorizationRequestPrincipal{}
	testUserAuthzPrincipal    = AuthorizationRequestPrincipal{SubjectID: "user-principal", Claims: []Claim{}}
	testServiceAuthzPrincipal = AuthorizationRequestPrincipal{SubjectID: "service-principal", Claims: []Claim{}}
)

func generateResponseFromPermissions(permissions []string) AuthorizationResponse {
	authzRequest := &AuthorizationRequest{
		RequestID:        testRequestIDOriginal,
		ServiceName:      "test-service-name",
		UserPrincipal:    &Principal{},
		ServicePrincipal: &Principal{},
		PhysicalAD:       "test-ad",
		RequestRegion:    "test-region",
	}
	for _, p := range permissions {
		authzRequest.SetPermissionVariables(p, []AuthzVariable{{Name: "test", Type: CtxVarTypeString, Value: "sup"}})
	}
	outboundRequest := authorizationRequestToOutbound(authzRequest)
	return AuthorizationResponse{OutboundAuthorizationRequest: *outboundRequest}
}

// Returns a 200 response given the body content for mocking authz service response
func validAuthzResponse(b []byte) *http.Response {
	response := &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewBuffer(b))}
	response.Header = http.Header{}
	response.Header.Add(requestIDHeader, testRequestID)
	return response
}

// Returns a 200 response given the body content for mocking authz association service response
func validAssociationResponseHTTP(b []byte) *http.Response {
	response := &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewBuffer(b))}
	return response
}

func TestNewAuthorizationClient(t *testing.T) {
	var testSigner httpsigner.RequestSigner = &MockRequestSigner{}
	client := NewSigningClient(testSigner, testKeyID)
	t.Run(
		`should return a client with non-nil signer`,
		func(t *testing.T) {
			// Test for NewAuthorizationClient
			c := NewAuthorizationClient(client, "http://localhost")
			assert.NotNil(t, c)
			assert.Equal(t, testSigner, c.(*authorizationClient).client.(*SigningClient).Signer())
			assert.Equal(t, testKeyID, c.(*authorizationClient).client.(*SigningClient).KeyID())
			assert.Equal(t, authorizeURITemplate, c.(*authorizationClient).uriTemplate)
			assert.False(t, c.(*authorizationClient).taggingEnabled)

			// Test for NewAuthorizationClientExtended
			extended := NewAuthorizationClientExtended(client, "http://localhost")
			assert.NotNil(t, extended)
			assert.Equal(t, testSigner, extended.(*authorizationClient).client.(*SigningClient).Signer())
			assert.Equal(t, testKeyID, extended.(*authorizationClient).client.(*SigningClient).KeyID())
			assert.Equal(t, authorizeURITemplate, extended.(*authorizationClient).uriTemplate)
			assert.False(t, extended.(*authorizationClient).taggingEnabled)

			// Test for NewAssociationAuthorizationClient
			association := NewAssociationAuthorizationClient(client, "http://localhost")
			assert.NotNil(t, association)
			assert.Equal(t, testSigner, association.(*authorizationClient).client.(*SigningClient).Signer())
			assert.Equal(t, testKeyID, association.(*authorizationClient).client.(*SigningClient).KeyID())
			assert.Equal(t, authorizeURITemplate, association.(*authorizationClient).uriTemplate)
			assert.False(t, association.(*authorizationClient).taggingEnabled)

			// Test for NewAuthorizationClientWithTags
			c = NewAuthorizationClientWithTags(client, "http://localhost")
			assert.NotNil(t, c)
			assert.Equal(t, testSigner, c.(*authorizationClient).client.(*SigningClient).Signer())
			assert.Equal(t, testKeyID, c.(*authorizationClient).client.(*SigningClient).KeyID())
			assert.Equal(t, authorizeWithTagsURITemplate, c.(*authorizationClient).uriTemplate)
			assert.True(t, c.(*authorizationClient).taggingEnabled)

			// Test for NewAssociationAuthorizationClientWithTags
			c = NewAssociationAuthorizationClientWithTags(client, "http://localhost")
			assert.NotNil(t, c)
			assert.Equal(t, testSigner, c.(*authorizationClient).client.(*SigningClient).Signer())
			assert.Equal(t, testKeyID, c.(*authorizationClient).client.(*SigningClient).KeyID())
			assert.Equal(t, authorizeWithTagsURITemplate, c.(*authorizationClient).uriTemplate)
			assert.True(t, c.(*authorizationClient).taggingEnabled)
		})
}

func TestNewAuthorizationClientError(t *testing.T) {
	var testSigner httpsigner.RequestSigner = &MockRequestSigner{}
	client := NewSigningClient(testSigner, testKeyID)
	testIO := []struct {
		tc       string
		client   httpsigner.Client
		endpoint string
	}{
		{tc: `should panic if client is nil`,
			client: nil, endpoint: "http://localhost"},
		{tc: `should panic if endpoint is empty string`,
			client: client, endpoint: ""},
	}
	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			assert.Panics(t, func() { NewAuthorizationClient(test.client, test.endpoint) })
			assert.Panics(t, func() { NewAuthorizationClientExtended(test.client, test.endpoint) })
			assert.Panics(t, func() { NewAssociationAuthorizationClient(test.client, test.endpoint) })
		})
	}
}

func TestNewAuthorizationRequest(t *testing.T) {
	t.Run(
		`should return a request with valid inputs`,
		func(t *testing.T) {
			userPrincipal := NewPrincipal(testSubjectID, testTenantID)
			principal := NewPrincipal("another-subject-id", testTenantID)
			request, err := NewAuthorizationRequest(testRequestID, testOperationID, testCompartmentID, testServiceName,
				userPrincipal, principal, testRequestRegion, testPhysicalAD)
			assert.Nil(t, err)
			err = request.SetActionKind(testActionKind)
			assert.Nil(t, err)
			assert.Equal(t, testRequestID, request.RequestID)
			assert.Equal(t, testOperationID, request.OperationID)
			assert.Equal(t, testCompartmentID, request.CompartmentID)
			assert.Equal(t, testServiceName, request.ServiceName)
			assert.Equal(t, userPrincipal, request.UserPrincipal)
			assert.Equal(t, principal, request.ServicePrincipal)
			assert.Equal(t, testRequestRegion, request.RequestRegion)
			assert.Equal(t, testPhysicalAD, request.PhysicalAD)
			assert.Equal(t, testActionKind, request.ActionKind)
		})
}

func TestNewAssociationAuthorizationRequest(t *testing.T) {
	userPrincipal := NewPrincipal(testSubjectID, testTenantID)
	principal := NewPrincipal("another-subject-id", testTenantID)
	request, _ := NewAuthorizationRequest(testRequestID, testOperationID, testCompartmentID, testServiceName,
		userPrincipal, principal, testRequestRegion, testPhysicalAD)
	testIO := []struct {
		tc       string
		requests int
		err      error
	}{
		{tc: `zero AuthorizationRequest in AssociationAuthorizationRequest should fail`, requests: 0, err: ErrAssociationInsufficientRequests},
		{tc: `one AuthorizationRequest in AssociationAuthorizationRequest should fail`, requests: 1, err: ErrAssociationInsufficientRequests},
		{tc: `two AuthorizationRequest in AssociationAuthorizationRequest should succeed`, requests: 2, err: nil},
		{tc: `ten AuthorizationRequest in AssociationAuthorizationRequest should succeed`, requests: 10, err: nil},
	}
	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			var requestList []AuthorizationRequest
			for i := 1; i <= test.requests; i++ {
				requestList = append(requestList, *request)
			}
			assocReq, err := NewAssociationAuthorizationRequest(requestList...)
			if test.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, test.requests, len(*assocReq))
			} else {
				assert.Equal(t, err, test.err)
				assert.Nil(t, assocReq)
			}
		})
	}
}

func TestNewAuthorizationRequestError(t *testing.T) {
	principal := NewPrincipal(testSubjectID, testTenantID)
	testIO := []struct {
		tc            string
		requestID     string
		operationID   string
		compartmentID string
		userPrincipal *Principal
		requestRegion string
		physicalAD    string
		expectedError error
	}{
		{tc: `should return ErrInvalidOperationID if the given operationID is empty`,
			requestID: testRequestID, operationID: "", compartmentID: testCompartmentID, userPrincipal: principal, requestRegion: testRequestRegion,
			physicalAD: testPhysicalAD, expectedError: ErrInvalidOperationID},
		{tc: `should return ErrInvalidCompartmentID if the given compartmentID is empty`,
			requestID: testRequestID, operationID: testOperationID, compartmentID: "", userPrincipal: principal, requestRegion: testRequestRegion,
			physicalAD: testPhysicalAD, expectedError: ErrInvalidCompartmentID},
		{tc: `should return ErrInvalidRequestRegion if the given requestRegion is empty`,
			requestID: testRequestID, operationID: testOperationID, compartmentID: testCompartmentID, userPrincipal: principal, requestRegion: "",
			physicalAD: testPhysicalAD, expectedError: ErrInvalidRequestRegion},
		{tc: `should return ErrInvalidPhysicalAD if the given physicalAD is empty`,
			requestID: testRequestID, operationID: testOperationID, compartmentID: testCompartmentID, userPrincipal: principal, requestRegion: testPhysicalAD,
			physicalAD: "", expectedError: ErrInvalidPhysicalAD},
		{tc: `should return ErrInvalidPrincipal if both user principal and service principal are nil`,
			requestID: testRequestID, operationID: testOperationID, compartmentID: testCompartmentID, userPrincipal: nil, requestRegion: testRequestRegion,
			physicalAD: testPhysicalAD, expectedError: ErrInvalidPrincipal},
	}
	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			request, err := NewAuthorizationRequest(test.requestID, test.operationID, test.compartmentID, "",
				test.userPrincipal, nil, test.requestRegion, test.physicalAD)
			assert.Nil(t, request)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestNewAuthorizationSetActionKindError(t *testing.T) {
	principal := NewPrincipal(testSubjectID, testTenantID)
	testIO := []struct {
		tc            string
		actionKind    ActionKind
		expectedError error
	}{
		{tc: `should return ErrInvalidActionKind if the given ActionKind is invalid`, actionKind: ActionKind("hohoho"), expectedError: ErrInvalidActionKind},
		{tc: `should return ErrInvalidActionKind if the given ActionKind is invalid`, actionKind: ActionKind(""), expectedError: ErrInvalidActionKind},
		{tc: `should return nil if the given ActionKind is valid`, actionKind: testActionKind, expectedError: nil},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			request, err := NewAuthorizationRequest(testRequestID, testOperationID, testCompartmentID, "",
				principal, nil, testRequestRegion, testPhysicalAD)
			assert.Nil(t, err)
			err = request.SetActionKind(test.actionKind)
			assert.Equal(t, test.expectedError, err)
			if test.expectedError == nil {
				assert.Equal(t, test.actionKind, request.ActionKind)
			}
		})
	}
}

// Ensure default ActionKindNotDefined set.
func TestNewAuthorizationDefaultActionKind(t *testing.T) {
	principal := NewPrincipal(testSubjectID, testTenantID)
	request, err := NewAuthorizationRequest(testRequestID, testOperationID, testCompartmentID, "",
		principal, nil, testRequestRegion, testPhysicalAD)
	assert.Nil(t, err)
	assert.Equal(t, request.ActionKind, ActionKindNotDefined)
}

func TestAuthorizationRequestToOutboundPrincipal(t *testing.T) {
	testUserPrincipal := &Principal{subject: "user-principal"}
	testServicePrincipal := &Principal{subject: "service-principal"}

	testIO := []struct {
		tc                       string
		az                       *AuthorizationRequest
		expectedPrincipal        AuthorizationRequestPrincipal
		expectedUserPrincipal    AuthorizationRequestPrincipal
		expectedServicePrincipal AuthorizationRequestPrincipal
		expectedOBOPrincipal     AuthorizationRequestPrincipal
	}{
		{tc: `should have all empty principal if no principal is available in authz request`,
			az:                       &AuthorizationRequest{},
			expectedPrincipal:        testEmptyAuthzPrincipal,
			expectedUserPrincipal:    testEmptyAuthzPrincipal,
			expectedServicePrincipal: testEmptyAuthzPrincipal,
			expectedOBOPrincipal:     testEmptyAuthzPrincipal,
		},
		{tc: `should set principal userPrincipal as USER principal if authz request service principal is nil`,
			az:                       &AuthorizationRequest{UserPrincipal: testUserPrincipal},
			expectedPrincipal:        testUserAuthzPrincipal,
			expectedUserPrincipal:    testUserAuthzPrincipal,
			expectedServicePrincipal: testEmptyAuthzPrincipal,
			expectedOBOPrincipal:     testEmptyAuthzPrincipal,
		},
		{tc: `should set principal servicePrincipal as SERVICE principal if authz request user principal is nil`,
			az:                       &AuthorizationRequest{ServicePrincipal: testServicePrincipal},
			expectedPrincipal:        testServiceAuthzPrincipal,
			expectedUserPrincipal:    testEmptyAuthzPrincipal,
			expectedServicePrincipal: testServiceAuthzPrincipal,
			expectedOBOPrincipal:     testEmptyAuthzPrincipal,
		},
		{tc: `should set principal, servicePrincipal, userPrincipal, oboPrincipal if neither principal is nil`,
			az: &AuthorizationRequest{
				UserPrincipal:    testUserPrincipal,
				ServicePrincipal: testServicePrincipal,
			},
			expectedPrincipal:        testServiceAuthzPrincipal,
			expectedUserPrincipal:    testUserAuthzPrincipal,
			expectedServicePrincipal: testServiceAuthzPrincipal,
			expectedOBOPrincipal:     testUserAuthzPrincipal,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			result := authorizationRequestToOutbound(test.az)
			assert.Equal(t, test.expectedPrincipal, result.Principal)
			assert.Equal(t, test.expectedUserPrincipal, result.UserPrincipal)
			assert.Equal(t, test.expectedServicePrincipal, result.ServicePrincipal)
			assert.Equal(t, test.expectedOBOPrincipal, result.OBOPrincipal)

			marshaled, err := json.Marshal(result)
			assert.Nil(t, err)

			var unmarshaled map[string]interface{}
			err = json.Unmarshal(marshaled, &unmarshaled)
			assert.Nil(t, err)

			expectedPrincipal := !test.expectedPrincipal.empty()
			_, ok := unmarshaled["principal"]
			assert.Equal(t, expectedPrincipal, ok)

			expectedUserPrincipal := !test.expectedUserPrincipal.empty()
			_, ok = unmarshaled["userPrincipal"]
			assert.Equal(t, expectedUserPrincipal, ok)

			expectedServicePrincipal := !test.expectedServicePrincipal.empty()
			_, ok = unmarshaled["svcPrincipal"]
			assert.Equal(t, expectedServicePrincipal, ok)

			expectedOBOPrincipal := !test.expectedOBOPrincipal.empty()
			_, ok = unmarshaled["oboPrincipal"]
			assert.Equal(t, expectedOBOPrincipal, ok)
		})
	}
}

func TestAuthorizationRequestPrincipalEmpty(t *testing.T) {
	testIO := []struct {
		tc        string
		principal AuthorizationRequestPrincipal
		empty     bool
	}{
		{
			tc:        `should return true for an empty principal`,
			principal: testEmptyAuthzPrincipal,
			empty:     true,
		},
		{
			tc:        `should return false for non-empty subject id`,
			principal: AuthorizationRequestPrincipal{SubjectID: "subject-id"},
			empty:     false,
		},
		{
			tc:        `should return false for non-empty tenant id`,
			principal: AuthorizationRequestPrincipal{TenantID: "tenant-id"},
			empty:     false,
		},
		{
			tc:        `should return false for non-nil claims`,
			principal: AuthorizationRequestPrincipal{Claims: []Claim{}},
			empty:     false,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			assert.Equal(t, test.empty, test.principal.empty())
		})
	}

}

func TestMarshalOutboundAuthorizationRequest(t *testing.T) {
	testIO := []struct {
		tc string
		az OutboundAuthorizationRequest
	}{
		{
			tc: `should marshal full outbound authorization request`,
			az: OutboundAuthorizationRequest{
				Client:           "client",
				RequestID:        "request-id",
				ServiceName:      "service-name",
				UserPrincipal:    testUserAuthzPrincipal,
				ServicePrincipal: testServiceAuthzPrincipal,
				OBOPrincipal:     testServiceAuthzPrincipal,
				Principal:        testUserAuthzPrincipal,
				Context: [][]ContextVariable{
					{{P: "PERMISSION"}, {Name: "name", Type: "STRING", Types: "", Value: "hello", Boolean: nil}},
				},
				Properties:    struct{}{},
				RequestRegion: "request-region",
				PhysicalAD:    "physical-ad",
			},
		},
		{
			tc: `should marshal tagging fields outbound authorization request`,
			az: OutboundAuthorizationRequest{
				Client:           "client",
				RequestID:        "request-id",
				ServiceName:      "service-name",
				UserPrincipal:    testUserAuthzPrincipal,
				ServicePrincipal: testServiceAuthzPrincipal,
				OBOPrincipal:     testServiceAuthzPrincipal,
				Principal:        testUserAuthzPrincipal,
				Context: [][]ContextVariable{
					{{P: "PERMISSION"}, {Name: "name", Type: "STRING", Types: "", Value: "hello", Boolean: nil}},
				},
				Properties:        struct{}{},
				RequestRegion:     "request-region",
				PhysicalAD:        "physical-ad",
				TagSlugOriginal:   testTagSlug,
				TagSlugChanges:    testTagSlug,
				TagSlugMerged:     testTagSlug,
				TagSlugError:      "The following tag namespaces / keys are not authorized or not found: 'foo'",
				ResponseErrorType: TaggingAuthzOrNotExistError,
			},
		},
		{
			tc: `should omit empty service principal from being marshaled`,
			az: OutboundAuthorizationRequest{
				Client:           "client",
				RequestID:        "request-id",
				ServiceName:      "service-name",
				UserPrincipal:    testUserAuthzPrincipal,
				ServicePrincipal: testEmptyAuthzPrincipal,
				OBOPrincipal:     testEmptyAuthzPrincipal,
				Principal:        testUserAuthzPrincipal,
				Context: [][]ContextVariable{
					{{P: "PERMISSION"}, {Name: "name", Type: "STRING", Types: "", Value: "hello", Boolean: nil}},
				},
				Properties:    struct{}{},
				RequestRegion: "request-region",
				PhysicalAD:    "physical-ad",
			},
		},
		{
			tc: `should omit empty user principal from being marshaled`,
			az: OutboundAuthorizationRequest{
				Client:           "client",
				RequestID:        "request-id",
				ServiceName:      "service-name",
				UserPrincipal:    testEmptyAuthzPrincipal,
				ServicePrincipal: testServiceAuthzPrincipal,
				OBOPrincipal:     testServiceAuthzPrincipal,
				Principal:        testEmptyAuthzPrincipal,
				Context: [][]ContextVariable{
					{{P: "PERMISSION"}, {Name: "name", Type: "STRING", Types: "", Value: "hello", Boolean: nil}},
				},
				Properties:    struct{}{},
				RequestRegion: "request-region",
				PhysicalAD:    "physical-ad",
			},
		},
		{
			tc: `should marshal empty authorization request`,
			az: OutboundAuthorizationRequest{},
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			marshaled, err := json.Marshal(test.az)
			assert.Nil(t, err)

			var unmarshaled OutboundAuthorizationRequest
			err = json.Unmarshal(marshaled, &unmarshaled)
			assert.Nil(t, err)
			assert.Equal(t, unmarshaled, test.az)

			// Ensure fields expected to be omitted are not present
			var unmarshaledMap map[string]interface{}
			err = json.Unmarshal(marshaled, &unmarshaledMap)
			assert.Nil(t, err)

			expectedPrincipal := !test.az.Principal.empty()
			_, ok := unmarshaledMap["principal"]
			assert.Equal(t, expectedPrincipal, ok)

			expectedUserPrincipal := !test.az.UserPrincipal.empty()
			_, ok = unmarshaledMap["userPrincipal"]
			assert.Equal(t, expectedUserPrincipal, ok)

			expectedServicePrincipal := !test.az.ServicePrincipal.empty()
			_, ok = unmarshaledMap["svcPrincipal"]
			assert.Equal(t, expectedServicePrincipal, ok)

			expectedOBOPrincipal := !test.az.OBOPrincipal.empty()
			_, ok = unmarshaledMap["oboPrincipal"]
			assert.Equal(t, expectedOBOPrincipal, ok)
		})
	}
}

func TestSetCommonPermissions(t *testing.T) {
	t.Run(`should set the expected common permissions`, func(t *testing.T) {
		r := AuthorizationRequest{CompartmentID: testCompartmentID, OperationID: testOperationID}
		r.SetCommonPermissions()

		// First should always be just the permission
		assert.Equal(t, r.Context[0][0].P, commonPermission)

		found := 0
		for _, record := range r.Context {
			for _, context := range record {
				if context.Name == "target.compartment.id" {
					assert.Equal(t, context.Type, CtxVarTypeEntity)
					assert.Equal(t, context.Value, testCompartmentID)
					found++
				}
				if context.Name == "request.operation" {
					assert.Equal(t, context.Type, CtxVarTypeString)
					assert.Equal(t, context.Value, testOperationID)
					found++
				}
			}
		}
		assert.Equal(t, found, 2)
	})
}

func TestSetPermissions(t *testing.T) {
	trueP := new(bool)
	falseP := new(bool)
	*trueP = true
	*falseP = false

	testIO := []struct {
		tc         string
		permission string
		contexts   []AuthzVariable
		expected   [][]ContextVariable
	}{
		{
			tc:         `should return expected context variables that contains string`,
			permission: "PERMISSION",
			contexts:   []AuthzVariable{{Name: "name", Type: CtxVarTypeString, Types: "", Value: "hello", Values: nil}},
			expected:   [][]ContextVariable{{{P: "PERMISSION"}, {Name: "name", Type: "STRING", Types: "", Value: "hello", Boolean: nil}}},
		},
		{
			tc:         `should return expected context variables that contains a list`,
			permission: "PERMISSION",
			contexts:   []AuthzVariable{{Name: "name", Type: CtxVarTypeList, Types: CtxVarTypeString, Values: []string{"a", "b", "c"}}},
			expected:   [][]ContextVariable{{{P: "PERMISSION"}, {Name: "name", Type: CtxVarTypeList, Types: CtxVarTypeString, Values: []string{"a", "b", "c"}, Boolean: nil}}},
		},
		{
			tc:         `should return expected context variables that contains a boolean (true)`,
			permission: "PERMISSION",
			contexts:   []AuthzVariable{{Name: "name", Type: CtxVarTypeBool, Types: "", Boolean: true}},
			expected:   [][]ContextVariable{{{P: "PERMISSION"}, {Name: "name", Type: "BOOLEAN", Types: "", Boolean: trueP}}},
		},
		{
			tc:         `should return expected context variables that contains a boolean (false)`,
			permission: "PERMISSION",
			contexts:   []AuthzVariable{{Name: "name", Type: CtxVarTypeBool, Types: "", Boolean: false}},
			expected:   [][]ContextVariable{{{P: "PERMISSION"}, {Name: "name", Type: "BOOLEAN", Types: "", Boolean: falseP}}},
		},
		{
			tc:         `should return expected context variables for multiple booleans`,
			permission: "PERMISSION",
			contexts: []AuthzVariable{
				{Name: "falsey", Type: CtxVarTypeBool, Types: "", Boolean: false},
				{Name: "truey", Type: CtxVarTypeBool, Types: "", Boolean: true},
			},
			expected: [][]ContextVariable{{
				{P: "PERMISSION"},
				{Name: "falsey", Type: "BOOLEAN", Types: "", Boolean: falseP},
				{Name: "truey", Type: "BOOLEAN", Types: "", Boolean: trueP},
			}},
		},
	}
	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			request := &AuthorizationRequest{}
			request.SetPermissionVariables(test.permission, test.contexts)
			assert.Equal(t, test.expected, request.Context)
		})
	}
}

func TestSetExistingPermissions(t *testing.T) {
	permission := "PERMISSION"
	existingVariables := []AuthzVariable{
		{Name: "existing", Type: CtxVarTypeString, Types: "", Value: "hello", Values: nil},
	}

	testIO := []struct {
		tc         string
		permission string
		contexts   []AuthzVariable
		expected   [][]ContextVariable
	}{
		{
			tc:         `should expand expected context variables that matches the given permission`,
			permission: permission,
			contexts:   []AuthzVariable{{Name: "name", Type: CtxVarTypeString, Types: "", Value: "hello", Values: nil}},
			expected: [][]ContextVariable{
				{
					{P: "PERMISSION"},
					{Name: "existing", Type: "STRING", Types: "", Value: "hello", Boolean: nil},
					{Name: "name", Type: "STRING", Types: "", Value: "hello", Boolean: nil},
				},
			},
		},
		{
			tc:         `should not expand existing context variables that matches the given permission`,
			permission: "PERMISSION_2",
			contexts:   []AuthzVariable{{Name: "name", Type: CtxVarTypeString, Types: "", Value: "hello", Values: nil}},
			expected: [][]ContextVariable{
				{
					{P: "PERMISSION"},
					{Name: "existing", Type: "STRING", Types: "", Value: "hello", Boolean: nil},
				},
				{
					{P: "PERMISSION_2"},
					{Name: "name", Type: "STRING", Types: "", Value: "hello", Boolean: nil},
				},
			},
		},
	}
	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			request := &AuthorizationRequest{}
			request.SetPermissionVariables(permission, existingVariables)
			request.SetPermissionVariables(test.permission, test.contexts)
			assert.Equal(t, test.expected, request.Context)
		})
	}
}

func TestSetCommonPermission(t *testing.T) {
	existingVariables := []AuthzVariable{
		{Name: "existing", Type: CtxVarTypeString, Types: "", Value: "hello", Values: nil},
	}

	testIO := []struct {
		tc       string
		contexts []AuthzVariable
		expected [][]ContextVariable
	}{
		{
			tc: `should expand expected common context variables`,
			contexts: []AuthzVariable{
				{Name: "name", Type: CtxVarTypeString, Types: "", Value: "hello", Values: nil},
			},
			expected: [][]ContextVariable{
				{
					{P: "__COMMON__"},
					{Name: "existing", Type: "STRING", Types: "", Value: "hello", Boolean: nil},
					{Name: "name", Type: "STRING", Types: "", Value: "hello", Boolean: nil},
				},
			},
		},
		{
			tc: `should expand multiple expected context variables`,
			contexts: []AuthzVariable{
				{Name: "name", Type: CtxVarTypeString, Types: "", Value: "hello", Values: nil},
				{Name: "something-else", Type: CtxVarTypeString, Types: "", Value: "hello", Values: nil},
			},
			expected: [][]ContextVariable{
				{
					{P: "__COMMON__"},
					{Name: "existing", Type: "STRING", Types: "", Value: "hello", Boolean: nil},
					{Name: "name", Type: "STRING", Types: "", Value: "hello", Boolean: nil},
					{Name: "something-else", Type: "STRING", Types: "", Value: "hello", Boolean: nil},
				},
			},
		},
	}
	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			request := &AuthorizationRequest{}
			request.SetPermissionVariables("__COMMON__", existingVariables)
			for _, ctx := range test.contexts {
				request.SetCommonPermission(ctx)
			}
			assert.Equal(t, test.expected, request.Context)
		})
	}
}

type claimTestState struct {
	// checked should be initialized to false and the test run will set it to
	// true after the claim in the resulting principal has been checked.
	checked bool
	value   string
}

var (
	federatedClaims = Claims{
		ClaimIssuer:              []Claim{{testIssuer, ClaimIssuer, testIssuer}},
		ClaimSubject:             []Claim{{testIssuer, ClaimSubject, testSubjectID}},
		ClaimAudience:            []Claim{{testIssuer, ClaimAudience, testAudience}},
		ClaimExpires:             []Claim{{testIssuer, ClaimExpires, testExp}},
		ClaimNotBefore:           []Claim{{testIssuer, ClaimNotBefore, testNbf}},
		ClaimIssuedAt:            []Claim{{testIssuer, ClaimIssuedAt, testIat}},
		ClaimJwtID:               []Claim{{testIssuer, ClaimJwtID, testJwtID}},
		ClaimServiceName:         []Claim{{testIssuer, ClaimServiceName, testServiceName}},
		ClaimFingerprint:         []Claim{{testIssuer, ClaimFingerprint, testFPrint}},
		ClaimPrincipalType:       []Claim{{testIssuer, ClaimPrincipalType, testPType}},
		ClaimTokenType:           []Claim{{testIssuer, ClaimTokenType, testTTypeSAML}},
		ClaimTenant:              []Claim{{testIssuer, ClaimTenant, testTenantID}},
		ClaimPrincipalSubType:    []Claim{{testIssuer, ClaimPrincipalSubType, testPSType}},
		ClaimFederatedUserGroups: []Claim{{testIssuer, ClaimFederatedUserGroups, testFederatedUserGroups}},
	}

	federatedCheckMap = map[string]*claimTestState{
		"iss":    {false, testIssuer},
		"sub":    {false, testSubjectID},
		"aud":    {false, testAudience},
		"exp":    {false, testExp},
		"nbf":    {false, testNbf},
		"iat":    {false, testIat},
		"jti":    {false, testJwtID},
		"svc":    {false, testServiceName},
		"fprint": {false, testFPrint},
		"ptype":  {false, testPType},
		"ttype":  {false, testTTypeSAML},
		"tenant": {false, testTenantID},
		"pstype": {false, testPSType},
		"grps":   {false, testFederatedUserGroups},
	}

	nonFederatedClaims = Claims{
		ClaimIssuer:        []Claim{{testIssuer, ClaimIssuer, testIssuer}},
		ClaimSubject:       []Claim{{testIssuer, ClaimSubject, testSubjectID}},
		ClaimAudience:      []Claim{{testIssuer, ClaimAudience, testAudience}},
		ClaimExpires:       []Claim{{testIssuer, ClaimExpires, testExp}},
		ClaimNotBefore:     []Claim{{testIssuer, ClaimNotBefore, testNbf}},
		ClaimIssuedAt:      []Claim{{testIssuer, ClaimIssuedAt, testIat}},
		ClaimJwtID:         []Claim{{testIssuer, ClaimJwtID, testJwtID}},
		ClaimServiceName:   []Claim{{testIssuer, ClaimServiceName, testServiceName}},
		ClaimFingerprint:   []Claim{{testIssuer, ClaimFingerprint, testFPrint}},
		ClaimPrincipalType: []Claim{{testIssuer, ClaimPrincipalType, testPType}},
		ClaimTokenType:     []Claim{{testIssuer, ClaimTokenType, testTTypeLogin}},
		ClaimTenant:        []Claim{{testIssuer, ClaimTenant, testTenantID}},
	}

	nonFederatedCheckMap = map[string]*claimTestState{
		"iss":    {false, testIssuer},
		"sub":    {false, testSubjectID},
		"aud":    {false, testAudience},
		"exp":    {false, testExp},
		"nbf":    {false, testNbf},
		"iat":    {false, testIat},
		"jti":    {false, testJwtID},
		"svc":    {false, testServiceName},
		"fprint": {false, testFPrint},
		"ptype":  {false, testPType},
		"ttype":  {false, testTTypeLogin},
		"tenant": {false, testTenantID},
	}
)

func TestPrincipalToAuthorizationPrincipalClaims(t *testing.T) {
	testIO := []struct {
		tc       string
		claims   Claims
		checkMap map[string]*claimTestState
	}{
		{
			tc:       `should return an authz request principal with expected claims for a non-federated user`,
			claims:   nonFederatedClaims,
			checkMap: nonFederatedCheckMap,
		},
		{
			tc:       `should return an authz request principal with expected claims for a federated user`,
			claims:   federatedClaims,
			checkMap: federatedCheckMap,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			subject := test.claims.GetString(ClaimSubject)
			tenant := test.claims.GetString(ClaimTenant)

			p := &Principal{subject, tenant, nil, test.claims}
			requestPrincipal := principalToAuthorizationPrincipal(p)

			if test.claims == nil {
				assert.Empty(t, requestPrincipal.Claims)
			}
			for _, c := range requestPrincipal.Claims {
				claim, ok := test.checkMap[c.Key]
				assert.True(t, ok)
				assert.False(t, claim.checked)
				assert.Equal(t, claim.value, c.Value)
				assert.Equal(t, testIssuer, c.Issuer)
				claim.checked = true
			}

			for k, v := range test.checkMap {
				msg := fmt.Sprintf("did not check claim %q", k)
				assert.True(t, v.checked, msg)
			}
		})
	}
}

func TestAuthzClientAuthorize(t *testing.T) {
	testErr := errors.New("test-err")
	invalidJSON := []byte(`{test: "test"`)
	validAuthzRequest := &AuthorizationRequest{
		RequestID:     testRequestIDOriginal,
		OperationID:   testOperationID,
		CompartmentID: testCompartmentID,
		UserPrincipal: NewPrincipal(testSubjectID, testTenantID),
		RequestRegion: testRequestRegion,
		PhysicalAD:    testPhysicalAD,
		ActionKind:    testActionKind,
	}
	validAuthzRequest.SetPermissionVariables("test", []AuthzVariable{})

	// Test tag slug
	tagSlug, err := tagging.NewTagSlug(testFreeformTagSet, testDefinedTagSet)
	assert.NoError(t, err)
	assert.NotEmpty(t, tagSlug)

	// Copy the valid authz request and add a new tag slug to it
	validAuthzRequestWithNewTagSlug := validAuthzRequest
	validAuthzRequestWithNewTagSlug.SetNewTagSlug(tagSlug)

	// Copy the valid authz request and add an existing tag slug to it
	validAuthzRequestWithExistingTagSlug := validAuthzRequest
	validAuthzRequestWithExistingTagSlug.SetExistingTagSlug(tagSlug)

	testIO := []struct {
		tc            string
		client        httpsigner.SigningClient
		keyID         string
		endpoint      string
		request       *AuthorizationRequest
		expected      *AuthorizationResponse
		expectedError error
		errorExpected bool
		withTags      bool
	}{
		{
			tc:            `should handle error response from missing required authorization request variable`,
			client:        &MockSigningClient{},
			keyID:         testKeyID,
			endpoint:      "http://localhost/",
			request:       &AuthorizationRequest{},
			expectedError: ErrInvalidOperationID,
		},
		{
			tc:       `should handle error response if no permissions are set in authz request`,
			client:   &MockSigningClient{},
			keyID:    testKeyID,
			endpoint: "http://localhost/",
			request: &AuthorizationRequest{
				RequestID:     testRequestIDOriginal,
				OperationID:   testOperationID,
				CompartmentID: testCompartmentID,
				UserPrincipal: NewPrincipal(testSubjectID, testTenantID),
				RequestRegion: testRequestRegion,
				PhysicalAD:    testPhysicalAD,
				ActionKind:    testActionKind,
			},
			expectedError: ErrNoPermissionsSet,
		},
		{
			tc:            `should handle url.EscapeError from building NewRequest`,
			client:        &MockSigningClient{},
			keyID:         testKeyID,
			endpoint:      "http://%en/",
			request:       validAuthzRequest,
			expectedError: &url.Error{Op: "parse", URL: "http://%en//authorization/authorizerequest", Err: url.EscapeError("%en")},
		},
		{
			tc:            `should handle test error returned from client.Do`,
			client:        &MockSigningClient{doError: testErr},
			keyID:         testKeyID,
			endpoint:      "http://localhost/",
			request:       validAuthzRequest,
			expectedError: testErr,
		},
		{
			tc:            `should handle test error returned from ioutil.ReadAll`,
			client:        &MockSigningClient{doResponse: &http.Response{StatusCode: http.StatusOK, Body: mockReadCloser{readError: testErr}}},
			keyID:         testKeyID,
			endpoint:      "http://localhost/",
			request:       validAuthzRequest,
			expectedError: testErr,
		},
		{
			tc:            `should handle unmarshal error`,
			client:        &MockSigningClient{doResponse: validAuthzResponse(invalidJSON)},
			keyID:         testKeyID,
			endpoint:      "http://localhost/",
			request:       validAuthzRequest,
			errorExpected: true,
		},
		{
			tc:            `should handle non-200 response`,
			client:        &MockSigningClient{doResponse: non200Response},
			keyID:         testKeyID,
			endpoint:      "http://localhost/",
			request:       validAuthzRequest,
			expectedError: &ServiceResponseError{non200Response},
		},
		{
			tc:       `should return expected authorization response`,
			client:   &MockSigningClient{doResponse: validAuthzResponse(validAuthorizationResponseJSON)},
			keyID:    testKeyID,
			endpoint: "http://localhost/",
			expected: validAuthzResponseWithContext(AuthorizationContext{RequestID: testRequestID}),
			request:  validAuthzRequest,
		},
		{
			tc: `should return expected authorization response with no request ID in the response header`,
			client: &MockSigningClient{
				doResponse: &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewBuffer(validAuthorizationResponseJSON))},
			},
			keyID:    testKeyID,
			endpoint: "http://localhost/",
			expected: validAuthzResponseWithContext(AuthorizationContext{RequestID: testRequestIDOriginal}),
			request:  validAuthzRequest,
		},
		{
			tc:       `should return expected authorization response with a new tag slug`,
			client:   &MockSigningClient{doResponse: validAuthzResponse(authzResponseJSON(validAuthzRespWithNewTagSlug))},
			keyID:    testKeyID,
			endpoint: "http://localhost/",
			expected: validAuthzResponseWithTags(*validAuthzRespWithNewTagSlug),
			withTags: true,
			request:  validAuthzRequestWithNewTagSlug,
		},
		{
			tc:       `should return expected authorization response with a new and existing tag slug`,
			client:   &MockSigningClient{doResponse: validAuthzResponse(authzResponseJSON(validAuthzRespWithExistingTagSlug))},
			keyID:    testKeyID,
			endpoint: "http://localhost/",
			expected: validAuthzResponseWithTags(*validAuthzRespWithExistingTagSlug),
			withTags: true,
			request:  validAuthzRequestWithExistingTagSlug,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			var c AuthorizationClient
			if test.withTags {
				c = NewAuthorizationClientWithTags(test.client, test.endpoint)
			} else {
				c = NewAuthorizationClient(test.client, test.endpoint)
			}
			r, e := c.(*authorizationClient).authorize(test.request)

			// Error expected, but cannot compare directly
			if test.errorExpected {
				assert.Nil(t, r)
				assert.NotNil(t, e)

				// Perform comparison on the expected error
			} else if test.expectedError != nil {
				assert.Nil(t, r)
				assert.Equal(t, e, test.expectedError)

				// Perform comparison on the expected result
			} else {
				assert.Nil(t, e)
				assert.Equal(t, test.expected, r)
			}
		})
	}
}

func TestAll(t *testing.T) {
	var testSigner httpsigner.RequestSigner = &MockRequestSigner{}
	client := NewSigningClient(testSigner, testKeyID)
	authzResponse := validAuthorizationResponse
	authzResponse.OutboundAuthorizationRequest.Context = validContextVariable
	responseJSON, _ := json.Marshal(authzResponse)

	testIO := []struct {
		tc               string
		input            []string
		grantExpected    bool
		responseJSON     []byte
		expectedResponse *AuthorizationResponse
		expectedError    error
	}{
		{
			tc:               `should deny request if not all of requested permissions were granted`,
			input:            []string{`permission-1`, `permission-2`, `permission-3`},
			responseJSON:     responseJSON,
			grantExpected:    false,
			expectedResponse: authzResponseWithContext(authzResponse, authzContext2Permissions),
		},
		{
			tc:               `should grant request if all of requested permissions were granted`,
			input:            []string{`permission-1`, `permission-2`},
			responseJSON:     responseJSON,
			grantExpected:    true,
			expectedResponse: authzResponseWithContext(authzResponse, authzContext2Permissions),
		},
		{
			tc:               `should deny request if no permissions are requested`,
			input:            []string{},
			responseJSON:     responseJSON,
			grantExpected:    false,
			expectedResponse: nil,
			expectedError:    ErrNoPermissionsSet,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			client := NewAuthorizationClient(client, "http://localhost")
			client.(*authorizationClient).client = &MockSigningClient{doResponse: validAuthzResponse(test.responseJSON)}
			request, err := NewAuthorizationRequest(testRequestID, testOperationID, testCompartmentID, testServiceName,
				NewPrincipal(testSubjectID, testTenantID), nil, testRequestRegion, testPhysicalAD)
			assert.Nil(t, err)
			err = request.SetActionKind(testActionKind)
			assert.Nil(t, err)
			for _, p := range test.input {
				request.SetPermissionVariables(p, []AuthzVariable{{Name: "sup"}})
			}
			// Attach the test request to the expected AuthorizationResponse if an error isn't expected
			if test.expectedError == nil {
				test.expectedResponse.AuthorizationRequest = request
			}
			granted, response, err := client.All(request)
			assert.Equal(t, test.grantExpected, granted)
			assert.Equal(t, test.expectedResponse, response)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestAllMissingRequiredVariables(t *testing.T) {
	var testSigner httpsigner.RequestSigner = &MockRequestSigner{}
	client := NewSigningClient(testSigner, testKeyID)
	t.Run(`should return false and nil if authz request is missing required variables`, func(t *testing.T) {
		client := NewAuthorizationClient(client, "http://localhost")
		request := &AuthorizationRequest{}
		granted, response, err := client.All(request)
		assert.Equal(t, false, granted)
		assert.Equal(t, err, ErrInvalidOperationID)
		assert.Nil(t, response)
	})
}

func TestAny(t *testing.T) {
	var testSigner httpsigner.RequestSigner = &MockRequestSigner{}
	client := NewSigningClient(testSigner, testKeyID)
	authzResponse := validAuthorizationResponse
	authzResponse.OutboundAuthorizationRequest.Context = validContextVariable
	responseJSON, _ := json.Marshal(authzResponse)

	testIO := []struct {
		tc               string
		input            []string
		grantExpected    bool
		responseJSON     []byte
		expectedResponse *AuthorizationResponse
		expectedError    error
	}{
		{
			tc:    `should deny request if no requested permissions were granted`,
			input: []string{`permission-1000`}, responseJSON: responseJSON,
			grantExpected:    false,
			expectedResponse: authzResponseWithContext(authzResponse, authzContext0Permission),
		},
		{
			tc:    `should grant request if one of requested permissions were granted`,
			input: []string{`permission-1`}, responseJSON: responseJSON,
			grantExpected:    true,
			expectedResponse: authzResponseWithContext(authzResponse, authzContext1Permission),
		},
		{
			tc:               `should grant request if all of requested permissions were granted`,
			input:            []string{`permission-1`, `permission-2`},
			responseJSON:     responseJSON,
			grantExpected:    true,
			expectedResponse: authzResponseWithContext(authzResponse, authzContext2Permissions),
		},
		{
			tc:               `should deny request if no permissions are requested`,
			input:            []string{},
			responseJSON:     responseJSON,
			grantExpected:    false,
			expectedResponse: nil,
			expectedError:    ErrNoPermissionsSet,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			client := NewAuthorizationClient(client, "http://localhost")
			client.(*authorizationClient).client = &MockSigningClient{doResponse: validAuthzResponse(test.responseJSON)}

			request, err := NewAuthorizationRequest(testRequestID, testOperationID, testCompartmentID, testServiceName,
				NewPrincipal(testSubjectID, testTenantID), nil, testRequestRegion, testPhysicalAD)
			assert.Nil(t, err)
			err = request.SetActionKind(testActionKind)
			assert.Nil(t, err)
			for _, p := range test.input {
				request.SetPermissionVariables(p, []AuthzVariable{{}})
			}
			// Attach the test request to the expected AuthorizationResponse if an error isn't expected
			if test.expectedError == nil {
				test.expectedResponse.AuthorizationRequest = request
			}
			granted, response, err := client.Any(request)
			assert.Equal(t, test.grantExpected, granted)
			assert.Equal(t, test.expectedResponse, response)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestAnyMissingRequiredVariables(t *testing.T) {
	var testSigner httpsigner.RequestSigner = &MockRequestSigner{}
	client := NewSigningClient(testSigner, testKeyID)
	t.Run(`should return false and nil if authz request is missing required variables`, func(t *testing.T) {
		client := NewAuthorizationClient(client, "http://localhost")
		request := &AuthorizationRequest{}
		granted, response, err := client.Any(request)
		assert.Equal(t, false, granted)
		assert.Nil(t, response)
		assert.Equal(t, err, ErrInvalidOperationID)
	})
}

func TestFilter(t *testing.T) {
	var testSigner httpsigner.RequestSigner = &MockRequestSigner{}
	client := NewSigningClient(testSigner, testKeyID)
	authzResponse := generateResponseFromPermissions([]string{`permission-1`, `permission-2`})
	responseJSON, _ := json.Marshal(authzResponse)
	noPermissionResponse := generateResponseFromPermissions([]string{})
	noPermissionJSON, _ := json.Marshal(noPermissionResponse)

	testIO := []struct {
		tc               string
		input            []string
		responseJSON     []byte
		permissions      []string
		expectedError    error
		expectedResponse *AuthorizationResponse
	}{
		{
			tc:               `should return an empty set if we get no permissions back from the service`,
			input:            []string{`permission-100`},
			responseJSON:     noPermissionJSON,
			permissions:      []string{},
			expectedResponse: authzResponseWithContext(noPermissionResponse, authzContext0Permission),
		},
		{
			tc:               `should return set that contains granted permissions if not all requested permissions were granted`,
			input:            []string{`permission-1`},
			responseJSON:     responseJSON,
			permissions:      []string{`permission-1`},
			expectedResponse: authzResponseWithContext(authzResponse, authzContext1Permission),
		},
		{
			tc:               `should return set that contains granted permissions when all request permissions were granted`,
			input:            []string{`permission-1`, `permission-2`},
			responseJSON:     responseJSON,
			permissions:      []string{`permission-1`, `permission-2`},
			expectedResponse: authzResponseWithContext(authzResponse, authzContext2Permissions),
		},
		{
			tc:               `should return empty set with error if no permissions are requested`,
			input:            []string{},
			responseJSON:     responseJSON,
			permissions:      []string{},
			expectedResponse: nil,
			expectedError:    ErrNoPermissionsSet,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			client := NewAuthorizationClient(client, "http://localhost")
			client.(*authorizationClient).client = &MockSigningClient{doResponse: validAuthzResponse(test.responseJSON)}

			request, err := NewAuthorizationRequest(testRequestID, testOperationID, testCompartmentID, testServiceName,
				NewPrincipal(testSubjectID, testTenantID), nil, testRequestRegion, testPhysicalAD)
			assert.Nil(t, err)
			err = request.SetActionKind(testActionKind)
			assert.Nil(t, err)
			for _, p := range test.input {
				request.SetPermissionVariables(p, []AuthzVariable{})
			}
			// Attach the test request to the expected AuthorizationResponse if an error isn't expected
			if test.expectedError == nil {
				test.expectedResponse.AuthorizationRequest = request
			}
			permissions, response, err := client.Filter(request)
			assert.Equal(t, permissions, test.permissions)
			assert.Equal(t, response, test.expectedResponse)
			assert.Equal(t, err, test.expectedError)
		})
	}
}

func TestFilterMissingRequiredVariables(t *testing.T) {
	var testSigner httpsigner.RequestSigner = &MockRequestSigner{}
	client := NewSigningClient(testSigner, testKeyID)
	t.Run(`should return empty list and nil if authz request is missing required variables`, func(t *testing.T) {
		client := NewAuthorizationClient(client, "http://localhost")
		request := &AuthorizationRequest{}
		permissions, response, err := client.Filter(request)
		assert.Equal(t, []string{}, permissions)
		assert.Nil(t, response)
		assert.Equal(t, err, ErrInvalidOperationID)
	})
}

func TestAuthorizationResponseAny(t *testing.T) {
	testIO := []struct {
		tc          string
		contextVars [][]ContextVariable
		input       []string
		isGranted   bool
	}{
		{
			tc:          `should deny request if no requested permissions were granted`,
			contextVars: validContextVariable,
			input:       []string{`permission-1000`},
			isGranted:   false,
		},
		{
			tc:          `should grant request if one of requested permissions were granted`,
			contextVars: validContextVariable,
			input:       []string{`permission-1`},
			isGranted:   true,
		},
		{
			tc:          `should grant request if all of requested permissions were granted`,
			contextVars: validContextVariable,
			input:       []string{`permission-1`, `permission-2`},
			isGranted:   true,
		},
		{
			tc:          `should deny request if no permissions are requested`,
			contextVars: validContextVariable,
			input:       []string{},
			isGranted:   false,
		},
		{
			tc:          `should deny request if user has no authorized permissions`,
			contextVars: validEmptyContextVariable,
			input:       []string{`permission-1`},
			isGranted:   false,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			authzResponse := validAuthorizationResponse
			authzResponse.OutboundAuthorizationRequest.Context = test.contextVars

			request, err := NewAuthorizationRequest(testRequestID, testOperationID, testCompartmentID, testServiceName,
				NewPrincipal(testSubjectID, testTenantID), nil, testRequestRegion, testPhysicalAD)
			assert.Nil(t, err)
			for _, p := range test.input {
				request.SetPermissionVariables(p, []AuthzVariable{})
			}
			authzResponse.AuthorizationRequest = request

			isGranted := authzResponse.Any()
			assert.Equal(t, test.isGranted, isGranted)
		})
	}
}

func TestAuthorizationResponseAll(t *testing.T) {
	testIO := []struct {
		tc          string
		contextVars [][]ContextVariable
		input       []string
		isGranted   bool
	}{
		{
			tc:          `should deny request if not all of requested permissions were granted`,
			contextVars: validContextVariable,
			input:       []string{`permission-1`, `permission-2`, `permission-3`},
			isGranted:   false,
		},
		{
			tc:          `should grant request if all of requested permissions were granted`,
			contextVars: validContextVariable,
			input:       []string{`permission-1`, `permission-2`},
			isGranted:   true,
		},
		{
			tc:          `should deny request if no permissions are requested`,
			contextVars: validContextVariable,
			input:       []string{},
			isGranted:   false,
		},
		{
			tc:          `should deny request if user has no authorized permissions`,
			contextVars: validEmptyContextVariable,
			input:       []string{`permission-1`},
			isGranted:   false,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			authzResponse := validAuthorizationResponse
			authzResponse.OutboundAuthorizationRequest.Context = test.contextVars

			request, err := NewAuthorizationRequest(testRequestID, testOperationID, testCompartmentID, testServiceName,
				NewPrincipal(testSubjectID, testTenantID), nil, testRequestRegion, testPhysicalAD)
			assert.Nil(t, err)

			for _, p := range test.input {
				request.SetPermissionVariables(p, []AuthzVariable{})
			}
			authzResponse.AuthorizationRequest = request

			isGranted := authzResponse.All()
			assert.Equal(t, test.isGranted, isGranted)
		})
	}
}

func TestAuthorizationResponseFilter(t *testing.T) {
	testIO := []struct {
		tc                   string
		contextVars          [][]ContextVariable
		input                []string
		expectedIntersection []string
	}{
		{
			tc:                   `should return the two authorized permissions, and not the single unauthorized one`,
			contextVars:          validContextVariable,
			input:                []string{`permission-1`, `permission-2`, `permission-3`},
			expectedIntersection: []string{`permission-1`, `permission-2`},
		},
		{
			tc:                   `should return both authorized permissions`,
			contextVars:          validContextVariable,
			input:                []string{`permission-1`, `permission-2`},
			expectedIntersection: []string{`permission-1`, `permission-2`},
		},
		{
			tc:                   `should return the single authorized permission requested out of two authorized permissions`,
			contextVars:          validContextVariable,
			input:                []string{`permission-2`},
			expectedIntersection: []string{`permission-2`},
		},
		{
			tc:                   `should return no permissions since the requested permission is not among the authorized`,
			contextVars:          validContextVariable,
			input:                []string{`permission-3`},
			expectedIntersection: []string{},
		},
		{
			tc:                   `should return no permissions because no permissions were requested`,
			contextVars:          validContextVariable,
			input:                []string{},
			expectedIntersection: []string{},
		},
		{
			tc:                   `should deny request if user has no authorized permissions`,
			contextVars:          validEmptyContextVariable,
			input:                []string{`permission-1`},
			expectedIntersection: []string{},
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			authzResponse := validAuthorizationResponse
			authzResponse.OutboundAuthorizationRequest.Context = test.contextVars

			request, err := NewAuthorizationRequest(testRequestID, testOperationID, testCompartmentID, testServiceName,
				NewPrincipal(testSubjectID, testTenantID), nil, testRequestRegion, testPhysicalAD)
			assert.Nil(t, err)

			for _, p := range test.input {
				request.SetPermissionVariables(p, []AuthzVariable{})
			}
			authzResponse.AuthorizationRequest = request

			sec := authzResponse.Filter()
			assert.Equal(t, test.expectedIntersection, sec)
		})
	}
}

func TestAuthorizationResponseSet(t *testing.T) {
	testIO := []struct {
		tc          string
		contextVars [][]ContextVariable
		input       []string
		isGranted   bool
	}{
		{
			tc:          `should deny request when missing a requested permission`,
			contextVars: validContextVariable,
			input:       []string{`permission-1`, `permission-2`, `permission-3`},
			isGranted:   false,
		},
		{
			tc:          `should grant request when requested permissions match authorized permissions`,
			contextVars: validContextVariable,
			input:       []string{`permission-1`, `permission-2`},
			isGranted:   true,
		},
		{
			tc:          `should grant request when requested permissions are a subset of authorized permissions`,
			contextVars: validContextVariable,
			input:       []string{`permission-2`},
			isGranted:   true,
		},
		{
			tc:          `should deny request when requested permissions are not a subset of the authorized permissions`,
			contextVars: validContextVariable,
			input:       []string{`permission-3`},
			isGranted:   false,
		},
		{
			tc:          `should deny request when requested permissions are empty`,
			contextVars: validContextVariable,
			input:       []string{},
			isGranted:   false,
		},
		{
			tc:          `should deny request if there are no authorized permissions`,
			contextVars: validEmptyContextVariable,
			input:       []string{`permission-1`},
			isGranted:   false,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			authzResponse := validAuthorizationResponse
			authzResponse.OutboundAuthorizationRequest.Context = test.contextVars

			request, err := NewAuthorizationRequest(testRequestID, testOperationID, testCompartmentID, testServiceName,
				NewPrincipal(testSubjectID, testTenantID), nil, testRequestRegion, testPhysicalAD)
			assert.Nil(t, err)

			for _, p := range test.input {
				request.SetPermissionVariables(p, []AuthzVariable{})
			}
			authzResponse.AuthorizationRequest = request

			isGranted := authzResponse.Set(test.input)
			assert.Equal(t, test.isGranted, isGranted)
		})
	}
}

func TestMakeAuthorizationCall(t *testing.T) {
	testErr := errors.New("test-err")
	invalidJSON := []byte(`{test: "test"`)
	validAuthzRequest := &AuthorizationRequest{
		RequestID:     testRequestIDOriginal,
		OperationID:   testOperationID,
		CompartmentID: testCompartmentID,
		UserPrincipal: NewPrincipal(testSubjectID, testTenantID),
		RequestRegion: testRequestRegion,
		PhysicalAD:    testPhysicalAD,
		ActionKind:    testActionKind,
	}

	validAuthzRequest.SetPermissionVariables("test", []AuthzVariable{})

	testIO := []struct {
		tc            string
		client        httpsigner.SigningClient
		keyID         string
		endpoint      string
		request       *AuthorizationRequest
		expected      *AuthorizationResponse
		expectedError error
		errorExpected bool
	}{
		{
			tc:            `should handle error response from missing required authorization request variable`,
			client:        &MockSigningClient{},
			keyID:         testKeyID,
			endpoint:      "http://localhost/",
			request:       &AuthorizationRequest{},
			expectedError: ErrInvalidOperationID,
		},
		{
			tc:       `should handle error response if no permissions are set in authz request`,
			client:   &MockSigningClient{},
			keyID:    testKeyID,
			endpoint: "http://localhost/",
			request: &AuthorizationRequest{
				RequestID:     testRequestIDOriginal,
				OperationID:   testOperationID,
				CompartmentID: testCompartmentID,
				UserPrincipal: NewPrincipal(testSubjectID, testTenantID),
				RequestRegion: testRequestRegion,
				PhysicalAD:    testPhysicalAD,
				ActionKind:    testActionKind,
			},
			expectedError: ErrNoPermissionsSet,
		},
		{
			tc:            `should handle url.EscapeError from building NewRequest`,
			client:        &MockSigningClient{},
			keyID:         testKeyID,
			endpoint:      "http://%en/",
			request:       validAuthzRequest,
			expectedError: &url.Error{Op: "parse", URL: "http://%en//authorization/authorizerequest", Err: url.EscapeError("%en")},
		},
		{
			tc:            `should handle test error returned from client.Do`,
			client:        &MockSigningClient{doError: testErr},
			keyID:         testKeyID,
			endpoint:      "http://localhost/",
			request:       validAuthzRequest,
			expectedError: testErr,
		},
		{
			tc:            `should handle test error returned from ioutil.ReadAll`,
			client:        &MockSigningClient{doResponse: &http.Response{StatusCode: http.StatusOK, Body: mockReadCloser{readError: testErr}}},
			keyID:         testKeyID,
			endpoint:      "http://localhost/",
			request:       validAuthzRequest,
			expectedError: testErr,
		},
		{
			tc:            `should handle unmarshal error`,
			client:        &MockSigningClient{doResponse: validAuthzResponse(invalidJSON)},
			keyID:         testKeyID,
			endpoint:      "http://localhost/",
			request:       validAuthzRequest,
			errorExpected: true,
		},
		{
			tc:            `should handle non-200 response`,
			client:        &MockSigningClient{doResponse: non200Response},
			keyID:         testKeyID,
			endpoint:      "http://localhost/",
			request:       validAuthzRequest,
			expectedError: &ServiceResponseError{non200Response},
		},
		{
			tc:       `should return expected authorization response`,
			client:   &MockSigningClient{doResponse: validAuthzResponse(validAuthorizationResponseJSON)},
			keyID:    testKeyID,
			endpoint: "http://localhost/",
			expected: validAuthzResponseWithContext(AuthorizationContext{RequestID: testRequestID}),
			request:  validAuthzRequest,
		},
		{
			tc: `should return expected authorization response with no request ID in the response header`,
			client: &MockSigningClient{
				doResponse: &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewBuffer(validAuthorizationResponseJSON))},
			},
			keyID:    testKeyID,
			endpoint: "http://localhost/",
			expected: validAuthzResponseWithContext(AuthorizationContext{RequestID: testRequestIDOriginal}),
			request:  validAuthzRequest,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			c := NewAuthorizationClient(test.client, test.endpoint)

			r, e := c.(*authorizationClient).MakeAuthorizationCall(test.request)

			// Error expected, but cannot compare directly
			if test.errorExpected {
				assert.Nil(t, r)
				assert.NotNil(t, e)
			} else if test.expectedError != nil {
				// Perform comparison on the expected error
				assert.Nil(t, r)
				assert.Equal(t, e, test.expectedError)
			} else {
				// Perform comparison on the expected result
				assert.Nil(t, e)
				test.expected.AuthorizationRequest = validAuthzRequest
				assert.Equal(t, test.expected, r)
			}
		})
	}
}

func TestMakeAssociationAuthorizationCall(t *testing.T) {
	testErr := errors.New("test-err")
	invalidJSON := []byte(`{test: "test"`)

	validAuthzRequest := AuthorizationRequest{
		RequestID:     testRequestIDOriginal,
		OperationID:   testOperationID,
		CompartmentID: testCompartmentID,
		UserPrincipal: NewPrincipal(testSubjectID, testTenantID),
		RequestRegion: testRequestRegion,
		PhysicalAD:    testPhysicalAD,
		ActionKind:    testActionKind,
	}
	validAuthzRequest.SetPermissionVariables("test", []AuthzVariable{})

	expectedAuthzRequest := AuthorizationRequest{
		RequestID:     testRequestIDOriginal,
		OperationID:   testOperationID,
		CompartmentID: testCompartmentID,
		UserPrincipal: NewPrincipal(testSubjectID, testTenantID),
		RequestRegion: testRequestRegion,
		PhysicalAD:    testPhysicalAD,
		ActionKind:    testActionKind,
	}
	expectedAuthzRequest.SetPermissionVariables("test", []AuthzVariable{})
	expectedAuthzRequest.SetCommonPermissions()
	expectedAssociationResponse := validAssociationAuthorizationResponse
	expectedAssociationResponse.Responses[0].AuthorizationRequest = &expectedAuthzRequest
	expectedAssociationResponse.Responses[1].AuthorizationRequest = &expectedAuthzRequest

	testIO := []struct {
		tc            string
		client        httpsigner.SigningClient
		keyID         string
		endpoint      string
		request       *AssociationAuthorizationRequest
		expected      *AssociationAuthorizationResponse
		expectedError error
		errorExpected bool
	}{
		{
			tc:            `should error when including no assigned authorization`,
			client:        &MockSigningClient{},
			keyID:         testKeyID,
			endpoint:      "http://localhost/",
			request:       &AssociationAuthorizationRequest{},
			expectedError: ErrAssociationInsufficientRequests,
		},
		{
			tc:            `should error when including just one assigned authorization request`,
			client:        &MockSigningClient{},
			keyID:         testKeyID,
			endpoint:      "http://localhost/",
			request:       &AssociationAuthorizationRequest{validAuthzRequest},
			expectedError: ErrAssociationInsufficientRequests,
		},
		{
			tc:       `should return error if no permissions are set in one authz request as a subset of AssociationAuthorizationRequest`,
			client:   &MockSigningClient{},
			keyID:    testKeyID,
			endpoint: "http://localhost/",
			request: &AssociationAuthorizationRequest{{
				RequestID:     testRequestIDOriginal,
				OperationID:   testOperationID,
				CompartmentID: testCompartmentID,
				UserPrincipal: NewPrincipal(testSubjectID, testTenantID),
				RequestRegion: testRequestRegion,
				PhysicalAD:    testPhysicalAD,
				ActionKind:    testActionKind},
				validAuthzRequest,
			},
			expectedError: ErrNoPermissionsSet,
		},
		{
			tc:            `should handle url.EscapeError from building NewRequest`,
			client:        &MockSigningClient{},
			keyID:         testKeyID,
			endpoint:      "http://%en/",
			request:       &AssociationAuthorizationRequest{validAuthzRequest, validAuthzRequest},
			expectedError: &url.Error{Op: "parse", URL: "http://%en//authorization/associaterequest", Err: url.EscapeError("%en")},
		},
		{
			tc:            `should handle test error returned from client.Do`,
			client:        &MockSigningClient{doError: testErr},
			keyID:         testKeyID,
			endpoint:      "http://localhost/",
			request:       &AssociationAuthorizationRequest{validAuthzRequest, validAuthzRequest},
			expectedError: testErr,
		},
		{
			tc:            `should handle test error returned from ioutil.ReadAll`,
			client:        &MockSigningClient{doResponse: &http.Response{StatusCode: http.StatusOK, Body: mockReadCloser{readError: testErr}}},
			keyID:         testKeyID,
			endpoint:      "http://localhost/",
			request:       &AssociationAuthorizationRequest{validAuthzRequest, validAuthzRequest},
			expectedError: testErr,
		},
		{
			tc:            `should handle unmarshal error`,
			client:        &MockSigningClient{doResponse: validAssociationResponseHTTP(invalidJSON)},
			keyID:         testKeyID,
			endpoint:      "http://localhost/",
			request:       &AssociationAuthorizationRequest{validAuthzRequest, validAuthzRequest},
			errorExpected: true,
		},
		{
			tc:            `should handle non-200 response`,
			client:        &MockSigningClient{doResponse: non200Response},
			keyID:         testKeyID,
			endpoint:      "http://localhost/",
			request:       &AssociationAuthorizationRequest{validAuthzRequest, validAuthzRequest},
			expectedError: &ServiceResponseError{non200Response},
		},
		{
			tc:            `should return ErrUnexpectedAssociationAuthzResponseLength if the response length does not match the number of authz requests`,
			client:        &MockSigningClient{doResponse: validAssociationResponseHTTP(validAssociationAuthorizationResponseJSON)},
			keyID:         testKeyID,
			endpoint:      "http://localhost/",
			request:       &AssociationAuthorizationRequest{validAuthzRequest, validAuthzRequest, validAuthzRequest},
			expectedError: ErrUnexpectedAssociationAuthzResponseLength,
		},
		{
			tc:       `should return expected authorization response`,
			client:   &MockSigningClient{doResponse: validAssociationResponseHTTP(validAssociationAuthorizationResponseJSON)},
			keyID:    testKeyID,
			endpoint: "http://localhost/",
			expected: &expectedAssociationResponse,
			request:  &AssociationAuthorizationRequest{validAuthzRequest, validAuthzRequest},
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			c := NewAssociationAuthorizationClient(test.client, test.endpoint)
			r, e := c.(*authorizationClient).MakeAssociationAuthorizationCall(test.request)

			// Error expected, but cannot compare directly
			if test.errorExpected {
				assert.Nil(t, r)
				assert.NotNil(t, e)
			} else if test.expectedError != nil {
				// Perform comparison on the expected error
				assert.Nil(t, r)
				assert.Equal(t, e, test.expectedError)
			} else {
				// Perform comparison on the expected result
				assert.Nil(t, e)
				assert.Equal(t, test.expected, r)
			}
		})
	}
}

func TestActionKindIsChange(t *testing.T) {
	isChange := map[ActionKind]bool{
		ActionKindCreate:           true,
		ActionKindUpdate:           true,
		ActionKindDelete:           true,
		ActionKindList:             true,
		ActionKindAttach:           true,
		ActionKindDetach:           true,
		ActionKindOther:            true,
		ActionKindNotDefined:       true,
		ActionKindSearch:           true,
		ActionKindUpdateRemoveOnly: true,
		ActionKindRead:             false}
	t.Run(`IsChange should only return true for Actions which initiate a change.`, func(t *testing.T) {
		for _, actionKind := range allActionKinds {
			isItAChange := actionKind.IsChange()
			assert.Equal(t, isChange[actionKind], isItAChange)
		}
	})
}

func TestActionKindIsSearch(t *testing.T) {
	isSearch := map[ActionKind]bool{
		ActionKindCreate:           false,
		ActionKindUpdate:           false,
		ActionKindDelete:           false,
		ActionKindList:             false,
		ActionKindAttach:           false,
		ActionKindDetach:           false,
		ActionKindOther:            false,
		ActionKindNotDefined:       false,
		ActionKindSearch:           true,
		ActionKindUpdateRemoveOnly: false,
		ActionKindRead:             false}
	t.Run(`IsSearch should only return true for Actions which are searches.`, func(t *testing.T) {
		for _, actionKind := range allActionKinds {
			isItASearch := actionKind.IsSearch()
			assert.Equal(t, isSearch[actionKind], isItASearch)
		}
	})
}

func TestActionKindIsCreate(t *testing.T) {
	isCreate := map[ActionKind]bool{
		ActionKindCreate:           true,
		ActionKindUpdate:           false,
		ActionKindDelete:           false,
		ActionKindList:             false,
		ActionKindAttach:           false,
		ActionKindDetach:           false,
		ActionKindOther:            false,
		ActionKindNotDefined:       false,
		ActionKindSearch:           false,
		ActionKindUpdateRemoveOnly: false,
		ActionKindRead:             false}
	t.Run(`isCreate should only return true for Actions which create.`, func(t *testing.T) {
		for _, actionKind := range allActionKinds {
			isItACreate := actionKind.IsCreate()
			assert.Equal(t, isCreate[actionKind], isItACreate)
		}
	})
}

func TestActionKindIsDeleteFriendly(t *testing.T) {
	isDeleteFriendly := map[ActionKind]bool{
		ActionKindCreate:           false,
		ActionKindUpdate:           false,
		ActionKindDelete:           true,
		ActionKindList:             true,
		ActionKindAttach:           false,
		ActionKindDetach:           true,
		ActionKindOther:            false,
		ActionKindNotDefined:       false,
		ActionKindSearch:           true,
		ActionKindUpdateRemoveOnly: true,
		ActionKindRead:             true}
	t.Run(`isDeleteFriendly should only return true for Actions which are delete friendly.`, func(t *testing.T) {
		for _, actionKind := range allActionKinds {
			isItDeleteFriendly := actionKind.IsDeleteFriendly()
			assert.Equal(t, isDeleteFriendly[actionKind], isItDeleteFriendly)
		}
	})
}

func TestGetTopActionKind(t *testing.T) {

	testIO := []struct {
		tc     string
		input  []ActionKind
		result ActionKind
	}{
		{tc: `2 invalid action kinds should return ActionKindNotDefined`,
			input: []ActionKind{ActionKind("lies"), ActionKind("videotape")}, result: ActionKindNotDefined,
		},
		{tc: `1 invalid action, 1 valid kinds should return the valid one: ActionKindDelete`,
			input: []ActionKind{ActionKindDelete, ActionKind("videotape")}, result: ActionKindDelete,
		},
		{tc: `2 valid kinds should return the valid one: ActionKindCreate should prevail`,
			input: []ActionKind{ActionKindCreate, ActionKindUpdate}, result: ActionKindCreate,
		},
		{tc: `2 valid kinds with wonky caps ActionKindCreate should prevail`,
			input: []ActionKind{ActionKind("CreatE"), ActionKind("UpDaTe")}, result: ActionKindCreate,
		},
		{tc: `1 valid kind should return self`,
			input: []ActionKind{ActionKindUpdate}, result: ActionKindUpdate,
		},
		{tc: `1 valid kind with wonky capitalization should return ActionKindUpdate`,
			input: []ActionKind{ActionKind("uPdAtE")}, result: ActionKindUpdate,
		},
		{tc: `1 invalid kind should return ActionKindNotDefined`,
			input: []ActionKind{ActionKind("wtf")}, result: ActionKindNotDefined,
		},
		{tc: `5 valid kind ActionKindDetach should prevail`,
			input: []ActionKind{ActionKindDetach, ActionKindUpdateRemoveOnly, ActionKindList, ActionKindRead, ActionKindSearch}, result: ActionKindDetach,
		},
		{tc: `same one twice, should result self`,
			input: []ActionKind{ActionKindDetach, ActionKindDetach}, result: ActionKindDetach,
		},
		{tc: `2 wonkys of ActionKindDetach, should return ActionKindDetach`,
			input: []ActionKind{ActionKind("Detach"), ActionKind("dETACH")}, result: ActionKindDetach,
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			topActionkind := GetTopActionKind(test.input)
			assert.Equal(t, test.result, topActionkind)

		})
	}
}

func TestAssociationAuthorizationResultIsSuccess(t *testing.T) {
	isSuccess := map[AssociationAuthorizationResult]bool{
		AssociationFailUnknown:        false,
		AssociationFailBadRequest:     false,
		AssociationFailMissingEndorse: false,
		AssociationFailMissingAdmit:   false,
		AssociationSuccess:            true,
	}
	t.Run(`isSuccess should only return true for AssociationAuthorizationResult which are successful.`, func(t *testing.T) {
		for result := range isSuccess {
			isItSuccessful := result.IsSuccess()
			assert.Equal(t, isSuccess[result], isItSuccessful)
		}
	})
}

func TestMarshalContextVariable(t *testing.T) {
	trueP := new(bool)
	*trueP = true

	testIO := []struct {
		tc       string
		ctxVar   ContextVariable
		expected interface{}
		errStr   string
	}{
		{
			tc: `should marshal and unmarshal permission context variable as expected`,
			ctxVar: ContextVariable{
				P: "permission",
			},
			expected: ctxVarPermission{
				P: "permission",
			},
		},
		{
			tc: `should marshal and unmarshal non-empty CtxVarTypeString as expected`,
			ctxVar: ContextVariable{
				Name:  "foo",
				Type:  CtxVarTypeString,
				Value: "bar",
			},
			expected: ctxVarDefault{
				Name:  "foo",
				Type:  CtxVarTypeString,
				Value: "bar",
			},
		},
		{
			tc: `should marshal and unmarshal CtxVarTypeString with empty fields`,
			ctxVar: ContextVariable{
				Type: CtxVarTypeString,
			},
			expected: ctxVarDefault{
				Name:  "",
				Type:  CtxVarTypeString,
				Value: "",
			},
		},
		{
			tc: `should marshal and unmarshal non-empty CtxVarTypeBool as expected`,
			ctxVar: ContextVariable{
				Name:    "foo",
				Type:    CtxVarTypeBool,
				Boolean: trueP,
			},
			expected: ctxVarBoolean{
				Name:    "foo",
				Type:    CtxVarTypeBool,
				Boolean: trueP,
			},
		},
		{
			tc: `should marshal and unmarshal CtxVarTypeBool with empty fields`,
			ctxVar: ContextVariable{
				Type: CtxVarTypeBool,
			},
			expected: ctxVarBoolean{
				Name:    "",
				Type:    CtxVarTypeBool,
				Boolean: nil,
			},
		},
		{
			tc: `should marshal and unmarshal empty CtxVarTypeInteger as expected`,
			ctxVar: ContextVariable{
				Type: CtxVarTypeInt,
			},
			expected: ctxVarDefault{
				Name:  "",
				Type:  CtxVarTypeInt,
				Value: "",
			},
		},
		{
			tc: `should marshal and unmarshal non-empty CtxVarTypeInteger as expected`,
			ctxVar: ContextVariable{
				Name:  "foo",
				Type:  CtxVarTypeInt,
				Value: "123",
			},
			expected: ctxVarDefault{
				Name:  "foo",
				Type:  CtxVarTypeInt,
				Value: "123",
			},
		},
		{
			tc: `should marshal and unmarshal non-empty CtxVarTypeList as expected`,
			ctxVar: ContextVariable{
				Name:   "foo",
				Type:   CtxVarTypeList,
				Types:  CtxVarTypeString,
				Values: []string{"bar"},
			},
			expected: ctxVarList{
				Name:   "foo",
				Type:   CtxVarTypeList,
				Types:  CtxVarTypeString,
				Values: []string{"bar"},
			},
		},
		{
			tc: `should marshal and unmarshal empty CtxVarTypeList as expected`,
			ctxVar: ContextVariable{
				Type: CtxVarTypeList,
			},
			expected: ctxVarList{
				Name:   "",
				Type:   CtxVarTypeList,
				Types:  "",
				Values: nil,
			},
		},
		{
			tc: `should marshal and unmarshal non-empty CtxVarTypeEntity as expected`,
			ctxVar: ContextVariable{
				Name:  "foo",
				Type:  CtxVarTypeEntity,
				Value: "bar",
			},
			expected: ctxVarDefault{
				Name:  "foo",
				Type:  CtxVarTypeEntity,
				Value: "bar",
			},
		},
		{
			tc: `should marshal and unmarshal non-empty CtxVarTypeSubnet as expected`,
			ctxVar: ContextVariable{
				Name:  "foo",
				Type:  CtxVarTypeSubnet,
				Value: "bar",
			},
			expected: ctxVarDefault{
				Name:  "foo",
				Type:  CtxVarTypeSubnet,
				Value: "bar",
			},
		},
		{
			tc: `should return ErrInvalidCtxVarType for unexpected context variable type`,
			ctxVar: ContextVariable{
				Type: "great-type",
			},
			expected: ctxVarDefault{},
			errStr:   "json: error calling MarshalJSON for type ociauthz.ContextVariable: invalid context variable type",
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			expectedJSON, err := json.Marshal(test.expected)
			assert.Nil(t, err)

			marshaled, err := json.Marshal(test.ctxVar)

			if test.errStr != "" {
				assert.Equal(t, test.errStr, err.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, expectedJSON, marshaled)

				// check unmarshaling equals the input
				var unmarshaled ContextVariable
				assert.Nil(t, json.Unmarshal(marshaled, &unmarshaled))
				assert.Equal(t, test.ctxVar, unmarshaled)

				// check expected fields were not omitted during marshaling
				var checkJSONFields map[string]interface{}
				var expectedFields []string

				switch test.expected.(type) {
				case ctxVarPermission:
					expectedFields = []string{"p"}
				case ctxVarBoolean:
					expectedFields = []string{"NAME", "BOOLEAN", "TYPE"}
				case ctxVarList:
					expectedFields = []string{"NAME", "VALUES", "TYPES", "TYPE"}
				case ctxVarDefault:
					expectedFields = []string{"NAME", "VALUE", "TYPE"}
				default:
					assert.Fail(t, "Unsupported context type")
				}

				assert.Nil(t, json.Unmarshal(marshaled, &checkJSONFields))

				for _, field := range expectedFields {
					_, ok := checkJSONFields[field]
					assert.True(t, ok)
				}
			}
		})
	}
}

func TestAuthorizationRequest_SetNewTagSlug(t *testing.T) {
	tests := []struct {
		name string
		arg  *tagging.TagSlug
		want *tagging.TagSlug
	}{
		{
			name: "empty tag slug",
			arg:  emptyTagSlug,
		},
		{
			name: "freeform tags only",
			arg:  freeformTagSlug,
		},
		{
			name: "defined tags only",
			arg:  definedTagSlug,
		},
		{
			name: "freeform and defined tags",
			arg:  fullTagSlug,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &AuthorizationRequest{}

			// Should be empty by default
			assert.Empty(t, request.TagSlugChanges)

			request.SetNewTagSlug(tt.arg)

			// Headers should always be added, even to empty slugs
			assert.True(t, bytes.HasPrefix(*request.TagSlugChanges, tagging.ProtobufHeaders))
		})
	}
}

func TestAuthorizationRequest_SetExistingTagSlug(t *testing.T) {
	tests := []struct {
		name string
		arg  *tagging.TagSlug
	}{
		{
			name: "empty tag slug",
			arg:  emptyTagSlug,
		},
		{
			name: "freeform tags only",
			arg:  freeformTagSlug,
		},
		{
			name: "defined tags only",
			arg:  definedTagSlug,
		},
		{
			name: "freeform tags only",
			arg:  fullTagSlug,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &AuthorizationRequest{}

			// Should be empty by default
			assert.Empty(t, request.TagSlugOriginal)

			request.SetExistingTagSlug(tt.arg)

			// Headers should always be added, even to empty slugs
			assert.True(t, bytes.HasPrefix(*request.TagSlugOriginal, tagging.ProtobufHeaders))
		})
	}
}

func TestAuthorizeTags(t *testing.T) {
	tagErr := errors.New("a tagging authz error")
	type fields struct {
		TagSlugError  error
		TagSlugMerged *tagging.TagSlug
	}
	type want struct {
		AuthorizeTags     bool
		TagError          error
		MergedTagSlug     *tagging.TagSlug
		ResponseErrorType AuthzResponseErrorType
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "has a merged tag slug when authorized",
			fields: fields{
				TagSlugError:  nil,
				TagSlugMerged: testTagSlug,
			},
			want: want{
				AuthorizeTags: true,
				TagError:      nil,
				MergedTagSlug: testTagSlug,
			},
		},
		{
			name: "has an error when not authorized",
			fields: fields{
				TagSlugError:  tagErr,
				TagSlugMerged: nil,
			},
			want: want{
				AuthorizeTags:     false,
				TagError:          tagErr,
				MergedTagSlug:     nil,
				ResponseErrorType: TaggingAuthzOrNotExistError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthorizationContext{
				TagSlugError:  tt.fields.TagSlugError,
				TagSlugMerged: tt.fields.TagSlugMerged,
			}
			assert.Equal(t, tt.want.AuthorizeTags, a.AuthorizeTags())
			assert.Equal(t, tt.want.TagError, a.GetTagError())
			assert.Equal(t, tt.want.MergedTagSlug, a.GetTagSlug())

			r := &AuthorizationResponse{}
			if tt.want.TagError != nil {
				r.OutboundAuthorizationRequest.TagSlugError = tt.want.TagError.Error()
				r.OutboundAuthorizationRequest.ResponseErrorType = tt.want.ResponseErrorType
			} else {
				r.OutboundAuthorizationRequest.TagSlugMerged = tt.fields.TagSlugMerged
			}
			assert.Equal(t, tt.want.AuthorizeTags, r.AuthorizeTags())
			assert.Equal(t, tt.want.TagError, r.GetTagError())
			assert.Equal(t, tt.want.MergedTagSlug, r.GetTagSlug())
			assert.Equal(t, tt.want.ResponseErrorType, r.GetErrorType())
		})
	}
}

// Verify GetTagSlug strips the slug headers
func TestGetTagSlug(t *testing.T) {
	ar := *newAuthzResponseWithTags(validAuthorizationResponse, nil, testTagSlug, testTagSlug, "")
	assert.Equal(t, testTagSlug, ar.GetTagSlug())
}
