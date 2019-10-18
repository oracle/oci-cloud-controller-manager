// Copyright (c) 2018, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"oracle.com/oci/httpsigner"
	"oracle.com/oci/tagging"
)

const (
	requestIDHeader  = "opc-request-id"
	commonPermission = "__COMMON__"
)

// Context var definitions. These are string versions of the definitions in httpiam and ocihttpiam, here to
// avoid a httpiam dependency
const (
	CtxVarTypeString = "STRING"
	CtxVarTypeBool   = "BOOLEAN"
	CtxVarTypeInt    = "INTEGER"
	CtxVarTypeList   = "LIST"
	CtxVarTypeEntity = "ENTITY"
	CtxVarTypeSubnet = "SUBNET"
)

// ActionProfile definitions. For mapping ActionKinds to ActionProfiles
const (
	ActionProfileReadOnly           = "READ_ONLY"
	ActionProfileConstructiveChange = "CONSTRUCTIVE_CHANGE"
	ActionProfileDestructiveChange  = "DESTRUCTIVE_CHANGE"
	ActionProfileAssociate          = "ASSOCIATE"
)

// AssociationAuthorizationResult constants
const (
	AssociationFailUnknown        AssociationAuthorizationResult = "FAIL_UNKNOWN"         // Association Failed for unknown reason
	AssociationFailBadRequest     AssociationAuthorizationResult = "FAIL_BAD_REQUEST"     // Association Failed due to bad request
	AssociationFailMissingEndorse AssociationAuthorizationResult = "FAIL_MISSING_ENDORSE" // Association Failed due to missing endorsement
	AssociationFailMissingAdmit   AssociationAuthorizationResult = "FAIL_MISSING_ADMIT"   // Association Failed due to missing admit
	AssociationSuccess            AssociationAuthorizationResult = "SUCCESS"              // Association Successful
)

// ActionKind definitions.
const (
	ActionKindCreate           = ActionKind("CREATE")
	ActionKindRead             = ActionKind("READ")
	ActionKindUpdate           = ActionKind("UPDATE")
	ActionKindDelete           = ActionKind("DELETE")
	ActionKindList             = ActionKind("LIST")
	ActionKindAttach           = ActionKind("ATTACH")
	ActionKindDetach           = ActionKind("DETACH")
	ActionKindOther            = ActionKind("OTHER")
	ActionKindNotDefined       = ActionKind("NOT_DEFINED")
	ActionKindSearch           = ActionKind("SEARCH")
	ActionKindUpdateRemoveOnly = ActionKind("UPDATE_REMOVE_ONLY")
)

// AuthzResponseErrorType is the set of tagging authorization response error types.
type AuthzResponseErrorType string

// Set of possible `AuthzResponseErrorType`s
const (
	TaggingNoError              AuthzResponseErrorType = "NO_ERROR"
	TaggingEmpty                AuthzResponseErrorType = "EMPTY"
	TaggingResourceAuthzError   AuthzResponseErrorType = "RESOURCE_AUTHORIZATION_ERROR"
	TaggingAuthzOrNotExistError AuthzResponseErrorType = "TAG_AUTHORIZATION_OR_NOT_EXIST_ERROR"
	TaggingValidationError      AuthzResponseErrorType = "TAG_VALIDATION_ERROR"
)

// ActionKindToActionProfileMap maps Action Kinds to their Profile Map
var ActionKindToActionProfileMap = map[ActionKind]string{
	ActionKindCreate:           ActionProfileConstructiveChange,
	ActionKindRead:             ActionProfileReadOnly,
	ActionKindUpdate:           ActionProfileConstructiveChange,
	ActionKindDelete:           ActionProfileDestructiveChange,
	ActionKindList:             ActionProfileReadOnly,
	ActionKindAttach:           ActionProfileAssociate,
	ActionKindDetach:           ActionProfileDestructiveChange,
	ActionKindOther:            ActionProfileConstructiveChange,
	ActionKindNotDefined:       ActionProfileConstructiveChange,
	ActionKindSearch:           ActionProfileReadOnly,
	ActionKindUpdateRemoveOnly: ActionProfileDestructiveChange,
}

// ActionsOrderedBySeverity is an ordered list of ActionKinds ordered by severity
var ActionsOrderedBySeverity = []ActionKind{
	ActionKindCreate,
	ActionKindUpdate,
	ActionKindAttach,
	ActionKindOther,
	ActionKindDelete,
	ActionKindDetach,
	ActionKindUpdateRemoveOnly,
	ActionKindList,
	ActionKindRead,
	ActionKindSearch,
}

