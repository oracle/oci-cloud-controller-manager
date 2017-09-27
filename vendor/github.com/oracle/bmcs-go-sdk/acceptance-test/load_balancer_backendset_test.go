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

func TestBackendSetCRUD(t *testing.T) {
	// TODO: this requires so much arrangement, can we simplify?
	client := helpers.GetClient("fixtures/backend_set")
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

	assert.Equal(t, compartmentID, lb.CompartmentID, "Get: CompartmentID")
	assert.Equal(t, "my test LB", lb.DisplayName, "Get: DisplayName")
	assert.Equal(t, "100Mbps", lb.Shape, "Get: Shape")
	assert.Len(t, lb.SubnetIDs, len(subnetIDs), "Get: SubnetIDs")
	// SubnetIDs should use Set equivalance
	for _, subnetID := range subnetIDs {
		assert.Contains(t,
			lb.SubnetIDs,
			subnetID,
			"Get: SubnetIDs",
		)
	}
	assert.Equal(t,
		bm.BackendSet{
			Backends: []bm.Backend{},
			Policy:   "ROUND_ROBIN",
			Name:     "backend-set-name",
			HealthChecker: &bm.HealthChecker{
				Protocol:          "HTTP",
				IntervalInMS:      10000,
				Port:              0,
				ResponseBodyRegex: ".*",
				URLPath:           "/",
				Retries:           3,
				ReturnCode:        200,
				TimeoutInMS:       3000,
			},
		},
		lb.BackendSets["backend-set-name"],
		"Get: BackendSet",
	)
	// Note: Listener & Certificate operations happen in other test suites
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
