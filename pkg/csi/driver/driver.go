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
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"

	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-node-driver/nodedriveroptions"
	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	csi_util "github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/instance/metadata"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

var (
	BlockVolumeDriverName string
	FSSDriverName         string
	LustreDriverName      string
)

func init() {
	BlockVolumeDriverName = getEnv("BLOCK_VOLUME_DRIVER_NAME", "blockvolume.csi.oraclecloud.com")
	FSSDriverName = getEnv("FSS_VOLUME_DRIVER_NAME", "fss.csi.oraclecloud.com")
	LustreDriverName = getEnv("LUSTRE_VOLUME_DRIVER_NAME", "lustre.csi.oraclecloud.com")
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

const (

	// BlockVolumeDriverVersion is the version of the CSI driver
	BlockVolumeDriverVersion = "0.1.0"

	// FSSDriverVersion is the version of the CSI driver
	FSSDriverVersion = "0.1.0"

	// LustreDriverVersion is the version of the CSI driver
	LustreDriverVersion = "0.1.0"
	// Default config file path
	configFilePath = "/etc/oci/config.yaml"

	CSIConfigMapName = "oci-csi-config"
)

type CSIDriver string

const (
	BV  CSIDriver = "BV"
	FSS CSIDriver = "FSS"

	Lustre CSIDriver = "Lustre"
)

// Driver implements only Identity interface and embed Controller and Node interface.
type Driver struct {
	controllerDriver       csi.ControllerServer
	nodeDriver             csi.NodeServer
	name                   string
	version                string
	endpoint               string
	srv                    *grpc.Server
	readyMu                sync.Mutex // protects ready
	ready                  bool
	logger                 *zap.SugaredLogger
	enableControllerServer bool
	csi.UnimplementedIdentityServer
}

// ControllerDriver implements CSI Controller interfaces
type ControllerDriver struct {
	KubeClient      kubernetes.Interface
	logger          *zap.SugaredLogger
	config          *providercfg.Config
	client          client.Interface
	util            *csi_util.Util
	metricPusher    *metrics.MetricPusher
	clusterIpFamily string
	csi.UnimplementedControllerServer
}

// BlockVolumeControllerDriver extends ControllerDriver
type BlockVolumeControllerDriver struct {
	ControllerDriver
}

// FSSControllerDriver extends ControllerDriver
type FSSControllerDriver struct {
	ControllerDriver
	serviceAccountLister listersv1.ServiceAccountLister
}

// NodeDriver implements CSI Node interfaces
type NodeDriver struct {
	nodeID       string
	KubeClient   kubernetes.Interface
	logger       *zap.SugaredLogger
	util         *csi_util.Util
	volumeLocks  *csi_util.VolumeLocks
	nodeMetadata *csi_util.NodeMetadata
	csi.UnimplementedNodeServer
	csiConfig *csi_util.CSIConfig
}

// BlockVolumeNodeDriver extends NodeDriver
type BlockVolumeNodeDriver struct {
	NodeDriver
}

// FSSNodeDriver extends NodeDriver
type FSSNodeDriver struct {
	NodeDriver
}

type LustreNodeDriver struct {
	NodeDriver
}

type ControllerDriverConfig struct {
	CsiEndpoint            string
	CsiKubeConfig          string
	CsiMaster              string
	EnableControllerServer bool
	DriverName             string
	DriverVersion          string
	ClusterIpFamily        string
}

type MetricPusherGetter func(logger *zap.SugaredLogger) (*metrics.MetricPusher, error)

func newControllerDriver(kubeClientSet kubernetes.Interface, logger *zap.SugaredLogger, config *providercfg.Config, c client.Interface, metricPusher *metrics.MetricPusher, clusterIpFamily string) ControllerDriver {
	return ControllerDriver{
		KubeClient:      kubeClientSet,
		logger:          logger,
		util:            &csi_util.Util{Logger: logger},
		config:          config,
		client:          c,
		metricPusher:    metricPusher,
		clusterIpFamily: clusterIpFamily,
	}
}

func newNodeDriver(nodeID string, nodeMetaData *csi_util.NodeMetadata, kubeClientSet kubernetes.Interface, logger *zap.SugaredLogger, csiConfig *csi_util.CSIConfig) NodeDriver {
	return NodeDriver{
		nodeID:       nodeID,
		KubeClient:   kubeClientSet,
		logger:       logger,
		util:         &csi_util.Util{Logger: logger},
		volumeLocks:  csi_util.NewVolumeLocks(),
		nodeMetadata: nodeMetaData,
		csiConfig:    csiConfig,
	}
}

func GetControllerDriver(name string, kubeClientSet kubernetes.Interface, logger *zap.SugaredLogger, config *providercfg.Config, c client.Interface, clusterIpFamily string) csi.ControllerServer {
	metricPusher, err := getMetricPusher(newMetricPusher, logger)
	if err != nil {
		logger.With("error", err).Error("Metrics collection could not be enabled")
		// disable metric collection
		metricPusher = nil
	}

	if name == BlockVolumeDriverName {
		return &BlockVolumeControllerDriver{ControllerDriver: newControllerDriver(kubeClientSet, logger, config, c, metricPusher, clusterIpFamily)}
	}
	if name == FSSDriverName {

		factory := informers.NewSharedInformerFactory(kubeClientSet, 5*time.Minute)
		serviceAccountInformer := factory.Core().V1().ServiceAccounts()
		go serviceAccountInformer.Informer().Run(wait.NeverStop)

		if !cache.WaitForCacheSync(wait.NeverStop, serviceAccountInformer.Informer().HasSynced) {
			utilruntime.HandleError(fmt.Errorf("timed out waiting for informers to sync"))
		}

		return &FSSControllerDriver{ControllerDriver: newControllerDriver(kubeClientSet, logger, config, c, metricPusher, clusterIpFamily), serviceAccountLister: serviceAccountInformer.Lister()}

	}
	return nil
}

func newMetricPusher(logger *zap.SugaredLogger) (*metrics.MetricPusher, error) {
	metricPusher, err := metrics.NewMetricPusher(logger)
	return metricPusher, err
}

func getMetricPusher(metricPusherGetter MetricPusherGetter, logger *zap.SugaredLogger) (*metrics.MetricPusher, error) {
	metricPusher, err := metricPusherGetter(logger)
	if err != nil {
		logger.With("error", err).Error("Failed to get metric pusher")
		return nil, err
	}
	if metricPusher == nil {
		logger.Info("Failed to get metric pusher. Got nil object")
		return nil, fmt.Errorf("failed to get metric pusher")
	}
	logger.Info("Metrics collection has been enabled")
	return metricPusher, nil
}

func GetNodeDriver(name string, nodeID string, nodeMetadata *csi_util.NodeMetadata, kubeClientSet kubernetes.Interface, logger *zap.SugaredLogger, csiConfig *csi_util.CSIConfig) csi.NodeServer {
	if name == BlockVolumeDriverName {
		return BlockVolumeNodeDriver{NodeDriver: newNodeDriver(nodeID, nodeMetadata, kubeClientSet, logger, csiConfig)}
	}
	if name == FSSDriverName {
		return FSSNodeDriver{NodeDriver: newNodeDriver(nodeID, nodeMetadata, kubeClientSet, logger, csiConfig)}
	}
	if name == LustreDriverName {
		return LustreNodeDriver{NodeDriver: newNodeDriver(nodeID, nodeMetadata, kubeClientSet, logger, csiConfig)}
	}
	return nil
}

// NewNodeDriver creates a new CSI node driver for OCI blockvolume
func NewNodeDriver(logger *zap.SugaredLogger, nodeOptions nodedriveroptions.NodeOptions) (*Driver, error) {
	logger.With("endpoint", nodeOptions.Endpoint, "kubeconfig", nodeOptions.Kubeconfig, "master", nodeOptions.Master, "nodeID",
		nodeOptions.NodeID).Info("Creating a new CSI Node driver.")

	kubeClientSet := csi_util.GetKubeClient(logger, nodeOptions.Master, nodeOptions.Kubeconfig)
	nodeMetadata := &csi_util.NodeMetadata{}
	csiConfig := &csi_util.CSIConfig{}

	return &Driver{
		controllerDriver:       nil,
		nodeDriver:             GetNodeDriver(nodeOptions.DriverName, nodeOptions.NodeID, nodeMetadata, kubeClientSet, logger, csiConfig),
		endpoint:               nodeOptions.Endpoint,
		logger:                 logger,
		enableControllerServer: nodeOptions.EnableControllerServer,
		name:                   nodeOptions.DriverName,
		version:                nodeOptions.DriverVersion,
	}, nil

}

// NewControllerDriver creates a new CSI driver
func NewControllerDriver(logger *zap.SugaredLogger, driverConfig ControllerDriverConfig) (*Driver, error) {
	logger.With("endpoint", driverConfig.CsiEndpoint, "kubeconfig", driverConfig.CsiKubeConfig, "master",
		driverConfig.CsiMaster).Info("Creating a new CSI Controller driver.")

	kubeClientSet := csi_util.GetKubeClient(logger, driverConfig.CsiMaster, driverConfig.CsiKubeConfig)

	cfg := getConfig(logger)

	c := getClient(logger)

	return &Driver{
		controllerDriver:       GetControllerDriver(driverConfig.DriverName, kubeClientSet, logger, cfg, c, driverConfig.ClusterIpFamily),
		nodeDriver:             nil,
		endpoint:               driverConfig.CsiEndpoint,
		logger:                 logger,
		enableControllerServer: driverConfig.EnableControllerServer,
		name:                   driverConfig.DriverName,
		version:                driverConfig.DriverVersion,
	}, nil
}

func (d *Driver) GetControllerDriver() csi.ControllerServer {
	if d.name == BlockVolumeDriverName {
		return d.controllerDriver.(*BlockVolumeControllerDriver)
	}
	if d.name == FSSDriverName {
		return d.controllerDriver.(*FSSControllerDriver)
	}
	return nil
}

func (d *Driver) GetNodeDriver() csi.NodeServer {
	if d.name == BlockVolumeDriverName {
		return d.nodeDriver.(BlockVolumeNodeDriver)
	}
	if d.name == FSSDriverName {
		return d.nodeDriver.(FSSNodeDriver)
	}
	if d.name == LustreDriverName {
		return d.nodeDriver.(LustreNodeDriver)
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
		csi.RegisterControllerServer(d.srv, d.GetControllerDriver())
	} else {
		csi.RegisterNodeServer(d.srv, d.GetNodeDriver())
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