// AuthorizationClient interface defines expected functionality of an authorization client
// TODO: #major change to return AuthorizationContext that contains AuthorizationResponse
type AuthorizationClient interface {
	// All should respond with true if all requested permissions in the request are allowed.  In addition to
	// the authorization decision, additional details are returned in authzResponse.  If there is an error
	// while processing the request, it should return granted=false, and an appropriate error.
	//
	// Deprecated: The permission checking functionality of All has been moved the AuthorizationResponse
	All(*AuthorizationRequest) (granted bool, authzResponse *AuthorizationResponse, err error)

	// Any should respond with true if at least one of the requested permissions in the request are allowed.  In addition to
	// the authorization decision, additional details are returned in authzResponse.  If there is an error
	// while processing the request, it should return granted=false, and an appropriate error.
	//
	// Deprecated: The permission checking functionality of Any has been moved the AuthorizationResponse
	Any(*AuthorizationRequest) (granted bool, authzResponse *AuthorizationResponse, err error)

	// Filter should respond with a list of requested permissions in the request that are allowed.  In addition to
	// the filtered list, additional details are returned in authzResponse.  If there is an error
	// while processing the request, it should return permissions=[]string{} and an appropriate error.
	//
	// Deprecated: The permission checking functionality of Filter has been moved the AuthorizationResponse
	Filter(*AuthorizationRequest) (permissions []string, authzResponse *AuthorizationResponse, err error)
}

// AuthorizationClientExtended is an extension to the AuthorizationClient, initially for supporting
// the MakeAuthorizationCall function on an AuthorizationClient
type AuthorizationClientExtended interface {
	AuthorizationClient

	// MakeAuthorizationCall executes an authorization request to identity and returns the
	// AuthorizationResponse on success, or passes back the error encountered for the request
	MakeAuthorizationCall(*AuthorizationRequest) (*AuthorizationResponse, error)
}

// AssociationAuthorizationClient builds upon the AuthorizationClientExtended and adds
// AssociationAuthorization functions.
type AssociationAuthorizationClient interface {
	AuthorizationClientExtended
	MakeAssociationAuthorizationCall(*AssociationAuthorizationRequest) (*AssociationAuthorizationResponse, error)
}

// authorizationClient implements the AuthorizationClient interface
type authorizationClient struct {
	client         httpsigner.Client
	endpoint       string
	uriTemplate    string
	taggingEnabled bool
}

// OutboundAuthorizationRequest represents authorization payload sent to the identity service for performing
// authorization. Note that JSON marshalling is done via MarshalJSON function below which initializes an anonymous
// struct which should be kept in sync if this struct is updated.
type OutboundAuthorizationRequest struct {
	Client            string                        `json:"client,omitempty"`
	RequestID         string                        `json:"requestId"`
	ServiceName       string                        `json:"serviceName"`
	UserPrincipal     AuthorizationRequestPrincipal `json:"userPrincipal,omitempty"`
	ServicePrincipal  AuthorizationRequestPrincipal `json:"svcPrincipal,omitempty"`
	OBOPrincipal      AuthorizationRequestPrincipal `json:"oboPrincipal,omitempty"`
	Principal         AuthorizationRequestPrincipal `json:"principal,omitempty"`
	Context           [][]ContextVariable           `json:"context"`
	Properties        struct{}                      `json:"properties"`
	RequestRegion     string                        `json:"region"`
	PhysicalAD        string                        `json:"physicalAD"`
	ActionKind        ActionKind                    `json:"actionKind"`
	TagSlugOriginal   *tagging.TagSlug              `json:"tagSlugO,omitempty"`
	TagSlugChanges    *tagging.TagSlug              `json:"tagSlugC,omitempty"`
	TagSlugMerged     *tagging.TagSlug              `json:"tagSlugM,omitempty"`
	TagSlugError      string                        `json:"tagError,omitempty"`
	ResponseErrorType AuthzResponseErrorType        `json:"responseError,omitempty"`
}

// OutboundAssociationAuthorizationRequest represents a slice of OutboundAuthorizationRequests
type OutboundAssociationAuthorizationRequest struct {
	Requests []OutboundAuthorizationRequest `json:"requests"`
}

// AuthorizationRequestPrincipal represents a principal that is marshaled inside the authorization request
type AuthorizationRequestPrincipal struct {
	TenantID  string  `json:"tenantId"`
	SubjectID string  `json:"subjectId"`
	Claims    []Claim `json:"claims"`
}

// ContextVariable represents context variable for the authorization request
type ContextVariable struct {
	P       string   `json:"p"`
	Name    string   `json:"NAME"`
	Type    string   `json:"TYPE"`
	Types   string   `json:"TYPES"`
	Value   string   `json:"VALUE"`
	Values  []string `json:"VALUES"`
	Boolean *bool    `json:"BOOLEAN"`
}

// ctxVarPermission represents a permission of context variables
type ctxVarPermission struct {
	P string `json:"p"`
}

// ctxVarDefault represents a default context variable JSON format
type ctxVarDefault struct {
	Name  string `json:"NAME"`
	Type  string `json:"TYPE"`
	Value string `json:"VALUE"`
}

// ctxVarBoolean represents a boolean type context variable JSON format
type ctxVarBoolean struct {
	Name    string `json:"NAME"`
	Type    string `json:"TYPE"`
	Boolean *bool  `json:"BOOLEAN"`
}

