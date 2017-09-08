// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,identity_policy recording,all !recording

package acceptance

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bm "github.com/MustWin/baremetal-sdk-go"
	"github.com/MustWin/baremetal-sdk-go/acceptance-test/helpers"
)

func TestPolicyLifecycle(t *testing.T) {
	client := helpers.GetClient("fixtures/identity/policy")
	defer client.Stop()
	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err)
	c, err := client.GetCompartment(compartmentID)
	require.NoError(t, err)
	compartmentName := c.Name
	// create group and get its name, we need this for statement
	gid, err := helpers.CreateGroup(client)
	require.NoError(t, err)
	defer func() {
		err = client.DeleteGroup(gid, nil)
		assert.NoError(t, err)
	}()
	group, err := client.GetGroup(gid)
	require.NoError(t, err)
	statements := []string{
		fmt.Sprintf("Allow group %s to manage users in compartment %s", group.Name, compartmentName),
	}
	pid, err := helpers.CreatePolicy(client, statements, compartmentID)
	require.NoError(t, err)
	defer func() {
		err = client.DeletePolicy(pid, nil)
		assert.NoError(t, err)
	}()

	policy, err := client.GetPolicy(pid)
	assert.NoError(t, err)
	assert.Len(t, policy.Statements, 1)
}

func TestPolicyUpdate(t *testing.T) {
	client := helpers.GetClient("fixtures/identity/updatepolicy")
	defer client.Stop()
	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err)
	c, err := client.GetCompartment(compartmentID)
	require.NoError(t, err)
	compartmentName := c.Name
	// create group and get its name, we need this for statement
	gid, err := helpers.CreateGroup(client)
	require.NoError(t, err)
	defer func() {
		err = client.DeleteGroup(gid, nil)
		assert.NoError(t, err)
	}()
	group, err := client.GetGroup(gid)
	require.NoError(t, err)
	statements := []string{
		fmt.Sprintf("Allow group %s to manage users in compartment %s", group.Name, compartmentName),
	}
	pid, err := helpers.CreatePolicy(client, statements, compartmentID)
	require.NoError(t, err)
	defer func() {
		err = client.DeletePolicy(pid, nil)
		assert.NoError(t, err)
	}()

	statements = append(statements, fmt.Sprintf("Allow group %s to inspect all-resources in compartment %s", group.Name, compartmentName))
	updateOpts := &bm.UpdatePolicyOptions{
		UpdateIdentityOptions: bm.UpdateIdentityOptions{
			Description: "new desc",
		},
		Statements: statements,
	}

	policy, err := client.UpdatePolicy(pid, updateOpts)
	require.NoError(t, err)
	assert.Len(t, policy.Statements, 2)
	assert.Equal(t, "new desc", policy.Description)
}

func TestListPolicy(t *testing.T) {
	client := helpers.GetClient("fixtures/identity/listpolicy")
	defer client.Stop()
	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err)
	c, err := client.GetCompartment(compartmentID)
	require.NoError(t, err)
	compartmentName := c.Name

	gid, err := helpers.CreateGroup(client)
	require.NoError(t, err)
	defer func() {
		err = client.DeleteGroup(gid, nil)
		assert.NoError(t, err)
	}()
	group, err := client.GetGroup(gid)
	require.NoError(t, err)
	statements := []string{
		fmt.Sprintf("Allow group %s to manage users in compartment %s", group.Name, compartmentName),
		fmt.Sprintf("Allow group %s to inspect all-resources in compartment %s", group.Name, compartmentName),
	}
	p1, err := helpers.CreatePolicy(client, []string{statements[0]}, compartmentID)
	require.NoError(t, err)
	defer func() {
		err = client.DeletePolicy(p1, nil)
		assert.NoError(t, err)
	}()
	p2, err := helpers.CreatePolicy(client, []string{statements[1]}, compartmentID)
	require.NoError(t, err)
	defer func() {
		err = client.DeletePolicy(p2, nil)
		assert.NoError(t, err)
	}()

	list, err := client.ListPolicies(compartmentID, nil)
	require.NoError(t, err)
	pols := map[string]bm.Policy{}
	for _, p := range list.Policies {
		pols[p.ID] = p
	}
	_, ok := pols[p1]
	assert.True(t, ok)
	_, ok = pols[p2]
	assert.True(t, ok)
}
