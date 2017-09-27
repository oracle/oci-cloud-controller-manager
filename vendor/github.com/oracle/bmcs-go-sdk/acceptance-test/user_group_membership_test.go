// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package acceptance

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bm "github.com/oracle/bmcs-go-sdk"
	"github.com/oracle/bmcs-go-sdk/acceptance-test/helpers"
)

func TestUserGroupMembershipsRUD(t *testing.T) {
	client := helpers.GetClient("fixtures/user_group_membership")
	defer client.Stop()
	// Create user
	uid, err := helpers.CreateUser(client)
	require.NoError(t, err, "Setup User")
	defer func() {
		err = client.DeleteUser(uid, nil)
		assert.NoError(t, err, "Teardown User")
	}()
	// Create group
	gid, err := helpers.CreateGroup(client)
	require.NoError(t, err, "Setup Group")
	defer func() {
		err = client.DeleteGroup(gid, nil)
		assert.NoError(t, err, "Teardown Group")
	}()

	// Create
	ugid, err := helpers.AddUserToGroup(client, uid, gid)
	assert.NoError(t, err, "Create")

	// Get
	ugm, err := client.GetUserGroupMembership(ugid)
	assert.NoError(t, err, "Get")
	assert.Equal(t, uid, ugm.UserID, "Get: UserID")
	assert.Equal(t, gid, ugm.GroupID, "Get: GroupID")

	// List with GID
	var opts bm.ListMembershipsOptions
	opts.Limit = 10
	opts.GroupID = gid
	list, err := client.ListUserGroupMemberships(&opts)
	assert.NoError(t, err, "List with GID")
	assert.Len(t, list.Memberships, 1, "List with GID")

	// List sans GID
	opts.GroupID = ""
	opts.UserID = uid
	list, err = client.ListUserGroupMemberships(&opts)
	assert.NoError(t, err, "List sans GID")
	assert.Len(t, list.Memberships, 1, "List sans GID")

	// Delete
	err = client.DeleteUserGroupMembership(ugid, nil)
	assert.NoError(t, err, "Delete")
}