// ctxVarList represents a list type context variable JSON format
type ctxVarList struct {
	Name   string   `json:"NAME"`
	Type   string   `json:"TYPE"`
	Types  string   `json:"TYPES"`
	Values []string `json:"VALUES"`
}

// MarshalJSON is called by json.Marshal. It delegates the work of marshaling
// the struct to one of `ctxVarPermission`, `ctxVarDefault`, `ctxVarBoolean` or
// `ctxVarList`.
func (c ContextVariable) MarshalJSON() ([]byte, error) {
	var ctxVar interface{}
	switch t := c.Type; t {
	case "":
		ctxVar = ctxVarPermission{P: c.P}
	case CtxVarTypeList:
		ctxVar = ctxVarList{
			Name:   c.Name,
			Type:   c.Type,
			Types:  c.Types,
			Values: c.Values,
		}
	case CtxVarTypeBool:
		ctxVar = ctxVarBoolean{
			Name:    c.Name,
			Type:    c.Type,
			Boolean: c.Boolean,
		}
	case CtxVarTypeEntity, CtxVarTypeInt, CtxVarTypeString, CtxVarTypeSubnet:
		ctxVar = ctxVarDefault{
			Name:  c.Name,
			Type:  c.Type,
			Value: c.Value,
		}
	default:
		return nil, ErrInvalidCtxVarType
	}
	return json.Marshal(ctxVar)
}

// AuthorizationRequest is an internal representation of authorization request.  This struct is converted
// into OutboundAuthorizationRequest internally during authorization.
type AuthorizationRequest struct {
	RequestID        string
	OperationID      string
	CompartmentID    string
	ServiceName      string
	UserPrincipal    *Principal
	ServicePrincipal *Principal
	Context          [][]ContextVariable
	Properties       struct{}
	RequestRegion    string
	PhysicalAD       string
	ActionKind       ActionKind
	TagSlugOriginal  *tagging.TagSlug
	TagSlugChanges   *tagging.TagSlug
}

// AssociationAuthorizationRequest is an internal representation of an association
// authorization request. This struct is converted into
// OutboundAssociationAuthorizationRequest internally during authorization.
type AssociationAuthorizationRequest []AuthorizationRequest

// AuthzVariable is an internal representation of authorization variables
type AuthzVariable struct {
	Name    string
	Type    string
	Types   string
	Value   string
	Values  []string
	Boolean bool
}

// AuthorizationContext is returned with the authorization result to give the caller the context as to why the
// authorization resulted in success or failure.  It contains RequestID which is the RequestID we get back
// from Identity, and Permissions which are list of permissions that the user is authorized to perform.
type AuthorizationContext struct {
	RequestID         string
	Permissions       []string
	ResponseErrorType AuthzResponseErrorType
	TagSlugError      error
	TagSlugMerged     *tagging.TagSlug
}

// AuthorizationResponse represents authz response returned from the OCI identity authorization service.  Additionally
// it contains an AuthorizationContext object and the original AuthorizationRequest sent for this AuthorizationResponse
type AuthorizationResponse struct {
	AuthorizationContext         AuthorizationContext
	AuthorizationRequest         *AuthorizationRequest
	OutboundAuthorizationRequest OutboundAuthorizationRequest `json:"authorizationRequest"`
	Duration                     string                       `json:"decisionCacheDuration"`
}

// AssociationAuthorizationResponse represents authz association response returned
// from the OCI identity authorization service.
type AssociationAuthorizationResponse struct {
	Responses        []AuthorizationResponse        `json:"responses"`
	AssocationResult AssociationAuthorizationResult `json:"associationResult"`
}

// empty returns true if the principal is empty value
func (a AuthorizationRequestPrincipal) empty() bool {
	if a.SubjectID == "" && a.TenantID == "" && a.Claims == nil {
		return true
	}
	return false
}

