// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,load_balancer_certificate recording,all !recording

package acceptance

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	bm "github.com/MustWin/baremetal-sdk-go"
)

type LoadBalancerCertificateTestSuite struct {
	availabilityDomains []bm.AvailabilityDomain
	shapes              []bm.LoadBalancerShape
	compartmentID       string
	vcnID               string
	subnetIDs           []string
	suite.Suite
}

func TestLoadBalancerCertificateTestSuite(t *testing.T) {
	suite.Run(t, new(LoadBalancerCertificateTestSuite))
}

func (s *LoadBalancerCertificateTestSuite) SetupSuite() {
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
	}

	time.Sleep(TIMEOUT) // TODO: can we verify the subnets have been created?
}

func (s *LoadBalancerCertificateTestSuite) TearDownSuite() {
	client := getClient("fixtures/load_balancer/setup")
	defer client.Stop()
	// VCN requires all subnets to be deleted first
	for _, subnetID := range s.subnetIDs {
		resourceApply(deleteSubnet(client, subnetID))
	}
	resourceApply(deleteVCN(client, s.vcnID))
}

func (s *LoadBalancerCertificateTestSuite) TestCreateLoadBalancerCertificate() {
	client := getClient("fixtures/load_balancer_certificate/create")
	defer client.Stop()

	// Minimal stub dependencies
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

	// Create SUT
	workRequestID, err = client.CreateCertificate(
		workRequest.LoadBalancerID,
		"certificate-name",
		"",
		"-----BEGIN RSA PRIVATE KEY-----\nMIIBOgIBAAJBAOUzyXPcEUkDrMGWwXreT1qM9WrdDVZCgdDePfnTwNEoh/Cp9X4L\nEvrdbd1mvAvhOuOqis/kJDfr4jo5YAsfbNUCAwEAAQJAJz8k4bfvJceBT2zXGIj0\noZa9d1z+qaSdwfwsNJkzzRyGkj/j8yv5FV7KNdSfsBbStlcuxUm4i9o5LXhIA+iQ\ngQIhAPzStAN8+Rz3dWKTjRWuCfy+Pwcmyjl3pkMPSiXzgSJlAiEA6BUZWHP0b542\nu8AizBT3b3xKr1AH2nkIx9OHq7F/QbECIHzqqpDypa8/QVuUZegpVrvvT/r7mn1s\nddS6cDtyJgLVAiEA1Z5OFQeuL2sekBRbMyP9WOW7zMBKakLL3TqL/3JCYxECIAkG\nl96uo1MjK/66X5zQXBG7F2DN2CbcYEz0r3c3vvfq\n-----END RSA PRIVATE KEY-----",
		"",
		"-----BEGIN CERTIFICATE-----\nMIIBNzCB4gIJAKtwJkxUgNpzMA0GCSqGSIb3DQEBCwUAMCMxITAfBgNVBAoTGElu\ndGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0xNzA0MTIyMTU3NTZaFw0xODA0MTIy\nMTU3NTZaMCMxITAfBgNVBAoTGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDBcMA0G\nCSqGSIb3DQEBAQUAA0sAMEgCQQDlM8lz3BFJA6zBlsF63k9ajPVq3Q1WQoHQ3j35\n08DRKIfwqfV+CxL63W3dZrwL4TrjqorP5CQ36+I6OWALH2zVAgMBAAEwDQYJKoZI\nhvcNAQELBQADQQCEjHVQJoiiVpIIvDWF+4YDRReVuwzrvq2xduWw7CIsDWlYuGZT\nQKVY6tnTy2XpoUk0fqUvMB/M2HGQ1WqZGHs6\n-----END CERTIFICATE-----",
		nil,
	)
	s.NoError(err)
	s.Require().NotEmpty(workRequestID)
	for {
		workRequest, err = client.GetWorkRequest(workRequestID, nil)
		s.NoError(err)
		if workRequest.State == "SUCCEEDED" {
			break
		}
		time.Sleep(TIMEOUT)
	}

	// Get SUT
	log.Printf("Get SUT\n")
	lb, err := client.GetLoadBalancer(workRequest.LoadBalancerID, nil)
	s.Require().NoError(err)

	s.Equal(
		bm.Certificate{
			CertificateName: "certificate-name",
			// PrivateKey should not be included in response, only in request
			PublicCertificate: "-----BEGIN CERTIFICATE-----\nMIIBNzCB4gIJAKtwJkxUgNpzMA0GCSqGSIb3DQEBCwUAMCMxITAfBgNVBAoTGElu\ndGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0xNzA0MTIyMTU3NTZaFw0xODA0MTIy\nMTU3NTZaMCMxITAfBgNVBAoTGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDBcMA0G\nCSqGSIb3DQEBAQUAA0sAMEgCQQDlM8lz3BFJA6zBlsF63k9ajPVq3Q1WQoHQ3j35\n08DRKIfwqfV+CxL63W3dZrwL4TrjqorP5CQ36+I6OWALH2zVAgMBAAEwDQYJKoZI\nhvcNAQELBQADQQCEjHVQJoiiVpIIvDWF+4YDRReVuwzrvq2xduWw7CIsDWlYuGZT\nQKVY6tnTy2XpoUk0fqUvMB/M2HGQ1WqZGHs6\n-----END CERTIFICATE-----",
		},
		lb.Certificates["certificate-name"],
	)

	// Note: Certificates have no Update operation

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
}
