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

package main

import (
	"flag"
	csicontrollerdriver "github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-controller-driver/csi-controller-driver"
	"github.com/oracle/oci-cloud-controller-manager/cmd/oci-csi-controller-driver/csioptions"
	"github.com/oracle/oci-cloud-controller-manager/pkg/logging"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util/signals"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

func main() {
	csiOptions := csioptions.CSIOptions{}
	flag.StringVar(&csiOptions.Endpoint, "endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	flag.StringVar(&csiOptions.Master, "master", "", "kube master")
	flag.StringVar(&csiOptions.Kubeconfig, "kubeconfig", "", "cluster kubeconfig")
	flag.Parse()
	stopCh := signals.SetupSignalHandler()
	log := logging.Logger()
	logger := log.Sugar()
	config, err := clientcmd.BuildConfigFromFlags(csiOptions.Master, csiOptions.Kubeconfig)
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
		logger.With(zap.Error(err)).Errorf("failed to get kube-apiserver version")
		return
	}
	logger.With("endpoint", csiOptions.Endpoint).Infof("Starting controller driver go routine.")
	go csicontrollerdriver.StartControllerDriver(csiOptions)
	<-stopCh
}
