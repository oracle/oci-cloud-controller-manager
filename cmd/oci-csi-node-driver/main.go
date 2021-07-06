// Copyright 2019 Oracle and/or its affiliates. All rights reserved.
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

package main

import (
	"flag"
	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-node-driver/nodedriver"
	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-node-driver/nodedriveroptions"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util/signals"
	"k8s.io/klog"
)

func main() {
	nodecsioptions := nodedriveroptions.NodeCSIOptions{}

	flag.StringVar(&nodecsioptions.Endpoint, "endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	flag.StringVar(&nodecsioptions.NodeID, "nodeid", "", "node id")
	flag.StringVar(&nodecsioptions.LogLevel, "loglevel", "info", "log level")
	flag.StringVar(&nodecsioptions.Master, "master", "", "kube master")
	flag.StringVar(&nodecsioptions.Kubeconfig, "kubeconfig", "", "cluster kubeconfig")

	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()
	stopCh := signals.SetupSignalHandler()

	go nodedriver.RunNodeDriver(nodecsioptions, stopCh)
	<-stopCh
}
