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
	"github.com/oracle/oci-cloud-controller-manager/pkg/csi/driver"
	"github.com/oracle/oci-cloud-controller-manager/pkg/logging"
	"go.uber.org/zap"
)

var (
	endpoint   = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	nodeID     = flag.String("nodeid", "", "node id")
	logLevel   = flag.String("loglevel", "info", "log level")
	master     = flag.String("master", "", "kube master")
	kubeconfig = flag.String("kubeconfig", "", "cluster kubeconfig")
)

func main() {
	flag.Parse()

	logger := logging.Logger().Sugar()
	logger.Sync()

	drv, err := driver.NewNodeDriver(logger, *endpoint, *nodeID, *kubeconfig, *master)
	if err != nil {
		logger.With(zap.Error(err)).Fatal("Failed to create Node driver.")
	}

	if err := drv.Run(); err != nil {
		logger.With(zap.Error(err)).Fatal("Failed to run the CSI driver.")
	}

	logger.Info("CSI driver exited")
}
