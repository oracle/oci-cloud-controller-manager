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
	"bytes"
	"fmt"
	"io"
	"net"
	"net/url"
	"os/exec"
	"strings"
	"syscall"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework/ginkgowrapper"
	batch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/pkg/api/testapi"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	appsinternal "k8s.io/kubernetes/pkg/apis/apps"
	batchinternal "k8s.io/kubernetes/pkg/apis/batch"
	api "k8s.io/kubernetes/pkg/apis/core"
	extensionsinternal "k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/controller/nodelifecycle"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"
	schedulercache "k8s.io/kubernetes/pkg/scheduler/cache"
	testutil "k8s.io/kubernetes/test/utils"
	uexec "k8s.io/utils/exec"
)

const (
	// How long to try single API calls (like 'get' or 'list'). Used to prevent
	// transient failures from failing tests.
	// TODO: client should not apply this timeout to Watch calls. Increased from 30s until that is fixed.
	SingleCallTimeout = 5 * time.Minute

	// Number of objects that gc can delete in a second.
	// GC issues 2 requestes for single delete.
	gcThroughput = 10
)

var (
	BusyBoxImage = "busybox"
	// SSL CAData is a CA certificate not being used anywhere else; it only is utilised to check the load
	// balancer SSL connection during tests
	SSLCAData = `-----BEGIN CERTIFICATE-----
MIIE5jCCAs6gAwIBAgIBATANBgkqhkiG9w0BAQsFADATMREwDwYDVQQDEwhzc2wt
dGVzdDAeFw0xODA5MDYxNTQzNTlaFw0yMDAzMDYxNTQzNTlaMBMxETAPBgNVBAMT
CHNzbC10ZXN0MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAuZkDWwTh
2/Oq00ViJw8yXUqAppGgH81QaHbSkGp1W4TbpieFufheWYkxjBksqynph/RD1JX/
J+k7WLha6QcAEFsDB+b9b+suQS8MPoLxE7Rw95p6PW9TlXYOh7SQSjOOFSA/SDKh
ewt7PIaGtvn0Qmda8mOTC/Bi/XbnF+ZdM39/WsHYGdqHfEsT40L8Kz3BvQaaLL6Q
ocJEMKmBCVUN6bLhc8GhFhM1mOn9fvM5wKNzLV07VeEWogH3Zz48qiCK1gosE3c8
oHynHxbEtnsjawIeYsy01A/i434i1dksqA0bRaNJu5YCQAf5o9JhyPJb16neUag8
wEuczF7x5Q0V/G5oU7k4Hh1Zcl1kIG7hHvgcpV6uX6+i5H/0Pmxa+9TVP2x3axwH
0XDZyis8aH0aLLndtH/EqIUJfYq9Q3N871cBYueY9/rqL5sinXgrH0Ud5WKJeREL
ilFWwFthafj6DMe00S1lKCNIGtRLKt9B9GGhuRdRyOdw4ZWOrxYc66uoeyaQowvi
iNJnjZ0yQruk3uLkTOMrKo6xXkJKuHyoIWHL6uNFMcVBLBeA9ERdjV36wCJx+zh7
y2DGydQjSbcGKTYHQr1z21E1n+QRMEnqxQ5cW1kWxhYecfrOja4zh/iGrXOA9ffO
BEHn0Fb/fA/Pijr0t7mvH4VdaICx8G3X2TUCAwEAAaNFMEMwDgYDVR0PAQH/BAQD
AgEGMBIGA1UdEwEB/wQIMAYBAf8CAQAwHQYDVR0OBBYEFDEbBw9aIWhE9MHxc+Cm
4xHNnmu8MA0GCSqGSIb3DQEBCwUAA4ICAQCLtbLoPS/jLWA1zgbw2wQA2rcA7kpB
D3iKKzA+6FnJlWSbMi9u4Ui/e3Z98P0KLafHZ+E/0jzMM4gq/yEMbWTP4KzY5aCG
sRJMtawSf3hCxK/1uRmdFbEiygQh2BzjmDb2TYXOeLMMSif+5LzXHD9Vzq4Rnzbf
Qqf0TnKh7EHIjptkEWgsIX0afBPfXSs0EzVQ8dJNioy/NqkuUoStCdB2OaKcFt/9
YoCuaWf/4hGvOY59WpUI1i1MXuig2s3o1cpkwu4jB08la2kPFCinv4Cpsv4s5BvY
zsxPFeMFjjMNLaZnp/LruvbvMq08D58Qvqsn4vpxrYPiRPNvpO3Dh1OeA6KppTNV
15NtOG0H0K2Wc7iJCMV+vnGBIWUZ9lh57RcKb+pf9E5F5vWBO2MdKUugrUQ60NH9
cubjZ7moO9OgjNLqe24PdEtx3CK933bwvTn2fT5pR1yYVOGVI0vbmFDgwD9Tg6+X
Q3xRbI67X9qqGWj1Nxl+WAOCFdzvQWaI6+f4Byhf1xwQ07CFSykUA1JpurXfemkT
R1F56uxXc5pWQFYU38r8ROKfrzD857kIGoMQhHbf3KBTavR1fCQdun3991wfkKsd
7Quf316BO4qqn1G3e5vIaRLAqePgvSdm7CRzOtja1RXvqA/o52/9zVourhZinK0x
UV5xGY0MDAKR0g==
-----END CERTIFICATE-----`
	SSLCertificateData = `-----BEGIN CERTIFICATE-----
MIIEJzCCAg+gAwIBAgIRAPtkjNA2AzZwFxKV/en8uDIwDQYJKoZIhvcNAQELBQAw
EzERMA8GA1UEAxMIc3NsLXRlc3QwHhcNMTgwOTA2MTU0NTI1WhcNMjAwMzA2MTU0
MzU4WjAYMRYwFAYDVQQDEw1zc2wtdGVzdC1jZXJ0MIIBIjANBgkqhkiG9w0BAQEF
AAOCAQ8AMIIBCgKCAQEAlFq7tzGl8ljlzyrddPI5WWs6BfHTwi0WOkhWCOMT2jzo
qfjZJ4WXCKIn0V3fmsG9MkmDkCRmoBwrwY8axUqwKopF5MQp39rF1W2L7EYaMPiH
Ry71gPwbfyzWfRcOCG0rlNuyZOkWh9AYKJQAlDPDZrQ9X6fzj7/FhqJLPqHbSgBT
PCQ7GnPUITYEGcND7d0t4KcojwJHog+7WRuChwq28MVM5gOoONl2eI/C3q6CeXSK
dlXBjZbBj5pGV+7lqDoSvzSrJYhucruNKGAT3loMvJicRih5+kS1auOOeWLhqupM
UyLgDSHTLoWzS+gY29UiK3Au/qc72igby/xHaMsZ3QIDAQABo3EwbzAOBgNVHQ8B
Af8EBAMCA7gwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMB0GA1UdDgQW
BBSvDgYQAAyY+0OKaCspaI/157dYJjAfBgNVHSMEGDAWgBQxGwcPWiFoRPTB8XPg
puMRzZ5rvDANBgkqhkiG9w0BAQsFAAOCAgEAH7wRpPqp+M6UCeH9PYuKYml0eqFc
mhV67Qu/wDUYT8yqy3IoM5fKVyarxK2PbiHZxv+BMDnDMvIFi+8o2DxQC/2mvkgE
1tQsMIDPphFN/UIAEnBEdjFrnyQljsbjYpazcGr77oM0B5gjzPHc6pyCrgfjfN7g
Qb8qgTv33feLagGuZ8wV+LQcQxRL+hBqWnl6NivFeKwEHLG/pMKJmsxufTtVDVXt
GPJBoi6jnOKluIyU+XZAbRl+Iz5lUeEjr24D0+JJlOleqate4uIm8L11AEjQ5NcX
lgz1Fkr/V5mZIS/CnzxWNdJpCVV39VT5KLAYzXNsnE3kutwW0ERvANz2MqmOW3oW
hayT/qUlqvJQwghMWPRwAbU0l9TQvSzLWS9t/wLjb9G9xz4UxIrVqUHJ4BhAQsFF
pYQ6heCqW0mkmgN3Tz/3pgUk9DZvrDPVB6r431geeoFtRfeJMjGtbkjvNufo9Xp9
sQDNlhZ5X8ZFRKjU4VedYqYrcfya8sjAJnvdktzLozgU41Op0zkykNUogPFK1Wma
330oZwh9LsEJ3bQnNOQd5P5hERmP4CXDl43qdmnjInp9Fn1kzJFWKzho68tmbdo7
9rm7AXKM1hM5GcTxKcalQQa92UQCCPbbjQKNUj/dpGHHpNQtEYXlLcvq6OOBfegU
VcU/2vioT6/IBgM=
-----END CERTIFICATE-----`
	SSLPrivateData = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAlFq7tzGl8ljlzyrddPI5WWs6BfHTwi0WOkhWCOMT2jzoqfjZ
J4WXCKIn0V3fmsG9MkmDkCRmoBwrwY8axUqwKopF5MQp39rF1W2L7EYaMPiHRy71
gPwbfyzWfRcOCG0rlNuyZOkWh9AYKJQAlDPDZrQ9X6fzj7/FhqJLPqHbSgBTPCQ7
GnPUITYEGcND7d0t4KcojwJHog+7WRuChwq28MVM5gOoONl2eI/C3q6CeXSKdlXB
jZbBj5pGV+7lqDoSvzSrJYhucruNKGAT3loMvJicRih5+kS1auOOeWLhqupMUyLg
DSHTLoWzS+gY29UiK3Au/qc72igby/xHaMsZ3QIDAQABAoIBAQCIZTs83Jqzy+PN
qp1dDe6INbttB78nQYi0IRlopk30UhefXjlYysvnlh6BsYsJn6Iw/8HD3pO0FPJt
zQJR7a82DVbhUzEkrOFCrqr8mh8ucvD9E0GIS0GElVbGP1IZkTWGcM79VAfw6fOs
Kxa1Kzo1zPS4Y5Pcz/XtKH2BohIZX/jEBM2bLaX1l+/kwZbFwUjSByy6EPl3cDMG
6FeYMm7mtyR/nS7LvtyQf+d8BwEEoD23OX+TUNdf7Bc72C+5Xm8NNFaI6b/ocy1M
9h5Y2KGg8opYbonY8EM6cqeUBYoxgd4azOxKehyIs6EqKfi3EkGHMWr9BAapqWKL
KaVytYiJAoGBAMWyb1ZA4MGJCtKt42XHs7QlYwpzfmY4JSTYrF7C6kk2+55HR6Wf
GR6u788Rs5t7WP1dss2LUMF8ZIYf8O2Xx5cJ75UrT6Yfj2A2mAjOPNEGqZyrJ3M8
q4bqEqwyQizsxwNuWTfE8uT5Sv3EU1THXM0nd1xc5GgfrRzxpylFYglzAoGBAMAb
EqNv8mzZHU9/xN17BmED/rmWQJeHBXAFXG0n0U4VzjkZLJthE00ko1fFvyt/+dc2
ZctUHBjMktwqtGRmJHVr7FlZAjesaEDCfNGq7GfHn9cNTFqwMpNHZCsl+YydHehV
2WglfKSgNd29cOluD+WgPi+bsIrHqN801B3flrtvAoGAHxkLZHD+KdNpzWwm+gqo
3OyIoMs8Fc49IYenzZwxiGTKvcIOpiHZrLbt6A+rxghoHirQBn8kq9rqYSLDHkyi
y8J3Wmqes84BGqxby/7NGEBJC+jsYrcncCh/2XBqcnCoeYSxKhGj95qzTinExfA6
S9cqcm/2Sd65t3TXy6krW7sCgYAUoVOpFZaVNeO969y6ZEHVHVa3m3koTKm60/iP
CF2j4xeYbimqLgyiljKsdGIJS98Ky5627TtvNlj3J+bjeUylB4gEOFjSncM8YSaC
ZbaplniF4bm3a4Ci7GHHeHaKMT1K/B0y4AO4sjPWskdz6gvu/vxupGubG8H45nV5
F8/aJQKBgHwkkWYhysyy1F+SyWOkcpFa5ORYapYvVQhnVjgMB8SSs8KZquMV5ixX
imlfs6WRsA4mr20dMC9a4Lb51lET2XPvc5Tef/kESpGfc7Zbgx5/Sm5Z2IpNlhgp
rTMMobnwm2ixgvgHMC4Uq9U0HVM9hyQ7l88m5QD+02Qr0vV/ZCMi
-----END RSA PRIVATE KEY-----`
	SSLPassphrase = ""
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
					if taint.MatchTaint(nodelifecycle.UnreachableTaintTemplate) || taint.MatchTaint(nodelifecycle.NotReadyTaintTemplate) {
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

func newExecPodSpec(ns, generateName string) *v1.Pod {
	immediate := int64(0)
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: generateName,
			Namespace:    ns,
		},
		Spec: v1.PodSpec{
			TerminationGracePeriodSeconds: &immediate,
			Containers: []v1.Container{
				{
					Name:    "exec",
					Image:   BusyBoxImage,
					Command: []string{"sh", "-c", "while true; do sleep 5; done"},
				},
			},
		},
	}
	return pod
}

// CreateExecPodOrFail creates a simple busybox pod in a sleep loop used as a
// vessel for kubectl exec commands.
// Returns the name of the created pod.
func CreateExecPodOrFail(client clientset.Interface, ns, generateName string, tweak func(*v1.Pod)) string {
	Logf("Creating new exec pod")
	execPod := newExecPodSpec(ns, generateName)
	if tweak != nil {
		tweak(execPod)
	}
	created, err := client.CoreV1().Pods(ns).Create(execPod)
	Expect(err).NotTo(HaveOccurred())
	err = wait.PollImmediate(Poll, 5*time.Minute, func() (bool, error) {
		retrievedPod, err := client.CoreV1().Pods(execPod.Namespace).Get(created.Name, metav1.GetOptions{})
		if err != nil {
			if IsRetryableAPIError(err) {
				return false, nil
			}
			return false, err
		}
		return retrievedPod.Status.Phase == v1.PodRunning, nil
	})
	Expect(err).NotTo(HaveOccurred())
	return created.Name
}

// KubectlCmd runs the kubectl executable through the wrapper script.
func KubectlCmd(args ...string) *exec.Cmd {
	defaultArgs := []string{}

	if kubeconfig != "" {
		defaultArgs = append(defaultArgs, "--"+clientcmd.RecommendedConfigPathFlag+"="+kubeconfig)
	} else {
		Fail("--kubeconfig not provided")
	}
	kubectlArgs := append(defaultArgs, args...)

	//We allow users to specify path to kubectl, so you can test either "kubectl" or "cluster/kubectl.sh"
	//and so on.
	cmd := exec.Command("kubectl", kubectlArgs...)

	//caller will invoke this and wait on it.
	return cmd
}

// kubectlBuilder is used to build, customize and execute a kubectl Command.
// Add more functions to customize the builder as needed.
type kubectlBuilder struct {
	cmd     *exec.Cmd
	timeout <-chan time.Time
}

func NewKubectlCommand(args ...string) *kubectlBuilder {
	b := new(kubectlBuilder)
	b.cmd = KubectlCmd(args...)
	return b
}

func (b *kubectlBuilder) WithEnv(env []string) *kubectlBuilder {
	b.cmd.Env = env
	return b
}

func (b *kubectlBuilder) WithTimeout(t <-chan time.Time) *kubectlBuilder {
	b.timeout = t
	return b
}

func (b kubectlBuilder) WithStdinData(data string) *kubectlBuilder {
	b.cmd.Stdin = strings.NewReader(data)
	return &b
}

func (b kubectlBuilder) WithStdinReader(reader io.Reader) *kubectlBuilder {
	b.cmd.Stdin = reader
	return &b
}

func (b kubectlBuilder) ExecOrDie() string {
	str, err := b.Exec()
	Logf("stdout: %q", str)
	// In case of i/o timeout error, try talking to the apiserver again after 2s before dying.
	// Note that we're still dying after retrying so that we can get visibility to triage it further.
	if isTimeout(err) {
		Logf("Hit i/o timeout error, talking to the server 2s later to see if it's temporary.")
		time.Sleep(2 * time.Second)
		retryStr, retryErr := RunKubectl("version")
		Logf("stdout: %q", retryStr)
		Logf("err: %v", retryErr)
	}
	Expect(err).NotTo(HaveOccurred())
	return str
}

func isTimeout(err error) bool {
	switch err := err.(type) {
	case net.Error:
		if err.Timeout() {
			return true
		}
	case *url.Error:
		if err, ok := err.Err.(net.Error); ok && err.Timeout() {
			return true
		}
	}
	return false
}

func (b kubectlBuilder) Exec() (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := b.cmd
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	Logf("Running '%s %s'", cmd.Path, strings.Join(cmd.Args[1:], " ")) // skip arg[0] as it is printed separately
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("error starting %v:\nCommand stdout:\n%v\nstderr:\n%v\nerror:\n%v\n", cmd, cmd.Stdout, cmd.Stderr, err)
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- cmd.Wait()
	}()
	select {
	case err := <-errCh:
		if err != nil {
			var rc int = 127
			if ee, ok := err.(*exec.ExitError); ok {
				rc = int(ee.Sys().(syscall.WaitStatus).ExitStatus())
				Logf("rc: %d", rc)
			}
			return "", uexec.CodeExitError{
				Err:  fmt.Errorf("error running %v:\nCommand stdout:\n%v\nstderr:\n%v\nerror:\n%v\n", cmd, cmd.Stdout, cmd.Stderr, err),
				Code: rc,
			}
		}
	case <-b.timeout:
		b.cmd.Process.Kill()
		return "", fmt.Errorf("timed out waiting for command %v:\nCommand stdout:\n%v\nstderr:\n%v\n", cmd, cmd.Stdout, cmd.Stderr)
	}
	Logf("stderr: %q", stderr.String())
	return stdout.String(), nil
}

// RunKubectl is a convenience wrapper over kubectlBuilder
func RunKubectl(args ...string) (string, error) {
	return NewKubectlCommand(args...).Exec()
}

// RunHostCmd runs the given cmd in the context of the given pod using `kubectl exec`
// inside of a shell.
func RunHostCmd(ns, name, cmd string) (string, error) {
	return RunKubectl("exec", fmt.Sprintf("--namespace=%v", ns), name, "--", "/bin/sh", "-c", cmd)
}

func getRuntimeObjectForKind(c clientset.Interface, kind schema.GroupKind, ns, name string) (runtime.Object, error) {
	switch kind {
	case api.Kind("ReplicationController"):
		return c.CoreV1().ReplicationControllers(ns).Get(name, metav1.GetOptions{})
	case api.Kind("Secrets"):
		return c.CoreV1().Secrets(ns).Get(name, metav1.GetOptions{})
	case extensionsinternal.Kind("ReplicaSet"), appsinternal.Kind("ReplicaSet"):
		return c.ExtensionsV1beta1().ReplicaSets(ns).Get(name, metav1.GetOptions{})
	case extensionsinternal.Kind("Deployment"), appsinternal.Kind("Deployment"):
		return c.ExtensionsV1beta1().Deployments(ns).Get(name, metav1.GetOptions{})
	case extensionsinternal.Kind("DaemonSet"):
		return c.ExtensionsV1beta1().DaemonSets(ns).Get(name, metav1.GetOptions{})
	case batchinternal.Kind("Job"):
		return c.BatchV1().Jobs(ns).Get(name, metav1.GetOptions{})
	default:
		return nil, fmt.Errorf("Unsupported kind when getting runtime object: %v", kind)
	}
}

func deleteResource(c clientset.Interface, kind schema.GroupKind, ns, name string, deleteOption *metav1.DeleteOptions) error {
	switch kind {
	case api.Kind("ReplicationController"):
		return c.CoreV1().ReplicationControllers(ns).Delete(name, deleteOption)
	case api.Kind("Secrets"):
		return c.CoreV1().Secrets(ns).Delete(name, deleteOption)
	case extensionsinternal.Kind("ReplicaSet"), appsinternal.Kind("ReplicaSet"):
		return c.ExtensionsV1beta1().ReplicaSets(ns).Delete(name, deleteOption)
	case extensionsinternal.Kind("Deployment"), appsinternal.Kind("Deployment"):
		return c.ExtensionsV1beta1().Deployments(ns).Delete(name, deleteOption)
	case extensionsinternal.Kind("DaemonSet"):
		return c.ExtensionsV1beta1().DaemonSets(ns).Delete(name, deleteOption)
	case batchinternal.Kind("Job"):
		return c.BatchV1().Jobs(ns).Delete(name, deleteOption)
	default:
		return fmt.Errorf("Unsupported kind when deleting: %v", kind)
	}
}

func getSelectorFromRuntimeObject(obj runtime.Object) (labels.Selector, error) {
	switch typed := obj.(type) {
	case *v1.ReplicationController:
		return labels.SelectorFromSet(typed.Spec.Selector), nil
	case *extensions.ReplicaSet:
		return metav1.LabelSelectorAsSelector(typed.Spec.Selector)
	case *extensions.Deployment:
		return metav1.LabelSelectorAsSelector(typed.Spec.Selector)
	case *extensions.DaemonSet:
		return metav1.LabelSelectorAsSelector(typed.Spec.Selector)
	case *batch.Job:
		return metav1.LabelSelectorAsSelector(typed.Spec.Selector)
	default:
		return nil, fmt.Errorf("Unsupported kind when getting selector: %v", obj)
	}
}

func getReplicasFromRuntimeObject(obj runtime.Object) (int32, error) {
	switch typed := obj.(type) {
	case *v1.ReplicationController:
		if typed.Spec.Replicas != nil {
			return *typed.Spec.Replicas, nil
		}
		return 0, nil
	case *extensions.ReplicaSet:
		if typed.Spec.Replicas != nil {
			return *typed.Spec.Replicas, nil
		}
		return 0, nil
	case *extensions.Deployment:
		if typed.Spec.Replicas != nil {
			return *typed.Spec.Replicas, nil
		}
		return 0, nil
	case *batch.Job:
		// TODO: currently we use pause pods so that's OK. When we'll want to switch to Pods
		// that actually finish we need a better way to do this.
		if typed.Spec.Parallelism != nil {
			return *typed.Spec.Parallelism, nil
		}
		return 0, nil
	default:
		return -1, fmt.Errorf("Unsupported kind when getting number of replicas: %v", obj)
	}
}

// DeleteRCAndWaitForGC deletes only the Replication Controller and waits for GC to delete the pods.
func DeleteRCAndWaitForGC(c clientset.Interface, ns, name string) error {
	return DeleteResourceAndWaitForGC(c, api.Kind("ReplicationController"), ns, name)
}

// podStoreForSelector creates a PodStore that monitors pods from given namespace matching given selector.
// It waits until the reflector does a List() before returning.
func podStoreForSelector(c clientset.Interface, ns string, selector labels.Selector) (*testutil.PodStore, error) {
	ps, err := testutil.NewPodStore(c, ns, selector, fields.Everything())
	if err != nil {
		return nil, err
	}
	err = wait.Poll(100*time.Millisecond, 2*time.Minute, func() (bool, error) {
		if len(ps.Reflector.LastSyncResourceVersion()) != 0 {
			return true, nil
		}
		return false, nil
	})
	return ps, err
}

// waitForPodsInactive waits until there are no active pods left in the PodStore.
// This is to make a fair comparison of deletion time between DeleteRCAndPods
// and DeleteRCAndWaitForGC, because the RC controller decreases status.replicas
// when the pod is inactvie.
func waitForPodsInactive(ps *testutil.PodStore, interval, timeout time.Duration) error {
	return wait.PollImmediate(interval, timeout, func() (bool, error) {
		pods := ps.List()
		for _, pod := range pods {
			if controller.IsPodActive(pod) {
				return false, nil
			}
		}
		return true, nil
	})
}

// waitForPodsGone waits until there are no pods left in the PodStore.
func waitForPodsGone(ps *testutil.PodStore, interval, timeout time.Duration) error {
	return wait.PollImmediate(interval, timeout, func() (bool, error) {
		if pods := ps.List(); len(pods) == 0 {
			return true, nil
		}
		return false, nil
	})
}

// DeleteResourceAndWaitForGC deletes only given resource and waits for GC to delete the pods.
func DeleteResourceAndWaitForGC(c clientset.Interface, kind schema.GroupKind, ns, name string) error {
	By(fmt.Sprintf("deleting %v %s in namespace %s, will wait for the garbage collector to delete the pods", kind, name, ns))

	rtObject, err := getRuntimeObjectForKind(c, kind, ns, name)
	if err != nil {
		if apierrs.IsNotFound(err) {
			Logf("%v %s not found: %v", kind, name, err)
			return nil
		}
		return err
	}
	selector, err := getSelectorFromRuntimeObject(rtObject)
	if err != nil {
		return err
	}
	replicas, err := getReplicasFromRuntimeObject(rtObject)
	if err != nil {
		return err
	}

	ps, err := testutil.NewPodStore(c, ns, selector, fields.Everything())
	if err != nil {
		return err
	}

	defer ps.Stop()
	falseVar := false
	deleteOption := &metav1.DeleteOptions{OrphanDependents: &falseVar}
	startTime := time.Now()
	if err := testutil.DeleteResourceWithRetries(c, kind, ns, name, deleteOption); err != nil {
		return err
	}
	deleteTime := time.Since(startTime)
	Logf("Deleting %v %s took: %v", kind, name, deleteTime)

	var interval, timeout time.Duration
	switch {
	case replicas < 100:
		interval = 100 * time.Millisecond
	case replicas < 1000:
		interval = 1 * time.Second
	default:
		interval = 10 * time.Second
	}
	if replicas < 5000 {
		timeout = 10 * time.Minute
	} else {
		timeout = time.Duration(replicas/gcThroughput) * time.Second
		// gcThroughput is pretty strict now, add a bit more to it
		timeout = timeout + 3*time.Minute
	}

	err = waitForPodsInactive(ps, interval, timeout)
	if err != nil {
		return fmt.Errorf("error while waiting for pods to become inactive %s: %v", name, err)
	}
	terminatePodTime := time.Since(startTime) - deleteTime
	Logf("Terminating %v %s pods took: %v", kind, name, terminatePodTime)

	err = waitForPodsGone(ps, interval, 10*time.Minute)
	if err != nil {
		return fmt.Errorf("error while waiting for pods gone %s: %v", name, err)
	}
	return nil
}
