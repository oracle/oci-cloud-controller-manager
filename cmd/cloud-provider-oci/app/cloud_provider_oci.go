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
	"flag"
	"fmt"
	"github.com/oracle/oci-cloud-controller-manager/pkg/logging"
	_ "github.com/oracle/oci-cloud-controller-manager/pkg/oci/client" // for oci client metric registration
	provisioner "github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/core"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apiserver/pkg/util/term"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/cli/globalflag"
	cloudControllerManager "k8s.io/kubernetes/cmd/cloud-controller-manager/app"
	cloudControllerManagerConfig "k8s.io/kubernetes/cmd/cloud-controller-manager/app/config"
	"k8s.io/kubernetes/cmd/cloud-controller-manager/app/options"
	"k8s.io/kubernetes/pkg/version/verflag"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	logLevel                                                 int8
	minVolumeSize, metricsEndpoint, logfilePath              string
	enableVolumeProvisioning, volumeRoundingEnabled, logJSON bool
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
			log := logging.Logger()
			defer log.Sync()
			zap.ReplaceGlobals(log)
			logger = log.Sugar()
			verflag.PrintAndExitIfRequested()
			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				logger.Infof("FLAG: --%s=%q", flag.Name, flag.Value)
			})

			c, err := s.Config(cloudControllerManager.KnownControllers(), cloudControllerManager.ControllersDisabledByDefault.List())
			if err != nil {
				logger.With(zap.Error(err)).Fatalf("Unable to create cloud controller manager config")
			}

			run(logger, c.Complete(), s)

		},
	}

	namedFlagSets := s.Flags(cloudControllerManager.KnownControllers(), cloudControllerManager.ControllersDisabledByDefault.List())
	// cloud controller manager flag set
	//ccmFlagSet := namedFlagSets.flagSet("cloud controller manager")
	//s.AddFlags(ccmFlagSet)

	// logging parameters flagset
	loggingFlagSet := namedFlagSets.FlagSet("logging variables")
	loggingFlagSet.Int8Var(&logLevel, "log-level", int8(zapcore.InfoLevel), "Adjusts the level of the logs that will be omitted.")
	loggingFlagSet.BoolVar(&logJSON, "log-json", false, "Log in json format.")
	loggingFlagSet.StringVar(&logfilePath, "logfile-path", "", "If specified, write log messages to a file at this path.")

	// prometheus metrics endpoint flagset
	metricsFlagSet := namedFlagSets.FlagSet("metrics endpoint")
	metricsFlagSet.StringVar(&metricsEndpoint, "metrics-endpoint", "0.0.0.0:8080", "The endpoint where to expose metrics")

	// volume provisioner flag set
	vpFlagSet := namedFlagSets.FlagSet("volume provisioner")
	vpFlagSet.BoolVar(&enableVolumeProvisioning, "enable-volume-provisioning", true, "When enabled volumes will be provisioned/deleted by cloud controller manager")
	vpFlagSet.BoolVar(&volumeRoundingEnabled, "rounding-enabled", true, "When enabled volumes will be rounded up if less than 'minVolumeSizeMB'")
	vpFlagSet.StringVar(&minVolumeSize, "min-volume-size", "50Gi", "The minimum size for a block volume. By default OCI only supports block volumes > 50GB")

	verflag.AddFlags(namedFlagSets.FlagSet("global"))
	globalflag.AddGlobalFlags(namedFlagSets.FlagSet("global"), command.Name())

	if flag.CommandLine.Lookup("cloud-provider-gce-lb-src-cidrs") != nil {
		// hoist this flag from the global flagset to preserve the commandline until
		// the gce cloudprovider is removed.
		globalflag.Register(namedFlagSets.FlagSet("generic"), "cloud-provider-gce-lb-src-cidrs")
	}
	for _, f := range namedFlagSets.FlagSets {
		command.Flags().AddFlagSet(f)
	}
	usageFmt := "Usage:\n  %s\n"
	cols, _, _ := term.TerminalSize(command.OutOrStdout())
	command.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Fprintf(cmd.OutOrStderr(), usageFmt, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStderr(), namedFlagSets, cols)
		return nil
	})
	command.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
	})

	viper.BindPFlags(command.Flags())

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
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(metricsEndpoint, nil); err != nil {
			logger.With(zap.Error(err)).Errorf("Error exposing metrics at %s/metrics", metricsEndpoint)
		}
		cancelFunc()
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
		// Run starts all the cloud controller manager control loops.
		// TODO move to newer cloudControllerManager dependency that provides a way to pass channel/context
		if err := cloudControllerManager.Run(config, ctx.Done()); err != nil {
			logger.With(zap.Error(err)).Error("Error running cloud controller manager")
		}
		cancelFunc()
	}()

	// wait for all the go routines to finish.
	wg.Wait()
}
