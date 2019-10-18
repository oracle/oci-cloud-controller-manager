// Copyright (c) 2018, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"time"

	"oracle.com/oci/httpsigner"
)

// OnBehalfOf constants
const (
	// default token expiration per com.oracle.pic.identity.authentication.entities.OnBehalfOfRequest
	DefaultTokenExpiration         = 21600 // 6 hours in seconds
	DefaultTokenExpirationDuration = 6 * time.Hour
	requestTypeOBO                 = "OBO"
	requestTypeDelegation          = "DELEGATION"
)

// onBehalfOfRequest represents an OBO payload sent to the identity service for obtaining an OBO
// token.
type onBehalfOfRequest struct {
	RequestHeaders     map[string][]string `json:"requestHeaders"`
	TargetServiceNames []string            `json:"targetServiceNames"`
	DelegateGroups     []string            `json:"delegateGroups"`
	OboToken           string              `json:"oboToken,omitempty"`
	RequestType        string              `json:"requestType"`
	Expiration         int                 `json:"expiration"`
}

// getRequestHeaders returns a map of request headers from the given ociauthz.Principal claims.
// The request map keys will be the header claim keys with HdrClaimPrefix removed.
// Returns an empty map for a nil Principal.
func getRequestHeaders(p *Principal) (headers map[string][]string) {
	headers = make(map[string][]string)

	if p == nil || p.Claims() == nil {
		return
	}

	for _, claim := range p.Claims().ToSlice() {
		if claim.Issuer == HdrClaimIssuer {
			hdrName := strings.Replace(claim.Key, HdrClaimPrefix, "", 1)
			headers[hdrName] = []string{claim.Value}
		}
	}
	return
}

// GetOboToken gets an Obo token from the Identity Service and returns the token string.
// If there is an error, it returns an empty string and an appropriate error.
// client should be a signing client using a principal with sufficient privilege for the OBO request (typically a service principal)
// endpoint is the base URL of the identity endpoint for the obo request (e.g. https://auth.us-phoenix-1.oraclecloud.com/v1 )
// requestPrincipal is the principal making the original request. It must have header claims for the
// request. In addition, if the delegatePrincipal is present, it is expected that `obo_tk` claim
// exists.
// delegatePrincipal can be nil if there is no delegate principal.
// requestID is the opc-request-id that will be sent to as the opc-request-id header for the OBO request.
// expiration is the duration in seconds until the OBO token will expire. ociauthz.DefaultTokenExpiration may be used as the default (6 hours).
// targetServiceNames is a list of services for which the OBO token is intended.
func GetOboToken(client httpsigner.Client, endpoint string, requestPrincipal *Principal, delegatePrincipal *Principal, requestID string, expiration int, targetServiceNames []string) (token string, err error) {
	// verify required args
	if client == nil {
		err = ErrInvalidClient
		return
	}
	if endpoint == "" {
		err = ErrInvalidEndpoint
		return
	}
	if requestPrincipal == nil {
		err = ErrInvalidRequestPrincipal
		return
	}
	if len(targetServiceNames) == 0 {
		err = ErrNoTargetServiceNames
		return
	}

	oboURI := fmt.Sprintf(OboURITemplate, endpoint)
	delegateGroups := make([]string, 0)
	duration := time.Duration(expiration) * time.Second

	token, err = getOboOrDelegationToken(client, oboURI, requestPrincipal, delegatePrincipal, requestID, duration, targetServiceNames, delegateGroups, requestTypeOBO)
	return
}

