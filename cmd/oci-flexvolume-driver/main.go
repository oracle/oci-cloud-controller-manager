// Copyright 2017 Oracle and/or its affiliates. All rights reserved.
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
	"fmt"
	"os"

	"github.com/oracle/oci-cloud-controller-manager/pkg/flexvolume"
	"github.com/oracle/oci-cloud-controller-manager/pkg/flexvolume/block"
	"github.com/oracle/oci-cloud-controller-manager/pkg/logging"

	"go.uber.org/zap"
)

// version/build is set at build time to the version of the driver being built.
var version string
var build string

// GetLogPath returns the default path to the driver log file.
func GetLogPath() string {
	path := os.Getenv("OCI_FLEXD_DRIVER_LOG_DIR")
	if path == "" {
		path = block.GetDriverDirectory()
	}
	return path + "/oci_flexvolume_driver.log"
}

func main() {
	l := logging.FileLogger(GetLogPath())
	defer l.Sync()
	zap.ReplaceGlobals(l)
	logger := l.Sugar()
	logger = logger.With(
		"pid", os.Getpid(),
		"version", version,
		"build", build,
		"component", "flexvolume-driver",
	)
	logger.Debug("OCI Flexvolume driver")
	d, err := block.NewOCIFlexvolumeDriver(logger)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating new driver: %v", err)
		logger.With(zap.Error(err)).Error("Error creating new driver")
		os.Exit(1)
	}
	flexvolume.ExecDriver(logger, d, os.Args)
}
