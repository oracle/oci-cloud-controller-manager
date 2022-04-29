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
	"flag"
	"syscall"

	"github.com/oracle/oci-cloud-controller-manager/pkg/logging"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util/signals"
	provisioner "github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/core"
	"go.uber.org/zap"
)

// version/build is set at build time to the version of the provisioner being built.
var version string
var build string

func main() {
	syscall.Umask(0)

	log := logging.Logger()
	defer log.Sync()
	zap.ReplaceGlobals(log)

	kubeconfig := flag.String("kubeconfig", "", "Path to Kubeconfig file with authorization and master location information.")
	volumeRoundingEnabled := flag.Bool("rounding-enabled", true, "When enabled volumes will be rounded up if less than 'minVolumeSizeMB'")
	minVolumeSize := flag.String("min-volume-size", "50Gi", "The minimum size for a block volume. By default OCI only supports block volumes > 50GB")
	master := flag.String("master", "", "The address of the Kubernetes API server (overrides any value in kubeconfig).")
	flag.Parse()

	logger := log.Sugar()

	logger.With("version", version, "build", build, "component", "volume-provisioner").Info("oci-volume-provisioner")

	// Set up signals so we handle the shutdown signal gracefully.
	stopCh := signals.SetupSignalHandler()

	if err := provisioner.Run(logger, *kubeconfig, *master, *minVolumeSize, *volumeRoundingEnabled, stopCh); err != nil {
		logger.With(zap.Error(err)).Fatal("error running volume provisioner")
	}
}
