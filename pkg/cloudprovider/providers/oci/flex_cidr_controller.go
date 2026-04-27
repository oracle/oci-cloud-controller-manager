// Copyright 2026 Oracle and/or its affiliates. All rights reserved.
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
	"sync"
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
	nodeInformer           coreinformers.NodeInformer
	serviceInformer        coreinformers.ServiceInformer
	kubeClient             clientset.Interface
	cloud                  *CloudProvider
	queue                  workqueue.RateLimitingInterface
	logger                 *zap.SugaredLogger
	ociClient              client.Interface
	expectedPodCIDRsMu     sync.RWMutex
	expectedPodCIDRsByNode map[string][]string
}

func NewFlexCIDRController(
	nodeInformer coreinformers.NodeInformer,
	serviceInformer coreinformers.ServiceInformer,
	kubeClient clientset.Interface,
	cloud *CloudProvider,
	logger *zap.SugaredLogger,
	ociClient client.Interface) *FlexCIDRController {

	controller := &FlexCIDRController{
		nodeInformer:           nodeInformer,
		serviceInformer:        serviceInformer,
		kubeClient:             kubeClient,
		cloud:                  cloud,
		queue:                  workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		logger:                 logger,
		ociClient:              ociClient,
		expectedPodCIDRsByNode: make(map[string][]string),
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
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err != nil {
				controller.logger.With(zap.Error(err)).Debug("failed to determine deleted node cache key")
				return
			}
			controller.deleteExpectedPodCIDRs(key)
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

	if expectedPodCIDRs, ok := fcc.getExpectedPodCIDRs(node.Name); ok && flexcidr.StringSlicesEqualIgnoreOrder(node.Spec.PodCIDRs, expectedPodCIDRs) {
		logger.Debugf("node already has cached expected podCIDRs %v", expectedPodCIDRs)
		return nil
	}

	instance, instanceID, err := fcc.getInstanceByNode(node, logger)
	if err != nil {
		return err
	}
	if instance == nil {
		return nil
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
	fcc.setExpectedPodCIDRs(node.Name, flexCIDRs)
	if flexcidr.StringSlicesEqualIgnoreOrder(node.Spec.PodCIDRs, flexCIDRs) {
		logger.Debugf("node already has expected podCIDRs %v", flexCIDRs)
		return nil
	}

	return flexcidr.PatchNodePodCIDRs(context.Background(), fcc.kubeClient, node.Name, flexCIDRs, logger)
}

func (fcc *FlexCIDRController) getExpectedPodCIDRs(nodeName string) ([]string, bool) {
	fcc.expectedPodCIDRsMu.RLock()
	defer fcc.expectedPodCIDRsMu.RUnlock()

	podCIDRs, ok := fcc.expectedPodCIDRsByNode[nodeName]
	if !ok {
		return nil, false
	}
	return append([]string(nil), podCIDRs...), true
}

func (fcc *FlexCIDRController) setExpectedPodCIDRs(nodeName string, podCIDRs []string) {
	fcc.expectedPodCIDRsMu.Lock()
	defer fcc.expectedPodCIDRsMu.Unlock()

	fcc.expectedPodCIDRsByNode[nodeName] = append([]string(nil), podCIDRs...)
}

func (fcc *FlexCIDRController) deleteExpectedPodCIDRs(nodeName string) {
	fcc.expectedPodCIDRsMu.Lock()
	defer fcc.expectedPodCIDRsMu.Unlock()

	delete(fcc.expectedPodCIDRsByNode, nodeName)
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
