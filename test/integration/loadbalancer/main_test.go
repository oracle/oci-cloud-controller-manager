package loadbalancer

import (
	"testing"

	"k8s.io/apiserver/pkg/util/flag"
	"k8s.io/apiserver/pkg/util/logs"

	"github.com/golang/glog"
	"github.com/oracle/oci-cloud-controller-manager/test/integration/framework"
)

var fw *framework.Framework

func TestMain(m *testing.M) {
	logs.InitLogs()
	defer logs.FlushLogs()

	err := fw.Init()
	if err != nil {
		glog.Fatal(err)
	}

	fw.Run(m.Run)
}

func init() {
	flag.InitFlags()
	fw = framework.New()
}
