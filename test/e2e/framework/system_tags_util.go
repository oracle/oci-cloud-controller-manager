package framework

import (
	cloudprovider "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci"
)

const (
	okeSystemTagKey = "Cluster"
)

func HasOkeSystemTags(systemTags map[string]map[string]interface{}) bool {
	Logf("actual system tags on the resource: %v", systemTags)
	if systemTags != nil {
		if okeSystemTag, okeSystemTagNsExists := systemTags[cloudprovider.OkeSystemTagNamesapce]; okeSystemTagNsExists {
			if _, okeSystemTagKeyExists := okeSystemTag[okeSystemTagKey]; okeSystemTagKeyExists {
				return true
			}
		}
		return false
	}
	return false
}
