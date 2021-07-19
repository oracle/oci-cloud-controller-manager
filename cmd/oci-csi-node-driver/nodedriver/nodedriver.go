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

package nodedriver

import (
	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-node-driver/nodedriveroptions"
	"github.com/oracle/oci-cloud-controller-manager/pkg/csi/driver"
	"github.com/oracle/oci-cloud-controller-manager/pkg/logging"
	"go.uber.org/zap"
)

//RunNodeDriver main function to start node driver
func RunNodeDriver(nodecsioptions nodedriveroptions.NodeCSIOptions, stopCh <-chan struct{}) error {
	logger := logging.Logger().Sugar()
	logger.Sync()

	drv, err := driver.NewNodeDriver(logger, nodecsioptions.Endpoint, nodecsioptions.NodeID, nodecsioptions.Kubeconfig, nodecsioptions.Master)
	if err != nil {
		logger.With(zap.Error(err)).Fatal("Failed to create Node driver.")
	}

	if err := drv.Run(); err != nil {
		logger.With(zap.Error(err)).Fatal("Failed to run the CSI driver.")
	}

	logger.Info("CSI driver exited")
	<-stopCh
	return nil
}
