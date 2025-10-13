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

package oci

import (
	"context"
	"reflect"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	"github.com/oracle/oci-go-sdk/v65/core"
)

const (
	reconcileRetryingMinutes       = 240
	workerRestartIntervalinMinutes = 2
	openshiftTagNamespace          = "openshift-tags"
	openshiftTagKey                = "openshift-resource"
	openshiftTagValue              = "openshift-resource-infra"
	definedTagLimit                = 64
)

type TaggingController struct {
	nodeInformer coreinformers.NodeInformer
	logger       *zap.SugaredLogger
	kubeClient   clientset.Interface
	cloud        *CloudProvider
	queue        workqueue.RateLimitingInterface
	ociClient    client.Interface
}

// TaggingControllerRateLimiter enforces at most one retry every ten minutes for tagging controller work queues.
func TaggingControllerRateLimiter() workqueue.RateLimiter {
	tenMinuteDelay := 10 * time.Minute
	return workqueue.NewMaxOfRateLimiter(
		// Ensure each retry is at least 10 minutes apart regardless of failures
		workqueue.NewItemExponentialFailureRateLimiter(tenMinuteDelay, tenMinuteDelay),
		// Limit overall processing to one operation per 10 minutes
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Every(tenMinuteDelay), 1)},
	)
}

// NewTaggingController creates a TaggingController object
func NewTaggingController(
	nodeInformer coreinformers.NodeInformer,
	kubeClient kubernetes.Interface,
	cloud *CloudProvider,
	logger *zap.SugaredLogger,
	ociClient client.Interface) *TaggingController {

	tc := &TaggingController{
		nodeInformer: nodeInformer,
		kubeClient:   kubeClient,
		logger:       logger,
		cloud:        cloud,
		queue:        workqueue.NewRateLimitingQueue(TaggingControllerRateLimiter()),
		ociClient:    ociClient,
	}

	// Use shared informer to listen to add nodes
	tc.nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node := obj.(*v1.Node)
			tc.queue.Add(node.Name)
		},
		UpdateFunc: func(_, newObj interface{}) {
			node := newObj.(*v1.Node)
			tc.queue.Add(node.Name)
		},
	})

	return tc

}

func (tc *TaggingController) Run(stopCh <-chan struct{}) {
	defer tc.queue.ShutDown()
	tc.logger.Info("Starting tagging controller")
	wait.Until(func() {
		if err := tc.runWorker(stopCh); err != nil {
			tc.logger.Error(err, "runWorker error", "TaggingController")
		}
	}, workerRestartIntervalinMinutes*time.Minute, stopCh)

}

func (tc *TaggingController) runWorker(stopCh <-chan struct{}) error {
	nodeLister := tc.nodeInformer.Lister()

	ticker := time.NewTicker(reconcileRetryingMinutes * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			tc.logger.Info("Tagging Controller stopped")
			return nil
		case <-ticker.C:
			nodes, err := nodeLister.List(labels.Everything())
			if err != nil {
				tc.logger.Error("Failed to list nodes: %v", err)
				return err
			}
			for _, node := range nodes {
				tc.logger.Info("processing node: ", node.Name)
				tc.ReconcileNodeTags(context.Background(), node)
			}
		}
	}
}

