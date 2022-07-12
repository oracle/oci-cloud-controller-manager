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

	JobCompletionTimeout       = 5 * time.Minute
	deploymentAvailableTimeout = 5 * time.Minute

	DefaultClusterKubeconfig = "/tmp/clusterkubeconfig"
	DefaultCloudConfig       = "/tmp/cloudconfig"

	ClassOCI          = "oci"
	ClassOCICSI       = "oci-bv"
	ClassOCICSIExpand = "oci-bv-expand"
	ClassOCILowCost   = "oci-bv-low"
	ClassOCIBalanced  = "oci-bal"
	ClassOCIHigh      = "oci-bv-high"
	ClassOCIKMS       = "oci-kms"
	ClassOCIExt3      = "oci-ext3"
	ClassOCIXfs       = "oci-xfs"
	MinVolumeBlock    = "50Gi"
	MaxVolumeBlock    = "100Gi"
	VolumeFss         = "1Gi"
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
	agnhost           string // Image for agnhost
	busyBoxImage      string // Image for busyBoxImage
	centos            string // Image for centos
	imagePullRepo     string // Repo to pull images from. Will pull public images if not specified.
	cmekKMSKey        string //KMS key for CMEK testing
	nsgOCIDS		  string // Testing CCM NSG feature
	reservedIP        string // Testing public reserved IP feature
	architecture      string
	volumeHandle      string // The FSS mount volume handle
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
	flag.StringVar(&volumeHandle, "volume-handle", "", "FSS volume handle used to mount the File System")

	flag.StringVar(&imagePullRepo, "image-pull-repo", "", "Repo to pull images from. Will pull public images if not specified.")
	flag.StringVar(&cmekKMSKey, "cmek-kms-key", "", "KMS key to be used for CMEK testing")
	flag.StringVar(&nsgOCIDS, "nsg-ocids", "", "NSG OCIDs to be used to associate to LB")
	flag.StringVar(&reservedIP, "reserved-ip", "", "Public reservedIP to be used for testing loadbalancer with reservedIP")
	flag.StringVar(&architecture, "architecture", "", "CPU architecture to be used for testing.")
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
	NsgOCIDS      string
	ReservedIP    string
	Architecture  string

	VolumeHandle string
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
		AdLocation:    adlocation,
		MntTargetOcid: mntTargetOCID,
		CMEKKMSKey:    cmekKMSKey,
		NsgOCIDS:	   nsgOCIDS,
		ReservedIP:    reservedIP,
		VolumeHandle:  volumeHandle,
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
	f.VolumeHandle = volumeHandle
	Logf("FSS Volume Handle is : %s", f.VolumeHandle)
	f.CMEKKMSKey = cmekKMSKey
	Logf("CMEK KMS Key: %s", f.CMEKKMSKey)
	f.NsgOCIDS = nsgOCIDS
	Logf("NSG OCIDS: %s", f.NsgOCIDS)
	f.ReservedIP = reservedIP
	Logf("Reserved IP: %s", f.ReservedIP)
	f.Architecture = architecture
	Logf("Architecture: %s", f.Architecture)
	f.Compartment1 = compartment1
	Logf("OCI compartment1 OCID: %s", f.Compartment1)
	f.setImages()
	f.ClusterKubeconfigPath = clusterkubeconfig
	f.CloudConfigPath = cloudConfigFile
}

func (f *Framework) setImages() {
	var Agnhost           = "agnhost:2.6"
	var BusyBoxImage      = "busybox:latest"
	var Nginx             = "nginx:stable-alpine"
	var Centos            = "centos:latest"

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
