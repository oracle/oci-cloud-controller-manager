// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,core recording,all !recording

package acceptance

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	bm "github.com/MustWin/baremetal-sdk-go"
	//"fmt"
	"time"
)

const vcnAddress = "172.16.0.0/16"

type CoreTestSuite struct {
	compartmentID       string
	availabilityDomains []bm.AvailabilityDomain
	suite.Suite
}

func TestCoreTestSuite(t *testing.T) {
	suite.Run(t, new(CoreTestSuite))
}

func (s *CoreTestSuite) SetupSuite() {
	client := getClient("fixtures/core/setup")
	defer client.Stop()
	// get a compartment, any compartment
	var listOpts bm.ListOptions
	listOpts.Limit = 1
	list, err := client.ListCompartments(&listOpts)
	s.Require().NoError(err)
	if len(list.Compartments) == 1 {
		s.compartmentID = list.Compartments[0].ID
	} else {
		id, err := resourceApply(createCompartment(client))
		s.Require().NoError(err)
		s.compartmentID = id
	}

	// Get Availability Domains

	ads, err := client.ListAvailabilityDomains(s.compartmentID)
	s.availabilityDomains = ads.AvailabilityDomains
	s.Require().NoError(err)
}

func (s *CoreTestSuite) TestCpeCreate() {
	client := getClient("fixtures/core/instance")
	defer client.Stop()

	cpe, err := client.CreateCpe(s.compartmentID, "120.90.41.18", nil)
	s.Require().NoError(err)
	defer func() {
		err := client.DeleteCpe(cpe.ID, nil)
		s.NoError(err)
	}()
	s.Require().NotEmpty(cpe.ID)
}

func (s *CoreTestSuite) TestCpeList() {
	client := getClient("fixtures/core/instance")
	defer client.Stop()

	cpe, err := client.CreateCpe(s.compartmentID, "120.90.41.18", nil)
	s.Require().NoError(err)
	defer func() {
		err := client.DeleteCpe(cpe.ID, nil)
		s.NoError(err)
	}()
	s.Require().NotEmpty(cpe.ID)

	cpes, err := client.ListCpes(s.compartmentID, nil)
	s.Require().NoError(err)
	found := false
	for _, ce := range cpes.Cpes {
		if strings.Compare(ce.ID, cpe.ID) == 0 {
			found = true
		}
	}
	s.Require().True(found, "Created Instance not found in list")
}

