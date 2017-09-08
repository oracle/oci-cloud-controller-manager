// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,identity_group recording,all !recording

package acceptance

import (
	"testing"

	"github.com/stretchr/testify/suite"

	bm "github.com/MustWin/baremetal-sdk-go"
	"github.com/MustWin/baremetal-sdk-go/acceptance-test/helpers"
)

type IdentityGroupTestSuite struct {
	suite.Suite
}

func (s *IdentityGroupTestSuite) TestGroup() {
	client := helpers.GetClient("fixtures/identity/group")
	defer client.Stop()
	id, err := helpers.CreateGroup(client)
	s.Require().NoError(err)
	defer func() {
		err := client.DeleteGroup(id, nil)
		s.NoError(err)
	}()
	grp, err := client.GetGroup(id)
	s.Require().NoError(err)
	s.NotNil(grp)
}

func (s *IdentityGroupTestSuite) TestGroupUpdates() {
	client := helpers.GetClient("fixtures/identity/updategroup")
	defer client.Stop()
	id, err := helpers.CreateGroup(client)
	s.Require().NoError(err)
	defer func() {
		err = client.DeleteGroup(id, nil)
		s.NoError(err)
	}()

	opt := bm.UpdateIdentityOptions{
		Description: "new description",
	}

	g, err := client.UpdateGroup(id, &opt)
	s.NoError(err)
	s.Equal("new description", g.Description)
}

func (s *IdentityGroupTestSuite) TestListGroups() {
	client := helpers.GetClient("fixtures/identity/listgroups")
	defer client.Stop()
	var gids []string
	defer func() {
		for _, id := range gids {
			err := client.DeleteGroup(id, nil)
			s.NoError(err)
		}
	}()
	for i := 0; i < 4; i++ {
		id, err := helpers.CreateGroup(client)
		s.Require().NoError(err)
		gids = append(gids, id)
	}
	var options bm.ListOptions
	options.Limit = 2
	listCalls := 0
	returned := 0
	for {
		list, err := client.ListGroups(&options)
		s.Require().NoError(err)
		s.NotNil(list)
		listCalls++
		returned += len(list.Groups)
		if list.NextPage == "" {
			break
		}
		options.Page = list.NextPage
	}
	s.True(listCalls >= 2)
	s.True(returned >= 3)
}

func TestIndentityGroupTestSuite(t *testing.T) {
	suite.Run(t, new(IdentityGroupTestSuite))
}
