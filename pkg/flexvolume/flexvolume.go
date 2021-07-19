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

package flexvolume

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"
)

// Defined to enable overriding in tests.
var out io.Writer = os.Stdout
var exit = os.Exit

// Status denotes the state of a Flexvolume call.
type Status string

// Options is the map (passed as JSON) to some Flexvolume calls.
type Options map[string]string

const (
	// StatusSuccess indicates that the driver call has succeeded.
	StatusSuccess Status = "Success"
	// StatusFailure indicates that the driver call has failed.
	StatusFailure Status = "Failure"
	// StatusNotSupported indicates that the driver call is not supported.
	StatusNotSupported Status = "Not supported"
)

// DriverStatus of a Flexvolume driver call.
type DriverStatus struct {
	// Status of the callout. One of "Success", "Failure" or "Not supported".
	Status Status `json:"status"`
	// Reason for success/failure.
	Message string `json:"message,omitempty"`
	// Path to the device attached. This field is valid only for attach calls.
	// e.g: /dev/sdx
	Device string `json:"device,omitempty"`
	// Represents volume is attached on the node.
	Attached bool `json:"attached,omitempty"`
}

// Option keys
const (
	OptionFSType    = "kubernetes.io/fsType"
	OptionReadWrite = "kubernetes.io/readwrite"
	OptionKeySecret = "kubernetes.io/secret"
	OptionFSGroup   = "kubernetes.io/fsGroup"
	OptionMountsDir = "kubernetes.io/mountsDir"

	OptionKeyPodName      = "kubernetes.io/pod.name"
	OptionKeyPodNamespace = "kubernetes.io/pod.namespace"
	OptionKeyPodUID       = "kubernetes.io/pod.uid"

	OptionKeyServiceAccountName = "kubernetes.io/serviceAccount.name"
)

// Driver is the main Flexvolume interface.
type Driver interface {
	Init(logger *zap.SugaredLogger) DriverStatus
	Attach(logger *zap.SugaredLogger, opts Options, nodeName string) DriverStatus
	Detach(logger *zap.SugaredLogger, mountDevice, nodeName string) DriverStatus
	WaitForAttach(mountDevice string, opts Options) DriverStatus
	IsAttached(logger *zap.SugaredLogger, opts Options, nodeName string) DriverStatus
	MountDevice(logger *zap.SugaredLogger, mountDir, mountDevice string, opts Options) DriverStatus
	UnmountDevice(logger *zap.SugaredLogger, mountDevice string) DriverStatus
	Mount(logger *zap.SugaredLogger, mountDir string, opts Options) DriverStatus
	Unmount(logger *zap.SugaredLogger, mountDir string) DriverStatus
}

// ExitWithResult outputs the given Result and exits with the appropriate exit
// code.
func ExitWithResult(logger *zap.SugaredLogger, result DriverStatus) {
	code := 1
	if result.Status == StatusSuccess || result.Status == StatusNotSupported {
		code = 0
	}

	res, err := json.Marshal(result)
	if err != nil {
		logger.With("result", result, zap.Error(err)).Error("Error marshaling result to JSON.")
		fmt.Fprintln(out, `{"status":"Failure","message":"Error marshaling result to JSON"}`)
	} else {
		s := string(res)
		logger.With("result", res).Debug("Command finished.")
		fmt.Fprintln(out, s)
	}
	exit(code)
}

// Fail creates a StatusFailure Result with a given message.
func Fail(logger *zap.SugaredLogger, a ...interface{}) DriverStatus {
	msg := fmt.Sprint(a...)
	logger.With("status", StatusFailure).Error(msg)
	return DriverStatus{
		Status:  StatusFailure,
		Message: msg,
	}
}

// Succeed creates a StatusSuccess Result with a given message.
func Succeed(logger *zap.SugaredLogger, a ...interface{}) DriverStatus {
	msg := fmt.Sprint(a...)
	logger.With("status", StatusSuccess).Info(msg)
	return DriverStatus{
		Status:  StatusSuccess,
		Message: msg,
	}
}

// NotSupported creates a StatusNotSupported Result with a given message.
func NotSupported(logger *zap.SugaredLogger, a ...interface{}) DriverStatus {
	msg := fmt.Sprint(a...)
	logger.With("status", StatusNotSupported).Warn(msg)
	return DriverStatus{
		Status:  StatusNotSupported,
		Message: msg,
	}
}

