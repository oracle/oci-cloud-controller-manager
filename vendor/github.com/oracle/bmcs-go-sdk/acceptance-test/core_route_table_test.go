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

func TestRouteTableCreate(t *testing.T) {
	client := helpers.GetClient("fixtures/core/route_table")
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

	rtID, err := helpers.CreateRouteTable(client, compartmentID, vcnID, igID)
	require.NoError(t, err)
	defer func() {
		_, err := helpers.DeleteRouteTable(client, rtID)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, rtID)
}

func TestRouteTableList(t *testing.T) {
	client := helpers.GetClient("fixtures/core/route_table_list")
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

	rtID, err := helpers.CreateRouteTable(client, compartmentID, vcnID, igID)
	require.NoError(t, err)
	defer func() {
		_, err := helpers.DeleteRouteTable(client, rtID)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, rtID)

	rts, err := client.ListRouteTables(compartmentID, vcnID, nil)
	require.NoError(t, err)
	found := false
	for _, rt := range rts.RouteTables {
		if strings.Compare(rtID, rt.ID) == 0 {
			found = true
		}
	}
	require.True(t, found, "Created route table not found in list")
}
