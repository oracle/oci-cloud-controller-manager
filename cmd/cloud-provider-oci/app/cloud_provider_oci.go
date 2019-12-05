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
	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-controller-driver/csi-controller"
	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-controller-driver/csioptions"
	"k8s.io/apiserver/pkg/util/term"
	"k8s.io/component-base/cli/globalflag"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	provisioner "github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/core"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	cliflag "k8s.io/component-base/cli/flag"
	utilflag "k8s.io/component-base/cli/flag"
	cloudControllerManager "k8s.io/kubernetes/cmd/cloud-controller-manager/app"
	cloudControllerManagerConfig "k8s.io/kubernetes/cmd/cloud-controller-manager/app/config"
	"k8s.io/kubernetes/cmd/cloud-controller-manager/app/options"
	"k8s.io/kubernetes/pkg/version/verflag"
)

var (
	minVolumeSize, resourcePrincipalFile                                             string
	enableCSI, enableVolumeProvisioning, volumeRoundingEnabled, useResourcePrincipal bool
	resourcePrincipalInitialTimeout                                                  time.Duration
)

var csioption = csioptions.CSIOptions{}

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

	// volume provisioner flag set
	vpFlagSet := namedFlagSets.FlagSet("volume provisioner")
	vpFlagSet.BoolVar(&enableVolumeProvisioning, "enable-volume-provisioning", true, "When enabled volumes will be provisioned/deleted by cloud controller manager")
	vpFlagSet.BoolVar(&volumeRoundingEnabled, "rounding-enabled", true, "When enabled volumes will be rounded up if less than 'minVolumeSizeMB'")
	vpFlagSet.StringVar(&minVolumeSize, "min-volume-size", "50Gi", "The minimum size for a block volume. By default OCI only supports block volumes > 50GB")

	// oci authentication mode flag set
	ociAuthFlagSet := namedFlagSets.FlagSet("oci authentication modes")
	ociAuthFlagSet.BoolVar(&useResourcePrincipal, "use-resource-principal", false, "If true use resource principal as authentication mode else use service principal as authentication mode")
	ociAuthFlagSet.StringVar(&resourcePrincipalFile, "resource-principal-file", "", "The filesystem path at which the serialized Resource Principal is stored")
	ociAuthFlagSet.DurationVar(&resourcePrincipalInitialTimeout, "resource-principal-initial-timeout", 1*time.Minute, "How long to wait for an initial Resource Principal before terminating with an error if one is not supplied")

	// csi flag set.
	csiFlagSet := namedFlagSets.FlagSet("CSI Controller")
	csiFlagSet.BoolVar(&enableCSI, "csi-enabled", false, "Whether to enable CSI feature in OKE")
	csiFlagSet.StringVar(&csioption.CsiAddress, "csi-address", "/run/csi/socket", "Address of the CSI driver socket.")
	csiFlagSet.StringVar(&csioption.Endpoint, "csi-endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	csiFlagSet.StringVar(&csioption.VolumeNamePrefix, "csi-volume-name-prefix", "pvc", "Prefix to apply to the name of a created volume.")
	csiFlagSet.IntVar(&csioption.VolumeNameUUIDLength, "csi-volume-name-uuid-length", -1, "Truncates generated UUID of a created volume to this length. Defaults behavior is to NOT truncate.")
	csiFlagSet.BoolVar(&csioption.ShowVersion, "csi-version", false, "Show version.")
	csiFlagSet.DurationVar(&csioption.RetryIntervalStart, "csi-retry-interval-start", time.Second, "Initial retry interval of failed provisioning or deletion. It doubles with each failure, up to retry-interval-max.")
	csiFlagSet.DurationVar(&csioption.RetryIntervalMax, "csi-retry-interval-max", 5*time.Minute, "Maximum retry interval of failed provisioning or deletion.")
	csiFlagSet.UintVar(&csioption.WorkerThreads, "csi-worker-threads", 100, "Number of provisioner worker threads, in other words nr. of simultaneous CSI calls.")
	csiFlagSet.DurationVar(&csioption.OperationTimeout, "csi-op-timeout", 10*time.Second, "Timeout for waiting for creation or deletion of a volume")
	csiFlagSet.BoolVar(&csioption.EnableLeaderElection, "csi-enable-leader-election", false, "Enables leader election. If leader election is enabled, additional RBAC rules are required. Please refer to the Kubernetes CSI documentation for instructions on setting up these RBAC rules.")
	csiFlagSet.StringVar(&csioption.LeaderElectionType, "csi-leader-election-type", "endpoints", "the type of leader election, options are 'endpoints' (default) or 'leases' (strongly recommended). The 'endpoints' option is deprecated in favor of 'leases'.")
	csiFlagSet.StringVar(&csioption.LeaderElectionNamespace, "csi-leader-election-namespace", "", "Namespace where the leader election resource lives. Defaults to the pod namespace if not set.")
	csiFlagSet.BoolVar(&csioption.StrictTopology, "csi-strict-topology", false, "Passes only selected node topology to CreateVolume Request, unlike default behavior of passing aggregated cluster topologies that match with topology keys of the selected node.")
	csiFlagSet.DurationVar(&csioption.Resync, "csi-resync", 10*time.Minute, "Resync interval of the controller.")
	csiFlagSet.DurationVar(&csioption.Timeout, "csi-timeout", 15*time.Second, "Timeout for waiting for attaching or detaching the volume.")
	csiFlagSet.Var(utilflag.NewMapStringBool(&csioption.FeatureGates), "csi-feature-gates", "A set of key=value pairs that describe feature gates for alpha/experimental features. ")

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

	if enableCSI == true {
		wg.Add(1)
		logger := logger.With(zap.String("component", "CSI controller driver"))
		logger.Info("CSI is enabled.")
		go func() {
			defer wg.Done()
			csioption.Master = options.Master
			csioption.Kubeconfig = options.Kubeconfig
			csicontroller.Run(csioption, ctx.Done())
			cancelFunc()
		}()
	} else {
		logger := logger.With(zap.String("component", "CSI controller driver"))
		logger.Info("CSI is disabled.")
	}

	// wait for all the go routines to finish.
	wg.Wait()
}
