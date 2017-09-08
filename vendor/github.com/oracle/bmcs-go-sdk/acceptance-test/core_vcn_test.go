// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,core_vcn recording,all !recording

package acceptance

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MustWin/baremetal-sdk-go/acceptance-test/helpers"
)

func TestVCNCreate(t *testing.T) {
	client := helpers.GetClient("fixtures/core/vcn")
	defer client.Stop()

	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err)

	id, err := helpers.CreateVCN(client, "172.16.0.0/16", compartmentID)
	require.NoError(t, err)
	defer func() {
		_, err = helpers.DeleteVCN(client, id)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, id)

	vcn, err := client.GetVirtualNetwork(id)

	require.NoError(t, err)
	assert.Equal(t, id, vcn.ID)
}

func TestVCNList(t *testing.T) {
	client := helpers.GetClient("fixtures/core/vcn_list")
	defer client.Stop()
	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err)

	id, err := helpers.CreateVCN(client, "172.16.0.0/16", compartmentID)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	defer func() {
		_, err = helpers.DeleteVCN(client, id)
		assert.NoError(t, err)
	}()

	vcns, err := client.ListVirtualNetworks(compartmentID, nil)

	require.NoError(t, err)
	found := false
	for _, vcn := range vcns.VirtualNetworks {
		if strings.Compare(vcn.ID, id) == 0 {
			found = true
		}
	}
	require.True(t, found, "Created VCN not found in list")
}
