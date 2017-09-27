// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package acceptance

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oracle/bmcs-go-sdk/acceptance-test/helpers"
)

func TestRouteTableCRUD(t *testing.T) {
	client := helpers.GetClient("fixtures/route_table")
	defer client.Stop()
	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err, "Setup Compartment")
	// Create VCN
	vcnID, err := helpers.CreateVCN(client, "172.16.0.0/16", compartmentID)
	require.NoError(t, err, "Setup VCN")
	require.NotEmpty(t, vcnID, "Setup VCN: ID")
	defer func() {
		_, err = helpers.DeleteVCN(client, vcnID)
		assert.NoError(t, err, "Teardown VCN")
	}()
	// Create internet gateway
	igID, err := helpers.CreateInternetGateway(client, compartmentID, vcnID)
	require.NoError(t, err, "Setup InternetGateway")
	require.NotEmpty(t, igID, "Setup InternetGateway: ID")
	defer func() {
		_, err := helpers.DeleteInternetGateway(client, igID)
		assert.NoError(t, err, "Teardown InternetGateway")
	}()

	// Create
	rtID, err := helpers.CreateRouteTable(client, compartmentID, vcnID, igID)
	assert.NoError(t, err, "Create")
	assert.NotEmpty(t, rtID, "Create: ID")

	// List
	rts, err := client.ListRouteTables(compartmentID, vcnID, nil)
	assert.NoError(t, err, "List")
	found := false
	for _, rt := range rts.RouteTables {
		if strings.Compare(rtID, rt.ID) == 0 {
			found = true
		}
	}
	assert.True(t, found, "List: Created RouteTable not found")

	// Delete
	_, err = helpers.DeleteRouteTable(client, rtID)
	assert.NoError(t, err, "Delete")
}
