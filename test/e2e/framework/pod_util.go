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
	j.DeleteAndAwaitPodOrFail(namespace, podName)

	command = "while true; do sleep 5; done"
	By("Recreating a pod with the same dynamically provisioned volume and waiting for it to be running")
	podName = j.CreateAndAwaitNginxPodOrFail(pvc.Namespace, pvc, command)

	By("Checking if the file exists on the newly created pod")
	j.CheckFileExists(namespace, podName, "/usr/share/nginx/html", "out.txt")
}

// DeleteAndAwaitPodOrFail deletes the pod definition based on the namespace and waits for pod to disappear
func (j *PVCTestJig) DeleteAndAwaitPodOrFail(ns string, podName string) {
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

func (j *PVCTestJig) CheckDataInBlockDevice(namespace string, podName string, fileName string) {
	By("check if the block device has data")
	command := "dd if=/dev/xvda bs=2048 count=1 status=none"
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

func (j *PVCTestJig) CheckMountOptions(namespace string, podName string, expectedPath string, expectedOptions []string) {
	By("check if NFS mount options are applied")
	command := fmt.Sprintf("mount -t nfs")
	if pollErr := wait.PollImmediate(K8sResourcePoll, DefaultTimeout, func() (bool, error) {
		stdout, err := RunHostCmd(namespace, podName, command)
		if err != nil {
			Logf("got err: %v, retry until timeout", err)
			return false, nil
		}
		if stdout == "" || !strings.Contains(stdout, expectedPath) {
			return false, errors.Errorf("NFS Mount not found for path %s. Mounted as %s", expectedPath, stdout)
		}
		for _, option := range expectedOptions {
			if !strings.Contains(stdout, option) {
				return false, errors.Errorf("NFS Mount Options check failed. Mounted as %s", stdout)
			}
		}
		return true, nil
	}); pollErr != nil {
		Failf("NFS mount with Mount Options failed in pod '%v' with error '%v'", podName, pollErr.Error())
	}
}
func (j *PVCTestJig) CheckLustreMountOptions(namespace string, podName string, expectedPath string, expectedOptions []string) {
	command := fmt.Sprintf("mount -t lustre")
	if pollErr := wait.PollImmediate(K8sResourcePoll, DefaultTimeout, func() (bool, error) {
		stdout, err := RunHostCmd(namespace, podName, command)
		if err != nil {
			Logf("got err: %v, retry until timeout", err)
			return false, nil
		}
		if stdout == "" || !strings.Contains(stdout, expectedPath) {
			return false, errors.Errorf("Lustre Mount not found for path %s. Mounted as %s", expectedPath, stdout)
		}
		for _, option := range expectedOptions {
			if !strings.Contains(stdout, option) {
				return false, errors.Errorf("Lustre Mount Options check failed. Mounted as %s", stdout)
			}
		}
		return true, nil
	}); pollErr != nil {
		Failf("Lustre mount with Mount Options failed in pod '%v' with error '%v'", podName, pollErr.Error())
	}
}
func (j *PVCTestJig) ExtractDataFromBlockDevice(namespace string, podName string, devicePath string, outFile string) {
	By("extract data from block device")
	command := fmt.Sprintf("dd if=%s count=1 | tr -d '\\000' > %s", devicePath, outFile)
	if pollErr := wait.PollImmediate(K8sResourcePoll, DefaultTimeout, func() (bool, error) {
		_, err := RunHostCmd(namespace, podName, command)
		if err != nil {
			Logf("got err: %v, retry until timeout", err)
			return false, nil
		}
		return true, nil
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
	j.DeleteAndAwaitPodOrFail(pvc.Namespace, podName)

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

// CheckExpandedVolumeReadWrite checks a pvc expanded pod with a dymincally provisioned volume
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

// CheckExpandedRawBlockVolumeReadWrite checks a pvc expanded pod with a dymincally provisioned raw block volume
func (j *PVCTestJig) CheckExpandedRawBlockVolumeReadWrite(namespace string, podName string) {
	text := fmt.Sprintf("Hello New World")
	command := fmt.Sprintf("echo '%s' > /tmp/test.txt; dd if=/tmp/test.txt of=/dev/xvda count=1; dd if=/dev/xvda bs=512 count=1", text)

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

// CheckUsableVolumeSizeInsidePod checks a pvc expanded pod with a dymincally provisioned volume
func (j *PVCTestJig) CheckUsableVolumeSizeInsidePod(namespace string, podName string, capacity string) {

	command := fmt.Sprintf("df -BG | grep '/data'")

	if pollErr := wait.PollImmediate(K8sResourcePoll, DefaultTimeout, func() (bool, error) {
		stdout, err := RunHostCmd(namespace, podName, command)
		if err != nil {
			Logf("got err: %v, retry until timeout", err)
			return false, nil
		}
		if strings.Fields(strings.TrimSpace(stdout))[1] != capacity {
			Logf("Expected capacity: %v, got capacity: %v", capacity, strings.Fields(strings.TrimSpace(stdout))[1])
			return false, nil
		} else {
			return true, nil
		}
	}); pollErr != nil {
		Failf("Check Usable Volume Size Inside Pod Test failed in pod '%v' after expanding pvc", podName)
	}

}

// CheckUsableVolumeSizeInsidePodBlock checks a pvc expanded pod with a dymincally provisioned volume
/*
   Example output of `fdisk -l /dev/block`:

   Disk /dev/block: 55 GiB, 59055800320 bytes, 115343360 sectors
   Units: sectors of 1 * 512 = 512 bytes
   Sector size (logical/physical): 512 bytes / 4096 bytes
   I/O size (minimum/optimal): 4096 bytes / 1048576 bytes

   The expression strings.Fields(strings.TrimSpace(stdout))[2] returns "55".
*/
func (j *PVCTestJig) CheckUsableVolumeSizeInsidePodBlock(namespace string, podName string, capacity string) {
	command := fmt.Sprintf("fdisk -l /dev/xvda")
	if pollErr := wait.PollImmediate(K8sResourcePoll, DefaultTimeout, func() (bool, error) {
		stdout, err := RunHostCmd(namespace, podName, command)
		if err != nil {
			Logf("got err: %v, retry until timeout", err)
			return false, nil
		}
		if strings.Fields(strings.TrimSpace(stdout))[2] != capacity {
			Logf("Expected capacity: %v, got capacity: %v", capacity, strings.Fields(strings.TrimSpace(stdout))[1])
			return false, nil
		} else {
			return true, nil
		}
	}); pollErr != nil {
		Failf("Check Usable Volume Size Inside Pod Test failed in pod '%v' after expanding pvc", podName)
	}
}

// CheckFilesystemTypeOfVolumeInsidePod Checks the volume is provisioned with FsType as requested
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
	volumeMode := v1.PersistentVolumeFilesystem
	if *pvc.Spec.VolumeMode != "" {
		volumeMode = *pvc.Spec.VolumeMode
	}

	By("Creating a pod with the dynamically provisioned volume or raw block volume")
	fsGroup := int64(1000)

	// Define the container spec based on volume mode
	container := v1.Container{
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
	}

	// Modify container spec for block volume mode
	if volumeMode == v1.PersistentVolumeBlock {
		container.VolumeMounts = nil
		container.VolumeDevices = []v1.VolumeDevice{
			{
				Name:       "nginx-storage",
				DevicePath: "/dev/xvda",
			},
		}
	}

	// Create the pod
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
			Containers: []v1.Container{container},
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
	err = j.WaitTimeoutForPodRunningInNamespace(pod.Name, ns, slowPodStartTimeout)
	if err != nil {
		Logf("Pod failed to come up, logging debug info\n")
		j.LogPodDebugInfo(namespace, pod.Name)
		Failf("Pod %q is not Running: %v", pod.Name, err)
	}

	return pod.Name
}

// WaitForPodNotFoundInNamespace waits default amount of time for the specified pod to be terminated.
// If the pod Get api returns IsNotFound then the wait stops and nil is returned. If the Get api returns
// an error other than "not found" then that error is returned and the wait stops.
func (j *PVCTestJig) waitTimeoutForPodNotFoundInNamespace(podName, namespace string, timeout time.Duration) error {
	return wait.PollImmediate(Poll, timeout, j.podNotFound(podName, namespace))
}

// WaitTimeoutForPodRunningInNamespace waits default amount of time (PodStartTimeout) for the specified pod to become running.
// Returns an error if timeout occurs first, or pod goes in to failed state.
func (j *PVCTestJig) WaitTimeoutForPodRunningInNamespace(podName, namespace string, timeout time.Duration) error {
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

func (j *PVCTestJig) GetNodeHostnameFromPod(podName, namespace string) string {
	pod, err := j.KubeClient.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		Failf("Failed to get pod %q: %v", podName, err)
	}
	hostName := pod.Labels[NodeHostnameLabel]
	return hostName
}

func (j *PVCTestJig) CheckVolumeOwnership(namespace, podName, mountPath, expectedOwner string) {
	cmd := "ls -l " + mountPath + " | awk 'NR==2 { print $4 }'"
	cmdOutput, err := RunHostCmd(namespace, podName, cmd)
	if err != nil {
		Failf("Failed to check volume ownership in pod %q: %v", podName, err)
	}
	cmdOutput = strings.ReplaceAll(cmdOutput, "\n", "")
	if cmdOutput == expectedOwner {
		Logf("Verified volume group owner for PV in pod %q is %v", podName, cmdOutput)
	} else {
		Failf("Actual Volume group ownership: %v and expected ownership: %v is not matching", cmdOutput, expectedOwner)
	}
}

func (j *PVCTestJig) GetNodeNameFromPod(podName, namespace string) string {
	pod, err := j.KubeClient.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		Failf("Failed to get pod %q: %v", podName, err)
	}
	return pod.Spec.NodeName
}

func (j *PVCTestJig) GetCSIPodNameRunningOnNode(nodeName string) string {

	// List all pods in the kube-system namespace
	pods, err := j.KubeClient.CoreV1().Pods("kube-system").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		Failf("Failed to list pods in kube-system namespace: %v", err)
	}

	// Find the csi-oci-node pod running on the specific node
	for _, p := range pods.Items {
		if p.Spec.NodeName == nodeName && p.Labels["app"] == "csi-oci-node" {
			return p.Name
		}
	}

	Failf("Failed to find csi-oci-node pod on node %v in kube-system namespace: %v", nodeName, err)
	return ""
}

// describePod fetches details about the given pod
func (j *PVCTestJig) describePod(namespace, podName string) (*v1.Pod, error) {
	pod, err := j.KubeClient.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return pod, nil
}

func (j *PVCTestJig) describePVCs(namespace string, pod *v1.Pod) ([]*v1.PersistentVolumeClaim, error) {
	var pvcs []*v1.PersistentVolumeClaim
	for _, vol := range pod.Spec.Volumes {
		if vol.PersistentVolumeClaim != nil {
			pvcName := vol.PersistentVolumeClaim.ClaimName
			pvc, err := j.KubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), pvcName, metav1.GetOptions{})
			if err != nil {
				Logf("Failed to get PVC %s: %v", pvcName, err)
				continue
			}
			pvcs = append(pvcs, pvc)
		}
	}
	return pvcs, nil
}

func (j *PVCTestJig) getPodEvents(ns, podName string) ([]*v1.Event, error) {
	// Get events for the specified pod in the given namespace
	eventsList, err := j.KubeClient.CoreV1().Events(ns).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s", podName),
	})
	if err != nil {
		return nil, err
	}

	var podEvents []*v1.Event
	for i := range eventsList.Items {
		podEvents = append(podEvents, &eventsList.Items[i])
	}
	return podEvents, nil
}

