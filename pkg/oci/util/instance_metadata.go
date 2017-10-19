package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	instanceMetaDataAPI      = "http://169.254.169.254"
	instanceMetaDataEndpoint = instanceMetaDataAPI + "/opc/v1/instance/"
)

// InstanceMetadata holds OCI API details about the node the driver is executing
// on.
type InstanceMetadata struct {
	InstanceOCID       string `json:"id"`
	CompartmentOCID    string `json:"compartmentId"`
	AvailabilityDomain string `json:"availabilityDomain"`
	Region             string `json:"region"`
}

// InstanceMetadataService defines a capability to obtain the OCI InstanceMetadata.
type InstanceMetadataService func() (*InstanceMetadata, error)

func httpGetInstanceMetadataJSONFromAPI() ([]byte, error) {
	rs, err := http.Get(instanceMetaDataEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to http get node api metadata: %s", err)

	}
	defer rs.Body.Close()

	body, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read node api metadata: %s", err)
	}

	return body, nil
}

func unmarshallInstanceMetadataJSON(metadataJSON []byte) (*InstanceMetadata, error) {
	metadata := &InstanceMetadata{}
	err := json.Unmarshal([]byte(metadataJSON), metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal node api metadata: %s", err)
	}
	return metadata, nil
}

// InstanceMetadataFromAPI is a concrete InstanceMetadataService implementation.
// An http-get metadata api based method to get InstanceMetadata.
func InstanceMetadataFromAPI() (*InstanceMetadata, error) {
	metadataJSON, err := httpGetInstanceMetadataJSONFromAPI()
	if err != nil {
		return nil, err
	}
	metadata, err := unmarshallInstanceMetadataJSON(metadataJSON)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}
