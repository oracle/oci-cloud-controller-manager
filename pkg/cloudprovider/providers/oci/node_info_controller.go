// Copyright 2020 Oracle and/or its affiliates. All rights reserved.
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

package oci

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v50/core"
)

// metadata labeling for placement info
const (
	FaultDomainLabel        = "oci.oraclecloud.com/fault-domain"
	CompartmentIDAnnotation = "oci.oraclecloud.com/compartment-id"
	timeout                 = 10 * time.Second
)

// NodeInfoController helps compute workers in the cluster
type NodeInfoController struct {
	nodeInformer  coreinformers.NodeInformer
	kubeClient    clientset.Interface
	recorder      record.EventRecorder
	cloud         *CloudProvider
	queue         workqueue.RateLimitingInterface
	logger        *zap.SugaredLogger
	instanceCache cache.Store
	ociClient     client.Interface
}

// NewNodeInfoController creates a NodeInfoController object
func NewNodeInfoController(
	nodeInformer coreinformers.NodeInformer,
	kubeClient clientset.Interface,
	cloud *CloudProvider,
	logger *zap.SugaredLogger,
	instanceCache cache.Store,
	ociClient client.Interface) *NodeInfoController {

	eventBroadcaster := record.NewBroadcaster()
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "node-info-controller"})
	eventBroadcaster.StartLogging(klog.Infof)
	if kubeClient != nil {
		cloud.logger.Info("Sending events to api server.")
		eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	} else {
		cloud.logger.Info("No api server defined - no events will be sent to API server.")
	}

	nic := &NodeInfoController{
		nodeInformer:  nodeInformer,
		kubeClient:    kubeClient,
		recorder:      recorder,
		cloud:         cloud,
		queue:         workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		logger:        logger,
		instanceCache: instanceCache,
		ociClient:     ociClient,
	}

	// Use shared informer to listen to add nodes
	nic.nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node := obj.(*v1.Node)
			nic.queue.Add(node.Name)
		},
		UpdateFunc: func(_, newObj interface{}) {
			node := newObj.(*v1.Node)
			nic.queue.Add(node.Name)
		},
	})

	return nic
}

// Run will start the NodeInfoController and manage shutdown
func (nic *NodeInfoController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()

	defer nic.queue.ShutDown()

	nic.logger.Info("Starting node info controller")

	if !cache.WaitForCacheSync(stopCh, nic.nodeInformer.Informer().HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	wait.Until(nic.runWorker, time.Second, stopCh)
}

//A function to run the worker which will process items in the queue
func (nic *NodeInfoController) runWorker() {
	for nic.processNextItem() {

	}
}

//Used to sequentially process the keys present in the queue
func (nic *NodeInfoController) processNextItem() bool {

	key, quit := nic.queue.Get()
	if quit {
		return false
	}

	defer nic.queue.Done(key)

	err := nic.processItem(key.(string))

	if err != nil {
		nic.logger.Errorf("Error processing node %s (will retry): %v", key, err)
		nic.queue.AddRateLimited(key)
	} else {
		nic.queue.Forget(key)
	}
	return true
}

//A function which is responsible for adding the fault domain label and CompartmentID annotation to the node if it
//is not already present. Also cache the instance information
func (nic *NodeInfoController) processItem(key string) error {

	logger := nic.logger.With("node", key)

	cacheNode, err := nic.nodeInformer.Lister().Get(key)

	if err != nil {
		return err
	}

	// if node has required labels already, don't process agin
	if validateNodeHasRequiredLabels(cacheNode) {
		logger.With("nodeName", cacheNode.Name).Debugf("The node has the fault domain label and compartmentID annotation already, will not process")
		return nil
	}

	instance, err := getInstanceByNode(cacheNode, nic, logger)
	if err != nil {
		return err
	}

	if err := nic.instanceCache.Add(instance); err != nil {
		logger.With(zap.Error(err)).Error("Failed to add instance in instanceCache")
		return err
	}

	nodePatchBytes := getPatchBytes(cacheNode, instance, logger)

	if nodePatchBytes == nil {
		return nil
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		_, err := nic.kubeClient.CoreV1().Nodes().Patch(context.Background(), cacheNode.Name, types.StrategicMergePatchType, nodePatchBytes, metav1.PatchOptions{})
		return err
	})
	if err != nil {
		logger.With(zap.Error(err)).Error("Error in applying patch in node %v", err)
		return err
	}

	return nil
}

