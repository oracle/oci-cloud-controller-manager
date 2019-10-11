// Copyright (c) 2019, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"crypto/rsa"
	"encoding/json"
	"time"

	"oracle.com/oci/httpsigner"
)

const (
	tokenDefaultLifeTime       = time.Hour * 2
	tokenDefaultAudience       = "oci"
	tokenResourcePrincipalType = "resource"
	// RPTFixedKeyID is the content of the kid field on an RPT ("asw" is temporary until identity supports resource/v2.2/ocid's)
	RPTFixedKeyID = "asw"
)

// ResourcePrincipalTokenClaimValues can be used to set the claims that are required to be included in the RPT blob.
// It includes the standard JWT claims and some resource principal specific ones that are listed here :
// https://confluence.oci.oraclecorp.com/pages/viewpage.action?spaceKey=ID&title=Guideline+for+creating+Resource-Principal-Token+aka+RPT+blob+for+resource+principal+v2
type ResourcePrincipalTokenClaimValues struct {
	Audience      string `json:"aud"`
	ID            string `json:"jti"`
	IssuedAt      int    `json:"iat"`
	NotBefore     int    `json:"nbf"`
	Subject       string `json:"sub"`
	Expiry        int    `json:"exp"`
	PrincipalType string `json:"ptype"`
	Issuer        string `json:"iss"`
	TokenType     string `json:"ttype"`
	ResourceType  string `json:"res_type"`
	TenantID      string `json:"res_tenant"`
	DupTenantID   string `json:"tenant"` //temporary - to remove once identity lands a fix
	CompartmentID string `json:"res_compartment"`
	ResourceID    string `json:"res_id"`
	ResourceTag   string `json:"res_tag"`
	PublicKey     string `json:"res_pbk"`
}

//RPTProvider is an interface exposing the method to get an RPT blob
type RPTProvider interface {
	GenerateRPT(spst, algName string, claims *ResourcePrincipalTokenClaimValues) (string, error)
}

//ResourcePrincipalTokenProvider contains the key and algorithm suppliers required for signing RPTs
type ResourcePrincipalTokenProvider struct {
	keySupplier       httpsigner.KeySupplier
	algorithmSupplier httpsigner.AlgorithmSupplier
}

//NewResourcePrincipalTokenProvider returns a ResourcePrincipalTokenProvider
func NewResourcePrincipalTokenProvider(keySupplier httpsigner.KeySupplier, algSupplier httpsigner.AlgorithmSupplier) *ResourcePrincipalTokenProvider {
	return &ResourcePrincipalTokenProvider{
		keySupplier:       keySupplier,
		algorithmSupplier: algSupplier,
	}
}

// buildRPTPublicKeySupplier builds a PublicKey Supplier for use in signing RPTs that must have a fixed keyID string.
func buildRPTPublicKeySupplier(fixedKeyID string, spstKeyID string, ks httpsigner.KeySupplier) (supplier httpsigner.KeySupplier, err error) {

	pk, err := ks.Key(spstKeyID)
	if err != nil {
		return nil, err
	}

	return httpsigner.NewStaticRSAPubKeySupplier(pk.(*rsa.PublicKey), fixedKeyID)
}

// buildRPTSigningKeySupplier builds a PrivateKey Supplier for use in signing RPTs that must have a fixed keyID string.
func buildRPTSigningKeySupplier(fixedKeyID string, spstKeyID string, ks httpsigner.KeySupplier) (supplier httpsigner.KeySupplier, err error) {

	pk, err := ks.Key(spstKeyID)
	if err != nil {
		if err2, ok := err.(*httpsigner.KeyRotationError); ok {
			spstKeyID = err2.ReplacementKeyID
			pk, err = ks.Key(spstKeyID)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return httpsigner.NewStaticRSAKeySupplier(pk.(*rsa.PrivateKey), fixedKeyID)
}

// GenerateRPT returns a JWT token that serves as an RPT blob required for obtaining RPST
// the RPT contains rpstPublicString embedded in the claims, and is signed by the spst privatekey provided by stsKeySupplier
// if the claims are invalid it returns the appropriate error and an empty token. if the passed signing algorithm
// is invalid it returns ErrInvalidSigningAlgorithm
// SECURITY NOTE: The provided key supplier should *only* be able to supply keys trusted for token signing.  For
// example, do not pass a composite key supplier capable of looking up both service and api keys.
func (rptp *ResourcePrincipalTokenProvider) GenerateRPT(spst string, signingAlg string, rpc *ResourcePrincipalTokenClaimValues) (string, error) {

	ec, err := rpc.EncodeClaims()
	if err != nil {
		return "", err
	}

	fks, err := buildRPTSigningKeySupplier(RPTFixedKeyID, spst, rptp.keySupplier)
	if err != nil {
		return "", err
	}

	rpt, err := generateJWT(RPTFixedKeyID, signingAlg, ec, fks, rptp.algorithmSupplier)
	if err != nil {
		return "", err
	}

	return rpt, nil
}

//EncodeClaims encodes the claims from ResourcePrincipalTokenClaimValues into a string
func (rpc *ResourcePrincipalTokenClaimValues) EncodeClaims() (string, error) {
	if rpc.Issuer == "" {
		return "", ErrInvalidClaimIssuer
	}

	if rpc.TokenType == "" {
		return "", ErrInvalidClaimTokenType
	}

	if rpc.ResourceType == "" {
		return "", ErrInvalidClaimResourceType
	}

	if rpc.TenantID == "" {
		return "", ErrInvalidClaimTenantID
	}

	if rpc.CompartmentID == "" {
		return "", ErrInvalidClaimCompartmentID
	}

	if rpc.ResourceID == "" {
		return "", ErrInvalidClaimResourceID
	}

	if rpc.ResourceTag == "" {
		return "", ErrInvalidClaimResourceTag
	}

	if rpc.PublicKey == "" {
		return "", ErrInvalidClaimPublicKey
	}

	if rpc.Expiry == 0 {
		rpc.Expiry = int(time.Now().Add(tokenDefaultLifeTime).Unix())
	}

	rpc.Audience = tokenDefaultAudience
	rpc.PrincipalType = tokenResourcePrincipalType
	uniqueID, err := tokenID(tokenIDLength)
	if err != nil {
		return "", ErrUnableToGenerateUniqueID
	}
	rpc.ID = uniqueID

	// clocks are not always in sync so we risk to set a time for issueAt and NotBefore in the future compared to when
	// identity received the request, if that happens Identity won't be able to proceed with the request.
	// clocks can't be out-of-sync badly but 2 seconds of difference should be enough
	issueAndNbf := int(time.Now().Add(-time.Second * 2).Unix())
	rpc.IssuedAt = issueAndNbf
	rpc.NotBefore = issueAndNbf

	rpc.Subject = rpc.ResourceID
	rpc.DupTenantID = rpc.TenantID // temporary - to remove when identity lands a fix

	rpclaims, err := json.Marshal(rpc)
	if err != nil {
		return "", err
	}

	return string(rpclaims), nil
}

//parseRPT parses a raw RPT string and returns a token of the type Token
func parseRPT(rawToken string, spstKeyID string, as httpsigner.AlgorithmSupplier, ks httpsigner.KeySupplier) (token *Token, err error) {

	rptks, err := buildRPTPublicKeySupplier(RPTFixedKeyID, spstKeyID, ks)
	if err != nil {
		return
	}
	return ParseToken(rawToken, rptks, as)
}
