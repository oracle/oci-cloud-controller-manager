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
	"testing"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

func TestProcessItemSkipsOCILookupsWhenNodeAlreadyHasCachedExpectedPodCIDRs(t *testing.T) {
	kubeClient := fake.NewSimpleClientset()
	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	nodeInformer := factory.Core().V1().Nodes()

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "worker-node-3"},
		Spec: corev1.NodeSpec{
			ProviderID: "oci://instance",
			PodCIDRs:   []string{"10.0.0.0/24", "2001:db8::/80"},
		},
	}
	if err := nodeInformer.Informer().GetStore().Add(node); err != nil {
		t.Fatalf("adding node to informer store: %v", err)
	}

	controller := &FlexCIDRController{
		nodeInformer:           nodeInformer,
		logger:                 zap.NewNop().Sugar(),
		expectedPodCIDRsByNode: make(map[string][]string),
	}
	controller.setExpectedPodCIDRs(node.Name, []string{"10.0.0.0/24", "2001:db8::/80"})

	if err := controller.processItem(node.Name); err != nil {
		t.Fatalf("processItem() error = %v, want nil", err)
	}
}

func TestDeleteExpectedPodCIDRsRemovesCachedValue(t *testing.T) {
	controller := &FlexCIDRController{
		expectedPodCIDRsByNode: make(map[string][]string),
	}

	controller.setExpectedPodCIDRs("worker-node-3", []string{"10.0.0.0/24"})
	controller.deleteExpectedPodCIDRs("worker-node-3")

	if _, ok := controller.getExpectedPodCIDRs("worker-node-3"); ok {
		t.Fatal("expected cached podCIDRs to be removed")
	}
}
