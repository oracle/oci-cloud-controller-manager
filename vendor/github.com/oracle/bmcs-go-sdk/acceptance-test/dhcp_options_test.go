// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package acceptance

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oracle/bmcs-go-sdk/acceptance-test/helpers"
)

func TestDHCPOptionCRUD(t *testing.T) {
	client := helpers.GetClient("fixtures/dhcp_options")
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
	dhcpID, err := helpers.CreateDhcpOption(client, compartmentID, vcnID)
	assert.NoError(t, err, "Create")
	assert.NotEmpty(t, dhcpID, "Create: ID")

	// TODO: Get
	// TODO: Update

	// List
	opts, err := client.ListDHCPOptions(compartmentID, vcnID, nil)
	assert.NoError(t, err, "List")
	found := false
	for _, opt := range opts.DHCPOptions {
		if strings.Compare(dhcpID, opt.ID) == 0 {
			found = true
		}
	}
	assert.True(t, found, "List: Created DHCPOptions not found")

	// Delete
	_, err = helpers.DeleteDhcpOption(client, dhcpID)
	assert.NoError(t, err, "Delete")
}
