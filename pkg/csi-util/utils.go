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

package csi_util

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/oracle/oci-go-sdk/v65/core"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
	kubeAPI "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util/disk"
)

const (
	// minimumVolumeSizeInBytes is used to validate that the user is not trying
	// to create a volume that is smaller than what we support
	MinimumVolumeSizeInBytes int64 = 50 * client.GiB

	// maximumVolumeSizeInBytes is used to validate that the user is not trying
	// to create a volume that is larger than what we support
	MaximumVolumeSizeInBytes int64 = 32 * client.TiB

	// defaultVolumeSizeInBytes is used when the user did not provide a size or
	// the size they provided did not satisfy our requirements
	defaultVolumeSizeInBytes int64 = MinimumVolumeSizeInBytes

	waitForPathDelay = 1 * time.Second

	// ociVolumeBackupID is the name of the oci volume backup id annotation.
	ociVolumeBackupID = "volume.beta.kubernetes.io/oci-volume-source"

	// Block Volume Performance Units
	VpusPerGB                     = "vpusPerGB"
	LowCostPerformanceOption      = 0
	BalancedPerformanceOption     = 10
	HigherPerformanceOption       = 20
	MaxUltraHighPerformanceOption = 120

	InTransitEncryptionPackageName = "oci-fss-utils"
	FIPS_ENABLED_FILE_PATH         = "/host/proc/sys/crypto/fips_enabled"
	CAT_COMMAND                    = "cat"
	RPM_COMMAND                    = "rpm-host"
	LabelIpFamilyPreferred         = "oci.oraclecloud.com/ip-family-preferred"
	LabelIpFamilyIpv4              = "oci.oraclecloud.com/ip-family-ipv4"
	LabelIpFamilyIpv6              = "oci.oraclecloud.com/ip-family-ipv6"
	IscsiIpv6Prefix                = "fd00:00c1::"

	Ipv6Stack = "IPv6"
	Ipv4Stack = "IPv4"

	// For Raw Block Volumes, the name of the bind-mounted file inside StagingTargetPath
	RawBlockStagingFile = "mountfile"

	AvailabilityDomainLabel = "csi-ipv6-full-ad-name"

)

// Util interface
type Util struct {
	Logger *zap.SugaredLogger
}

var (
	DiskByPathPatternPV    = `/dev/disk/by-path/pci-\w{4}:\w{2}:\w{2}\.\d+-scsi-\d+:\d+:\d+:\d+$`
	DiskByPathPatternISCSI = `/dev/disk/by-path/ip-[[?\w\.\:]+]?:\d+-iscsi-[\w\.\-:]+-lun-\d+$`
)

type FSSVolumeHandler struct {
	FilesystemOcid       string
	MountTargetIPAddress string
	FsExportPath         string
}

type NodeMetadata struct {
	PreferredNodeIpFamily  string
	Ipv4Enabled            bool
	Ipv6Enabled            bool
	AvailabilityDomain     string
	FullAvailabilityDomain string
	IsNodeMetadataLoaded   bool
}

// CSIConfig represents the structure of the ConfigMap data.
type CSIConfig struct {
	Lustre *DriverConfig `yaml:"lustre"`
	IsLoaded bool
}

// DriverConfig represents driver-specific configurations.
type DriverConfig struct {
	SkipNodeUnstage bool `yaml:"skipNodeUnstage"`
	SkipLustreParameters bool `yaml:"skipLustreParameters"`

}

func (u *Util) LookupNodeID(k kubernetes.Interface, nodeName string) (string, error) {
	n, err := k.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		u.Logger.With(zap.Error(err)).With("node", nodeName).Error("Failed to get Node by name.")
		return "", fmt.Errorf("fail to get the node %s", nodeName)
	}
	if n.Spec.ProviderID == "" {
		u.Logger.With("node", nodeName).Error("ProvideID is missing.")
		return "", fmt.Errorf("missing provider id for node %s", nodeName)
	}
	u.Logger.With("node", nodeName).Info("Node is found.")
	return n.Spec.ProviderID, nil
}

