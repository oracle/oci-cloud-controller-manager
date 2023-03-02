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

package driver

import (
	"fmt"
	"github.com/oracle/oci-cloud-controller-manager/pkg/logging"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	"go.uber.org/zap"
	"testing"
)

func Test_getMetricPusher(t *testing.T) {
	l := logging.Logger().Sugar()
	tests := map[string]struct {
		metricPusherGetter  MetricPusherGetter
		logger              *zap.SugaredLogger
		wantMetricPusherNil bool
		wantErr             bool
	}{
		"Get Metric Pusher Object Success": {
			metricPusherGetter:  getMetricPusherSuccess,
			logger:              l,
			wantMetricPusherNil: false,
			wantErr:             false,
		},
		"Get Metric Pusher Object Failure": {
			metricPusherGetter:  getMetricPusherFailure,
			logger:              l,
			wantMetricPusherNil: true,
			wantErr:             true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			metricPusherResponse, err := getMetricPusher(tt.metricPusherGetter, tt.logger)
			if tt.wantErr && err == nil {
				t.Errorf("Wanted error but got no error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Wanted no error but got error")
			}
			if tt.wantMetricPusherNil && metricPusherResponse != nil {
				t.Errorf("Wanted nil object but got non-nil object")
			}
			if !tt.wantMetricPusherNil && metricPusherResponse == nil {
				t.Errorf("Wanted non-nil object but got nil object")
			}
		})
	}
}

func getMetricPusherSuccess(logger *zap.SugaredLogger) (*metrics.MetricPusher, error) {
	metricPusher := &metrics.MetricPusher{}
	return metricPusher, nil
}

func getMetricPusherFailure(logger *zap.SugaredLogger) (*metrics.MetricPusher, error) {
	return nil, fmt.Errorf("failed to get metric pusher")
}