func processOpts(optsStr string) (Options, error) {
	opts := make(Options)
	if err := json.Unmarshal([]byte(optsStr), &opts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal options %q: %v", optsStr, err)
	}

	opts, err := DecodeKubeSecrets(opts)
	if err != nil {
		return nil, err
	}

	return opts, nil
}

// ExecDriver executes the appropriate FlexvolumeDriver command based on
// recieved call-out.
func ExecDriver(logger *zap.SugaredLogger, driver Driver, args []string) {
	if len(args) < 2 {
		ExitWithResult(logger, Fail(logger, "Expected at least one argument."))
	}

	logger = logger.With("command", args[1])
	logger.With("binary", args[0], "arguments", args[2:]).Debug("Exec called")

	switch args[1] {
	// <driver executable> init
	case "init":
		ExitWithResult(logger, driver.Init(logger))

	// <driver executable> getvolumename <json options>
	// Currently broken as of lates kube release (1.6.4). Work around hardcodes
	// exiting with StatusNotSupported.
	// TODO(apryde): Investigate current situation and version support
	// requirements.
	case "getvolumename":
		ExitWithResult(logger, NotSupported(logger, "getvolumename is broken as of kube 1.6.4"))

	// <driver executable> attach <json options> <node name>
	case "attach":
		if len(args) != 4 {
			ExitWithResult(logger, Fail(logger, "attach expected exactly 4 arguments; got ", args))
		}

		opts, err := processOpts(args[2])
		if err != nil {
			ExitWithResult(logger, Fail(logger, err))
		}

		nodeName := args[3]
		ExitWithResult(logger, driver.Attach(logger, opts, nodeName))

	// <driver executable> detach <mount device> <node name>
	case "detach":
		if len(args) != 4 {
			ExitWithResult(logger, Fail(logger, "detach expected exactly 4 arguments; got ", args))
		}

		mountDevice := args[2]
		nodeName := args[3]
		ExitWithResult(logger, driver.Detach(logger, mountDevice, nodeName))

	// <driver executable> waitforattach <mount device> <json options>
	case "waitforattach":
		if len(args) != 4 {
			ExitWithResult(logger, Fail(logger, "waitforattach expected exactly 4 arguments; got ", args))
		}

		mountDevice := args[2]
		opts, err := processOpts(args[3])
		if err != nil {
			ExitWithResult(logger, Fail(logger, err))
		}

		ExitWithResult(logger, driver.WaitForAttach(mountDevice, opts))

	// <driver executable> isattached <json options> <node name>
	case "isattached":
		if len(args) != 4 {
			ExitWithResult(logger, Fail(logger, "isattached expected exactly 4 arguments; got ", args))
		}

		opts, err := processOpts(args[2])
		if err != nil {
			ExitWithResult(logger, Fail(logger, err))
		}
		nodeName := args[3]
		ExitWithResult(logger, driver.IsAttached(logger, opts, nodeName))

	// <driver executable> mountdevice <mount dir> <mount device> <json options>
	case "mountdevice":
		if len(args) != 5 {
			ExitWithResult(logger, Fail(logger, "mountdevice expected exactly 5 arguments; got ", args))
		}

		mountDir := args[2]
		mountDevice := args[3]

		opts, err := processOpts(args[4])
		if err != nil {
			ExitWithResult(logger, Fail(logger, err))
		}

		ExitWithResult(logger, driver.MountDevice(logger, mountDir, mountDevice, opts))

	// <driver executable> unmountdevice <mount dir>
	case "unmountdevice":
		if len(args) != 3 {
			ExitWithResult(logger, Fail(logger, "unmountdevice expected exactly 3 arguments; got ", args))
		}

		mountDir := args[2]
		ExitWithResult(logger, driver.UnmountDevice(logger, mountDir))

	// <driver executable> mount <mount dir> <json options>
	case "mount":
		if len(args) != 4 {
			ExitWithResult(logger, Fail(logger, "mount expected exactly 4 arguments; got ", args))
		}

		mountDir := args[2]

		opts, err := processOpts(args[3])
		if err != nil {
			ExitWithResult(logger, Fail(logger, err))
		}

		ExitWithResult(logger, driver.Mount(logger, mountDir, opts))

	// <driver executable> unmount <mount dir>
	case "unmount":
		if len(args) != 3 {
			ExitWithResult(logger, Fail(logger, "mount expected exactly 3 arguments; got ", args))
		}

		mountDir := args[2]
		ExitWithResult(logger, driver.Unmount(logger, mountDir))

	default:
		ExitWithResult(logger, Fail(logger, "Invalid command; got ", args))
	}
}