func (u *Util) WaitForKubeApiServerToBeReachableWithContext(ctx context.Context, k kubernetes.Interface, backOffCap time.Duration) {

	waitForKubeApiServerCtx, waitForKubeApiServerCtxCancel := context.WithTimeout(ctx, time.Second * 45)
	defer waitForKubeApiServerCtxCancel()

	backoff := wait.Backoff{
		Duration: 1 * time.Second,
		Factor:   2.0,
		Steps:    5,
		Cap: backOffCap,
	}

	wait.ExponentialBackoffWithContext(
		waitForKubeApiServerCtx,
		backoff,
		func(waitForKubeApiServerCtx context.Context) (bool, error) {
			attemptCtx, attemptCancel := context.WithTimeout(waitForKubeApiServerCtx, backoff.Step())
			defer attemptCancel()
			_, err := k.CoreV1().RESTClient().Get().AbsPath("/readyz").Do(attemptCtx).Raw()
			if err != nil {
				u.Logger.With(zap.Error(err)).Errorf("Waiting for kube api server to be reachable, Retrying..")
				return false, nil
			}
			u.Logger.Infof("Kube Api Server is Reachable")
			return true, nil
		},
	)
}

func (u *Util) LoadNodeMetadataFromApiServer(ctx context.Context, k kubernetes.Interface, nodeID string, nodeMetadata *NodeMetadata) (error) {

	u.WaitForKubeApiServerToBeReachableWithContext(ctx, k, time.Second * 30)

	node, err := k.CoreV1().Nodes().Get(ctx, nodeID, metav1.GetOptions{})

	if err != nil {
		u.Logger.With(zap.Error(err)).With("nodeId", nodeID).Error("Failed to get Node information from kube api server, Please check if kube api server is accessible.")
		return fmt.Errorf("Failed to get node information from kube api server, please check if kube api server is accessible.")
	}

	var ok bool
	if node.Labels != nil {
		nodeMetadata.AvailabilityDomain, ok = node.Labels[kubeAPI.LabelTopologyZone]
		if !ok {
			nodeMetadata.AvailabilityDomain, ok = node.Labels[kubeAPI.LabelZoneFailureDomain]
		}
		if ok {
			nodeMetadata.FullAvailabilityDomain, _ = node.Labels[AvailabilityDomainLabel]
		}

		if preferredIpFamily, ok := node.Labels[LabelIpFamilyPreferred]; ok {
			nodeMetadata.PreferredNodeIpFamily = FormatValidIpStackInK8SConvention(preferredIpFamily)
		}
		if ipv4Enabled, ok := node.Labels[LabelIpFamilyIpv4]; ok && strings.EqualFold(ipv4Enabled, "true") {
			nodeMetadata.Ipv4Enabled = true
		}
		if ipv6Enabled, ok := node.Labels[LabelIpFamilyIpv6]; ok && strings.EqualFold(ipv6Enabled, "true") {
			nodeMetadata.Ipv6Enabled = true
		}
	}
	if !nodeMetadata.Ipv4Enabled && !nodeMetadata.Ipv6Enabled {
		nodeMetadata.PreferredNodeIpFamily = Ipv4Stack
		nodeMetadata.Ipv4Enabled = true
		u.Logger.With("nodeId", nodeID, "nodeMetadata", nodeMetadata).Info("No IP family labels identified on node, defaulting to ipv4.")
	} else {
		u.Logger.With("nodeId", nodeID, "nodeMetadata", nodeMetadata).Info("Node IP family identified.")
	}
	nodeMetadata.IsNodeMetadataLoaded = true
	return  nil
}

// waitForPathToExist waits for for a given filesystem path to exist.
func (u *Util) WaitForPathToExist(path string, maxRetries int) bool {
	for i := 0; i < maxRetries; i++ {
		var err error
		_, err = os.Stat(path)
		if err == nil {
			return true
		}
		if err != nil && !os.IsNotExist(err) {
			return false
		}
		if i == maxRetries-1 {
			break
		}
		time.Sleep(waitForPathDelay)
	}
	return false
}

