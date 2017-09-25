// Copyright 2017 The OCI Cloud Controller Manager Authors
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
