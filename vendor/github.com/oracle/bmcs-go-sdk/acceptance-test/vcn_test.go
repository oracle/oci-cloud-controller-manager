// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package acceptance

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oracle/bmcs-go-sdk/acceptance-test/helpers"
)

func TestVCNCRUD(t *testing.T) {
	client := helpers.GetClient("fixtures/vcn")
	defer client.Stop()
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err, "Setup Compartment")

	// Create
	id, err := helpers.CreateVCN(client, "172.16.0.0/16", compartmentID)
	assert.NoError(t, err, "Create")
	assert.NotEmpty(t, id, "Create: ID")

	// Get
	vcn, err := client.GetVirtualNetwork(id)
	assert.NoError(t, err, "Get")
	assert.Equal(t, id, vcn.ID, "Get: ID")

	// TODO: Update

	// List
	vcns, err := client.ListVirtualNetworks(compartmentID, nil)
	assert.NoError(t, err, "List")
	found := false
	for _, vcn := range vcns.VirtualNetworks {
		if strings.Compare(vcn.ID, id) == 0 {
			found = true
		}
	}
	assert.True(t, found, "List: Created VCN not found")

	// Delete
	_, err = helpers.DeleteVCN(client, id)
	assert.NoError(t, err, "Delete")
}
