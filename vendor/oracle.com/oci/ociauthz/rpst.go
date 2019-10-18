// Copyright (c) 2019, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"

	"oracle.com/oci/httpsigner"
)

// RPSTProvider is an interface exposing the method to get a resource principal session token
type RPSTProvider interface {
	GetRPST(*ResourcePrincipalTokenClaimValues) (string, error)
}

// request to be passed to identity in the call to get RPST token
type resourcePrincipalSessionTokenRequest struct {
	ResourcePrincipalToken       string `json:"resourcePrincipalToken,omitempty"`
	ServicePrincipalSessionToken string `json:"servicePrincipalSessionToken,omitempty"`
	SessionPublicKey             string `json:"sessionPublicKey,omitempty"`
}

func newRPSTRequest(rptSignedToken string, spst string, rpstPublicKey string) *resourcePrincipalSessionTokenRequest {

	return &resourcePrincipalSessionTokenRequest{
		ResourcePrincipalToken:       rptSignedToken,
		ServicePrincipalSessionToken: spst,
		SessionPublicKey:             rpstPublicKey,
	}
}

// ResourcePrincipalSessionTokenProvider contains the stsprovider thst signs the rpt, client that signs the http request to identity, and the RPT provider
// to generate the RPT to make RPST request with.
type ResourcePrincipalSessionTokenProvider struct {
	mx sync.Mutex
	// Use this client to sign http request to get RPST from identity
	stsClient httpsigner.Client
	// Use this key supplier to sign RPT and retrieve SPST
	stsKeySupplier                        *STSKeySupplier
	resourcePrincipalSessionTokenEndpoint string
	signingAlgorithm                      string
	rptProvider                           *ResourcePrincipalTokenProvider
}

// NewResourcePrincipalSessionTokenProvider creates and returns a new RPST provider. it takes as input
// 1. the endpoint to make the request at and get the RPST from
// 2. A key supplier that provides the private key for the signature on the RPT
// 3. An algorithm supplier that provides the algorithm for signing the RPT
// 3. a signing Client that signs the http request to this endpoint
// 4. a JWT standard format string as defined in https://tools.ietf.org/html/rfc7518#section-3 that specifies the signing algorithm.
func NewResourcePrincipalSessionTokenProvider(rpstEndpoint string, stsKeySupplier *STSKeySupplier, algSupplier httpsigner.AlgorithmSupplier, stsClient httpsigner.Client, signingAlgString string) *ResourcePrincipalSessionTokenProvider {
	rptp := NewResourcePrincipalTokenProvider(stsKeySupplier, algSupplier)
	return &ResourcePrincipalSessionTokenProvider{
		resourcePrincipalSessionTokenEndpoint: rpstEndpoint,
		stsClient:        stsClient,
		stsKeySupplier:   stsKeySupplier,
		signingAlgorithm: signingAlgString,
		rptProvider:      rptp,
	}
}

// GetRPST takes as input the claims required for generating an RPT and returns a
// Resource principal session token by making a request to identity with the generated RPT
func (rpstp *ResourcePrincipalSessionTokenProvider) GetRPST(rpc *ResourcePrincipalTokenClaimValues) (string, error) {

	// 1. get the service principal session token
	spst, err := rpstp.stsKeySupplier.KeyID()
	if err != nil {
		return "", err
	}

	// 2. Generate the RPT blob (jwt token with claims embedded and signed by spst private key)
	signedRPT, err := rpstp.rptProvider.GenerateRPT(spst, rpstp.signingAlgorithm, rpc)
	if err != nil {
		return "", err
	}

	// 3. Make a call to identity to get RPST
	rpstRequest := newRPSTRequest(signedRPT, spst, rpc.PublicKey)

	rpst, err := rpstp.getRPSTFromIdentity(rpstRequest)
	if err != nil {
		return "", err
	}

	return rpst, nil

}

// getRPSTFromIdentity uses the resourcePrincipalSessionTokenRequest and makes a http request to identity to get the rpst back and returns it.
// The http request is signed by STSClient in the ResourcePrincipalSessionTokenProvider.
func (rpstp *ResourcePrincipalSessionTokenProvider) getRPSTFromIdentity(rpstRequest *resourcePrincipalSessionTokenRequest) (string, error) {

	b, err := json.Marshal(rpstRequest)
	if err != nil {
		return "", err
	}
	rpstHTTPRequest, err := http.NewRequest("POST", rpstp.resourcePrincipalSessionTokenEndpoint, bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}

	rpstHTTPResponse, err := rpstp.stsClient.Do(rpstHTTPRequest)
	if err != nil {
		return "", err
	}

	if rpstHTTPResponse.StatusCode != http.StatusOK {
		err = &ServiceResponseError{Response: rpstHTTPResponse}
		return "", err
	}

	defer rpstHTTPResponse.Body.Close()
	body, err := ioutil.ReadAll(rpstHTTPResponse.Body)
	if err != nil {
		return "", err
	}

	var result S2SResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	return result.Token, nil
}
