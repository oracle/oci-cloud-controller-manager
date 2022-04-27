// Copyright 2020 Oracle and/or its affiliates. All rights reserved.
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
	"github.com/onsi/ginkgo"
	sharedfw "github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
)

var setupF *sharedfw.Framework

var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {

	setupF = sharedfw.New()

	sharedfw.Logf("CloudProviderFramework Setup")
	sharedfw.Logf("Running tests with existing cluster.")
	return nil
}, func(data []byte) {})

var _ = ginkgo.SynchronizedAfterSuite(func() {
	sharedfw.Logf("Running AfterSuite actions on all node")
	sharedfw.RunCleanupActions()
}, func() {})
