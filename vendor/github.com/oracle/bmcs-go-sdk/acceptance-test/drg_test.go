// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package acceptance

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oracle/bmcs-go-sdk/acceptance-test/helpers"
)

func TestDRGDRUD(t *testing.T) {
	client := helpers.GetClient("fixtures/drg")
	defer client.Stop()
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err, "Setup Compartment")

	// Create
	drgID, err := helpers.CreateDrg(client, compartmentID)
	assert.NoError(t, err, "Create")
	assert.NotEmpty(t, drgID, "Create: ID")

	// TODO: Get
	// TODO: Update

	// List
	drgs, err := client.ListDrgs(compartmentID, nil)
	assert.NoError(t, err, "List")
	found := false
	for _, drg := range drgs.Drgs {
		if strings.Compare(drgID, drg.ID) == 0 {
			found = true
		}
	}
	assert.True(t, found, "List: Created DRG not found")

	// Delete
	_, err = helpers.DeleteDrg(client, drgID)
	assert.NoError(t, err, "Delete")
}
