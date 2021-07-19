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
	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"go.uber.org/zap"
)

//GetClient returns the client for given Configuration
func GetClient(logger *zap.SugaredLogger, cfg *config.Config) (Interface, error) {
	cp, err := config.NewConfigurationProvider(cfg)
	if err != nil {
		logger.With(zap.Error(err)).Fatal("Unable to create client.")
		return nil, err
	}

	rateLimiter := NewRateLimiter(logger, cfg.RateLimiter)

	c, err := New(logger, cp, &rateLimiter)
	return c, err
}
