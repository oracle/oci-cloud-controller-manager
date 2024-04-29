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

package disk

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/kubernetes/pkg/volume/util/hostutil"
	"k8s.io/mount-utils"
	utilexec "k8s.io/utils/exec"
)

const (
	directoryDeletePollInterval = 5 * time.Second
	errNotMounted               = "not mounted"

	EncryptedUmountCommand = "encrypt-umount"

	EncryptionMountCommand = "encrypt-mount"
	UnmountCommand         = "umount"
)

func MountWithEncrypt(logger *zap.SugaredLogger, source string, target string, fstype string, options []string) error {
	mountArgs, mountArgsLogStr := MakeMountArgs(source, target, fstype, options)
	mountArgsLogStr = EncryptionMountCommand + " " + mountArgsLogStr

	logger.Debug("Mounting cmd (%s) with arguments (%s)", EncryptionMountCommand, mountArgsLogStr)
	command := exec.Command(EncryptionMountCommand, mountArgs...)
	output, err := command.CombinedOutput()
	if err != nil {
		if err.Error() == "wait: no child processes" {
			if command.ProcessState.Success() {
				// We don't consider errNoChildProcesses an error if the process itself succeeded (see - k/k issue #103753).
				return nil
			}
			// Rewrite err with the actual exit error of the process.
			err = &exec.ExitError{ProcessState: command.ProcessState}
		}
		logger.Errorf("Mount failed: %v\nMounting command: %s\nMounting arguments: %s\nOutput: %s\n", err, EncryptionMountCommand, mountArgsLogStr, string(output))
		return fmt.Errorf("mount failed: %v\nMounting command: %s\nMounting arguments: %s\nOutput: %s",
			err, EncryptionMountCommand, mountArgsLogStr, string(output))
	}
	return err
}

func MakeMountArgs(source, target, fstype string, options []string) (mountArgs []string, mountArgsLogStr string) {
	// Build mount command as follows:
	//   mount [$mountFlags] [-t $fstype] [-o $options] [$source] $target
	mountArgs = []string{}
	mountArgsLogStr = ""

	if len(fstype) > 0 {
		mountArgs = append(mountArgs, "-t", fstype)
		mountArgsLogStr += strings.Join(mountArgs, " ")
	}
	if len(options) > 0 {
		mountArgs = append(mountArgs, "-o", strings.Join(options, ","))
		mountArgsLogStr += " -o " + strings.Join(options, ",")
	}
	if len(source) > 0 {
		mountArgs = append(mountArgs, source)
		mountArgsLogStr += " " + source
	}
	mountArgs = append(mountArgs, target)
	mountArgsLogStr += " " + target

	return mountArgs, mountArgsLogStr
}

// UnmountPath is a common unmount routine that unmounts the given path and
// deletes the remaining directory if successful.
func UnmountPath(logger *zap.SugaredLogger, mountPath string, mounter mount.Interface) error {
	if pathExists, pathErr := mount.PathExists(mountPath); pathErr != nil {
		return fmt.Errorf("Error checking if path exists: %v", pathErr)
	} else if !pathExists {
		logger.With("mount path", mountPath).Warn("Unmount skipped because path does not exist.")
		return nil
	}

	notMnt, err := mounter.IsLikelyNotMountPoint(mountPath)
	if err != nil {
		return err
	}
	if notMnt {
		logger.With("mount path", mountPath).Warn("Mount path is not a mount point. Removing directory.")
		return os.Remove(mountPath)
	}

	// Unmount the mount path
	if err := mounter.Unmount(mountPath); err != nil {
		return err
	}
	notMnt, mntErr := mounter.IsLikelyNotMountPoint(mountPath)
	if mntErr != nil {
		return err
	}
	if notMnt {
		logger.With("mount path", mountPath).Info("Mount path is unmounted. Removing directory.")
		return WaitForDirectoryDeletion(logger, mountPath)
	}
	return fmt.Errorf("Failed to unmount path %v", mountPath)
}

func WaitForDirectoryDeletion(logger *zap.SugaredLogger, mountPath string) error {
	var err error
	// Try removing the mount path thrice, else suppress the error
	for loopCounter := 0; loopCounter < 3; loopCounter += 1 {
		if err = os.Remove(mountPath); err != nil {
			logger.With("mount path", mountPath, "error", err).Warn("Mount path couldn't be deleted. Trying again...")
			time.Sleep(directoryDeletePollInterval)
		} else {
			logger.With("mount path", mountPath).Info("Mount path deleted.")
			return nil
		}
	}
	logger.With("mount path", mountPath, "error", err).Warn("Mount path couldn't be deleted.")
	return nil
}

// Unmount the target that is in-transit encryption enabled
func UnmountWithEncrypt(logger *zap.SugaredLogger, target string) error {
	logger.With("target", target).Info("Unmounting in-transit encryption mount point.")
	command := exec.Command(EncryptedUmountCommand, target)
	output, err := command.CombinedOutput()
	if err != nil {
		logger.With(
			zap.Error(err),
			"command", EncryptedUmountCommand,
			"target", target,
			"output", string(output),
		).Error("Unmount failed.")
		return fmt.Errorf("Unmount failed: %v\nUnmounting command: %s\nUnmounting arguments: %s\nOutput: %v\n", err, EncryptedUmountCommand, target, string(output))
	}
	logger.Debugf("unmount output: %v", string(output))
	return nil
}

func getDiskFormat(ex utilexec.Interface, disk string, logger *zap.SugaredLogger) (string, error) {
	args := []string{"-p", "-s", "TYPE", "-s", "PTTYPE", "-o", "export", disk}
	logger.With(disk).Infof("Attempting to determine if disk %q is formatted using blkid with args: %q", disk, args)
	dataOut, err := ex.Command("blkid", args...).CombinedOutput()
	output := string(dataOut)
	logger.Infof("Output: %q", output)

	if err != nil {
		if exit, ok := err.(utilexec.ExitError); ok {
			if exit.ExitStatus() == 2 {
				// Disk device is unformatted.
				// For `blkid`, if the specified token (TYPE/PTTYPE, etc) was
				// not found, or no (specified) devices could be identified, an
				// exit code of 2 is returned.
				return "", nil
			}
		}
		logger.With(disk).Errorf("Could not determine if disk %q is formatted (%v)", disk, err)
		return "", err
	}

	var fstype, pttype string

	lines := strings.Split(output, "\n")
	for _, l := range lines {
		if len(l) <= 0 {
			// Ignore empty line.
			continue
		}
		cs := strings.Split(l, "=")
		if len(cs) != 2 {
			return "", fmt.Errorf("blkid returns invalid output: %s", output)
		}
		// TYPE is filesystem type, and PTTYPE is partition table type, according
		// to https://www.kernel.org/pub/linux/utils/util-linux/v2.21/libblkid-docs/.
		if cs[0] == "TYPE" {
			fstype = cs[1]
		} else if cs[0] == "PTTYPE" {
			pttype = cs[1]
		}
	}

	if len(pttype) > 0 {
		logger.With(disk).Infof("Disk %s detected partition table type: %s", disk, pttype)
		// Returns a special non-empty string as filesystem type, then kubelet
		// will not format it.
		return "unknown data, probably partitions", nil
	}

	return fstype, nil
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

func UnmountWithForce(targetPath string) error {
	command := exec.Command(UnmountCommand, "-f", targetPath)
	output, err := command.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), errNotMounted) {
			return nil
		}
		return status.Errorf(codes.Internal, err.Error())
	}
	return nil
}
