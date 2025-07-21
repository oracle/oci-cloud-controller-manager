// Copyright 2017 Oracle and/or its affiliates. All rights reserved.
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

package disk

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	cmdexec "os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/oracle/oci-go-sdk/v65/core"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/mount-utils"
	"k8s.io/utils/exec"
)

const (
	iscsiadmCommand = "iscsiadm"
	mountCommand    = "/bin/mount"

	// ISCSIDEVICE is the map key to get or save iscci device
	ISCSIDEVICE = "iscsi_device"
	// ISCSIIQN is the map key to get or save iSCSI IQN
	ISCSIIQN = "iscci_iqn"
	// ISCSIIP is the map key to get or save iSCSI IP
	ISCSIIP = "iscsi_ip"
	// ISCSIPORT is the map key to get or save iSCSI Port
	ISCSIPORT         = "iscsi_port"
	loginPollInterval = 5 * time.Second
	pathPollInterval  = 2 * time.Second
	pathPollTimeout   = 3 * time.Minute
	waitForPathDelay  = 1 * time.Second

	LIST_PATHS_COMMAND  = "ls -f /dev/disk/by-path"
	DISK_BY_PATH_FOLDER = "/dev/disk/by-path/"
)

// ErrMountPointNotFound is returned when a given path does not appear to be
// a mount point.
var ErrMountPointNotFound = errors.New("mount point not found")

// diskByPathPattern is the regex for extracting the iSCSI connection details
// from /dev/disk/by-path/<disk>.
var diskByPathPattern = regexp.MustCompile(
	`/dev/disk/by-path/ip-(?P<IscsiIp>[[?\w\.\:]+]?):(?P<Port>\d+)-iscsi-(?P<IQN>[\w\.\-:]+)-lun-\d+`,
)

// Interface mounts iSCSI volumes.
type Interface interface {
	// AddToDB adds the iSCSI node record for the target.
	AddToDB() error
	// FormatAndMount formats the given disk, if needed, and mounts it.  That is
	// if the disk is not formatted and it is not being mounted as read-only it
	// will format it first then mount it. Otherwise, if the disk is already
	// formatted or it is being mounted as read-only, it will be mounted without
	// formatting.
	FormatAndMount(source string, target string, fstype string, options []string) error

	//Mount only mounts the disk. In case if formatting is handled by different functionality.
	// This function doesn't bother for checking the format again.
	Mount(source string, target string, fstype string, options []string) error

	// Login logs into the iSCSI target.
	Login() error

	// Logout logs out the iSCSI target.
	Logout() error

	DeviceOpened(pathname string) (bool, error)

	IsMounted(devicePath string, targetPath string)	(bool, error)

	// updates the queue depth for iSCSI target
	UpdateQueueDepth() error

	// RemoveFromDB removes the iSCSI target from the database.
	RemoveFromDB() error

	// SetAutomaticLogin sets the iSCSI node to automatically login at machine
	// start-up.
	SetAutomaticLogin() error

	// UnmountPath is a common unmount routine that unmounts the given path and
	// deletes the remaining directory if successful.
	UnmountPath(path string) error

	Rescan(devicePath string) error

	Resize(devicePath string, volumePath string) (bool, error)

	WaitForVolumeLoginOrTimeout(ctx context.Context, multipathDevices []core.MultipathDevice) error

	GetDiskFormat(devicePath string) (string, error)

	WaitForPathToExist(path string, maxRetries int) bool

	ISCSILogoutOnFailure() error
}

// iSCSIMounter implements Interface.
type iSCSIMounter struct {
	disk *Disk

	runner  exec.Interface
	mounter mount.Interface

	// iscsiadmPath is the cached absolute path to iscsiadm.
	iscsiadmPath string
	logger       *zap.SugaredLogger
}

// Disk interface
type Disk struct {
	IQN     string
	IscsiIp string
	Port    int
}

func (sd *Disk) String() string {
	return fmt.Sprintf("%s:%d-%s", sd.IscsiIp, sd.Port, sd.IQN)
}

