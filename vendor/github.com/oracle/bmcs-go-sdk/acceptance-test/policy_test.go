// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package acceptance

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bm "github.com/oracle/bmcs-go-sdk"
	"github.com/oracle/bmcs-go-sdk/acceptance-test/helpers"
)

func TestPolicyCRUD(t *testing.T) {
	client := helpers.GetClient("fixtures/policy")
	defer client.Stop()
	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err, "Setup Compartment")
	c, err := client.GetCompartment(compartmentID)
	require.NoError(t, err, "Setup Compartment Get")
	compartmentName := c.Name
	// Create group and get its name, we need this for statement
	gid, err := helpers.CreateGroup(client)
	require.NoError(t, err, "Setup Group")
	defer func() {
		err = client.DeleteGroup(gid, nil)
		assert.NoError(t, err, "Teardown Group")
	}()
	group, err := client.GetGroup(gid)
	require.NoError(t, err, "Setup Group Get")

	// Create
	statements := []string{
		fmt.Sprintf("Allow group %s to manage users in compartment %s", group.Name, compartmentName),
		fmt.Sprintf("Allow group %s to inspect all-resources in compartment %s", group.Name, compartmentName),
	}
	p1, err := helpers.CreatePolicy(client, []string{statements[0]}, compartmentID)
	assert.NoError(t, err, "Create p1")
	p2, err := helpers.CreatePolicy(client, []string{statements[1]}, compartmentID)
	assert.NoError(t, err, "Create p2")

	// TODO: Get

	// Update
	updateOpts := &bm.UpdatePolicyOptions{
		UpdateIdentityOptions: bm.UpdateIdentityOptions{
			Description: "new desc",
		},
		Statements: statements,
	}
	policy, err := client.UpdatePolicy(p1, updateOpts)
	assert.NoError(t, err, "Update")
	assert.Len(t, policy.Statements, 2, "Update: Statements")
	assert.Equal(t, "new desc", policy.Description, "Update: Description")

	// List
	list, err := client.ListPolicies(compartmentID, nil)
	assert.NoError(t, err, "List")
	pols := map[string]bm.Policy{}
	for _, p := range list.Policies {
		pols[p.ID] = p
	}
	_, ok := pols[p1]
	assert.True(t, ok, "List: Policy p1 ok")
	_, ok = pols[p2]
	assert.True(t, ok, "List: Policy p2 ok")

	// Delete
	err = client.DeletePolicy(p1, nil)
	assert.NoError(t, err, "Delete p1")
	err = client.DeletePolicy(p2, nil)
	assert.NoError(t, err, "Delete p2")

}
