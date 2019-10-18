package types

import (
	"errors"
	"fmt"
)

const (
	BMCCloudType = "bmc"
	awsCloudType = "aws"

	CloudAuthRoleClusterEnv   = "clusterEnv"
	CloudAuthRoleCustomerAide = "customerAide"
)

// CloudAuth represents the auth details fora specific cloud provider.  If
// providers are modified, `ValidateCloudAuthForCreation` and
// CloudAuthV1.Sanitize() should be modified too.
type CloudAuth struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	AWSAuth   *AWSCloudAuth `json:"aws,omitempty"`
	BMCAuth   *BMCCloudAuth `json:"bmc,omitempty"`
	CloudType string        `json:"cloudType,omitempty"`
	Role      string        `json:"role,omitempty"`
}

// FIXME: TenancyID is not the same as TenantID, but cloud auth is being deprecrated
func (ca *CloudAuth) TenancyID() string {
	if ca.BMCAuth != nil {
		return ca.BMCAuth.Tenancy
	}
	return ""
}

func (ca *CloudAuth) ToBMConfig() (*BMConfig, error) {
	if ca == nil {
		return nil, fmt.Errorf("invalid empty cloud auth")
	}
	if ca.BMCAuth == nil {
		return nil, fmt.Errorf("invalid empty bmc auth in cloud auth")
	}

	return &BMConfig{
		UserOCID:    ca.BMCAuth.User,
		Fingerprint: ca.BMCAuth.Fingerprint,
		PrivateKey:  []byte(ca.BMCAuth.PrivateKey),
		Tenancy:     ca.BMCAuth.Tenancy,
		Region:      ca.BMCAuth.Region,
	}, nil
}

type CloudAuthV1 struct {
	CloudAuth
	ResourceOwnerID string `json:"resourceOwnerId"`
}

type AWSCloudAuth struct {
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
}

type BMCCloudAuth struct {
	User        string `json:"user"`
	Fingerprint string `json:"fingerprint"`
	PrivateKey  string `json:"privateKey"`
	Tenancy     string `json:"tenancy"`
	Region      string `json:"region"`
}

// ToV1 converts a CloudAuth to a response that can be marshaled to a consumer.
// Auth details are removed and replaced with an identifier in the CloudType
// attribute instead.
func (c *CloudAuth) ToV1() CloudAuthV1 {
	var v1 CloudAuthV1
	if c != nil {
		v1.ID = c.ID
		v1.Name = c.Name
	}
	if c.BMCAuth != nil {
		v1.CloudType = BMCCloudType
	} else if c.AWSAuth != nil {
		v1.CloudType = awsCloudType
	}
	return v1
}

// ToProto converts a CloudAuthV1 to a CloudAuth.  A CloudAuth is not actually
// a protobuf, but the method name here is used for consistency.
func (v1 *CloudAuthV1) ToProto() *CloudAuth {
	c := new(CloudAuth)
	if v1 != nil {
		c.ID = v1.ID
		c.Name = v1.Name
		if v1.AWSAuth != nil {
			c.AWSAuth = v1.AWSAuth
		}
		if v1.BMCAuth != nil {
			c.BMCAuth = v1.BMCAuth
		}
	}
	return c
}

func (v1 *CloudAuthV1) SetResourceOwnerID(id string) {
	v1.ResourceOwnerID = id
}

type CloudAuthListV1 struct {
	AuthConfigs []*CloudAuthV1 `json:"authConfigs"`
}

func (c *CloudAuthListV1) SetResourceOwnerID(id string) {
	for _, auth := range c.AuthConfigs {
		auth.ResourceOwnerID = id
	}
}

type CloudAuthList struct {
	AuthConfigs []*CloudAuth
}

// TODO: remove CloudAuthList
func (c *CloudAuthList) Filter(tenancyID string) {
	filteredAuths := make([]*CloudAuth, 0, len(c.AuthConfigs))
	for _, auth := range c.AuthConfigs {
		if auth.BMCAuth != nil {
			if auth.BMCAuth.Tenancy == tenancyID {
				filteredAuths = append(filteredAuths, auth)
			}
		}
	}
	c.AuthConfigs = filteredAuths
}

func (c *CloudAuthList) ToV1() CloudAuthListV1 {
	var v1 CloudAuthListV1
	if c != nil {
		if c.AuthConfigs != nil {
			v1.AuthConfigs = []*CloudAuthV1{}
			for _, auth := range c.AuthConfigs {
				v1Auth := auth.ToV1()
				v1.AuthConfigs = append(v1.AuthConfigs, &v1Auth)
			}
		}
	}
	return v1
}

type CloudAuthListOptions struct {
	RolesInclusions []string
}

func ValidateCloudAuthForCreation(auth *CloudAuth) error {
	if auth == nil {
		return errors.New("nil cloud auth")
	}
	if auth.Name == "" {
		return errors.New("cloud auth missing name")
	}
	if (auth.AWSAuth != nil) == (auth.BMCAuth != nil) {
		return errors.New("cloud auth must have exactly one provider")
	}
	return nil
}
