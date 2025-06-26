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

package framework

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/storage/driver"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Config struct {
	UseInstancePrincipals bool   `yaml:"useInstancePrincipals"`
	Compartment           string `yaml:"compartment"`
	VCN                   string `yaml:"vcn"`
}

const (
	secretName       = "oci-volume-provisioner"
	secretNamespace  = "kube-system"
	k8sSecretKeyName = "config.yaml"
)

func InstallCustomDriver(clusterKubeconfigPath string, customHandle string, compartment string, vcn string) {
	config, err := clientcmd.BuildConfigFromFlags("", clusterKubeconfigPath)
	if err != nil {
		Failf("Failed to build kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		Failf("Failed to create Kubernetes client: %v", err)
	}

	// needed for snapshot sidecars, sidecars crash otherwise
	InstallSnapshotCRDs()
	CreateInstancePrincipalSecret(clientset, compartment, vcn)
	// controller pod manifest has node selector to only schedule on pods marked as control plane
	LabelNodesAsControlPlane(clientset)
	HelmInstall(clusterKubeconfigPath, customHandle)
}

func InstallSnapshotCRDs() {
	_, err := RunKubectl("apply", "-f", "https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/client/config/crd/snapshot.storage.k8s.io_volumesnapshotclasses.yaml")
	if err != nil {
		Failf("Failed to install snapshotclass CRD: %v", err)
	}
	_, err = RunKubectl("apply", "-f", "https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/client/config/crd/snapshot.storage.k8s.io_volumesnapshotcontents.yaml")
	if err != nil {
		Failf("Failed to install snapshotcontent CRD: %v", err)
	}
	_, err = RunKubectl("apply", "-f", "https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/client/config/crd/snapshot.storage.k8s.io_volumesnapshots.yaml")
	if err != nil {
		Failf("Failed to install snapshot CRD: %v", err)
	}
}

func HelmInstall(clusterKubeconfigPath string, customHandle string) {
	chartPath := "../../../manifests/container-storage-interface/csi"
	chart, err := loader.Load(chartPath)
	if err != nil {
		Failf("failed to load Helm chart: %v", err)
	}

	releaseName := "custom-bv-fss"
	releaseNamespace := "kube-system"

	helmCfg := new(action.Configuration)
	err = helmCfg.Init(
		kube.GetConfig(clusterKubeconfigPath, "", releaseNamespace),
		releaseNamespace,
		os.Getenv("HELM_DRIVER"),
		func(format string, v ...interface{}) {
			fmt.Sprintf(format, v...)
		},
	)
	if err != nil {
		Failf("failed to initialize Helm config: %v", err)
	}

	getter := action.NewGet(helmCfg)
	_, err = getter.Run(releaseName)
	if err == nil {
		Logf("Helm release %q already exists, skipping install", releaseName)
		return
	}

	if !errors.Is(err, driver.ErrReleaseNotFound) && !strings.Contains(err.Error(), "release: not found") {
		Failf("failed to check if Helm release exists: %v", err)
	}

	installer := action.NewInstall(helmCfg)
	installer.Namespace = releaseNamespace
	installer.ReleaseName = releaseName

	valuesMap := map[string]interface{}{
		"customHandle": customHandle,
	}

	release, err := installer.Run(chart, valuesMap)
	if err != nil {
		Failf("helm install failed: %v", err)
	}

	Logf("Helm release installed: %s", release.Name)
}

func CreateInstancePrincipalSecret(clientset *kubernetes.Clientset, compartment string, vcn string) {
	cfg := Config{
		UseInstancePrincipals: true,
		Compartment:           compartment,
		VCN:                   vcn,
	}

	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		Failf("Failed to marshal config to YAML: %v", err)
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: secretNamespace,
		},
		Data: map[string][]byte{
			k8sSecretKeyName: yamlData,
		},
		Type: v1.SecretTypeOpaque,
	}

	ctx := context.Background()
	_, err = clientset.CoreV1().Secrets(secretNamespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		_, err = clientset.CoreV1().Secrets(secretNamespace).Update(ctx, secret, metav1.UpdateOptions{})
		if err != nil {
			Failf("Failed to create or update secret: %v", err)
		}
		Logf("Updated existing secret: %s", secretName)
	} else {
		Logf("Created new secret: %s", secretName)
	}
}

func LabelNodesAsControlPlane(clientset *kubernetes.Clientset) {
	ctx := context.Background()

	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		Failf("Failed to list nodes: %v", err)
	}

	for _, node := range nodes.Items {
		labels := node.GetLabels()
		if labels == nil {
			labels = make(map[string]string)
		}
		labels["node-role.kubernetes.io/control-plane"] = ""
		node.SetLabels(labels)

		_, err := clientset.CoreV1().Nodes().Update(ctx, &node, metav1.UpdateOptions{})
		if err != nil {
			Failf("Failed to label node %s: %v", node.Name, err)
		} else {
			Logf("Labeled node %s as control-plane", node.Name)
		}
	}
}
