// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package acceptance

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"time"

	bm "github.com/oracle/bmcs-go-sdk"
	"github.com/oracle/bmcs-go-sdk/acceptance-test/helpers"
)

func TestSecurityListCRUD(t *testing.T) {
	client := helpers.GetClient("fixtures/security_list")
	defer client.Stop()
	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err, "Setup Compartment")
	vcnID, err := helpers.CreateVCN(client, "172.16.0.0/16", compartmentID)
	require.NoError(t, err, "Setup VCN")
	require.NotEmpty(t, vcnID, "Setup VCN: ID")
	defer func() {
		_, err = helpers.DeleteVCN(client, vcnID)
		assert.NoError(t, err, "Teardown VCN")
	}()

	// Create
	ingressRules := []bm.IngressSecurityRule{
		{
			Source:   "0.0.0.0/0",
			Protocol: "6", // TCP
			TCPOptions: &bm.TCPOptions{
				DestinationPortRange: &bm.PortRange{
					Min: 1,
					Max: 2,
				},
			},
		},
	}
	egressRules := []bm.EgressSecurityRule{
		{
			Destination: "0.0.0.0/0",
			Protocol:    "17", // UDP
			UDPOptions: &bm.UDPOptions{
				DestinationPortRange: &bm.PortRange{
					Min: 1,
					Max: 2,
				},
				SourcePortRange: &bm.PortRange{
					Min: 3,
					Max: 4,
				},
			},
		},
	}
	securityList, err := client.CreateSecurityList(compartmentID, vcnID, egressRules, ingressRules, nil)
	assert.NoError(t, err, "Create")
	assert.NotEmpty(t, securityList.ID, "Create: ID")

	// Get
	startTime := time.Now()
	for {
		securityList, _ := client.GetSecurityList(securityList.ID)
		if securityList.State == bm.ResourceAvailable {
			break
		}
		helpers.Sleep(2 * time.Second)
		if time.Now().Sub(startTime) > 5*time.Minute {
			assert.FailNow(t, "Timeout while waiting for Security List provisioning.")
		}
	}

	assert.Equal(t, bm.ResourceAvailable, securityList.State, "Get: State")
	assert.Equal(t, 1, len(securityList.EgressSecurityRules))
	assert.Equal(t, uint64(2), securityList.EgressSecurityRules[0].UDPOptions.DestinationPortRange.Max)
	assert.Equal(t, uint64(3), securityList.EgressSecurityRules[0].UDPOptions.SourcePortRange.Min)
	assert.Equal(t, 1, len(securityList.IngressSecurityRules))
	assert.Equal(t, uint64(1), securityList.IngressSecurityRules[0].TCPOptions.DestinationPortRange.Min)
	assert.Equal(t, "6", securityList.IngressSecurityRules[0].Protocol)

	// List
	sls, err := client.ListSecurityLists(compartmentID, vcnID, nil)
	assert.NoError(t, err, "List")
	found := false
	for _, sl := range sls.SecurityLists {
		if strings.Compare(securityList.ID, sl.ID) == 0 {
			found = true
		}
	}
	assert.True(t, found, "List: Created SecurityList not found")

	// Delete
	_, err = helpers.DeleteSecurityList(client, securityList.ID)
	assert.NoError(t, err, "Delete")
}
