// Copyright 2018 Oracle and/or its affiliates. All rights reserved.
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
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	v1 "k8s.io/api/core/v1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (f *Framework) createSecret(secretName string) string {
	content, err := ioutil.ReadFile(TestContext.OCIConfig)
	if err != nil {
		Failf("Failed to read the ociconfig file")
	}

	Logf("Creating secret %q", types.NamespacedName{Namespace: f.Namespace.Name, Name: secretName})
	s := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: f.Namespace.Name,
		},
		Data: map[string][]byte{
			"config.yaml": content,
		},
		Type: v1.SecretTypeOpaque,
	}
	_, err = f.ClientSet.CoreV1().Secrets(f.Namespace.Name).Create(s)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			Logf("Secret %q already exists. Upgrading secret.", s.Name)
			_, err = f.ClientSet.CoreV1().Secrets(f.Namespace.Name).Update(s)
			if err != nil {
				Logf("Failed to upgrade secret %q: %v", s.Name, err)
			}
		} else {
			Failf("Failed to create secret: %v", err)
		}
	}
	return s.Name
}

func (f *Framework) createClusterRole() string {
	cr := &rbacv1beta1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "oci-provisioner-runner",
		},
		Rules: []rbacv1beta1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumes"},
				Verbs:     []string{"get", "list", "watch", "create", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumeclaims"},
				Verbs:     []string{"get", "list", "watch", "update", "create"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"storageclasses"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"list", "watch", "create", "update", "patch"},
			},
		},
	}
	_, err := f.ClientSet.RbacV1beta1().ClusterRoles().Create(cr)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			Logf("ClusterRole %q already exists. Upgrading cluster role.", cr.Name)
			_, err = f.ClientSet.RbacV1beta1().ClusterRoles().Update(cr)
			if err != nil {
				Logf("Failed to upgrade cluster role %q: %v", cr.Name, err)
			}
		} else {
			Failf("Failed to create cluster role: %v", err)
		}
	}
	return cr.Name
}

func (f *Framework) createClusterRoleBinding(sa string, clusterRole string) string {
	cr := &rbacv1beta1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "run-oracle-provisioner",
		},
		Subjects: []rbacv1beta1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      sa,
				Namespace: f.Namespace.Name,
			},
		},
		RoleRef: rbacv1beta1.RoleRef{
			Name:     "oci-provisioner-runner",
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	_, err := f.ClientSet.RbacV1beta1().ClusterRoleBindings().Create(cr)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			Logf("ClusterRoleBinding %q already exists. Upgrading cluster role binding.", cr.Name)
			_, err = f.ClientSet.RbacV1beta1().ClusterRoleBindings().Update(cr)
			if err != nil {
				Logf("Failed to upgrade cluster role binding %q: %v", cr.Name, err)
			}
		} else {
			Failf("Failed to create cluster role binding: %v", err)
		}
	}
	return cr.Name
}

func (f *Framework) createServiceAccount() string {

	sa := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "oci-volume-provisioner",
			Namespace: f.Namespace.Name,
		},
	}
	_, err := f.ClientSet.CoreV1().ServiceAccounts(f.Namespace.Name).Create(sa)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			Logf("Service account %q already exists. Upgrading service account.", sa.Name)
			_, err = f.ClientSet.CoreV1().ServiceAccounts(f.Namespace.Name).Update(sa)
			if err != nil {
				Logf("Failed to upgrade service account %q: %v", sa.Name, err)
			}
		} else {
			Failf("Failed to create service account: %v", sa)
		}
	}
	return sa.Name
}

func (f *Framework) createProvisioner(secretName string, serviceAccountName string, provName string, provisionerType string) {
	secret, err := f.ClientSet.CoreV1().Secrets(f.Namespace.Name).Get(secretName, metav1.GetOptions{})
	if err != nil {
		Failf("Failed to get secret: %v", err)
	}
	serviceAccount, err := f.ClientSet.CoreV1().ServiceAccounts(f.Namespace.Name).Get(serviceAccountName, metav1.GetOptions{})
	if err != nil {
		Failf("Failed to get service account: %v", err)
	}
	Logf("Installing provisioner")
	replica := int32(1)
	mounts := []v1.VolumeMount{
		{
			Name:      "config",
			MountPath: "/etc/oci/",
			ReadOnly:  true,
		},
	}
	volumes := []v1.Volume{
		{
			Name: "config",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: secret.Name,
				},
			},
		},
	}

	provisionerPodSpec := v1.PodSpec{
		ServiceAccountName: serviceAccount.Name,
		Containers: []v1.Container{
			{
				Name:            secret.Name,
				Image:           TestContext.Image,
				Command:         []string{"/usr/local/bin/oci-volume-provisioner"},
				ImagePullPolicy: v1.PullAlways,
				Env: []v1.EnvVar{
					{
						Name: "NODE_NAME",
						ValueFrom: &v1.EnvVarSource{
							FieldRef: &v1.ObjectFieldSelector{
								FieldPath: "spec.nodeName",
							},
						},
					},
					{
						Name:  "PROVISIONER_TYPE",
						Value: provisionerType,
					},
				},
				VolumeMounts: mounts,
			},
		},
		Volumes: volumes,
	}
	provisionerDeployment := &appsv1beta1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      provName,
			Namespace: f.Namespace.Name,
		},
		Spec: appsv1beta1.DeploymentSpec{
			Replicas: &replica,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": secret.Name,
					},
				},
				Spec: provisionerPodSpec,
			},
		},
	}

	deploymentCreate, err := f.ClientSet.AppsV1beta1().Deployments(f.Namespace.Name).Create(provisionerDeployment)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			Logf("Provisioner already exists.")
		} else {
			Failf("Failed to create %s deployment: %v", deploymentCreate.Name, err)
		}
	}
	Logf("Created deployment %s in namespace %s", deploymentCreate.Name, deploymentCreate.Namespace)

}

// CheckandInstallProvisioner checks if a provisioner is installed, if installed uses the following provisioner.
// Otherwise, creates a provisioner in the test namespace such that after the test it can go back to its original state.
func (f *Framework) CheckandInstallProvisioner(provisionerName string, provisionerType string) bool {
	Logf("Installing block provisioner")
	_, err := f.ClientSet.AppsV1beta1().Deployments(KubeSystemNS).Get(provisionerName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			Logf("Provisioner is not installed")
			By("Creating Secret")
			s := f.createSecret(SecretNameDefault)
			By("Configuring RBAC and service account")
			cr := f.createClusterRole()
			f.createClusterRoleBinding(s, cr)
			sa := f.createServiceAccount()
			By("Installing OCI volume provisioner")
			f.createProvisioner(s, sa, provisionerName, provisionerType)
			return true
		}
		Failf("Failed to get %q provisioner: %v", provisionerName, err)
		return false
	}
	return true
}