// Target returns the target to connect to in the format ip:port for Ipv4 and [ip]:port for ipv6.
func (sd *Disk) Target() string {
	if net.ParseIP(sd.IscsiIp).To4() != nil {
		return fmt.Sprintf("%s:%d", sd.IscsiIp, sd.Port)
	}
	return fmt.Sprintf("[%s]:%d", sd.IscsiIp, sd.Port)
}

func newWithMounter(logger *zap.SugaredLogger, mounter mount.Interface, iqn, iSCSIIp string, port int) Interface {
	return &iSCSIMounter{
		disk: &Disk{
			IQN:     iqn,
			IscsiIp: iSCSIIp,
			Port:    port,
		},
		runner:  exec.New(),
		mounter: mounter,
		logger:  logger,
	}
}

// New creates a new iSCSI handler.
func New(logger *zap.SugaredLogger, iqn, iSCSIIp string, port int) Interface {
	return newWithMounter(logger, mount.New(mountCommand), iqn, iSCSIIp, port)
}

// NewFromISCSIDisk creates a new iSCSI handler from ISCSIDisk.
func NewFromISCSIDisk(logger *zap.SugaredLogger, sd *Disk) Interface {
	return &iSCSIMounter{
		disk: sd,

		runner:  exec.New(),
		mounter: mount.New(mountCommand),
		logger:  logger,
	}
}

// NewFromDevicePath extracts the IQN, IscsiIp address, and port from a
// iSCSI mount device path.
// i.e. /dev/disk/by-path/ip-<ip>:<port>-iscsi-<IQN>-lun-1
func NewFromDevicePath(logger *zap.SugaredLogger, mountDevice string) (Interface, error) {
	m := diskByPathPattern.FindStringSubmatch(mountDevice)
	if len(m) != 4 {
		return nil, fmt.Errorf("mount device path %q did not match pattern; got %v", mountDevice, m)
	}

	port, err := strconv.Atoi(m[2])
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}

	return New(logger, m[3], m[1], port), nil
}

// FindFromDevicePath extracts the IQN, IscsiIp address, and port from a
// iSCSI mount device path.
// i.e. /dev/disk/by-path/ip-<ip>:<port>-iscsi-<IQN>-lun-1
func FindFromDevicePath(logger *zap.SugaredLogger, mountDevice string) ([]string, error) {
	m := diskByPathPattern.FindStringSubmatch(mountDevice)
	if len(m) != 4 {
		return nil, fmt.Errorf("mount device path %q did not match pattern; got %v", mountDevice, m)
	}
	return m, nil
}

// NewFromMountPointPath gets /dev/disk/by-path/ip-<ip>:<port>-iscsi-<IQN>-lun-1
// from the given mount point path.
func NewFromMountPointPath(logger *zap.SugaredLogger, mountPath string) (Interface, error) {
	mounter := mount.New(mountCommand)
	mountPoint, err := getMountPointForPath(mounter, mountPath)
	if err != nil {
		return nil, err
	}
	diskByPaths, err := diskByPathsForMountPoint(mountPoint)
	if err != nil {
		return nil, err
	}
	for _, diskByPath := range diskByPaths {
		iface, err := NewFromDevicePath(logger, diskByPath)
		if err == nil {
			return iface, nil
		}
	}
	return nil, errors.New("iSCSI information not found for mount point")
}

// FindFromMountPointPath gets /dev/disk/by-path/ip-<ip>:<port>-iscsi-<IQN>-lun-1
// from the given mount point path.
func FindFromMountPointPath(logger *zap.SugaredLogger, diskByPaths []string) ([]string, error) {

	for _, diskByPath := range diskByPaths {
		m, err := FindFromDevicePath(logger, diskByPath)
		if err == nil {
			return m, nil
		}
	}
	return nil, errors.New("iSCSI information not found for mount point")
}