// MarshalJSON is called by json.Marshal. It takes care of omitting empty structs.
// Deprecated: removed for 1.0
func (a OutboundAuthorizationRequest) MarshalJSON() ([]byte, error) {
	var userPrincipal, servicePrincipal, oboPrincipal, principal *AuthorizationRequestPrincipal

	if !a.UserPrincipal.empty() {
		userPrincipal = &a.UserPrincipal
	}
	if !a.OBOPrincipal.empty() {
		oboPrincipal = &a.OBOPrincipal
	}
	if !a.ServicePrincipal.empty() {
		servicePrincipal = &a.ServicePrincipal
	}
	if !a.Principal.empty() {
		principal = &a.Principal
	}

	// Note that this anonymous struct represents OutboundAuthorizationRequest. It should be kept in sync with the
	// OutboundAuthorizationRequest struct defined above.
	return json.Marshal(struct {
		Client            string                         `json:"client,omitempty"`
		RequestID         string                         `json:"requestId"`
		ServiceName       string                         `json:"serviceName"`
		UserPrincipal     *AuthorizationRequestPrincipal `json:"userPrincipal,omitempty"`
		ServicePrincipal  *AuthorizationRequestPrincipal `json:"svcPrincipal,omitempty"`
		OBOPrincipal      *AuthorizationRequestPrincipal `json:"oboPrincipal,omitempty"`
		Principal         *AuthorizationRequestPrincipal `json:"principal,omitempty"`
		Context           [][]ContextVariable            `json:"context"`
		Properties        struct{}                       `json:"properties"`
		RequestRegion     string                         `json:"region"`
		PhysicalAD        string                         `json:"physicalAD"`
		TagSlugOriginal   *tagging.TagSlug               `json:"tagSlugO,omitempty"`
		TagSlugChanges    *tagging.TagSlug               `json:"tagSlugC,omitempty"`
		TagSlugMerged     *tagging.TagSlug               `json:"tagSlugM,omitempty"`
		TagSlugError      string                         `json:"tagError,omitempty"`
		ResponseErrorType AuthzResponseErrorType         `json:"responseError,omitempty"`
	}{
		Client:            a.Client,
		RequestID:         a.RequestID,
		ServiceName:       a.ServiceName,
		UserPrincipal:     userPrincipal,
		ServicePrincipal:  servicePrincipal,
		OBOPrincipal:      oboPrincipal,
		Principal:         principal,
		Context:           a.Context,
		Properties:        a.Properties,
		RequestRegion:     a.RequestRegion,
		PhysicalAD:        a.PhysicalAD,
		TagSlugOriginal:   a.TagSlugOriginal,
		TagSlugChanges:    a.TagSlugChanges,
		TagSlugMerged:     a.TagSlugMerged,
		TagSlugError:      a.TagSlugError,
		ResponseErrorType: a.ResponseErrorType,
	})
}

// principalToAuthorizationPrincipal will return an AuthorizationRequestPrincipal from the given ociauthz.Principal
func principalToAuthorizationPrincipal(p *Principal) AuthorizationRequestPrincipal {
	if p == nil {
		return AuthorizationRequestPrincipal{}
	}

	claims := p.Claims().ToSlice()

	return AuthorizationRequestPrincipal{
		TenantID:  p.TenantID(),
		SubjectID: p.ID(),
		Claims:    claims,
	}
}

// authorizationRequestToOutbound converts the given AuthorizationRequest into outbound format understood
// by the identity service
func authorizationRequestToOutbound(r *AuthorizationRequest) *OutboundAuthorizationRequest {
	// legacy attributes and behaviors, to remove
	// (user, null) -> principal: user, obo: null
	// (svc, null)  -> principal: svc, obo: null
	// (svc, user)  -> principal: svc, obo: user
	var principal, userPrincipal, servicePrincipal, oboPrincipal AuthorizationRequestPrincipal
	if r.UserPrincipal != nil && r.ServicePrincipal == nil {
		principal = principalToAuthorizationPrincipal(r.UserPrincipal)
		userPrincipal = principalToAuthorizationPrincipal(r.UserPrincipal)
	} else if r.UserPrincipal == nil && r.ServicePrincipal != nil {
		principal = principalToAuthorizationPrincipal(r.ServicePrincipal)
		servicePrincipal = principalToAuthorizationPrincipal(r.ServicePrincipal)
	} else {
		principal = principalToAuthorizationPrincipal(r.ServicePrincipal)
		userPrincipal = principalToAuthorizationPrincipal(r.UserPrincipal)
		servicePrincipal = principalToAuthorizationPrincipal(r.ServicePrincipal)
		oboPrincipal = principalToAuthorizationPrincipal(r.UserPrincipal)
	}

	return &OutboundAuthorizationRequest{
		RequestID:        r.RequestID,
		ServiceName:      r.ServiceName,
		UserPrincipal:    userPrincipal,
		ServicePrincipal: servicePrincipal,
		OBOPrincipal:     oboPrincipal,
		Principal:        principal,
		Context:          r.Context,
		Properties:       r.Properties,
		RequestRegion:    r.RequestRegion,
		PhysicalAD:       r.PhysicalAD,
		ActionKind:       r.ActionKind,
	}
}

// associationAuthorizationRequestToOutbound
// converts AssociationAuthorizationRequest to an OutboundAssociationAuthorizationRequest
func associationAuthorizationRequestToOutbound(r *AssociationAuthorizationRequest) *OutboundAssociationAuthorizationRequest {
	outbound := &OutboundAssociationAuthorizationRequest{}
	for _, request := range *r {
		outboundRequest := authorizationRequestToOutbound(&request)
		outbound.Requests = append(outbound.Requests, *outboundRequest)
	}
	return outbound
}

