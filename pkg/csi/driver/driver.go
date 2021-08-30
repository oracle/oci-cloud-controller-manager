package driver

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

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/instance/metadata"
)

const (
	// DriverName defines the driver name to be used in Kubernetes
	DriverName = "blockvolume.csi.oraclecloud.com"

	// DriverVersion is the version of the CSI driver
	DriverVersion = "0.1.0"

	// Default config file path
	configFilePath = "/etc/oci/config.yaml"
)

//Driver implements only Identity interface and embed Controller and Node interface.
type Driver struct {
	*ControllerDriver
	*NodeDriver
	endpoint string
	srv      *grpc.Server
	readyMu  sync.Mutex // protects ready
	ready    bool
	logger   *zap.SugaredLogger
}

// ControllerDriver implements CSI Controller interfaces
type ControllerDriver struct {
	KubeClient   kubernetes.Interface
	logger       *zap.SugaredLogger
	config       *providercfg.Config
	client       client.Interface
	util         *Util
	metricPusher *metrics.MetricPusher
}

// NodeDriver implements CSI Node interfaces
type NodeDriver struct {
	nodeID     string
	KubeClient kubernetes.Interface
	logger     *zap.SugaredLogger
	util       *Util
	volumeLocks *VolumeLocks
}

// NewNodeDriver creates a new CSI node driver for OCI blockvolume
func NewNodeDriver(logger *zap.SugaredLogger, endpoint, nodeID, kubeconfig, master string) (*Driver, error) {
	logger.With("endpoint", endpoint, "kubeconfig", kubeconfig, "master", master, "nodeID",
		nodeID).Info("Creating a new CSI Node driver.")

	kubeClientSet := getKubeClient(logger, master, kubeconfig)

	drv := NodeDriver{
		nodeID:     nodeID,
		KubeClient: kubeClientSet,
		logger:     logger,
		util:       &Util{logger: logger},
		volumeLocks: NewVolumeLocks(),
	}

	return &Driver{
		ControllerDriver: nil,
		NodeDriver:       &drv,
		endpoint:         endpoint,
		logger:           logger,
	}, nil

}

// NewControllerDriver creates a new CSI driver for OCI blockvolume
func NewControllerDriver(logger *zap.SugaredLogger, endpoint, kubeconfig, master string) (*Driver, error) {
	logger.With("endpoint", endpoint, "kubeconfig", kubeconfig, "master",
		master).Info("Creating a new CSI Controller driver.")

	kubeClientSet := getKubeClient(logger, master, kubeconfig)

	cfg := getConfig(logger)

	c := getClient(logger)

	drv := ControllerDriver{
		KubeClient: kubeClientSet,
		logger:     logger,
		util:       &Util{logger: logger},
		config:     cfg,
		client:     c,
	}

	return &Driver{
		ControllerDriver: &drv,
		NodeDriver:       nil,
		endpoint:         endpoint,
		logger:           logger,
	}, nil
}

// Run starts a gRPC server on the given endpoint
func (d *Driver) Run() error {
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
	csi.RegisterControllerServer(d.srv, d)
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

	d.logger.Info("CSI ControllerDriver has started.")
	return d.srv.Serve(listener)
}

func getConfig(logger *zap.SugaredLogger) *providercfg.Config {
	configPath, ok := os.LookupEnv("CONFIG_YAML_FILENAME")
	if !ok {
		configPath = configFilePath
	}

	cfg, err := providercfg.FromFile(configPath)
	if err != nil {
		logger.With(zap.Error(err)).With("config", configPath).Fatal("Failed to load configuration file from given path.")
	}

	err = cfg.Validate()
	if err != nil {
		logger.With(zap.Error(err)).With("config", configPath).Fatal("Failed to validate. Invalid configuration.")
	}

	if cfg.CompartmentID == "" {
		metadata, err := metadata.New().Get()
		if err != nil {
			logger.With(zap.Error(err)).With("config", configPath).Fatalf("Neither CompartmentID is given. Nor able to retrieve compartment OCID from metadata.")
		}
		cfg.CompartmentID = metadata.CompartmentID
	}

	return cfg
}

func getClient(logger *zap.SugaredLogger) client.Interface {
	cfg := getConfig(logger)

	c, err := client.GetClient(logger, cfg)

	if err != nil {
		logger.With(zap.Error(err)).Fatal("client can not be generated.")
	}
	return c
}

// Stop stops the plugin
func (d *Driver) Stop() {
	d.logger.Info("Stopping the gRPC server")
	d.srv.Stop()
}
