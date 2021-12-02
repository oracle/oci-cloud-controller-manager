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
func RunNodeDriver(nodeOptions nodedriveroptions.NodeOptions, stopCh <-chan struct{}) error {
	logger := logging.Logger().Sugar()
	logger.Sync()

	csiDriver, err := driver.NewNodeDriver(logger.Named(nodeOptions.Name), nodeOptions)
	if err != nil {
		logger.With(zap.Error(err)).Fatalf("Failed to create %s Node driver.", nodeOptions.Name)
	}

	if err := csiDriver.Run(); err != nil {
		logger.With(zap.Error(err)).Fatalf("Failed to run the %s CSI driver.", nodeOptions.Name)
	}
	logger.Infof("%s CSI driver exited", nodeOptions.Name)
	<-stopCh
	return nil
}
