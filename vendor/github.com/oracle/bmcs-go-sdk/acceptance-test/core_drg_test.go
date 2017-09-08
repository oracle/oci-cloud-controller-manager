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

func TestDrgCreate(t *testing.T) {
	client := helpers.GetClient("fixtures/core/drg")
	defer client.Stop()

	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err)

	drgID, err := helpers.CreateDrg(client, compartmentID)
	require.NoError(t, err)
	defer func() {
		_, err = helpers.DeleteDrg(client, drgID)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, drgID)

}

func TestDrgList(t *testing.T) {
	client := helpers.GetClient("fixtures/core/drg_list")
	defer client.Stop()

	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err)

	drgID, err := helpers.CreateDrg(client, compartmentID)
	require.NoError(t, err)
	defer func() {
		_, err = helpers.DeleteDrg(client, drgID)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, drgID)

	drgs, err := client.ListDrgs(compartmentID, nil)
	require.NoError(t, err)
	found := false
	for _, drg := range drgs.Drgs {
		if strings.Compare(drgID, drg.ID) == 0 {
			found = true
		}
	}
	require.True(t, found, "Created Drg not found in list")
}
