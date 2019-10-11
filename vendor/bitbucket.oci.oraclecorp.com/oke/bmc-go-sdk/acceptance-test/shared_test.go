// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package acceptance

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/dnaeon/go-vcr/recorder"

	bm "github.com/MustWin/baremetal-sdk-go"
)

type Runmode string

const (
	RunmodeRecord Runmode = "RECORD"
	RunmodeReplay Runmode = "REPLAY"
)

type stopper interface {
	Stop() error
}

type testClient struct {
	*bm.Client
	recorder *recorder.Recorder
}

func (tc *testClient) Stop() error {
	return tc.recorder.Stop()
}

type testTransport struct {
	requestCount  int
	realTransport http.RoundTripper
	mtx           sync.Mutex
}

func newTestTransport(t http.RoundTripper) http.RoundTripper {
	tt := &testTransport{
		realTransport: http.DefaultTransport,
	}
	if t != nil {
		tt.realTransport = t
	}
	return tt
}

func (tt *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Request-Number", strconv.Itoa(tt.requestCount))
	tt.mtx.Lock()
	tt.requestCount++
	tt.mtx.Unlock()
	return tt.realTransport.RoundTrip(req)
}

var validNameChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.-_"

func randomText(keySize int) string {
	key := make([]byte, keySize)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Read(key)
	for i := range key {
		key[i] = validNameChars[key[i]%byte(len(validNameChars))]
	}
	return string(key)
}

func boolPtr(v bool) *bool {
	return &v
}

const waitDuration = 2 * time.Minute

type resourceCommandResult struct {
	id  string
	err error
}

var errResourceCommandTimeout = errors.New("timed out waiting to apply command to resource")

// commandFunc takes a channel; when called should push a resourceCreateResponse onto the channel and return whether the finished executing
type commandFunc func(c chan<- resourceCommandResult) (finished bool)

// resourceApply takes a commandFunc; executes it; and returns an id.
// If the command fails to execute, returns an error.
func resourceApply(f commandFunc) (id string, err error) {
	idChan := make(chan resourceCommandResult)
	go func(c chan<- resourceCommandResult) {
		for {
			if finished := f(idChan); finished {
				return
			}
			if RUNMODE == RunmodeRecord {
				time.Sleep(3 * time.Second)
			}
		}
	}(idChan)
	for {
		select {
		case r := <-idChan:
			return r.id, r.err
		case <-time.After(waitDuration):
			return "", errResourceCommandTimeout
		}
	}
}

func getCreatedState(resource interface{}) string {
	v := reflect.ValueOf(resource).Elem().FieldByName("State")

	if !v.IsValid() {
		panic("Developer error, resource passed to getCreatedState does not have a State field")
	}

	return v.Interface().(string)
}

func getID(resource interface{}) string {
	v := reflect.ValueOf(resource)
	if !v.IsValid() {
		panic("Developer error, resource passed to getID is not valid")
	}
	if v.IsNil() {
		return ""
	}
	id := v.Elem().FieldByName("ID")

	return id.Interface().(string)
}

func isResourceGoneOrTerminated(resource interface{}, err error) (bool, error) {
	if bmError, ok := err.(*bm.Error); ok {
		if bmError.Code == bm.NotAuthorizedOrNotFound {
			return true, nil
		}
	}
	if err != nil {
		return false, err
	}

	v := reflect.ValueOf(resource).Elem().FieldByName("State")

	if !v.IsValid() {
		panic("Developer error, resource passed to isResourceGoneOrTerminated does not have a state field")
	}

	state := v.Interface().(string)

	if state == bm.ResourceTerminated {
		return true, nil
	}
	return false, nil
}

func getCreateFn(err error, resource interface{}, completedState string, resourceFetchFn func() (interface{}, error)) commandFunc {
	id := getID(resource) // This wont be used if err != nil
	return func(c chan<- resourceCommandResult) bool {
		fmt.Println("-----------------------------")
		if err != nil {
			c <- resourceCommandResult{"", err}
			return true
		}
		fresh, err := resourceFetchFn()
		if err != nil {
			c <- resourceCommandResult{"", err}
			return true
		}
		if getCreatedState(fresh) == completedState {
			c <- resourceCommandResult{id, nil}
			return true
		}
		return false
	}
}

// getDeleteFn takes an error, an id, and a func to fetch the resource; returns a createFunc.
// the createFunc will fetch the resource and check if it is gone, returning
func getDeleteFn(err error, id string, resourceFetchFn func() (interface{}, error)) commandFunc {
	return func(c chan<- resourceCommandResult) bool {
		if err != nil {
			c <- resourceCommandResult{"", err}
			return true
		}
		res, erra := resourceFetchFn()
		if erra != nil {
			c <- resourceCommandResult{"", erra}
			return true
		}
		gone, errb := isResourceGoneOrTerminated(res, erra)
		if errb != nil {
			c <- resourceCommandResult{"", errb}
			return true
		}
		if gone {
			c <- resourceCommandResult{id, nil}
			return true
		}
		return false
	}
}