func (j *PVCTestJig) logPodDetails(ns string, pod *v1.Pod) {
	fmt.Printf("Pod Description:\n")
	fmt.Printf("Name: %s\n", pod.Name)
	fmt.Printf("Namespace: %s\n", pod.Namespace)
	fmt.Printf("Node Name: %s\n", pod.Spec.NodeName)
	fmt.Printf("Status: %s\n", pod.Status.Phase)
	fmt.Printf("Conditions: %+v\n", pod.Status.Conditions)
	fmt.Printf("Labels: %+v\n", pod.Labels)
	fmt.Printf("Annotations: %+v\n", pod.Annotations)
	fmt.Printf("Containers:\n")
	for _, container := range pod.Spec.Containers {
		fmt.Printf("  Name: %s\n  Image: %s\n  Ready: %v\n", container.Name, container.Image, container.ReadinessProbe)
	}
	fmt.Println()

	// Fetch and print Pod events
	events, err := j.getPodEvents(ns, pod.Name)
	if err != nil {
		Logf("Error retrieving events for pod: %v", err)
	} else {
		fmt.Printf("Pod Events:\n")
		for _, event := range events {
			fmt.Printf("  Type: %s  Reason: %s  Message: %s  Time: %s\n", event.Type, event.Reason, event.Message, event.LastTimestamp)
		}
	}
	fmt.Println()
}

