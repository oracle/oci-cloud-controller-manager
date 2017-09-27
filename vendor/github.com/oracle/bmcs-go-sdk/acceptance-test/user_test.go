// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package acceptance

import (
	"testing"

	"github.com/stretchr/testify/assert"

	bm "github.com/oracle/bmcs-go-sdk"
	"github.com/oracle/bmcs-go-sdk/acceptance-test/helpers"
)

func TestUserCRUD(t *testing.T) {
	client := helpers.GetClient("fixtures/user")
	defer client.Stop()

	// Create
	var uids []string
	for i := 0; i < 5; i++ {
		id, err := helpers.CreateUser(client)
		assert.NoError(t, err, "Create %v", i)
		uids = append(uids, id)
	}

	// TODO: Get

	// List with pagination
	var options bm.ListOptions
	options.Limit = 1
	lenUsers := 0
	for i := 0; i < 5; i++ {
		list, err := client.ListUsers(&options)
		assert.NoError(t, err, "List")
		assert.NotNil(t, list, "List")
		lenUsers += len(list.Users)
		if list.NextPage == "" {
			break
		}
		options.Page = list.NextPage
	}
	assert.Equal(t, 5, lenUsers, "List: len")

	// Update
	id := uids[0]
	opt := bm.UpdateIdentityOptions{
		Description: "new description",
	}
	u, err := client.UpdateUser(id, &opt)
	assert.NoError(t, err, "Update")
	assert.Equal(t, "new description", u.Description, "Update: Description")

	// Update State
	state := bm.UpdateUserStateOptions{
		Blocked: helpers.BoolPtr(false),
	}
	u, err = client.UpdateUserState(id, &state)
	assert.NoError(t, err, "UpdateState")

	// Delete
	for i, id := range uids {
		err := client.DeleteUser(id, nil)
		assert.NoError(t, err, "Delete %v", i)
	}
}
