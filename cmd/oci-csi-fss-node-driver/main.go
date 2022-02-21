package main

import (
	"flag"
	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-node-driver/nodedriveroptions"
	"github.com/oracle/oci-cloud-controller-manager/pkg/csi-fss"
	"github.com/oracle/oci-cloud-controller-manager/pkg/logging"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util/signals"
	"go.uber.org/zap"
	"k8s.io/klog"
)

func main() {
	nodecsioptions := nodedriveroptions.NodeCSIOptions{}

	flag.StringVar(&nodecsioptions.Endpoint, "endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	flag.StringVar(&nodecsioptions.NodeID, "nodeid", "", "node id")
	flag.StringVar(&nodecsioptions.LogLevel, "loglevel", "info", "log level")
	flag.StringVar(&nodecsioptions.Master, "master", "", "kube master")
	flag.StringVar(&nodecsioptions.Kubeconfig, "kubeconfig", "", "cluster kubeconfig")
	flag.DurationVar(&nodecsioptions.ConnectionTimeout, "connection-timeout", 0, "The --connection-timeout flag is deprecated")
	flag.StringVar(&nodecsioptions.CsiAddress, "csi-address", "/run/csi/socket", "Path of the CSI driver socket that the node-driver-registrar will connect to.")
	flag.StringVar(&nodecsioptions.KubeletRegistrationPath, "kubelet-registration-path", "", "Path of the CSI driver socket on the Kubernetes host machine.")

	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()
	stopCh := signals.SetupSignalHandler()

	logger := logging.Logger().Sugar()
	logger.Sync()

	drv, err := csi_fss.NewNodeDriver(logger, nodecsioptions.Endpoint, nodecsioptions.NodeID, nodecsioptions.Kubeconfig, nodecsioptions.Master)
	if err != nil {
		logger.With(zap.Error(err)).Fatal("Failed to create CSI FSS Node driver.")
	}

	if err := drv.Run(); err != nil {
		logger.With(zap.Error(err)).Fatal("Failed to run the CSI FSS node driver.")
	}

	logger.Info("CSI FSS driver exited")

	<-stopCh
}
