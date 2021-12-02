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
	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-node-driver/nodedriverregistrar"
	"github.com/oracle/oci-cloud-controller-manager/pkg/csi/driver"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util/signals"
	"k8s.io/klog"
)

func main() {
	nodecsioptions := nodedriveroptions.NodeCSIOptions{}

	flag.StringVar(&nodecsioptions.Endpoint, "endpoint", "unix://tmp/csi.sock", "Block Volume CSI endpoint")
	flag.StringVar(&nodecsioptions.NodeID, "nodeid", "", "node id")
	flag.StringVar(&nodecsioptions.LogLevel, "loglevel", "info", "log level")
	flag.StringVar(&nodecsioptions.Master, "master", "", "kube master")
	flag.StringVar(&nodecsioptions.Kubeconfig, "kubeconfig", "", "cluster kubeconfig")
	flag.DurationVar(&nodecsioptions.ConnectionTimeout, "connection-timeout", 0, "The --connection-timeout flag is deprecated")
	flag.StringVar(&nodecsioptions.CsiAddress, "csi-address", "/run/csi/socket", "Path of the Block Volume CSI driver socket that the node-driver-registrar will connect to.")
	flag.StringVar(&nodecsioptions.KubeletRegistrationPath, "kubelet-registration-path", "", "Path of the Block Volume CSI driver socket on the Kubernetes host machine.")
	flag.StringVar(&nodecsioptions.FssEndpoint, "fss-endpoint", "unix://tmp/fss/csi.sock", "FSS CSI endpoint")
	flag.StringVar(&nodecsioptions.FssCsiAddress, "fss-csi-address", "/run/fss/socket", "Path of the FSS CSI driver socket that the node-driver-registrar will connect to.")
	flag.StringVar(&nodecsioptions.FssKubeletRegistrationPath, "fss-kubelet-registration-path", "", "Path of the FSS CSI driver socket on the Kubernetes host machine.")
	flag.BoolVar(&nodecsioptions.EnableFssDriver, "fss-csi-driver-enabled", false, "Handle flag to enable FSS CSI driver")

	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()

	blockvolumeNodeOptions := nodedriveroptions.NodeOptions{
		Name:                   "BV",
		Endpoint:               nodecsioptions.Endpoint,
		NodeID:                 nodecsioptions.NodeID,
		Kubeconfig:             nodecsioptions.Kubeconfig,
		Master:                 nodecsioptions.Master,
		DriverName:             driver.BlockVolumeDriverName,
		DriverVersion:          driver.BlockVolumeDriverVersion,
		EnableControllerServer: false,
	}
	fssNodeOptions := nodedriveroptions.NodeOptions{
		Name:                   "FSS",
		Endpoint:               nodecsioptions.FssEndpoint,
		NodeID:                 nodecsioptions.NodeID,
		Kubeconfig:             nodecsioptions.Kubeconfig,
		Master:                 nodecsioptions.Master,
		DriverName:             driver.FSSDriverName,
		DriverVersion:          driver.FSSDriverVersion,
		EnableControllerServer: false,
	}

	stopCh := signals.SetupSignalHandler()
	go nodedriver.RunNodeDriver(blockvolumeNodeOptions, stopCh)
	if nodecsioptions.EnableFssDriver {
		go nodedriver.RunNodeDriver(fssNodeOptions, stopCh)
	}
	<-stopCh
}
