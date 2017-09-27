// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package acceptance

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oracle/bmcs-go-sdk/acceptance-test/helpers"
)

func TestCPECRUD(t *testing.T) {
	client := helpers.GetClient("fixtures/cpe")
	defer client.Stop()
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err, "Setup Compartment")

	// Create
	cpe, err := client.CreateCpe(compartmentID, "120.90.41.18", nil)
	assert.NoError(t, err, "Create")
	assert.NotEmpty(t, cpe.ID, "Create: ID")

	// TODO: Get
	// TODO: Update

	// List
	cpes, err := client.ListCpes(compartmentID, nil)
	assert.NoError(t, err, "List")
	found := false
	for _, ce := range cpes.Cpes {
		if strings.Compare(ce.ID, cpe.ID) == 0 {
			found = true
		}
	}
	assert.True(t, found, "List: Created CPE not found")

	// Delete
	err = client.DeleteCpe(cpe.ID, nil)
	assert.NoError(t, err, "Delete")
}
