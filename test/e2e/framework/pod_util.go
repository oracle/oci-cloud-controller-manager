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
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubernetes/pkg/client/conditions"
)

// CheckVolumeReadWrite creates a pod with a dymincally provisioned volume
func (j *PVCTestJig) CheckVolumeReadWrite(namespace string, pvcParam *v1.PersistentVolumeClaim) {
	pvc, err := j.KubeClient.CoreV1().PersistentVolumeClaims(pvcParam.Namespace).Get(pvcParam.Name, metav1.GetOptions{})
	pv, err := j.KubeClient.CoreV1().PersistentVolumes().Get(pvc.Spec.VolumeName, metav1.GetOptions{})
	if err != nil {
		Failf("Failed to get persistent volume %q: %v", pvc.Spec.VolumeName, err)
	}
	By("checking the created volume is writable and has the PV's mount options")
	command := "while true; do echo 'hello world' >> /usr/share/nginx/html/out.txt; sleep 5; done"
	// We give the first pod the secondary responsibility of checking the volume has
	// been mounted with the PV's mount options, if the PV was provisioned with any
	for _, option := range pv.Spec.MountOptions {
		// Get entry, get mount options at 6th word, replace brackets with commas
		command += fmt.Sprintf(" && ( mount | grep 'on /usr/share/nginx/html/out.txt' | awk '{print $6}' | sed 's/^(/,/; s/)$/,/' | grep -q ,%s, )", option)
	}
	j.CreateAndAwaitNginxPodOrFail(pvc.Namespace, pvc, command)

	By("checking the created volume is readable and retains data")
	j.CreateAndAwaitNginxPodOrFail(pvc.Namespace, pvc, "grep 'hello world' /mnt/test/data")
}

// CreateAndAwaitNginxPodOrFail returns a pod definition based on the namespace using nginx image
func (j *PVCTestJig) CreateAndAwaitNginxPodOrFail(ns string, pvc *v1.PersistentVolumeClaim, command string) {
	By("Creating a pod with the dynamically provisioned volume")
	pod, err := j.KubeClient.CoreV1().Pods(ns).Create(&v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "pvc-pod-tester-" + pvc.Labels["testID"],
			Namespace:    ns,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "write-pod",
					Image: nginx,
					Ports: []v1.ContainerPort{
						{
							Name:          "http-server",
							ContainerPort: 80,
						},
					},
					Command: []string{"/bin/sh"},
					Args:    []string{"-c", command},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "nginx-storage",
							MountPath: "/usr/share/nginx/html/",
						},
					},
				},
			},
			Volumes: []v1.Volume{
				{
					Name: "nginx-storage",
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvc.Name,
						},
					},
				},
			},
		},
	})
	if err != nil {
		Failf("Pod %q Create API error: %v", pod.Name, err)
	}

	// Waiting for pod to be running
	err = j.waitTimeoutForPodRunningInNamespace(pod.Name, ns, slowPodStartTimeout)
	if err != nil {
		Failf("Pod %q is not Running: %v", pod.Name, err)
	}
}

// WaitTimeoutForPodRunningInNamespace aits default amount of time (PodStartTimeout) for the specified pod to become running.
// Returns an error if timeout occurs first, or pod goes in to failed state.
func (j *PVCTestJig) waitTimeoutForPodRunningInNamespace(podName, namespace string, timeout time.Duration) error {
	return wait.PollImmediate(Poll, timeout, j.podRunning(podName, namespace))
}

func (j *PVCTestJig) podRunning(podName, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		pod, err := j.KubeClient.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		switch pod.Status.Phase {
		case v1.PodRunning:
			return true, nil
		case v1.PodFailed, v1.PodSucceeded:
			return false, conditions.ErrPodCompleted
		}
		return false, nil
	}
}
