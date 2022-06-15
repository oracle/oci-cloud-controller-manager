// Copyright 2017 Oracle and/or its affiliates. All rights reserved.
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

package core

import (
	"context"
	"math/rand"
	"os"
	"strings"
	"time"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/instance/metadata"
	"github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/block"
	"github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/fss"
	"github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/plugin"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/informers"
	informersv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/v8/controller"
)

const (
	// ProvisionerNameDefault is the name of the default OCI volume provisioner (block)
	ProvisionerNameDefault = "oracle.com/oci"
	// ProvisionerNameBlock is the name of the OCI block volume provisioner
	ProvisionerNameBlock = "oracle.com/oci-block"
	// ProvisionerNameFss is the name of the OCI FSS dedicated storage provisioner
	ProvisionerNameFss     = "oracle.com/oci-fss"
	ociProvisionerIdentity = "ociProvisionerIdentity"
	ociAvailabilityDomain  = "ociAvailabilityDomain"
	ociCompartment         = "ociCompartment"
	configFilePath         = "/etc/oci/config.yaml"
)

const (
	resyncPeriod              = 15 * time.Second
	minResyncPeriod           = 12 * time.Hour
	exponentialBackOffOnError = false
	failedRetryThreshold      = 5
)

// OCIProvisioner is a dynamic volume provisioner that satisfies
// the Kubernetes external storage Provisioner controller interface
type OCIProvisioner struct {
	client     client.Interface
	kubeClient kubernetes.Interface

	nodeLister       listersv1.NodeLister
	nodeListerSynced cache.InformerSynced

	provisioner plugin.ProvisionerPlugin

	compartmentID string

	logger *zap.SugaredLogger
}

// NewOCIProvisioner creates a new OCI provisioner.
func NewOCIProvisioner(logger *zap.SugaredLogger, kubeClient kubernetes.Interface, nodeInformer informersv1.NodeInformer, provisionerType string, volumeRoundingEnabled bool, minVolumeSize resource.Quantity) (*OCIProvisioner, error) {
	configPath, ok := os.LookupEnv("CONFIG_YAML_FILENAME")
	if !ok {
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

	metadata, mdErr := metadata.New().Get()
	if mdErr != nil {
		logger.With(zap.Error(mdErr)).Warnf("unable to retrieve instance metadata.")
	}

	if cfg.CompartmentID == "" {
		if metadata == nil {
			return nil, errors.Wrap(mdErr, "unable to get compartment OCID")
		}

		logger.With("compartmentID", metadata.CompartmentID).Infof("'CompartmentID' not given. Using compartment OCID from instance metadata.")
		cfg.CompartmentID = metadata.CompartmentID
	}

	cp, err := providercfg.NewConfigurationProvider(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create volume provisioner client.")
	}

	tenancyID, err := cp.TenancyOCID()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to detrimine tenancy")
	}

	logger = logger.With(
		"compartmentID", cfg.CompartmentID,
		"tenancyID", tenancyID,
	)

	rateLimiter := client.NewRateLimiter(logger, cfg.RateLimiter)

	client, err := client.New(logger, cp, &rateLimiter)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to construct OCI client")
	}

	region, ok := os.LookupEnv("OCI_SHORT_REGION")
	if !ok {
		if mdErr != nil {
			return nil, errors.Wrap(err, "region not provided and cant detect from metadata")
		}
		region = metadata.Region
	}

	var provisioner plugin.ProvisionerPlugin
	switch provisionerType {
	case ProvisionerNameDefault, ProvisionerNameBlock:
		provisioner = block.NewBlockProvisioner(
			logger,
			client,
			region,
			cfg.CompartmentID,
			volumeRoundingEnabled,
			minVolumeSize,
		)
	case ProvisionerNameFss:
		provisioner = fss.NewFilesystemProvisioner(logger, client, region, cfg.CompartmentID)
	default:
		return nil, errors.Errorf("invalid provisioner type %q", provisionerType)
	}
	return &OCIProvisioner{
		client:           client,
		kubeClient:       kubeClient,
		nodeLister:       nodeInformer.Lister(),
		nodeListerSynced: nodeInformer.Informer().HasSynced,
		provisioner:      provisioner,
		compartmentID:    cfg.CompartmentID,
		logger:           logger,
	}, nil
}

var _ controller.Provisioner = &OCIProvisioner{}

func roundUpSize(volumeSizeBytes int64, allocationUnitBytes int64) int64 {
	return (volumeSizeBytes + allocationUnitBytes - 1) / allocationUnitBytes
}

// mapAvailabilityDomainToFailureDomain maps a given Availability Domain to a
// k8s label compat. string.
func mapAvailabilityDomainToFailureDomain(AD string) string {
	parts := strings.SplitN(AD, ":", 2)
	if parts == nil {
		return ""
	}
	return parts[len(parts)-1]
}

