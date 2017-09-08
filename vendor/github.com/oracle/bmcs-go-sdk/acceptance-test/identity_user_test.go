// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,identity_user recording,all !recording

package acceptance

import (
	"testing"

	"github.com/stretchr/testify/suite"

	bm "github.com/MustWin/baremetal-sdk-go"
	"github.com/MustWin/baremetal-sdk-go/acceptance-test/helpers"
)

type IdentityUserTestSuite struct {
	suite.Suite
}

func (s *IdentityUserTestSuite) TestUser() {
	client := helpers.GetClient("fixtures/identity/user")
	defer client.Stop()

	id, err := helpers.CreateUser(client)

	s.Require().NoError(err)
	defer func() {
		err = client.DeleteUser(id, nil)
		s.NoError(err)
	}()
	s.Require().NotEmpty(id)
	u, err := client.GetUser(id)

	s.Require().NoError(err)
	s.Equal(id, u.ID)

}

func (s *IdentityUserTestSuite) TestUserUpdates() {
	client := helpers.GetClient("fixtures/identity/updateuser")
	defer client.Stop()
	id, err := helpers.CreateUser(client)
	s.Require().NoError(err)
	defer func() {
		client.DeleteUser(id, nil)
	}()

	opt := bm.UpdateIdentityOptions{
		Description: "new description",
	}

	u, err := client.UpdateUser(id, &opt)
	s.NoError(err)
	s.Equal("new description", u.Description)

	state := bm.UpdateUserStateOptions{
		Blocked: helpers.BoolPtr(false),
	}
	u, err = client.UpdateUserState(id, &state)
	s.NoError(err)

}

// create 10 users, list them, then delete
func (s *IdentityUserTestSuite) TestListUsers() {
	client := helpers.GetClient("fixtures/identity/listusers")
	defer client.Stop()
	var uids []string
	defer func() {
		for _, id := range uids {
			err := client.DeleteUser(id, nil)
			s.Require().NoError(err)
		}
	}()

	for i := 0; i < 5; i++ {
		id, err := helpers.CreateUser(client)
		s.Require().NoError(err)
		uids = append(uids, id)
	}

	var options bm.ListOptions
	options.Limit = 2
	listCalls := 0
	usersReturned := 0
	for {
		list, err := client.ListUsers(&options)
		s.Require().NoError(err)
		s.NotNil(list)
		listCalls++
		usersReturned += len(list.Users)
		if list.NextPage == "" {
			break
		}
		options.Page = list.NextPage
	}
	s.True(listCalls >= 2)
	s.True(usersReturned >= 5)
}

func TestIdentityUserTestSuite(t *testing.T) {
	suite.Run(t, new(IdentityUserTestSuite))
}