// ReconcileNodeTags  retrieves instance details, merges required tags with existing defined tags,
// and calls the UpdateInstance API to apply the changes.
func (tc *TaggingController) ReconcileNodeTags(ctx context.Context, node *v1.Node) {

	if node == nil {
		tc.logger.Error("node is nil")
		return
	}
	tc.logger.Infow("Getting instanceOcid for node", "node", node.Name)
	instanceOCID, err := MapProviderIDToResourceID(node.Spec.ProviderID)
	if err != nil {
		tc.logger.Error("Failed to get/retrieve instanceOCID for node %s: %v", node.Name, err)
		return
	}
	logger := tc.logger.With("nodeName", node.Name, "instanceId", instanceOCID)

	logger.Info("Resolved instance OCID")
	instance, err := tc.ociClient.Compute().GetInstance(ctx, instanceOCID)
	if err != nil {
		logger.Errorw("Failed to get instance for node", "error", err)
		return
	}

	logger.Infof("Existing defined tags on node: %v", instance.DefinedTags)

	var t *config.TagConfig
	if tc.cloud.config.Tags == nil || tc.cloud.config.Tags.Common == nil {
		logger.Warnf("Tag config is nil; using empty TagConfig for node")
		t = &config.TagConfig{
			FreeformTags: map[string]string{},
			DefinedTags:  map[string]map[string]interface{}{},
		}
	} else {
		t = tc.cloud.config.Tags.Common
	}

	// Ensure the OpenShift defined tag namespace, key and value are present on the TagConfig.
	ns, ok := t.DefinedTags[openshiftTagNamespace]
	if !ok {
		t.DefinedTags[openshiftTagNamespace] = map[string]interface{}{
			openshiftTagKey: openshiftTagValue,
		}
	} else if val, ok := ns[openshiftTagKey]; !ok || !reflect.DeepEqual(val, openshiftTagValue) {
		ns[openshiftTagKey] = openshiftTagValue
	}

	logger.Infow(
		"Defined tags to reconcile on node",
		"definedTagsToReconcile", t.DefinedTags,
		"definedTagsExistingOnNode", instance.DefinedTags)

	if tc.hasRequiredDefinedTags(instance.DefinedTags, t.DefinedTags) {
		logger.Infof("Node already has required defined tags; skipping update")
		return
	}

	instanceTagConfig := &config.TagConfig{
		FreeformTags: instance.FreeformTags,
		DefinedTags:  instance.DefinedTags,
	}

	// MergeTags - Retrieve all defined tags currently set on the instance,
	// and merge them with the required tags that should be present.
	tags := util.MergeTagConfig(instanceTagConfig, t)

	// If the instance has already reached the defined tag limit, skip update to avoid API failure
	if countDefinedTags(instance.DefinedTags) >= definedTagLimit {
		logger.Warnf("Instance has %d defined tags which is the maximum allowed; cannot add required tags. Skipping update.", countDefinedTags(instance.DefinedTags))
		return
	}

	_, err = tc.ociClient.Compute().UpdateInstance(ctx, core.UpdateInstanceRequest{
		InstanceId: &instanceOCID,
		UpdateInstanceDetails: core.UpdateInstanceDetails{
			DefinedTags:  tags.DefinedTags,
			FreeformTags: tags.FreeformTags,
		},
	})
	if err != nil {
		logger.Error("Failed to update defined tags for node : %v", err)
		return
	}
	logger.Info("Successfully updated defined tags for node ")
}

// hasRequiredDefinedTags verifies that all required defined tags exist on the instance.
// This includes the OpenShift defined tags.
func (tc *TaggingController) hasRequiredDefinedTags(instanceDefinedTags, requiredDefinedTags map[string]map[string]interface{}) bool {
	if requiredDefinedTags == nil {
		return true
	}
	if len(requiredDefinedTags) > 0 {
		if len(instanceDefinedTags) == 0 {
			return false
		}
		for namespace, tags := range requiredDefinedTags {
			existingNamespace, ok := instanceDefinedTags[namespace]
			if !ok {
				return false
			}
			for key, value := range tags {
				existingValue, ok := existingNamespace[key]
				if !ok || !reflect.DeepEqual(existingValue, value) {
					return false
				}
			}
		}
	}
	return true
}

// countDefinedTags returns the total count of defined tag key-value pairs across all namespaces.
func countDefinedTags(tags map[string]map[string]interface{}) int {
	if len(tags) == 0 {
		return 0
	}
	total := 0
	for _, ns := range tags {
		total += len(ns)
	}
	return total
}
