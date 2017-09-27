// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package acceptance

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oracle/bmcs-go-sdk/acceptance-test/helpers"
)

func TestListAvailabilityDomains(t *testing.T) {
	client := helpers.GetClient("fixtures/availability_domain")
	defer client.Stop()
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err)

	// List
	list, err := client.ListAvailabilityDomains(compartmentID)
	require.NoError(t, err)
	require.NotNil(t, list)
	assert.Len(t, list.AvailabilityDomains, 3, "List")
}
