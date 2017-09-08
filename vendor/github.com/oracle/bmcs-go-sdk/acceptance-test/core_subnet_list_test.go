// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,core recording,all !recording

package acceptance

import (
	"strings"
	"testing"

	"github.com/MustWin/baremetal-sdk-go/acceptance-test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityListCreate(t *testing.T) {
	client := helpers.GetClient("fixtures/core/security_list")
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

	slID, err := helpers.CreateSecurityList(client, compartmentID, vcnID)
	require.NoError(t, err)
	defer func() {
		_, err := helpers.DeleteSecurityList(client, slID)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, slID)
}

func TestSecurityListList(t *testing.T) {
	client := helpers.GetClient("fixtures/core/security_list_list")
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

	slID, err := helpers.CreateSecurityList(client, compartmentID, vcnID)
	require.NoError(t, err)
	defer func() {
		_, err := helpers.DeleteSecurityList(client, slID)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, slID)

	sls, err := client.ListSecurityLists(compartmentID, vcnID, nil)
	require.NoError(t, err)
	found := false
	for _, sl := range sls.SecurityLists {
		if strings.Compare(slID, sl.ID) == 0 {
			found = true
		}
	}
	require.True(t, found, "Created security list not found in list")
}
