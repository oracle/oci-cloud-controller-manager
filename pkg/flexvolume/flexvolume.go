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
	"log"
	"os"
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
	Init() DriverStatus
	Attach(opts Options, nodeName string) DriverStatus
	Detach(mountDevice, nodeName string) DriverStatus
	WaitForAttach(mountDevice string, opts Options) DriverStatus
	IsAttached(opts Options, nodeName string) DriverStatus
	MountDevice(mountDir, mountDevice string, opts Options) DriverStatus
	UnmountDevice(mountDevice string) DriverStatus
	Mount(mountDir string, opts Options) DriverStatus
	Unmount(mountDir string) DriverStatus
}

// ExitWithResult outputs the given Result and exits with the appropriate exit
// code.
func ExitWithResult(result DriverStatus) {
	code := 1
	if result.Status == StatusSuccess || result.Status == StatusNotSupported {
		code = 0
	}

	res, err := json.Marshal(result)
	if err != nil {
		log.Printf("Error marshaling result: %v", err)
		fmt.Fprintln(out, `{"status":"Failure","message":"Error marshaling result to JSON"}`)
	} else {
		s := string(res)
		log.Printf("Command result: %s", s)
		fmt.Fprintln(out, s)
	}
	exit(code)
}

// Fail creates a StatusFailure Result with a given message.
func Fail(a ...interface{}) DriverStatus {
	msg := fmt.Sprint(a...)
	return DriverStatus{
		Status:  StatusFailure,
		Message: msg,
	}
}

// Succeed creates a StatusSuccess Result with a given message.
func Succeed(a ...interface{}) DriverStatus {
	return DriverStatus{
		Status:  StatusSuccess,
		Message: fmt.Sprint(a...),
	}
}

// NotSupported creates a StatusNotSupported Result with a given message.
func NotSupported(a ...interface{}) DriverStatus {
	return DriverStatus{
		Status:  StatusNotSupported,
		Message: fmt.Sprint(a...),
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
func ExecDriver(driver Driver, args []string) {
	if len(args) < 2 {
		ExitWithResult(Fail("Expected at least one argument"))
	}

	log.Printf("'%s %s' called with %s", args[0], args[1], args[2:])

	switch args[1] {
	// <driver executable> init
	case "init":
		ExitWithResult(driver.Init())

	// <driver executable> getvolumename <json options>
	// Currently broken as of lates kube release (1.6.4). Work around hardcodes
	// exiting with StatusNotSupported.
	// TODO(apryde): Investigate current situation and version support
	// requirements.
	case "getvolumename":
		ExitWithResult(NotSupported("getvolumename is broken as of kube 1.6.4"))

	// <driver executable> attach <json options> <node name>
	case "attach":
		if len(args) != 4 {
			ExitWithResult(Fail("attach expected exactly 4 arguments; got ", args))
		}

		opts, err := processOpts(args[2])
		if err != nil {
			ExitWithResult(Fail(err))
		}

		nodeName := args[3]
		ExitWithResult(driver.Attach(opts, nodeName))

	// <driver executable> detach <mount device> <node name>
	case "detach":
		if len(args) != 4 {
			ExitWithResult(Fail("detach expected exactly 4 arguments; got ", args))
		}

		mountDevice := args[2]
		nodeName := args[3]
		ExitWithResult(driver.Detach(mountDevice, nodeName))

	// <driver executable> waitforattach <mount device> <json options>
	case "waitforattach":
		if len(args) != 4 {
			ExitWithResult(Fail("waitforattach expected exactly 4 arguments; got ", args))
		}

		mountDevice := args[2]
		opts, err := processOpts(args[3])
		if err != nil {
			ExitWithResult(Fail(err))
		}

		ExitWithResult(driver.WaitForAttach(mountDevice, opts))

	// <driver executable> isattached <json options> <node name>
	case "isattached":
		if len(args) != 4 {
			ExitWithResult(Fail("isattached expected exactly 4 arguments; got ", args))
		}

		opts, err := processOpts(args[2])
		if err != nil {
			ExitWithResult(Fail(err))
		}
		nodeName := args[3]
		ExitWithResult(driver.IsAttached(opts, nodeName))

	// <driver executable> mountdevice <mount dir> <mount device> <json options>
	case "mountdevice":
		if len(args) != 5 {
			ExitWithResult(Fail("mountdevice expected exactly 5 arguments; got ", args))
		}

		mountDir := args[2]
		mountDevice := args[3]

		opts, err := processOpts(args[4])
		if err != nil {
			ExitWithResult(Fail(err))
		}

		ExitWithResult(driver.MountDevice(mountDir, mountDevice, opts))

	// <driver executable> unmountdevice <mount dir>
	case "unmountdevice":
		if len(args) != 3 {
			ExitWithResult(Fail("unmountdevice expected exactly 3 arguments; got ", args))
		}

		mountDir := args[2]
		ExitWithResult(driver.UnmountDevice(mountDir))

	// <driver executable> mount <mount dir> <json options>
	case "mount":
		if len(args) != 4 {
			ExitWithResult(Fail("mount expected exactly 4 arguments; got ", args))
		}

		mountDir := args[2]

		opts, err := processOpts(args[3])
		if err != nil {
			ExitWithResult(Fail(err))
		}

		ExitWithResult(driver.Mount(mountDir, opts))

	// <driver executable> unmount <mount dir>
	case "unmount":
		if len(args) != 3 {
			ExitWithResult(Fail("mount expected exactly 3 arguments; got ", args))
		}

		mountDir := args[2]
		ExitWithResult(driver.Unmount(mountDir))

	default:
		ExitWithResult(Fail("Invalid command; got ", args))
	}
}