func createVCN(client *testClient, cidr, compartmentID string) commandFunc {
	log.Printf("[DEBUG] Create VCN with CIDR:%v", cidr)
	unique := randomText(8)
	displayName := "vcn_" + unique
	retryToken := "retry_token_" + unique
	opts := &bm.CreateVcnOptions{
		CreateOptions: bm.CreateOptions{
			DisplayNameOptions: bm.DisplayNameOptions{DisplayName: displayName},
			RetryTokenOptions:  bm.RetryTokenOptions{RetryToken: retryToken},
		},
	}
	vcn, err := client.CreateVirtualNetwork(cidr, compartmentID, opts)
	return getCreateFn(err, vcn, bm.ResourceAvailable, func() (interface{}, error) {
		return client.GetVirtualNetwork(vcn.ID)
	})
}

func deleteVCN(client *testClient, id string) commandFunc {
	log.Printf("[DEBUG] Delete VCN")
	err := client.DeleteVirtualNetwork(id, nil)
	// FIXME: always returns an error because the VCN jumps directly from State:TERMINATING to http status:404
	return getDeleteFn(err, id, func() (interface{}, error) {
		log.Printf("[DEBUG] Get VCN.State for Delete")
		return client.GetVirtualNetwork(id)
	})
}

func createSubnetWithCIDR(client *testClient, compartmentID, availabilityDomainName, vcnID, cidr string) commandFunc {
	log.Printf("[DEBUG] Create Subnet with CIDR:%v", cidr)
	subOpts := &bm.CreateSubnetOptions{}
	subnet, err := client.CreateSubnet(availabilityDomainName, cidr, compartmentID, vcnID, subOpts)
	return getCreateFn(err, subnet, bm.ResourceAvailable, func() (interface{}, error) {
		log.Printf("[DEBUG] Get Subnet.State for Create")
		return client.GetSubnet(subnet.ID)
	})
}

func createSubnet(client *testClient, compartmentID, availabilityDomainName, vcnID string) commandFunc {
	cidr := "172.16.0.0/16"
	return createSubnetWithCIDR(client, compartmentID, availabilityDomainName, vcnID, cidr)
}

func deleteSubnet(client *testClient, subnetID string) commandFunc {
	log.Printf("[DEBUG] Delete Subnet")
	err := client.DeleteSubnet(subnetID, nil)
	// FIXME: always returns an error because the subnet jumps directly from State:TERMINATING to http status:404
	return getDeleteFn(err, subnetID, func() (interface{}, error) {
		log.Printf("[DEBUG] Get Subnet.State for Delete")
		return client.GetSubnet(subnetID)
	})
}

func createInternetGateway(client *testClient, compartmentID, vcnID string) commandFunc {
	log.Printf("[DEBUG] Create InternetGateway")
	ig, err := client.CreateInternetGateway(compartmentID, vcnID, true, nil)
	return getCreateFn(err, ig, bm.ResourceAvailable, func() (interface{}, error) {
		return client.GetInternetGateway(ig.ID)
	})
}

func deleteInternetGateway(client *testClient, id string) commandFunc {
	log.Printf("[DEBUG] Delete InternetGateway")
	err := client.DeleteInternetGateway(id, nil)
	return getDeleteFn(err, id, func() (interface{}, error) {
		return client.GetInternetGateway(id)
	})
}

func createSecurityList(client *testClient, compartmentID, vcnID string) commandFunc {
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
	return getCreateFn(err, sl, bm.ResourceAvailable, func() (interface{}, error) {
		return client.GetSecurityList(sl.ID)
	})
}

func deleteSecurityList(client *testClient, id string) commandFunc {
	log.Printf("[DEBUG] Delete SecurityList")
	err := client.DeleteSecurityList(id, nil)
	return getDeleteFn(err, id, func() (interface{}, error) {
		return client.GetSecurityList(id)
	})
}

func createRouteTable(client *testClient, compartmentID, vcnID, targetID string) commandFunc {
	log.Printf("[DEBUG] Create RouteTable")
	rules := []bm.RouteRule{
		{
			NetworkEntityID: targetID,
			CidrBlock:       "0.0.0.0/0",
		},
	}
	rt, err := client.CreateRouteTable(compartmentID, vcnID, rules, nil)
	return getCreateFn(err, rt, bm.ResourceAvailable, func() (interface{}, error) {
		return client.GetRouteTable(rt.ID)
	})
}

func deleteRouteTable(client *testClient, id string) commandFunc {
	log.Printf("[DEBUG] Delete RouteTable")
	err := client.DeleteRouteTable(id, nil)
	return getDeleteFn(err, id, func() (interface{}, error) {
		return client.GetRouteTable(id)
	})
}

