// +build linux

/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mount

import (
	"bufio"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"path/filepath"

	"go.uber.org/zap"
	utilexec "k8s.io/utils/exec"
)

const (
	// How many times to retry for a consistent read of /proc/mounts.
	maxListTries = 3
	// Number of fields per line in /proc/mounts as per the fstab man page.
	expectedNumFieldsPerLine = 6
	// Location of the mount file to use
	procMountsPath = "/proc/mounts"
	FIPS_ENABLED_FILE_PATH = "/host/proc/sys/crypto/fips_enabled"
	ENCRYPTED_UMOUNT_COMMAND = "umount.oci-fss"
	UMOUNT_COMMAND = "umount"
    FINDMNT_COMMAND = "findmnt"
    CAT_COMMAND = "cat"
    RPM_COMMAND = "rpm"
	// 'fsck' found errors and corrected them
	fsckErrorsCorrected = 1
	// 'fsck' found errors but exited without correcting them
	fsckErrorsUncorrected = 4
)

// Mounter provides the default implementation of mount.Interface
// for the linux platform.  This implementation assumes that the
// kubelet is running in the host's root mount namespace.
type Mounter struct {
	mounterPath string
	logger      *zap.SugaredLogger
}

// Mount mounts source to target as fstype with given options. 'source' and 'fstype' must
// be an emtpy string in case it's not required, e.g. for remount, or for auto filesystem
// type, where kernel handles fs type for you. The mount 'options' is a list of options,
// currently come from mount(8), e.g. "ro", "remount", "bind", etc. If no more option is
// required, call Mount with an empty string list or nil.
func (mounter *Mounter) Mount(source string, target string, fstype string, options []string) error {
	// Path to mounter binary if containerized mounter is needed. Otherwise, it is set to empty.
	// All Linux distros are expected to be shipped with a mount utility that an support bind mounts.
	mounterPath := ""
	bind, bindRemountOpts := isBind(options)
	if bind {
		err := doMount(mounter.logger, mounterPath, defaultMountCommand, source, target, fstype, []string{"bind"})
		if err != nil {
			return err
		}
		return doMount(mounter.logger, mounterPath, defaultMountCommand, source, target, fstype, bindRemountOpts)
	}
	// The list of filesystems that require containerized mounter on GCI image cluster
	fsTypesNeedMounter := []string{"nfs", "glusterfs", "ceph", "cifs"}
	for _, fst := range fsTypesNeedMounter {
		if fst == fstype {
			mounterPath = mounter.mounterPath
		}
	}
	return doMount(mounter.logger, mounterPath, defaultMountCommand, source, target, fstype, options)
}

// isBind detects whether a bind mount is being requested and makes the remount options to
// use in case of bind mount, due to the fact that bind mount doesn't respect mount options.
// The list equals:
//   options - 'bind' + 'remount' (no duplicate)
func isBind(options []string) (bool, []string) {
	bindRemountOpts := []string{"remount"}
	bind := false

	if len(options) != 0 {
		for _, option := range options {
			switch option {
			case "bind":
				bind = true
				break
			case "remount":
				break
			default:
				bindRemountOpts = append(bindRemountOpts, option)
			}
		}
	}

	return bind, bindRemountOpts
}

// doMount runs the mount command. mounterPath is the path to mounter binary if containerized mounter is used.
func doMount(logger *zap.SugaredLogger, mounterPath string, mountCmd string, source string, target string, fstype string, options []string) error {
	mountArgs := makeMountArgs(source, target, fstype, options)
	if len(mounterPath) > 0 {
		mountArgs = append([]string{mountCmd}, mountArgs...)
		mountCmd = mounterPath
	}
	logger.With("command", mountCmd, "args", mountArgs).Info("Mounting")
	command := exec.Command(mountCmd, mountArgs...)
	output, err := command.CombinedOutput()
	if err != nil {
		logger.With(
			zap.Error(err),
			"command", mountCmd,
			"source", source,
			"target", target,
			"fsType", fstype,
			"options", options,
			"output", string(output),
		).Error("Mount failed.")
		return fmt.Errorf("mount failed: %v\nMounting command: %s\nMounting arguments: %s %s %s %v\nOutput: %v\n",
			err, mountCmd, source, target, fstype, options, string(output))
	}
	logger.Debugf("Mount output: %v", string(output))
	return err
}

// makeMountArgs makes the arguments to the mount(8) command.
func makeMountArgs(source, target, fstype string, options []string) []string {
	// Build mount command as follows:
	//   mount [-t $fstype] [-o $options] [$source] $target
	mountArgs := []string{}
	if len(fstype) > 0 {
		mountArgs = append(mountArgs, "-t", fstype)
	}
	if len(options) > 0 {
		mountArgs = append(mountArgs, "-o", strings.Join(options, ","))
	}
	if len(source) > 0 {
		mountArgs = append(mountArgs, source)
	}
	mountArgs = append(mountArgs, target)

	return mountArgs
}

