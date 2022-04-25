// Copyright 2018 Oracle and/or its affiliates. All rights reserved.
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
	"fmt"
	"strings"

	"github.com/oracle/oci-go-sdk/v50/identity"
	"github.com/pkg/errors"
)

// IdentityInterface defines the interface to the OCI identity service consumed
// by the volume provisioner.
type IdentityInterface interface {
	GetAvailabilityDomainByName(ctx context.Context, compartmentID, name string) (*identity.AvailabilityDomain, error)
	ListAvailabilityDomains(ctx context.Context, compartmentID string) ([]identity.AvailabilityDomain, error)
}

func (c *client) ListAvailabilityDomains(ctx context.Context, compartmentID string) ([]identity.AvailabilityDomain, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "ListAvailabilityDomains")
	}

	resp, err := c.identity.ListAvailabilityDomains(ctx, identity.ListAvailabilityDomainsRequest{
		CompartmentId:   &compartmentID,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, listVerb, availabilityDomainResource)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return resp.Items, nil
}

func (c *client) GetAvailabilityDomainByName(ctx context.Context, compartmentID, name string) (*identity.AvailabilityDomain, error) {
	ads, err := c.ListAvailabilityDomains(ctx, compartmentID)
	if err != nil {
		return nil, err
	}
	// TODO: Add paging when implemented in oci-go-sdk.
	for _, ad := range ads {
		if strings.HasSuffix(strings.ToLower(*ad.Name), strings.ToLower(name)) {
			return &ad, nil
		}
	}
	return nil, fmt.Errorf("error looking up availability domain '%s':%v", name, ads)
}
