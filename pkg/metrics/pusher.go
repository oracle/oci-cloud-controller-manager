package metrics

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/oracle/oci-go-sdk/v50/common"
	"github.com/oracle/oci-go-sdk/v50/common/auth"
	"github.com/oracle/oci-go-sdk/v50/monitoring"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	ociclient "github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
)

const (
	metricSubmissionTimeout       = time.Second * 5
	jitterFactor                  = 0.1
	configFilePath                = "/etc/oci/config.yaml"
	telemetryIngestionServiceName = "telemetry-ingestion"
	configFileName                = "config.yaml"
)

// MonitoringClient is wrapper interface over the oci golang monitoring client
type MonitoringClient interface {
	PostMetricData(ctx context.Context, request monitoring.PostMetricDataRequest) (response monitoring.PostMetricDataResponse, err error)
}

// MetricPusher is the wrapper used to push metrics to OCI monitoring service.
type MetricPusher struct {
	// namespace in OCI monitoring
	namespace string
	// resourceGroup for OCI monitoring
	resourceGroup string
	// compartmentOCID is the the compartment OCID to be
	// used by OCI monitoring service.
	compartmentOCID string
	// metricPrefix is the prefix which should to added to
	// every metric
	metricPrefix string
	// telemetryClient is the monitoring client.
	telemetryClient MonitoringClient
	logger          *zap.SugaredLogger
}

// NewMetricPusher creates a new OCI Metric pusher
func NewMetricPusher(logger *zap.SugaredLogger) (*MetricPusher, error) {
	// we need the following information to push to OCI monitoring service
	// 1. Compartment for OCI Monitoring
	// 2. Monitoring Namespace
	// 3. Monitoring Resource Group
	// More details are available on this public doc
	// https://docs.cloud.oracle.com/en-us/iaas/Content/Monitoring/Concepts/monitoringoverview.htm#MetricsOverview
	var cpoOk bool
	var fvdOk bool
	var cpoConfig string
	var fvdConfig string
	var configPath string

	// Enable config file for CPO to push metrics
	cpoConfig, cpoOk = os.LookupEnv("CONFIG_YAML_FILENAME")
	// Enable config file for FVD to push metrics
	fvdConfig, fvdOk = os.LookupEnv("OCI_FLEXD_DRIVER_DIRECTORY")
	if cpoOk {
		configPath = cpoConfig
	} else if fvdOk {
		configPath = fmt.Sprintf("%s/%s", fvdConfig, configFileName)
	} else {
		configPath = configFilePath
	}

	cfg, err := providercfg.FromFile(configPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load configuration file at path %s", configPath)
	}

	err = cfg.Validate()
	if err != nil {
		return nil, errors.Wrapf(err, "invalid configuration")
	}
	// metrics collection will not be enabled
	if cfg.Metrics == nil {
		return nil, nil
	}

	cp, err := auth.InstancePrincipalConfigurationProvider()
	if err != nil {
		logger.With("error", err).Error("error occurred while creating auth provider")
		return nil, err
	}

	telemetryEndpoint := common.StringToRegion(cfg.RegionKey).Endpoint(telemetryIngestionServiceName)
	client, err := monitoring.NewMonitoringClientWithConfigurationProvider(cp)
	if err != nil {
		logger.With(err).Error("error occurred while creating monitoring client")
		return nil, err
	}
	client.Host = telemetryEndpoint

	return &MetricPusher{
		resourceGroup:   cfg.Metrics.ResourceGroup,
		namespace:       cfg.Metrics.Namespace,
		compartmentOCID: cfg.Metrics.CompartmentID,
		telemetryClient: client,
		logger:          logger,
		metricPrefix:    cfg.Metrics.Prefix,
	}, nil
}

// SendMetricData sends a metric to OCI monitoring.
// name is the metric name
// value is the metric value which is float by default
// dimensions are the custom dimensions to be added to the metric
func (p *MetricPusher) sendMetricData(name string, value float64, dimensions map[string]string) {
	now := common.SDKTime{Time: time.Now()}

	dataPoint := monitoring.Datapoint{
		Value:     &value,
		Timestamp: &now,
	}

	metricNameWithPrefix := p.metricPrefix + name
	metricData := monitoring.MetricDataDetails{
		Namespace:     &p.namespace,
		ResourceGroup: &p.resourceGroup,
		CompartmentId: &p.compartmentOCID,
		Name:          &metricNameWithPrefix,
		Dimensions:    dimensions,
		Datapoints:    []monitoring.Datapoint{dataPoint},
	}

	metricDataRequest := monitoring.PostMetricDataRequest{
		PostMetricDataDetails: monitoring.PostMetricDataDetails{
			MetricData: []monitoring.MetricDataDetails{metricData},
		},
		RequestMetadata: common.RequestMetadata{
			RetryPolicy: ociclient.NewRetryPolicyWithMaxAttempts(2),
		},
	}
	context, cancel := context.WithTimeout(context.Background(), metricSubmissionTimeout)
	defer cancel()
	response, err := p.telemetryClient.PostMetricData(context, metricDataRequest)
	if err != nil {
		p.logger.With("error", err).Errorf("error occurred while pushing metrics to OCI monitoring")
		return
	}
	if *response.FailedMetricsCount > 0 {
		p.logger.With("failedMetrics", response.FailedMetrics).Warnf("metrics could not be submitted successfully")
	} else {
		p.logger.With("metricName", metricNameWithPrefix).Info("metrics were submitted successfully")
	}
}

// SendMetricData is used to send the metric
func SendMetricData(metricPusher *MetricPusher, metric string, value float64, dimensionsMap map[string]string) {
	if metricPusher == nil {
		return
	}

	// in case of empty dimension value, fill it with unknown
	dimensions := prepareDimensions(dimensionsMap)
	metricPusher.sendMetricData(metric, value, dimensions)
}

// in case of empty values, store unknown. empty values are not accepted in dimensions
func prepareDimensions(dimensions map[string]string) map[string]string {
	for k, v := range dimensions {
		if len(v) == 0 {
			dimensions[k] = "unknown"
		}
	}
	return dimensions
}
