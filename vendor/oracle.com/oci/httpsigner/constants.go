// Copyright (c) 2017 Oracle and/or its affiliates. All rights reserved.

package httpsigner

// Header constants
const (
	HdrAuthorization = `Authorization`
	HdrDate          = `Date`
	HdrSignature     = `Signature`

	// HdrRequestTarget represents the special header name used to refer to the HTTP verb and URI in the signature.
	HdrRequestTarget = `(request-target)`

	// Host is not included in http.Request.Header, but managed as a special request field (must be lowercase)
	HdrHost = `host`
)

// Signature Field Names
const (
	SigFieldHeaders   = `headers`
	SigFieldKeyID     = `keyid`
	SigFieldSignature = `signature`
	SigFieldAlgorithm = `algorithm`
	SigFieldName      = `version`
)
