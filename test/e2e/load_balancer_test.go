package e2e

import (
	"testing"
	"time"

	"k8s.io/api/core/v1"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
)

func TestLoadBalancerBasic(t *testing.T) {
	serviceName := "basic-lb-test"

	jig := framework.NewServiceTestJig(t, fw.Client, serviceName)
	jig.RunOrFail(fw.Namespace, nil)

	svc := jig.CreateLoadBalancerService(fw.Namespace, serviceName, 5*time.Minute, func(s *v1.Service) {
		s.Annotations = map[string]string{oci.ServiceAnnotationLoadBalancerShape: "400Mbps"}
	})
	jig.SanityCheckService(svc, v1.ServiceTypeLoadBalancer)
}
