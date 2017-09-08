// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,identity_compartment recording,all !recording

package acceptance

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bm "github.com/MustWin/baremetal-sdk-go"
	"github.com/MustWin/baremetal-sdk-go/acceptance-test/helpers"
)

// TODO: find a safe way to test creating compartments, which cannot be deleted
// func TestCompartment(t *testing.T) {
// 	client := helpers.GetClient("fixtures/identity/compartment")
// 	defer client.Stop()
// 	id, err := helpers.CreateCompartment(client)
// 	require.NoError(t, err)

// 	compartment, err := client.GetCompartment(id)
// 	require.NoError(t, err)
// 	require.NotNil(t, compartment)
// 	updateCompartment := bm.UpdateIdentityOptions{
// 		Description: "new desc",
// 	}
// 	compartment, err = client.UpdateCompartment(id, &updateCompartment)
// 	assert.NoError(t, err)

// 	assert.Equal(t, "new desc", compartment.Description)
// }

func TestListCompartment(t *testing.T) {
	client := helpers.GetClient("fixtures/identity/listcompartments")
	defer client.Stop()
	var options bm.ListOptions
	calls := 0
	compartments := 0
	options.Limit = 4
	for {
		list, err := client.ListCompartments(&options)
		require.NoError(t, err)
		require.NotNil(t, list)
		calls++
		compartments += len(list.Compartments)
		if list.NextPage == "" {
			break
		}
		options.Page = list.NextPage
	}
	assert.Equal(t, 12, calls)
	assert.Equal(t, 43, compartments)
}
