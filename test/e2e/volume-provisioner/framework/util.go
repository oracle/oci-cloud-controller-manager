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
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework/ginkgowrapper"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// Poll defines how regularly to poll kubernetes resources.
	Poll = 2 * time.Second
	// DefaultTimeout is how long we wait for long-running operations in the
	// test suite before giving up.
	DefaultTimeout = 10 * time.Minute
	// Some pods can take much longer to get ready due to volume attach/detach latency.
	slowPodStartTimeout = 15 * time.Minute
)

type podCondition func(pod *v1.Pod) (bool, error)

func nowStamp() string {
	return time.Now().Format(time.StampMilli)
}

func log(level string, format string, args ...interface{}) {
	fmt.Fprintf(GinkgoWriter, nowStamp()+": "+level+": "+format+"\n", args...)
}

// Logf used in framework
func Logf(format string, args ...interface{}) {
	log("INFO", format, args...)
}

// Failf used in framework
func Failf(format string, args ...interface{}) {
	FailfWithOffset(1, format, args...)
}

// FailfWithOffset calls "Fail" and logs the error at "offset" levels above its caller
// (for example, for call chain f -> g -> FailfWithOffset(1, ...) error would be logged for "f").
func FailfWithOffset(offset int, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log("INFO", msg)
	ginkgowrapper.Fail(nowStamp()+": "+msg, 1+offset)
}

// KubectlCmd runs the kubectl executable through the wrapper script.
func KubectlCmd(args ...string) *exec.Cmd {
	defaultArgs := []string{}

	if TestContext.KubeConfig != "" {
		defaultArgs = append(defaultArgs, "--"+clientcmd.RecommendedConfigPathFlag+"="+TestContext.KubeConfig)
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