func logPVCDetails(pvc *v1.PersistentVolumeClaim) {
	fmt.Printf("PVC Description:\n")
	fmt.Printf("Name: %s\n", pvc.Name)
	fmt.Printf("Status: %s\n", pvc.Status.Phase)
	fmt.Printf("Access Modes: %v\n", pvc.Spec.AccessModes)
	fmt.Printf("Volume: %s\n", pvc.Spec.VolumeName)
	if pvc.Spec.StorageClassName != nil {
		fmt.Printf("Storage Class: %s\n", *pvc.Spec.StorageClassName) // Dereferencing the pointer
	} else {
		fmt.Println("Storage Class: Not set")
	}
	fmt.Printf("Resources Requests: %v\n", pvc.Spec.Resources.Requests)
	fmt.Printf("Labels: %v\n", pvc.Labels)
	fmt.Printf("Annotations: %v\n", pvc.Annotations)
	fmt.Println()
}

func (j *PVCTestJig) getPVCEvents(ns, pvcName string) ([]*v1.Event, error) {
	eventsList, err := j.KubeClient.CoreV1().Events(ns).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s", pvcName),
	})
	if err != nil {
		return nil, err
	}

	var pvcEvents []*v1.Event
	for i := range eventsList.Items {
		pvcEvents = append(pvcEvents, &eventsList.Items[i])
	}
	return pvcEvents, nil
}

