// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package acceptance

import (
	"testing"

	"github.com/stretchr/testify/assert"

	bm "github.com/oracle/bmcs-go-sdk"
	"github.com/oracle/bmcs-go-sdk/acceptance-test/helpers"
)

func TestCompartmentCRUD(t *testing.T) {
	client := helpers.GetClient("fixtures/compartment")
	defer client.Stop()

	// Create
	// TODO: find a safe way to test creating compartments, which cannot be deleted
	// id, err := helpers.CreateCompartment(client)
	// assert.NoError(t, err, "Create")

	// TODO: Get
	// compartment, err := client.GetCompartment(id)
	// assert.NoError(t, err, "Get")
	// assert.NotNil(t, compartment, "Get")

	// TODO: Update
	// updateCompartment := bm.UpdateIdentityOptions{
	// 	Description: "new desc",
	// }
	// compartment, err = client.UpdateCompartment(id, &updateCompartment)
	// assert.NoError(t, err, "Update")
	// assert.Equal(t, "new desc", compartment.Description, "Update: Description")

	// List sans pagination
	var options bm.ListOptions
	options.Limit = 1

	list, err := client.ListCompartments(&options)
	assert.NoError(t, err, "List sans pagination")
	assert.NotNil(t, list, "List sans pagination")

	assert.Len(t, list.Compartments, 1, "List sans pagination")

	// List with pagination
	compartments := 0
	wantLen := 2
	for i := 0; i < wantLen; i++ {
		list, err := client.ListCompartments(&options)
		assert.NoError(t, err, "List with pagination, page %v", i)
		assert.NotNil(t, list, "List with pagination, page %v", i)

		compartments += len(list.Compartments)
		if list.NextPage == "" {
			break
		}
		options.Page = list.NextPage
	}

	assert.Equal(t, wantLen, compartments, "List with pagination: len")

	// Note: compartments do not support Delete
}
