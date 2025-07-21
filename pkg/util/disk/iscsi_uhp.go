package disk

import (
	"context"
	"fmt"
	"os"
	cmdexec "os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/oracle/oci-go-sdk/v65/core"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/mount-utils"
	"k8s.io/utils/exec"
)

const (
	volumeLoginTimeout  = 3 * time.Minute

	pathPollIntervalUHP  = 10 * time.Second

	CHROOT_BASH_COMMAND = "chroot-bash"
)

// iSCSIUHPMounter implements Interface.
type iSCSIUHPMounter struct {
	runner  exec.Interface
	mounter mount.Interface

	logger *zap.SugaredLogger
}

func NewISCSIUHPMounter(logger *zap.SugaredLogger) Interface {
	return &iSCSIUHPMounter{

		runner:  exec.New(),
		mounter: mount.New(mountCommand),
		logger:  logger,
	}
}

func (c *iSCSIUHPMounter) AddToDB() error {
	c.logger.Info("Attachment type ISCSI for UHP. AddToDB() not needed for UHP ISCSI attachment")
	return nil
}

func (c *iSCSIUHPMounter) FormatAndMount(source string, target string, fstype string, options []string) error {
	safeMounter := &mount.SafeFormatAndMount{
		Interface: c.mounter,
		Exec:      c.runner,
	}
	return formatAndMount(source, target, fstype, options, safeMounter)
}

func (c *iSCSIUHPMounter) Mount(source string, target string, fstype string, options []string) error {
	safeMounter := &mount.SafeFormatAndMount{
		Interface: c.mounter,
		Exec:      c.runner,
	}
	return mnt(source, target, fstype, options, safeMounter)
}

func (c *iSCSIUHPMounter) Login() error {
	c.logger.Info("Attachment type ISCSI for UHP. Login() not needed for UHP ISCSI attachment")
	return nil
}

func (c *iSCSIUHPMounter) Logout() error {
	c.logger.Info("Attachment type ISCSI for UHP. Logout() not needed for UHP ISCSI attachment")
	return nil
}

func (c *iSCSIUHPMounter) UpdateQueueDepth() error {
	c.logger.Info("Attachment type ISCSI for UHP. UpdateQueueDepth() not needed for UHP ISCSI attachment")
	return nil
}

func (c *iSCSIUHPMounter) RemoveFromDB() error {
	c.logger.Info("Attachment type ISCSI for UHP. RemoveFromDB() not needed for UHP ISCSI attachment")
	return nil
}

func (c *iSCSIUHPMounter) SetAutomaticLogin() error {
	c.logger.Info("Attachment type ISCSI for UHP. SetAutomaticLogin() not needed for UHP ISCSI attachment")
	return nil
}

func (c *iSCSIUHPMounter) UnmountPath(path string) error {
	return UnmountPath(c.logger, path, c.mounter)
}

func (c *iSCSIUHPMounter) Rescan(devicePath string) error {
	deviceMapperPath, err := ReadLink(devicePath, c.logger)
	if err != nil {
		c.logger.With(zap.Error(err)).Error("error during getting device mapper path for multipath device, will retry")
		return fmt.Errorf("failed to obtain device mapper path from multipath device path %v: %v", devicePath, err)
	}
	deviceMapperPathBase := path.Base(deviceMapperPath)
	c.logger.With("deviceMapperPathBase", deviceMapperPathBase).Info("Found device mapper base path")

	devicePathsFolder := fmt.Sprintf(`/sys/block/%s/slaves`, deviceMapperPathBase)
	fileNames, err := os.ReadDir(devicePathsFolder)
	if err != nil {
		c.logger.With(zap.Error(err)).Errorf("error getting list of device paths from %s", devicePathsFolder)
		return fmt.Errorf("failed to obtain device paths from device paths folder %v: %v", devicePathsFolder, err)
	}

	var rescanPaths []string
	for _, file := range fileNames {
		if strings.HasPrefix(file.Name(), "sd") {
			subDevicePath := fmt.Sprintf(`/sys/block/%s/device/rescan`, file.Name())
			rescanPaths = append(rescanPaths, subDevicePath)
		}
	}
	for _, rescanPath := range rescanPaths {
		cmdStr := fmt.Sprintf("echo 1 | tee %s", rescanPath)
		cmd := cmdexec.Command("bash", "-c", cmdStr)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("command failed: %v\narguments: %s\nOutput: %v\n", err, cmdStr, string(output))
		}
		c.logger.With("command", cmdStr, "output", string(output)).Debug("rescan output")
	}
	c.logger.With("devicePath", devicePath).Info("Rescanned multipath device successfully")

	c.logger.With("deviceMapperPathBase", deviceMapperPathBase).Info("Resizing multipath device")
	args := []string{"multipathd resize map", deviceMapperPathBase}
	command := cmdexec.Command(CHROOT_BASH_COMMAND, args...)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %v\narguments: %s\nOutput: %v\n", err, CHROOT_BASH_COMMAND, string(output))
	}

	return nil
}

func (c *iSCSIUHPMounter) Resize(devicePath string, volumePath string) (bool, error) {
	resizefs := mount.NewResizeFs(c.runner)
	return resizefs.Resize(devicePath, volumePath)
}

func (c *iSCSIUHPMounter) WaitForPathToExist(path string, maxRetries int) bool {
	return true
}