func getPatchBytes(cacheNode *v1.Node, instance *core.Instance, logger *zap.SugaredLogger) []byte {
	if validateNodeHasRequiredLabels(cacheNode) {
		return nil
	}
	_, isFaultDomainLabelPresent := cacheNode.ObjectMeta.Labels[FaultDomainLabel]
	_, isCompartmentIDAnnotationPresent := cacheNode.ObjectMeta.Annotations[CompartmentIDAnnotation]

	var nodePatchBytes []byte
	if isFaultDomainLabelPresent {
		//In this case CompartmentIDAnnotation not present but FaultDomainLabel present
		logger.Infof("Adding node annotation from cloud provider: %s=%s", CompartmentIDAnnotation, *instance.CompartmentId)
		nodePatchBytes = []byte(fmt.Sprintf("{\"metadata\": {\"annotations\": {\"%s\":\"%s\"}}}",
			CompartmentIDAnnotation, *instance.CompartmentId))
	} else if isCompartmentIDAnnotationPresent {
		//In this case FaultDomainLabel not present but CompartmentIDAnnotation present
		logger.Infof("Adding node label from cloud provider: %s=%s", FaultDomainLabel, *instance.FaultDomain)
		nodePatchBytes = []byte(fmt.Sprintf("{\"metadata\": {\"labels\": {\"%s\":\"%s\"}}}", FaultDomainLabel, *instance.FaultDomain))
	} else {
		//In this case none of FaultDomainLabel or CompartmentIDAnnotation present
		logger.Infof("Adding node label from cloud provider: %s=%s", FaultDomainLabel, *instance.FaultDomain)
		logger.Infof("Adding node annotation from cloud provider: %s=%s", CompartmentIDAnnotation, *instance.CompartmentId)
		nodePatchBytes = []byte(fmt.Sprintf("{\"metadata\": {\"labels\": {\"%s\":\"%s\"},\"annotations\": {\"%s\":\"%s\"}}}",
			FaultDomainLabel, *instance.FaultDomain, CompartmentIDAnnotation, *instance.CompartmentId))
	}
	return nodePatchBytes
}

func getInstanceByNode(cacheNode *v1.Node, nic *NodeInfoController, logger *zap.SugaredLogger) (*core.Instance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	providerID := cacheNode.Spec.ProviderID
	var err error
	if providerID == "" {
		providerID, err = nic.cloud.InstanceID(ctx, types.NodeName(cacheNode.Name))
		if err != nil {
			logger.With(zap.Error(err)).Error("Failed to map provider ID to instance ID")
			return nil, err
		}
	}

	instanceID, err := MapProviderIDToInstanceID(providerID)
	if err != nil {
		logger.With(zap.Error(err)).Error("Failed to map providerID to instanceID")
		return nil, err
	}
	instance, err := nic.ociClient.Compute().GetInstance(ctx, instanceID)
	if err != nil {
		logger.With(zap.Error(err)).Error("Failed to get instance from instance ID")
		return nil, err
	}
	return instance, nil
}

func validateNodeHasRequiredLabels(node *v1.Node) bool {
	_, isFaultDomainLabelPresent := node.ObjectMeta.Labels[FaultDomainLabel]
	_, isCompartmentIDAnnotationPresent := node.ObjectMeta.Annotations[CompartmentIDAnnotation]
	if isFaultDomainLabelPresent && isCompartmentIDAnnotationPresent {
		return true
	}
	return false
}
