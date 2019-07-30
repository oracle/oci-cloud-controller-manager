// Copyright 2019 Oracle and/or its affiliates. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"context"
	identitymeta "github.com/oracle/oci-cloud-controller-manager/pkg/oci/identity"
	"github.com/oracle/oci-go-sdk/identity"
	"github.com/pkg/errors"
)

// IdentityMetadataSvcInterface defines the interface to the OCI identity metadata service consumed
// by the volume driver.
type IdentityMetadataSvcInterface interface {
	GetTenantByCompartment(ctx context.Context, compartmentID string) (*identity.Tenancy, error)
}

func (c *client) GetTenantByCompartment(ctx context.Context, compartmentID string) (*identity.Tenancy, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetTenantByCompartment")
	}
	resp, err := c.metadata.GetTenantByCompartment(ctx, identitymeta.GetTenantByCompartmentRequest{CompartmentId: &compartmentID})
	incRequestCounter(err, getVerb, identityMetadataResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.Tenancy, nil
}
