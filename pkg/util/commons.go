// Copyright 2020 Oracle and/or its affiliates. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-go-sdk/v65/common"
	metricErrors "github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

const (
	// CompartmentIDAnnotation is the annotation for OCI compartment
	CompartmentIDAnnotation = "oci.oraclecloud.com/compartment-id"

	// Error codes
	Err429             = "429"
	Err4XX             = "4XX"
	Err5XX             = "5XX"
	ErrValidation      = "VALIDATION_ERROR"
	ErrLimitExceeded   = "LIMIT_EXCEEDED"
	ErrCtxTimeout      = "CTX_TIMEOUT"
	ErrTagLimitReached = "TAG_LIMIT_REACHED"
	Success            = "SUCCESS"
	BackupCreating     = "CREATING"

	// Components generating errors
	// Load Balancer
	LoadBalancerType = "LB"
	NSGType          = "NSG"
	// storage types
	CSIStorageType = "CSI"
	FVDStorageType = "FVD"

	// Errorcode prefixes
	SystemTagErrTypePrefix = "SYSTEM_TAG_"
)

// LookupNodeCompartment returns the compartment OCID for the given nodeName.
func LookupNodeCompartment(k kubernetes.Interface, nodeName string) (string, error) {
	node, err := k.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if compartmentID, ok := node.ObjectMeta.Annotations[CompartmentIDAnnotation]; ok {
		if compartmentID != "" {
			return compartmentID, nil
		}
	}
	return "", errors.New("CompartmentID annotation is not present")
}

func GetError(err error) string {
	if err == nil {
		return ""
	}
	err = metricErrors.Cause(err)

	cause := err.Error()
	if cause == "" {
		return ""
	}

	// ErrWaitTimeout is the same var in use by wait.PollUntil in AwaitWorkRequest client method
	if errors.Is(err, wait.ErrWaitTimeout) {
		return ErrCtxTimeout
	}

	re := regexp.MustCompile(`(?i)http status code:\s*(\d+)`)
	if match := re.FindStringSubmatch(cause); match != nil {
		if status, er := strconv.Atoi(match[1]); er == nil {
			if status >= 500 {
				return Err5XX
			} else if status >= 400 {
				if strings.Contains(cause, "LimitExceeded") {
					return ErrLimitExceeded
				}
				if status == 429 {
					return Err429
				}
				return Err4XX
			}
		}
	}
	return ErrValidation
}

func GetMetricDimensionForComponent(err string, component string) string {
	if err == "" || component == "" {
		return ""
	}
	return fmt.Sprintf("%s_%s", component, err)
}

func GetHttpStatusCode(err error) int {
	statusCode := 200
	err = metricErrors.Cause(err)
	if err != nil {
		if serviceErr, ok := err.(common.ServiceError); ok {
			statusCode = serviceErr.GetHTTPStatusCode()
		} else {
			statusCode = 555 // ¯\_(ツ)_/¯
		}
	}
	return statusCode
}

func mergeFreeFormTags(freefromTags ...map[string]string) map[string]string {
	merged := make(map[string]string)
	for _, t := range freefromTags {
		for k, v := range t {
			merged[k] = v
		}
	}
	return merged
}

func mergeDefinedTags(definedTags ...map[string]map[string]interface{}) map[string]map[string]interface{} {
	merged := make(map[string]map[string]interface{})
	for _, t := range definedTags {
		for k, v := range t {
			merged[k] = v
		}
	}
	return merged
}

// MergeTagConfig merges TagConfig's where dstTagConfig takes precedence
func MergeTagConfig(srcTagConfig, dstTagConfig *config.TagConfig) *config.TagConfig {
	var mergedTag config.TagConfig
	mergedTag.FreeformTags = mergeFreeFormTags(srcTagConfig.FreeformTags, dstTagConfig.FreeformTags)
	mergedTag.DefinedTags = mergeDefinedTags(srcTagConfig.DefinedTags, dstTagConfig.DefinedTags)

	return &mergedTag
}

// IsCommonTagPresent return true if Common tags are initialised in config
func IsCommonTagPresent(initialTags *config.InitialTags) bool {

	return initialTags != nil && initialTags.Common != nil
}
