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
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pkg/errors"
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
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	v1helper "k8s.io/component-helpers/scheduling/corev1"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	appsinternal "k8s.io/kubernetes/pkg/apis/apps"
	batchinternal "k8s.io/kubernetes/pkg/apis/batch"
	api "k8s.io/kubernetes/pkg/apis/core"
	extensionsinternal "k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/controller/nodelifecycle"
	scheduler "k8s.io/kubernetes/pkg/scheduler/framework"
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
	// SSL CAData is a CA certificate not being used anywhere else; it only is utilised to check the load
	// balancer SSL connection during tests
	SSLCAData = `-----BEGIN CERTIFICATE-----
MIICpDCCAYwCCQDgCiDM+0GBsjANBgkqhkiG9w0BAQsFADAUMRIwEAYDVQQDDAkx
MC4wLjEwLjIwHhcNMjAwMzMxMTE1MjMwWhcNNDcwODE3MTE1MjMwWjAUMRIwEAYD
VQQDDAkxMC4wLjEwLjIwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDN
C9q4cAUAUuh2IXm78xCheKYASZ3Q1/eHLrwkNw2/mB6mjLefCFxxqsbCs4fWNile
iOmLUkZIyR17K82YNZIOlekUDzGHh3HFbUnC8uoJk4hr4sy0Qh5D7Ebb8x/s5JE7
Ghi7PVq2vKd/RCSJ6zs94McosBKJ7HxfHPWL3e59AHKroWXkT4JoikXcazE+vP37
3mjZyFotpDbDuzc2bVLOsCmgaNUJgn9wz5yESpiqY5TnEDYw4CbTW3luqnQSh+Ju
6CKYYsZPBzXoUByXAixSnFZUZu+US0xlJs80+J3Rxy1NN3qO5sNzdAulDwj/3wOO
j82Q+2Uoo9xz0pEiXoTrAgMBAAEwDQYJKoZIhvcNAQELBQADggEBAHxANdp0bAWG
st3UfWS6djkEIHdQgLADkwFAwMMRzJPJ1AV2t4h1SKT7c39hvfN7XuJcCaYYHUIQ
s4wmg+y02z7mx7OsKwvbbXKmBw80+BFRbTTK8yzuuCq1RZGWMm5cWSzDPuNu6Si7
ux7/+PSKo0zqcmFzq+AmuLo0g29cAATP3H5VL4W82eZTw/ABcDe4x3M9DDirxrfM
8s8UeORlXe7jvecjIXeYILhbdtp6htJ63P7MGXxQOyxzlhc/BKV+B2ydKS7GrCrM
N9L6fYa+pBVYivpEo0LN8AgfJjvpMEFWlYwxjXhGzMrvCijZo+OMAsyz2Wq58Bd8
PfQy9LDBAxc=
-----END CERTIFICATE-----`
	SSLCertificateData = `-----BEGIN CERTIFICATE-----
MIIEAzCCAuugAwIBAgIJALYtmrEW3ho9MA0GCSqGSIb3DQEBBQUAMBQxEjAQBgNV
BAMMCTEwLjAuMTAuMjAeFw0yMDAzMzExMTU2MThaFw00NzA4MTcxMTU2MThaMGgx
CzAJBgNVBAYTAklOMRIwEAYDVQQIDAlLQVJOQVRBS0ExEjAQBgNVBAcMCUJFTkdB
TFVSVTEPMA0GA1UECgwGT1JBQ0xFMQwwCgYDVQQLDANPQ0kxEjAQBgNVBAMMCTEw
LjAuMTAuMjCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMsTxDva6MGE
Ov7I6Q+FcZxnQoaJKuG/C8jnJyBUTLnooZfvEmqam3vegN3rUGjAQDcdhXLhbfRX
D0iWiVPd4vYZMPxT2uXdviq+8IBN6cDj1z6jiMbmnkH1CGeEHBjUrNbrbEm4/4t0
byQ4K1DiPkcqU+Hq6UF9nEGebaZWg3dksQWFlsmVJFSET1ysiZuGH7dQrCg7WxM8
yHYevQ5X8n1kQ2FsXw/Q/0X5Zg2n27p16NTgtlCdzB4vrQ9jcYAhVbzloohPgTA9
vHt5hlYOlr6BbXLrWoZ76xaHBw1p6DCtuvQGyK2dVrlCgcTpW1i8/V227MRWA7JE
ajZNthra0N0CAwEAAaOCAQIwgf8wLgYDVR0jBCcwJaEYpBYwFDESMBAGA1UEAwwJ
MTAuMC4xMC4yggkA4AogzPtBgbIwCQYDVR0TBAIwADALBgNVHQ8EBAMCBDAwHQYD
VR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMIGVBgNVHREEgY0wgYqCCmt1YmVy
bmV0ZXOCEmt1YmVybmV0ZXMuZGVmYXVsdIIWa3ViZXJuZXRlcy5kZWZhdWx0LnN2
Y4Iea3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVzdGVygiRrdWJlcm5ldGVzLmRl
ZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWyHBAoACgKHBAoACgEwDQYJKoZIhvcNAQEF
BQADggEBAF9qgvdINum9t8kLWmSKBTfuOGozWD7C4+gZgjNGyCmJ3kU/4iP3MWKu
QyDQujfuX5v+y2uRczJ3Vn65bTsc7UsuRU3X6m4TFnd/HXWL2PpX3nq8bDyQkNYr
4kAO4eDa8hf0HPR2Vnrmh6NvOEL0tXpDyc1mtzLg/SIPntu0dT8Up9uf4OUmYuBH
QVDHkZ8d6CzWXFng+NSXjR6nUAQUZFU4LHADQnKIcyDWaiKZFAEqbqQRSx9wxl1Q
ZtHtnRrs4vJdPDIRU57azHEyfuwDDGW8jJdc9hGaCaXAdfMmr7iALilLdMbupZQo
WKEIlNmKmQj8eVAvGrzD0tGVnRxP9zw=
-----END CERTIFICATE-----`
	SSLPrivateData = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAyxPEO9rowYQ6/sjpD4VxnGdChokq4b8LyOcnIFRMueihl+8S
