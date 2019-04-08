// Copyright (c) 2016, 2017, Oracle and/or its affiliates. All rights reserved.

// Identity and Access Management Service API
//
// APIs for managing users, groups, compartments, and policies.
//

package integtest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/oracle/oci-go-sdk/audit"
	"github.com/oracle/oci-go-sdk/common"
)

// ListEvents test
func TestAuditClient_ListEvents(t *testing.T) {
	c, clerr := audit.NewAuditClientWithConfigurationProvider(configurationProvider())
	failIfError(t, clerr)

	// list events for last 5 hour
	req := audit.ListEventsRequest{
		CompartmentId: common.String(getCompartmentID()),
		StartTime:     &common.SDKTime{time.Now().Add(time.Hour * -5)},
		EndTime:       &common.SDKTime{time.Now()},
	}

	resp, err := c.ListEvents(context.Background(), req)
	failIfError(t, err)
	assert.NotEmpty(t, resp)
}
