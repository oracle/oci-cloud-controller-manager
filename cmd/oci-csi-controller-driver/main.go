package main

import (
	"flag"
	"github.com/oracle/oci-cloud-controller-manager/pkg/csi/driver"
	"github.com/oracle/oci-cloud-controller-manager/pkg/logging"
	"go.uber.org/zap"
)

var (
	endpoint   = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	nodeID     = flag.String("nodeid", "", "node id")
	logLevel   = flag.String("loglevel", "info", "log level")
	master     = flag.String("master", "", "kube master")
	kubeconfig = flag.String("kubeconfig", "", "cluster kube config")
)

func main() {

	flag.Parse()

	logger := logging.Logger().Sugar()
	logger.Sync()

	drv, err := driver.NewControllerDriver(logger, *endpoint, *kubeconfig, *master)
	if err != nil {
		logger.With(zap.Error(err)).Fatalf("Failed to create controller driver.")
	}

	if err := drv.Run(); err != nil {
		logger.With(zap.Error(err)).Fatalf("Failed to run the CSI driver.")
	}

	logger.Info("CSI driver exited")
}