func createDhcpOption(client *testClient, compartmentID, vcnID string) commandFunc {
	log.Printf("[DEBUG] Create DHCPOption")
	dhcp := []bm.DHCPDNSOption{
		{
			Type:             "DomainNameServer",
			CustomDNSServers: []string{"202.44.61.9"},
			ServerType:       "CustomDnsServer",
		},
	}
	dhcpOpt, err := client.CreateDHCPOptions(compartmentID, vcnID, dhcp, nil)
	return getCreateFn(err, dhcpOpt, bm.ResourceAvailable, func() (interface{}, error) {
		return client.GetDHCPOptions(dhcpOpt.ID)
	})
}

func deleteDhcpOption(client *testClient, id string) commandFunc {
	log.Printf("[DEBUG] Delete DHCPOption")
	err := client.DeleteDHCPOptions(id, nil)
	return getDeleteFn(err, id, func() (interface{}, error) {
		return client.GetDHCPOptions(id)
	})
}

func createDrg(client *testClient, compartmentID string) commandFunc {
	log.Printf("[DEBUG] Create DRG")
	drg, err := client.CreateDrg(compartmentID, nil)
	return getCreateFn(err, drg, bm.ResourceAvailable, func() (interface{}, error) {
		return client.GetDrg(drg.ID)
	})
}

func deleteDrg(client *testClient, id string) commandFunc {
	log.Printf("[DEBUG] Delete DRG")
	err := client.DeleteDrg(id, nil)
	return getDeleteFn(err, id, func() (interface{}, error) {
		return client.GetDrg(id)
	})
}

func createInstance(client *testClient, compartmentID, availabilityDomainName, image, shape, subnetID string) commandFunc {
	log.Printf("[DEBUG] Create Instance")
	opts := &bm.LaunchInstanceOptions{CreateOptions: bm.CreateOptions{DisplayNameOptions: bm.DisplayNameOptions{DisplayName: "instance"}}}
	instance, err := client.LaunchInstance(availabilityDomainName, compartmentID, image, shape, subnetID, opts)
	return getCreateFn(err, instance, bm.ResourceRunning, func() (interface{}, error) {
		return client.GetInstance(instance.ID)
	})
}

func deleteInstance(client *testClient, id string) commandFunc {
	log.Printf("[DEBUG] Delete Instance")
	err := client.TerminateInstance(id, nil)
	return getDeleteFn(err, id, func() (interface{}, error) {
		return client.GetInstance(id)
	})
}

func createPolicy(client *testClient, statements []string, compartmentID string) commandFunc {
	log.Printf("[DEBUG] Create Policy")
	unique := randomText(8)
	name := "test_policy_" + unique
	retry := "retry_tokey_" + unique
	opts := &bm.CreatePolicyOptions{
		RetryTokenOptions: bm.RetryTokenOptions{
			RetryToken: retry,
		},
	}
	policy, err := client.CreatePolicy(name, "desc", compartmentID, statements, opts)
	return getCreateFn(err, policy, bm.ResourceActive, func() (interface{}, error) {
		return client.GetPolicy(policy.ID)
	})
}

func createUser(client *testClient) commandFunc {
	log.Printf("[DEBUG] Create User")
	uniqueName := randomText(8)
	testUser := "test_user_" + uniqueName
	retryToken := "retry_token_" + uniqueName
	user, err := client.CreateUser(testUser, "test user", &bm.RetryTokenOptions{RetryToken: retryToken})
	return getCreateFn(err, user, bm.ResourceActive, func() (interface{}, error) {
		return client.GetUser(user.ID)
	})
}

func createGroup(client *testClient) commandFunc {
	log.Printf("[DEBUG] Create Group")
	uniqueName := randomText(8)
	tg := "test_group_" + uniqueName
	rt := "retry_token_" + uniqueName
	group, err := client.CreateGroup(tg, "test group", &bm.RetryTokenOptions{RetryToken: rt})
	return getCreateFn(err, group, bm.ResourceActive, func() (interface{}, error) {
		return client.GetGroup(group.ID)
	})
}

func addUserToGroup(client *testClient, uid, gid string) commandFunc {
	log.Printf("[DEBUG] Add User to Group")
	retryToken := "token_" + randomText(8)
	userGroup, err := client.AddUserToGroup(uid, gid, &bm.RetryTokenOptions{RetryToken: retryToken})
	return getCreateFn(err, userGroup, bm.ResourceActive, func() (interface{}, error) {
		return client.GetUserGroupMembership(userGroup.ID)
	})
}

func createCompartment(client *testClient) commandFunc {
	log.Printf("[DEBUG] Create Compartment")
	unique := randomText(8)
	compartmentName := "test_compartment_" + unique
	retryToken := "retry_token_" + unique
	compartment, err := client.CreateCompartment(compartmentName, "test", &bm.RetryTokenOptions{RetryToken: retryToken})
	return getCreateFn(err, compartment, bm.ResourceActive, func() (interface{}, error) {
		return client.GetCompartment(compartment.ID)
	})
}