// convert "zkJl:US-ASHBURN-AD-1" to "US-ASHBURN-AD-1"
func (u *Util) GetAvailableDomainInNodeLabel(fullAD string) string {
	adElements := strings.Split(fullAD, ":")
	if len(adElements) > 0 {
		realAD := adElements[len(adElements)-1]
		u.Logger.Infof("Converted %q to %q", fullAD, realAD)
		return realAD

	}
	u.Logger.With("fullAD", fullAD).Error("Available Domain for Node Label not found.")
	return ""
}

func ExtractISCSIInformation(attributes map[string]string) (*disk.Disk, error) {
	iqn, ok := attributes[disk.ISCSIIQN]
	if !ok {
		return nil, fmt.Errorf("unable to get the IQN from the attribute list")
	}
	iSCSIIp, ok := attributes[disk.ISCSIIP]
	if !ok {
		return nil, fmt.Errorf("unable to get the iSCSIIp from the attribute list")
	}
	port, ok := attributes[disk.ISCSIPORT]
	if !ok {
		return nil, fmt.Errorf("unable to get the port from the attribute list")
	}

	nPort, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %s, error: %v", port, err)
	}

	return &disk.Disk{
		IQN:     iqn,
		IscsiIp: iSCSIIp,
		Port:    nPort,
	}, nil
}

// Extracts the vpusPerGB as int64 from given string input
func ExtractBlockVolumePerformanceLevel(attribute string) (int64, error) {
	vpusPerGB, err := strconv.ParseInt(attribute, 10, 64)
	if err != nil {
		return 0, status.Errorf(codes.InvalidArgument, "unable to parse performance level value %s as int64", attribute)
	}
	if vpusPerGB < LowCostPerformanceOption || vpusPerGB > MaxUltraHighPerformanceOption {
		return 0, status.Errorf(codes.InvalidArgument, "invalid performance option : %s provided  for "+
			"storage class. Supported values for performance options are between %d and %d", attribute, LowCostPerformanceOption, MaxUltraHighPerformanceOption)
	}
	return vpusPerGB, nil
}

func ExtractISCSIInformationFromMountPath(logger *zap.SugaredLogger, diskPath []string) (*disk.Disk, error) {

	logger.Info("Getting ISCSIInfo for the mount path: ", diskPath)
	m, err := disk.FindFromMountPointPath(logger, diskPath)
	if err != nil {
		logger.With(zap.Error(err)).With("mount path", diskPath).Error("Invalid mount path")
		return nil, err
	}

	port, err := strconv.Atoi(m[2])
	if err != nil {
		logger.With(zap.Error(err)).With("mount path", diskPath, "port", port).Error("Invalid port")
		return nil, err
	}

	logger.With("IQN", m[3], "IscsiIP", m[1], "Port", port).Info("Found ISCSIInfo for the mount path: ", diskPath)
	return &disk.Disk{
		IQN:     m[3],
		IscsiIp: m[1],
		Port:    port,
	}, nil
}

func GetKubeClient(logger *zap.SugaredLogger, master, kubeconfig string) *kubernetes.Clientset {
	var (
		config *rest.Config
		err    error
	)

	if master != "" || kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags(master, kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			logger.With(zap.Error(err)).Fatal("Failed to get the kubeconfig in cluster.")
		}
	}

	kubeClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.With(zap.Error(err)).Fatal("Failed to create a kubernetes clientset.")
	} else {
		logger.Info("Created kubernetes client successfully.")
	}
	return kubeClientSet
}

// Get the staging target filepath inside the given stagingTargetPath, to be used for raw block volume support
func GetPathForBlock(volumePath string) string {
	pathForBlock := filepath.Join(volumePath, RawBlockStagingFile)
	return pathForBlock
}

