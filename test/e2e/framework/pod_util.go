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
	"context"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/pkg/errors"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubernetes/pkg/client/conditions"
)

// CheckVolumeMount creates a pod with a dynamically provisioned volume
func (j *PVCTestJig) CheckVolumeMount(namespace string, pvcParam *v1.PersistentVolumeClaim) {
	pvc, err := j.KubeClient.CoreV1().PersistentVolumeClaims(pvcParam.Namespace).Get(context.Background(), pvcParam.Name, metav1.GetOptions{})
	pv, err := j.KubeClient.CoreV1().PersistentVolumes().Get(context.Background(), pvc.Spec.VolumeName, metav1.GetOptions{})
	if err != nil {
		Failf("Failed to get persistent volume %q: %v", pvc.Spec.VolumeName, err)
	}
	By("checking the created volume is writable and has the PV's mount options")
	command := "while true; do echo 'hello world' >> /usr/share/nginx/html/out.txt; sleep 5; done"
	// We give the first pod the responsibility of checking the volume has
	// been mounted with the PV's mount options
	for _, option := range pv.Spec.MountOptions {
		// Get entry, get mount options at 6th word, replace brackets with commas
		command += fmt.Sprintf(" && ( mount | grep 'on /usr/share/nginx/html/out.txt' | awk '{print $6}' | sed 's/^(/,/; s/)$/,/' | grep -q ,%s, )", option)
	}
	podName := j.CreateAndAwaitNginxPodOrFail(pvc.Namespace, pvc, command)

	j.CheckFileExists(namespace, podName, "/usr/share/nginx/html", "out.txt")

	By("Wait for pod with dynamically provisioned volume to be deleted")
	j.DeleteAndAwaitNginxPodOrFail(namespace, podName)

	command = "while true; do sleep 5; done"
	By("Recreating a pod with the same dynamically provisioned volume and waiting for it to be running")
	podName = j.CreateAndAwaitNginxPodOrFail(pvc.Namespace, pvc, command)

	By("Checking if the file exists on the newly created pod")
	j.CheckFileExists(namespace, podName, "/usr/share/nginx/html", "out.txt")
}

// DeleteAndAwaitNginxPodOrFail deletes the pod definition based on the namespace and waits for pod to disappear
func (j *PVCTestJig) DeleteAndAwaitNginxPodOrFail(ns string, podName string) {
	err := j.KubeClient.CoreV1().Pods(ns).Delete(context.Background(), podName, metav1.DeleteOptions{})
	if err != nil {
		Failf("Pod %q Delete API error: %v", podName, err)
	}

	err = j.waitTimeoutForPodNotFoundInNamespace(podName, ns, DefaultTimeout)
	if err != nil {
		Failf("Pod %q is not deleted: %v", podName, err)
	}
}

func (j *PVCTestJig) CheckFileExists(namespace string, podName string, dir string, fileName string) {
	By("check if the file exists")
	command := fmt.Sprintf("ls %s", dir)
	if pollErr := wait.PollImmediate(K8sResourcePoll, DefaultTimeout, func() (bool, error) {
		stdout, err := RunHostCmd(namespace, podName, command)
		if err != nil {
			Logf("got err: %v, retry until timeout", err)
			return false, nil
		}
		return strings.Contains(stdout, fileName), nil
	}); pollErr != nil {
		Failf("File does not exist in pod '%v'", podName)
	}
}

func (j *PVCTestJig) CheckFileCorruption(namespace string, podName string, dir string, fileName string) {
	By("check if the file is corrupt")
	md5hash := "e59ff97941044f85df5297e1c302d260"
	command := fmt.Sprintf("md5sum %s/%s", dir, fileName)
	if pollErr := wait.PollImmediate(K8sResourcePoll, DefaultTimeout, func() (bool, error) {
		stdout, err := RunHostCmd(namespace, podName, command)
		if err != nil {
			Logf("got err: %v, retry until timeout", err)
			return false, nil
		}
		return strings.Contains(stdout, md5hash), nil
	}); pollErr != nil {
		Failf("MD5 hash does not match, file is corrupt in pod '%v'", podName)
	}
}

