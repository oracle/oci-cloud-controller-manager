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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"go.uber.org/zap"
	"k8s.io/kubernetes/pkg/volume/util/hostutil"
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
	ISCSIPORT = "iscsi_port"
)

// ErrMountPointNotFound is returned when a given path does not appear to be
// a mount point.
var ErrMountPointNotFound = errors.New("mount point not found")

// diskByPathPattern is the regex for extracting the iSCSI connection details
// from /dev/disk/by-path/<disk>.
var diskByPathPattern = regexp.MustCompile(
	`/dev/disk/by-path/ip-(?P<IPv4>[\w\.]+):(?P<Port>\d+)-iscsi-(?P<IQN>[\w\.\-:]+)-lun-1`,
)

// Interface mounts iSCSI voumes.
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

	Resize(devicePath string, volumePath string) (bool, error)
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

//Disk interface
type Disk struct {
	IQN  string
	IPv4 string
	Port int
}

func (sd *Disk) String() string {
	return fmt.Sprintf("%s:%d-%s", sd.IPv4, sd.Port, sd.IQN)
}

// Target returns the target to connect to in the format ip:port.
func (sd *Disk) Target() string {
	return fmt.Sprintf("%s:%d", sd.IPv4, sd.Port)
}

func newWithMounter(logger *zap.SugaredLogger, mounter mount.Interface, iqn, ipv4 string, port int) Interface {
	return &iSCSIMounter{
		disk: &Disk{
			IQN:  iqn,
			IPv4: ipv4,
			Port: port,
		},
		runner:  exec.New(),
		mounter: mounter,
		logger:  logger,
	}
}

// New creates a new iSCSI handler.
func New(logger *zap.SugaredLogger, iqn, ipv4 string, port int) Interface {
	return newWithMounter(logger, mount.New(mountCommand), iqn, ipv4, port)
}

//NewFromISCSIDisk creates a new iSCSI handler from ISCSIDisk.
func NewFromISCSIDisk(logger *zap.SugaredLogger, sd *Disk) Interface {
	return &iSCSIMounter{
		disk: sd,

		runner:  exec.New(),
		mounter: mount.New(mountCommand),
		logger:  logger,
	}
}

// NewFromDevicePath extracts the IQN, IPv4 address, and port from a
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

// FindFromDevicePath extracts the IQN, IPv4 address, and port from a
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
	diskByPaths, err := diskByPathsForMountPoint(mountPoint)
	if err != nil {
		return nil, err
	}
	logger.Infof("diskByPaths is %v", diskByPaths)
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
	return deviceOpened(pathname, c.logger)
}

func deviceOpened(pathname string, logger *zap.SugaredLogger) (bool, error) {
	hostUtil := hostutil.NewHostUtil()
	exists, err := hostUtil.PathExists(pathname)
	if err != nil {
		logger.With(zap.Error(err)).Errorf("Failed to find is path exists %s", pathname)
		return false, err
	}
	if !exists {
		logger.Infof("Path does not exist %s", pathname)
		return false, nil
	}
	return hostUtil.DeviceOpened(pathname)
}

func (c *iSCSIMounter) UnmountPath(path string) error {
	return UnmountPath(c.logger, path, c.mounter)
}

func (c *iSCSIMounter) Resize(devicePath string, volumePath string) (bool, error) {
	resizefs := mount.NewResizeFs(c.runner)
	return resizefs.Resize(devicePath, volumePath)
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
