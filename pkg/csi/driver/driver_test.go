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
