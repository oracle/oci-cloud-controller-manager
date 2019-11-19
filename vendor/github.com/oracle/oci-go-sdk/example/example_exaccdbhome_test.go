// Copyright (c) 2016, 2019, Oracle and/or its affiliates. All rights reserved.
//
// Example code for Database API as it pertains to DbHomes in ExaCC
//
package example

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/database"
	"github.com/oracle/oci-go-sdk/example/helpers"
)

func ExampleExaCCCreateDbHome() {
	client, errors := database.NewDatabaseClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(errors)
	ctx := context.Background()

	rand.Seed(time.Now().Unix())

	vmClusterOcid := os.Getenv("VM_CLUSTER_ID")

	dbName := strings.ToLower(helpers.GetRandomString(8))
	dbUniqueName := dbName + "_" + strings.ToLower(helpers.GetRandomString(20))
	dbVersion := "18.0.0.0"
	adminPassword := "--11##AAbb" + helpers.GetRandomString(14)
	displayName := helpers.GetRandomString(32)

	// create database details
	createDatabaseDetails := database.CreateDatabaseDetails{AdminPassword: &adminPassword, DbName: &dbName, DbUniqueName: &dbUniqueName}

	// create dbhome details
	createDbHomeDetails := database.CreateDbHomeWithVmClusterIdDetails{DisplayName: &displayName, Database: &createDatabaseDetails, VmClusterId: &vmClusterOcid, DbVersion: &dbVersion}

	// create dbome request
	request := database.CreateDbHomeRequest{CreateDbHomeWithDbSystemIdDetails: createDbHomeDetails}

	_, createErrors := client.CreateDbHome(ctx, request)

	helpers.FatalIfError(createErrors)

	fmt.Printf("Create DB Home with vmClusterId completed")

	// Output:
	// Create DB Home with vmClusterId completed
}

func ExampleExaCCGetDbHome() {
	client, errors := database.NewDatabaseClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(errors)
	ctx := context.Background()

	rand.Seed(time.Now().Unix())

	dbHomeId := os.Getenv("DB_HOME_ID")

	// get dbhome request
	request := database.GetDbHomeRequest{DbHomeId: &dbHomeId}

	_, getErrors := client.GetDbHome(ctx, request)

	helpers.FatalIfError(getErrors)

	fmt.Printf("Get DB Home with DbHomeId completed")

	// Output:
	// Get DB Home with DbHomeId completed
}

func ExampleExaCCListDbHome() {
	client, errors := database.NewDatabaseClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(errors)
	ctx := context.Background()

	rand.Seed(time.Now().Unix())

	vmClusterOcid := os.Getenv("VM_CLUSTER_ID")

	// get compartmentId
	getVmClusterRequest := database.GetVmClusterRequest{VmClusterId: &vmClusterOcid}
	getResponse, getErrors := client.GetVmCluster(ctx, getVmClusterRequest)

	helpers.FatalIfError(getErrors)

	compartmentId := *getResponse.CompartmentId

	// list dbome request
	request := database.ListDbHomesRequest{VmClusterId: &vmClusterOcid, CompartmentId: &compartmentId}

	_, listErrors := client.ListDbHomes(ctx, request)

	helpers.FatalIfError(listErrors)

	fmt.Printf("List DB Home with vmClusterId completed")

	// Output:
	// List DB Home with vmClusterId completed
}
