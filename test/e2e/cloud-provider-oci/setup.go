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
}, func(data []byte) {
	setupF = sharedfw.New()
},
)

var _ = ginkgo.SynchronizedAfterSuite(func() {}, func() {
	sharedfw.Logf("Running AfterSuite actions on all node")
	sharedfw.RunCleanupActions()
})
