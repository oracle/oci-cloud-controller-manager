// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,load_balancer_certificate recording,all !recording

package acceptance

import (
	"fmt"
	"log"
	"testing"

	bm "github.com/MustWin/baremetal-sdk-go"
	"github.com/MustWin/baremetal-sdk-go/acceptance-test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateLoadBalancerCertificate(t *testing.T) {
	client := helpers.GetClient("fixtures/load_balancer_certificate/create")
	defer client.Stop()

	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err)

	// Get Availability Domains
	ads, err := client.ListAvailabilityDomains(compartmentID)
	require.NoError(t, err)
	availabilityDomains := ads.AvailabilityDomains

	// populate shapeName from ListShapes() {
	shapeList, err := client.ListLoadBalancerShapes(compartmentID, nil)
	require.NoError(t, err)
	shapes := shapeList.LoadBalancerShapes

	vcnID, err := helpers.CreateVCN(client, "172.16.0.0/16", compartmentID)
	require.NoError(t, err)
	require.NotEmpty(t, vcnID)

	// Load Balancers require 2 subnets, each in seperate Availability Domains
	subnetIDs := make([]string, 2)
	for i := range subnetIDs {
		subnetIDs[i], err = helpers.CreateSubnetWithCIDR(
			client,
			compartmentID,
			availabilityDomains[i].Name,
			vcnID,
			fmt.Sprintf("172.16.%d.0/24", i),
		)
		require.NoError(t, err)
	}

	helpers.Sleep(TIMEOUT) // TODO: can we verify the subnets have been created?

	// Minimal stub dependencies
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
	require.NoError(t, err)
	require.NotEmpty(t, workRequestID)
	var workRequest *bm.WorkRequest
	for {
		workRequest, err = client.GetWorkRequest(workRequestID, nil)
		assert.NoError(t, err)
		if workRequest.State == "SUCCEEDED" {
			break
		}
		helpers.Sleep(TIMEOUT) // wait until create is complete
	}
	require.NotEmpty(t, workRequest.LoadBalancerID)

	// Create SUT
	workRequestID, err = client.CreateCertificate(
		workRequest.LoadBalancerID,
		"certificate-name",
		"",
		"-----BEGIN RSA PRIVATE KEY-----\nMIIBOgIBAAJBAOUzyXPcEUkDrMGWwXreT1qM9WrdDVZCgdDePfnTwNEoh/Cp9X4L\nEvrdbd1mvAvhOuOqis/kJDfr4jo5YAsfbNUCAwEAAQJAJz8k4bfvJceBT2zXGIj0\noZa9d1z+qaSdwfwsNJkzzRyGkj/j8yv5FV7KNdSfsBbStlcuxUm4i9o5LXhIA+iQ\ngQIhAPzStAN8+Rz3dWKTjRWuCfy+Pwcmyjl3pkMPSiXzgSJlAiEA6BUZWHP0b542\nu8AizBT3b3xKr1AH2nkIx9OHq7F/QbECIHzqqpDypa8/QVuUZegpVrvvT/r7mn1s\nddS6cDtyJgLVAiEA1Z5OFQeuL2sekBRbMyP9WOW7zMBKakLL3TqL/3JCYxECIAkG\nl96uo1MjK/66X5zQXBG7F2DN2CbcYEz0r3c3vvfq\n-----END RSA PRIVATE KEY-----",
		"",
		"-----BEGIN CERTIFICATE-----\nMIIBNzCB4gIJAKtwJkxUgNpzMA0GCSqGSIb3DQEBCwUAMCMxITAfBgNVBAoTGElu\ndGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0xNzA0MTIyMTU3NTZaFw0xODA0MTIy\nMTU3NTZaMCMxITAfBgNVBAoTGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDBcMA0G\nCSqGSIb3DQEBAQUAA0sAMEgCQQDlM8lz3BFJA6zBlsF63k9ajPVq3Q1WQoHQ3j35\n08DRKIfwqfV+CxL63W3dZrwL4TrjqorP5CQ36+I6OWALH2zVAgMBAAEwDQYJKoZI\nhvcNAQELBQADQQCEjHVQJoiiVpIIvDWF+4YDRReVuwzrvq2xduWw7CIsDWlYuGZT\nQKVY6tnTy2XpoUk0fqUvMB/M2HGQ1WqZGHs6\n-----END CERTIFICATE-----",
		nil,
	)
	assert.NoError(t, err)
	require.NotEmpty(t, workRequestID)
	for {
		workRequest, err = client.GetWorkRequest(workRequestID, nil)
		assert.NoError(t, err)
		if workRequest.State == "SUCCEEDED" {
			break
		}
		helpers.Sleep(TIMEOUT)
	}

	// Get SUT
	log.Printf("Get SUT\n")
	lb, err := client.GetLoadBalancer(workRequest.LoadBalancerID, nil)
	require.NoError(t, err)

	assert.Equal(t,
		bm.Certificate{
			CertificateName: "certificate-name",
			// PrivateKey should not be included in response, only in request
			PublicCertificate: "-----BEGIN CERTIFICATE-----\nMIIBNzCB4gIJAKtwJkxUgNpzMA0GCSqGSIb3DQEBCwUAMCMxITAfBgNVBAoTGElu\ndGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0xNzA0MTIyMTU3NTZaFw0xODA0MTIy\nMTU3NTZaMCMxITAfBgNVBAoTGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDBcMA0G\nCSqGSIb3DQEBAQUAA0sAMEgCQQDlM8lz3BFJA6zBlsF63k9ajPVq3Q1WQoHQ3j35\n08DRKIfwqfV+CxL63W3dZrwL4TrjqorP5CQ36+I6OWALH2zVAgMBAAEwDQYJKoZI\nhvcNAQELBQADQQCEjHVQJoiiVpIIvDWF+4YDRReVuwzrvq2xduWw7CIsDWlYuGZT\nQKVY6tnTy2XpoUk0fqUvMB/M2HGQ1WqZGHs6\n-----END CERTIFICATE-----",
		},
		lb.Certificates["certificate-name"],
	)

	// Note: Certificates have no Update operation

	// Delete SUT
	workRequestID, err = client.DeleteLoadBalancer(lb.ID, nil)
	assert.NoError(t, err)
	require.NotEmpty(t, workRequestID)
	for {
		workRequest, err = client.GetWorkRequest(workRequestID, nil)
		assert.NoError(t, err)
		if workRequest.State == "SUCCEEDED" {
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
