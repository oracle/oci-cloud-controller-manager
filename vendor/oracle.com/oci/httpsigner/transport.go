// Copyright (c) 2017 Oracle and/or its affiliates. All rights reserved.

package httpsigner

import "net/http"

// Transport is an http.RoundTripper which embeddeds an actual
// http.RoundTripper and leverages a provided httpsigner.SigningClient for
// dealing with signing outgoing HTTP requests
type Transport struct {
	Client SigningClient

	transport http.RoundTripper
}

// NewTransport creates a new Transport instance with the provided
// SigningClient, and base transport. If a nil transport is provided,
// http.DefaultTransport is used.
func NewTransport(client SigningClient, transport http.RoundTripper) *Transport {
	// default to standard http RoundTripper
	if transport == nil {
		transport = http.DefaultTransport
	}

	return &Transport{
		Client:    client,
		transport: transport,
	}
}

// RoundTrip implements the http.RoundTripper interface and is responsible for
// signing the outgoing http request via the Transports SigningClient and
// delegating the actual HTTP request to the underlying RoundTripper
func (s *Transport) RoundTrip(r *http.Request) (response *http.Response, err error) {
	// Sign the request
	sreq, err := s.Client.SignRequest(r)
	if err != nil {
		return nil, err
	}

	// Delegate to underlying transport for executing the request
	return s.transport.RoundTrip(sreq)
}
