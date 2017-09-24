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

type Framework struct {
	configFile    string
	nodeSubnetOne string
	nodeSubnetTwo string

	Config    *client.Config
	Client    client.Interface
	Instances []*baremetal.Instance
}

func New() *Framework {
	return &Framework{
		configFile: path.Join(os.Getenv("HOME"), ".oci", "cloud-provider.cfg"),
	}
}

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

	f.Client = f.Client.Compartment(f.Config.Global.CompartmentOCID)

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

func (f *Framework) Run(run func() int) {
	os.Exit(run())
}

func (f *Framework) NodeSubnets() []string {
	return []string{f.nodeSubnetOne, f.nodeSubnetTwo}
}

func (f *Framework) WaitForInstance(id string) error {
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

func (f *Framework) CreateInstance(availabilityDomain string, subnetID string) (*baremetal.Instance, error) {
	instance, err := f.Client.LaunchInstance(
		availabilityDomain,
		f.Config.Global.CompartmentOCID,
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

func (f *Framework) Cleanup() {
	for _, instance := range f.Instances {
		err := f.Client.TerminateInstance(instance.ID, nil)
		if client.IsNotFound(err) {
			continue
		}
		if err != nil {
			glog.Errorf("unable to terminate instance: %v", err)
		}
	}
	f.Instances = []*baremetal.Instance{}
}
