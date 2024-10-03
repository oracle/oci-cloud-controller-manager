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

	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/pkg/errors"
)

// IdentityInterface defines the interface to the OCI identity service consumed
// by the volume provisioner.
type IdentityInterface interface {
	GetAvailabilityDomainByName(ctx context.Context, compartmentID, name string) (*identity.AvailabilityDomain, error)
}

func (c *client) GetAvailabilityDomainByName(ctx context.Context, compartmentID, name string) (*identity.AvailabilityDomain, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "ListAvailabilityDomains")
	}

	var availabilityDomains []identity.AvailabilityDomain
	var err error

	// TODO: Uncomment when compartments is available in OCI Go-SDK
	//if IsIpv6SingleStackCluster() {
	//	availabilityDomains, err = c.listAvailabilityDomainsV6(ctx, compartmentID)
	//} else {
	availabilityDomains, err = c.listAvailabilityDomains(ctx, compartmentID)
	//}

	if err != nil {
		return nil, err
	}

	// Find the desired availability domain by name
	for _, ad := range availabilityDomains {
		if strings.HasSuffix(strings.ToLower(*ad.Name), strings.ToLower(name)) {
			return &ad, nil
		}
	}
	return nil, fmt.Errorf("availability domain '%s' not found in list: %v", name, availabilityDomains)
}

// listAvailabilityDomainsV6 lists availability domains for IPv6 single-stack clusters.
//func (c *client) listAvailabilityDomainsV6(ctx context.Context, compartmentID string) ([]identity.AvailabilityDomain, error) {
//	resp, err := c.compartment.ListAvailabilityDomains(ctx, compartments.ListAvailabilityDomainsRequest{
//		CompartmentId:   &compartmentID,
//		RequestMetadata: c.requestMetadata,
//	})
//	if resp.OpcRequestId != nil {
//		c.logger.With("service", "Compartment", "verb", listVerb).
//			With("OpcRequestId", *(resp.OpcRequestId)).With("statusCode", util.GetHttpStatusCode(err)).
//			Info("OPC Request ID recorded for Compartment ListAvailabilityDomains call.")
//	}
//	incRequestCounter(err, listVerb, availabilityDomainResource)
//	if err != nil {
//		return nil, errors.WithStack(err)
//	}
//
//	availabilityDomains := make([]identity.AvailabilityDomain, len(resp.Items))
//	for i, ad := range resp.Items {
//		availabilityDomains[i] = identity.AvailabilityDomain{Name: ad.Name}
//	}
//	return availabilityDomains, nil
//}

// listAvailabilityDomains lists availability domains for regular clusters.
func (c *client) listAvailabilityDomains(ctx context.Context, compartmentID string) ([]identity.AvailabilityDomain, error) {
	resp, err := c.identity.ListAvailabilityDomains(ctx, identity.ListAvailabilityDomainsRequest{
		CompartmentId:   &compartmentID,
		RequestMetadata: c.requestMetadata,
	})
	if resp.OpcRequestId != nil {
		c.logger.With("service", "Identity", "verb", listVerb).
			With("OpcRequestId", *(resp.OpcRequestId)).With("statusCode", util.GetHttpStatusCode(err)).
			Info("OPC Request ID recorded for ListAvailabilityDomains call.")
	}
	incRequestCounter(err, listVerb, availabilityDomainResource)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return resp.Items, nil
}