// NewAuthorizationRequest will construct a new authorization request object that users may use to perform authorization
func NewAuthorizationRequest(requestID, operationID, compartmentID, serviceName string,
	userPrincipal *Principal, servicePrincipal *Principal, requestRegion, physicalAD string) (*AuthorizationRequest, error) {

	// Default to ActionKindNotDefined
	request := &AuthorizationRequest{
		RequestID:        requestID,
		OperationID:      operationID,
		CompartmentID:    compartmentID,
		ServiceName:      serviceName,
		UserPrincipal:    userPrincipal,
		ServicePrincipal: servicePrincipal,
		RequestRegion:    requestRegion,
		PhysicalAD:       physicalAD,
		ActionKind:       ActionKindNotDefined,
	}

	if err := request.CheckRequiredVariables(); err != nil {
		return nil, err
	}

	return request, nil
}

// NewAssociationAuthorizationRequest builds an AssociationAuthorizationRequest object
// from a variadic slice of AuthorizationRequest
func NewAssociationAuthorizationRequest(reqs ...AuthorizationRequest) (*AssociationAuthorizationRequest, error) {
	if len(reqs) < 2 {
		return nil, ErrAssociationInsufficientRequests
	}
	for _, request := range reqs {
		if err := request.CheckRequiredVariables(); err != nil {
			return nil, err
		}
	}
	newAssociationAuthorizationRequest := AssociationAuthorizationRequest(reqs)
	return &newAssociationAuthorizationRequest, nil
}

// SetActionKind sets the ActionKind for this request
func (request *AuthorizationRequest) SetActionKind(actionKind ActionKind) error {
	if _, ok := ActionKindToActionProfileMap[actionKind]; !ok {
		return ErrInvalidActionKind
	}
	request.ActionKind = actionKind
	return nil
}

// SetPermissionVariables makes it east to add new permission and a list of context variables
// Context is presented as a array of records, where each record contains a permission and zero or more context variables
// e.g., "CONTEXT" : [[ { "p" : "__COMMON__"} ]] is an empty context,
// while "CONTEXT" : [
//   [ { "p" : "__COMMON__"},
//     { "NAME" : "target.compartment.id", "TYPE" : "ENTITY", "VALUE" : "ocid1.ten..."} ],
//   [ { "p" : "A_VERB" } ]
// ] is a context for permission A_VERB with a target.compartment.id.
func (request *AuthorizationRequest) SetPermissionVariables(permission string, contexts []AuthzVariable) {
	result := make([]ContextVariable, 0, len(contexts))

	for _, context := range contexts {
		var c ContextVariable
		if context.Type == CtxVarTypeBool {
			// We set Boolean to a pointer, to avoid sending false as the nil value for other context var types
			if context.Boolean {
				trueP := new(bool)
				*trueP = true
				c = ContextVariable{P: "", Name: context.Name, Type: context.Type, Types: context.Types, Boolean: trueP}
			} else {
				falseP := new(bool)
				*falseP = false
				c = ContextVariable{P: "", Name: context.Name, Type: context.Type, Types: context.Types, Boolean: falseP}
			}
		} else {
			c = ContextVariable{P: "", Name: context.Name, Type: context.Type, Types: context.Types, Value: context.Value, Values: context.Values}
		}
		result = append(result, c)
	}

	// check if the given permission already exists. If it does, append the
	// context variables to the same permission
	var permissionExists bool
	for ii, ctx := range request.Context {
		if ctx[0].P == permission {
			permissionExists = true
			request.Context[ii] = append(request.Context[ii], result...)
		}
	}

	if !permissionExists {
		c := ContextVariable{P: permission, Name: "", Type: "", Value: ""}
		result = append([]ContextVariable{c}, result...)
		request.Context = append(request.Context, result)
	}
}

// SetCommonPermission sets the given authz variable under the __COMMON__ permission.
func (request *AuthorizationRequest) SetCommonPermission(ctx AuthzVariable) {
	request.SetPermissionVariables(commonPermission, []AuthzVariable{ctx})
}

// SetCommonPermissions sets common context variables on the given AuthorizationRequest
func (request *AuthorizationRequest) SetCommonPermissions() {
	// Set common context variables
	request.SetPermissionVariables(commonPermission, []AuthzVariable{
		{"target.compartment.id", CtxVarTypeEntity, "", request.CompartmentID, nil, false},
		{"request.operation", CtxVarTypeString, "", request.OperationID, nil, false},
	})
}

// CheckRequiredVariables will return an error if any required variables are missing
func (request *AuthorizationRequest) CheckRequiredVariables() error {
	if request.OperationID == "" {
		return ErrInvalidOperationID
	}
	if request.CompartmentID == "" {
		return ErrInvalidCompartmentID
	}
	if request.RequestRegion == "" {
		return ErrInvalidRequestRegion
	}
	if request.PhysicalAD == "" {
		return ErrInvalidPhysicalAD
	}
	if _, ok := ActionKindToActionProfileMap[request.ActionKind]; !ok {
		return ErrInvalidActionKind
	}
	if request.UserPrincipal == nil && request.ServicePrincipal == nil {
		return ErrInvalidPrincipal
	}
	return nil
}

