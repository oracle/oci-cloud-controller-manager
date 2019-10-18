// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,identity_user_group_membership recording,all !recording

package acceptance

import (
	"testing"

	"github.com/stretchr/testify/suite"

	bm "github.com/MustWin/baremetal-sdk-go"
)

type IdentityUserGroupMembershipSuite struct {
	suite.Suite
}

func (s *IdentityUserGroupMembershipSuite) TestUserGroups() {
	client := getClient("fixtures/identity/usergroups")
	defer client.Stop()
	uid, err := resourceApply(createUser(client))
	s.Require().NoError(err)
	defer func() {
		err = client.DeleteUser(uid, nil)
		s.NoError(err)
	}()
	gid, err := resourceApply(createGroup(client))
	s.Require().NoError(err)
	defer func() {
		err = client.DeleteGroup(gid, nil)
		s.NoError(err)
	}()
	ugid, err := resourceApply(addUserToGroup(client, uid, gid))
	s.Require().NoError(err)
	ugm, err := client.GetUserGroupMembership(ugid)
	s.Require().NoError(err)
	s.Equal(uid, ugm.UserID)
	s.Equal(gid, ugm.GroupID)

	var opts bm.ListMembershipsOptions
	opts.Limit = 10
	opts.GroupID = gid
	list, err := client.ListUserGroupMemberships(&opts)
	s.Require().NoError(err)
	s.True(len(list.Memberships) == 1)

	opts.Limit = 10
	opts.GroupID = ""
	opts.UserID = uid
	list, err = client.ListUserGroupMemberships(&opts)
	s.Require().NoError(err)
	s.True(len(list.Memberships) == 1)

	err = client.DeleteUserGroupMembership(ugid, nil)
	s.NoError(err)
}

func TestIdentityUserGroupMembershipSuite(t *testing.T) {
	suite.Run(t, new(IdentityUserGroupMembershipSuite))
}
