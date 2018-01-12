package acceptance

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bm "github.com/oracle/bmcs-go-sdk"
	"github.com/oracle/bmcs-go-sdk/acceptance-test/helpers"
)

func TestDatabaseCRUD(t *testing.T) {
	// Arrange
	client := helpers.GetClient("fixtures/database")
	defer client.Stop()
	// Get compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err, "Setup Compartment")
	// Get Availability Domain
	ads, err := client.ListAvailabilityDomains(compartmentID)
	require.NoError(t, err, "Setup AvailabilityDomains")
	availabilityDomainName := ads.AvailabilityDomains[0].Name
	// Create VCN (DnsLabel is required for DB Systems)
	vcnOpts := &bm.CreateVcnOptions{
		DnsLabel: "database",
	}
	vcnID, err := helpers.CreateVCNWithOptions(client, "172.16.0.0/16", compartmentID, vcnOpts)
	require.NoError(t, err, "Setup VCN")
	require.NotEmpty(t, vcnID, "Setup VCN: ID")
	defer func() {
		_, err = helpers.DeleteVCN(client, vcnID)
		assert.NoError(t, err, "Teardown VCN")
	}()
	// Create Subnet
	subnetOpts := &bm.CreateSubnetOptions{
		DNSLabel: "test",
	}
	subnetID, err := helpers.CreateSubnetWithOptions(client, compartmentID, availabilityDomainName, vcnID, "172.16.0.0/16", subnetOpts)
	require.NoError(t, err, "Setup Subnet")
	require.NotEmpty(t, subnetID, "Setup Subnet: ID")
	defer func() {
		_, err = helpers.DeleteSubnet(client, subnetID)
		assert.NoError(t, err, "Teardown Subnet")
		helpers.Sleep(2 * time.Second)
	}()

	versions, err := client.ListDBVersions(compartmentID, nil)
	assert.NoError(t, err, "ListDBVersions")
	version := versions.DBVersions[1]

	shapes, err := client.ListDBSystemShapes(availabilityDomainName, compartmentID, nil)
	assert.NoError(t, err, "ListDBSystemShapes")
	assert.NotEmpty(t, shapes, "ListDBSystemShapes")

	dbHomeOpts := &bm.CreateDBHomeOptions{
		DisplayNameOptions: bm.DisplayNameOptions{DisplayName: "dbHomeDisplayName"},
	}
	dbOpts := &bm.CreateDatabaseOptions{
		CharacterSet:  "AL32UTF8",
		NCharacterSet: "AL16UTF16",
		DBWorkload:    "OLTP",
		PDBName:       "pdbName",
	}
	db := bm.NewCreateDatabaseDetails("ABab_#789", "dbname", dbOpts)
	dbHome := bm.NewCreateDBHomeDetails(db, version.Version, dbHomeOpts)
	opts := &bm.LaunchDBSystemOptions{
		Domain:                     "testDBDomain",
		DataStoragePercentage:      80,
		DiskRedundancy:             "HIGH",
		InitialDataStorageSizeInGB: 256,
		LicenseModel:               bm.BringYourOwnLicense,
		NodeCount:                  1,
	}
	opts.DisplayName = "dbDisplayName"
	sys, err := client.LaunchDBSystem(
		availabilityDomainName,
		compartmentID,
		2, // this parameter is not used because the core count is inferred from the shape
		bm.DatabaseEditionStandard,
		dbHome,
		"test-db-system-hostname",
		"VM.Standard1.2",
		[]string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDWm/fWAxfDy2DxlJLvIRubenc/aO77QaoSHXotCAxCkgttaxv+YNGyJIxO1hGDbmxwlBfyYivHCAg+LMBX6vrp8esA5B3Gnd9kLcvnfazGFvCmGJAecoZFwGvGJb5UeFZI6jCmELp/QbAx7wL2iOvCB+HY3K18sVft0kk4vd/p9iXiXrDPBytdZcYtR6hBU8pal6+FR1o0UlGbK8vvTi3r57IJ/U+DMs1wRHYvIEBWGoCBeuCXqL5PQU+HxGp1SwmicQGZbXS4x1XW1Hvc4pudvfoC0YOVXmVYIE3ZgWYzyid/IPm/JEs9wlbPN1zoCbHQjxKd7o15B2nSvDj0/gTT gotascii@gmail.com"},
		subnetID,
		opts,
	)
	assert.NoError(t, err, "LaunchDBSystem")
	dbSystemID := sys.ID

	sys, err = client.GetDBSystem(dbSystemID)
	assert.NoError(t, err, "GetDBSystem")
	assert.Equal(t, dbSystemID, sys.ID)
	assert.Equal(t, bm.BringYourOwnLicense, sys.LicenseModel)

	systems, err := client.ListDBSystems(compartmentID, nil)
	assert.NoError(t, err, "ListDBSystems")
	found := false
	for _, system := range systems.DBSystems {
		if strings.Compare(dbSystemID, system.ID) == 0 {
			found = true
		}
	}
	assert.True(t, found, "ListDBSystems: Launched DBSystem not found")

	nodes, err := client.ListDBNodes(compartmentID, dbSystemID, nil)
	assert.NoError(t, err, "ListDBNodes")
	assert.NotEmpty(t, nodes, "ListDBNodes")
	for _, node := range nodes.DBNodes {
		assert.Equal(t, dbSystemID, node.DBSystemID, "ListDBNodes")
	}

	homes, err := client.ListDBHomes(compartmentID, dbSystemID, nil)
	assert.NoError(t, err, "ListDBHomes")
	assert.NotEmpty(t, homes, "ListDBHomes")
	for _, home := range homes.DBHomes {
		assert.Equal(t, dbSystemID, home.DBSystemID, "ListDBHomes")
	}

	err = client.TerminateDBSystem(dbSystemID, nil)
	assert.NoError(t, err, "TerminateDBSystem")
}