// Unmount unmounts the target.
func (mounter *Mounter) Unmount(target string) error {
	return mounter.unmount(target, UMOUNT_COMMAND)
}

func (mounter *Mounter) unmount(target string, unmountCommand string) error {
	mounter.logger.With("target", target).Info("Unmounting.")
	command := exec.Command(unmountCommand, target)
	output, err := command.CombinedOutput()
	if err != nil {
		mounter.logger.With(
			zap.Error(err),
			"command", unmountCommand,
			"target", target,
			"output", string(output),
		).Error("Unmount failed.")
		return fmt.Errorf("Unmount failed: %v\nUnmounting command: %s\nUnmounting arguments: %s\nOutput: %v\n", err, unmountCommand, target, string(output))
	}
	mounter.logger.Debugf("unmount output: %v", string(output))
	return nil
}

// Unmount unmounts the target.
func (mounter *Mounter) UnmountWithEncrypt(target string) error {
	return mounter.unmount(target, ENCRYPTED_UMOUNT_COMMAND)
}

func FindMount(mounter Interface, target string) ([]string, error) {
	mountArgs := []string{"-n", "-o", "SOURCE", "-T", target}
	command := exec.Command(FINDMNT_COMMAND, mountArgs...)
	output, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("findmnt failed: %v\narguments: %s\nOutput: %v\n", err, mountArgs, string(output))
	}

	sources := strings.Fields(string(output))
	return sources, nil
}

func IsFipsEnabled(mounter Interface) (string, error) {
	command := exec.Command(CAT_COMMAND, FIPS_ENABLED_FILE_PATH)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %v\narguments: %s\nOutput: %v\n", err, CAT_COMMAND, string(output))
	}

	return string(output), nil
}

func IsInTransitEncryptionPackageInstalled(mounter Interface) (bool, error) {
	args := []string{"-q", "-a", "--root=/host"}
	command := exec.Command(RPM_COMMAND, args...)
	output, err := command.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("command failed: %v\narguments: %s\nOutput: %v\n", err, RPM_COMMAND, string(output))
	}

	if len(output) > 0 {
		list := string(output)
		if strings.Contains(list, InTransitEncryptionPackageName) {
			return true, nil
		}
		return false, nil
	}
	return false, nil
}

// List returns a list of all mounted filesystems.
func (*Mounter) List() ([]MountPoint, error) {
	return listProcMounts(procMountsPath)
}

// IsLikelyNotMountPoint determines if a directory is not a mountpoint.
// It is fast but not necessarily ALWAYS correct. If the path is in fact
// a bind mount from one part of a mount to another it will not be detected.
// mkdir /tmp/a /tmp/b; mount --bin /tmp/a /tmp/b; IsLikelyNotMountPoint("/tmp/b")
// will return true. When in fact /tmp/b is a mount point. If this situation
// if of interest to you, don't use this function...
func (mounter *Mounter) IsLikelyNotMountPoint(file string) (bool, error) {
	stat, err := os.Stat(file)
	if err != nil {
		return true, err
	}
	rootStat, err := os.Lstat(file + "/..")
	if err != nil {
		return true, err
	}
	// If the directory has a different device as parent, then it is a mountpoint.
	if stat.Sys().(*syscall.Stat_t).Dev != rootStat.Sys().(*syscall.Stat_t).Dev {
		return false, nil
	}

	return true, nil
}

// IsNotMountPoint determines if a directory is a mountpoint.
// It should return ErrNotExist when the directory does not exist.
// IsNotMountPoint is more expensive than IsLikelyNotMountPoint.
// IsNotMountPoint detects bind mounts in linux.
// IsNotMountPoint enumerates all the mountpoints using List() and
// the list of mountpoints may be large, then it uses
// isMountPointMatch to evaluate whether the directory is a mountpoint.
func IsNotMountPoint(mounter Interface, file string) (bool, error) {
	// IsLikelyNotMountPoint provides a quick check
	// to determine whether file IS A mountpoint.
	notMnt, notMntErr := mounter.IsLikelyNotMountPoint(file)
	if notMntErr != nil && os.IsPermission(notMntErr) {
		// We were not allowed to do the simple stat() check, e.g. on NFS with
		// root_squash. Fall back to /proc/mounts check below.
		notMnt = true
		notMntErr = nil
	}
	if notMntErr != nil {
		return notMnt, notMntErr
	}
	// identified as mountpoint, so return this fact.
	if notMnt == false {
		return notMnt, nil
	}

	// Resolve any symlinks in file, kernel would do the same and use the resolved path in /proc/mounts.
	resolvedFile, err := filepath.EvalSymlinks(file)
	if err != nil {
		return true, err
	}

	// check all mountpoints since IsLikelyNotMountPoint
	// is not reliable for some mountpoint types.
	mountPoints, mountPointsErr := mounter.List()
	if mountPointsErr != nil {
		return notMnt, mountPointsErr
	}
	for _, mp := range mountPoints {
		if isMountPointMatch(mp, resolvedFile) {
			notMnt = false
			break
		}
	}
	return notMnt, nil
}

