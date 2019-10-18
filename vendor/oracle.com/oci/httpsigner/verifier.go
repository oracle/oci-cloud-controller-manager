// Copyright (c) 2017-2019 Oracle and/or its affiliates. All rights reserved.

package httpsigner

import (
	"encoding/base64"
	"net/http"
	"strings"
)

// VerifyRequest will verify an http.Request that contains a Cavage http signature in its Authorization header.  The
// caller must provide a KeySupplier that can supply the public key associated with the keyid field present in the
// header and an AlgorithmSupplier that can supply an Algorithm implementation associated with the algorithm field
// present in the header.  If the signature is valid, the function returns with a nil error.  If the signature is
// invalid then an appropriate error is returned.
func VerifyRequest(request *http.Request, keySupplier KeySupplier, algSupplier AlgorithmSupplier) (err error) {

	// basic input validation
	if request == nil {
		return ErrInvalidRequest
	}
	if keySupplier == nil {
		return ErrInvalidKeySupplier
	}
	if algSupplier == nil {
		return ErrInvalidAlgorithmSupplier
	}

	// Extract and parse Authorization header
	headers, keyID, sig, algName, err := ExtractSignatureFields(request)
	if err != nil {
		return
	}

	// look up key
	key, err := keySupplier.Key(keyID)
	if err != nil {
		return
	}

	// look up algorithm
	alg, err := algSupplier.Algorithm(algName)
	if err != nil {
		return
	}

	// construct signed value
	sigTarget := getStringToSign(request, headers)

	// decode signature
	sigBytes, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		return
	}

	// delegate to algorithm
	return alg.Verify([]byte(sigTarget), sigBytes, key)
}

// ExtractSignatureFields parses the Authorization header and pulls out the headers, keyid, signature, and algorithm
// name.  Missing fields are left to their zero value.  Unknown fields are ignored. It will return ErrMissingAuthzHeader
// if no `Authorization` header is present in the request or its value is empty string.  It will return
// ErrUnsupportedScheme for any scheme other than `Signature`.
func ExtractSignatureFields(request *http.Request) (headers []string, keyID, sig, alg string, err error) {

	// pull header
	authzHdr := request.Header.Get(HdrAuthorization)
	if authzHdr == "" {
		err = ErrMissingAuthzHeader
		return
	}

	// verify auth-scheme
	hdrTokens := strings.SplitN(authzHdr, " ", 2)
	if hdrTokens[0] != HdrSignature {
		err = ErrUnsupportedScheme
		return
	}

	// pull out values (ignore unknown fields)
	if len(hdrTokens) > 1 {
		for _, field := range strings.Split(hdrTokens[1], ",") {
			key, value := extractKeyValue(strings.Trim(field, ` `))
			switch k := strings.ToLower(key); k {
			case SigFieldHeaders:
				headers = strings.Fields(value)
			case SigFieldKeyID:
				keyID = value
			case SigFieldSignature:
				sig = value
			case SigFieldAlgorithm:
				alg = value
			}
		}
	}

	return
}

// extractKeyValue will take a key value pair in the format `key="value"` and return the two fields without the quotes
func extractKeyValue(field string) (key, value string) {
	parts := strings.SplitN(field, "=", 2)
	if len(parts) > 0 {
		key = parts[0]
	}
	if len(parts) > 1 {
		value = strings.Trim(parts[1], `"`)
	}
	return
}

// RequestVerifier represents an object that a caller can use to validate an http.Request which has been signed using
// Cavage http signatures without the need to manage a KeySupplier or AlgorithmSupplier.
type RequestVerifier interface {

	// VerifyRequest will verify an http.Request that contains a Cavage http signature in its Authorization header.  If
	// the signature is valid, the function returns with a nil error.  If the signature is invalid then an appropriate
	// error is returned.
	VerifyRequest(request *http.Request) (err error)
}

// requestVerifier is an internal implementation of the RequestVerifier interface which delegates to the package
// function VerifyRequest using its member KeySupplier and AlgorithmSupplier.
type requestVerifier struct {
	keySupplier KeySupplier
	algSupplier AlgorithmSupplier
}

// VerifyRequest delegates to the package function VerifyRequest passing the KeySupplier and AlgorithmSupplier from the
// requestVerifier instance.
func (rv *requestVerifier) VerifyRequest(request *http.Request) (err error) {
	return VerifyRequest(request, rv.keySupplier, rv.algSupplier)
}

// NewRequestVerifier builds a new RequestVerifier with the supplied KeySupplier and AlgorithmSupplier.  The function
// will panic if either is nil.
func NewRequestVerifier(ks KeySupplier, as AlgorithmSupplier) RequestVerifier {
	if ks == nil || as == nil {
		panic("Programmer error: must provide both KeySupplier and AlgorithmSupplier")
	}
	return &requestVerifier{keySupplier: ks, algSupplier: as}
}