// Creates a file on the specified path after creating the containing directory
func CreateFilePath(logger *zap.SugaredLogger, path string) error {
	pathDir := filepath.Dir(path)

	err := os.MkdirAll(pathDir, 0750)
	if err != nil {
		logger.With(zap.Error(err)).Fatal("failed to create surrounding directory")
		return err
	}

	file, fileErr := os.OpenFile(path, os.O_CREATE, 0640)
	if fileErr != nil && !os.IsExist(fileErr) {
		logger.With(zap.Error(err)).Fatal("failed to create/open the target file")
		return fileErr
	}

	fileErr = file.Close()
	if fileErr != nil {
		logger.With(zap.Error(err)).Fatal("failed to close the target file")
		return fileErr
	}

	return nil
}

func MaxOfInt(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func FormatBytes(inputBytes int64) string {
	output := float64(inputBytes)
	unit := ""

	switch {
	case inputBytes >= client.TiB:
		output = output / client.TiB
		unit = "Ti"
	case inputBytes >= client.GiB:
		output = output / client.GiB
		unit = "Gi"
	case inputBytes >= client.MiB:
		output = output / client.MiB
		unit = "Mi"
	case inputBytes >= client.KiB:
		output = output / client.KiB
		unit = "Ki"
	case inputBytes == 0:
		return "0"
	}

	result := strconv.FormatFloat(output, 'f', 1, 64)
	result = strings.TrimSuffix(result, ".0")
	return result + unit
}

func ValidateFsType(logger *zap.SugaredLogger, fsType string) string {
	defaultFsType := "ext4"
	if fsType == "ext4" || fsType == "ext3" || fsType == "xfs" {
		return fsType
	} else if fsType != "" {
		//TODO: Remove this code when we support other than ext4 || ext3.
		logger.With("fsType", fsType).Warn("Supporting only 'ext4/ext3/xfs' as fsType.")
		return defaultFsType
	} else {
		//No fsType provided returning ext4
		return defaultFsType
	}
}

type VolumeLocks struct {
	locks sets.String
	mux   sync.Mutex
}

func NewVolumeLocks() *VolumeLocks {
	return &VolumeLocks{
		locks: sets.NewString(),
	}
}

func (vl *VolumeLocks) TryAcquire(volumeID string) bool {
	vl.mux.Lock()
	defer vl.mux.Unlock()
	if vl.locks.Has(volumeID) {
		return false
	}
	vl.locks.Insert(volumeID)
	return true
}

func (vl *VolumeLocks) Release(volumeID string) {
	vl.mux.Lock()
	defer vl.mux.Unlock()
	vl.locks.Delete(volumeID)
}

// extractStorage extracts the storage size in bytes from the given capacity
// range. If the capacity range is not satisfied it returns the default volume
// size. If the capacity range is below or above supported sizes, it returns an
// error.
func ExtractStorage(capRange *csi.CapacityRange) (int64, error) {
	if capRange == nil {
		return defaultVolumeSizeInBytes, nil
	}

	requiredBytes := capRange.GetRequiredBytes()
	requiredSet := 0 < requiredBytes
	limitBytes := capRange.GetLimitBytes()
	limitSet := 0 < limitBytes

	if !requiredSet && !limitSet {
		return defaultVolumeSizeInBytes, nil
	}

	if requiredSet && limitSet && limitBytes < requiredBytes {
		return 0, fmt.Errorf("limit (%v) can not be less than required (%v) size", FormatBytes(limitBytes), FormatBytes(requiredBytes))
	}

	if requiredSet && !limitSet {
		return MaxOfInt(requiredBytes, MinimumVolumeSizeInBytes), nil
	}

	if limitSet {
		return MaxOfInt(limitBytes, MinimumVolumeSizeInBytes), nil
	}

	if requiredSet && requiredBytes > MaximumVolumeSizeInBytes {
		return 0, fmt.Errorf("required (%v) can not exceed maximum supported volume size (%v)", FormatBytes(requiredBytes), FormatBytes(MaximumVolumeSizeInBytes))
	}

	if !requiredSet && limitSet && limitBytes > MaximumVolumeSizeInBytes {
		return 0, fmt.Errorf("limit (%v) can not exceed maximum supported volume size (%v)", FormatBytes(limitBytes), FormatBytes(MaximumVolumeSizeInBytes))
	}

	if requiredSet && limitSet {
		return MaxOfInt(requiredBytes, limitBytes), nil
	}

	if requiredSet {
		return requiredBytes, nil
	}

	if limitSet {
		return limitBytes, nil
	}

	return defaultVolumeSizeInBytes, nil
}

func RoundUpSize(volumeSizeBytes int64, allocationUnitBytes int64) int64 {
	return (volumeSizeBytes + allocationUnitBytes - 1) / allocationUnitBytes
}

func RoundUpMinSize() int64 {
	return RoundUpSize(MinimumVolumeSizeInBytes, 1*client.GiB)
}

func IsFipsEnabled() (string, error) {
	command := exec.Command(CAT_COMMAND, FIPS_ENABLED_FILE_PATH)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %v\narguments: %s\nOutput: %v\n", err, CAT_COMMAND, string(output))
	}

	return string(output), nil
}
func IsInTransitEncryptionPackageInstalled() (bool, error) {
	args := []string{"-q", InTransitEncryptionPackageName}
	command := exec.Command(RPM_COMMAND, args...)
	output, err := command.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("command failed: %v\narguments: %s\nOutput: %v\n", err, RPM_COMMAND, string(output))
	}

	if len(output) > 0 {
		rpmSearchOutput := string(output)
		if strings.Contains(rpmSearchOutput, InTransitEncryptionPackageName) && !strings.Contains(rpmSearchOutput, "not installed") {
			return true, nil
		}
		return false, nil
	}
	return false, nil
}

