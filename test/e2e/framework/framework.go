// Copyright 2018 Oracle and/or its affiliates. All rights reserved.
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
	"math/rand"
	"strings"
	"time"

	oke "github.com/oracle/oci-go-sdk/v65/containerengine"
	imageutils "k8s.io/kubernetes/test/utils/image"
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

	JobCompletionTimeout       = 5 * time.Minute
	deploymentAvailableTimeout = 5 * time.Minute
	CloneAvailableTimeout      = 10 * time.Minute

	DefaultClusterKubeconfig = "/tmp/clusterkubeconfig"
	DefaultCloudConfig       = "/tmp/cloudconfig"

	ClassOCI           = "oci"
	ClassOCICSI        = "oci-bv"
	ClassCustom        = "oci-bv-custom"
	ClassOCICSIExpand  = "oci-bv-expand"
	ClassOCILowCost    = "oci-bv-low"
	ClassOCIBalanced   = "oci-bal"
	ClassOCIHigh       = "oci-bv-high"
	ClassOCIUHP        = "oci-uhp"
	ClassOCIKMS        = "oci-kms"
	ClassOCIExt3       = "oci-ext3"
	ClassOCIXfs        = "oci-xfs"
	ClassFssDynamic    = "oci-file-storage-test"
	FssProvisionerType = "fss.csi.oraclecloud.com"
	ClassSnapshot      = "oci-snapshot-sc"
	MinVolumeBlock     = "50Gi"
	MaxVolumeBlock     = "100Gi"
	VolumeFss          = "1Gi"

	VSClassDefault    = "oci-snapclass"
	NodeHostnameLabel = "kubernetes.io/hostname"
)

var (
	compartment1                  string
	adlocation                    string
	clusterkubeconfig             string // path to kubeconfig file
	deleteNamespace               bool   // whether or not to delete test namespaces
	cloudConfigFile               string // path to cloud provider config file
	nodePortTest                  bool   // whether or not to test the connectivity of node ports.
	ccmSeclistID                  string // The ocid of the loadbalancer subnet seclist. Optional.
	k8sSeclistID                  string // The ocid of the k8s worker subnet seclist. Optional.
	mntTargetOCID                 string // Mount Target ID is specified to identify the mount target to be attached to the volumes. Optional.
	mntTargetSubnetOCID           string // mntTargetSubnetOCID is required for testing MT creation in FSS dynamic
	mntTargetCompartmentOCID      string // mntTargetCompartmentOCID is required for testing MT cross compartment creation in FSS dynamic
	nginx                         string // Image for nginx
	agnhost                       string // Image for agnhost
	busyBoxImage                  string // Image for busyBoxImage
	centos                        string // Image for centos
	imagePullRepo                 string // Repo to pull images from. Will pull public images if not specified.
	cmekKMSKey                    string // KMS key for CMEK testing
	nsgOCIDS                      string // Testing CCM NSG feature
	backendNsgIds                 string // Testing Rule management Backend NSG feature
	reservedIP                    string // Testing public reserved IP feature
	architecture                  string
	volumeHandle                  string // The FSS mount volume handle
	lustreVolumeHandle			  string // The Lustre mount volume handle
	lustreSubnetCidr              string // The Lustre Subnet Cidr
	staticSnapshotCompartmentOCID string // Compartment ID for cross compartment snapshot test
	runUhpE2E                     bool   // Whether to run UHP E2Es, requires Volume Management Plugin enabled on the node and 16+ cores (check blockvolumeperformance public doc for the exact requirements)
	enableParallelRun			  bool
	addOkeSystemTags              bool
	clusterID                     string              // Ocid of the newly created E2E cluster
	clusterType                   string              // Cluster type can be BASIC_CLUSTER or ENHANCED_CLUSTER (Default: BASIC_CLUSTER)
	clusterTypeEnum               oke.ClusterTypeEnum // Enum for OKE Cluster Type
	maxPodsPerNode                int
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

	flag.StringVar(&mntTargetOCID, "mnt-target-id", "", "Mount Target ID is required for creating storage class for FSS dynamic testing")
	flag.StringVar(&mntTargetSubnetOCID, "mnt-target-subnet-id", "", "Mount Target Subnet is required for creating storage class for FSS dynamic testing")
	flag.StringVar(&mntTargetCompartmentOCID, "mnt-target-compartment-id", "", "Mount Target Compartment is required for creating storage class for FSS dynamic testing with cross compartment")
	flag.StringVar(&volumeHandle, "volume-handle", "", "FSS volume handle used to mount the File System")
	flag.StringVar(&lustreVolumeHandle, "lustre-volume-handle", "", "Lustre volume handle used to mount the File System")
	flag.StringVar(&lustreSubnetCidr, "lustre-subnet-cidr", "", "Lustre subnet cidr to identify SVNIC in lustre subnet to configure lnet.")

	flag.StringVar(&imagePullRepo, "image-pull-repo", "", "Repo to pull images from. Will pull public images if not specified.")
	flag.StringVar(&cmekKMSKey, "cmek-kms-key", "", "KMS key to be used for CMEK testing")
	flag.StringVar(&nsgOCIDS, "nsg-ocids", "", "NSG OCIDs to be used to associate to LB")
	flag.StringVar(&backendNsgIds, "backend-nsg-ocids", "", "backend NSG Ids associated with backends of LB")
	flag.StringVar(&reservedIP, "reserved-ip", "", "Public reservedIP to be used for testing loadbalancer with reservedIP")
	flag.StringVar(&architecture, "architecture", "", "CPU architecture to be used for testing.")

	flag.StringVar(&staticSnapshotCompartmentOCID, "static-snapshot-compartment-id", "", "Compartment ID for cross compartment snapshot test")
	flag.BoolVar(&runUhpE2E, "run-uhp-e2e", false, "Run UHP E2Es as well")
	flag.BoolVar(&enableParallelRun, "enable-parallel-run", true, "Enables parallel running of test suite")
	flag.BoolVar(&addOkeSystemTags, "add-oke-system-tags", false, "Adds oke system tags to new and existing loadbalancers and storage resources")
	flag.IntVar(&maxPodsPerNode, "maxpodspernode", MAX_PODS_PER_NODE, "maxPods per node for OCI_VCN_IP_NATIVE")
	flag.StringVar(&clusterType, "cluster-type", "BASIC_CLUSTER", "Cluster type can be BASIC_CLUSTER or ENHANCED_CLUSTER")
}