apqbe96A3etQaMBANx2FcuFt9FcPSJaJU93i9hkw/FPa5d2+Kr7wgE3pwOPXPqOI
xuaeQfUIZ4QcGNSs1utsSbj/i3RvJDgrUOI+RypT4erpQX2cQZ5tplaDd2SxBYWW
yZUkVIRPXKyJm4Yft1CsKDtbEzzIdh69DlfyfWRDYWxfD9D/RflmDafbunXo1OC2
UJ3MHi+tD2NxgCFVvOWiiE+BMD28e3mGVg6WvoFtcutahnvrFocHDWnoMK269AbI
rZ1WuUKBxOlbWLz9XbbsxFYDskRqNk22GtrQ3QIDAQABAoIBACU1cfclnRAYElcs
qMdXRAHMSbws1daXEqm08M5To9tMbI9SFqXBvktr8WC4BPusfhebKSBrfaIPcZVz
P6ZGOZet9fPFyY3kmztp0Ncxb2sQVBf+Dsmi58xeATQ2WI+UKDcY27aGVwxOQS75
u7YOPir77nKugB6nzUGYra6Um3H8hYNWTgWyiATb8Y0V4njCf8pAepGOptClyI1I
i5fsEE6q52jbGeFRK2JTysG8ovABBdGYsS8XOUuZ+O/QktF/iFwFtMWdEur5tcOO
RoPSrc/4H8pNpL7IhF0Iy/hpNoNsin7Gj4UBNi6dhrtcGz3zCGSKtldsootgSC2C
KWd/rAECgYEA5sF6OZsLguVfCqmj3WiLM5I+YWC/HAmV9grb9puW35cQxfQegmdj
InWk+rcotuFTBcTKjXDKT4C8vCZid2p0WnSWqLPWhPYg0p2awobZgjRy0HzvUgGJ
/gWAEydzsUc8ojHrUBdJ2iyvjy+I8JWQcyQkBUGlPZj0IC5VUgODYD0CgYEA4Usg
UCJqo35pLq0TmPSfUuMPzTV3StIft+r7S3g4HWpvrBQNKf6p96/Fjt2WaPhvAABB
ww8Pg2B97iSqR6Rg4Ba4BQQEfHtWCHQ2NuNOoNkRLTJqOxREk7+741Qy9EwgeDJ6
rQqgrde1dLJPZDzQpbFoCLkIkQ6CL3jTkyDenSECgYEAmvZ1STgoy9eTMsrnY2mw
iYp9X9GjpYV+coOqYfrsn+yH9BfTYUli1qJgj4nuypmYsngMel2zTx6qIEQ6vez8
hD5lapeSySmssyPp6Ra7/OeR7xbndI/aBn/VGYfV9shbHKUfXGK3Us/Nef+3G7Gl
Ft2/XtRNzobn8rCK1Y/MaxUCgYB6RFpKAxOanS0aLsX2+bNJuX7G4KBYE8cw+i7d
G2Zg2HW4jr1CMDov+M2fpjRNzZ34AyutX4wMwZ42UuGytcv5cXr3BeIlaI4dUmxl
x2DRvFwtCjJK08oP4TtnuTdaC8KHWOXo6V6gWfPZXDfn73VQpwIN0dWLW7NdbhZs
v6bw4QKBgEXYPIf827EVz0XU+1wkjaLt+G40J9sAPk/a6qybF33BBbBhjDxMnest
ArGIjYo4IcYu5hzwnPy/B9WIFgz1iY31l01eP90zJ6q+xpCO5qSdSnjkfq1zrwzK
Bs7B72+hgS7VwRowRUbNanaZIZt0ZAiwQWN1+Dh7Bj+VbSxc/fna
-----END RSA PRIVATE KEY-----`
	SSLPassphrase = ""
)

type KubeClient struct {
	Client clientset.Interface
	config *restclient.Config
}

func NewKubeClient(kubeConfig string) *KubeClient {
	tmpfile, err := ioutil.TempFile("", "kubeconfig")
	Expect(err).NotTo(HaveOccurred())
	_, err = tmpfile.Write([]byte(kubeConfig))

	Expect(err).NotTo(HaveOccurred())
	err = tmpfile.Close()
	Expect(err).NotTo(HaveOccurred())
	defer os.Remove(tmpfile.Name())
	config, err := clientcmd.BuildConfigFromFlags("", tmpfile.Name())
	Expect(err).NotTo(HaveOccurred())

	Logf("kubeclient.NewKubeClient exec provider is %#v", config.ExecProvider)

	client, err := clientset.NewForConfig(config)
	Expect(err).NotTo(HaveOccurred())
	return &KubeClient{
		Client: client,
		config: config,
	}
}

func (kc *KubeClient) NamespaceExists(ns string) bool {
	namespaces, err := kc.Client.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())
	for _, namespace := range namespaces.Items {
		if namespace.Name == ns {
			return true
		}
	}
	return false
}

func (kc *KubeClient) DeletePod(namespace string, name string, timeout time.Duration) error {
	Logf("deleting pod %s/%s", namespace, name)
	err := kc.Client.CoreV1().Pods(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if apierrs.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	for start := time.Now(); time.Since(start) < timeout; time.Sleep(Poll) {
		_, err := kc.Client.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if apierrs.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}
	}

	return errors.New("unable to delete pod within timeout")
}

func (kc *KubeClient) GetPodIP(namespace string, name string) string {
	pod, err := kc.Client.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
	Expect(err).To(BeNil())
	return pod.Status.PodIP
}

func (kc *KubeClient) CreatePod(pod *v1.Pod) (*v1.Pod, error) {
	return kc.Client.CoreV1().Pods(pod.Namespace).Create(context.Background(), pod, metav1.CreateOptions{})
}

func (kc *KubeClient) WaitForPodRunning(namespace string, name string, timeout time.Duration) {
	Expect(WaitForPodCondition(kc.Client, namespace, name, "running or failure", timeout, func(pod *v1.Pod) (bool, error) {
		switch pod.Status.Phase {
		case v1.PodFailed:
			logs, err := kc.PodLogs(namespace, name)
			if err != nil {
				logs = fmt.Sprintf("unable to fetch pod logs: %v", err)
			}
			return true, fmt.Errorf("pod %q failed with reason: %q, message: %q, logs:\n%s", name, pod.Status.Reason, pod.Status.Message, logs)
		case v1.PodRunning:
			return true, nil
		default:
			return false, nil
		}
	},
	)).To(Succeed(), "wait for pod %s/%s to run", namespace, name)
}

func (kc *KubeClient) WaitForPodSuccess(namespace string, name string, timeout time.Duration) {
	Expect(WaitForPodCondition(kc.Client, namespace, name, "success", timeout, func(pod *v1.Pod) (bool, error) {
		switch pod.Status.Phase {
		case v1.PodFailed:
			logs, err := kc.PodLogs(namespace, name)
			if err != nil {
				logs = fmt.Sprintf("unable to fetch pod logs: %v", err)
			}
			return true, fmt.Errorf("pod %q failed with reason: %q, message: %q, logs:\n%s", name, pod.Status.Reason, pod.Status.Message, logs)
		case v1.PodSucceeded:
			return true, nil
		default:
			return false, nil
		}
	},
	)).To(Succeed(), "wait for pod %s/%s to success", namespace, name)
}

func (kc *KubeClient) WaitForPodFailure(namespace string, name string, errorMessage string, timeout time.Duration) {
	Expect(WaitForPodCondition(kc.Client, namespace, name, "failure", timeout, func(pod *v1.Pod) (bool, error) {
		switch pod.Status.Phase {
		case v1.PodFailed:
			logs, err := kc.PodLogs(namespace, name)
			if err != nil {
				logs = fmt.Sprintf("unable to fetch pod logs: %v", err)
			}
			if strings.Contains(logs, errorMessage) {
				return true, nil
			}
			return true, fmt.Errorf("pod %q failed with reason: %q, message: %q, logs:\n%s", name, pod.Status.Reason, pod.Status.Message, logs)
		case v1.PodSucceeded:
			return true, fmt.Errorf("pod %q expected to fail but succeeded", name)
		default:
			return false, nil
		}
	},
	)).To(Succeed(), "wait for pod %s/%s to fail with error %s", namespace, name, errorMessage)
}

func (kc *KubeClient) PodLogs(namespace string, name string) (string, error) {
	req := kc.Client.CoreV1().Pods(namespace).GetLogs(name, &v1.PodLogOptions{})
	reader, err := req.Stream(context.Background())
	if err != nil {
		return "", errors.Wrap(err, "unable to open log stream")
	}

	logs, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", errors.Wrap(err, "unable to read logs")
	}

	return string(logs), nil
}

func (kc *KubeClient) CheckNodes(expectedNumNodes int) {
	nodes, err := kc.Client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())
	Expect(len(nodes.Items)).To(Equal(expectedNumNodes))
	for _, node := range nodes.Items {
		Expect(isNodeSchedulable(&node)).To(BeTrue())
	}
}

// Check the number of pods reported by k8s in the cluster
func (kc *KubeClient) CheckPods(expectedNumPods int) {
	pods, err := kc.Client.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())
	//	Expect(len(pods.Items)).To(Equal(expectedNumPods))
	// TODO: after debug, make this a fatal error again
	if len(pods.Items) != expectedNumPods {
		Logf("Error: pods found == %d; expected %d", len(pods.Items), expectedNumPods)
		Logf("Continuing test...")
	}
	for _, pod := range pods.Items {
		Logf("found pod: %s", pod.Name)
	}
}

// Check that k8s reports the expected server version
func (kc *KubeClient) CheckVersion(serverVersion string) {
	versionInfo, err := kc.Client.Discovery().ServerVersion()
	Expect(err).NotTo(HaveOccurred())

	// k8s returns version in this format: v1.9.7-2+ff9181f92914d6
	Logf("k8s server version:%s", versionInfo.GitVersion)
	Logf("k8s expected server version:%s", serverVersion)
	Expect(strings.HasPrefix(versionInfo.GitVersion, serverVersion)).To(BeTrue())
}

// CheckVersionSucceeds checks that the client can successfully fetch the version
func (kc *KubeClient) CheckVersionSucceeds() (string, error) {
	versionInfo, err := kc.Client.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return versionInfo.GitVersion, nil
}

// Exec executes a command in the specified container,
// returning stdout, stderr and error. `options` allowed for
// additional parameters to be passed.
func (kc *KubeClient) Exec(namespace, podName, containerName string, command []string) (string, error) {
	// Prepare the API URL used to execute another process within the Pod.  In
	// this case, we'll run a remote shell.

	req := kc.Client.CoreV1().RESTClient().
		Post().
		Namespace(namespace).
		Resource("pods").
		Name(podName).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: containerName,
			Command:   command,

			Stdin:  true,
			Stdout: true,
			Stderr: true,
			TTY:    true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(kc.config, "POST", req.URL())
	if err != nil {
		return "", err
	}

	rw := &bytes.Buffer{}
	// Connect this process' std{in,out,err} to the remote shell process.
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: rw,
		Stderr: rw,
		Tty:    false,
	})

	return rw.String(), err
}

// Node is schedulable if:
// 1) doesn't have "unschedulable" field set
// 2) it's Ready condition is set to true
// 3) doesn't have NetworkUnavailable condition set to true
func isNodeSchedulable(node *v1.Node) bool {
	var status bool
	//retry 3 times before deciding node is not ready
	for index := 0; index < 3; index++ {
		nodeReady := isNodeConditionSetAsExpected(node, v1.NodeReady, true, false)
		networkReady := isNodeConditionUnset(node, v1.NodeNetworkUnavailable) ||
			isNodeConditionSetAsExpected(node, v1.NodeNetworkUnavailable, false, true)
		status = !node.Spec.Unschedulable && nodeReady && networkReady
		if status {
			return status
		}
		time.Sleep(5 * time.Second)
	}
	return status
}

func isNodeConditionUnset(node *v1.Node, conditionType v1.NodeConditionType) bool {
	for _, cond := range node.Status.Conditions {
		if cond.Type == conditionType {
			return false
		}
	}
	return true
}

type podCondition func(pod *v1.Pod) (bool, error)

// The function labels the node specified by the nodeName with labels specified by map labelValues.
func (kc *KubeClient) AddLabelsToNode(nodeName string, labelValues map[string]string) {
	Logf("Retrieving kubernetes node with hostname %s", nodeName)
	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"hostname": nodeName}}
	nodes, err := kc.Client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{LabelSelector: labels.Set(labelSelector.MatchLabels).String()})
	Expect(err).NotTo(HaveOccurred())
	Expect(len(nodes.Items)).Should(BeNumerically(">", 0))

	node := &nodes.Items[0]
	//Add all labels
	for k, v := range labelValues {
		node.Labels[k] = v
	}

	resultNode, err := kc.Client.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
	if err != nil {
		Logf("Could not label node %q with labels %+v. Error: %v", nodeName, labelValues, err)
	}
	Expect(err).NotTo(HaveOccurred())
	Logf("Labeled node %s Successfully. Hostname for the node is %s. "+
		"New labels after operation are %+v. ", node.Name, nodeName, resultNode.Labels)
}

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
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(K8sResourcePoll) {
		pod, err := c.CoreV1().Pods(ns).Get(context.Background(), podName, metav1.GetOptions{})
		if err != nil {
			if apierrs.IsNotFound(err) {
				Logf("Pod %q in namespace %q not found. Error: %v", podName, ns, err)
				return err
			}
			Logf("Get pod %q in namespace %q failed, ignoring for %v. Error: %v", podName, ns, K8sResourcePoll, err)
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
	if wait.PollImmediate(K8sResourcePoll, SingleCallTimeout, func() (bool, error) {
		nodes, err = c.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{FieldSelector: fields.Set{
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

// Test whether a fake pod can be scheduled on "node", given its current taints.
func isNodeUntainted(node *v1.Node) bool {
	fakePod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
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
	nodeInfo := scheduler.NewNodeInfo()
	nodeInfo.SetNode(node)
	fit, err := PodToleratesNodeTaints(fakePod, nodeInfo)
	if err != nil {
		Failf("Can't test predicates for node %s: %v", node.Name, err)
		return false
	}
	return fit
}

func PodToleratesNodeTaints(pod *v1.Pod, nodeInfo *scheduler.NodeInfo) (bool, error) {
	if nodeInfo == nil || nodeInfo.Node() == nil {
		return false, nil
	}

	return podToleratesNodeTaints(pod, nodeInfo, func(t *v1.Taint) bool {
		// PodToleratesNodeTaints is only interested in NoSchedule and NoExecute taints.
		return t.Effect == v1.TaintEffectNoSchedule || t.Effect == v1.TaintEffectNoExecute
	})
}

func podToleratesNodeTaints(pod *v1.Pod, nodeInfo *scheduler.NodeInfo, filter func(t *v1.Taint) bool) (bool, error) {
	taints := nodeInfo.Node().Spec.Taints
	if len(taints) == 0 {
		return true, nil
	}

	_, matchingFlag := v1helper.FindMatchingUntoleratedTaint(taints, pod.Spec.Tolerations, filter)

	if !matchingFlag {
		return true, nil
	}
	return false, nil
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
					Image:   busyBoxImage,
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
	created, err := client.CoreV1().Pods(ns).Create(context.Background(), execPod, metav1.CreateOptions{})
	Expect(err).NotTo(HaveOccurred())
	err = wait.PollImmediate(K8sResourcePoll, 5*time.Minute, func() (bool, error) {
		retrievedPod, err := client.CoreV1().Pods(execPod.Namespace).Get(context.Background(), created.Name, metav1.GetOptions{})
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

	defaultArgs = append(defaultArgs, "--"+clientcmd.RecommendedConfigPathFlag+"="+clusterkubeconfig)
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
		return c.CoreV1().ReplicationControllers(ns).Get(context.Background(), name, metav1.GetOptions{})
	case api.Kind("Secrets"):
		return c.CoreV1().Secrets(ns).Get(context.Background(), name, metav1.GetOptions{})
	case extensionsinternal.Kind("ReplicaSet"), appsinternal.Kind("ReplicaSet"):
		return c.ExtensionsV1beta1().ReplicaSets(ns).Get(context.Background(), name, metav1.GetOptions{})
	case extensionsinternal.Kind("Deployment"), appsinternal.Kind("Deployment"):
		return c.ExtensionsV1beta1().Deployments(ns).Get(context.Background(), name, metav1.GetOptions{})
	case extensionsinternal.Kind("DaemonSet"):
		return c.ExtensionsV1beta1().DaemonSets(ns).Get(context.Background(), name, metav1.GetOptions{})
	case batchinternal.Kind("Job"):
		return c.BatchV1().Jobs(ns).Get(context.Background(), name, metav1.GetOptions{})
	default:
		return nil, fmt.Errorf("Unsupported kind when getting runtime object: %v", kind)
	}
}

func deleteResource(c clientset.Interface, kind schema.GroupKind, ns, name string, deleteOption *metav1.DeleteOptions) error {
	switch kind {
	case api.Kind("ReplicationController"):
		return c.CoreV1().ReplicationControllers(ns).Delete(context.Background(), name, *deleteOption)
	case api.Kind("Secrets"):
		return c.CoreV1().Secrets(ns).Delete(context.Background(), name, *deleteOption)
	case extensionsinternal.Kind("ReplicaSet"), appsinternal.Kind("ReplicaSet"):
		return c.ExtensionsV1beta1().ReplicaSets(ns).Delete(context.Background(), name, *deleteOption)
	case extensionsinternal.Kind("Deployment"), appsinternal.Kind("Deployment"):
		return c.ExtensionsV1beta1().Deployments(ns).Delete(context.Background(), name, *deleteOption)
	case extensionsinternal.Kind("DaemonSet"):
		return c.ExtensionsV1beta1().DaemonSets(ns).Delete(context.Background(), name, *deleteOption)
	case batchinternal.Kind("Job"):
		return c.BatchV1().Jobs(ns).Delete(context.Background(), name, *deleteOption)
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
	if err := testutil.DeleteResourceWithRetries(c, kind, ns, name, *deleteOption); err != nil {
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