// isMountPointMatch returns true if the path in mp is the same as dir.
// Handles case where mountpoint dir has been renamed due to stale NFS mount.
func isMountPointMatch(mp MountPoint, dir string) bool {
	deletedDir := fmt.Sprintf("%s\\040(deleted)", dir)
	return ((mp.Path == dir) || (mp.Path == deletedDir))
}

// DeviceOpened checks if block device in use by calling Open with O_EXCL flag.
// If pathname is not a device, log and return false with nil error.
// If open returns errno EBUSY, return true with nil error.
// If open returns nil, return false with nil error.
// Otherwise, return false with error
func (mounter *Mounter) DeviceOpened(pathname string) (bool, error) {
	return exclusiveOpenFailsOnDevice(mounter.logger, pathname)
}

// PathIsDevice uses FileInfo returned from os.Stat to check if path refers
// to a device.
func (mounter *Mounter) PathIsDevice(pathname string) (bool, error) {
	return pathIsDevice(pathname)
}

func exclusiveOpenFailsOnDevice(logger *zap.SugaredLogger, pathname string) (bool, error) {
	isDevice, err := pathIsDevice(pathname)
	if err != nil {
		return false, fmt.Errorf(
			"PathIsDevice failed for path %q: %v",
			pathname,
			err)
	}
	if !isDevice {
		logger.With("path", pathname).Warn("Path does not refer to a device.")
		return false, nil
	}
	fd, errno := syscall.Open(pathname, syscall.O_RDONLY|syscall.O_EXCL, 0)
	// If the device is in use, open will return an invalid fd.
	// When this happens, it is expected that Close will fail and throw an error.
	defer syscall.Close(fd)
	if errno == nil {
		// device not in use
		return false, nil
	} else if errno == syscall.EBUSY {
		// device is in use
		return true, nil
	}
	// error during call to Open
	return false, errno
}

func pathIsDevice(pathname string) (bool, error) {
	finfo, err := os.Stat(pathname)
	if os.IsNotExist(err) {
		return false, nil
	}
	// err in call to os.Stat
	if err != nil {
		return false, err
	}
	// path refers to a device
	if finfo.Mode()&os.ModeDevice != 0 {
		return true, nil
	}
	// path does not refer to device
	return false, nil
}

//GetDeviceNameFromMount: given a mount point, find the device name from its global mount point
func (mounter *Mounter) GetDeviceNameFromMount(mountPath, pluginDir string) (string, error) {
	return getDeviceNameFromMount(mounter.logger, mounter, mountPath, pluginDir)
}

func listProcMounts(mountFilePath string) ([]MountPoint, error) {
	hash1, err := readProcMounts(mountFilePath, nil)
	if err != nil {
		return nil, err
	}

	for i := 0; i < maxListTries; i++ {
		mps := []MountPoint{}
		hash2, err := readProcMounts(mountFilePath, &mps)
		if err != nil {
			return nil, err
		}
		if hash1 == hash2 {
			// Success
			return mps, nil
		}
		hash1 = hash2
	}
	return nil, fmt.Errorf("failed to get a consistent snapshot of %v after %d tries", mountFilePath, maxListTries)
}

