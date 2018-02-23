/*
Copyright 2014 The Kubernetes Authors.

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
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/api/testapi"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	"k8s.io/kubernetes/plugin/pkg/scheduler/algorithm"
	"k8s.io/kubernetes/plugin/pkg/scheduler/algorithm/predicates"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"
	testutil "k8s.io/kubernetes/test/utils"

	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework/ginkgowrapper"
)

const (
	// How long to try single API calls (like 'get' or 'list'). Used to prevent
	// transient failures from failing tests.
	// TODO: client should not apply this timeout to Watch calls. Increased from 30s until that is fixed.
	SingleCallTimeout = 5 * time.Minute
)

func nowStamp() string {
	return time.Now().Format(time.StampMilli)
}

func log(level string, format string, args ...interface{}) {
	fmt.Fprintf(GinkgoWriter, nowStamp()+": "+level+": "+format+"\n", args...)
}

func Logf(format string, args ...interface{}) {
	log("INFO", format, args...)
}

func Failf(format string, args ...interface{}) {
	FailfWithOffset(1, format, args...)
}

func ExpectNoError(err error, explain ...interface{}) {
	ExpectNoErrorWithOffset(1, err, explain...)
}

// ExpectNoErrorWithOffset checks if "err" is set, and if so, fails assertion while logging the error at "offset" levels above its caller
// (for example, for call chain f -> g -> ExpectNoErrorWithOffset(1, ...) error would be logged for "f").
func ExpectNoErrorWithOffset(offset int, err error, explain ...interface{}) {
	if err != nil {
		Logf("Unexpected error occurred: %v", err)
	}
	ExpectWithOffset(1+offset, err).NotTo(HaveOccurred(), explain...)
}

// FailfWithOffset calls "Fail" and logs the error at "offset" levels above its caller
// (for example, for call chain f -> g -> FailfWithOffset(1, ...) error would be logged for "f").
func FailfWithOffset(offset int, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log("INFO", msg)
	ginkgowrapper.Fail(nowStamp()+": "+msg, 1+offset)
}

type podCondition func(pod *v1.Pod) (bool, error)

// CheckPodsRunningReady returns whether all pods whose names are listed in
// podNames in namespace ns are running and ready, using c and waiting at most
// timeout.
func CheckPodsRunningReady(c clientset.Interface, ns string, podNames []string, timeout time.Duration) bool {
	return CheckPodsCondition(c, ns, podNames, timeout, testutil.PodRunningReady, "running and ready")
}

// CheckPodsCondition returns whether all pods whose names are listed in podNames
// in namespace ns are in the condition, using c and waiting at most timeout.
func CheckPodsCondition(c clientset.Interface, ns string, podNames []string, timeout time.Duration, condition podCondition, desc string) bool {
	np := len(podNames)
	Logf("Waiting up to %v for %d pods to be %s: %s", timeout, np, desc, podNames)
	type waitPodResult struct {
		success bool
		podName string
	}
	result := make(chan waitPodResult, len(podNames))
	for _, podName := range podNames {
		// Launch off pod readiness checkers.
		go func(name string) {
			err := WaitForPodCondition(c, ns, name, desc, timeout, condition)
			result <- waitPodResult{err == nil, name}
		}(podName)
	}
	// Wait for them all to finish.
	success := true
	for range podNames {
		res := <-result
		if !res.success {
			Logf("Pod %[1]s failed to be %[2]s.", res.podName, desc)
			success = false
		}
	}
	Logf("Wanted all %d pods to be %s. Result: %t. Pods: %v", np, desc, success, podNames)
	return success
}

// WaitForPodCondition waits for a Pod to satisfy a condition.
func WaitForPodCondition(c clientset.Interface, ns, podName, desc string, timeout time.Duration, condition podCondition) error {
	Logf("Waiting up to %v for pod %q in namespace %q to be %q", timeout, podName, ns, desc)
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(Poll) {
		pod, err := c.CoreV1().Pods(ns).Get(podName, metav1.GetOptions{})
		if err != nil {
			if apierrs.IsNotFound(err) {
				Logf("Pod %q in namespace %q not found. Error: %v", podName, ns, err)
				return err
			}
			Logf("Get pod %q in namespace %q failed, ignoring for %v. Error: %v", podName, ns, Poll, err)
			continue
		}
		// log now so that current pod info is reported before calling `condition()`
		Logf("Pod %q: Phase=%q, Reason=%q, readiness=%t. Elapsed: %v",
			podName, pod.Status.Phase, pod.Status.Reason, podutil.IsPodReady(pod), time.Since(start))
		if done, err := condition(pod); done {
			if err == nil {
				Logf("Pod %q satisfied condition %q", podName, desc)
			}
			return err
		}
	}
	return fmt.Errorf("Gave up after waiting %v for pod %q to be %q", timeout, podName, desc)
}

// Filters nodes in NodeList in place, removing nodes that do not
// satisfy the given condition
// TODO: consider merging with pkg/client/cache.NodeLister
func FilterNodes(nodeList *v1.NodeList, fn func(node v1.Node) bool) {
	var l []v1.Node

	for _, node := range nodeList.Items {
		if fn(node) {
			l = append(l, node)
		}
	}
	nodeList.Items = l
}

// waitListSchedulableNodesOrDie is a wrapper around listing nodes supporting retries.
func waitListSchedulableNodesOrDie(c clientset.Interface) *v1.NodeList {
	var nodes *v1.NodeList
	var err error
	if wait.PollImmediate(Poll, SingleCallTimeout, func() (bool, error) {
		nodes, err = c.CoreV1().Nodes().List(metav1.ListOptions{FieldSelector: fields.Set{
			"spec.unschedulable": "false",
		}.AsSelector().String()})
		if err != nil {
			if IsRetryableAPIError(err) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	}) != nil {
		ExpectNoError(err, "Non-retryable failure or timed out while listing nodes for e2e cluster.")
	}
	return nodes
}

func isNodeConditionSetAsExpected(node *v1.Node, conditionType v1.NodeConditionType, wantTrue, silent bool) bool {
	// TODO(apryde): replace with vars from k8s.io/kubernetes/pkg/controller/nodelifecycle
	// once k8s dep bumped.
	unreachableTaintTemplate := &v1.Taint{
		Key:    algorithm.TaintNodeUnreachable,
		Effect: v1.TaintEffectNoExecute,
	}
	notReadyTaintTemplate := &v1.Taint{
		Key:    algorithm.TaintNodeNotReady,
		Effect: v1.TaintEffectNoExecute,
	}
	// Check the node readiness condition (logging all).
	for _, cond := range node.Status.Conditions {
		// Ensure that the condition type and the status matches as desired.
		if cond.Type == conditionType {
			// For NodeReady condition we need to check Taints as well
			if cond.Type == v1.NodeReady {
				hasNodeControllerTaints := false
				// For NodeReady we need to check if Taints are gone as well
				taints := node.Spec.Taints
				for _, taint := range taints {
					if taint.MatchTaint(unreachableTaintTemplate) || taint.MatchTaint(notReadyTaintTemplate) {
						hasNodeControllerTaints = true
						break
					}
				}
				if wantTrue {
					if (cond.Status == v1.ConditionTrue) && !hasNodeControllerTaints {
						return true
					} else {
						msg := ""
						if !hasNodeControllerTaints {
							msg = fmt.Sprintf("Condition %s of node %s is %v instead of %t. Reason: %v, message: %v",
								conditionType, node.Name, cond.Status == v1.ConditionTrue, wantTrue, cond.Reason, cond.Message)
						} else {
							msg = fmt.Sprintf("Condition %s of node %s is %v, but Node is tainted by NodeController with %v. Failure",
								conditionType, node.Name, cond.Status == v1.ConditionTrue, taints)
						}
						if !silent {
							Logf(msg)
						}
						return false
					}
				} else {
					// TODO: check if the Node is tainted once we enable NC notReady/unreachable taints by default
					if cond.Status != v1.ConditionTrue {
						return true
					}
					if !silent {
						Logf("Condition %s of node %s is %v instead of %t. Reason: %v, message: %v",
							conditionType, node.Name, cond.Status == v1.ConditionTrue, wantTrue, cond.Reason, cond.Message)
					}
					return false
				}
			}
			if (wantTrue && (cond.Status == v1.ConditionTrue)) || (!wantTrue && (cond.Status != v1.ConditionTrue)) {
				return true
			} else {
				if !silent {
					Logf("Condition %s of node %s is %v instead of %t. Reason: %v, message: %v",
						conditionType, node.Name, cond.Status == v1.ConditionTrue, wantTrue, cond.Reason, cond.Message)
				}
				return false
			}
		}

	}
	if !silent {
		Logf("Couldn't find condition %v on node %v", conditionType, node.Name)
	}
	return false
}

func IsNodeConditionSetAsExpected(node *v1.Node, conditionType v1.NodeConditionType, wantTrue bool) bool {
	return isNodeConditionSetAsExpected(node, conditionType, wantTrue, false)
}

func IsNodeConditionSetAsExpectedSilent(node *v1.Node, conditionType v1.NodeConditionType, wantTrue bool) bool {
	return isNodeConditionSetAsExpected(node, conditionType, wantTrue, true)
}

func IsNodeConditionUnset(node *v1.Node, conditionType v1.NodeConditionType) bool {
	for _, cond := range node.Status.Conditions {
		if cond.Type == conditionType {
			return false
		}
	}
	return true
}

// Node is schedulable if:
// 1) doesn't have "unschedulable" field set
// 2) it's Ready condition is set to true
// 3) doesn't have NetworkUnavailable condition set to true
func isNodeSchedulable(node *v1.Node) bool {
	nodeReady := IsNodeConditionSetAsExpected(node, v1.NodeReady, true)
	networkReady := IsNodeConditionUnset(node, v1.NodeNetworkUnavailable) ||
		IsNodeConditionSetAsExpectedSilent(node, v1.NodeNetworkUnavailable, false)
	return !node.Spec.Unschedulable && nodeReady && networkReady
}

// Test whether a fake pod can be scheduled on "node", given its current taints.
func isNodeUntainted(node *v1.Node) bool {
	fakePod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: testapi.Groups[v1.GroupName].GroupVersion().String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-not-scheduled",
			Namespace: "fake-not-scheduled",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "fake-not-scheduled",
					Image: "fake-not-scheduled",
				},
			},
		},
	}
	nodeInfo := schedulercache.NewNodeInfo()
	nodeInfo.SetNode(node)
	fit, _, err := predicates.PodToleratesNodeTaints(fakePod, nil, nodeInfo)
	if err != nil {
		Failf("Can't test predicates for node %s: %v", node.Name, err)
		return false
	}
	return fit
}

func IsRetryableAPIError(err error) bool {
	return apierrs.IsTimeout(err) || apierrs.IsServerTimeout(err) || apierrs.IsTooManyRequests(err)
}

// GetReadySchedulableNodesOrDie addresses the common use case of getting nodes you can do work on.
// 1) Needs to be schedulable.
// 2) Needs to be ready.
// If EITHER 1 or 2 is not true, most tests will want to ignore the node entirely.
func GetReadySchedulableNodesOrDie(c clientset.Interface) (nodes *v1.NodeList) {
	nodes = waitListSchedulableNodesOrDie(c)
	// previous tests may have cause failures of some nodes. Let's skip
	// 'Not Ready' nodes, just in case (there is no need to fail the test).
	FilterNodes(nodes, func(node v1.Node) bool {
		return isNodeSchedulable(&node) && isNodeUntainted(&node)
	})
	return nodes
}
