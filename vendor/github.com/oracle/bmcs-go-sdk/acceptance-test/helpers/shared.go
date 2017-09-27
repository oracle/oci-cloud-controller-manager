// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package helpers

import (
	"log"
	"time"

	bm "github.com/oracle/bmcs-go-sdk"
)

func Sleep(d time.Duration) {
	if RUNMODE == RunmodeRecord {
		time.Sleep(d)
	}
}

func CreateVCNWithOptions(client *TestClient, cidr, compartmentID string, opts *bm.CreateVcnOptions) (string, error) {
	log.Printf("[DEBUG] Create VCN with CIDR:%v", cidr)
	if opts == nil {
		opts = &bm.CreateVcnOptions{}
	}
	vcn, err := client.CreateVirtualNetwork(cidr, compartmentID, opts)
	return resourceApply(getCreateFn(err, vcn, bm.ResourceAvailable, func() (interface{}, error) {
		return client.GetVirtualNetwork(vcn.ID)
	}))
}

func CreateVCN(client *TestClient, cidr, compartmentID string) (string, error) {
	unique := RandomText(8)
	displayName := "vcn_" + unique
	retryToken := "retry_token_" + unique
	opts := &bm.CreateVcnOptions{
		CreateOptions: bm.CreateOptions{
			DisplayNameOptions: bm.DisplayNameOptions{DisplayName: displayName},
			RetryTokenOptions:  bm.RetryTokenOptions{RetryToken: retryToken},
		},
	}
	return CreateVCNWithOptions(client, cidr, compartmentID, opts)
}

func DeleteVCN(client *TestClient, id string) (string, error) {
	log.Printf("[DEBUG] Delete VCN")
	err := client.DeleteVirtualNetwork(id, nil)
	// FIXME: always returns an error because the VCN jumps directly from State:TERMINATING to http status:404
	return resourceApply(getDeleteFn(err, id, func() (interface{}, error) {
		log.Printf("[DEBUG] Get VCN.State for Delete")
		return client.GetVirtualNetwork(id)
	}))
}

func CreateSubnetWithOptions(client *TestClient, compartmentID, availabilityDomainName, vcnID, cidr string, opts *bm.CreateSubnetOptions) (string, error) {
	log.Printf("[DEBUG] Create Subnet with CIDR:%v", cidr)
	if opts == nil {
		opts = &bm.CreateSubnetOptions{}
	}
	subnet, err := client.CreateSubnet(availabilityDomainName, cidr, compartmentID, vcnID, opts)
	return resourceApply(getCreateFn(err, subnet, bm.ResourceAvailable, func() (interface{}, error) {
		log.Printf("[DEBUG] Get Subnet.State for Create")
		return client.GetSubnet(subnet.ID)
	}))
}

func CreateSubnet(client *TestClient, compartmentID, availabilityDomainName, vcnID string) (string, error) {
	cidr := "172.16.0.0/16"
	return CreateSubnetWithOptions(client, compartmentID, availabilityDomainName, vcnID, cidr, nil)
}

func DeleteSubnet(client *TestClient, subnetID string) (string, error) {
	log.Printf("[DEBUG] Delete Subnet")
	err := client.DeleteSubnet(subnetID, nil)
	// FIXME: always returns an error because the subnet jumps directly from State:TERMINATING to http status:404
	return resourceApply(getDeleteFn(err, subnetID, func() (interface{}, error) {
		log.Printf("[DEBUG] Get Subnet.State for Delete")
		s, err := client.GetSubnet(subnetID)
		if err == nil {
			log.Printf("[WARN] Subnet state: %v", s.State)
		}
		return s, err
	}))
}

func CreateInternetGateway(client *TestClient, compartmentID, vcnID string) (string, error) {
	log.Printf("[DEBUG] Create InternetGateway")
	ig, err := client.CreateInternetGateway(compartmentID, vcnID, true, nil)
	return resourceApply(getCreateFn(err, ig, bm.ResourceAvailable, func() (interface{}, error) {
		return client.GetInternetGateway(ig.ID)
	}))
}

func DeleteInternetGateway(client *TestClient, id string) (string, error) {
	log.Printf("[DEBUG] Delete InternetGateway")
	err := client.DeleteInternetGateway(id, nil)
	return resourceApply(getDeleteFn(err, id, func() (interface{}, error) {
		return client.GetInternetGateway(id)
	}))
}

func CreateSecurityList(client *TestClient, compartmentID, vcnID string) (string, error) {
	log.Printf("[DEBUG] Create SecurityList")
	ingressRules := []bm.IngressSecurityRule{
		{
			Source:   "0.0.0.0/0",
			Protocol: "all",
		},
	}
	egressRules := []bm.EgressSecurityRule{
		{
			Destination: "0.0.0.0/0",
			Protocol:    "all",
		},
	}
	sl, err := client.CreateSecurityList(compartmentID, vcnID, egressRules, ingressRules, nil)
	return resourceApply(getCreateFn(err, sl, bm.ResourceAvailable, func() (interface{}, error) {
		return client.GetSecurityList(sl.ID)
	}))
}