func (s *CoreTestSuite) TestDhcpOptionsCreate() {
	client := getClient("fixtures/core/dhcp_options")
	defer client.Stop()

	vcnID, err := resourceApply(createVCN(client, "172.16.0.0/16", s.compartmentID))
	s.Require().NoError(err)
	defer func() {
		_, err = resourceApply(deleteVCN(client, vcnID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(vcnID)

	dhcpID, err := resourceApply(createDhcpOption(client, s.compartmentID, vcnID))
	s.Require().NoError(err)
	defer func() {
		_, err = resourceApply(deleteDhcpOption(client, dhcpID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(dhcpID)

}

func (s *CoreTestSuite) TestDhcpOptionsList() {
	client := getClient("fixtures/core/dhcp_options_list")
	defer client.Stop()

	vcnID, err := resourceApply(createVCN(client, "172.16.0.0/16", s.compartmentID))
	s.Require().NoError(err)
	defer func() {
		_, err = resourceApply(deleteVCN(client, vcnID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(vcnID)

	dhcpID, err := resourceApply(createDhcpOption(client, s.compartmentID, vcnID))
	s.Require().NoError(err)
	defer func() {
		_, err = resourceApply(deleteDhcpOption(client, dhcpID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(dhcpID)

	opts, err := client.ListDHCPOptions(s.compartmentID, vcnID, nil)
	s.Require().NoError(err)
	found := false
	for _, opt := range opts.DHCPOptions {
		if strings.Compare(dhcpID, opt.ID) == 0 {
			found = true
		}
	}
	s.Require().True(found, "Created Instance not found in list")
}

func (s *CoreTestSuite) TestDrgCreate() {
	client := getClient("fixtures/core/drg")
	defer client.Stop()

	drgID, err := resourceApply(createDrg(client, s.compartmentID))
	s.Require().NoError(err)
	defer func() {
		_, err = resourceApply(deleteDrg(client, drgID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(drgID)

}

func (s *CoreTestSuite) TestDrgList() {
	client := getClient("fixtures/core/drg_list")
	defer client.Stop()

	drgID, err := resourceApply(createDrg(client, s.compartmentID))
	s.Require().NoError(err)
	defer func() {
		_, err = resourceApply(deleteDrg(client, drgID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(drgID)

	drgs, err := client.ListDrgs(s.compartmentID, nil)
	s.Require().NoError(err)
	found := false
	for _, drg := range drgs.Drgs {
		if strings.Compare(drgID, drg.ID) == 0 {
			found = true
		}
	}
	s.Require().True(found, "Created Drg not found in list")
}

func (s *CoreTestSuite) TestInternetGatewayCreate() {
	client := getClient("fixtures/core/internet_gateway")
	defer client.Stop()

	vcnID, err := resourceApply(createVCN(client, "172.16.0.0/16", s.compartmentID))
	s.Require().NoError(err)
	defer func() {
		_, err = resourceApply(deleteVCN(client, vcnID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(vcnID)

	igID, err := resourceApply(createInternetGateway(client, s.compartmentID, vcnID))
	s.Require().NoError(err)
	defer func() {
		_, err := resourceApply(deleteInternetGateway(client, igID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(igID)
}

func (s *CoreTestSuite) TestInternetGatewayList() {
	client := getClient("fixtures/core/internet_gateway_list")
	defer client.Stop()

	vcnID, err := resourceApply(createVCN(client, "172.16.0.0/16", s.compartmentID))
	s.Require().NoError(err)
	defer func() {
		_, err = resourceApply(deleteVCN(client, vcnID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(vcnID)

	igID, err := resourceApply(createInternetGateway(client, s.compartmentID, vcnID))
	s.Require().NoError(err)
	defer func() {
		_, err := resourceApply(deleteInternetGateway(client, igID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(igID)

	igs, err := client.ListInternetGateways(s.compartmentID, vcnID, nil)
	s.Require().NoError(err)
	found := false
	for _, ig := range igs.Gateways {
		if strings.Compare(igID, ig.ID) == 0 {
			found = true
		}
	}
	s.Require().True(found, "Created Internet Gateway not found in list")
}

func (s *CoreTestSuite) TestRouteTableCreate() {
	client := getClient("fixtures/core/route_table")
	defer client.Stop()

	vcnID, err := resourceApply(createVCN(client, "172.16.0.0/16", s.compartmentID))
	s.Require().NoError(err)
	defer func() {
		_, err = resourceApply(deleteVCN(client, vcnID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(vcnID)

	subnetID, err := resourceApply(createSubnet(client, s.compartmentID, s.availabilityDomains[0].Name, vcnID))
	s.Require().NoError(err)
	defer func() {
		_, err := resourceApply(deleteSubnet(client, subnetID))
		s.NoError(err)
		time.Sleep(2 * time.Second) // Subnet deletes don't happen immediately
	}()
	s.Require().NotEmpty(subnetID)

	igID, err := resourceApply(createInternetGateway(client, s.compartmentID, vcnID))
	s.Require().NoError(err)
	defer func() {
		_, err := resourceApply(deleteInternetGateway(client, igID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(igID)

	rtID, err := resourceApply(createRouteTable(client, s.compartmentID, vcnID, igID))
	s.Require().NoError(err)
	defer func() {
		_, err := resourceApply(deleteRouteTable(client, rtID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(rtID)
}

func (s *CoreTestSuite) TestRouteTableList() {
	client := getClient("fixtures/core/route_table_list")
	defer client.Stop()

	vcnID, err := resourceApply(createVCN(client, "172.16.0.0/16", s.compartmentID))
	s.Require().NoError(err)
	defer func() {
		_, err = resourceApply(deleteVCN(client, vcnID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(vcnID)

	subnetID, err := resourceApply(createSubnet(client, s.compartmentID, s.availabilityDomains[0].Name, vcnID))
	s.Require().NoError(err)
	defer func() {
		_, err := resourceApply(deleteSubnet(client, subnetID))
		s.NoError(err)
		time.Sleep(2 * time.Second)
	}()
	s.Require().NotEmpty(subnetID)

	igID, err := resourceApply(createInternetGateway(client, s.compartmentID, vcnID))
	s.Require().NoError(err)
	defer func() {
		_, err := resourceApply(deleteInternetGateway(client, igID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(igID)

	rtID, err := resourceApply(createRouteTable(client, s.compartmentID, vcnID, igID))
	s.Require().NoError(err)
	defer func() {
		_, err := resourceApply(deleteRouteTable(client, rtID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(rtID)

	rts, err := client.ListRouteTables(s.compartmentID, vcnID, nil)
	s.Require().NoError(err)
	found := false
	for _, rt := range rts.RouteTables {
		if strings.Compare(rtID, rt.ID) == 0 {
			found = true
		}
	}
	s.Require().True(found, "Created Internet Gateway not found in list")
}

func (s *CoreTestSuite) TestSecurityListCreate() {
	client := getClient("fixtures/core/security_list")
	defer client.Stop()

	vcnID, err := resourceApply(createVCN(client, "172.16.0.0/16", s.compartmentID))
	s.Require().NoError(err)
	defer func() {
		_, err = resourceApply(deleteVCN(client, vcnID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(vcnID)

	slID, err := resourceApply(createSecurityList(client, s.compartmentID, vcnID))
	s.Require().NoError(err)
	defer func() {
		_, err := resourceApply(deleteSecurityList(client, slID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(slID)
}

func (s *CoreTestSuite) TestSecurityListList() {
	client := getClient("fixtures/core/security_list_list")
	defer client.Stop()

	vcnID, err := resourceApply(createVCN(client, "172.16.0.0/16", s.compartmentID))
	s.Require().NoError(err)
	defer func() {
		_, err = resourceApply(deleteVCN(client, vcnID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(vcnID)

	subnetID, err := resourceApply(createSubnet(client, s.compartmentID, s.availabilityDomains[0].Name, vcnID))
	s.Require().NoError(err)
	defer func() {
		_, err := resourceApply(deleteSubnet(client, subnetID))
		s.NoError(err)
		time.Sleep(2 * time.Second)
	}()
	s.Require().NotEmpty(subnetID)

	slID, err := resourceApply(createSecurityList(client, s.compartmentID, vcnID))
	s.Require().NoError(err)
	defer func() {
		_, err := resourceApply(deleteSecurityList(client, slID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(slID)

	sls, err := client.ListSecurityLists(s.compartmentID, vcnID, nil)
	s.Require().NoError(err)
	found := false
	for _, sl := range sls.SecurityLists {
		if strings.Compare(slID, sl.ID) == 0 {
			found = true
		}
	}
	s.Require().True(found, "Created Internet Gateway not found in list")
}

func (s *CoreTestSuite) TestSubnetCreate() {
	client := getClient("fixtures/core/subnet")
	defer client.Stop()

	vcnID, err := resourceApply(createVCN(client, "172.16.0.0/16", s.compartmentID))
	s.Require().NoError(err)
	defer func() {
		_, err = resourceApply(deleteVCN(client, vcnID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(vcnID)

	id, err := resourceApply(createSubnet(client, s.compartmentID, s.availabilityDomains[0].Name, vcnID))
	s.Require().NoError(err)
	defer func() {
		_, err := resourceApply(deleteSubnet(client, id))
		s.NoError(err)
		time.Sleep(2 * time.Second)
	}()
	s.Require().NotEmpty(id)

	subnet, err := client.GetSubnet(id)
	s.Require().NoError(err)
	s.Equal(id, subnet.ID)
}

func (s *CoreTestSuite) TestSubnetList() {
	client := getClient("fixtures/core/subnet_list")
	defer client.Stop()

	vcnID, err := resourceApply(createVCN(client, "172.16.0.0/16", s.compartmentID))
	s.Require().NoError(err)
	defer func() {
		_, err = resourceApply(deleteVCN(client, vcnID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(vcnID)

	id, err := resourceApply(createSubnet(client, s.compartmentID, s.availabilityDomains[0].Name, vcnID))
	s.Require().NoError(err)
	defer func() {
		_, err := resourceApply(deleteSubnet(client, id))
		s.NoError(err)
		time.Sleep(2 * time.Second)
	}()
	s.Require().NotEmpty(id)

	subnets, err := client.ListSubnets(s.compartmentID, vcnID, nil)
	s.Require().NoError(err)

	found := false
	for _, sub := range subnets.Subnets {
		if strings.Compare(sub.ID, id) == 0 {
			found = true
		}
	}
	s.Require().True(found, "Created subnet not found in list")
}

func (s *CoreTestSuite) TestVCNCreate() {
	client := getClient("fixtures/core/vcn")
	defer client.Stop()

	id, err := resourceApply(createVCN(client, "172.16.0.0/16", s.compartmentID))
	s.Require().NoError(err)
	defer func() {
		_, err = resourceApply(deleteVCN(client, id))
		s.NoError(err)
	}()
	s.Require().NotEmpty(id)

	vcn, err := client.GetVirtualNetwork(id)

	s.Require().NoError(err)
	s.Equal(id, vcn.ID)
}

func (s *CoreTestSuite) TestVCNList() {
	client := getClient("fixtures/core/vcn_list")
	defer client.Stop()

	id, err := resourceApply(createVCN(client, "172.16.0.0/16", s.compartmentID))

	s.Require().NoError(err)
	s.Require().NotEmpty(id)

	defer func() {
		_, err = resourceApply(deleteVCN(client, id))
		s.NoError(err)
	}()

	vcns, err := client.ListVirtualNetworks(s.compartmentID, nil)

	s.Require().NoError(err)
	found := false
	for _, vcn := range vcns.VirtualNetworks {
		if strings.Compare(vcn.ID, id) == 0 {
			found = true
		}
	}
	s.Require().True(found, "Created VCN not found in list")
}

func (s *CoreTestSuite) TestVCNUpdate() {
	s.T().Skip("UpdateVirtualNetwork is broken, it says it's 'Terminating'")

	client := getClient("fixtures/core/vcn_update")
	defer client.Stop()

	vcnID, err := resourceApply(createVCN(client, vcnAddress, s.compartmentID))

	s.Require().NoError(err)
	s.Require().NotEmpty(vcnID)

	defer func() {
		_, err = resourceApply(deleteVCN(client, vcnID))
		s.NoError(err)
	}()

	newName := "vcn2"
	var vcn *bm.VirtualNetwork
	opts := &bm.IfMatchDisplayNameOptions{DisplayNameOptions: bm.DisplayNameOptions{DisplayName: newName}}

	_, err = resourceApply(func(c chan<- resourceCommandResult) bool {
		vcn, err = client.UpdateVirtualNetwork(vcnID, opts)
		if err != nil {
			c <- resourceCommandResult{"", err}
			return true
		} else if vcn.State == bm.ResourceAvailable {
			c <- resourceCommandResult{vcn.ID, nil}
			return true
		}
		return false
	})

	s.Require().NoError(err)
	s.Require().Equal(vcn.DisplayName, newName)
}
