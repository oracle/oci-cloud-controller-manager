// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
//
// Example code for Core Services API
//

package example

import (
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-go-sdk/example/helpers"
)

const (
	vnicId      = "" // REQUIRED: VNIC ID on which the IPv6 attachment will be created
	displayName = "OCI-GOSDK-CreateIpv6-Example"
)

func ExampleCreateIpv6() {
	client, err := core.NewVirtualNetworkClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(err)
	ctx := context.Background()

	// create the request
	request := core.CreateIpv6Request{}
	request.VnicId = common.String(vnicId)
	request.DisplayName = common.String(displayName)
	request.RequestMetadata = helpers.GetRequestMetadataWithDefaultRetryPolicy()

	_, err = client.CreateIpv6(ctx, request)
	helpers.FatalIfError(err)

	fmt.Println("IPv6 VNIC attachment created")

	// Output:
	// IPv6 VNIC attachment created
}