// GetDiskPathFromMountPath resolves a directory to a block device
func GetDiskPathFromMountPath(logger *zap.SugaredLogger, mountPath string) ([]string, error) {
	mounter := mount.New(mountCommand)
	mountPoint, err := getMountPointForPath(mounter, mountPath)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(mountPoint.Device, "/dev/mapper") {
		return []string{mountPoint.Device}, nil
	}
	diskByPaths, err := diskByPathsForMountPoint(mountPoint)
	if err != nil {
		return nil, err
	}
	logger.Infof("diskByPaths is %v", diskByPaths)
	return diskByPaths, nil
}

// Looping through sanitizedDevice - "sanitizedDevice": "/sdc"
// Finding device name - "deviceName": "/sdc"
// Finding disk by path - "diskByPaths": ["/dev/disk/by-path/ip-<ip>-iscsi-iqn.2015-12.com.oracleiaas:uniqfier-lun-2"]

// Gets the diskPath for a bind-mounted device file
func GetDiskPathFromBindDeviceFilePath(logger *zap.SugaredLogger, mountPath string) ([]string, error) {
	// Get the block device for the given mount path
	devices, err := FindMount(mountPath)

	if err != nil {
		logger.With(zap.Error(err)).Warnf("Unable to get block device for mount path: %s", mountPath)
		return nil, err
	}

	var sanitizedDevices []string
	for _, dev := range devices {
		if prefixEnd := strings.Index(dev, "["); prefixEnd != -1 {
			sanitizedDevice := dev[prefixEnd+1:] // Start after `[`
			sanitizedDevice = strings.TrimSuffix(sanitizedDevice, "]")
			sanitizedDevice = filepath.Clean(sanitizedDevice) // Fix extra slashes
			sanitizedDevices = append(sanitizedDevices, sanitizedDevice)
		}
	}

	if len(sanitizedDevices) != 1 {
		logger.Warn("Found multiple or no block devices for the mount path")
		return nil, fmt.Errorf("did not find exactly a single block device on %s, found devices: %v", mountPath, sanitizedDevices)
	}

	deviceName := sanitizedDevices[0]

	// Convert the device name to the correct path format
	devicePath := filepath.Join("/dev", deviceName)

	// Create a mount.MountPoint struct
	mountPoint := mount.MountPoint{
		Path:   mountPath,
		Device: devicePath,
	}

	// Use the device path to get diskByPaths
	diskByPaths, err := diskByPathsForMountPoint(mountPoint)
	if err != nil {
		logger.With(zap.Error(err)).Warn("Unable to find diskByPaths for device")
		return nil, err
	}

	return diskByPaths, nil
}

// getISCSIAdmPath gets the absolute path to the iscsiadm executable on the
// $PATH.
func (c *iSCSIMounter) getISCSIAdmPath() (string, error) {
	if c.iscsiadmPath != "" {
		return c.iscsiadmPath, nil
	}

	path, err := c.runner.LookPath(iscsiadmCommand)
	if err != nil {
		return "", err
	}
	c.iscsiadmPath = path
	c.logger.With("iscsiadm", c.iscsiadmPath).Info("Full iscsiadm path found.")
	return path, nil
}

