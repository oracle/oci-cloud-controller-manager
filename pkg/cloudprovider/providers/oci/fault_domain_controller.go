package oci

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
	"github.com/oracle/oci-go-sdk/core"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

const MaxRetries = 20
const FaultDomainLabel = "oci.oraclecloud.com/fault-domain"

type FaultDomainController struct {
	nodeInformer coreinformers.NodeInformer
	kubeClient   clientset.Interface
	recorder     record.EventRecorder
	cloud        *CloudProvider
	queue        workqueue.RateLimitingInterface
}

// NewFaultDomainController creates a FaultDomainController object
func NewFaultDomainController(
	nodeInformer coreinformers.NodeInformer,
	kubeClient clientset.Interface,
	cloud *CloudProvider) *FaultDomainController {

	eventBroadcaster := record.NewBroadcaster()
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "fault-domain-controller"})
	eventBroadcaster.StartLogging(glog.Infof)
	if kubeClient != nil {
		cloud.logger.Info("Sending events to api server.")
		eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	} else {
		cloud.logger.Info("No api server defined - no events will be sent to API server.")
	}

	fdc := &FaultDomainController{
		nodeInformer: nodeInformer,
		kubeClient:   kubeClient,
		recorder:     recorder,
		cloud:        cloud,
		queue:        workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}

	// Use shared informer to listen to add nodes
	fdc.nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node := obj.(*v1.Node)
			fdc.queue.Add(node.Name)
		},
		UpdateFunc: func(_,newObj interface{}) {
			node := newObj.(*v1.Node)
			fdc.queue.Add(node.Name)
		},
	})

	return fdc
}

func (fdc *FaultDomainController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()

	defer fdc.queue.ShutDown()

	fdc.cloud.logger.Info("Starting fault domain controller")

	if !cache.WaitForCacheSync(stopCh, fdc.nodeInformer.Informer().HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	wait.Until(fdc.runWorker, time.Second, stopCh)
}

//A function to run the worker which will process items in the queue
func (fdc *FaultDomainController) runWorker() {
	for fdc.processNextItem() {

	}
}

//Used to sequentially process the keys present in the queue
func (fdc *FaultDomainController) processNextItem() bool {

	key, quit := fdc.queue.Get()
	if quit {
		return false
	}

	defer fdc.queue.Done(key)

	err := fdc.processItem(key.(string))

	if err == nil {
		fdc.queue.Forget(key)
	} else if fdc.queue.NumRequeues(key) < MaxRetries {
		fdc.cloud.logger.Errorf("Error processing node %s (will retry): %v", key, err)
		fdc.queue.AddRateLimited(key)
	} else {
		fdc.cloud.logger.Errorf("Error processing node %s (giving up): %v", key, err)
		fdc.queue.Forget(key)
		utilruntime.HandleError(err)
	}
	return true
}

//A function which is responsible for adding the fault domain label to the node if it is not already present
func (fdc *FaultDomainController) processItem(key string) error {
	cacheNode, err := fdc.nodeInformer.Lister().Get(key)

	if err != nil {
		return err
	}

	_, present := cacheNode.ObjectMeta.Labels[FaultDomainLabel]
	if present {
		fdc.cloud.logger.Debugf("The node %s has fault domain label already, will not process",cacheNode.Name)
	}else{
		curNode := cacheNode.DeepCopy()

		var instanceID string
		var instance *core.Instance
		instanceID, err = fdc.cloud.InstanceID(context.TODO(), types.NodeName(curNode.Name))
		if err != nil {
			fdc.cloud.logger.With(zap.Error(err)).Error("Failed to map provider ID to instance ID")
			return err
		}
		instance, err = fdc.cloud.client.Compute().GetInstance(context.TODO(), instanceID)
		if err != nil {
			fdc.cloud.logger.With(zap.Error(err)).Error("Failed to get instance from instance ID")
			return err
		}

		fdc.cloud.logger.Infof("Adding node label from cloud provider: %s=%s", FaultDomainLabel, *instance.FaultDomain)
		curNode.ObjectMeta.Labels[FaultDomainLabel] = *instance.FaultDomain

		_, err = fdc.kubeClient.CoreV1().Nodes().Update(curNode)
		if err != nil {
			return err
		}
	}
	return nil
}
