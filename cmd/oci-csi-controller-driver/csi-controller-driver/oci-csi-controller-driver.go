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

package csicontrollerdriver

import (
	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-controller-driver/csioptions"
	"github.com/oracle/oci-cloud-controller-manager/pkg/csi/driver"
	"github.com/oracle/oci-cloud-controller-manager/pkg/logging"
	"go.uber.org/zap"
)

//StartControllerDriver main function to start CSI Controller Driver
func StartControllerDriver(csioptions csioptions.CSIOptions) {

	logger := logging.Logger().Sugar()
	logger.Sync()

	drv, err := driver.NewControllerDriver(logger.Named("BV").With(zap.String("component", "csi-controller")), csioptions.Endpoint, csioptions.Kubeconfig, csioptions.Master,
		true, driver.BlockVolumeDriverName, driver.BlockVolumeDriverVersion)
	if err != nil {
		logger.With(zap.Error(err)).Fatal("Failed to create controller driver.")
	}

	if err := drv.Run(); err != nil {
		logger.With(zap.Error(err)).Fatal("Failed to run the CSI driver.")
	}

	logger.Info("CSI driver exited")
}