func (c *iSCSIMounter) iscsiadm(parts ...string) (string, error) {
	iscsiadmPath, err := c.getISCSIAdmPath()
	if err != nil {
		return "", err
	}

	cmd := c.runner.Command(iscsiadmPath, parts...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (c *iSCSIMounter) AddToDB() error {
	c.logger.With("IQN", c.disk.IQN, "target", c.disk.Target()).Info("Adding node record to db.")

	_, err := c.iscsiadm(
		"-m", "node",
		"-o", "new",
		"-T", c.disk.IQN,
		"-p", c.disk.Target())
	if err != nil {
		return fmt.Errorf("iscsi: error adding node record to db: %v", err)
	}

	c.logger.With("IQN", c.disk.IQN, "target", c.disk.Target()).Info("Added node record to db.")

	return nil
}

func (c *iSCSIMounter) SetAutomaticLogin() error {
	c.logger.With("IQN", c.disk.IQN).Info("Configuring automatic node login.")

	_, err := c.iscsiadm(
		"-m", "node",
		"-o", "update",
		"-T", c.disk.IQN,
		"-n", "node.startup",
		"-v", "automatic")
	if err != nil {
		return fmt.Errorf("iscsi: error configuring automatic node login: %v", err)
	}

	c.logger.With("IQN", c.disk.IQN).Info("Configured automatic node login.")

	return nil
}

func (c *iSCSIMounter) Login() error {
	c.logger.With("IQN", c.disk.IQN, "target", c.disk.Target()).Info("Logging in.")

	_, err := c.iscsiadm(
		"-m", "node",
		"-T", c.disk.IQN,
		"-p", c.disk.Target(),
		"-l")
	if err != nil {
		return fmt.Errorf("iscsi: error logging in target: %v", err)
	}

	c.logger.With("IQN", c.disk.IQN, "target", c.disk.Target()).Info("Logged in.")

	return nil
}

// Logout logs out the iSCSI target.
// sudo iscsiadm -m node -T <IQN> -p <ip>:<port>  -u
func (c *iSCSIMounter) Logout() error {
	c.logger.With("IQN", c.disk.IQN, "target", c.disk.Target()).Info("Logging out.")
	_, err := c.iscsiadm(
		"-m", "node",
		"-T", c.disk.IQN,
		"-p", c.disk.Target(),
		"-u")
	if err != nil {
		return fmt.Errorf("iscsi: error logging out target: %v", err)
	}

	c.logger.With("IQN", c.disk.IQN, "target", c.disk.Target()).Info("Logged out.")

	return nil
}

func (c *iSCSIMounter) UpdateQueueDepth() error {
	c.logger.With("IQN", c.disk.IQN, "target", c.disk.Target()).Info("Updating queue depth to 128.")

	_, err := c.iscsiadm(
		"-m", "node",
		"-T", c.disk.IQN,
		"-p", c.disk.Target(),
		"-o", "update",
		"-n", "node.session.queue_depth",
		"-v", "128")
	if err != nil {
		return fmt.Errorf("iscsi: error updating queue depth in target: %v", err)
	}

	c.logger.With("IQN", c.disk.IQN, "target", c.disk.Target()).Info("Updated queue depth.")

	return nil
}

func (c *iSCSIMounter) RemoveFromDB() error {
	c.logger.With("IQN", c.disk.IQN, "target", c.disk.Target()).Info("Removing from database.")
	_, err := c.iscsiadm(
		"-m", "node",
		"-o", "delete",
		"-T", c.disk.IQN,
		"-p", c.disk.Target())
	if err != nil {
		return fmt.Errorf("iscsi: error removing target from database: %v", err)
	}

	c.logger.With("IQN", c.disk.IQN, "target", c.disk.Target()).Info("Removed from database.")

	return nil
}

func (c *iSCSIMounter) WaitForVolumeLoginOrTimeout(ctx context.Context, multipathDevices []core.MultipathDevice) error {
	c.logger.Info("Attachment type ISCSI. WaitForVolumeLoginOrTimeout() not needed for iscsi attachment")
	return nil
}

func (c *iSCSIMounter) FormatAndMount(source string, target string, fstype string, options []string) error {
	safeMounter := &mount.SafeFormatAndMount{
		Interface: c.mounter,
		Exec:      c.runner,
	}
	return formatAndMount(source, target, fstype, options, safeMounter)
}

func formatAndMount(source string, target string, fstype string, options []string, sm *mount.SafeFormatAndMount) error {
	return sm.FormatAndMount(source, target, fstype, options)
}

func (c *iSCSIMounter) GetDiskFormat(disk string) (string, error) {
	return getDiskFormat(c.runner, disk, c.logger)
}

func (c *iSCSIMounter) Mount(source string, target string, fstype string, options []string) error {
	safeMounter := &mount.SafeFormatAndMount{
		Interface: c.mounter,
		Exec:      c.runner,
	}
	return mnt(source, target, fstype, options, safeMounter)
}

func mnt(source string, target string, fstype string, options []string, sm *mount.SafeFormatAndMount) error {
	return sm.Mount(source, target, fstype, options)
}

func (c *iSCSIMounter) DeviceOpened(pathname string) (bool, error) {
	var err error
	pathname, err = GetIscsiDevicePath(c.disk)
	if err != nil {
		if strings.Contains(err.Error(), "No such file or directory") {
			return false, nil
		} else {
			return false, err
		}
	}
	return deviceOpened(pathname, c.logger)
}

func (c *iSCSIMounter) IsMounted(devicePath string, targetPath string) (bool, error) {
	var diskByPath string
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

	diskByPath, err = GetIscsiDevicePath(c.disk)
	if err != nil {
		if strings.Contains(err.Error(), "No such file or directory") {
			return false, fmt.Errorf("ISCSI login not complete for volume but staging path is a mount point, mapped to wrong device")
		} else {
			return false, fmt.Errorf("failed to find /dev/disk/by-path path for volume: %v", c.disk)
		}
	}

	resolvedDevicePath, err := filepath.EvalSymlinks(diskByPath)
	if err != nil {
		return false, fmt.Errorf("failed to resolve symlink for /dev/disk/by-path path %s: %v", diskByPath, err)
	}

	mounts, err := c.mounter.List()
	if err != nil {
		return false, fmt.Errorf("could not list mount points: %v", err)
	}

	for _, m := range mounts {
		if m.Path == targetPath {
			if m.Device == resolvedDevicePath {
				return true, nil
			}
			return false, fmt.Errorf("expected device %s but found %s mounted at %s", resolvedDevicePath, m.Device, targetPath)
		}
	}
	return false, nil
}

func (c *iSCSIMounter) UnmountPath(path string) error {
	return UnmountPath(c.logger, path, c.mounter)
}

func (c *iSCSIMounter) Rescan(devicePath string) error {
	return Rescan(c.logger, devicePath)
}

func (c *iSCSIMounter) Resize(devicePath string, volumePath string) (bool, error) {
	resizefs := mount.NewResizeFs(c.runner)
	return resizefs.Resize(devicePath, volumePath)
}

func (c *iSCSIMounter) WaitForPathToExist(path string, maxRetries int) bool {
	return true
}

func (c *iSCSIMounter) ISCSILogoutOnFailure() error {
	err := c.Logout()
	if err != nil {
		c.logger.With(zap.Error(err)).Error("failed to logout from the iSCSI target")
		return status.Error(codes.Internal, err.Error())
	}

	err = c.RemoveFromDB()
	if err != nil {
		c.logger.With(zap.Error(err)).Error("failed to remove the iSCSI node record")
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

// getMountPointForPath returns the mount.MountPoint for a given path. If the
// given path is not a mount point
func getMountPointForPath(ml mount.Interface, path string) (mount.MountPoint, error) {
	mountPoints, err := ml.List()
	if err != nil {
		return mount.MountPoint{}, err
	}

	for _, mountPoint := range mountPoints {
		if mountPoint.Path == path {
			return mountPoint, nil
		}
	}

	return mount.MountPoint{}, ErrMountPointNotFound
}

// TODO(apryde): Need to think about how best to test this/make it more
// testable.
func diskByPathsForMountPoint(mountPoint mount.MountPoint) ([]string, error) {
	diskByPaths := []string{}
	err := filepath.Walk("/dev/disk/by-path/", func(path string, info os.FileInfo, err error) error {
		target, err := filepath.EvalSymlinks(path)
		if err != nil {
			return err
		}
		if target == mountPoint.Device {
			diskByPaths = append(diskByPaths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(diskByPaths) == 0 {
		return nil, errors.New("disk by path link not found")
	}
	return diskByPaths, nil
}

func GetIscsiDevicePath(disk *Disk) (string, error) {
	// run command ls -l /dev/disk/by-path
	cmdStr := fmt.Sprintf("ls -f /dev/disk/by-path/ip-%s:%d-iscsi-%s-lun-*", disk.IscsiIp, disk.Port, disk.IQN)
	cmd := cmdexec.Command("bash", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %v\ncommand: %s\nOutput: %s\n", err, LIST_PATHS_COMMAND, string(output))
	}
	for _, line := range strings.Split(string(output), "\n") {
		re := regexp.MustCompile(fmt.Sprintf(`(ip-%s:%d-iscsi-%s-lun-\d+)`, disk.IscsiIp, disk.Port, disk.IQN))
		match := re.FindStringSubmatch(line)
		if len(match) > 0 {
			fileName := match[1]
			devicePath := DISK_BY_PATH_FOLDER + fileName
			return devicePath, nil
		}
	}
	return "", fmt.Errorf("cannot find device path")
}

func WaitForDevicePathToExist(ctx context.Context, disk *Disk, logger *zap.SugaredLogger) (string, error) {
	logger.With("disk", disk).Info("Waiting for iscsi device path to exist")

	ctxt, cancel := context.WithTimeout(ctx, pathPollTimeout)
	defer cancel()

	var iscsiDevicePath string

	if err := wait.PollImmediateUntil(pathPollInterval, func() (done bool, err error) {
		devicePath, err := GetIscsiDevicePath(disk)
		if err != nil {
			if !strings.Contains(err.Error(), "No such file or directory") {
				return false, err
			}
			return false, nil
		} else {
			iscsiDevicePath = devicePath
			return true, nil
		}
	}, ctxt.Done()); err != nil {
		return "", err
	}

	return iscsiDevicePath, nil
}

func Rescan(logger *zap.SugaredLogger, devicePath string) error {
	lsblkargs := []string{"-n", "-o", "NAME", devicePath}
	lsblkcmd := cmdexec.Command("lsblk", lsblkargs...)
	lsblkoutput, err := lsblkcmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Failed to find device name associated with devicePath %s", devicePath)
	}
	deviceName := strings.TrimSpace(string(lsblkoutput))
	if strings.HasPrefix(deviceName, "/dev/") {
		deviceName = strings.TrimPrefix(deviceName, "/dev/")
	}
	logger.With("deviceName", deviceName).Info("Rescanning")

	// run command dd iflag=direct if=/dev/<device_name> of=/dev/null count=1
	// https://docs.oracle.com/en-us/iaas/Content/Block/Tasks/rescanningdisk.htm#Rescanni
	devicePathFileArg := fmt.Sprintf("if=%s", devicePath)
	args := []string{"iflag=direct", devicePathFileArg, "of=/dev/null", "count=1"}
	cmd := cmdexec.Command("dd", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %v\narguments: %s\nOutput: %v\n", err, "dd", string(output))
	}
	logger.With("command", "dd", "output", string(output)).Debug("dd output")
	// run command echo 1 | tee /sys/class/block/%s/device/rescan
	// https://docs.oracle.com/en-us/iaas/Content/Block/Tasks/rescanningdisk.htm#Rescanni
	cmdStr := fmt.Sprintf("echo 1 | tee /sys/class/block/%s/device/rescan", deviceName)
	cmd = cmdexec.Command("bash", "-c", cmdStr)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %v\narguments: %s\nOutput: %v\n", err, cmdStr, string(output))
	}
	logger.With("command", cmdStr, "output", string(output)).Debug("rescan output")

	return nil
}

func GetScsiInfo(mountDevice string) (*Disk, error) {
	m := diskByPathPattern.FindStringSubmatch(mountDevice)
	if len(m) != 4 {
		return nil, fmt.Errorf("mount device path %q did not match pattern; got %v", mountDevice, m)
	}

	port, err := strconv.Atoi(m[2])
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}

	return &Disk{
		IQN:     m[3],
		IscsiIp: m[1],
		Port:    port,
	}, nil
}
