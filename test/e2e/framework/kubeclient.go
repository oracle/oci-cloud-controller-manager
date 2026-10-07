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
	"k8s.io/klog/v2"
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
MIIDGTCCAgGgAwIBAgIUZTsInKDv91f33o5hyBz76Wt7UgIwDQYJKoZIhvcNAQEL
BQAwFDESMBAGA1UEAwwJMTAuMC4xMC4yMB4XDTI2MDkwMzE2MDQyMFoXDTM2MDgz
MTE2MDQyMFowFDESMBAGA1UEAwwJMTAuMC4xMC4yMIIBIjANBgkqhkiG9w0BAQEF
AAOCAQ8AMIIBCgKCAQEA1xWwjJ/N6jCpI74JuxMVy5IGNDo2r1cxO7BujUbp0VHV
bU0sVWL5+imbkHIRXaCWUFN6TFu5RakNVWChRMRUB/eEs+bfOonK74/4PTCCNYvC
CNq3/201R68zM20zR5EfTL8EzNqGELvlGbhPN6URptXSWKaifPxmP/eBegFB18Dg
hZ/sK9CRXuh4hrRPCdKOBn4EtZy6tN57zY77cIMNMeBgUfcSpuWxjs0Z8vIvNtc4
tv9QPLaADlZ2r29c+Xq9/t1e9S4ibqJCakeCbw58Ncwc34+NqUYaqdPQP+3e6vBi
uQ81qwIr6T3tfyWD2WzRwAn6NTEu3sGXFEJ1nsQA0QIDAQABo2MwYTAdBgNVHQ4E
FgQUYYV9SrmVycRmM5TEa/C0CmTWRicwHwYDVR0jBBgwFoAUYYV9SrmVycRmM5TE
a/C0CmTWRicwDwYDVR0TAQH/BAUwAwEB/zAOBgNVHQ8BAf8EBAMCAQYwDQYJKoZI
hvcNAQELBQADggEBAJXJDvy/ZQ2AxauCK+ZZsqRTLQX27YVUIbALZ4s9IeUthMU9
oIh2rQ3nzjEdB6wYUamGAb4/ah8C/skoeaKJWZ2EILaidXqdsWgxVSXPF+NlrrF8
pX9My33l6Oz3nfJF0fPzqhz8h8KnynSagZw9Tugjrtu8QPbSwyJjjnDf7cZmut1x
y3EfxdeH2GRPzK6UklcAcLTCnE03McV4Scycay8GOXHQmsooGRcewIQ2wBh/M3aF
/viqaXzQrmkL3EemRjRFt+K4/aWrtLK9VCj/LavbUveFK7UJgEb5duJV8NGIVfbJ
qmQ1wQAL4/vFZRrIQ1jl7mrsRJX7rhStBhuKbfM=
-----END CERTIFICATE-----`
	SSLCertificateData = `-----BEGIN CERTIFICATE-----
