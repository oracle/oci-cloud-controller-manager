package e2e

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"k8s.io/apiserver/pkg/util/logs"

	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework/ginkgowrapper"
)

func TestE2E(t *testing.T) {
	logs.InitLogs()
	defer logs.FlushLogs()

	gomega.RegisterFailHandler(ginkgowrapper.Fail)
	RunSpecs(t, "CCM E2E suite")
}
