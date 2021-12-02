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

	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-node-driver/nodedriveroptions"
	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	csi_util "github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/instance/metadata"
)

const (
	// BlockVolumeDriverName defines the driver name to be used in Kubernetes
	BlockVolumeDriverName = "blockvolume.csi.oraclecloud.com"

	// BlockVolumeDriverVersion is the version of the CSI driver
	BlockVolumeDriverVersion = "0.1.0"

	// FSSDriverName defines the driver name to be used in Kubernetes
	FSSDriverName = "fss.csi.oraclecloud.com"

	// FSSDriverVersion is the version of the CSI driver
	FSSDriverVersion = "0.1.0"

	// Default config file path
	configFilePath = "/etc/oci/config.yaml"
)

//Driver implements only Identity interface and embed Controller and Node interface.
type Driver struct {
	*ControllerDriver
	nodeDriver             csi.NodeServer
	name                   string
	version                string
	endpoint               string
	srv                    *grpc.Server
	readyMu                sync.Mutex // protects ready
	ready                  bool
	logger                 *zap.SugaredLogger
	enableControllerServer bool
}

// ControllerDriver implements CSI Controller interfaces
type ControllerDriver struct {
	KubeClient   kubernetes.Interface
	logger       *zap.SugaredLogger
	config       *providercfg.Config
	client       client.Interface
	util         *csi_util.Util
	metricPusher *metrics.MetricPusher
}

// NodeDriver implements CSI Node interfaces
type NodeDriver struct {
	nodeID      string
	KubeClient  kubernetes.Interface
	logger      *zap.SugaredLogger
	util        *csi_util.Util
	volumeLocks *csi_util.VolumeLocks
}

// BlockVolumeNodeDriver extends NodeDriver
type BlockVolumeNodeDriver struct {
	NodeDriver
}

// FSSNodeDriver extends NodeDriver
type FSSNodeDriver struct {
	NodeDriver
}

func newNodeDriver(nodeID string, kubeClientSet kubernetes.Interface, logger *zap.SugaredLogger) NodeDriver {
	return NodeDriver{
		nodeID:      nodeID,
		KubeClient:  kubeClientSet,
		logger:      logger,
		util:        &csi_util.Util{Logger: logger},
		volumeLocks: csi_util.NewVolumeLocks(),
	}
}

func GetNodeDriver(name string, nodeID string, kubeClientSet kubernetes.Interface, logger *zap.SugaredLogger) csi.NodeServer {
	if name == BlockVolumeDriverName {
		return BlockVolumeNodeDriver{NodeDriver: newNodeDriver(nodeID, kubeClientSet, logger)}
	}
	if name == FSSDriverName {
		return FSSNodeDriver{NodeDriver: newNodeDriver(nodeID, kubeClientSet, logger)}
	}
	return nil
}

// NewNodeDriver creates a new CSI node driver for OCI blockvolume
func NewNodeDriver(logger *zap.SugaredLogger, nodeOptions nodedriveroptions.NodeOptions) (*Driver, error) {
	logger.With("endpoint", nodeOptions.Endpoint, "kubeconfig", nodeOptions.Kubeconfig, "master", nodeOptions.Master, "nodeID",
		nodeOptions.NodeID).Info("Creating a new CSI Node driver.")

	kubeClientSet := csi_util.GetKubeClient(logger, nodeOptions.Master, nodeOptions.Kubeconfig)

	return &Driver{
		ControllerDriver:       nil,
		nodeDriver:             GetNodeDriver(nodeOptions.DriverName, nodeOptions.NodeID, kubeClientSet, logger),
		endpoint:               nodeOptions.Endpoint,
		logger:                 logger,
		enableControllerServer: nodeOptions.EnableControllerServer,
		name:                   nodeOptions.DriverName,
		version:                nodeOptions.DriverVersion,
	}, nil

}

// NewControllerDriver creates a new CSI driver for OCI blockvolume
func NewControllerDriver(logger *zap.SugaredLogger, endpoint, kubeconfig, master string, enableControllerServer bool, name, version string) (*Driver, error) {
	logger.With("endpoint", endpoint, "kubeconfig", kubeconfig, "master",
		master).Info("Creating a new CSI Controller driver.")

	kubeClientSet := csi_util.GetKubeClient(logger, master, kubeconfig)

	cfg := getConfig(logger)

	c := getClient(logger)

	drv := ControllerDriver{
		KubeClient: kubeClientSet,
		logger:     logger,
		util:       &csi_util.Util{Logger: logger},
		config:     cfg,
		client:     c,
	}

	return &Driver{
		ControllerDriver:       &drv,
		nodeDriver:             nil,
		endpoint:               endpoint,
		logger:                 logger,
		enableControllerServer: enableControllerServer,
		name:                   name,
		version:                version,
	}, nil
}

func (d *Driver) GetNodeDriver() csi.NodeServer {
	if d.name == BlockVolumeDriverName {
		return d.nodeDriver.(BlockVolumeNodeDriver)
	}
	if d.name == FSSDriverName {
		return d.nodeDriver.(FSSNodeDriver)
	}
	return nil
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
		d.logger.With("address", addr).With("msg", "Failed to listen").Error(err)
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
	if d.enableControllerServer {
		csi.RegisterControllerServer(d.srv, d)
	} else {
		csi.RegisterNodeServer(d.srv, d.GetNodeDriver())
	}

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

	d.logger.Info("CSI Driver has started.")
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
