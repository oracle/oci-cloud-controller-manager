package csi_util

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"github.com/container-storage-interface/spec/lib/go/csi"
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
)

//Util interface
type Util struct {
	Logger *zap.SugaredLogger
}

var (
	DiskByPathPatternPV    = `/dev/disk/by-path/pci-\d+:\d+:\d+\.\d+-scsi-\d+:\d+:\d+:\d+$`
	DiskByPathPatternISCSI = `/dev/disk/by-path/ip-[\w\.]+:\d+-iscsi-[\w\.\-:]+-lun-1$`
)

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

func (u *Util) LookupNodeAvailableDomain(k kubernetes.Interface, nodeID string) (string, error) {
	n, err := k.CoreV1().Nodes().Get(context.Background(), nodeID, metav1.GetOptions{})
	if err != nil {
		u.Logger.With(zap.Error(err)).With("nodeId", nodeID).Error("Failed to get Node by name.")
		return "", fmt.Errorf("failed to get node %s", nodeID)
	}
	if n.Labels != nil {
		ad, ok := n.Labels[kubeAPI.LabelZoneFailureDomain]
		if ok {
			return ad, nil
		}
	}

	errMsg := fmt.Sprint("Did not find the label for the fault domain.")
	u.Logger.With("nodeId", nodeID, "label", kubeAPI.LabelZoneFailureDomain).Error(errMsg)
	return "", fmt.Errorf(errMsg)
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

func GetDevicePath(sd *disk.Disk) string {
	return fmt.Sprintf("/dev/disk/by-path/ip-%s:%d-iscsi-%s-lun-1", sd.IPv4, sd.Port, sd.IQN)
}

func ExtractISCSIInformation(attributes map[string]string) (*disk.Disk, error) {
	iqn, ok := attributes[disk.ISCSIIQN]
	if !ok {
		return nil, fmt.Errorf("Unable to get the IQN from the attribute list")
	}
	ipv4, ok := attributes[disk.ISCSIIP]
	if !ok {
		return nil, fmt.Errorf("Unable to get the ipv4 from the attribute list")
	}
	port, ok := attributes[disk.ISCSIPORT]
	if !ok {
		return nil, fmt.Errorf("Unable to get the port from the attribute list")
	}

	nPort, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("Invalid port number: %s, error: %v", port, err)
	}

	return &disk.Disk{
		IQN:  iqn,
		IPv4: ipv4,
		Port: nPort,
	}, nil
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

	logger.With("IQN", m[3], "IPv4", m[1], "Port", port).Info("Found ISCSIInfo for the mount path: ", diskPath)
	return &disk.Disk{
		IQN:  m[3],
		IPv4: m[1],
		Port: port,
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
	if fsType == "ext4" || fsType == "ext3" {
		return fsType
	} else if fsType != "" {
		//TODO: Remove this code when we support other than ext4 || ext3.
		logger.With("fsType", fsType).Warn("Supporting only 'ext4' as fsType.")
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
