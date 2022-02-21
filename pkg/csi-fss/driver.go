package csi_fss

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes"

	"github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
)

const (
	// DriverName defines the driver name to be used in Kubernetes
	DriverName = "fss.csi.oraclecloud.com"

	// DriverVersion is the version of the CSI driver
	DriverVersion = "0.1.0"
)

//Driver implements only Identity interface and embed Controller and Node interface.
type Driver struct {
	*NodeDriver
	endpoint     string
	srv          *grpc.Server
	readyMu      sync.Mutex // protects ready
	ready        bool
	logger       *zap.SugaredLogger
	metricPusher *metrics.MetricPusher
}

// NodeDriver implements CSI Node interfaces
type NodeDriver struct {
	nodeID     string
	KubeClient kubernetes.Interface
	logger     *zap.SugaredLogger
	util       *csi_util.Util
	volumeLocks *csi_util.VolumeLocks
}

// NewNodeDriver creates a new CSI node driver for OCI FSS
func NewNodeDriver(logger *zap.SugaredLogger, endpoint, nodeID, kubeconfig, master string) (*Driver, error) {
	logger.With("endpoint", endpoint, "kubeconfig", kubeconfig, "master", master, "nodeID",
		nodeID).Info("Creating a new CSI FSS Node driver.")

	kubeClientSet := csi_util.GetKubeClient(logger, master, kubeconfig)

	drv := NodeDriver{
		nodeID:      nodeID,
		KubeClient:  kubeClientSet,
		logger:      logger,
		util:        &csi_util.Util{Logger: logger},
		volumeLocks: csi_util.NewVolumeLocks(),
	}

	return &Driver{
		NodeDriver:       &drv,
		endpoint:         endpoint,
		logger:           logger,
	}, nil

}

// Run starts a gRPC server on the given endpoint
func (d *Driver) Run() error {
	d.logger.Info("Running CSI FSS node driver")
	u, err := url.Parse(d.endpoint)
	if err != nil {
		d.logger.With("endpoint", d.endpoint).With("Failed to parse address").Error(err)
		return fmt.Errorf("failed to parse the address: %s", d.endpoint)
	}

	addr := path.Join(u.Host, filepath.FromSlash(u.Path))
	if u.Host == "" {
		addr = filepath.FromSlash(u.Path)
	}

	// CSI plugins talk only over UNIX sockets currently
	if u.Scheme != "unix" {
		d.logger.With("schema", u.Scheme).With("Currently only unix domain sockets are supported").Error(err)
		return fmt.Errorf("currently only unix domain sockets are supported, have: %s", u.Scheme)
	}

	// remove the socket if it's already there. This can happen if we
	// deploy a new version and the socket was created from the old running plugin.
	d.logger.With("address", addr).Info("Removing socket.")
	if err := os.Remove(addr); err != nil && !os.IsNotExist(err) {
		d.logger.With("address", addr).With("Failed to remove unix domain socket file").Error(err)
		return fmt.Errorf("failed to remove unix domain socket file %s", addr)
	}

	listener, err := net.Listen(u.Scheme, addr)
	if err != nil {
		d.logger.With("address", addr).With("Failed to listen").Error(err)
		return fmt.Errorf("failed to listen")
	}

	errHandler := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			d.logger.With(zap.Error(err)).With("method", info.FullMethod, "request", protosanitizer.StripSecrets(req)).Error("Failed to process gRPC request.")
		} else {
			d.logger.With("method", info.FullMethod, "response", protosanitizer.StripSecrets(resp)).Info("gRPC response is sent successfully.")
		}

		return resp, err
	}

	d.ready = true
	d.srv = grpc.NewServer(grpc.UnaryInterceptor(errHandler))
	csi.RegisterIdentityServer(d.srv, d)
	csi.RegisterNodeServer(d.srv, d)

	metricPusher, err := metrics.NewMetricPusher(d.logger)
	if err != nil {
		d.logger.With("error", err).Error("Metrics collection could not be enabled")
		// disable metric collection
		metricPusher = nil
	}
	if metricPusher != nil {
		d.logger.Info("Metrics collection has been enabled")
		d.metricPusher = metricPusher
	} else {
		d.logger.Info("Metrics collection is not enabled")
	}

	d.logger.Info("CSI FSS Driver has started.")
	return d.srv.Serve(listener)
}

// Stop stops the plugin
func (d *Driver) Stop() {
	d.logger.Info("Stopping the gRPC server")
	d.srv.Stop()
}
