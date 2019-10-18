// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	baremetal "github.com/MustWin/baremetal-sdk-go"
	tt "github.com/MustWin/baremetal-sdk-go/test/shared"
)

func main() {
	log.SetFlags(log.Lshortfile)
	var err error

	keyPath := tt.TestVals["BAREMETAL_PRIVATE_KEY_PATH"]
	tenancyOCID := tt.TestVals["BAREMETAL_TENANCY_OCID"]
	userOCID := tt.TestVals["BAREMETAL_USER_OCID"]
	fingerprint := tt.TestVals["BAREMETAL_FINGERPRINT"]
	password := tt.TestVals["BAREMETAL_KEY_PASSWORD"]
	compartmentID := tt.TestVals["TEST_COMPARTMENT_ID"]

	var client *baremetal.Client
	if client, err = baremetal.NewFromKeyPath(userOCID, tenancyOCID, fingerprint, keyPath, password); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}

	tt.PrintTestHeader("ListDBSupportedOperations")

	var ops *baremetal.ListSupportedOperations
	if ops, err = client.ListSupportedOperations(); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	fmt.Printf("%+v\n", ops)
	tt.PrintTestFooter()

	tt.PrintTestHeader("ListDBVersions")
	var versions *baremetal.ListDBVersions
	if versions, err = client.ListDBVersions(compartmentID, 100, nil); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	version := versions.DBVersions[0]
	fmt.Printf("%+v\n", version)
	tt.PrintTestFooter()

	// tt.PrintTestHeader("ListDBHomes")
	// var homes *baremetal.ListDBHomes
	// if homes, err = client.ListDBHomes(compartmentID, 100, nil); err != nil {
	// 	log.Println(err)
	// 	os.Exit(tt.ERR)
	// }
	// fmt.Printf("%+v\n", homes)
	// tt.PrintTestFooter()

	tt.PrintTestHeader("ListAvailabilityDomains")
	var ads *baremetal.ListAvailabilityDomains
	if ads, err = client.ListAvailabilityDomains(compartmentID); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	ad := ads.AvailabilityDomains[0]
	fmt.Printf("%+v\n", ads)
	fmt.Printf("%+v\n", ad)
	tt.PrintTestFooter()

	tt.PrintTestHeader("ListVirtualNetworks")
	var vcnList *baremetal.ListVirtualNetworks
	if vcnList, err = client.ListVirtualNetworks(compartmentID, nil); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	vcn := vcnList.VirtualNetworks[0]
	fmt.Printf("%+v\n", vcnList)
	fmt.Printf("%+v\n", vcn)
	tt.PrintTestFooter()

	tt.PrintTestHeader("ListSubnets")
	var subnetList *baremetal.ListSubnets
	if subnetList, err = client.ListSubnets(compartmentID, vcn.ID, nil); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	subnet := subnetList.Subnets[0]
	fmt.Printf("%+v\n", subnetList)
	fmt.Printf("%+v\n", subnet)
	tt.PrintTestFooter()

	tt.PrintTestHeader("ListDBSystemShapes")
	var shapes *baremetal.ListDBSystemShapes
	if shapes, err = client.ListDBSystemShapes(ad.Name, compartmentID, 1, nil); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	fmt.Printf("%+v\n", shapes)
	shape := shapes.DBSystemShapes[0]
	fmt.Printf("%+v\n", shape)
	tt.PrintTestFooter()

	tt.PrintTestHeader("DBSystem")
	dbHome := baremetal.NewCreateDBHomeDetails("ABab_#789", "dbname", version.Version, nil)
	tt.PrintTestHeader("DBSystem:DBHome")
	fmt.Printf("%+v\n", dbHome)
	opts := &baremetal.LaunchDBSystemOptions{
		DatabaseEdition: baremetal.DatabaseEditionStandard,
		DBHome:          dbHome,
		Domain:          "test-db-domain",
		Hostname:        "test-db-system-hostname",
	}
	var sys *baremetal.DBSystem
	// TODO: shape.Name should be used instead of BM.DenseIO1.36, but
	// DBSystemShapes is only returning one shape.
	if sys, err = client.LaunchDBSystem(
		ad.Name, compartmentID, "BM.DenseIO1.36", subnet.ID,
		[]string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDWm/fWAxfDy2DxlJLvIRubenc/aO77QaoSHXotCAxCkgttaxv+YNGyJIxO1hGDbmxwlBfyYivHCAg+LMBX6vrp8esA5B3Gnd9kLcvnfazGFvCmGJAecoZFwGvGJb5UeFZI6jCmELp/QbAx7wL2iOvCB+HY3K18sVft0kk4vd/p9iXiXrDPBytdZcYtR6hBU8pal6+FR1o0UlGbK8vvTi3r57IJ/U+DMs1wRHYvIEBWGoCBeuCXqL5PQU+HxGp1SwmicQGZbXS4x1XW1Hvc4pudvfoC0YOVXmVYIE3ZgWYzyid/IPm/JEs9wlbPN1zoCbHQjxKd7o15B2nSvDj0/gTT gotascii@gmail.com"},
		2, opts,
	); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	fmt.Printf("%+v\n", sys)
	tt.PrintTestFooter()

	tt.PrintTestHeader("TerminateDBSystem")
	time.Sleep(10 * time.Second)
	if err = client.TerminateDBSystem(sys.ID, nil); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	tt.PrintTestFooter()

	/*tt.PrintTestHeader("ListDBNodes")
	dbSystemID := "system-id-filled-in-from-above" // TODO: <---
	var nodes *baremetal.ListDBNodes
	if nodes, err = client.ListDBNodes(compartmentID, dbSystemID, 100, nil); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	fmt.Printf("%+v\n", nodes)
	tt.PrintTestFooter()


	tt.PrintTestHeader("ListDBHomes")

	var homes *baremetal.ListDBHomes
	if homes, err = client.ListDBHomes(compartmentID, dbSystemID, 100, nil); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	fmt.Printf("%+v\n", homes)
	tt.PrintTestFooter()


	*/

	fmt.Println("PASS")
}
