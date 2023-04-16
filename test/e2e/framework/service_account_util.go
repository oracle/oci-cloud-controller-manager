/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package framework

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// newServiceAccountTemplate returns the default v1.ServiceAccount template for this jig, but
// does not actually create the ServiceAccount. The default ServiceAccount has the same name
// as the jig.
func (j *ServiceTestJig) newServiceAccountTemplate(namespace, name string) *v1.ServiceAccount {
	serviceAccount := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels:    j.Labels,
		},
	}
	return serviceAccount
}

// CreateServiceAccountOrFail creates a new Service Account based on the default SA template
// in the namespace and with a name provided by the caller.
// Callers can provide a function to tweak the Service object before it is created.
func (j *ServiceTestJig) CreateServiceAccountOrFail(namespace, name string, tweak func(svc *v1.ServiceAccount)) *v1.ServiceAccount {
	sa := j.newServiceAccountTemplate(namespace, name)
	if tweak != nil {
		tweak(sa)
	}
	result, err := j.Client.CoreV1().ServiceAccounts(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err == nil {
		return result
	}

	result, err = j.Client.CoreV1().ServiceAccounts(namespace).Create(context.Background(), sa, metav1.CreateOptions{})
	if err != nil {
		Failf("Failed to create Service Account %q: %v", sa.Name, err)
	}
	return result
}
