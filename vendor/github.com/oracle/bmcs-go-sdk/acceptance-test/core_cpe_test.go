// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,core recording,all !recording

package acceptance

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MustWin/baremetal-sdk-go/acceptance-test/helpers"
)

func TestCpeCreate(t *testing.T) {
	client := helpers.GetClient("fixtures/core/cpe_create")
	defer client.Stop()

	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err)

	cpe, err := client.CreateCpe(compartmentID, "120.90.41.18", nil)
	require.NoError(t, err)
	defer func() {
		err := client.DeleteCpe(cpe.ID, nil)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, cpe.ID)
}

func TestCpeList(t *testing.T) {
	client := helpers.GetClient("fixtures/core/cpe_list")
	defer client.Stop()

	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err)

	cpe, err := client.CreateCpe(compartmentID, "120.90.41.18", nil)
	require.NoError(t, err)
	defer func() {
		err := client.DeleteCpe(cpe.ID, nil)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, cpe.ID)

	cpes, err := client.ListCpes(compartmentID, nil)
	require.NoError(t, err)
	found := false
	for _, ce := range cpes.Cpes {
		if strings.Compare(ce.ID, cpe.ID) == 0 {
			found = true
		}
	}
	require.True(t, found, "Created cpe not found in list")
}