func GetBlockSizeBytes(logger *zap.SugaredLogger, devicePath string) (int64, error) {
	args := []string{"--getsize64", devicePath}
	cmd := exec.Command("blockdev", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return -1, fmt.Errorf("command failed: %v\narguments: %s\nOutput: %v\n", err, "blockdev", string(output))
	}
	strOut := strings.TrimSpace(string(output))
	logger.With("devicePath", devicePath, "command", "blockdev", "output", strOut).Debugf("Get block device size in bytes successful")
	gotSizeBytes, err := strconv.ParseInt(strOut, 10, 64)
	if err != nil {
		return -1, fmt.Errorf("failed to parse size %s into an int64 size", strOut)
	}
	return gotSizeBytes, nil
}

func ValidateDNSName(name string) bool {
	pattern := `^([a-zA-Z0-9]+(-[a-zA-Z0-9]+)*\.)+[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(pattern, name)
	return match
}

func ValidateFssId(id string) *FSSVolumeHandler {
	volumeHandler := &FSSVolumeHandler{"", "", ""}
	if id == "" {
		return volumeHandler
	}
	//As ipv6 mount target address contains colons we need to find index of first and last colon to get the three parts
	firstColon := strings.Index(id, ":")
	lastColon := strings.LastIndex(id, ":")
	if firstColon > 0 && lastColon < len(id)-1 && firstColon != lastColon {
		//To handle ipv6  ex.[fd00:00c1::a9fe:202] trim brackets to get fd00:00c1::a9fe:202 which is parsable
		if net.ParseIP(strings.Trim(id[firstColon+1:lastColon], "[]")) != nil || ValidateDNSName(id[firstColon+1:lastColon]) {
			volumeHandler.FilesystemOcid = id[:firstColon]
			volumeHandler.MountTargetIPAddress = id[firstColon+1 : lastColon]
			volumeHandler.FsExportPath = id[lastColon+1:]
			return volumeHandler
		}
	}
	return volumeHandler
}

func GetIsFeatureEnabledFromEnv(logger *zap.SugaredLogger, featureName string, defaultValue bool) bool {
	enableFeature := defaultValue
	enableFeatureEnvVar, ok := os.LookupEnv(featureName)
	if ok {
		var err error
		enableFeature, err = strconv.ParseBool(enableFeatureEnvVar)
		if err != nil {
			logger.With(zap.Error(err)).Errorf("failed to parse %s envvar, defaulting to %t", featureName, defaultValue)
			return defaultValue
		}
	}
	return enableFeature
}

func ConvertIscsiIpFromIpv4ToIpv6(ipv4IscsiIp string) (string, error) {
	ipv4IscsiIP := net.ParseIP(ipv4IscsiIp).To4()
	if ipv4IscsiIP == nil {
		return "", fmt.Errorf("invalid iSCSIIp identified %s", ipv4IscsiIp)
	}
	ipv6IscsiIp := net.ParseIP(IscsiIpv6Prefix)
	ipv6IscsiIpBytes := ipv6IscsiIp.To16()
	copy(ipv6IscsiIpBytes[12:], ipv4IscsiIP.To4())
	return ipv6IscsiIpBytes.String(), nil
}

func FormatValidIp(ipAddress string) string {
	if net.ParseIP(ipAddress).To4() != nil {
		return ipAddress
	} else if net.ParseIP(ipAddress).To16() != nil {
		return fmt.Sprintf("[%s]", strings.Trim(ipAddress, "[]"))
	}
	return ipAddress
}

func FormatValidIpStackInK8SConvention(ipStack string) string {
	if strings.EqualFold(ipStack, Ipv4Stack) {
		return Ipv4Stack
	} else if strings.EqualFold(ipStack, Ipv6Stack) {
		return Ipv6Stack
	}
	return ipStack
}

func IsIpv4(ipAddress string) bool {
	return net.ParseIP(ipAddress).To4() != nil
}

func IsIpv6(ipAddress string) bool {
	return net.ParseIP(ipAddress).To4() == nil && net.ParseIP(strings.Trim(ipAddress, "[]")).To16() != nil
}

func IsIpv4SingleStackSubnet(subnet *core.Subnet) bool {
	return !IsDualStackSubnet(subnet) && subnet.CidrBlock != nil && len(*subnet.CidrBlock) > 0 && !strings.Contains(*subnet.CidrBlock, "null")
}

func IsIpv6SingleStackSubnet(subnet *core.Subnet) bool {
	return !IsDualStackSubnet(subnet) && !IsIpv4SingleStackSubnet(subnet)
}

func IsDualStackSubnet(subnet *core.Subnet) bool {
	return subnet.CidrBlock != nil && len(*subnet.CidrBlock) > 0 && !strings.Contains(*subnet.CidrBlock, "null") &&
		((subnet.Ipv6CidrBlock != nil && len(*subnet.Ipv6CidrBlock) > 0) || len(subnet.Ipv6CidrBlocks) > 0)
}

func IsValidIpFamilyPresentInClusterIpFamily(clusterIpFamily string) bool {
	return len(clusterIpFamily) > 0 && (strings.Contains(clusterIpFamily, Ipv4Stack) || strings.Contains(clusterIpFamily, Ipv6Stack))
}

func IsIpv6SingleStackNode(nodeMetadata *NodeMetadata) bool {
	if nodeMetadata == nil {
		return false
	}
	return nodeMetadata.Ipv6Enabled == true && nodeMetadata.Ipv4Enabled == false
}

func LoadCSIConfigFromConfigMap(csiConfig *CSIConfig, k kubernetes.Interface, configMapName string, logger *zap.SugaredLogger) {
	// Get the ConfigMap
	// Parse the configuration for each driver
	cm, err := k.CoreV1().ConfigMaps("kube-system").Get(context.Background(), configMapName, metav1.GetOptions{})
	if err != nil {
		logger.Debugf("Failed to load ConfigMap %v due to error %v. Using default configuration.", configMapName, err)
		return
	}

	if lustreConfig, exists := cm.Data["lustre"]; exists {
		if err := yaml.Unmarshal([]byte(lustreConfig), &csiConfig.Lustre); err != nil {
			logger.Debugf("Failed to parse lustre key in config map %v. Error: %v",configMapName,  err)
			return
		}
		logger.Infof("Successfully loaded ConfigMap %v. Using customized configuration for csi driver.", configMapName)
	}
}
