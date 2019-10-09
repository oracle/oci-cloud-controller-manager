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

package app

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	provisioner "github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/core"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	cloudControllerManager "k8s.io/kubernetes/cmd/cloud-controller-manager/app"
	cloudControllerManagerConfig "k8s.io/kubernetes/cmd/cloud-controller-manager/app/config"
	"k8s.io/kubernetes/cmd/cloud-controller-manager/app/options"
	"k8s.io/kubernetes/pkg/version/verflag"
)

var (
	master, kubeconfig, minVolumeSize               string
	enableVolumeProvisioning, volumeRoundingEnabled bool
)

// NewCloudProviderOCICommand creates a *cobra.Command object with default parameters
func NewCloudProviderOCICommand(logger *zap.SugaredLogger) *cobra.Command {

	// FIXME Create CLoudProviderOCIOptions struct that shall contain options for all the components
	s, err := options.NewCloudControllerManagerOptions()
	if err != nil {
		logger.With(zap.Error(err)).Fatalf("unable to initialize command options")
	}

	command := &cobra.Command{
		Use: "cloud-provider-oci",
		Long: `The cloud provider oci daemon is a agglomeration of oci cloud controller
manager and oci volume provisioner. It embeds the cloud specific control loops shipped with Kubernetes.`,
		Run: func(cmd *cobra.Command, args []string) {
			verflag.PrintAndExitIfRequested()
			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				logger.Infof("FLAG: --%s=%q", flag.Name, flag.Value)
			})

			c, err := s.Config()
			if err != nil {
				logger.With(zap.Error(err)).Fatalf("Unable to create cloud controller manager config")
			}

			run(logger, c.Complete(), s)

		},
	}

	// cloud controller manager flag set
	ccmFlagSet := pflag.NewFlagSet("cloud controller manager", pflag.ContinueOnError)
	s.AddFlags(ccmFlagSet)

	// volume provisioner flag set
	vpFlagSet := pflag.NewFlagSet("volume provisioner", pflag.ContinueOnError)
	vpFlagSet.BoolVar(&enableVolumeProvisioning, "enable-volume-provisioning", true, "When enabled volumes will be provisioned/deleted by cloud controller manager")
	vpFlagSet.BoolVar(&volumeRoundingEnabled, "rounding-enabled", true, "When enabled volumes will be rounded up if less than 'minVolumeSizeMB'")
	vpFlagSet.StringVar(&minVolumeSize, "min-volume-size", "50Gi", "The minimum size for a block volume. By default OCI only supports block volumes > 50GB")

	// add complete flag set to command
	command.Flags().AddFlagSet(ccmFlagSet)
	command.Flags().AddFlagSet(vpFlagSet)

	return command
}

func run(logger *zap.SugaredLogger, config *cloudControllerManagerConfig.CompletedConfig, options *options.CloudControllerManagerOptions) {
	var wg sync.WaitGroup
	ctx, cancelFunc := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 2)
	defer close(sigs)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancelFunc()
		<-sigs
		os.Exit(1)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger := logger.With(zap.String("component", "volume-provisioner"))
		if !enableVolumeProvisioning {
			logger.Info("Volume provisioning is disabled")
			return
		}
		// TODO Pass an options/config struct instead of config variables
		if err := provisioner.Run(logger, options.Kubeconfig, options.Master, minVolumeSize, volumeRoundingEnabled, ctx.Done()); err != nil {
			logger.With(zap.Error(err)).Error("Error running volume provisioner")
		}
		cancelFunc()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		runOnce := make(chan struct{}, 1)
		runOnce <- struct{}{}
		defer close(runOnce)
		// cloudControllerManager.Run does not accepts context. workaround to accept context
		for {
			select {
			case <-ctx.Done():
				return
			case <-runOnce:
				go func() {
					// Run starts all the cloud controller manager control loops.
					// TODO move to newer cloudControllerManager dependency that provides a way to pass channel/context
					if err := cloudControllerManager.Run(config); err != nil {
						logger.With(zap.Error(err)).Error("Error running cloud controller manager")
					}
					cancelFunc()
				}()

			}
		}
	}()

	// wait for all the go routines to finish.
	wg.Wait()
}