// SetExistingTagSlug sets the existing (aka original) tag slug
func (request *AuthorizationRequest) SetExistingTagSlug(existingSlug *tagging.TagSlug) {
	request.TagSlugOriginal = existingSlug
}

// SetNewTagSlug sets the new, or updated, tag slug changes
func (request *AuthorizationRequest) SetNewTagSlug(newSlug *tagging.TagSlug) {
	request.TagSlugChanges = newSlug
}

func newAuthorizationClient(client httpsigner.Client, endpoint, uriTemplate string, taggingEnabled bool) *authorizationClient {
	if endpoint == "" || client == nil || uriTemplate == "" {
		panic("programmer error: must provide non-empty endpoint and authorizeURITemplate and non-nil client")
	}
	return &authorizationClient{
		client:         client,
		endpoint:       endpoint,
		uriTemplate:    uriTemplate,
		taggingEnabled: taggingEnabled,
	}
}

// NewAuthorizationClient constructs a new authorization client from the given parameters
func NewAuthorizationClient(client httpsigner.Client, endpoint string) AuthorizationClient {
	return newAuthorizationClient(client, endpoint, authorizeURITemplate, false)
}

// NewAuthorizationClientExtended constructs a new authorization client with extended
// functionality from the given parameters
func NewAuthorizationClientExtended(client httpsigner.Client, endpoint string) AuthorizationClientExtended {
	return newAuthorizationClient(client, endpoint, authorizeURITemplate, false)
}

// NewAssociationAuthorizationClient constructs a new authorization client with association
// functionality from the given parameters
func NewAssociationAuthorizationClient(client httpsigner.Client, endpoint string) AssociationAuthorizationClient {
	return newAuthorizationClient(client, endpoint, authorizeURITemplate, false)
}

// NewAuthorizationClientWithTags constructs a new authorization client with tagging authorization enabled.
func NewAuthorizationClientWithTags(client httpsigner.Client, endpoint string) AuthorizationClient {
	return newAuthorizationClient(client, endpoint, authorizeWithTagsURITemplate, true)
}

// NewAssociationAuthorizationClientWithTags constructs a new association authorization client with tagging authorization enabled.
func NewAssociationAuthorizationClientWithTags(client httpsigner.Client, endpoint string) AssociationAuthorizationClient {
	return newAuthorizationClient(client, endpoint, authorizeWithTagsURITemplate, true)
}

// commonChecksAndPermissions performs some basic validation and sets common permissions
func (request *AuthorizationRequest) commonChecksAndPermissions() error {
	// Make sure all required variables are set
	err := request.CheckRequiredVariables()
	if err != nil {
		return err
	}

	// Must have permissions inside the request
	if len(request.Context) == 0 {
		return ErrNoPermissionsSet
	}

	// Set common permissions
	request.SetCommonPermissions()
	return nil
}

// performRequestToIdentity performs the request to identity.
func (a *authorizationClient) performRequestToIdentity(req *http.Request) (*http.Response, []byte, error) {
	// Perform the request
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	// Error if status code is not 200
	if resp.StatusCode != http.StatusOK {
		return nil, nil, &ServiceResponseError{resp}
	}

	// Read the response body
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	return resp, body, nil
}

// authorize performs authorization on the given AuthorizationRequest
func (a *authorizationClient) authorize(r *AuthorizationRequest) (*AuthorizationResponse, error) {
	// Make sure all required variables are set and common permissions set.
	err := r.commonChecksAndPermissions()
	if err != nil {
		return nil, err
	}

	// Convert AuthorizationRequest into outbound format
	outboundRequest := authorizationRequestToOutbound(r)

	// Add tag slugs and use the tagging authorization URI
	if a.taggingEnabled {
		outboundRequest.TagSlugChanges = r.TagSlugChanges
		outboundRequest.TagSlugOriginal = r.TagSlugOriginal
	}

	// Marshal the given request into JSON
	b, err := json.Marshal(*outboundRequest)
	if err != nil {
		return nil, err
	}

	// Create a new request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(a.uriTemplate, a.endpoint), bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Add(requestIDHeader, outboundRequest.RequestID)

	resp, body, err := a.performRequestToIdentity(req)
	if err != nil {
		return nil, err
	}

	// Attempt to unmarshal the response body into AuthorizationResponse
	bytes := []byte(body)
	var authorizationResponse AuthorizationResponse
	err = json.Unmarshal(bytes, &authorizationResponse)
	if err != nil {
		return nil, err
	}

	requestID := resp.Header.Get(requestIDHeader)
	if requestID == "" {
		requestID = outboundRequest.RequestID
	}
	authorizationResponse.AuthorizationContext.RequestID = requestID

	authorizationResponse.AuthorizationContext.TagSlugError = authorizationResponse.GetTagError()
	authorizationResponse.AuthorizationContext.TagSlugMerged = authorizationResponse.GetTagSlug()

	// ResponseErrorType allows clients to differentiate between
	// authorization errors (e.g. user not authz to resource and/or
	// tag key) and validation errors (e.g. invalid tag key, namespace, etc.)
	authorizationResponse.AuthorizationContext.ResponseErrorType = authorizationResponse.GetErrorType()

	// Success!
	return &authorizationResponse, nil
}

