// Copyright (C) 2026, Oracle and/or its affiliates.
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
	"reflect"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core"
	corev1informers "k8s.io/client-go/informers/core/v1"
	internalinterfaces "k8s.io/client-go/informers/internalinterfaces"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var nodeGroupVersionResource = corev1.SchemeGroupVersion.WithResource("nodes")

// NewNodeFilteredSharedInformerFactory creates a shared informer factory whose
// Node informer is restricted by requirements. Informers for all other
// resources remain unfiltered.
func NewNodeFilteredSharedInformerFactory(
	client kubernetes.Interface,
	resyncPeriod time.Duration,
	requirements string,
) (informers.SharedInformerFactory, error) {
	baseFactory := informers.NewSharedInformerFactory(client, resyncPeriod)
	requirements = strings.TrimSpace(requirements)
	if requirements == "" {
		return baseFactory, nil
	}

	selector, err := labels.Parse(requirements)
	if err != nil {
		return nil, err
	}

	nodeFactory := informers.NewSharedInformerFactoryWithOptions(
		client,
		resyncPeriod,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = selector.String()
		}),
	)

	return &nodeFilteredSharedInformerFactory{
		SharedInformerFactory: baseFactory,
		nodeFactory:           nodeFactory,
	}, nil
}

// nodeFilteredSharedInformerFactory delegates non-Node resources to the
// unfiltered factory and Nodes to the filtered factory.
type nodeFilteredSharedInformerFactory struct {
	informers.SharedInformerFactory
	nodeFactory informers.SharedInformerFactory
}

func (f *nodeFilteredSharedInformerFactory) Start(stopCh <-chan struct{}) {
	f.SharedInformerFactory.Start(stopCh)
	f.nodeFactory.Start(stopCh)
}

func (f *nodeFilteredSharedInformerFactory) Shutdown() {
	f.SharedInformerFactory.Shutdown()
	f.nodeFactory.Shutdown()
}

func (f *nodeFilteredSharedInformerFactory) WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool {
	synced := f.SharedInformerFactory.WaitForCacheSync(stopCh)
	for resourceType, result := range f.nodeFactory.WaitForCacheSync(stopCh) {
		synced[resourceType] = result
	}
	return synced
}

func (f *nodeFilteredSharedInformerFactory) ForResource(resource schema.GroupVersionResource) (informers.GenericInformer, error) {
	if resource == nodeGroupVersionResource {
		return f.nodeFactory.ForResource(resource)
	}
	return f.SharedInformerFactory.ForResource(resource)
}

func (f *nodeFilteredSharedInformerFactory) InformerFor(
	obj runtime.Object,
	newFunc internalinterfaces.NewInformerFunc,
) cache.SharedIndexInformer {
	if _, ok := obj.(*corev1.Node); ok {
		// A caller-provided newFunc may create an unfiltered Node informer, so
		// always return the canonical informer configured with the Node selector.
		return f.nodeFactory.Core().V1().Nodes().Informer()
	}
	return f.SharedInformerFactory.InformerFor(obj, newFunc)
}

func (f *nodeFilteredSharedInformerFactory) Core() coreinformers.Interface {
	return &nodeFilteredCoreInformers{
		Interface:  f.SharedInformerFactory.Core(),
		nodeCoreV1: f.nodeFactory.Core().V1(),
	}
}

type nodeFilteredCoreInformers struct {
	coreinformers.Interface
	nodeCoreV1 corev1informers.Interface
}

func (i *nodeFilteredCoreInformers) V1() corev1informers.Interface {
	return &nodeFilteredCoreV1Informers{
		Interface:  i.Interface.V1(),
		nodeCoreV1: i.nodeCoreV1,
	}
}

type nodeFilteredCoreV1Informers struct {
	corev1informers.Interface
	nodeCoreV1 corev1informers.Interface
}

func (i *nodeFilteredCoreV1Informers) Nodes() corev1informers.NodeInformer {
	return i.nodeCoreV1.Nodes()
}