func DeleteSecurityList(client *TestClient, id string) (string, error) {
	log.Printf("[DEBUG] Delete SecurityList")
	err := client.DeleteSecurityList(id, nil)
	return resourceApply(getDeleteFn(err, id, func() (interface{}, error) {
		return client.GetSecurityList(id)
	}))
}

func CreateRouteTable(client *TestClient, compartmentID, vcnID, targetID string) (string, error) {
	log.Printf("[DEBUG] Create RouteTable")
	rules := []bm.RouteRule{
		{
			NetworkEntityID: targetID,
			CidrBlock:       "0.0.0.0/0",
		},
	}
	rt, err := client.CreateRouteTable(compartmentID, vcnID, rules, nil)
	return resourceApply(getCreateFn(err, rt, bm.ResourceAvailable, func() (interface{}, error) {
		return client.GetRouteTable(rt.ID)
	}))
}

func DeleteRouteTable(client *TestClient, id string) (string, error) {
	log.Printf("[DEBUG] Delete RouteTable")
	err := client.DeleteRouteTable(id, nil)
	return resourceApply(getDeleteFn(err, id, func() (interface{}, error) {
		return client.GetRouteTable(id)
	}))
}

func CreateDhcpOption(client *TestClient, compartmentID, vcnID string) (string, error) {
	log.Printf("[DEBUG] Create DHCPOption")
	dhcp := []bm.DHCPDNSOption{
		{
			Type:             "DomainNameServer",
			CustomDNSServers: []string{"202.44.61.9"},
			ServerType:       "CustomDnsServer",
		},
	}
	dhcpOpt, err := client.CreateDHCPOptions(compartmentID, vcnID, dhcp, nil)
	return resourceApply(getCreateFn(err, dhcpOpt, bm.ResourceAvailable, func() (interface{}, error) {
		return client.GetDHCPOptions(dhcpOpt.ID)
	}))
}

func DeleteDhcpOption(client *TestClient, id string) (string, error) {
	log.Printf("[DEBUG] Delete DHCPOption")
	err := client.DeleteDHCPOptions(id, nil)
	return resourceApply(getDeleteFn(err, id, func() (interface{}, error) {
		return client.GetDHCPOptions(id)
	}))
}

func CreateDrg(client *TestClient, compartmentID string) (string, error) {
	log.Printf("[DEBUG] Create DRG")
	drg, err := client.CreateDrg(compartmentID, nil)
	return resourceApply(getCreateFn(err, drg, bm.ResourceAvailable, func() (interface{}, error) {
		return client.GetDrg(drg.ID)
	}))
}

func DeleteDrg(client *TestClient, id string) (string, error) {
	log.Printf("[DEBUG] Delete DRG")
	err := client.DeleteDrg(id, nil)
	return resourceApply(getDeleteFn(err, id, func() (interface{}, error) {
		return client.GetDrg(id)
	}))
}

func CreateInstance(client *TestClient, compartmentID, availabilityDomainName, image, shape, subnetID string) (string, error) {
	log.Printf("[DEBUG] Create Instance")
	opts := &bm.LaunchInstanceOptions{CreateOptions: bm.CreateOptions{DisplayNameOptions: bm.DisplayNameOptions{DisplayName: "instance"}}}
	instance, err := client.LaunchInstance(availabilityDomainName, compartmentID, image, shape, subnetID, opts)
	return resourceApply(getCreateFn(err, instance, bm.ResourceRunning, func() (interface{}, error) {
		return client.GetInstance(instance.ID)
	}))
}

func DeleteInstance(client *TestClient, id string) (string, error) {
	log.Printf("[DEBUG] Delete Instance")
	err := client.TerminateInstance(id, nil)
	return resourceApply(getDeleteFn(err, id, func() (interface{}, error) {
		return client.GetInstance(id)
	}))
}

func CreateVolumeAttachment(client *TestClient, instanceID, volumeID string) (string, error) {
	log.Printf("[DEBUG] Create VolumeAttachment")
	av, err := client.AttachVolume("iscsi", instanceID, volumeID, nil)
	return resourceApply(getCreateFn(err, av, bm.ResourceRunning, func() (interface{}, error) {
		return client.GetVolumeAttachment(av.ID)
	}))
}

func DeleteVolumeAttachment(client *TestClient, id string) (string, error) {
	log.Printf("[DEBUG] Delete VolumeAttachment")
	err := client.DetachVolume(id, nil)
	return resourceApply(getDeleteFn(err, id, func() (interface{}, error) {
		return client.GetVolumeAttachment(id)
	}))
}

