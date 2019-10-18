package types

import (
	"bytes"
	"encoding/base64"
	"encoding/json"

	"oracle.com/oci/ociauthz"
)

const (
	// These consts represent resources to be acted up on by API requests.
	// They are used in the context of authorization requests made by TM.
	ClusterAuthResource    = "cluster"
	AuthConfigAuthResource = "authConfig"
	WorkItemAuthResource   = "workItem"
	BMCAuthResource        = "bmc"
	LimitAuthResource      = "limits"

	// This resource type is associated service metadata
	OKEAuthResource    = "oke"
	DefaultContextType = "OCI"
)

// TMAuthenticationRequest is the payload to submit to an authentication
// provider such as identity adaptor
type TMAuthenticationRequest struct {
	Token string `json:"token"`
}

type Org struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type User struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// TMAuthenticationResponse is the response type that is received from an
// authentication provider such as identity adaptor
type TMAuthenticationResponse struct {
	Authenticated bool   `json:"authenticated"`
	UserName      string `json:"username"`
	Error         string `json:"error"`
	User          User   `json:"user"`
	Orgs          []Org  `json:"orgs"`
}

// TMAuthorizationRequest is the payload to submit to an authorization
// provider such as identity adaptor.
type TMAuthorizationRequest struct {
	Token    string `json:"token"`
	Group    string `json:"group"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// TMAuthorizationResponse is the response type that is received from an
// authorization provider such as identity adaptor.
type TMAuthorizationResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}

// Encode serializes OCIUser to a string
func (o *OCIUser) Encode() (string, error) {
	bs, err := json.Marshal(o)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(bs), nil
}

// Decode deserializes OCIUser from a string
func (o *OCIUser) Decode(encoded string) error {
	bdec := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(encoded))
	return json.NewDecoder(bdec).Decode(o)
}

// OCIUser describes a user or service making the request.
// This is a wrapper around ociauthz.Principal
type OCIUser struct {
	// OCID is the UserOCID identitying this user
	OCID string
	// TenancyID is the tenancy of the user
	TenancyID string
	// ContextType is "OCI"
	ContextType string
	// Claims of the principal.
	Claims ociauthz.Claims `json:"claims,omitempty"`
	// Delegate principal that made the request, if applicable.
	Delegate *OCIUser `json:"ociUser,omitempty"`
}

// ToPrincipal converts the OCIUser object to the principal used by authn/authz.
// This will never return nil for the principal.
func (o *OCIUser) ToPrincipal() *ociauthz.Principal {
	principal := ociauthz.NewPrincipal(o.OCID, o.TenancyID)
	if o.Claims != nil {
		for _, c := range o.Claims.ToSlice() {
			principal.AddClaim(c)
		}
	}

	return principal
}

func (o *OCIUser) DelegatePrincipal() *ociauthz.Principal {
	if o.Delegate == nil {
		return nil
	}

	return o.Delegate.ToPrincipal()
}
