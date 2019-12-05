// Copyright 2019 Oracle and/or its affiliates. All rights reserved.
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

package csicontroller

import (
	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-controller-driver/csi-attacher"
	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-controller-driver/csi-controller-driver"
	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-controller-driver/csi-provisioner"
	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-controller-driver/csioptions"
	"github.com/oracle/oci-cloud-controller-manager/pkg/logging"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

//Run main function to start CSI Controller
func Run(csioptions csioptions.CSIOptions, stopCh <-chan struct{}) error {
	log := logging.Logger()
	logger := log.Sugar()
	config, err := clientcmd.BuildConfigFromFlags(csioptions.Master, csioptions.Kubeconfig)
	clientset, err := kubernetes.NewForConfig(config)
	err = wait.PollUntil(15*time.Second, func() (done bool, err error) {
		_, err = clientset.Discovery().ServerVersion()
		if err != nil {
			logger.With(zap.Error(err)).Info("failed to get kube-apiserver version, will retry again")
			return false, nil
		}
		return true, nil
	}, stopCh)
	if err != nil {
		return errors.Wrapf(err, "failed to get kube-apiserver version")
	}

	go csiprovisioner.StartCSIProvisioner(csioptions)
	go csiattacher.StartCSIAttacher(csioptions)
	go csicontrollerdriver.StartControllerDriver(csioptions)
	<-stopCh
	return nil
}
