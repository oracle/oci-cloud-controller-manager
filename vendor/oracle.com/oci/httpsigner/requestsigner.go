// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package httpsigner

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

// Header templates
const (
	// AuthzV1HdrTempl is a template for the Authorization header Version 1
	AuthzV1HdrTempl string = `Signature version="1",headers="%s",keyId="%s",algorithm="%s",signature="%s"`
)

// DefaultAuthzHdrTempl is the default Authorization Header Template.
// The default is set to AuthzV1HdrTempl (Version 1)
var defaultAuthzHdrTempl = AuthzV1HdrTempl

// SignRequest signs an http.Request per draft-cavage-http-signatures-06
// (https://tools.ietf.org/html/draft-cavage-http-signatures-06) using the given key, header list, and signing
// algorithm, adds an `Authorization` header to the request with the signature, and returns a pointer to the updated
// request.  It will return ErrKeyNotFound if the keySupplier is not able to return a key with the given key ID.
func SignRequest(request *http.Request, keyID string, keySupplier KeySupplier, headersToSign []string, algorithm Algorithm) (sreq *http.Request, err error) {
	err = checkRequiredArgs(request, keyID, keySupplier, algorithm)

	if err != nil {
		return
	}

	// lookup the key
	key, err := keySupplier.Key(keyID)
	if err != nil {
		return
	}

	// default an empty headers list to 'Date' per https://tools.ietf.org/html/draft-cavage-http-signatures-06#section-2.1.3
	if len(headersToSign) == 0 {
		headersToSign = []string{"Date"}
	}

	// generate signature
	stringtoSign := getStringToSign(request, headersToSign)
	sig, err := algorithm.Sign([]byte(stringtoSign), key)
	if err != nil {
		return
	}

	// prepare/attach Authorization header
	encodedSig := base64.StdEncoding.EncodeToString(sig)
	authzHdrVal := fmt.Sprintf(defaultAuthzHdrTempl, concatenateHeaders(headersToSign), keyID, algorithm.Name(), encodedSig)
	request.Header.Add(HdrAuthorization, authzHdrVal)

	return request, nil
}

// checkRequiredArgs returns an error if any required arguments are empty/nil
func checkRequiredArgs(request *http.Request, keyID string, keySupplier KeySupplier, algorithm Algorithm) error {
	if request == nil {
		return ErrInvalidRequest
	}
	if keyID == "" {
		return ErrInvalidKeyID
	}
	if keySupplier == nil {
		return ErrInvalidKeySupplier
	}
	if algorithm == nil {
		return ErrInvalidAlgorithm
	}
	return nil
}

// getStringToSign generates the `message` string used as input to the signing algorithm.
// The `message` includes header names and values from the given headers list
// It also includes the value of the special (request-target) header field name
func getStringToSign(request *http.Request, headersToSign []string) string {
	stringToSign := ""
	for _, header := range headersToSign {
		if stringToSign != "" {
			stringToSign += "\n"
		}

		if header == HdrRequestTarget {
			stringToSign += fmt.Sprintf("%s: %s", strings.ToLower(header), getRequestTarget(request))
		} else if strings.ToLower(header) == HdrHost {
			// Host header is specially handled as a request field
			stringToSign += fmt.Sprintf("%s: %s", HdrHost, request.Host)
		} else {
			stringToSign += fmt.Sprintf("%s: %s", strings.ToLower(header), request.Header.Get(header))
		}
	}
	return stringToSign
}

// getRequestTarget returns the value of the special (request-target) header field name
// per https://tools.ietf.org/html/draft-cavage-http-signatures-06#section-2.3
func getRequestTarget(request *http.Request) string {
	lowercaseMethod := strings.ToLower(request.Method)
	return fmt.Sprintf("%s %s", lowercaseMethod, request.URL.RequestURI())
}

// concatenateHeaders returns a space delimited, lowercased string of headers
func concatenateHeaders(headers []string) (concatenated string) {
	for _, header := range headers {
		if len(concatenated) > 0 {
			concatenated += " "
		}
		concatenated += strings.ToLower(header)
	}
	return
}

// RequestSigner represents an object that can sign a request with a predetermined KeySupplier and signing algorithm,
// reducing the state to be managed by the consumer of the signer.
type RequestSigner interface {

	// SignRequest will sign the specified headers of a request with the key identified by keyID using a predetermined
	// algorithm.  It will return an error ErrKeyNotFound if the predetermined KeySupplier is unable to find the key
	// associated with keyID. It may also return ErrInvalidArg.
	SignRequest(request *http.Request, keyID string, headersToSign []string) (signedReq *http.Request, err error)
}

// requestSigner is an internal struct to hold a KeySupplier and Algorithm it delegates to the package function for
// signing.
type requestSigner struct {
	keySupplier KeySupplier
	algorithm   Algorithm
}

// SignRequest delegates to the package function SignRequest passing the KeySupplier and Algorithm from the
// requestSigner instance.
func (rs *requestSigner) SignRequest(req *http.Request, keyID string, headersToSign []string) (*http.Request, error) {
	return SignRequest(req, keyID, rs.keySupplier, headersToSign, rs.algorithm)
}

// NewRequestSigner will build a basic RequestSupplier given a KeySupplier and Algorithm.  Neither can be nil.
func NewRequestSigner(keySupplier KeySupplier, algorithm Algorithm) RequestSigner {
	if keySupplier == nil || algorithm == nil {
		panic("Programmer error, must provide both keySupplier and algorithm.")
	}
	return &requestSigner{keySupplier: keySupplier, algorithm: algorithm}
}
