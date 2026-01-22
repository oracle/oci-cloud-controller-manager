package driver

import (
	"fmt"
	"runtime/debug"

	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MakeCSIPanicRecovery returns a defer-able closure that recovers from panics in CSI RPCs,
// logs the panic with a stack trace, and emits a panic metric. Use like:
//	defer MakeCSIPanicRecovery(logger, metricPusher, "CreateVolume", map[string]string{ metrics.ResourceOCIDDimension: req.GetName() }, &err, codes.Internal)()

func MakeCSIPanicRecoveryWithError(logger *zap.SugaredLogger, metricPusher *metrics.MetricPusher, op string, extraDims map[string]string,
	outErr *error, code codes.Code) func() {
	return func() {
		if rec := recover(); rec != nil {
			err := fmt.Errorf("panic recovered %v stack is %s", rec, string(debug.Stack()))
			logger.With(zap.Error(err)).With("operation", op).Error("Recovered from panic in CSI RPC")

			dimensionsMap := map[string]string{}
			for k, v := range extraDims {
				dimensionsMap[k] = v
			}
			metricDimension := util.GetMetricDimensionForComponent(util.PANIC, util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = metricDimension
			metrics.SendMetricData(metricPusher, metricDimension, 1, dimensionsMap)

			// If the method didn't already set an error, convert the panic to a gRPC error.
			if outErr != nil && *outErr == nil {
				*outErr = status.Errorf(code, "Internal error occurred while processing request.")
			}
		}
	}
}
