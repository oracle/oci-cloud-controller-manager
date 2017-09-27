// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package main

import (
	"fmt"
	"log"
	"os"

	baremetal "github.com/oracle/bmcs-go-sdk"
	tt "github.com/oracle/bmcs-go-sdk/test/shared"
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

	client, err := baremetal.NewClient(
		userOCID,
		tenancyOCID,
		fingerprint,
		baremetal.PrivateKeyFilePath(keyPath),
		baremetal.PrivateKeyPassword(password),
	)

	if err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}

	tt.PrintTestHeader("ListShapes")

	var shapes *baremetal.ListShapes
	if shapes, err = client.ListShapes(compartmentID, nil); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	fmt.Printf("%+v\n", shapes)
	tt.PrintTestFooter()

	tt.PrintTestHeader("CreateCpe")

	cpeopts := &baremetal.CreateOptions{}
	cpeopts.DisplayName = "cpe1"
	var cpe *baremetal.Cpe
	if cpe, err = client.CreateCpe(compartmentID, "121.122.123.124", cpeopts); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}

	fmt.Printf("%+v\n", cpe)

	tt.PrintTestFooter()

	tt.PrintTestHeader("ListCpes")
	var cpes *baremetal.ListCpes
	if cpes, err = client.ListCpes(compartmentID, nil); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}

	tt.PrintResults(cpes)

	tt.PrintTestHeader("GetCpe")

	if cpe, err = client.GetCpe(cpe.ID); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}

	tt.PrintResults(cpe)

	tt.PrintTestHeader("DeleteCpe")

	if err = client.DeleteCpe(cpe.ID, nil); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}

	tt.PrintTestFooter()

	tt.PrintTestHeader("CreateVirtualNetwork")

	vcnopts := &baremetal.CreateVcnOptions{}
	vcnopts.DisplayName = "vcn1"
	var vcn *baremetal.VirtualNetwork
	if vcn, err = client.CreateVirtualNetwork("172.16.0.0/16", compartmentID, vcnopts); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	fmt.Printf("%+v\n", vcn)

	tt.PrintTestFooter()
	vcnETag := vcn.ETag
	tt.PrintTestHeader("GetVirtualNetwork")

	if vcn, err = client.GetVirtualNetwork(vcn.ID); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	fmt.Printf("%+v\n", vcn)

	tt.PrintTestFooter()

	tt.PrintTestHeader("DeleteVirtualNetwork")

	if err = client.DeleteVirtualNetwork(vcn.ID, &baremetal.IfMatchOptions{IfMatch: vcnETag}); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}

	tt.PrintTestFooter()

	tt.PrintTestHeader("Instance Feature")
	testInstanceFeature(client)
	tt.PrintTestFooter()

	fmt.Println("PASS")

}

func testInstanceFeature(client *baremetal.Client) {
	var err error

	compartmentID := tt.TestVals["TEST_COMPARTMENT_ID"]

	metadata := map[string]string{
		"ssh_authorized_keys": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDyE8kXMHt8O5i" +
			"zGIesqBy6tSl57c9pc6/3Y/5eE42pNYQ7/jjhgN5b8iXbWV/ZkXWKMchN+6H0uhse27ex6VB1Cl" +
			"7jxXU0hyHQ9yldI5YTvn8HByxi1FLkhHVYnVKe1YY6lgffCCv0EX/VCnZDugjCM4CNS4pGbTU1k" +
			"j21166K7Cr5Jz2XzfGV424Wq3HZNV2GYzcGSzHX4ei9CDzWgQTRQs+6eU3TGUmzQQnF1Gbq/np+L" +
			"wZnoewTeu5ZoxHEG4wiDK+uPiRSp/klNSppt1wVX6qYpdVEl5tRkxmuLhId6TTRksUABNZi+K0JtIk6RvcdR1poEEqihE/AIB91g8UX new mac",
	}

	var ads *baremetal.ListAvailabilityDomains
	if ads, err = client.ListAvailabilityDomains(compartmentID); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	ad := ads.AvailabilityDomains[0]

	var shapesList *baremetal.ListShapes
	if shapesList, err = client.ListShapes(compartmentID, nil); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	shape := shapesList.Shapes[0]

	var vcnList *baremetal.ListVirtualNetworks
	if vcnList, err = client.ListVirtualNetworks(compartmentID, nil); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	vcn := vcnList.VirtualNetworks[0]

	var subnetList *baremetal.ListSubnets
	if subnetList, err = client.ListSubnets(compartmentID, vcn.ID, nil); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	subnet := subnetList.Subnets[0]

	var imgList *baremetal.ListImages
	if imgList, err = client.ListImages(compartmentID, nil); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	img := imgList.Images[0]

	opts := &baremetal.LaunchInstanceOptions{}
	opts.DisplayName = "foobar"
	opts.Metadata = metadata

	instance, err := client.LaunchInstance(
		ad.Name,
		compartmentID,
		img.ID,
		shape.Name,
		subnet.ID,
		opts,
	)

	if err != nil {
		fmt.Println(err)
		os.Exit(tt.ERR)
	}

	fmt.Printf("%+v\n\n", instance)

	tt.PrintTestHeader("ListInstances")

	var instances *baremetal.ListInstances
	if instances, err = client.ListInstances(compartmentID, nil); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	for _, tempInstance := range instances.Instances {
		fmt.Println(tempInstance.DisplayName)
	}
	fmt.Printf("%+v\n", instances)

	tt.PrintTestFooter()

	tt.PrintTestHeader("DeleteInstance")

	if err = client.TerminateInstance(instance.ID, &baremetal.IfMatchOptions{
		IfMatch: instance.ETag,
	}); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}

	tt.PrintTestFooter()

	// instance, err := client.GetInstance("ocid1.instance.oc1.phx.abyhqljsor6xyqrihupbhe3mv7tky3fwyqm5mafnpsf2lihgq3osz4gzr3rq")
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(tt.ERR)
	// }
	// fmt.Printf("%+v\n\n", instance)
	// //
}