func (j *PVCTestJig) logPVCEvents(ns, pvcName string) {
	pvcEvents, err := j.getPVCEvents(ns, pvcName)
	if err != nil {
		Logf("Error retrieving events for PVC %s: %v", pvcName, err)
	} else {
		fmt.Printf("PVC Events for %s:\n", pvcName)
		for _, event := range pvcEvents {
			fmt.Printf("  Type: %s  Reason: %s  Message: %s  Time: %s\n", event.Type, event.Reason, event.Message, event.LastTimestamp)
		}
	}
	fmt.Println()
}

func (j *PVCTestJig) logPVCs(ns string, pod *v1.Pod) {
	pvcs, err := j.describePVCs(ns, pod)
	if err != nil {
		Logf("Error describing PVCs: %v", err)
		return
	}
	for _, pvc := range pvcs {
		logPVCDetails(pvc)
		j.logPVCEvents(ns, pvc.Name)
	}
}

// getNodeDriverLogs fetches the last 15 lines of logs from the csi-oci-node pod running on the same node
func (j *PVCTestJig) getNodeDriverLogs(nodeName string) (string, error) {
	// Find the csi-oci-node-* pod running on the given node
	podList, err := j.KubeClient.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app=csi-oci-node",
	})
	if err != nil {
		return "", err
	}

	var nodeDriverPod *v1.Pod
	for _, pod := range podList.Items {
		if pod.Spec.NodeName == nodeName {
			nodeDriverPod = &pod
			break
		}
	}

	if nodeDriverPod == nil {
		return "", fmt.Errorf("no csi-oci-node pod found on node %s", nodeName)
	}

	// Fetch logs (last 15 lines)
	logTailLines := int64(15)
	req := j.KubeClient.CoreV1().Pods("kube-system").GetLogs(nodeDriverPod.Name, &v1.PodLogOptions{
		//This function does not exist for some reason
		TailLines: &logTailLines,
	})
	logs, err := req.DoRaw(context.TODO())
	if err != nil {
		return "", err
	}

	return string(logs), nil
}

func (j *PVCTestJig) logNodeDriverLogs(pod *v1.Pod) {
	if pod.Spec.NodeName == "" {
		Logf("Pod is not scheduled yet, skipping node driver logs.")
		return
	}
	nodeName := pod.Spec.NodeName
	logs, err := j.getNodeDriverLogs(nodeName)
	if err != nil {
		Logf("Error getting node driver logs: %v", err)
	} else {
		fmt.Printf("Node Driver Logs:\n%s\n", logs)
	}
}

func (j *PVCTestJig) LogPodDebugInfo(ns string, podName string) {
	pod, err := j.describePod(ns, podName)
	if err != nil {
		Logf("Error describing pod: %v", err)
		if apierrors.IsNotFound(err) {
			pods, listErr := j.KubeClient.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{})
			if listErr != nil {
				Logf("Error listing pods: %v", listErr)
			}
			Logf("Pods in namespace list: %v", pods.Items)
		}
		return
	}

	j.logPodDetails(ns, pod)
	j.logPVCs(ns, pod)
	j.logNodeDriverLogs(pod)
}