// readProcMounts reads the given mountFilePath (normally /proc/mounts) and produces a hash
// of the contents.  If the out argument is not nil, this fills it with MountPoint structs.
func readProcMounts(mountFilePath string, out *[]MountPoint) (uint32, error) {
	file, err := os.Open(mountFilePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return readProcMountsFrom(file, out)
}

func readProcMountsFrom(file io.Reader, out *[]MountPoint) (uint32, error) {
	hash := fnv.New32a()
	scanner := bufio.NewReader(file)
	for {
		line, err := scanner.ReadString('\n')
		if err == io.EOF {
			break
		}
		fields := strings.Fields(line)
		if len(fields) != expectedNumFieldsPerLine {
			return 0, fmt.Errorf("wrong number of fields (expected %d, got %d): %s", expectedNumFieldsPerLine, len(fields), line)
		}

		fmt.Fprintf(hash, "%s", line)

		if out != nil {
			mp := MountPoint{
				Device: fields[0],
				Path:   fields[1],
				Type:   fields[2],
				Opts:   strings.Split(fields[3], ","),
			}

			freq, err := strconv.Atoi(fields[4])
			if err != nil {
				return 0, err
			}
			mp.Freq = freq

			pass, err := strconv.Atoi(fields[5])
			if err != nil {
				return 0, err
			}
			mp.Pass = pass

			*out = append(*out, mp)
		}
	}
	return hash.Sum32(), nil
}

// formatAndMount uses unix utils to format and mount the given disk
func (mounter *SafeFormatAndMount) formatAndMount(source string, target string, fstype string, options []string) error {
	options = append(options, "defaults")
	mounter.Logger = mounter.Logger.With(
		"source", source,
		"target", target,
		"fstype", fstype,
		"options", options,
	)
	// Run fsck on the disk to fix repairable issues
	mounter.Logger.Info("Checking disk for issues using 'fsck'.")
	args := []string{"-a", source}
	cmd := mounter.Runner.Command("fsck", args...)
	out, err := cmd.CombinedOutput()
	mounter.Logger = mounter.Logger.With("output", out)
	if err != nil {
		ee, isExitError := err.(utilexec.ExitError)
		switch {
		case err == utilexec.ErrExecutableNotFound:
			mounter.Logger.Info("'fsck' not found on system; continuing mount without running 'fsck'.")
		case isExitError && ee.ExitStatus() == fsckErrorsCorrected:
			mounter.Logger.Info("Device has errors that were corrected with 'fsck'.")
		case isExitError && ee.ExitStatus() == fsckErrorsUncorrected:
			mounter.Logger.Info("'fsck' found errors on device but was unable to correct them.")
			return fmt.Errorf("'fsck' found errors on device %s but could not correct them: %s.", source, string(out))
		case isExitError && ee.ExitStatus() > fsckErrorsUncorrected:
			mounter.Logger.Error("'fsck' error.")
		}
	}

	// Try to mount the disk
	mounter.Logger.Info("Attempting to mount disk.")
	mountErr := mounter.Interface.Mount(source, target, fstype, options)
	if mountErr != nil {
		// Mount failed. This indicates either that the disk is unformatted or
		// it contains an unexpected filesystem.
		existingFormat, err := mounter.getDiskFormat(source)
		if err != nil {
			return err
		}
		if existingFormat == "" {
			// Disk is unformatted so format it.
			args = []string{source}
			// Use 'ext4' as the default
			if len(fstype) == 0 {
				fstype = "ext4"
			}

			if fstype == "ext4" || fstype == "ext3" {
				args = []string{"-F", source}
			}
			mounter.Logger.With("argruments", args).Info("Disk appears to be unformatted, attempting to format.")
			cmd := mounter.Runner.Command("mkfs."+fstype, args...)
			_, err := cmd.CombinedOutput()
			if err == nil {
				// the disk has been formatted successfully try to mount it again.
				mounter.Logger.Info("Disk successfully formatted.")
				return mounter.Interface.Mount(source, target, fstype, options)
			}
			mounter.Logger.With(zap.Error(err)).Error("Format of disk failed.")
			return err
		} else {
			// Disk is already formatted and failed to mount
			if len(fstype) == 0 || fstype == existingFormat {
				// This is mount error
				return mountErr
			} else {
				// Block device is formatted with unexpected filesystem, let the user know
				return fmt.Errorf("failed to mount the volume as %q, it already contains %s. Mount error: %v", fstype, existingFormat, mountErr)
			}
		}
	}
	return mountErr
}

// getDiskFormat uses 'lsblk' to determine a given disk's format
func (mounter *SafeFormatAndMount) getDiskFormat(disk string) (string, error) {
	args := []string{"-n", "-o", "FSTYPE", disk}
	cmd := mounter.Runner.Command("lsblk", args...)
	mounter.Logger.With("disk", disk, "arguments", args).Info("Attempting to determine if disk is formatted using lsblk.")
	dataOut, err := cmd.CombinedOutput()
	output := string(dataOut)

	if err != nil {
		mounter.Logger.With(zap.Error(err), "disk", disk).Error("Could not determine if disk is formatted.")
		return "", err
	}
	mounter.Logger.With("output", output).Info("Disk format determined.")

	// Split lsblk output into lines. Unformatted devices should contain only
	// "\n". Beware of "\n\n", that's a device with one empty partition.
	output = strings.TrimSuffix(output, "\n") // Avoid last empty line
	lines := strings.Split(output, "\n")
	if lines[0] != "" {
		// The device is formatted
		return lines[0], nil
	}

	if len(lines) == 1 {
		// The device is unformatted and has no dependent devices
		return "", nil
	}

	// The device has dependent devices, most probably partitions (LVM, LUKS
	// and MD RAID are reported as FSTYPE and caught above).
	return "unknown data, probably partitions", nil
}
