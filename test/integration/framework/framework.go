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

package framework

import (
	"errors"
	"os"
	"path"
	"time"

	"github.com/golang/glog"
	baremetal "github.com/oracle/bmcs-go-sdk"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
)

const (
	// ubuntu image ocid
	instanceImageID = "ocid1.image.oc1.phx.aaaaaaaa2wjumduuoq6rqprrsmgu53eeyzp47vjztn355tkvsr3m2p57woqq"
	instanceShape   = "VM.Standard1.1"
)

// Framework used to help with integration testing.
type Framework struct {
	configFile    string
	nodeSubnetOne string
	nodeSubnetTwo string

	Config    *client.Config
	Client    client.Interface
	Instances []*baremetal.Instance
}

// New testing framework.
func New() *Framework {
	return &Framework{
		configFile: path.Join(os.Getenv("HOME"), ".oci", "cloud-provider.yaml"),
	}
}

// Init the framework which validates the configuration & env vars.
func (f *Framework) Init() error {
	if os.Getenv("OCI_CONFIG_FILE") != "" {
		f.configFile = os.Getenv("OCI_CONFIG_FILE")
	}

	file, err := os.Open(f.configFile)
	if err != nil {
		return err
	}

	f.Config, err = client.ReadConfig(file)
	if err != nil {
		return err
	}

	f.Client, err = client.New(f.Config)
	if err != nil {
		return err
	}

	f.nodeSubnetOne = os.Getenv("NODE_SUBNET_ONE")
	if f.nodeSubnetOne == "" {
		return errors.New("env var `NODE_SUBNET_ONE` is required")
	}

	f.nodeSubnetTwo = os.Getenv("NODE_SUBNET_TWO")
	if f.nodeSubnetTwo == "" {
		return errors.New("env var `NODE_SUBNET_TWO` is required")
	}

	return nil
}

// Run the tests and exit with the status code
func (f *Framework) Run(run func() int) {
	os.Exit(run())
}

// NodeSubnets returns the node subnets that should be used for testing.
func (f *Framework) NodeSubnets() []string {
	return []string{f.nodeSubnetOne, f.nodeSubnetTwo}
}

// WaitForInstance waits until the instance has a state of RUNNING.
func (f *Framework) WaitForInstance(id string) error {
	glog.Infof("Waiting for instance `%s` to be running", id)

	sleepTime := 30 * time.Second
	for {
		instance, err := f.Client.GetInstance(id)
		if err != nil {
			return err
		}
		if instance.State == baremetal.ResourceRunning {
			time.Sleep(sleepTime)
			return nil
		}
		glog.Infof("Instance is not running (%s)... sleeping for %v", instance.ID, sleepTime)
		time.Sleep(sleepTime)
	}
}

// CreateInstance creates an instance and stores a reference for cleanup.
func (f *Framework) CreateInstance(availabilityDomain string, subnetID string) (*baremetal.Instance, error) {
	instance, err := f.Client.LaunchInstance(
		availabilityDomain,
		f.Config.Auth.CompartmentOCID,
		instanceImageID,
		instanceShape,
		subnetID,
		&baremetal.LaunchInstanceOptions{},
	)
	if err != nil {
		return nil, err
	}

	f.Instances = append(f.Instances, instance)
	return instance, nil
}

// Cleanup all the instances created by the test framework.
func (f *Framework) Cleanup() {
	glog.Info("Running instance cleanup")
	for _, instance := range f.Instances {
		glog.Infof("Terminating instance for cleanup `%s`", instance.ID)
		err := f.Client.TerminateInstance(instance.ID, nil)
		if client.IsNotFound(err) {
			continue
		}
		if err != nil {
			glog.Errorf("unable to terminate instance: %v", err)
		}
	}
	f.Instances = []*baremetal.Instance{}
	glog.Info("Instance cleanup is done")
}
