package oci

import (
	"context"
	"fmt"
	"time"

	"github.com/oracle/oci-cloud-controller-manager/pkg/flexcidr"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const flexCIDRRetryDelay = time.Minute

type FlexCIDRController struct {
	nodeInformer    coreinformers.NodeInformer
	serviceInformer coreinformers.ServiceInformer
	kubeClient      clientset.Interface
	cloud           *CloudProvider
	queue           workqueue.RateLimitingInterface
	logger          *zap.SugaredLogger
	instanceCache   cache.Store
	ociClient       client.Interface
}

func NewFlexCIDRController(
	nodeInformer coreinformers.NodeInformer,
	serviceInformer coreinformers.ServiceInformer,
	kubeClient clientset.Interface,
	cloud *CloudProvider,
	logger *zap.SugaredLogger,
	instanceCache cache.Store,
	ociClient client.Interface) *FlexCIDRController {

	controller := &FlexCIDRController{
		nodeInformer:    nodeInformer,
		serviceInformer: serviceInformer,
		kubeClient:      kubeClient,
		cloud:           cloud,
		queue:           workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		logger:          logger,
		instanceCache:   instanceCache,
		ociClient:       ociClient,
	}

	controller.nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node := obj.(*v1.Node)
			controller.queue.Add(node.Name)
		},
		UpdateFunc: func(_, newObj interface{}) {
			node := newObj.(*v1.Node)
			controller.queue.Add(node.Name)
		},
	})

	return controller
}

func (fcc *FlexCIDRController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer fcc.queue.ShutDown()

	fcc.logger.Info("Starting flex CIDR controller")

	if !cache.WaitForCacheSync(stopCh, fcc.nodeInformer.Informer().HasSynced, fcc.serviceInformer.Informer().HasSynced) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for flex CIDR controller caches to sync"))
		return
	}

	wait.Until(fcc.runWorker, time.Second, stopCh)
}

func (fcc *FlexCIDRController) runWorker() {
	for fcc.processNextItem() {
	}
}

func (fcc *FlexCIDRController) processNextItem() bool {
	key, quit := fcc.queue.Get()
	if quit {
		return false
	}
	defer fcc.queue.Done(key)

	if err := fcc.processItem(key.(string)); err != nil {
		fcc.logger.Errorf("Error processing flex CIDR for node %s (will retry): %v", key, err)
		fcc.queue.AddRateLimited(key)
	} else {
		fcc.queue.Forget(key)
	}

	return true
}

func (fcc *FlexCIDRController) processItem(key string) error {
	logger := fcc.logger.With("node", key)

	node, err := fcc.nodeInformer.Lister().Get(key)
	if err != nil {
		return err
	}

	if len(node.Spec.PodCIDRs) > 0 && len(node.Spec.ProviderID) == 0 {
		logger.Debug("node already has podCIDRs but providerID is empty, skipping")
		return nil
	}

	instance, instanceID, err := fcc.getInstanceByNode(node, logger)
	if err != nil {
		return err
	}
	if instance == nil {
		return nil
	}

	if err := fcc.instanceCache.Add(instance); err != nil {
		logger.With(zap.Error(err)).Debug("failed to add instance to cache")
	}

	if instance.LifecycleState != core.InstanceLifecycleStateRunning {
		logger.Infof("instance %s not running yet, requeueing", instanceID)
		fcc.queue.AddAfter(key, flexCIDRRetryDelay)
		return nil
	}

	config, hasConfig := flexcidr.ParsePrimaryVnicConfig(instance)
	if !hasConfig {
		logger.Debug("instance metadata does not include flex CIDR configuration, skipping")
		return nil
	}

	clusterIPFamily, err := flexcidr.GetClusterIpFamily(context.Background(), fcc.serviceInformer.Lister())
	if err != nil {
		logger.With(zap.Error(err)).Info("cluster IP family not ready yet, requeueing")
		fcc.queue.AddAfter(key, flexCIDRRetryDelay)
		return nil
	}

	primaryVNIC, err := fcc.ociClient.Compute().GetPrimaryVNICForInstance(context.Background(), *instance.CompartmentId, instanceID)
	if err != nil {
		return errors.Wrap(err, "GetPrimaryVNICForInstance")
	}

	flexCIDRManager := &flexcidr.FlexCIDR{
		Logger:            logger,
		PrimaryVnicConfig: config,
		ClusterIpFamily:   clusterIPFamily,
		OciCoreClient:     fcc.ociClient.Networking(nil),
	}

	flexCIDRs, err := flexCIDRManager.GetOrCreateFlexCidrList(*primaryVNIC.Id)
	if err != nil {
		return err
	}
	if !flexCIDRManager.ValidateFlexCidrList(flexCIDRs) {
		return fmt.Errorf("computed flex CIDRs %v are invalid", flexCIDRs)
	}
	if flexcidr.StringSlicesEqualIgnoreOrder(node.Spec.PodCIDRs, flexCIDRs) {
		logger.Debugf("node already has expected podCIDRs %v", flexCIDRs)
		return nil
	}

	return flexcidr.PatchNodePodCIDRs(context.Background(), fcc.kubeClient, node.Name, flexCIDRs, logger)
}

func (fcc *FlexCIDRController) getInstanceByNode(node *v1.Node, logger *zap.SugaredLogger) (*core.Instance, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	providerID := node.Spec.ProviderID
	var err error
	if providerID == "" {
		providerID, err = fcc.cloud.InstanceID(ctx, types.NodeName(node.Name))
		if err != nil {
			return nil, "", err
		}
	}

	instanceID, err := MapProviderIDToResourceID(providerID)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to map providerID to instanceID")
		return nil, "", err
	}

	instance, err := fcc.ociClient.Compute().GetInstance(ctx, instanceID)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to fetch instance")
		return nil, "", err
	}

	return instance, instanceID, nil
}
