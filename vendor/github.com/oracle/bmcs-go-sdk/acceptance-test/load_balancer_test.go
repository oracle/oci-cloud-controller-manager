// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package acceptance

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bm "github.com/oracle/bmcs-go-sdk"
	"github.com/oracle/bmcs-go-sdk/acceptance-test/helpers"
)

const TIMEOUT = (2 * time.Second)

func TestLoadBalancerCRUD(t *testing.T) {
	client := helpers.GetClient("fixtures/load_balancer")
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

	// Create
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
	assert.NoError(t, err, "Create")
	assert.NotEmpty(t, workRequestID, "Create: WorkRequest ID")
	log.Printf("[DEBUG] Load Balancer Create Requested: %v", workRequestID)
	var workRequest *bm.WorkRequest
	for {
		workRequest, err = client.GetWorkRequest(workRequestID, nil)
		assert.NoError(t, err, "Create: WorkRequest Get")
		if workRequest.State == bm.WorkRequestSucceeded {
			break
		}
		helpers.Sleep(TIMEOUT) // wait until create is complete
	}
	assert.NotEmpty(t, workRequest.LoadBalancerID, "Create: ID")

	// Get
	lb, err := client.GetLoadBalancer(workRequest.LoadBalancerID, nil)
	assert.NoError(t, err, "Get")
	assert.Equal(t, compartmentID, lb.CompartmentID, "Get: CompartmentID")
	assert.Equal(t, "my test LB", lb.DisplayName, "Get: DisplayName")
	assert.Equal(t, "100Mbps", lb.Shape, "Get: Shape")
	// SubnetIDs should use Set equivalance
	assert.Len(t, lb.SubnetIDs, len(subnetIDs), "Get: SubnetIDs")
	for _, subnetID := range subnetIDs {
		assert.Contains(t,
			lb.SubnetIDs,
			subnetID,
			"Get",
		)
	}
	// Note: Backend, Listener & Certificate operations happen in other tests
	assert.Equal(t, map[string]bm.BackendSet{}, lb.BackendSets, "Get: BackendSets")
	assert.Equal(t, map[string]bm.Listener{}, lb.Listeners, "Get: Listeners")
	assert.Equal(t, map[string]bm.Certificate{}, lb.Certificates, "Get: Certificates")
	// Computed
	assert.NotEmpty(t, lb.ID, "Get: ID")
	assert.NotEmpty(t, lb.IPAddresses, "Get: IPAddresses")
	assert.NotEmpty(t, lb.TimeCreated, "Get: TimeCreated")

	// TODO: Update

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