// All will result in authorization grant value of true as long as all of request permissions were granted
//
// Deprecated: The permission checking functionality of All has been moved the AuthorizationResponse
func (a *authorizationClient) All(request *AuthorizationRequest) (bool, *AuthorizationResponse, error) {
	response, err := a.MakeAuthorizationCall(request)
	if err != nil {
		return false, nil, err
	}
	response.AuthorizationContext.Permissions = response.Filter()
	return response.All(), response, nil
}

// Any will result in authorization grant value of true as long as one of request permissions was granted
//
// Deprecated: The permission checking functionality of Any has been moved the AuthorizationResponse
func (a *authorizationClient) Any(request *AuthorizationRequest) (bool, *AuthorizationResponse, error) {
	response, err := a.MakeAuthorizationCall(request)
	if err != nil {
		return false, nil, err
	}
	response.AuthorizationContext.Permissions = response.Filter()
	return response.Any(), response, nil
}

// Filter returns a filtered set of Permissions that the actor has access to
//
// Deprecated: The permission checking functionality of Any has been moved the AuthorizationResponse
func (a *authorizationClient) Filter(request *AuthorizationRequest) ([]string, *AuthorizationResponse, error) {
	response, err := a.MakeAuthorizationCall(request)
	if err != nil {
		return []string{}, nil, err
	}
	set := response.Filter()
	response.AuthorizationContext.Permissions = set
	return set, response, nil
}

// MakeAssociationAuthorizationCall executes an associationAuthorization request to identity and
// returns the AssociationAuthorizationResponse upon success, or passes back the error encountered for the request
func (a *authorizationClient) MakeAssociationAuthorizationCall(r *AssociationAuthorizationRequest) (*AssociationAuthorizationResponse, error) {
	// We can only associate if there is something to associate with.
	if len(*r) < 2 {
		return nil, ErrAssociationInsufficientRequests
	}

	// Make sure all required variables are set and common permissions set for each underlying AuthorizationRequest.
	for ii := range *r {
		err := (*r)[ii].commonChecksAndPermissions()
		if err != nil {
			return nil, err
		}
	}

	// Convert AssociationAuthorizationRequest into outbound format
	outboundRequest := associationAuthorizationRequestToOutbound(r)

	// Marshal the given request into JSON
	b, err := json.Marshal(*outboundRequest)
	if err != nil {
		return nil, err
	}
	// Create a new request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(associationAuthorizeURITemplate, a.endpoint), bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	_, body, err := a.performRequestToIdentity(req)
	if err != nil {
		return nil, err
	}

	// Attempt to unmarshal the response body into AssociationAuthorizationResponse
	bytes := []byte(body)
	var associationAuthorizationResponse AssociationAuthorizationResponse
	err = json.Unmarshal(bytes, &associationAuthorizationResponse)
	if err != nil {
		return nil, err
	}

	// Save the original authorization request with the response so we can figure out
	// permissions in the AuthorizationResponse object
	if len(*r) != len(associationAuthorizationResponse.Responses) {
		return nil, ErrUnexpectedAssociationAuthzResponseLength
	}
	for ii := range associationAuthorizationResponse.Responses {
		request := (*r)[ii]
		associationAuthorizationResponse.Responses[ii].AuthorizationRequest = &request
	}

	return &associationAuthorizationResponse, nil

}

// MakeAuthorizationCall executes an authorization request to identity and returns the
// AuthorizationResponse on success, or passes back the error encountered for the request
func (a *authorizationClient) MakeAuthorizationCall(r *AuthorizationRequest) (*AuthorizationResponse, error) {
	// Continue to do all the validation, marshalling, sending of the request, and unmarshalling
	// we are currently doing in authorize call
	response, err := a.authorize(r)
	if err != nil {
		return nil, err
	}

	// Save the original authorization request with the response so we can figure out
	// permissions in the AuthorizationResponse object
	response.AuthorizationRequest = r

	return response, nil
}

// RequestedPermissions returns a list of permission strings that were requested
func (a AuthorizationResponse) RequestedPermissions() []string {
	return getPermissionsFromContextVariables(a.AuthorizationRequest.Context)
}

// GrantedPermissions returns a list of permission strings that were granted
func (a AuthorizationResponse) GrantedPermissions() []string {
	return getPermissionsFromContextVariables(a.OutboundAuthorizationRequest.Context)
}

// Filter returns the intersecting set of requested permissions and permissions the user is validated
// for from identity
func (a *AuthorizationResponse) Filter() []string {
	return permissionsIntersect(a.RequestedPermissions(), a.GrantedPermissions())
}