// MAX_PODS_PER_NODE : If CNI_TYPE is OCI_VCN_NATIVE MAX_PODS_PER_NODE is set to 12
const MAX_PODS_PER_NODE = 12

// Framework is the context of the text execution.
type Framework struct {
	// The compartment1 the cluster is running in.
	Compartment1 string

	// Cluster Type
	ClusterType oke.ClusterTypeEnum

	// Default adLocation
	AdLocation string

	// Default adLocation
	AdLabel string

	//is cluster creation required
	EnableCreateCluster bool

	ClusterKubeconfigPath string

	CloudConfigPath string

	MntTargetOcid            string
	MntTargetSubnetOcid      string
	MntTargetCompartmentOcid string
	CMEKKMSKey               string
	NsgOCIDS                 string
	BackendNsgOcid           string
	ReservedIP               string
	Architecture             string

	VolumeHandle string
	LustreVolumeHandle string

	LustreSubnetCidr string

	// Compartment ID for cross compartment snapshot test
	StaticSnapshotCompartmentOcid string
	RunUhpE2E                     bool
	AddOkeSystemTags        bool
}

// New creates a new a framework that holds the context of the test
// execution.
func New() *Framework {
	flag.Parse()
	return NewWithConfig()
}

// NewWithConfig creates a new Framework instance and configures the instance as per the configuration options in the given config.
func NewWithConfig() *Framework {
	rand.Seed(time.Now().UTC().UnixNano())

	f := &Framework{
		AdLocation:                    adlocation,
		MntTargetOcid:                 mntTargetOCID,
		MntTargetSubnetOcid:           mntTargetSubnetOCID,
		MntTargetCompartmentOcid:      mntTargetCompartmentOCID,
		CMEKKMSKey:                    cmekKMSKey,
		NsgOCIDS:                      nsgOCIDS,
		ReservedIP:                    reservedIP,
		VolumeHandle:                  volumeHandle,
		LustreVolumeHandle:            lustreVolumeHandle,
		LustreSubnetCidr:              lustreSubnetCidr,
		StaticSnapshotCompartmentOcid: staticSnapshotCompartmentOCID,
		RunUhpE2E:                     runUhpE2E,
		AddOkeSystemTags:              addOkeSystemTags,
		ClusterType:                   clusterTypeEnum,
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
	f.MntTargetSubnetOcid = mntTargetSubnetOCID
	Logf("OCI Mount Target Subnet OCID: %s", f.MntTargetSubnetOcid)
	f.MntTargetCompartmentOcid = mntTargetCompartmentOCID
	Logf("OCI Mount Target Compartment OCID: %s", f.MntTargetCompartmentOcid)
	f.VolumeHandle = volumeHandle
	Logf("FSS Volume Handle is : %s", f.VolumeHandle)
	f.LustreVolumeHandle = lustreVolumeHandle
	Logf("Lustre Volume Handle is : %s", f.LustreVolumeHandle)
	f.LustreSubnetCidr = lustreSubnetCidr
	Logf("Lustre Subnet CIDR is : %s", f.LustreSubnetCidr)
	f.StaticSnapshotCompartmentOcid = staticSnapshotCompartmentOCID
	Logf("Static Snapshot Compartment OCID: %s", f.StaticSnapshotCompartmentOcid)
	f.RunUhpE2E = runUhpE2E
	Logf("Run Uhp E2Es as well: %v", f.RunUhpE2E)
	f.CMEKKMSKey = cmekKMSKey
	Logf("CMEK KMS Key: %s", f.CMEKKMSKey)
	f.NsgOCIDS = nsgOCIDS
	Logf("NSG OCIDS: %s", f.NsgOCIDS)
	f.BackendNsgOcid = backendNsgIds
	Logf("Backend NSG OCIDS: %s", f.BackendNsgOcid)
	f.ReservedIP = reservedIP
	Logf("Reserved IP: %s", f.ReservedIP)
	f.Architecture = architecture
	Logf("Architecture: %s", f.Architecture)
	f.Compartment1 = compartment1
	Logf("OCI compartment1 OCID: %s", f.Compartment1)
	f.MaxPodsPerNode = maxPodsPerNode
	Logf("Max pods per node: %s", f.MaxPodsPerNode)
	f.setImages()
	if strings.ToUpper(clusterType) == "ENHANCED_CLUSTER" {
		clusterTypeEnum = oke.ClusterTypeEnhancedCluster
	} else {
		clusterTypeEnum = oke.ClusterTypeBasicCluster
	}
	f.ClusterType = clusterTypeEnum
	Logf("Cluster Type: %s", f.ClusterType)
	f.ClusterKubeconfigPath = clusterkubeconfig
	f.CloudConfigPath = cloudConfigFile
}

func (f *Framework) setImages() {
	var Agnhost = "agnhost:2.6"
	var BusyBoxImage = "busybox:latest"
	var Nginx = "nginx:stable-alpine"
	var Centos = "centos:latest"

	if architecture == "ARM" {
		Agnhost = "agnhost-arm:2.6"
		BusyBoxImage = "busybox-arm:latest"
		Nginx = "nginx-arm:latest"
		Centos = "centos-arm:latest"
	}

	if imagePullRepo != "" {
		agnhost = fmt.Sprintf("%s%s", imagePullRepo, Agnhost)
		busyBoxImage = fmt.Sprintf("%s%s", imagePullRepo, BusyBoxImage)
		nginx = fmt.Sprintf("%s%s", imagePullRepo, Nginx)
		centos = fmt.Sprintf("%s%s", imagePullRepo, Centos)
	} else {
		agnhost = imageutils.GetE2EImage(imageutils.Agnhost)
		busyBoxImage = BusyBoxImage
		nginx = Nginx
		centos = Centos
	}
}

func (f *CloudProviderFramework) GetCompartmentId(setupF Framework) string {
	compartmentId := ""
	if setupF.Compartment1 != "" {
		compartmentId = setupF.Compartment1
	} else if f.CloudProviderConfig.CompartmentID != "" {
		compartmentId = f.CloudProviderConfig.CompartmentID
	} else if f.CloudProviderConfig.Auth.CompartmentID != "" {
		compartmentId = f.CloudProviderConfig.Auth.CompartmentID
	} else {
		Failf("Compartment Id undefined.")
	}
	return compartmentId
}
