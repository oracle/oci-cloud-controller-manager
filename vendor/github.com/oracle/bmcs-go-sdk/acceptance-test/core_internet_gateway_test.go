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

func TestInternetGatewayCreate(t *testing.T) {
	client := helpers.GetClient("fixtures/core/internet_gateway")
	defer client.Stop()
	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err)

	vcnID, err := helpers.CreateVCN(client, "172.16.0.0/16", compartmentID)
	require.NoError(t, err)
	defer func() {
		_, err = helpers.DeleteVCN(client, vcnID)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, vcnID)

	igID, err := helpers.CreateInternetGateway(client, compartmentID, vcnID)
	require.NoError(t, err)
	defer func() {
		_, err := helpers.DeleteInternetGateway(client, igID)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, igID)
}

func TestInternetGatewayList(t *testing.T) {
	client := helpers.GetClient("fixtures/core/internet_gateway_list")
	defer client.Stop()
	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err)

	vcnID, err := helpers.CreateVCN(client, "172.16.0.0/16", compartmentID)
	require.NoError(t, err)
	defer func() {
		_, err = helpers.DeleteVCN(client, vcnID)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, vcnID)

	igID, err := helpers.CreateInternetGateway(client, compartmentID, vcnID)
	require.NoError(t, err)
	defer func() {
		_, err := helpers.DeleteInternetGateway(client, igID)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, igID)

	igs, err := client.ListInternetGateways(compartmentID, vcnID, nil)
	require.NoError(t, err)
	found := false
	for _, ig := range igs.Gateways {
		if strings.Compare(igID, ig.ID) == 0 {
			found = true
		}
	}
	require.True(t, found, "Created Internet Gateway not found in list")
}
