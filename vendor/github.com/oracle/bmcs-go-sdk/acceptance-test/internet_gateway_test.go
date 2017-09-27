// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package acceptance

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oracle/bmcs-go-sdk/acceptance-test/helpers"
)

func TestInternetGatewayCRUD(t *testing.T) {
	client := helpers.GetClient("fixtures/internet_gateway")
	defer client.Stop()
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err, "Setup Compartment")
	vcnID, err := helpers.CreateVCN(client, "172.16.0.0/16", compartmentID)
	require.NoError(t, err, "Setup VCN")
	require.NotEmpty(t, vcnID, "Setup VCN: ID")
	defer func() {
		_, err = helpers.DeleteVCN(client, vcnID)
		assert.NoError(t, err, "Teardown VCN")
	}()

	// Create
	igID, err := helpers.CreateInternetGateway(client, compartmentID, vcnID)
	assert.NoError(t, err, "Create")
	assert.NotEmpty(t, igID, "Create: ID")

	// TODO: Get
	// TODO: Update

	// List
	igs, err := client.ListInternetGateways(compartmentID, vcnID, nil)
	assert.NoError(t, err, "List")
	found := false
	for _, ig := range igs.Gateways {
		if strings.Compare(igID, ig.ID) == 0 {
			found = true
		}
	}
	assert.True(t, found, "List: Created InternetGateway not found")

	// Delete
	_, err = helpers.DeleteInternetGateway(client, igID)
	assert.NoError(t, err, "Delete")
}
