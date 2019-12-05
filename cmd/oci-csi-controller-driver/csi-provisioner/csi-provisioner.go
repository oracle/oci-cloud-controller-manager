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

package csiprovisioner

import (
	"context"
	"fmt"
	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-controller-driver/csioptions"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kubernetes-csi/csi-lib-utils/leaderelection"
	ctrl "github.com/kubernetes-csi/external-provisioner/pkg/controller"
	snapclientset "github.com/kubernetes-csi/external-snapshotter/pkg/client/clientset/versioned"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/controller"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	utilfeature "k8s.io/apiserver/pkg/util/feature"
	csitranslationlib "k8s.io/csi-translation-lib"
)

var (
	provisionController *controller.ProvisionController
)

type leaderElection interface {
	Run() error
	WithNamespace(namespace string)
}

//StartCSIProvisioner main function to start CSI Controller Provisioner
func StartCSIProvisioner(csioptions csioptions.CSIOptions) {
	var config *rest.Config
	var err error
	version := "unknown"

	if err := utilfeature.DefaultMutableFeatureGate.SetFromMap(csioptions.FeatureGates); err != nil {
		klog.Fatal(err)
	}

	if csioptions.ShowVersion {
		fmt.Println(os.Args[0], version)
		os.Exit(0)
	}
	klog.Infof("Version: %s", version)

	// get the KUBECONFIG from env if specified (useful for local/debug cluster)
	kubeconfigEnv := os.Getenv("KUBECONFIG")

	if kubeconfigEnv != "" {
		klog.Infof("Found KUBECONFIG environment variable set, using that..")
		csioptions.Kubeconfig = kubeconfigEnv
	}

	if csioptions.Master != "" || csioptions.Kubeconfig != "" {
		klog.Infof("Either master or kubeconfig specified. building kube config from that..")
		config, err = clientcmd.BuildConfigFromFlags(csioptions.Master, csioptions.Kubeconfig)
	} else {
		klog.Infof("Building kube configs for running in cluster...")
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		klog.Fatalf("Failed to create config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Failed to create client: %v", err)
	}
	// snapclientset.NewForConfig creates a new Clientset for VolumesnapshotV1alpha1Client
	snapClient, err := snapclientset.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Failed to create snapshot client: %v", err)
	}

	// The controller needs to know what the server version is because out-of-tree
	// provisioners aren't officially supported until 1.5
	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		klog.Fatalf("Error getting server version: %v", err)
	}

	grpcClient, err := ctrl.Connect(csioptions.CsiAddress)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	err = ctrl.Probe(grpcClient, csioptions.OperationTimeout)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	// Autodetect provisioner name
	provisionerName, err := ctrl.GetDriverName(grpcClient, csioptions.OperationTimeout)
	if err != nil {
		klog.Fatalf("Error getting CSI driver name: %s", err)
	}
	klog.V(2).Infof("Detected CSI driver %s", provisionerName)

	pluginCapabilities, controllerCapabilities, err := ctrl.GetDriverCapabilities(grpcClient, csioptions.OperationTimeout)
	if err != nil {
		klog.Fatalf("Error getting CSI driver capabilities: %s", err)
	}

	// Generate a unique ID for this provisioner
	timeStamp := time.Now().UnixNano() / int64(time.Millisecond)
	identity := strconv.FormatInt(timeStamp, 10) + "-" + strconv.Itoa(rand.Intn(10000)) + "-" + provisionerName

	provisionerOptions := []func(*controller.ProvisionController) error{
		controller.LeaderElection(false), // Always disable leader election in provisioner lib. Leader election should be done here in the CSI provisioner level instead.
		controller.FailedProvisionThreshold(0),
		controller.FailedDeleteThreshold(0),
		controller.RateLimiter(workqueue.NewItemExponentialFailureRateLimiter(csioptions.RetryIntervalStart, csioptions.RetryIntervalMax)),
		controller.Threadiness(int(csioptions.WorkerThreads)),
		controller.CreateProvisionedPVLimiter(workqueue.DefaultControllerRateLimiter()),
	}

	supportsMigrationFromInTreePluginName := ""
	if csitranslationlib.IsMigratedCSIDriverByName(provisionerName) {
		supportsMigrationFromInTreePluginName, err = csitranslationlib.GetInTreeNameFromCSIName(provisionerName)
		if err != nil {
			klog.Fatalf("Failed to get InTree plugin name for migrated CSI plugin %s: %v", provisionerName, err)
		}
		klog.V(2).Infof("Supports migration from in-tree plugin: %s", supportsMigrationFromInTreePluginName)
		provisionerOptions = append(provisionerOptions, controller.AdditionalProvisionerNames([]string{supportsMigrationFromInTreePluginName}))
	}

	// Create the provisioner: it implements the Provisioner interface expected by
	// the controller
	csiProvisioner := ctrl.NewCSIProvisioner(clientset, csioptions.OperationTimeout, identity, csioptions.VolumeNamePrefix, csioptions.VolumeNameUUIDLength, grpcClient, snapClient, provisionerName, pluginCapabilities, controllerCapabilities, supportsMigrationFromInTreePluginName, csioptions.StrictTopology)
	provisionController = controller.NewProvisionController(
		clientset,
		provisionerName,
		csiProvisioner,
		serverVersion.GitVersion,
		provisionerOptions...,
	)

	run := func(context.Context) {
		provisionController.Run(wait.NeverStop)
	}

	if !csioptions.EnableLeaderElection {
		run(context.TODO())
	} else {
		// this lock name pattern is also copied from sigs.k8s.io/sig-storage-lib-external-provisioner/controller
		// to preserve backwards compatibility
		lockName := strings.Replace(provisionerName, "/", "-", -1)

		var le leaderElection
		if csioptions.LeaderElectionType == "endpoints" {
			klog.Warning("The 'endpoints' leader election type is deprecated and will be removed in a future release. Use '--leader-election-type=leases' instead.")
			le = leaderelection.NewLeaderElectionWithEndpoints(clientset, lockName, run)
		} else if csioptions.LeaderElectionType == "leases" {
			le = leaderelection.NewLeaderElection(clientset, lockName, run)
		} else {
			klog.Error("--leader-election-type must be either 'endpoints' or 'leases'")
			os.Exit(1)
		}

		if csioptions.LeaderElectionNamespace != "" {
			le.WithNamespace(csioptions.LeaderElectionNamespace)
		}

		if err := le.Run(); err != nil {
			klog.Fatalf("failed to initialize leader election: %v", err)
		}
	}

}