// CheckVolumeReadWrite creates a pod with a dynamically provisioned volume
func (j *PVCTestJig) CheckVolumeReadWrite(namespace string, pvcParam *v1.PersistentVolumeClaim) {
	pvc, err := j.KubeClient.CoreV1().PersistentVolumeClaims(pvcParam.Namespace).Get(context.Background(), pvcParam.Name, metav1.GetOptions{})
	pv, err := j.KubeClient.CoreV1().PersistentVolumes().Get(context.Background(), pvc.Spec.VolumeName, metav1.GetOptions{})
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
	podName := j.CreateAndAwaitNginxPodOrFail(pvc.Namespace, pvc, command)

	By("Delete the pod to which volume is already attached")
	j.DeleteAndAwaitNginxPodOrFail(pvc.Namespace, podName)

	By("checking the created volume is readable and retains data")
	j.CreateAndAwaitNginxPodOrFail(pvc.Namespace, pvc, "grep 'hello world' /mnt/test/data")
}

func (j *PVCTestJig) checkFileOwnership(namespace string, podName string, dir string) {
	By("check if the file exists")
	command := fmt.Sprintf("stat -c '%%g' %s", dir)
	stdout, err := RunHostCmd(namespace, podName, command)
	if err != nil {
		Logf("got err: %v, retry until timeout", err)
	}
	fsGroup := strings.TrimSpace(stdout)
	if fsGroup != "1000" {
		Failf("Not expected group owner, group owner is %v but should be 1000", fsGroup)
	}
	Logf("Expected group owner, group owner is %v ", fsGroup)
}

// CheckVolumeDirectoryOwnership creates a pod with a dynamically provisioned volume
func (j *PVCTestJig) CheckVolumeDirectoryOwnership(namespace string, pvcParam *v1.PersistentVolumeClaim) {
	pvc, err := j.KubeClient.CoreV1().PersistentVolumeClaims(pvcParam.Namespace).Get(context.Background(), pvcParam.Name, metav1.GetOptions{})

	if err != nil {
		Failf("Failed to get persistent volume %q: %v", pvc.Spec.VolumeName, err)
	}
	By("checking the created volume is writable and has the PV's mount options")
	command := "while true; do echo 'hello world' >> /usr/share/nginx/html/out.txt; sleep 5; done"

	podName := j.CreateAndAwaitNginxPodOrFail(pvc.Namespace, pvc, command)

	j.checkFileOwnership(namespace, podName, "/usr/share/nginx/html/out.txt")
}

//CheckExpandedVolumeReadWrite checks a pvc expanded pod with a dymincally provisioned volume
func (j *PVCTestJig) CheckExpandedVolumeReadWrite(namespace string, podName string) {
	pattern := "ReadWriteTest"
	text := fmt.Sprintf("hello expanded pvc pod %s", pattern)
	command := fmt.Sprintf("echo '%s' > /data/test1; grep '%s'  /data/test1 ", text, pattern)

	if pollErr := wait.PollImmediate(K8sResourcePoll, DefaultTimeout, func() (bool, error) {
		stdout, err := RunHostCmd(namespace, podName, command)
		if err != nil {
			Logf("got err: %v, retry until timeout", err)
			return false, nil
		}
		return strings.Contains(stdout, text), nil
	}); pollErr != nil {
		Failf("Write Test failed in pod '%v' after expanding pvc", podName)
	}

}

//CheckUsableVolumeSizeInsidePod checks a pvc expanded pod with a dymincally provisioned volume
func (j *PVCTestJig) CheckUsableVolumeSizeInsidePod(namespace string, podName string, capacity string) {

	command := fmt.Sprintf("df -BG | grep '/data'")

	if pollErr := wait.PollImmediate(K8sResourcePoll, DefaultTimeout, func() (bool, error) {
		stdout, err := RunHostCmd(namespace, podName, command)
		if err != nil {
			Logf("got err: %v, retry until timeout", err)
			return false, nil
		}
		if strings.Fields(strings.TrimSpace(stdout))[1] != capacity {
			return false, nil
		} else {
			return true, nil
		}
	}); pollErr != nil {
		Failf("Write Test failed in pod '%v' after expanding pvc", podName)
	}

}

