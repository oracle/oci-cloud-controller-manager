package types

import (
	"io/ioutil"
	"path"
	"strings"

	"github.com/spf13/viper"
)

// RegionV1 is the data type to display the supported region
type RegionV1 struct {
	Regions string `json:"name" yaml:"name"`
}

// RegionRequestV1 holds the request data for getting the supported region
type RegionRequestV1 struct {
}

// GetSupportedRegion returns supported region
func GetSupportedRegion() (*RegionV1, error) {
	region, err := ioutil.ReadFile(path.Join(viper.GetString("controlplane-creds-dir"), "region")) // just pass the file name
	if err != nil {
		return nil, err
	}

	return &RegionV1{strings.TrimSpace(string(region))}, nil
}
