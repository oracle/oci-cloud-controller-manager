// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,load_balancer recording,all !recording

package acceptance

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	bm "github.com/MustWin/baremetal-sdk-go"
)

const TIMEOUT = (2 * time.Second)

type LoadBalancerTestSuite struct {
	availabilityDomains []bm.AvailabilityDomain
	shapes              []bm.LoadBalancerShape
	compartmentID       string
	vcnID               string
	subnetIDs           []string
	suite.Suite
}

func TestLoadBalancerTestSuite(t *testing.T) {
	suite.Run(t, new(LoadBalancerTestSuite))
}

func (s *LoadBalancerTestSuite) SetupSuite() {
	client := getClient("fixtures/load_balancer/setup")
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
	s.Require().NoError(err)
	s.availabilityDomains = ads.AvailabilityDomains

	// populate shapeName from ListShapes() {
	shapeList, err := client.ListLoadBalancerShapes(s.compartmentID, nil)
	s.Require().NoError(err)
	s.shapes = shapeList.LoadBalancerShapes

	s.vcnID, err = resourceApply(createVCN(client, "172.16.0.0/16", s.compartmentID))
	s.Require().NoError(err)
	s.Require().NotEmpty(s.vcnID)

	// Load Balancers require 2 subnets, each in seperate Availability Domains
	s.subnetIDs = make([]string, 2)
	for i := range s.subnetIDs {
		s.subnetIDs[i], err = resourceApply(createSubnetWithCIDR(
			client,
			s.compartmentID,
			s.availabilityDomains[i].Name,
			s.vcnID,
			fmt.Sprintf("172.16.%d.0/24", i),
		))
		s.Require().NoError(err)
		s.Require().NotEmpty(s.subnetIDs[i])
	}

	time.Sleep(TIMEOUT) // TODO: can we verify the subnets have been created?
}

func (s *LoadBalancerTestSuite) TearDownSuite() {
	client := getClient("fixtures/load_balancer/setup")
	defer client.Stop()
	// VCN requires all subnets to be deleted first
	for _, subnetID := range s.subnetIDs {
		resourceApply(deleteSubnet(client, subnetID))
	}
	resourceApply(deleteVCN(client, s.vcnID))
}

func (s *LoadBalancerTestSuite) TestCreateLoadBalancer() {
	client := getClient("fixtures/load_balancer/create")
	defer client.Stop()

	// Create SUT
	workRequestID, err := client.CreateLoadBalancer(
		nil,
		nil,
		s.compartmentID,
		nil,
		s.shapes[0].Name,
		s.subnetIDs,
		&bm.CreateOptions{
			DisplayNameOptions: bm.DisplayNameOptions{
				DisplayName: "my test LB",
			},
		},
	)
	s.Require().NoError(err)
	s.Require().NotEmpty(workRequestID)
	log.Printf("[DEBUG] Load Balancer Create Requested: %v", workRequestID)

	var workRequest *bm.WorkRequest
	for {
		workRequest, err = client.GetWorkRequest(workRequestID, nil)
		s.NoError(err)
		if workRequest.State == "SUCCEEDED" {
			break
		}
		time.Sleep(TIMEOUT) // wait until create is complete
	}
	s.Require().NotEmpty(workRequest.LoadBalancerID)

	// Get SUT
	lb, err := client.GetLoadBalancer(workRequest.LoadBalancerID, nil)
	s.Require().NoError(err)

	s.Equal(s.compartmentID, lb.CompartmentID)
	s.Equal("my test LB", lb.DisplayName)
	s.Equal("100Mbps", lb.Shape)
	// SubnetIDs should use Set equivalance
	s.Len(lb.SubnetIDs, len(s.subnetIDs))
	for _, subnetID := range s.subnetIDs {
		s.Contains(
			lb.SubnetIDs,
			subnetID,
		)
	}
	// Note: Backend, Listener & Certificate operations happen in other test suites
	s.Equal(bm.BackendSet{}, lb.BackendSets)
	s.Equal(bm.Listener{}, lb.Listeners)
	s.Equal(bm.Certificate{}, lb.Certificates)
	// Computed
	s.NotEmpty(lb.ID)
	s.NotEmpty(lb.IPAddresses)
	s.NotEmpty(lb.TimeCreated)

	// TODO: Update SUT

	// Delete SUT
	workRequestID, err = client.DeleteLoadBalancer(lb.ID, nil)
	s.NoError(err)
	s.Require().NotEmpty(workRequestID)
	for {
		workRequest, err = client.GetWorkRequest(workRequestID, nil)
		s.NoError(err)
		if workRequest.State == "SUCCEEDED" {
			break
		}
		time.Sleep(TIMEOUT) // wait until delete is complete
	}

	// lb, err = client.GetLoadBalancer()
}
