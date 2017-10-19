package util

import (
	"errors"
	"fmt"
	"testing"
)

func mockInstanceMetadataJSON() string {
	return `{
			"availabilityDomain" : "NWuj:PHX-AD-1",
  			"compartmentId" : "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq",
  			"displayName" : "trjl-kb8s-master",
  			"id" : "ocid1.instance.oc1.phx.abyhqljtj775udgtbu7nddt6j2hqgxdsgrnpweepogvvsmqfppefewile5zq",
  			"image" : "ocid1.image.oc1.phx.aaaaaaaaamx6ta37uxltor6n5lxfgd5lkb3lwmoqurlpn2x4dz5ockekiuea",
  			"metadata" : {
    			"ssh_authorized_keys" : "ssh-rsa some-key-data tlangfor@tlangfor-mac\n"
  			},
  			"region" : "phx",
  			"shape" : "VM.Standard1.1",
  			"state" : "Provisioning",
  			"timeCreated" : 1496415602152
			}`
}

func mockInstanceMetadata() *InstanceMetadata {
	return &InstanceMetadata{
		InstanceOCID:       "ocid1.instance.oc1.phx.abyhqljtj775udgtbu7nddt6j2hqgxdsgrnpweepogvvsmqfppefewile5zq",
		CompartmentOCID:    "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq",
		AvailabilityDomain: "NWuj:PHX-AD-1",
		Region:             "phx",
	}
}
func mockInstanceMetadataFromAPIFailure() (*InstanceMetadata, error) {
	return nil, fmt.Errorf("failed to http get node api metadata: %s",
		errors.New("Get http://169.254.169.254/opc/v1/instance/: dial tcp 169.254.169.254:80: i/o timeout"))
}

func TestUnmarshallInstanceMetadataJSON(t *testing.T) {
	mockJSON := mockInstanceMetadataJSON()
	metadata, err := unmarshallInstanceMetadataJSON([]byte(mockJSON))
	expected := mockInstanceMetadata()

	if err != nil {
		t.Fatalf("Got unexpected error %s", err)
	}
	if metadata.InstanceOCID != expected.InstanceOCID {
		t.Fatalf("%v != %v", metadata.InstanceOCID, expected.InstanceOCID)
	}
	if metadata.CompartmentOCID != expected.CompartmentOCID {
		t.Fatalf("%v != %v", metadata.CompartmentOCID, expected.CompartmentOCID)
	}
	if metadata.AvailabilityDomain != expected.AvailabilityDomain {
		t.Fatalf("%v != %v", metadata.AvailabilityDomain, expected.AvailabilityDomain)
	}
	if metadata.Region != expected.Region {
		t.Fatalf("%v != %v", metadata.Region, expected.Region)
	}
}
