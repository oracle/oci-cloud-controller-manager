// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package acceptance

import (
	"testing"

	"github.com/stretchr/testify/assert"

	bm "github.com/oracle/bmcs-go-sdk"
	"github.com/oracle/bmcs-go-sdk/acceptance-test/helpers"
)

func TestGroupCRUD(t *testing.T) {
	client := helpers.GetClient("fixtures/group")
	defer client.Stop()

	// Create
	var gids []string
	for i := 0; i < 4; i++ {
		id, err := helpers.CreateGroup(client)
		assert.NoError(t, err, "Create %v", i)
		gids = append(gids, id)
	}

	// TODO: Get

	// Update
	id := gids[0]

	opt := bm.UpdateIdentityOptions{
		Description: "new description",
	}
	g, err := client.UpdateGroup(id, &opt)
	assert.NoError(t, err, "Update")
	assert.Equal(t, opt.Description, g.Description, "Update: Description")

	// List with pagination
	var options bm.ListOptions
	options.Limit = 1
	lenGroups := 0
	for i := 0; i < 2; i++ {
		list, err := client.ListGroups(&options)
		assert.NoError(t, err, "List with pagination: page %v", i)
		assert.NotNil(t, list, "List with pagination: page %v", i)
		lenGroups += len(list.Groups)
		if list.NextPage == "" {
			break
		}
		options.Page = list.NextPage
	}
	assert.Equal(t, 2, lenGroups, "List with pagination: len")

	// Delete
	for _, id := range gids {
		err := client.DeleteGroup(id, nil)
		assert.NoError(t, err, "Delete")
	}
}
