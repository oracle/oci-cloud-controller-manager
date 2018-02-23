package e2e

import (
	"time"

	. "github.com/onsi/ginkgo"

	"k8s.io/api/core/v1"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
)

var _ = Describe("Service type:LoadBalancer", func() {
	f := framework.NewDefaultFramework("service")

	It("should be possible to create a Service type:LoadBalancer", func() {
		serviceName := "basic-lb-test"
		ns := f.Namespace.Name

		jig := framework.NewServiceTestJig(f.ClientSet, serviceName)
		jig.RunOrFail(ns, nil)

		svc := jig.CreateLoadBalancerService(ns, serviceName, 5*time.Minute, func(s *v1.Service) {
			s.Annotations = map[string]string{oci.ServiceAnnotationLoadBalancerShape: "400Mbps"}
		})
		jig.SanityCheckService(svc, v1.ServiceTypeLoadBalancer)

		By("hitting the TCP service's NodePort")
		jig.TestReachableHTTP(nodeIP, tcpNodePort, framework.KubeProxyLagTimeout)
	})
})
