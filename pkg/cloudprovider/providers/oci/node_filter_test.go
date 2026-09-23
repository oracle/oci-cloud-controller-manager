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
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

func TestNewNodeFilteredSharedInformerFactory(t *testing.T) {
	ociNode := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "oci-node",
			Labels: map[string]string{
				"platform": "oci",
			},
		},
	}
	otherNode := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "other-node",
			Labels: map[string]string{
				"platform": "other",
			},
		},
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "load-balancer",
			Namespace: "default",
		},
	}
	kubeClient := fake.NewSimpleClientset(ociNode, otherNode, service)

	factory, err := NewNodeFilteredSharedInformerFactory(
		kubeClient,
		0,
		"platform=oci",
	)
	if err != nil {
		t.Fatalf("NewNodeFilteredSharedInformerFactory returned an error: %v", err)
	}

	nodeInformer := factory.Core().V1().Nodes()
	serviceInformer := factory.Core().V1().Services()
	nodeSharedInformer := nodeInformer.Informer()
	serviceSharedInformer := serviceInformer.Informer()
	stopCh := make(chan struct{})
	t.Cleanup(func() {
		close(stopCh)
	})
	factory.Start(stopCh)

	if !cache.WaitForCacheSync(
		stopCh,
		nodeSharedInformer.HasSynced,
		serviceSharedInformer.HasSynced,
	) {
		t.Fatal("timed out waiting for informer caches to sync")
	}

	nodes, err := nodeInformer.Lister().List(labels.Everything())
	if err != nil {
		t.Fatalf("failed to list Nodes: %v", err)
	}
	if len(nodes) != 1 || nodes[0].Name != ociNode.Name {
		t.Fatalf("expected only %q, got %#v", ociNode.Name, nodeNames(nodes))
	}

	services, err := serviceInformer.Lister().List(labels.Everything())
	if err != nil {
		t.Fatalf("failed to list Services: %v", err)
	}
	if len(services) != 1 || services[0].Name != service.Name {
		t.Fatalf("expected the Node selector not to filter Services, got %d Services", len(services))
	}
}

func TestNewNodeFilteredSharedInformerFactoryInformerForMaintainsNodeFilter(t *testing.T) {
	ociNode := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "oci-node",
			Labels: map[string]string{
				"platform": "oci",
			},
		},
	}
	otherNode := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "other-node",
			Labels: map[string]string{
				"platform": "other",
			},
		},
	}
	kubeClient := fake.NewSimpleClientset(ociNode, otherNode)

	factory, err := NewNodeFilteredSharedInformerFactory(kubeClient, 0, "platform=oci")
	if err != nil {
		t.Fatalf("NewNodeFilteredSharedInformerFactory returned an error: %v", err)
	}

	directInformer := factory.InformerFor(
		&corev1.Node{},
		func(client kubernetes.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
			return corev1informers.NewNodeInformer(
				client,
				resyncPeriod,
				cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
			)
		},
	)
	nodeInformer := factory.Core().V1().Nodes()
	typedInformer := nodeInformer.Informer()
	if directInformer != typedInformer {
		t.Fatal("expected direct and typed Node informer accessors to share an informer")
	}

	stopCh := make(chan struct{})
	t.Cleanup(func() {
		close(stopCh)
	})
	factory.Start(stopCh)

	if !cache.WaitForCacheSync(stopCh, directInformer.HasSynced) {
		t.Fatal("timed out waiting for informer cache to sync")
	}

	nodes, err := nodeInformer.Lister().List(labels.Everything())
	if err != nil {
		t.Fatalf("failed to list Nodes: %v", err)
	}
	if len(nodes) != 1 || nodes[0].Name != ociNode.Name {
		t.Fatalf("expected only %q, got %#v", ociNode.Name, nodeNames(nodes))
	}
}

func TestNewNodeFilteredSharedInformerFactoryEmptyRequirements(t *testing.T) {
	kubeClient := fake.NewSimpleClientset(
		&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-one"}},
		&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-two"}},
	)

	factory, err := NewNodeFilteredSharedInformerFactory(kubeClient, 0, " ")
	if err != nil {
		t.Fatalf("NewNodeFilteredSharedInformerFactory returned an error: %v", err)
	}

	nodeInformer := factory.Core().V1().Nodes()
	nodeSharedInformer := nodeInformer.Informer()
	stopCh := make(chan struct{})
	t.Cleanup(func() {
		close(stopCh)
	})
	factory.Start(stopCh)

	if !cache.WaitForCacheSync(stopCh, nodeSharedInformer.HasSynced) {
		t.Fatal("timed out waiting for informer cache to sync")
	}

	nodes, err := nodeInformer.Lister().List(labels.Everything())
	if err != nil {
		t.Fatalf("failed to list Nodes: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected an empty selector to include both Nodes, got %d", len(nodes))
	}
}

func TestNewNodeFilteredSharedInformerFactoryInvalidRequirements(t *testing.T) {
	kubeClient := fake.NewSimpleClientset()

	_, err := NewNodeFilteredSharedInformerFactory(kubeClient, time.Minute, "platform in (")
	if err == nil {
		t.Fatal("expected invalid Node filter requirements to return an error")
	}
}

func nodeNames(nodes []*corev1.Node) []string {
	names := make([]string, 0, len(nodes))
	for _, node := range nodes {
		names = append(names, node.Name)
	}
	return names
}
