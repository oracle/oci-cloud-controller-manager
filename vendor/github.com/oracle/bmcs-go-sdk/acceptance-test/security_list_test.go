// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package acceptance

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oracle/bmcs-go-sdk/acceptance-test/helpers"
)

func TestSecurityListCRUD(t *testing.T) {
	client := helpers.GetClient("fixtures/security_list")
	defer client.Stop()
	// get a compartment, any compartment
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
	slID, err := helpers.CreateSecurityList(client, compartmentID, vcnID)
	assert.NoError(t, err, "Create")
	assert.NotEmpty(t, slID, "Create: ID")

	// List
	sls, err := client.ListSecurityLists(compartmentID, vcnID, nil)
	assert.NoError(t, err, "List")
	found := false
	for _, sl := range sls.SecurityLists {
		if strings.Compare(slID, sl.ID) == 0 {
			found = true
		}
	}
	assert.True(t, found, "List: Created SecurityList not found")

	// Delete
	_, err = helpers.DeleteSecurityList(client, slID)
	assert.NoError(t, err, "Delete")
}