// GetDelegationToken gets a delegation token from the Identity Service and returns the token string.
// If there is an error, it returns an empty string and an appropriate error.
// client should be a signing client using a principal with sufficient privilege for the OBO request (typically a service principal)
// endpoint is the base URL of the identity endpoint for the obo request (e.g. https://auth.us-phoenix-1.oraclecloud.com/v1 )
// requestPrincipal is the principal making the original request. It must have header claims for the
// request. In addition, if the delegatePrincipal is present, it is expected that `obo_tk` claim
// exists.
// delegatePrincipal Set this to nil when requestPrincipal is NOT from an OBO token. When requestPrincipal comes from an
// OBO/Delegation token, set this to the principal who made the request using that OBO token
// requestID is the opc-request-id that will be sent to as the opc-request-id header for the OBO request.
// expiration is the duration  until the OBO token will expire. ociauthz.DefaultTokenExpirationDuration may be used as the default (6 hours).
// targetServiceNames is a list of services for which the OBO token is intended.
/// delegateGroups ocids of the delegate groups. These are the only resources that can use the delegation token
func GetDelegationToken(client httpsigner.Client, endpoint string, requestPrincipal *Principal, delegatePrincipal *Principal, requestID string, expiration time.Duration, targetServiceNames []string, delegateGroups []string) (token string, err error) {
	// verify required args
	if client == nil {
		err = ErrInvalidClient
		return
	}
	if endpoint == "" {
		err = ErrInvalidEndpoint
		return
	}
	if requestPrincipal == nil {
		err = ErrInvalidRequestPrincipal
		return
	}
	if len(targetServiceNames) == 0 {
		err = ErrNoTargetServiceNames
		return
	}
	if len(delegateGroups) == 0 {
		err = ErrInvalidDelegateGroups
		return
	}

	oboURI := fmt.Sprintf(OboURITemplate, endpoint)
	token, err = getOboOrDelegationToken(client, oboURI, requestPrincipal, delegatePrincipal, requestID, expiration, targetServiceNames, delegateGroups, requestTypeDelegation)
	return
}

// getOboOrDelegationToken gets a security token from the Identity Service and returns the token string.
// If there is an error, it returns an empty string and an appropriate error.
func getOboOrDelegationToken(client httpsigner.Client, oboURI string, requestPrincipal *Principal, delegatePrincipal *Principal, requestID string, expiration time.Duration, targetServiceNames []string, delegateGroups []string, requestType string) (token string, err error) {
	// verify required args
	if client == nil {
		err = ErrInvalidClient
		return
	}
	if oboURI == "" {
		err = ErrInvalidEndpoint
		return
	}
	if requestPrincipal == nil {
		err = ErrInvalidRequestPrincipal
		return
	}

	if len(targetServiceNames) == 0 {
		err = ErrNoTargetServiceNames
		return
	}
	if !(requestType == requestTypeOBO || requestType == requestTypeDelegation) {
		err = ErrInvalidRequestType
		return
	}

	caller := requestPrincipal
	var oboToken string

	if delegatePrincipal != nil {
		caller = delegatePrincipal
		if oboToken = requestPrincipal.Claims().GetSingleClaim(ClaimOBOToken).Value; oboToken == "" {
			//Invalid claim in request principal
			err = ErrInvalidRequestPrincipal
			return
		}
	}
	requestHeaders := getRequestHeaders(caller)

	oboRequest := &onBehalfOfRequest{
		RequestHeaders:     requestHeaders,
		TargetServiceNames: targetServiceNames,
		DelegateGroups:     delegateGroups,
		OboToken:           oboToken,
		Expiration:         int(expiration.Seconds()),
		RequestType:        requestType,
	}

	// Marshal the given request into JSON
	b, err := json.Marshal(*oboRequest)
	if err != nil {
		return
	}

	// Create a new request
	req, err := http.NewRequest(http.MethodPost, oboURI, bytes.NewBuffer(b))
	if err != nil {
		return
	}

	if requestID != "" {
		req.Header.Add(requestIDHeader, requestID)
	}

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	// Error if status code is not 200
	if resp.StatusCode != http.StatusOK {
		err = &ServiceResponseError{resp}
		return
	}

	// Read the response body
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// Attempt to unmarshal the response body into STSTokenResponse
	bodyBytes := []byte(body)
	var response STSTokenResponse
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return
	}

	token = response.Token
	return
}
