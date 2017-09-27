// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"

	bm "github.com/oracle/bmcs-go-sdk"
)

var compartmentID string
var allCompartments bool

func init() {
	flag.StringVar(&compartmentID, "compartment", "", "compartment ID")
	flag.BoolVar(&allCompartments, "all", false, "all compartments")
}

func main() {
	client := getUnrecordedClient()

	flag.Parse()

	if compartmentID == "" && !allCompartments {
		fmt.Println("no compartment provided. Set with parameter -compartment=ocid1.compartment.oc1...")
		os.Exit(1)
	}

	var compartmentIDs []string
	if allCompartments {
		list, err := client.ListCompartments(nil)
		if err != nil {
			panic(fmt.Sprintf("could not list compartments: %v", err))
		}
		compartmentIDs = make([]string, len(list.Compartments))
		for i := range list.Compartments {
			compartmentIDs[i] = list.Compartments[i].ID
		}
	} else {
		compartmentIDs = []string{compartmentID}
	}

	for _, id := range compartmentIDs {
		err := scrubCompartment(client, id)
		if err != nil {
			fmt.Printf("Failed to clean compartment:%v, reason: %v\n", id, err)
		}
		fmt.Printf("\n")
	}

}

func scrubCompartment(client *bm.Client, compartmentID string) error {
	fmt.Printf("CLEAN compartment:%v\n", compartmentID)
	// TODO:
	// - CPEs
	// - DRGs
	// - Images
	// - IPSecConnections
	// - InternetGateways
	// - RouteGateways
	// - VolumeAttachments
	// - Volumes
	// - Databases
	// - LoadBalancers

	// instance
	is, err := client.ListInstances(compartmentID, nil)
	if err != nil {
		return fmt.Errorf("Failed to list Instances, reason: %v", err)
	}
	terminatedCount := 0
	for _, i := range is.Instances {
		if i.State == bm.ResourceTerminated {
			terminatedCount += 1
		}
	}
	fmt.Printf("- instances: %d (incl %d terminated)\n", len(is.Instances), terminatedCount)
	for _, i := range is.Instances {
		if i.State != bm.ResourceTerminated {
			fmt.Printf("DELETE instance:%v\n", i.ID)
			err := client.TerminateInstance(i.ID, nil)
			if err != nil {
				fmt.Printf("Error deleting instance:%v, %v\n", i.ID, err)
			}
		}
	}

	// TODO: is there anything we can do to clean these?
	// vnic-attachment
	// vas, err := client.ListVnicAttachments(compartmentID, nil)
	// if err != nil {
	// 	return fmt.Errorf("Failed to list VNIC Attachments, reason: %v", err)
	// }
	// fmt.Printf("%d vnic-attachments\n", len(vas.Attachments))
	// for _, va := range vas.Attachments {
	// 	fmt.Printf("vnic-attachment:%v %#v\n", va.ID, va)
	// }

	// scrub virtual-network
	vcns, err := client.ListVirtualNetworks(compartmentID, nil)
	if err != nil {
		return fmt.Errorf("Failed to list Virtual Networks, reason: %v", err)
	}
	fmt.Printf("- virtual-networks: %d\n", len(vcns.VirtualNetworks))
	for _, vcn := range vcns.VirtualNetworks {
		fmt.Printf("CLEAN virtual-network:%v\n", vcn.ID)

		// virtual-network/subnet
		subnets, err := client.ListSubnets(compartmentID, vcn.ID, nil)
		if err != nil {
			return err
		}
		fmt.Printf("-- subnets: %d\n", len(subnets.Subnets))
		for _, sub := range subnets.Subnets {
			fmt.Printf("DELETE subnet:%v\n", sub.ID)
			err := client.DeleteSubnet(sub.ID, nil)
			if err != nil {
				fmt.Printf("Failed to delete subnet:%v, %v\n", sub.ID, err)
			}
		}

		// virtual-network/route-table
		rts, err := client.ListRouteTables(compartmentID, vcn.ID, nil)
		if err != nil {
			return err
		}
		fmt.Printf("-- route-tables: %d (incl 1 default)\n", len(rts.RouteTables))
		for _, rt := range rts.RouteTables {
			if rt.ID != vcn.DefaultRouteTableID {
				fmt.Printf("DELETE route-table:%v\n", rt.ID)
				err := client.DeleteRouteTable(rt.ID, nil)
				if err != nil {
					fmt.Printf("Failed to delete route-table:%v, %v\n", rt.ID, err)
				}
			}
		}
		// virtual-network/security-lists
		ss, err := client.ListSecurityLists(compartmentID, vcn.ID, nil)
		if err != nil {
			return err
		}
		fmt.Printf("-- security-lists: %d (incl 1 default)\n", len(ss.SecurityLists))
		for _, s := range ss.SecurityLists {
			if s.ID != vcn.DefaultSecurityListID {
				fmt.Printf("DELETE security-list:%v\n", s.ID)
				err := client.DeleteSecurityList(s.ID, nil)
				if err != nil {
					fmt.Printf("Failed to delete security-lists:%v, %v\n", s.ID, err)
				}
			}
		}

		// virtual-network
		time.Sleep(10 * time.Second)
		fmt.Printf("DELETE virtual-networks:%v\n", vcn.ID)
		err = client.DeleteVirtualNetwork(vcn.ID, nil)
		if err != nil {
			fmt.Printf("Failed to delete virtual-network:%v, %v\n", vcn.ID, err)
		}
	}
	return nil
}

func getUnrecordedClient() *bm.Client {
	var apiParams = map[string]string{}
	if err := godotenv.Load(); err != nil {
		fmt.Printf("No .env file found, let's hope everything is already in environment variables!\n")
	}

	keys := []string{
		"OBMCS_PRIVATE_KEY_PATH",
		"OBMCS_TENANCY_OCID",
		"OBMCS_USER_OCID",
		"OBMCS_FINGERPRINT",
		"OBMCS_KEY_PASSWORD",
	}

	for _, key := range keys {
		val := os.Getenv(key)
		fmt.Println(key, "=>", val)
		apiParams[key] = val
	}

	c, err := bm.NewClient(
		apiParams["OBMCS_USER_OCID"],
		apiParams["OBMCS_TENANCY_OCID"],
		apiParams["OBMCS_FINGERPRINT"],
		bm.PrivateKeyFilePath(apiParams["OBMCS_PRIVATE_KEY_PATH"]),
		bm.PrivateKeyPassword(apiParams["OBMCS_KEY_PASSWORD"]),
	)

	if err != nil {
		panic(fmt.Sprintf("could not create new client: %v", err))
	}
	return c
}
