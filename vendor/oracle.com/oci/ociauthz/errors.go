// Copyright (c) 2018-2019, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"errors"
	"fmt"
	"net/http"
)

// Error Definitions
var (
	// ErrInvalidArg error means a function argument is invalid
	ErrInvalidArg = errors.New("invalid argument")

	// X509Supplier invalid arg errors
	ErrInvalidCertificate         = errors.New("certificate must be a non-nil value")
	ErrInvalidIntermediateCerts   = errors.New("intermediate certificates must be a non-nil value")
	ErrInvalidPrivateKey          = errors.New("private key must be a non-nil value")
	ErrInvalidTenantID            = errors.New("tenant id must be a non-empty string")
	ErrInvalidCertPEM             = errors.New("certificate PEM data is invalid")
	ErrInvalidIntermediateCertPEM = errors.New("intermediate certificate PEM data is invalid")
	ErrInvalidPrivateKeyPEM       = errors.New("private Key PEM data is invalid")

	// STSKeySupplier invalid arg errors
	ErrInvalidX509Supplier = errors.New("x509Supplier must be a non-nil value")
	ErrInvalidClient       = errors.New("client must be a non-nil value")
	ErrInvalidEndpoint     = errors.New("endpoint must be a non-empty string")

	// Key Service errors
	ErrKeyNotFound          = errors.New("key not found")
	ErrUnsupportedKeyFormat = errors.New("the format of the requested keyID is not supported")

	// JWT errors
	ErrJWTMalformed     = errors.New("jWT token should consist of three parts")
	ErrTokenNotValidYet = errors.New("token not valid yet")
	ErrTokenExpired     = errors.New("token expired")

	// JWK errors
	ErrUnsupportedJWKType      = errors.New("the JWK key type is not supported")
	ErrNoJWK                   = errors.New("token does not contain a JWK")
	ErrInvalidJWK              = errors.New("invalid JWK")
	ErrUnsupportedExponentSize = errors.New("unsupported public key exponent size")

	// PKIX errors
	ErrPEMDecodeError = errors.New("failed to decode public key PEM block")

	// CheckRequiredHeaders
	ErrRequiredHeaderMissing = errors.New("required headers missing from the request object")

	// Model errors
	ErrInvalidToken = errors.New("token must be a non-nil value")

	// Util errors
	ErrParsePEM             = errors.New("failed to parse PEM")
	ErrExpectedRSAPublicKey = errors.New("couldn't cast key to rsa.PublicKey")

	// Authz
	ErrInvalidOperationID   = errors.New("operationID must be a non-empty value")
	ErrInvalidCompartmentID = errors.New("compartmentID must be a non-empty value")
	ErrInvalidRequestRegion = errors.New("requestRegion must be a non-empty value")
	ErrInvalidPhysicalAD    = errors.New("physicalAD must be a non-empty value")
	ErrInvalidActionKind    = errors.New("invalid ActionKind")
	ErrInvalidPrincipal     = errors.New("either UserPrincipal or ServicePrincipal must be a non-nil value")
	ErrNoPermissionsSet     = errors.New("authorization request must contain permissions")
	ErrInvalidCtxVarType    = errors.New("invalid context variable type")

	// Associations
	ErrAssociationInsufficientRequests          = errors.New("association call must include at least 2 authorization requests")
	ErrUnexpectedAssociationAuthzResponseLength = errors.New("authorization response does not equal the length of authorization requests inside association request")

	// OBO
	ErrInvalidRequestPrincipal = errors.New("requestPrincipal must be a non-nil value")
	ErrNoTargetServiceNames    = errors.New("targetServiceNames must be non-empty")
	ErrInvalidRequestType      = errors.New("requestType must be 'OBO' or 'DELEGATION'")
	ErrInvalidDelegateGroups   = errors.New("delegateGroups must not be empty")
)

// ServiceResponseError is thrown if the response from an external service returns a non-2xx status code.
// The error contains the response object so that the caller may choose to inspect the response if needed.
type ServiceResponseError struct {
	Response *http.Response
}

const serviceResponseErrorDefault = "unknown"

var serviceResponseTemplate = `Unexpected status code of '%s' returned from %s. OPC-Request-ID: %s`

// Error returns a string which describes the error
func (e ServiceResponseError) Error() string {
	statusCode := serviceResponseErrorDefault
	target := serviceResponseErrorDefault
	requestID := serviceResponseErrorDefault

	if e.Response != nil {
		if e.Response.Status != "" {
			statusCode = e.Response.Status
		}
		if e.Response.Header != nil {
			if x := e.Response.Header.Get("opc-request-id"); x != "" {
				requestID = x
			}
		}
		request := e.Response.Request
		if request != nil {
			if request.URL != nil {
				target = e.Response.Request.URL.String()
			}
		}
	}

	return fmt.Sprintf(serviceResponseTemplate, statusCode, target, requestID)
}
