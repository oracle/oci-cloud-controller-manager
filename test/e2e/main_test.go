package e2e

import (
	"flag"
	"testing"

	"github.com/golang/glog"

	"k8s.io/apiserver/pkg/util/logs"

	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
)

var fw *framework.Framework

func TestMain(m *testing.M) {
	logs.InitLogs()
	defer logs.FlushLogs()

	kubeconfig := flag.String("kubeconfig", "", "Path to Kubeconfig file with authorization and master location information.")
	namespace := flag.String("namespace", "default", "The Kubernetes namespace within which test resources will be created.")
	flag.Parse()

	err := fw.Init(*kubeconfig, *namespace)
	if err != nil {
		glog.Fatal(err)
	}

	fw.Run(m.Run)
}

func init() {
	fw = framework.New()
}
