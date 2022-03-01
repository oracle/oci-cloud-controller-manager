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
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

const directoryDeletePollInterval = 5 * time.Second

// UnmountPath is a common unmount routine that unmounts the given path and
// deletes the remaining directory if successful.
func UnmountPath(logger *zap.SugaredLogger, mountPath string, mounter Interface) error {
	if pathExists, pathErr := PathExists(mountPath); pathErr != nil {
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

// PathExists returns true if the specified path exists.
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}