func CreatePolicy(client *TestClient, statements []string, compartmentID string) (string, error) {
	log.Printf("[DEBUG] Create Policy")
	unique := RandomText(8)
	name := "test_policy_" + unique
	retry := "retry_tokey_" + unique
	opts := &bm.CreatePolicyOptions{
		RetryTokenOptions: bm.RetryTokenOptions{
			RetryToken: retry,
		},
	}
	policy, err := client.CreatePolicy(name, "desc", compartmentID, statements, opts)
	return resourceApply(getCreateFn(err, policy, bm.ResourceActive, func() (interface{}, error) {
		return client.GetPolicy(policy.ID)
	}))
}

func CreateUser(client *TestClient) (string, error) {
	log.Printf("[DEBUG] Create User")
	uniqueName := RandomText(8)
	testUser := "test_user_" + uniqueName
	retryToken := "retry_token_" + uniqueName
	user, err := client.CreateUser(testUser, "test user", &bm.RetryTokenOptions{RetryToken: retryToken})
	return resourceApply(getCreateFn(err, user, bm.ResourceActive, func() (interface{}, error) {
		return client.GetUser(user.ID)
	}))
}

func CreateGroup(client *TestClient) (string, error) {
	log.Printf("[DEBUG] Create Group")
	uniqueName := RandomText(8)
	tg := "test_group_" + uniqueName
	rt := "retry_token_" + uniqueName
	group, err := client.CreateGroup(tg, "test group", &bm.RetryTokenOptions{RetryToken: rt})
	return resourceApply(getCreateFn(err, group, bm.ResourceActive, func() (interface{}, error) {
		return client.GetGroup(group.ID)
	}))
}

func AddUserToGroup(client *TestClient, uid, gid string) (string, error) {
	log.Printf("[DEBUG] Add User to Group")
	retryToken := "token_" + RandomText(8)
	userGroup, err := client.AddUserToGroup(uid, gid, &bm.RetryTokenOptions{RetryToken: retryToken})
	return resourceApply(getCreateFn(err, userGroup, bm.ResourceActive, func() (interface{}, error) {
		return client.GetUserGroupMembership(userGroup.ID)
	}))
}

func CreateCompartment(client *TestClient) (string, error) {
	log.Printf("[DEBUG] Create Compartment")
	unique := RandomText(8)
	compartmentName := "test_compartment_" + unique
	retryToken := "retry_token_" + unique
	compartment, err := client.CreateCompartment(compartmentName, "test", &bm.RetryTokenOptions{RetryToken: retryToken})
	return resourceApply(getCreateFn(err, compartment, bm.ResourceActive, func() (interface{}, error) {
		return client.GetCompartment(compartment.ID)
	}))
}

// TODO:
func CreateLoadBalancer(client *TestClient, compartmentID string, subnets []string) (string, error) {
	log.Printf("[DEBUG] Create LB")
	unique := RandomText(8)
	displayName := "lb_" + unique
	retryToken := "retry_token_" + unique
	opts := &bm.CreateLoadBalancerOptions{
		DisplayNameOptions: bm.DisplayNameOptions{DisplayName: displayName},
	}
	opts.RetryTokenOptions = bm.RetryTokenOptions{RetryToken: retryToken}
	wr, err := client.CreateLoadBalancer(nil, nil, compartmentID, nil, "100Mbps", subnets, opts)
	return resourceApply(getCreateFn(err, wr, bm.WorkRequestSucceeded, func() (interface{}, error) {
		return client.GetWorkRequest(wr, nil)
	}))
}

func DeleteLoadBalancer(client *TestClient, id string) (string, error) {
	log.Printf("[DEBUG] Delete VCN")
	wr, err := client.DeleteLoadBalancer(id, nil)
	// FIXME: always returns an error because the VCN jumps directly from State:TERMINATING to http status:404
	return resourceApply(getDeleteFn(err, wr, func() (interface{}, error) {
		log.Printf("[DEBUG] Get LB.State for Delete")
		return client.GetLoadBalancer(id, nil)
	}))
}

// FindOrCreateCompartment selects an arbitrary compartment if any exist, or creates a new one if none already exist
func FindOrCreateCompartmentID(client *TestClient) (string, error) {
	// get a compartment, any compartment
	var listOpts bm.ListOptions
	listOpts.Limit = 1
	list, err := client.ListCompartments(&listOpts)
	if err != nil {
		return "", err
	}
	if len(list.Compartments) == 1 {
		return list.Compartments[0].ID, nil
	} else {
		id, err := CreateCompartment(client)
		if err != nil {
			return "", err
		}
		return id, nil
	}
}