MIIEHzCCAwegAwIBAgIUZqGh/1O6CpSV7Mut5MNZs8kwC7gwDQYJKoZIhvcNAQEL
BQAwFDESMBAGA1UEAwwJMTAuMC4xMC4yMB4XDTI2MDkwMzE2MDQyMFoXDTM2MDgz
MTE2MDQyMFowaDELMAkGA1UEBhMCSU4xEjAQBgNVBAgMCUtBUk5BVEFLQTESMBAG
A1UEBwwJQkVOR0FMVVJVMQ8wDQYDVQQKDAZPUkFDTEUxDDAKBgNVBAsMA09DSTES
MBAGA1UEAwwJMTAuMC4xMC4yMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKC
AQEAmjyiturUbclElVBUGdziK96UvBbYCpoy7O3CdQugpeNHvYgsx1Sh679fTZAP
UhzVN/FmV3Uru7Bal4Rgz1vcIbfqYg/6nIh7mF5Gqw1wr9utmOdqpi0Gx7n96/cj
MSyAUR3VUwXwhhNoSPGvYj2OhSb+LhMEVAYSUHSQRe6hyn460EfoyD+5lKHQpdhx
wMYS4+MiCk5ewmU9dVjwpVSxfKNMxBn/I7aiGwKBEQcC6V4YyVjcRnyjTMFFddPd
NfhZQs6ZPO1yPwmeesoHWx4+m84rPiA9RIQzu+davWPN8j4MDXSqRefZo7hZPwLy
jgSG0LhnJmr1trMDhcrrdCPDhwIDAQABo4IBEzCCAQ8wDAYDVR0TAQH/BAIwADAO
BgNVHQ8BAf8EBAMCBaAwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMIGP
BgNVHREEgYcwgYSCCmt1YmVybmV0ZXOCEmt1YmVybmV0ZXMuZGVmYXVsdIIWa3Vi
ZXJuZXRlcy5kZWZhdWx0LnN2Y4Iea3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVz
dGVygiRrdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWyHBAoACgIw
HQYDVR0OBBYEFH4JAXK5+xF6mQPwweeKRzWURNAPMB8GA1UdIwQYMBaAFGGFfUq5
lcnEZjOUxGvwtApk1kYnMA0GCSqGSIb3DQEBCwUAA4IBAQB2rzfBHBQvmBznh3p+
yhIz6FMAGCemNQ7fqDu9Grrodha3YUhaODKS0aGOWaztMHTN/QqrYsDYgJRdbRy9
9QEiLKf8Yjz5KxrkDTPUXiJ/IoNiTpB8E0+34ZlIZFR3KpbHwEMnt+h1T+uX/6gW
OqbNjuAH5tgH4OvJ7Y+r+kYYzP54e+X5/KrAI2LZbQm3NU+T8OKYxeMzuxF8g3ok
R+3CyPficWSnqgSc2hcz4VUPPT6Lnn5KSdtY7+VAUbSPumyLVQzX6qM/exfWYvtc
45JQ/jc9rs/73mIkXYBhfccxNA3qZyItrjAwesiFvSY0l3xHwpRUQl/iTT+SqFDm
ykZ3
-----END CERTIFICATE-----`
	SSLPrivateData = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAmjyiturUbclElVBUGdziK96UvBbYCpoy7O3CdQugpeNHvYgs
x1Sh679fTZAPUhzVN/FmV3Uru7Bal4Rgz1vcIbfqYg/6nIh7mF5Gqw1wr9utmOdq
pi0Gx7n96/cjMSyAUR3VUwXwhhNoSPGvYj2OhSb+LhMEVAYSUHSQRe6hyn460Efo
yD+5lKHQpdhxwMYS4+MiCk5ewmU9dVjwpVSxfKNMxBn/I7aiGwKBEQcC6V4YyVjc
RnyjTMFFddPdNfhZQs6ZPO1yPwmeesoHWx4+m84rPiA9RIQzu+davWPN8j4MDXSq
RefZo7hZPwLyjgSG0LhnJmr1trMDhcrrdCPDhwIDAQABAoIBACSQtX6qA3yXl6PS
bG3JOPFvjlFdFRDGZ8ZVw6EeBZLuZAah7wO+o7iRIRXxTkCIR2BA4aNgSuLvNzru
jkj6nSB9Spq+8QoFLU/9IcHRqOJ4MRqr2lPVHuNgy7sSVwyEYVNZwEYzhUcz+Kg6
a/rdXxlgGihwJ7mHyvW5/K4nmcG4dx68if+Ehgi9JdD0OmM4r52PgeFcNE5y6+Xi
cORmlEUDGe5mzkXRolJkpHz84CrFRc516/4tYctgk4lB339EI2o9vB1ZWx6IAaxv
cXpToBWjaHwtksU0Jc4XB0ms2Eaf+O4aYVnNo0BY0A8spZqzusSutTuuzlzYBY3B
TwKtKOECgYEAyCE9hn/2MZRA9GAGCgJ35K2pX8JpV/sfpLqndKrYpnaZryTJMWLF
wuCgwgITRm66qeM53BIUvlOnyiTKdJS0lENoE0aCgdU4OYODyyz7h9IdZNJZgccR
6rREVMi+TFvCZyZZR0PcCTHakthOJXqDTptRPPt0J9RYxwW9TKFBLlECgYEAxUuK
9OzIcWQXyw2FG2HrNQM+kBg+CiU4A5WhoyV7MRNDJ1O9q9nBqeRoDxNpS69Kf5gk
EXRL80MsUOyreoFF1Zmat0a3x3hqEKWY0Xa7DTy9aScVZLoVF6Ubud7lG4GeiGnD
S1gjOhPqIDuf0DN22RCWAngCSa4dXU8m/6ANJlcCgYBakQBgz6PASBElBhd1jCxp
plVR6o71q8VkLLv/RhmJK37dDc6mtMY+LJ1TbtD+PLnoi9XxS5VrlDwIdKHSJEGc
Hu0IXA5PZwhsrqGD3rVtf56hs7ehzU7EYhPSMo47zAKr32TjpUf8OT1q2sxylYC6
n/shl8G3DJeoaWaDOS5gIQKBgQCEnOFa/elBJmlDx+OnYyro6DReQJ06zoeXCTWr
Zp8mfm8N+SCtaWHeIzO6pm6JO9rUZtwfi08dxRH9lwcwAcKB74xqErOm9Q4+AED2
0lqqbCBYlLexi85vpUA8sFDJK1f3Ezf85dJP0GD3p3wlQuJoxtg98pJ/GfSM6o4p
FligYwKBgGRISue2MhcGj4IEu4ZwwiTE/W+FuXC5BEaJmXJSPT6HrnSB8dervy2O
ff75Ib3H291gxOIFzgtR3EvxomcjzGiNCA3wyFuzbN1QHYDR2YnsOKXjBSOvAjH4
FtNxcnWs7ordOwtVBQONx8hhKtbwg05UBFJZP1+VO+RQElmyhlV7
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

	_, matchingFlag := v1helper.FindMatchingUntoleratedTaint(klog.Background(), taints, pod.Spec.Tolerations, filter, false)

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

func RunHostCmdUsingChroot(ns, name, cmd string) (string, error) {
	return RunKubectl("exec", fmt.Sprintf("--namespace=%v", ns), name, "--", "chroot-bash", cmd)
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
	var gracePeriod int64 = 0
	deleteOption := &metav1.DeleteOptions{OrphanDependents: &falseVar, GracePeriodSeconds: &gracePeriod}
	startTime := time.Now()
	deleteFunc := func() (bool, error) {
		err := testutil.DeleteResource(c, kind, ns, name, *deleteOption)
		if err == nil || apierrs.IsNotFound(err) {
			return true, nil
		}
		return false, fmt.Errorf("failed to delete object with non-retriable error: %v", err)
	}
	if err := testutil.RetryWithExponentialBackOff(deleteFunc); err != nil {
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
