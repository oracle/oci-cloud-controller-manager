// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,core recording,all !recording

package acceptance

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MustWin/baremetal-sdk-go/acceptance-test/helpers"
)

func TestDhcpOptionsCreate(t *testing.T) {
	client := helpers.GetClient("fixtures/core/dhcp_options")
	defer client.Stop()

	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err)

	vcnID, err := helpers.CreateVCN(client, "172.16.0.0/16", compartmentID)
	require.NoError(t, err)
	defer func() {
		_, err = helpers.DeleteVCN(client, vcnID)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, vcnID)

	dhcpID, err := helpers.CreateDhcpOption(client, compartmentID, vcnID)
	require.NoError(t, err)
	defer func() {
		_, err = helpers.DeleteDhcpOption(client, dhcpID)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, dhcpID)

}

func TestDhcpOptionsList(t *testing.T) {
	client := helpers.GetClient("fixtures/core/dhcp_options_list")
	defer client.Stop()

	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err)

	vcnID, err := helpers.CreateVCN(client, "172.16.0.0/16", compartmentID)
	require.NoError(t, err)
	defer func() {
		_, err = helpers.DeleteVCN(client, vcnID)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, vcnID)

	dhcpID, err := helpers.CreateDhcpOption(client, compartmentID, vcnID)
	require.NoError(t, err)
	defer func() {
		_, err = helpers.DeleteDhcpOption(client, dhcpID)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, dhcpID)

	opts, err := client.ListDHCPOptions(compartmentID, vcnID, nil)
	require.NoError(t, err)
	found := false
	for _, opt := range opts.DHCPOptions {
		if strings.Compare(dhcpID, opt.ID) == 0 {
			found = true
		}
	}
	require.True(t, found, "Created options not found in list")
}
