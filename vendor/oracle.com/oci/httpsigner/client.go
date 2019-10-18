// Copyright (c) 2017 Oracle and/or its affiliates. All rights reserved.

package httpsigner

import (
	"net/http"
)

// Client is an interface that exposes some of the same methods implemented by http.Client to allow extension.
type Client interface {
	Do(request *http.Request) (*http.Response, error)
}

// SigningClient extends Client to expose the signing mechanism.
type SigningClient interface {
	Client
	SignRequest(request *http.Request) (*http.Request, error)
}

// PrepareRequestFn is a type alias to a function that should be used to perform the work of request signing.  This function
// type is called from DefaultSigningClient.SignRequest to allow extended clients to modify the request for signing.
// This function should return the modified request, list of HTTP headers the signer should sign, and error if any.
type PrepareRequestFn func(request *http.Request) (prepared *http.Request, headers []string, err error)

// DefaultSigningClient is a simple SigningClient wrapper around the standard library http.Client to provide transparent
// signing of outbound requests.  It can be used as an alternative to http.Client, but not as a direct replacement since
// http.Client is a struct.
type DefaultSigningClient struct {
	signer     RequestSigner
	keyID      string
	httpClient *http.Client

	prepareRequestFn PrepareRequestFn
}

// DefaultHeadersToSign is used by Client to pass to the underlying signer
var DefaultHeadersToSign = []string{HdrDate, HdrRequestTarget}

// Do signs the given request before delegating to the underlying http.Client.  May return any errors produced by either
// http.Client.Do() or the underlying RequestSigner.  The client does not modify the headers other than to add an
// Authorization header and expects Date to be populated.
func (c *DefaultSigningClient) Do(request *http.Request) (*http.Response, error) {

	// Sign the request
	sreq, err := c.SignRequest(request)
	if err != nil {
		return nil, err
	}

	// Delegate to underlying http.Client
	return c.httpClient.Do(sreq)
}

// SignRequest signs the proffered request using the client's RequestSigner and keyID. It also provides an extension
// point for adopters to add custom logic.
func (c *DefaultSigningClient) SignRequest(request *http.Request) (*http.Request, error) {
	// Prepare request
	request, headers, err := c.prepareRequestFn(request)
	if err != nil {
		return nil, err
	}

	// Sign request
	signed, err := c.signer.SignRequest(request, c.keyID, headers)

	// Handle key expiration
	if expired, ok := err.(*KeyRotationError); ok {
		if expired.ReplacementKeyID == "" {
			return nil, ErrReplacementKeyIDEmpty
		}
		c.keyID = expired.ReplacementKeyID
		return c.signer.SignRequest(request, c.keyID, headers)
	}

	return signed, err
}

// Signer returns the request signer attached to this client
func (c *DefaultSigningClient) Signer() RequestSigner {
	return c.signer
}

// KeyID returns the keyid associated with this client
func (c *DefaultSigningClient) KeyID() string {
	return c.keyID
}

// SetKeyID allows override of the current keyid associated with the client
func (c *DefaultSigningClient) SetKeyID(keyID string) {
	c.keyID = keyID
}

// defaultPrepareRequestFn is the default function called for preparing http.Request for signing. It simply returns the given *http.Request,
// and httpsigner.DefaultHeadersToSign.
func defaultPrepareRequestFn(request *http.Request) (*http.Request, []string, error) {
	return request, DefaultHeadersToSign, nil
}

// NewSimpleClient builds a new request client from a signer and a keyID with a default http client and the default request signing method.
// A nil signer will result in a panic.
func NewSimpleClient(signer RequestSigner, keyID string) SigningClient {
	if signer == nil {
		panic(`Programmer Error: must provide a non-nil signer.`)
	}
	return &DefaultSigningClient{signer, keyID, &http.Client{}, defaultPrepareRequestFn}
}

// NewClient builds a new request client from a signer and a keyID.  A nil signer, client, or prepareRequestFn will result in a panic.
func NewClient(signer RequestSigner, keyID string, client *http.Client, prepareRequestFn PrepareRequestFn) SigningClient {
	if signer == nil {
		panic(`Programmer Error: must provide a non-nil signer.`)
	}
	if client == nil {
		panic(`Programmer Error: must provide a non-nil client.`)
	}
	if prepareRequestFn == nil {
		panic(`Programmer Error: must provide a non-nil prepareRequestFn.`)
	}
	return &DefaultSigningClient{signer, keyID, client, prepareRequestFn}
}
