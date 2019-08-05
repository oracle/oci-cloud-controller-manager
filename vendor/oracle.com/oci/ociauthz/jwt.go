// Copyright (c) 2018, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"encoding/base64"
	"encoding/json"
	"oracle.com/oci/httpsigner"
	"strings"
	"time"
)

// HdrClaimPrefix is a prefix added to a claim key that came from the request header
// See: https://bitbucket.oci.oraclecorp.com/projects/IDENT/repos/authorization-sdk/browse/authentication-common/src/main/java/com/oracle/pic/identity/authentication/Constants.java#82
const HdrClaimPrefix = "h_"

// HdrClaimIssuer is an Issuer value to use if the claim came from request header.
// See: https://bitbucket.oci.oraclecorp.com/projects/IDENT/repos/authorization-sdk/browse/authentication-common/src/main/java/com/oracle/pic/identity/authentication/Constants.java#41
const HdrClaimIssuer = "h"

// Standard claims (from https://tools.ietf.org/html/rfc7519)
var (
	ClaimIssuer    = "iss"
	ClaimSubject   = "sub"
	ClaimAudience  = "aud"
	ClaimExpires   = "exp"
	ClaimNotBefore = "nbf"
	ClaimIssuedAt  = "iat"
	ClaimJwtID     = "jti"
	ClaimJWK       = "jwk"
)

// Additional claims (see: https://bitbucket.oci.oraclecorp.com/projects/IDENT/repos/authorization-sdk/browse/authentication-common/src/main/java/com/oracle/pic/identity/authentication/ClaimType.java)
var (
	ClaimCN                        = "CN"
	ClaimOrgUnit                   = "OU"
	ClaimOrg                       = "O"
	ClaimIssuerCN                  = "IssuerCN"
	ClaimSerial                    = "Serial"
	ClaimEmail                     = "Email"
	ClaimNotBeforeCN               = "NotBefore"
	ClaimNotAfterCN                = "NotAfter"
	ClaimMFAVerified               = "mfa_verified"
	ClaimServiceName               = "svc"
	ClaimFingerprint               = "fprint"
	ClaimPrincipalType             = "ptype"
	ClaimPrincipalSubType          = "pstype"
	ClaimTokenType                 = "ttype"
	ClaimTenant                    = "tenant"
	ClaimTargetTenant              = "tgt"
	ClaimTargetTenantIDs           = "tgts"
	ClaimTargetServiceName         = "tgt_name"
	ClaimTargetServiceNames        = "tgt_names"
	ClaimDelegateGroups            = "dgrps"
	ClaimCallChain                 = "chain"
	ClaimPreviousTokenID           = "pti"
	ClaimCurrentTokenID            = "jti"
	ClaimOwner                     = "own"
	ClaimFederatedUserGroups       = "grps"
	ClaimOBOToken                  = "obo_tk"
	ClaimBody                      = "body"
	ClaimSessionExpiration         = "sess_exp"
	ClaimCrossTenancyRequestHeader = HdrClaimPrefix + "x-cross-tenancy-request"
	ClaimSubscriptionHeader        = HdrClaimPrefix + "x-subscription"

	// deprecated
	ClaimFingerPrint = "fprint"
)

// Sparta Cert Data (see: https://confluence.oci.oraclecorp.com/display/SP/Sparta+Cert+Data)
var (
	ClaimOPCCertType    = "opc-certtype"
	ClaimOPCInstance    = "opc-instance"
	ClaimOPCCompartment = "opc-compartment"
	ClaimOPCHostname    = "opc-hostname"
	ClaimOPCTag         = "opc-tag"
	ClaimOPCBump        = "opc-bump"
	ClaimOPCEpoch       = "opc-epoch"
	ClaimOPCIdentity    = "opc-identity"
	ClaimOPCTenant      = "opc-tenant"
)

// tokenVerificationPolicy type to perform token validation
type tokenVerificationPolicy func(string) bool

// alwaysVerifyTokenPolicy verify token always
var alwaysVerifyTokenPolicy = func(token string) bool { return true }

// Token represents a JSON Web Token (JWT)
// See: https://tools.ietf.org/html/rfc7519
type Token struct {
	Header Header
	Claims Claims
}

// NewToken returns a new JWT Token instance using a token string, failing to validate a token returns an error.
// SECURITY NOTE: The provided key supplier should *only* be able to supply keys trusted for token signing.  For
// example, do not pass a composite key supplier capable of looking up both service and api keys.
func NewToken(token string, ks httpsigner.KeySupplier) (*Token, error) {
	return newTokenWithTokenVerificationPolicy(token, ks, alwaysVerifyTokenPolicy)
}

// newTokenWithTokenVerificationPolicy returns a new JWT Token instance using a token string
func newTokenWithTokenVerificationPolicy(token string, ks httpsigner.KeySupplier, shouldVerifyFn tokenVerificationPolicy) (*Token, error) {
	if token == "" {
		return nil, ErrInvalidArg
	}

	return parseTokenWithTokenVerificationPolicy(token, ks, httpsigner.JWTAlgorithms, shouldVerifyFn)
}