func (c *iSCSIUHPMounter) GetISCSILoginState(multipathDevices []core.MultipathDevice) (bool, error) {
	c.logger.Info("Getting login state")
	cmdStr := fmt.Sprintf(DISK_BY_PATH_FOLDER)
	cmd := cmdexec.Command("ls", "-f", cmdStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("command failed: %v\ncommand: %s\nOutput: %v\n", err, LIST_PATHS_COMMAND, string(output))
	}
	for _, device := range multipathDevices {
		ip, port, iqn := device.Ipv4, device.Port, device.Iqn
		pathStr := fmt.Sprintf(`ip-%s:%d-iscsi-%s`, *ip, *port, *iqn)
		pathFound := strings.Contains(string(output), pathStr)
		if !pathFound {
			c.logger.With("multipathDevice", device, "listDiskPathsOutput", string(output), "pathStr", pathStr).Info("Error finding path for device in output")
			return false, nil
		}
	}
	return true, nil
}

func (c *iSCSIUHPMounter) WaitForVolumeLoginOrTimeout(ctx context.Context, multipathDevices []core.MultipathDevice) error {
	ctx, cancel := context.WithTimeout(ctx, volumeLoginTimeout)
	defer cancel()

	if err := wait.PollImmediateUntil(loginPollInterval, func() (done bool, err error) {
		loggedIn, err := c.GetISCSILoginState(multipathDevices)
		if err != nil {
			c.logger.With(zap.Error(err)).Error("error during waiting for automatic login through block volume management plugin")
			return false, err
		}
		if loggedIn {
			return true, nil
		}
		return false, nil
	}, ctx.Done()); err != nil {
		c.logger.With(zap.Error(err)).Error("error during waiting for automatic login through block volume management plugin")
		return err
	}
	return nil
}

func (c *iSCSIUHPMounter) GetDiskFormat(devicePath string) (string, error) {
	return getDiskFormat(c.runner, devicePath, c.logger)
}

func (c *iSCSIUHPMounter) DeviceOpened(pathname string) (bool, error) {
	return deviceOpened(pathname, c.logger)
}

func (c *iSCSIUHPMounter) IsMounted(devicePath string, targetPath string) (bool, error) {
	notMnt, err := c.mounter.IsLikelyNotMountPoint(targetPath)
	if err != nil {
		if os.IsNotExist(err){
			return false, nil
		}
		return false, fmt.Errorf("failed to check if %s is a mount point: %v", targetPath, err)
	}
	if notMnt {
		return false, nil
	}
	mounts, err := c.mounter.List()
	if err != nil {
		return false, fmt.Errorf("could not list mount points: %v", err)
	}

	for _, m := range mounts {
		if m.Path == targetPath {
			if m.Device == devicePath {
				return true, nil
			}
			return false, fmt.Errorf("expected device %s but found %s mounted at %s", devicePath, m.Device, targetPath)
		}
	}
	return false, nil
}

func (c *iSCSIUHPMounter) ISCSILogoutOnFailure() error {
	return nil
}

func GetMultipathIscsiDevicePath(ctx context.Context, consistentDevicePath string, logger *zap.SugaredLogger) (string, error) {
	logger.With("consistentDevicePath", consistentDevicePath).Info("Getting friendly name of multipath device using consistent device path")

	ctxt, cancel := context.WithTimeout(ctx, pathPollTimeout)
	defer cancel()

	var friendlyName string

	if err := wait.PollImmediateUntil(pathPollIntervalUHP, func() (done bool, err error) {
		multiDevicePath, err := GetMultipathFriendlyName(consistentDevicePath, logger)
		if err != nil {
			return false, nil
		} else {
			friendlyName = multiDevicePath
			return true, nil
		}
	}, ctxt.Done()); err != nil {
		return "", err
	}

	devicePath := friendlyName
	return devicePath, nil
}

func GetMultipathFriendlyName(consistentDevicePath string, logger *zap.SugaredLogger) (string, error) {
	friendlyName, err := ReadLink(consistentDevicePath, logger)
	if err != nil {
		logger.With(zap.Error(err)).Error("error during getting friendly name for multipath device, will retry")
		return "", fmt.Errorf("failed to get friendly name for volume")
	} else if !strings.HasPrefix(friendlyName, "/dev/mapper/") {
		logger.With(zap.Error(err)).Error("friendly name does not point to multipath device, will retry")
		return "", fmt.Errorf("failed to get multipath device path")
	}
	return friendlyName, nil
}

func ReadLink(symbolicLink string, logger *zap.SugaredLogger) (string, error) {
	files, err := filepath.Glob(symbolicLink)
	if err != nil {
		logger.With("symbolicLink", symbolicLink).Error("Error finding linked path for multipath device consistent path")
		return "", err
	}

	if len(files) == 0 {
		logger.With("symbolicLink", symbolicLink).Error("Error finding linked path for multipath device consistent path")
		errMsg := fmt.Sprintf("The linked path for symbolic link %s is not found", symbolicLink)
		return "", fmt.Errorf(errMsg)
	}

	linkedPath, err := os.Readlink(files[0])
	if err != nil {
		logger.With("symbolicLink", symbolicLink).With(zap.Error(err)).Error("Error finding linked path for multipath device consistent path")
		return "", err
	}

	return linkedPath, nil
}
