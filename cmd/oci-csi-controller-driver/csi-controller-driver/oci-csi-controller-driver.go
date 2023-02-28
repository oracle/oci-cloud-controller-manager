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

const (
	bvCsiDriver = "BV"
)

//StartControllerDriver main function to start CSI Controller Driver
func StartControllerDriver(csioptions csioptions.CSIOptions, csiDriver driver.CSIDriver) {

	logger := logging.Logger().Sugar()
	logger.Sync()

	logger = logger.Named(string(csiDriver)).With(zap.String("component", "csi-controller"))
	var drv *driver.Driver
	var err error

	if csiDriver == bvCsiDriver {
		controllerDriverConfig := &driver.ControllerDriverConfig{CsiEndpoint: csioptions.Endpoint, CsiKubeConfig: csioptions.Kubeconfig, CsiMaster: csioptions.Master, EnableControllerServer: true, DriverName: driver.BlockVolumeDriverName, DriverVersion: driver.BlockVolumeDriverVersion}
		drv, err = driver.NewControllerDriver(logger, *controllerDriverConfig)
	} else {
		controllerDriverConfig := &driver.ControllerDriverConfig{CsiEndpoint: csioptions.FssEndpoint, CsiKubeConfig: csioptions.Kubeconfig, CsiMaster: csioptions.Master, EnableControllerServer: true, DriverName: driver.FSSDriverName, DriverVersion: driver.FSSDriverVersion}
		drv, err = driver.NewControllerDriver(logger, *controllerDriverConfig)
	}
	if err != nil {
		logger.With(zap.Error(err)).Fatal("Failed to create %s controller driver.", csiDriver)
	}
	if err := drv.Run(); err != nil {
		logger.With(zap.Error(err)).Fatal("Failed to run the CSI driver for %s.", csiDriver)
	}

	logger.Info("CSI driver exited")
}