//CheckFilesystemTypeOfVolumeInsidePod Checks the volume is provisioned with FsType as requested
func (j *PVCTestJig) CheckFilesystemTypeOfVolumeInsidePod(namespace string, podName string, expectedFsType string) {
	command := fmt.Sprintf("df -Th | grep '/data'")
	stdout, err := RunHostCmd(namespace, podName, command)
	if err != nil {
		Logf("got err: %v, retry until timeout", err)
	}
	actualFsType := strings.Fields(strings.TrimSpace(stdout))[1]
	if actualFsType != expectedFsType {
		Failf("Filesystem type: %s does not match expected: %s", actualFsType, expectedFsType)
	}
	Logf("Filesystem type: %s is as expected", actualFsType)
}

// CreateAndAwaitNginxPodOrFail returns a pod definition based on the namespace using nginx image
func (j *PVCTestJig) CreateAndAwaitNginxPodOrFail(ns string, pvc *v1.PersistentVolumeClaim, command string) string {
	By("Creating a pod with the dynamically provisioned volume")
	fsGroup := int64(1000)
	pod, err := j.KubeClient.CoreV1().Pods(ns).Create(context.Background(), &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "pvc-pod-tester-" + pvc.Labels["testID"],
			Namespace:    ns,
		},
		Spec: v1.PodSpec{
			SecurityContext: &v1.PodSecurityContext{
				FSGroup: &fsGroup,
			},
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
	}, metav1.CreateOptions{})
	if err != nil {
		Failf("Pod %q Create API error: %v", pod.Name, err)
	}

	// Waiting for pod to be running
	err = j.waitTimeoutForPodRunningInNamespace(pod.Name, ns, slowPodStartTimeout)
	if err != nil {
		Failf("Pod %q is not Running: %v", pod.Name, err)
	}
	return pod.Name
}

// WaitForPodNotFoundInNamespace waits default amount of time for the specified pod to be terminated.
// If the pod Get api returns IsNotFound then the wait stops and nil is returned. If the Get api returns
//an error other than "not found" then that error is returned and the wait stops.
func (j *PVCTestJig) waitTimeoutForPodNotFoundInNamespace(podName, namespace string, timeout time.Duration) error {
	return wait.PollImmediate(Poll, timeout, j.podNotFound(podName, namespace))
}

// WaitTimeoutForPodRunningInNamespace waits default amount of time (PodStartTimeout) for the specified pod to become running.
// Returns an error if timeout occurs first, or pod goes in to failed state.
func (j *PVCTestJig) waitTimeoutForPodRunningInNamespace(podName, namespace string, timeout time.Duration) error {
	return wait.PollImmediate(Poll, timeout, j.podRunning(podName, namespace))
}

func (j *PVCTestJig) podRunning(podName, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		pod, err := j.KubeClient.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
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

// WaitTimeoutForPodRunningInNamespace waits default amount of time (PodStartTimeout) for the specified pod to become running.
// Returns an error if timeout occurs first, or pod goes in to failed state.
func (j *PVCTestJig) waitTimeoutForPodCompletedSuccessfullyInNamespace(podName, namespace string, timeout time.Duration) error {
	return wait.PollImmediate(Poll, timeout, j.podCompleted(podName, namespace))
}

func (j *PVCTestJig) podCompleted(podName, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		pod, err := j.KubeClient.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		switch pod.Status.Phase {
		case v1.PodSucceeded:
			return true, nil
		case v1.PodFailed:
			return false, errors.Errorf("Pod exited: %s", pod.Status.Reason)
		}
		return false, nil
	}
}

func (j *PVCTestJig) podNotFound(podName, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		_, err := j.KubeClient.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return true, nil // done
		}
		if err != nil {
			return true, err // stop wait with error
		}
		return false, nil
	}
}
