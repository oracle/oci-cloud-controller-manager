// Copyright 2017 Oracle and/or its affiliates. All rights reserved.
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

package oci

import (
	"testing"
)

func TestBuildRateLimiterWithConfig(t *testing.T) {
	var qpsRead float32 = 6.0
	bucketRead := 20
	var qpsWrite float32 = 8.0
	bucketWrite := 20

	rateLimiterConfig := &RateLimiterConfig{
		RateLimitQPSRead:     qpsRead,
		RateLimitBucketRead:  bucketRead,
		RateLimitQPSWrite:    qpsWrite,
		RateLimitBucketWrite: bucketWrite,
	}

	rateLimiter := buildNewRateLimiter(rateLimiterConfig)

	if rateLimiter.Reader.QPS() != qpsRead {
		t.Errorf("unexpected QPS (read) value: expected %f but found %f", qpsRead, rateLimiter.Reader.QPS())
	}

	if rateLimiter.Writer.QPS() != qpsWrite {
		t.Errorf("unexpected QPS (write) value: expected %f but found %f", qpsWrite, rateLimiter.Writer.QPS())
	}
}

func TestBuildRateLimiterWithDefaults(t *testing.T) {
	rateLimiterConfig := &RateLimiterConfig{}

	rateLimiter := buildNewRateLimiter(rateLimiterConfig)

	if rateLimiter.Reader.QPS() != RateLimitQPSDefault {
		t.Errorf("unexpected QPS (read) value: expected %f but found %f", RateLimitQPSDefault, rateLimiter.Reader.QPS())
	}

	if rateLimiter.Writer.QPS() != RateLimitQPSDefault {
		t.Errorf("unexpected QPS (write) value: expected %f but found %f", RateLimitQPSDefault, rateLimiter.Writer.QPS())
	}
}
