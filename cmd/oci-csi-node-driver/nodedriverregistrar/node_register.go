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

package nodedriverregistrar

import (
	"fmt"
	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-node-driver/nodedriveroptions"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	"k8s.io/klog"
	registerapi "k8s.io/kubernetes/pkg/kubelet/apis/pluginregistration/v1alpha1"
	"net"
	"os"
	"runtime"
)

//nodeRegister is the main function to start node register
func nodeRegister(csiDriverName string, nodecsioptions nodedriveroptions.NodeCSIOptions) {
	// When kubeletRegistrationPath is specified then driver-registrar ONLY acts
	// as gRPC server which replies to registration requests initiated by kubelet's
	// pluginswatcher infrastructure. Node labeling is done by kubelet's csi code.
	registrar := newRegistrationServer(csiDriverName, nodecsioptions.KubeletRegistrationPath, []string{"1.0.0"})
	socketPath := fmt.Sprintf("/registration/%s-reg.sock", csiDriverName)
	fi, err := os.Stat(socketPath)
	if err == nil && (fi.Mode()&os.ModeSocket) != 0 {
		// Remove any socket, stale or not, but fall through for other files
		if err := os.Remove(socketPath); err != nil {
			klog.Errorf("failed to remove stale socket %s with error: %+v", socketPath, err)
			os.Exit(1)
		}
	}
	if err != nil && !os.IsNotExist(err) {
		klog.Errorf("failed to stat the socket %s with error: %+v", socketPath, err)
		os.Exit(1)
	}

	var oldmask int
	if runtime.GOOS == "linux" {
		// Default to only user accessible socket, caller can open up later if desired
		oldmask, _ = umaskLinux(0077)
	}

	klog.Infof("Starting Registration Server at: %s\n", socketPath)
	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		klog.Errorf("failed to listen on socket: %s with error: %+v", socketPath, err)
		os.Exit(1)
	}
	if runtime.GOOS == "linux" {
		umaskLinux(oldmask)
	}
	klog.Infof("Registration Server started at: %s\n", socketPath)
	grpcServer := grpc.NewServer()
	// Registers kubelet plugin watcher api.
	registerapi.RegisterRegistrationServer(grpcServer, registrar)

	// Starts service
	if err := grpcServer.Serve(lis); err != nil {
		klog.Errorf("Registration Server stopped serving: %v", err)
		os.Exit(1)
	}
	// If gRPC server is gracefully shutdown, exit
	os.Exit(0)
}

func umaskLinux(mask int) (int, error) {
	return unix.Umask(mask), nil
}
