// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,load_balancer recording,all !recording

package acceptance

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bm "github.com/MustWin/baremetal-sdk-go"
	"github.com/MustWin/baremetal-sdk-go/acceptance-test/helpers"
)

const TIMEOUT = (2 * time.Second)

func TestCreateLoadBalancer(t *testing.T) {
	client := helpers.GetClient("fixtures/load_balancer/create")
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
		require.NotEmpty(t, subnetIDs[i])
	}

	helpers.Sleep(TIMEOUT) // TODO: can we verify the subnets have been created?

	// Create SUT
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
	log.Printf("[DEBUG] Load Balancer Create Requested: %v", workRequestID)

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

	// Get SUT
	lb, err := client.GetLoadBalancer(workRequest.LoadBalancerID, nil)
	require.NoError(t, err)

	assert.Equal(t, compartmentID, lb.CompartmentID)
	assert.Equal(t, "my test LB", lb.DisplayName)
	assert.Equal(t, "100Mbps", lb.Shape)
	// SubnetIDs should use Set equivalance
	assert.Len(t, lb.SubnetIDs, len(subnetIDs))
	for _, subnetID := range subnetIDs {
		assert.Contains(t,
			lb.SubnetIDs,
			subnetID,
		)
	}
	// Note: Backend, Listener & Certificate operations happen in other test suites
	assert.Equal(t, map[string]bm.BackendSet{}, lb.BackendSets)
	assert.Equal(t, map[string]bm.Listener{}, lb.Listeners)
	assert.Equal(t, map[string]bm.Certificate{}, lb.Certificates)
	// Computed
	assert.NotEmpty(t, lb.ID)
	assert.NotEmpty(t, lb.IPAddresses)
	assert.NotEmpty(t, lb.TimeCreated)

	// TODO: Update SUT

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

	// lb, err = client.GetLoadBalancer()

	// VCN requires all subnets to be deleted first
	for _, subnetID := range subnetIDs {
		helpers.DeleteSubnet(client, subnetID)
	}
	helpers.DeleteVCN(client, vcnID)
}
