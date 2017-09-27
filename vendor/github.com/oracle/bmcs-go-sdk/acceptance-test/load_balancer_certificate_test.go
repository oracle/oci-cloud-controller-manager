// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package acceptance

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bm "github.com/oracle/bmcs-go-sdk"
	"github.com/oracle/bmcs-go-sdk/acceptance-test/helpers"
)

func TestCertificateCRUD(t *testing.T) {
	// TODO: this requires so much arrangement, can we simplify?
	client := helpers.GetClient("fixtures/certificate")
	defer client.Stop()
	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err, "Setup Compartment")
	// Get Availability Domains
	ads, err := client.ListAvailabilityDomains(compartmentID)
	require.NoError(t, err, "Setup AvailabilityDomains")
	availabilityDomains := ads.AvailabilityDomains
	// populate shapeName from ListShapes()
	shapeList, err := client.ListLoadBalancerShapes(compartmentID, nil)
	require.NoError(t, err, "Setup LoadBalancer Shapes")
	shapes := shapeList.LoadBalancerShapes
	// Create VCN
	vcnID, err := helpers.CreateVCN(client, "172.16.0.0/16", compartmentID)
	require.NoError(t, err, "Setup VCN")
	require.NotEmpty(t, vcnID, "Setup VCN: ID")
	// Create subnets. Load Balancers require 2, each in seperate Availability Domains
	subnetIDs := make([]string, 2)
	for i := range subnetIDs {
		subnetIDs[i], err = helpers.CreateSubnetWithOptions(
			client,
			compartmentID,
			availabilityDomains[i].Name,
			vcnID,
			fmt.Sprintf("172.16.%d.0/24", i),
			nil,
		)
		require.NoError(t, err, "Setup Subnet %v", i)
	}
	helpers.Sleep(TIMEOUT) // TODO: can we verify the subnets have been created?
	// Create minimal load balander
	workRequestID, err := client.CreateLoadBalancer(
		nil,
		nil,
		compartmentID,
		nil,
		shapes[0].Name,
		subnetIDs,
		&bm.CreateLoadBalancerOptions{
			DisplayNameOptions: bm.DisplayNameOptions{
				DisplayName: "my test LB",
			},
		},
	)
	require.NoError(t, err, "Setup LoadBalancer")
	require.NotEmpty(t, workRequestID, "Setup LoadBalancer: WorkRequest ID")
	var workRequest *bm.WorkRequest
	for {
		workRequest, err = client.GetWorkRequest(workRequestID, nil)
		require.NoError(t, err, "Setup LoadBalancer: WorkRequest Get")
		if workRequest.State == bm.WorkRequestSucceeded {
			break
		}
		helpers.Sleep(TIMEOUT) // wait until create is complete
	}
	require.NotEmpty(t, workRequest.LoadBalancerID, "Setup LoadBalancer: ID")

	// Create
	workRequestID, err = client.CreateCertificate(
		workRequest.LoadBalancerID,
		"certificate-name",
		"",
		"-----BEGIN RSA PRIVATE KEY-----\nMIIBOgIBAAJBAOUzyXPcEUkDrMGWwXreT1qM9WrdDVZCgdDePfnTwNEoh/Cp9X4L\nEvrdbd1mvAvhOuOqis/kJDfr4jo5YAsfbNUCAwEAAQJAJz8k4bfvJceBT2zXGIj0\noZa9d1z+qaSdwfwsNJkzzRyGkj/j8yv5FV7KNdSfsBbStlcuxUm4i9o5LXhIA+iQ\ngQIhAPzStAN8+Rz3dWKTjRWuCfy+Pwcmyjl3pkMPSiXzgSJlAiEA6BUZWHP0b542\nu8AizBT3b3xKr1AH2nkIx9OHq7F/QbECIHzqqpDypa8/QVuUZegpVrvvT/r7mn1s\nddS6cDtyJgLVAiEA1Z5OFQeuL2sekBRbMyP9WOW7zMBKakLL3TqL/3JCYxECIAkG\nl96uo1MjK/66X5zQXBG7F2DN2CbcYEz0r3c3vvfq\n-----END RSA PRIVATE KEY-----",
		"",
		"-----BEGIN CERTIFICATE-----\nMIIBNzCB4gIJAKtwJkxUgNpzMA0GCSqGSIb3DQEBCwUAMCMxITAfBgNVBAoTGElu\ndGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0xNzA0MTIyMTU3NTZaFw0xODA0MTIy\nMTU3NTZaMCMxITAfBgNVBAoTGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDBcMA0G\nCSqGSIb3DQEBAQUAA0sAMEgCQQDlM8lz3BFJA6zBlsF63k9ajPVq3Q1WQoHQ3j35\n08DRKIfwqfV+CxL63W3dZrwL4TrjqorP5CQ36+I6OWALH2zVAgMBAAEwDQYJKoZI\nhvcNAQELBQADQQCEjHVQJoiiVpIIvDWF+4YDRReVuwzrvq2xduWw7CIsDWlYuGZT\nQKVY6tnTy2XpoUk0fqUvMB/M2HGQ1WqZGHs6\n-----END CERTIFICATE-----",
		nil,
	)
	assert.NoError(t, err, "Create")
	assert.NotEmpty(t, workRequestID, "Create: WorkRequest ID")
	for {
		workRequest, err = client.GetWorkRequest(workRequestID, nil)
		assert.NoError(t, err, "Create: WorkRequest Get")
		if workRequest.State == bm.WorkRequestSucceeded {
			break
		}
		helpers.Sleep(TIMEOUT)
	}

	// Get
	lb, err := client.GetLoadBalancer(workRequest.LoadBalancerID, nil)
	assert.NoError(t, err, "Create")

	assert.Equal(t,
		bm.Certificate{
			CertificateName: "certificate-name",
			// PrivateKey should not be included in response, only in request
			PublicCertificate: "-----BEGIN CERTIFICATE-----\nMIIBNzCB4gIJAKtwJkxUgNpzMA0GCSqGSIb3DQEBCwUAMCMxITAfBgNVBAoTGElu\ndGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0xNzA0MTIyMTU3NTZaFw0xODA0MTIy\nMTU3NTZaMCMxITAfBgNVBAoTGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDBcMA0G\nCSqGSIb3DQEBAQUAA0sAMEgCQQDlM8lz3BFJA6zBlsF63k9ajPVq3Q1WQoHQ3j35\n08DRKIfwqfV+CxL63W3dZrwL4TrjqorP5CQ36+I6OWALH2zVAgMBAAEwDQYJKoZI\nhvcNAQELBQADQQCEjHVQJoiiVpIIvDWF+4YDRReVuwzrvq2xduWw7CIsDWlYuGZT\nQKVY6tnTy2XpoUk0fqUvMB/M2HGQ1WqZGHs6\n-----END CERTIFICATE-----",
		},
		lb.Certificates["certificate-name"],
		"Create: Certificate",
	)

	// Note: Certificates have no Update operation

	// Delete
	workRequestID, err = client.DeleteLoadBalancer(lb.ID, nil)
	assert.NoError(t, err, "Delete")
	assert.NotEmpty(t, workRequestID, "Delete: WorkRequest ID")
	for {
		workRequest, err = client.GetWorkRequest(workRequestID, nil)
		assert.NoError(t, err, "Delete: WorkRequest Get")
		if workRequest.State == bm.WorkRequestSucceeded {
			break
		}
		helpers.Sleep(TIMEOUT) // wait until delete is complete
	}

	// VCN requires all subnets to be deleted first
	for _, subnetID := range subnetIDs {
		helpers.DeleteSubnet(client, subnetID)
	}
	helpers.DeleteVCN(client, vcnID)
}
