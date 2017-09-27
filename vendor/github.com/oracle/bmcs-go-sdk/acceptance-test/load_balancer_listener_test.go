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

func TestListenerCRUD(t *testing.T) {
	// TODO: this requires so much arrangement, can we simplify?
	client := helpers.GetClient("fixtures/listener")
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
	// Create load balancer backend set
	workRequestID, err = client.CreateBackendSet(
		workRequest.LoadBalancerID,
		"backend-set-name",
		"ROUND_ROBIN",
		[]bm.Backend{},
		&bm.HealthChecker{
			Protocol: "HTTP",
			URLPath:  "/",
		},
		nil, // &bm.SSLConfiguration{},
		nil,
		nil,
	)
	require.NoError(t, err, "Setup LoadBalancer BackendSet")
	require.NotEmpty(t, workRequestID, "Setup LoadBalancer BackendSet: WorkRequest ID")
	for {
		workRequest, err = client.GetWorkRequest(workRequestID, nil)
		require.NoError(t, err, "Setup LoadBalancer BackendSet: WorkRequest Get")
		if workRequest.State == bm.WorkRequestSucceeded {
			break
		}
		helpers.Sleep(TIMEOUT)
	}
	protos, err := client.ListLoadBalancerProtocols(
		compartmentID,
		nil,
	)
	require.NoError(t, err, "Setup LoadBalancer Protocols")

	// Create
	workRequestID, err = client.CreateListener(
		workRequest.LoadBalancerID,
		"listener-name",
		"backend-set-name",
		protos.LoadBalancerProtocols[0].Name,
		1234,
		nil,
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
	assert.NoError(t, err, "Get")

	assert.Equal(t,
		bm.Listener{
			DefaultBackendSetName: "backend-set-name",
			Name:      "listener-name",
			Port:      1234,
			Protocol:  "HTTP",
			SSLConfig: (*bm.SSLConfiguration)(nil),
		},
		lb.Listeners["listener-name"],
		"Get",
	)

	// TODO: Update SUT

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
