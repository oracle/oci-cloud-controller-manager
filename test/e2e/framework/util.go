package framework

import (
	"fmt"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	testutil "k8s.io/kubernetes/test/utils"
)

type podCondition func(pod *v1.Pod) (bool, error)

// CheckPodsRunningReady returns whether all pods whose names are listed in
// podNames in namespace ns are running and ready, using c and waiting at most
// timeout.
func CheckPodsRunningReady(t *testing.T, c kubernetes.Interface, ns string, podNames []string, timeout time.Duration) bool {
	return CheckPodsCondition(t, c, ns, podNames, timeout, testutil.PodRunningReady, "running and ready")
}

// CheckPodsCondition returns whether all pods whose names are listed in podNames
// in namespace ns are in the condition, using c and waiting at most timeout.
func CheckPodsCondition(t *testing.T, c kubernetes.Interface, ns string, podNames []string, timeout time.Duration, condition podCondition, desc string) bool {
	np := len(podNames)
	t.Logf("Waiting up to %v for %d pods to be %s: %s", timeout, np, desc, podNames)
	type waitPodResult struct {
		success bool
		podName string
	}
	result := make(chan waitPodResult, len(podNames))
	for _, podName := range podNames {
		// Launch off pod readiness checkers.
		go func(name string) {
			err := WaitForPodCondition(t, c, ns, name, desc, timeout, condition)
			result <- waitPodResult{err == nil, name}
		}(podName)
	}
	// Wait for them all to finish.
	success := true
	for range podNames {
		res := <-result
		if !res.success {
			t.Logf("Pod %[1]s failed to be %[2]s.", res.podName, desc)
			success = false
		}
	}
	t.Logf("Wanted all %d pods to be %s. Result: %t. Pods: %v", np, desc, success, podNames)
	return success
}

// WaitForPodCondition waits for a Pod to satisfy a condition.
func WaitForPodCondition(t *testing.T, c kubernetes.Interface, ns, podName, desc string, timeout time.Duration, condition podCondition) error {
	t.Logf("Waiting up to %v for pod %q in namespace %q to be %q", timeout, podName, ns, desc)
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(Poll) {
		pod, err := c.CoreV1().Pods(ns).Get(podName, metav1.GetOptions{})
		if err != nil {
			if apierrs.IsNotFound(err) {
				t.Logf("Pod %q in namespace %q not found. Error: %v", podName, ns, err)
				return err
			}
			t.Logf("Get pod %q in namespace %q failed, ignoring for %v. Error: %v", podName, ns, Poll, err)
			continue
		}
		// log now so that current pod info is reported before calling `condition()`
		t.Logf("Pod %q: Phase=%q, Reason=%q, readiness=%t. Elapsed: %v",
			podName, pod.Status.Phase, pod.Status.Reason, podutil.IsPodReady(pod), time.Since(start))
		if done, err := condition(pod); done {
			if err == nil {
				t.Logf("Pod %q satisfied condition %q", podName, desc)
			}
			return err
		}
	}
	return fmt.Errorf("Gave up after waiting %v for pod %q to be %q", timeout, podName, desc)
}
