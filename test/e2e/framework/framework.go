// Copyright 2020 Oracle and/or its affiliates. All rights reserved.
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

package framework

import (
	"flag"
	"fmt"
	imageutils "k8s.io/kubernetes/test/utils/image"
	"math/rand"
	"strings"
	"time"
)

const (
	// Poll is the default polling period when checking lifecycle status.
	Poll = 15 * time.Second
	// Poll defines how regularly to poll kubernetes resources.
	K8sResourcePoll = 2 * time.Second
	// DefaultTimeout is how long we wait for long-running operations in the
	// test suite before giving up.
	DefaultTimeout = 10 * time.Minute
	// Some pods can take much longer to get ready due to volume attach/detach latency.
	slowPodStartTimeout = 15 * time.Minute

	DefaultClusterKubeconfig = "/tmp/clusterkubeconfig"
	DefaultCloudConfig       = "/tmp/cloudconfig"

	ClassOCI          = "oci"
	ClassOCICSI       = "oci-bv"
	ClassOCIExt3      = "oci-ext3"
	ClassOCIMntFss    = "oci-fss-mnt"
	ClassOCISubnetFss = "oci-fss-subnet"
	MinVolumeBlock    = "50Gi"
	MaxVolumeBlock    = "100Gi"
	VolumeFss         = "1Gi"
	Netexec           = "netexec:1.1"
	BusyBoxImage      = "busybox:latest"
	Nginx             = "nginx:stable-alpine"
	Centos            = "centos:latest"
)

var (
	compartment1      string
	adlocation        string
	clusterkubeconfig string // path to kubeconfig file
	deleteNamespace   bool   // whether or not to delete test namespaces
	cloudConfigFile   string // path to cloud provider config file
	nodePortTest      bool   // whether or not to test the connectivity of node ports.
	ccmSeclistID      string // The ocid of the loadbalancer subnet seclist. Optional.
	k8sSeclistID      string // The ocid of the k8s worker subnet seclist. Optional.
	mntTargetOCID     string // Mount Target ID is specified to identify the mount target to be attached to the volumes. Optional.
	nginx             string // Image for nginx
	netexec           string // Image for netexec
	busyBoxImage      string // Image for busyBoxImage
	centos            string // Image for centos
	imagePullRepo     string // Repo to pull images from. Will pull public images if not specified.
	cmekKMSKey        string //KMS key for CMEK testing
)

func init() {
	flag.StringVar(&compartment1, "compartment1", "", "OCID of the compartment1 in which to manage clusters.")

	flag.StringVar(&adlocation, "adlocation", "", "Default Ad Location.")

	//Below two flags need to be provided if test cluster already exists.
	flag.StringVar(&clusterkubeconfig, "cluster-kubeconfig", DefaultClusterKubeconfig, "Path to Cluster's Kubeconfig file with authorization and master location information. Only provide if test cluster exists.")
	flag.StringVar(&cloudConfigFile, "cloud-config", DefaultCloudConfig, "The path to the cloud provider configuration file. Empty string for no configuration file. Only provide if test cluster exists.")

	flag.BoolVar(&nodePortTest, "nodeport-test", false, "If true test will include 'nodePort' connectectivity tests.")
	flag.StringVar(&ccmSeclistID, "ccm-seclist-id", "", "The ocid of the loadbalancer subnet seclist. Enables additional seclist rule tests. If specified the 'k8s-seclist-id parameter' is also required.")
	flag.StringVar(&k8sSeclistID, "k8s-seclist-id", "", "The ocid of the k8s worker subnet seclist. Enables additional seclist rule tests. If specified the 'ccm-seclist-id parameter' is also required.")
	flag.BoolVar(&deleteNamespace, "delete-namespace", true, "If true tests will delete namespace after completion. It is only designed to make debugging easier, DO NOT turn it off by default.")

	flag.StringVar(&mntTargetOCID, "mnt-target-id", "", "Mount Target ID is specified to identify the mount target to be attached to the volumes")

	flag.StringVar(&imagePullRepo, "image-pull-repo", "", "Repo to pull images from. Will pull public images if not specified.")
	flag.StringVar(&cmekKMSKey, "cmek-kms-key", "", "KMS key to be used for CMEK testing")
	flag.Parse()
}

// Framework is the context of the text execution.
type Framework struct {
	// The compartment1 the cluster is running in.
	Compartment1 string
	// Default adLocation
	AdLocation string

	// Default adLocation
	AdLabel string

	//is cluster creation required
	EnableCreateCluster bool

	ClusterKubeconfigPath string

	CloudConfigPath string

	MntTargetOcid string
	CMEKKMSKey    string
}

// New creates a new a framework that holds the context of the test
// execution.
func New() *Framework {
	return NewWithConfig()
}

// NewWithConfig creates a new Framework instance and configures the instance as per the configuration options in the given config.
func NewWithConfig() *Framework {
	rand.Seed(time.Now().UTC().UnixNano())

	f := &Framework{
		AdLocation:    adlocation,
		MntTargetOcid: mntTargetOCID,
		CMEKKMSKey:    cmekKMSKey,
	}

	f.CloudConfigPath = cloudConfigFile
	f.ClusterKubeconfigPath = clusterkubeconfig

	f.Initialize()

	return f
}

// BeforeEach will be executed before each Ginkgo test is executed.
func (f *Framework) Initialize() {
	Logf("initializing framework")
	f.AdLocation = adlocation
	Logf("OCI AdLocation: %s", f.AdLocation)
	if adlocation != "" {
		splitString := strings.Split(adlocation, ":")
		if len(splitString) == 2 {
			f.AdLabel = splitString[1]
		} else {
			Failf("Invalid Availability Domain %s. Expecting format: `Uocm:PHX-AD-1`", adlocation)
		}
	}
	Logf("OCI AdLabel: %s", f.AdLabel)
	f.MntTargetOcid = mntTargetOCID
	Logf("OCI Mount Target OCID: %s", f.MntTargetOcid)
	f.CMEKKMSKey = cmekKMSKey
	Logf("CMEK KMS Key: %s", f.CMEKKMSKey)
	f.Compartment1 = compartment1
	Logf("OCI compartment1 OCID: %s", f.Compartment1)
	f.setImages()
	f.ClusterKubeconfigPath = clusterkubeconfig
	f.CloudConfigPath = cloudConfigFile
}

func (f *Framework) setImages() {
	if imagePullRepo != "" {
		netexec = fmt.Sprintf("%s%s", imagePullRepo, Netexec)
		busyBoxImage = fmt.Sprintf("%s%s", imagePullRepo, BusyBoxImage)
		nginx = fmt.Sprintf("%s%s", imagePullRepo, Nginx)
		centos = fmt.Sprintf("%s%s", imagePullRepo, Centos)
	} else {
		netexec = imageutils.GetE2EImage(imageutils.Netexec)
		busyBoxImage = BusyBoxImage
		nginx = Nginx
		centos = Centos
	}
}
