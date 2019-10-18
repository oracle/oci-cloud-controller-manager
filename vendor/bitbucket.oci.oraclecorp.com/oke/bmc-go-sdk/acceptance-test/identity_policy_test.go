// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,identity_policy recording,all !recording

package acceptance

import (
	"fmt"
	"testing"

	bm "github.com/MustWin/baremetal-sdk-go"

	"github.com/stretchr/testify/suite"
)

type IdentityPolicyTestSuite struct {
	suite.Suite
	compartmentID   string
	compartmentName string
}

func (s *IdentityPolicyTestSuite) SetupSuite() {
	client := getClient("fixtures/identity/policysetup")
	defer client.Stop()
	// get a compartment, any compartment
	var listOpts bm.ListOptions
	listOpts.Limit = 1
	list, err := client.ListCompartments(&listOpts)
	s.Require().NoError(err)
	if len(list.Compartments) == 1 {
		s.compartmentID = list.Compartments[0].ID
		s.compartmentName = list.Compartments[0].Name
	} else {
		id, err := resourceApply(createCompartment(client))
		s.Require().NoError(err)
		c, err := client.GetCompartment(id)
		s.Require().NoError(err)
		s.compartmentID = id
		s.compartmentName = c.Name
	}
}

func (s *IdentityPolicyTestSuite) TestPolicyLifecycle() {
	client := getClient("fixtures/identity/policy")
	defer client.Stop()
	// create group and get its name, we need this for statement
	gid, err := resourceApply(createGroup(client))
	s.Require().NoError(err)
	defer func() {
		err = client.DeleteGroup(gid, nil)
		s.NoError(err)
	}()
	group, err := client.GetGroup(gid)
	s.Require().NoError(err)
	statements := []string{
		fmt.Sprintf("Allow group %s to manage users in compartment %s", group.Name, s.compartmentName),
	}
	pid, err := resourceApply(createPolicy(client, statements, s.compartmentID))
	s.Require().NoError(err)
	defer func() {
		err = client.DeletePolicy(pid, nil)
		s.NoError(err)
	}()

	policy, err := client.GetPolicy(pid)
	s.NoError(err)
	s.Len(policy.Statements, 1)
}

func (s *IdentityPolicyTestSuite) TestPolicyUpdate() {
	client := getClient("fixtures/identity/updatepolicy")
	defer client.Stop()
	// create group and get its name, we need this for statement
	gid, err := resourceApply(createGroup(client))
	s.Require().NoError(err)
	defer func() {
		err = client.DeleteGroup(gid, nil)
		s.NoError(err)
	}()
	group, err := client.GetGroup(gid)
	s.Require().NoError(err)
	statements := []string{
		fmt.Sprintf("Allow group %s to manage users in compartment %s", group.Name, s.compartmentName),
	}
	pid, err := resourceApply(createPolicy(client, statements, s.compartmentID))
	s.Require().NoError(err)
	defer func() {
		err = client.DeletePolicy(pid, nil)
		s.NoError(err)
	}()

	statements = append(statements, fmt.Sprintf("Allow group %s to inspect all-resources in compartment %s", group.Name, s.compartmentName))
	updateOpts := &bm.UpdatePolicyOptions{
		UpdateIdentityOptions: bm.UpdateIdentityOptions{
			Description: "new desc",
		},
		Statements: statements,
	}

	policy, err := client.UpdatePolicy(pid, updateOpts)
	s.Require().NoError(err)
	s.Len(policy.Statements, 2)
	s.Equal("new desc", policy.Description)
}

func (s *IdentityPolicyTestSuite) TestListPolicy() {
	client := getClient("fixtures/identity/listpolicy")
	defer client.Stop()

	gid, err := resourceApply(createGroup(client))
	s.Require().NoError(err)
	defer func() {
		err = client.DeleteGroup(gid, nil)
		s.NoError(err)
	}()
	group, err := client.GetGroup(gid)
	s.Require().NoError(err)
	statements := []string{
		fmt.Sprintf("Allow group %s to manage users in compartment %s", group.Name, s.compartmentName),
		fmt.Sprintf("Allow group %s to inspect all-resources in compartment %s", group.Name, s.compartmentName),
	}
	p1, err := resourceApply(createPolicy(client, []string{statements[0]}, s.compartmentID))
	s.Require().NoError(err)
	defer func() {
		err = client.DeletePolicy(p1, nil)
		s.NoError(err)
	}()
	p2, err := resourceApply(createPolicy(client, []string{statements[1]}, s.compartmentID))
	s.Require().NoError(err)
	defer func() {
		err = client.DeletePolicy(p2, nil)
		s.NoError(err)
	}()

	list, err := client.ListPolicies(s.compartmentID, nil)
	s.Require().NoError(err)
	pols := map[string]bm.Policy{}
	for _, p := range list.Policies {
		pols[p.ID] = p
	}
	_, ok := pols[p1]
	s.True(ok)
	_, ok = pols[p2]
	s.True(ok)

}

func TestIdentityPolicyTestSuite(t *testing.T) {
	suite.Run(t, new(IdentityPolicyTestSuite))
}
