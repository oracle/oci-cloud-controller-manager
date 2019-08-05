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
	goflag "flag"
	"github.com/oracle/oci-cloud-controller-manager/cmd/cloud-provider-oci/app"
	_ "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci"
	"github.com/oracle/oci-cloud-controller-manager/pkg/logging"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	_ "k8s.io/kubernetes/pkg/client/metrics/prometheus" // for client metric registration
	_ "k8s.io/kubernetes/pkg/version/prometheus"        // for version metric registration
	"math/rand"
	"syscall"
	"time"
)

var version string
var build string

func main() {
	syscall.Umask(0)
	rand.Seed(time.Now().UTC().UnixNano())

	log := logging.Logger()
	defer log.Sync()
	zap.ReplaceGlobals(log)
	logger := log.Sugar()

	command := app.NewCloudProviderOCICommand(logger)

	// TODO: once we switch everything over to Cobra commands, we can go back to calling
	// utilflag.InitFlags() (by removing its pflag.Parse() call). For now, we have to set the
	// normalize func and add the go flag set by hand.
	pflag.CommandLine.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	goflag.CommandLine.Parse([]string{})
	logs.InitLogs()
	defer logs.FlushLogs()

	logger.With("version", version, "build", build).Info("oci-cloud-controller-manager")

	// run the main cloud controller loop
	if err := command.Execute(); err != nil {
		logger.With(zap.Error(err)).Fatal("error running cloud provider OCI")
	}
}
