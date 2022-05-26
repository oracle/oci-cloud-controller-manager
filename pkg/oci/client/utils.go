// Copyright 2019 Oracle and/or its affiliates. All rights reserved.
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

package client

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"strings"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-go-sdk/v50/common"
	"go.uber.org/zap"
	"k8s.io/client-go/util/flowcontrol"
)

const (
	// providerName uniquely identifies the Oracle Cloud Infrastructure
	// (OCI) cloud-provider.
	providerName   = "oci"
	providerPrefix = providerName + "://"

	rateLimitQPSDefault    = 20.0
	rateLimitBucketDefault = 5
)

// MapProviderIDToInstanceID parses the provider id and returns the instance ocid.
func MapProviderIDToInstanceID(providerID string) string {
	if strings.HasPrefix(providerID, providerPrefix) {
		return strings.TrimPrefix(providerID, providerPrefix)
	}
	return providerID
}

// IsRetryable returns true if the given error is retriable.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	serviceErr, ok := common.IsServiceError(err)
	return ok && serviceErr.GetHTTPStatusCode() == http.StatusTooManyRequests
}

// NewRateLimiter builds and returns a struct containing read and write
// rate limiters. Defaults are used where no (0) value is provided.
func NewRateLimiter(logger *zap.SugaredLogger, config *providercfg.RateLimiterConfig) RateLimiter {
	if config == nil {
		config = &providercfg.RateLimiterConfig{}
	}

	//If RateLimiter is disabled we would use FakeAlwaysRateLimiter that always accepts the request
	if config.DisableRateLimiter {
		logger.Info("Cloud Provider OCI rateLimiter is disabled")
		return RateLimiter{
			Reader: flowcontrol.NewFakeAlwaysRateLimiter(),
			Writer: flowcontrol.NewFakeAlwaysRateLimiter(),
		}
	}

	// Set to default values if configuration not declared
	if config.RateLimitQPSRead == 0 {
		config.RateLimitQPSRead = rateLimitQPSDefault
	}
	if config.RateLimitBucketRead == 0 {
		config.RateLimitBucketRead = rateLimitBucketDefault
	}
	if config.RateLimitQPSWrite == 0 {
		config.RateLimitQPSWrite = rateLimitQPSDefault
	}
	if config.RateLimitBucketWrite == 0 {
		config.RateLimitBucketWrite = rateLimitBucketDefault
	}

	rateLimiter := RateLimiter{
		Reader: flowcontrol.NewTokenBucketRateLimiter(
			config.RateLimitQPSRead,
			config.RateLimitBucketRead),
		Writer: flowcontrol.NewTokenBucketRateLimiter(
			config.RateLimitQPSWrite,
			config.RateLimitBucketWrite),
	}

	logger.Infof("OCI using read rate limit configuration: QPS=%g, bucket=%d",
		config.RateLimitQPSRead,
		config.RateLimitBucketRead)

	logger.Infof("OCI using write rate limit configuration: QPS=%g, bucket=%d",
		config.RateLimitQPSWrite,
		config.RateLimitBucketWrite)

	return rateLimiter
}

// source: oci-go-sdk common/http.go
func generateRandUUID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	uuid := fmt.Sprintf("%x%x%x%x%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	return uuid, nil
}