// Any returns true if the user has any authorized permissions
func (a *AuthorizationResponse) Any() bool {
	return len(a.Filter()) > 0
}

// All returns true if all of the requested permissions are authorized
func (a *AuthorizationResponse) All() bool {
	return a.Set(a.RequestedPermissions()) && a.AuthorizeTags()
}

// Set returns true if all the passed in permissions are found in the permissions the
// user is authorized to perform
func (a *AuthorizationResponse) Set(permissions []string) bool {
	if len(permissions) == 0 {
		return false
	}
	responsePermissions := a.GrantedPermissions()
	if len(responsePermissions) == 0 {
		return false
	}
	diff := permissionsDifference(permissions, responsePermissions)
	return len(diff) == 0
}

// AuthorizeTags returns true if the request was tag annotated and there was no tag error.
func (a *AuthorizationResponse) AuthorizeTags() bool {
	return a.GetTagError() == nil
}

// GetTagError returns any tag slug errors, due to an authorization failure or issue with the tag slug encoding.
func (a *AuthorizationResponse) GetTagError() error {
	err := a.OutboundAuthorizationRequest.TagSlugError
	if err != "" {
		return errors.New(err)
	}
	return nil
}

// GetTagSlug returns the authorized and merged tag slug.
func (a *AuthorizationResponse) GetTagSlug() *tagging.TagSlug {
	return a.OutboundAuthorizationRequest.TagSlugMerged
}

// GetErrorType returns the error type associated with an authorization failure.
func (a *AuthorizationResponse) GetErrorType() AuthzResponseErrorType {
	return a.OutboundAuthorizationRequest.ResponseErrorType
}

// AuthorizeTags returns true if the request was tag annotated and there was no tag error.
func (a *AuthorizationContext) AuthorizeTags() bool {
	return a.GetTagError() == nil
}

// GetTagError returns any tag slug errors, due to an authorization failure or issue with the tag slug encoding.
func (a *AuthorizationContext) GetTagError() error {
	return a.TagSlugError
}

// GetTagSlug returns the authorized and merged tag slug.
func (a *AuthorizationContext) GetTagSlug() *tagging.TagSlug {
	return a.TagSlugMerged
}

// GetErrorType returns the error type associated with an authorization failure.
func (a *AuthorizationContext) GetErrorType() AuthzResponseErrorType {
	return a.ResponseErrorType
}

// ActionKind is a type of activity hint added to an AuthorizationRequest
type ActionKind string

// asFormattedString returns actionKind as an Uppercase string.
func (ak ActionKind) asFormattedString() string {
	return strings.ToUpper(string(ak))
}

// ActionKindByName returns ActionKind as an uppercase string
func (ak ActionKind) ActionKindByName() string {
	if _, ok := ActionKindToActionProfileMap[ak]; ok {
		return ak.asFormattedString()
	}
	return string(ActionKindNotDefined)
}

// IsChange returns true or false if ActionKind Creates
func (ak ActionKind) IsChange() bool {
	return ak != ActionKindRead
}

// IsSearch returns true or false if ActionKind is Search
func (ak ActionKind) IsSearch() bool {
	return ak == ActionKindSearch
}

// IsCreate returns true or false if ActionKind is Create
func (ak ActionKind) IsCreate() bool {
	return ak == ActionKindCreate
}

// IsDeleteFriendly returns true or false if ActionKind is delete friendly
func (ak ActionKind) IsDeleteFriendly() bool {
	switch ActionKindToActionProfileMap[ak] {
	case ActionProfileReadOnly, ActionProfileDestructiveChange:
		return true
	default:
		return false
	}
}

// GetTopActionKind returns the top (most aggressive) ActionKind from a
// list of ActionKinds
func GetTopActionKind(aks []ActionKind) ActionKind {
	// Nothing there? return ActionKindNotDefined
	if aks == nil {
		return ActionKindNotDefined
	}
	// Just one?
	if len(aks) == 1 {
		// make sure its actually a valid ActionKind.
		if _, ok := ActionKindToActionProfileMap[ActionKind(aks[0].asFormattedString())]; ok {
			// return with proper capitalization
			return ActionKind(aks[0].asFormattedString())
		}
		// if not, return ActionKindNotDefined
		return ActionKindNotDefined
	}
	// first match is the highest ranked ActionKind
	for _, orderedActionKind := range ActionsOrderedBySeverity {
		for _, actionKind := range aks {
			if ActionKind(actionKind.asFormattedString()) == orderedActionKind {
				return orderedActionKind
			}
		}
	}
	// No matches at all? return ActionKindNotDefined
	return ActionKindNotDefined

}

// AssociationAuthorizationResult type
type AssociationAuthorizationResult string

// IsSuccess returns boolean value indicating if AssociationAuthorizationResult is Successful
// or not.
func (ak AssociationAuthorizationResult) IsSuccess() bool {
	return ak == AssociationSuccess
}
