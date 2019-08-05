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

package e2e

import (
	"os"
	"testing"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	sharedfw "github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework/ginkgowrapper"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/volume-provisioner/framework"
	"k8s.io/component-base/logs"
)

var lockAquired bool
var installDisabled bool

var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {
	cs, err := framework.NewClientSetFromFlags()
	Ω(err).ShouldNot(HaveOccurred())

	err = sharedfw.AquireRunLock(cs, "oci-volume-provisioner-e2e-tests")
	Ω(err).ShouldNot(HaveOccurred())

	lockAquired = true

	_, installDisabled = os.LookupEnv("INSTALL_DISABLED")
	if !installDisabled {
		err = framework.InstallFlexvolumeDriver(cs)
		Ω(err).ShouldNot(HaveOccurred())

		err = framework.InstallVolumeProvisioner(cs)
		Ω(err).ShouldNot(HaveOccurred())
	}

	return nil
}, func(data []byte) {})

func TestE2E(t *testing.T) {
	logs.InitLogs()
	defer logs.FlushLogs()

	RegisterFailHandler(ginkgowrapper.Fail)
	ginkgo.RunSpecs(t, "Volume Provisioner E2E Test Suite")
}

var _ = ginkgo.SynchronizedAfterSuite(func() {
	framework.Logf("Running AfterSuite actions on all node")
	framework.RunCleanupActions()

	// Only delete resources if we aquired the lock and deployed them in the
	// first place.
	if lockAquired && !installDisabled {
		cs, err := framework.NewClientSetFromFlags()
		Ω(err).ShouldNot(HaveOccurred())

		framework.DeleteFlexvolumeDriver(cs)
		framework.DeleteVolumeProvisioner(cs)
	}
}, func() {})