// Provision creates a storage asset and returns a PV object representing it.
func (p *OCIProvisioner) Provision(ctx context.Context, options controller.ProvisionOptions) (*v1.PersistentVolume, controller.ProvisioningState, error) {
	availabilityDomainName, availabilityDomain, err := p.chooseAvailabilityDomain(context.Background(), options.PVC)
	if err != nil {
		return nil, controller.ProvisioningFinished, err
	}
	persistentVolume, err := p.provisioner.Provision(options, availabilityDomain)
	if err == nil {
		persistentVolume.ObjectMeta.Annotations[ociProvisionerIdentity] = ociProvisionerIdentity
		persistentVolume.ObjectMeta.Annotations[ociAvailabilityDomain] = availabilityDomainName
		persistentVolume.ObjectMeta.Annotations[ociCompartment] = p.compartmentID
		persistentVolume.ObjectMeta.Labels[v1.LabelZoneFailureDomain] = mapAvailabilityDomainToFailureDomain(*availabilityDomain.Name)
	}
	return persistentVolume, controller.ProvisioningFinished, err
}

// Delete removes the storage asset that was created by Provision represented
// by the given PV.
func (p *OCIProvisioner) Delete(ctx context.Context, volume *v1.PersistentVolume) error {
	identity, ok := volume.Annotations[ociProvisionerIdentity]
	if !ok {
		return errors.New("identity annotation not found on PV")
	}
	if identity != ociProvisionerIdentity {
		return &controller.IgnoredError{Reason: "identity annotation on PV does not match ours"}
	}
	return p.provisioner.Delete(volume)
}

// Ready waits unitl the the nodeLister has been synced.
func (p *OCIProvisioner) Ready(stopCh <-chan struct{}) error {
	if !cache.WaitForCacheSync(stopCh, p.nodeListerSynced) {
		return errors.New("unable to sync caches for OCI Volume Provisioner")
	}
	return nil
}

// informerResyncPeriod computes the time interval a shared informer waits
// before resyncing with the API server.
func informerResyncPeriod(minResyncPeriod time.Duration) func() time.Duration {
	return func() time.Duration {
		factor := rand.Float64() + 1
		return time.Duration(float64(minResyncPeriod.Nanoseconds()) * factor)
	}
}

// Run runs the volume provisoner control loop
func Run(logger *zap.SugaredLogger, kubeconfig string, master string, minVolumeSize string, volumeRoundingEnabled bool, stopCh <-chan struct{}) error {

	// Create an InClusterConfig and use it to create a client for the controller
	// to use to communicate with Kubernetes
	config, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
	if err != nil {
		return errors.Wrapf(err, "failed to load config")
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrapf(err, "failed to create Kubernetes client")
	}

	if err != nil {
		return errors.Wrapf(err, "failed to get kube-apiserver version")
	}

	// Decides what type of provider to deploy, either block or fss
	provisionerType := os.Getenv("PROVISIONER_TYPE")
	if provisionerType == "" {
		provisionerType = ProvisionerNameDefault
	}
	logger = logger.With("provisionerType", provisionerType)
	logger.Infof("Starting volume provisioner in %q mode", provisionerType)

	// TODO Should use an existing informer factory instead of constructing a new one
	sharedInformerFactory := informers.NewSharedInformerFactory(clientset, informerResyncPeriod(minResyncPeriod)())

	volumeSizeLowerBound, err := resource.ParseQuantity(minVolumeSize)
	if err != nil {
		return errors.Wrapf(err, "failed to parse minimum volume size")
	}

	// Create the provisioner: it implements the Provisioner interface expected by
	// the controller
	ociProvisioner, err := NewOCIProvisioner(logger, clientset, sharedInformerFactory.Core().V1().Nodes(), provisionerType, volumeRoundingEnabled, volumeSizeLowerBound)
	if err != nil {
		return errors.Wrapf(err, "cannot create volume provisioner.")
	}
	// Start the provision controller which will dynamically provision oci
	// PVs
	pc := controller.NewProvisionController(
		clientset,
		provisionerType,
		ociProvisioner,
		controller.ResyncPeriod(resyncPeriod),
		controller.ExponentialBackOffOnError(exponentialBackOffOnError),
		controller.FailedProvisionThreshold(failedRetryThreshold),
		controller.Threadiness(2),
	)
	go sharedInformerFactory.Start(stopCh)
	// We block waiting for Ready() after the shared informer factory has
	// started so we don't deadlock waiting for caches to sync.
	if err := ociProvisioner.Ready(stopCh); err != nil {
		return errors.Wrapf(err, "failed to start volume provisioner")
	}
	pc.Run(context.Background())
	return errors.Errorf("unreachable")
}