// ValidFor checks to see if token is in the validity period for the time given.
func (t *Token) ValidFor(clock time.Time) error {

	// issued for the future?
	nbf, err := t.Claims.GetInt(ClaimNotBefore)
	if err != nil {
		return err
	}
	if nbf > 0 {
		if clock.Before(time.Unix(nbf, 0)) {
			return ErrTokenNotValidYet
		}
	}

	// expired?
	exp, err := t.Claims.GetInt(ClaimExpires)
	if err != nil {
		return err
	}
	if clock.After(time.Unix(exp, 0)) {
		return ErrTokenExpired
	}

	return nil
}

// Header represents the decoded header part of the token
type Header struct {
	KeyID     string `json:"kid,omitempty"`
	Algorithm string `json:"alg,omitempty"`
}

// ParseToken takes an encoded JWT, verifies its signature, and if valid, decodes it into a Token instance.  The
// keySupplier must be able to provide the public key associated with the kid field of the JWT header.  If no suitable
// key can be found, httpsigner.ErrKeyNotFound is returned.  If the signature is invalid, rsa.ErrVerification is
// returned.  If the JWT content fails to decode an appropriate error message is returned.
// Deprecated use parseTokenWithTokenVerificationPolicy
// SECURITY NOTE: The provided key supplier should *only* be able to supply keys trusted for token signing.  For
// example, do not pass a composite key supplier capable of looking up both service and api keys.
func ParseToken(rawToken string, ks httpsigner.KeySupplier, as httpsigner.AlgorithmSupplier) (token *Token, err error) {
	return parseTokenWithTokenVerificationPolicy(rawToken, ks, as, alwaysVerifyTokenPolicy)
}

// parseTokenWithTokenVerificationPolicy takes an encoded JWT, verifies its signature, and if valid, decodes it into a Token instance.  The
// keySupplier must be able to provide the public key associated with the kid field of the JWT header.  If no suitable
// key can be found, httpsigner.ErrKeyNotFound is returned.  The signature is verified based on the return of the tokenVerificationPolicy and
// If it is invalid, rsa.ErrVerification is returned.
// If the JWT content fails to decode an appropriate error message is returned.
// SECURITY NOTE: The provided key supplier should *only* be able to supply keys trusted for token signing.  For
// example, do not pass a composite key supplier capable of looking up both service and api keys.
func parseTokenWithTokenVerificationPolicy(rawToken string, ks httpsigner.KeySupplier, as httpsigner.AlgorithmSupplier, shouldVerify tokenVerificationPolicy) (token *Token, err error) {

	header, body, signature, err := extractParts(rawToken)
	if err != nil {
		return
	}

	// Verify by default(zero value present)
	if shouldVerify == nil || shouldVerify(rawToken) {
		err = verifyToken(header, signature, rawToken, ks, as)
		if err != nil {
			return
		}
	}

	// parse claims
	var claims Claims
	claims, err = UnmarshalClaims(body)
	if err != nil {
		return
	}

	return &Token{
		Claims: claims,
		Header: header,
	}, nil
}

func extractParts(rawToken string) (header Header, body []byte, signature []byte, err error) {
	parts := strings.Split(rawToken, ".")
	if len(parts) != 3 {
		err = ErrJWTMalformed
		return
	}

	// Parse header
	decodedHdr, err := decodeTokenPart(parts[0])
	if err != nil {
		return
	}

	err = json.Unmarshal(decodedHdr, &header)
	if err != nil {
		return
	}

	//Parse the payload
	body, err = decodeTokenPart(parts[1])
	if err != nil {
		return
	}

	//Parse the signature
	signature, err = decodeTokenPart(parts[2])
	if err != nil {
		return
	}
	return
}

// verifyToken takes an encoded JWT, KeySupplier, and Algorithm supplier, verifies the token's signature and if invalid
// it returns an error
// SECURITY NOTE: The provided key supplier should *only* be able to supply keys trusted for token signing.  For
// example, do not pass a composite key supplier capable of looking up both service and api keys.
func verifyToken(header Header, signature []byte, token string, ks httpsigner.KeySupplier, as httpsigner.AlgorithmSupplier) (err error) {

	if signature == nil ||
		ks == nil ||
		as == nil {
		return ErrInvalidArg
	}

	payloadDelimeter := strings.LastIndex(token, ".")
	if payloadDelimeter == -1 {
		return ErrJWTMalformed
	}

	// fetch key
	key, err := ks.Key(header.KeyID)
	if err != nil {
		return
	}

	// fetch algorithm
	alg, err := as.Algorithm(header.Algorithm)
	if err != nil {
		return
	}

	// header plus payload
	message := []byte(token[0:payloadDelimeter])

	// check signature
	err = alg.Verify(message, signature, key)
	return err
}

// decodeTokenPart decodes part of the JSON Web Token
func decodeTokenPart(part string) ([]byte, error) {
	// The token part is base64 URL encoded, which may need some padding in order
	// to be understood by the parser.
	// See: https://en.wikipedia.org/wiki/Base64#URL_applications
	if pad := len(part) % 4; pad > 0 {
		part += strings.Repeat("=", 4-pad)
	}

	return base64.URLEncoding.DecodeString(part)
}
