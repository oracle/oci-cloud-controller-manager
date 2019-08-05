// Copyright (c) 2018, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"oracle.com/oci/httpsigner"
)

// Principal implements Subject from httpiam
type Principal struct {
	subject  string
	tenantID string
	delegate *Principal
	claims   Claims
}

const subjectUndetected = "<subject-not-specified>"
const hdrAuthorization = "Authorization"

// These constants define types of principal
const (
	PrincipalTypeUser     = "user"
	PrincipalTypeService  = "service"
	PrincipalTypeInstance = "instance"
)

// These constants define sub types of principals
const (
	PrincipalSubTypeFederated = "fed"
)

// InvalidClaimType represents a string to use when the type of the claim value is not supported
const InvalidClaimType = "<INVALID_CLAIM_TYPE>"

// NewPrincipal creates a new Principal with the given subject and tenant IDs.
func NewPrincipal(subject string, tenantID string) *Principal {
	if subject == "" {
		subject = subjectUndetected
	}
	return &Principal{subject: subject, tenantID: tenantID}
}

// NewPrincipalFromToken creates a new Principal with the given JWT token.  The caller may optionally provide a
// delegate.
func NewPrincipalFromToken(token *Token, delegate *Principal) (*Principal, error) {
	if token == nil {
		return nil, ErrInvalidToken
	}
	subject := subjectFromClaims(token.Claims)

	return &Principal{
		subject:  subject,
		tenantID: token.Claims.GetString(ClaimTenant),
		delegate: delegate,
		claims:   token.Claims,
	}, nil
}

// NewPrincipalFromTokenAndRequest creates a new Principal with the given JWT token and http request.  Headers are
// pulled out of the request and will be added as claims.  The caller may optionally provide a delegate.
func NewPrincipalFromTokenAndRequest(token *Token, delegate *Principal, request *http.Request) (*Principal, error) {
	p, err := NewPrincipalFromToken(token, delegate)
	if err != nil {
		return nil, err
	}
	err = p.AddHeaderClaims(request)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// ID returns the subject ID of the principal
func (p Principal) ID() string {
	return p.subject
}

// TenantID returns the tenant ID of the principal
func (p Principal) TenantID() string {
	return p.tenantID
}

// Claims returns the claims attached to the principal
func (p *Principal) Claims() Claims {
	return p.claims
}

// Delegate returns the Principal acting on behalf of this principal if any.
func (p *Principal) Delegate() *Principal {
	return p.delegate
}

// Type returns the principal type of this principal.  In absence of JWT claims or empty PrincipalType field,
// it will default to PrincipalTypeUser
func (p *Principal) Type() string {
	if p.Claims() == nil {
		return PrincipalTypeUser
	}
	ptype := p.Claims().GetString(ClaimPrincipalType)
	if ptype == "" {
		return PrincipalTypeUser
	}
	return ptype
}

// getRequestTarget returns the value of the special (request-target) header field name
// per https://tools.ietf.org/html/draft-cavage-http-signatures-06#section-2.3
// TODO use the copy of this function from httpsigner after making it public there
func getRequestTarget(request *http.Request) string {
	lowercaseMethod := strings.ToLower(request.Method)
	return fmt.Sprintf("%s %s", lowercaseMethod, request.URL.RequestURI())
}

// AddHeaderClaims adds signed headers from the given http.Request as Claims in the principal struct
func (p *Principal) AddHeaderClaims(r *http.Request) error {
	headers, _, _, _, err := httpsigner.ExtractSignatureFields(r)
	if err != nil {
		return err
	}
	headers = append(headers, hdrAuthorization)
	for _, name := range headers {
		var value string
		switch name {
		case httpsigner.HdrRequestTarget:
			value = getRequestTarget(r)
		case httpsigner.HdrHost:
			value = r.Host
		default:
			value = r.Header.Get(name)
		}
		claimKey := fmt.Sprintf("%s%s", HdrClaimPrefix, strings.ToLower(name))
		p.AddClaim(Claim{HdrClaimIssuer, claimKey, value})
	}
	return nil
}

// AddClaim adds single claim struct to the principal
func (p *Principal) AddClaim(claim Claim) {
	if p.claims == nil {
		p.claims = Claims{}
	}
	p.Claims().Add(claim)
}

// subjectFromClaims return the best subject name based on fields found in Claims
func subjectFromClaims(claims Claims) string {
	subject := claims.GetString(ClaimSubject)
	if subject == "" {
		subject = subjectUndetected
	}
	return subject
}

// Claim is a representation of a JWT claim
type Claim struct {
	Issuer string `json:"issuer"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

// IsEmpty returns true if the claim struct is an empty value
func (c Claim) IsEmpty() bool {
	return c == Claim{}
}

// Claims represents a collection of JWT claims
type Claims map[string][]Claim

// GetSingleClaim returns single claim given a claim type.
func (c Claims) GetSingleClaim(key string) Claim {
	claims := c[key]
	if len(claims) > 0 {
		return claims[0]
	}
	return Claim{}
}

// GetString returns the claim value given a claim type.
func (c Claims) GetString(key string) string {
	claim := c.GetSingleClaim(key)
	return claim.Value
}

// GetInt returns the claim value which corresponds to the given key.  The value will be coerced into an int.  If the
// key is not found then 0 will be returned.  If the value cannot be represented as an int, then a zero-value and
// an error from strconv.ParseInt are returned.
func (c Claims) GetInt(key string) (result int64, err error) {
	value := c.GetString(key)
	if value == "" {
		return 0, nil
	}
	result, err = strconv.ParseInt(value, 10, 64)
	return
}

// Add adds the given claim to the Claims object
func (c Claims) Add(claim Claim) {
	c[claim.Key] = append(c[claim.Key], claim)
}

// ToSlice returns the Claims collection as a slice of Claim structs
func (c Claims) ToSlice() []Claim {
	result := make([]Claim, 0, len(c))
	for _, claims := range c {
		result = append(result, claims...)
	}
	return result
}

// UnmarshalClaims unmarshals the given JWT data and return Claims object
func UnmarshalClaims(data []byte) (Claims, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	var decoded map[string]interface{}
	if err := decoder.Decode(&decoded); err != nil {
		return nil, err
	}

	normalized := make(map[string]string, len(decoded))
	for key, value := range decoded {
		switch v := value.(type) {
		case string:
			normalized[key] = v
		case json.Number:
			normalized[key] = v.String()
		case bool:
			normalized[key] = strconv.FormatBool(v)
		default:
			normalized[key] = InvalidClaimType
		}
	}

	issuer := normalized[ClaimIssuer]
	claims := make(Claims, len(normalized))
	for k, v := range normalized {
		claims.Add(Claim{issuer, k, v})
	}
	return claims, nil
}
