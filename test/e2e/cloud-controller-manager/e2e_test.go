/*
Copyright 2015 The Kubernetes Authors.

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

package e2e

import (
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/cloud-controller-manager/framework"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/cloud-controller-manager/framework/ginkgowrapper"
	"k8s.io/apiserver/pkg/util/logs"
)

func TestE2E(t *testing.T) {
	logs.InitLogs()
	defer logs.FlushLogs()

	gomega.RegisterFailHandler(ginkgowrapper.Fail)
	ginkgo.RunSpecs(t, "CCM E2E suite")
}

var _ = ginkgo.SynchronizedAfterSuite(func() {
	framework.Logf("Running AfterSuite actions on all node")
	framework.RunCleanupActions()
}, func() {})
