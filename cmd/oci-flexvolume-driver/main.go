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
	"log"
	"os"

	"github.com/oracle/oci-cloud-controller-manager/pkg/flexvolume"
	"github.com/oracle/oci-cloud-controller-manager/pkg/flexvolume/block"
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
	// TODO: Maybe use sirupsen/logrus?
	f, err := os.OpenFile(GetLogPath(), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening log file: %v", err)
		os.Exit(1)
	}
	defer f.Close()

	log.SetPrefix(fmt.Sprintf("%d ", os.Getpid()))

	log.SetOutput(f)

	log.Printf("OCI FlexVolume Driver version: %s (%s)", version, build)
	d, err := block.NewOCIFlexvolumeDriver()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating new driver: %v", err)
		log.Printf("error creating new driver: %v", err)
		os.Exit(1)
	}
	flexvolume.ExecDriver(d, os.Args)
}
