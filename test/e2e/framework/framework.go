// Copyright 2018 Oracle and/or its affiliates. All rights reserved.
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
	"fmt"
	"os"
	"time"

	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
)

// AquireRunLock blocks until the test run lock is required or a timeout
// elapses. A lock is required as only one test run can safely be executed on
// the same cluster at any given time.
func AquireRunLock(client clientset.Interface, lockName string) error {
	lec, err := makeLeaderElectionConfig(client, lockName)
	if err != nil {
		return err
	}

	readyCh := make(chan struct{})
	lec.Callbacks = leaderelection.LeaderCallbacks{
		OnStartedLeading: func(stop <-chan struct{}) {
			Logf("Test run lock aquired")
			readyCh <- struct{}{}
		},
		OnStoppedLeading: func() {
			Failf("Lost test run lock unexpectedly")
		},
	}

	le, err := leaderelection.NewLeaderElector(*lec)
	if err != nil {
		return err
	}

	go le.Run()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-readyCh:
			return nil
		case <-ticker.C:
			Logf("Waiting to aquire test run lock. %q currently has it.", le.GetLeader())
		case <-time.After(30 * time.Minute):
			return errors.New("timed out trying to aquire test run lock")
		}
	}
	panic("unreachable")
}

func makeLeaderElectionConfig(client clientset.Interface, lockName string) (*leaderelection.LeaderElectionConfig, error) {
	eventBroadcaster := record.NewBroadcaster()
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: lockName})

	id := os.Getenv("WERCKER_STEP_ID")
	if id == "" {
		id = uuid.NewUUID().String()
	}

	Logf("Test run lock id: %q", id)

	rl, err := resourcelock.New(
		resourcelock.ConfigMapsResourceLock,
		"kube-system",
		lockName,
		client.CoreV1(),
		resourcelock.ResourceLockConfig{
			Identity:      id,
			EventRecorder: recorder,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("couldn't create resource lock: %v", err)
	}

	return &leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: 10 * time.Second,
		RenewDeadline: 5 * time.Second,
		RetryPeriod:   1 * time.Second,
	}, nil
}

// CreateAndAwaitDaemonSet creates/updates the given DaemonSet and waits for it
// to be ready.
func CreateAndAwaitDaemonSet(client clientset.Interface, desired *appsv1.DaemonSet) error {
	actual, err := client.AppsV1().DaemonSets(desired.Namespace).Create(desired)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrapf(err, "failed to create %q DaemonSet", desired.Name)
		}
		Logf("%q DaemonSet already exists. Updating.", desired.Name)
		actual, err = client.AppsV1().DaemonSets(desired.Namespace).Update(desired)
		if err != nil {
			return errors.Wrapf(err, "updating DaemonSet %q", desired.Name)
		}
	} else {
		Logf("Created DaemonSet %q in namespace %q", actual.Name, actual.Namespace)
	}

	return wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
		actual, err := client.AppsV1().DaemonSets(actual.Namespace).Get(actual.Name, metav1.GetOptions{})
		if err != nil {
			return false, errors.Wrap(err, "waiting for DaemonSet to be ready")
		}

		if actual.Status.DesiredNumberScheduled != 0 && actual.Status.NumberReady == actual.Status.DesiredNumberScheduled {
			return true, nil
		}

		Logf("%q DaemonSet not yet ready (diesired=%d, ready=%d). Waiting...",
			actual.Name, actual.Status.DesiredNumberScheduled, actual.Status.NumberReady)
		return false, nil
	})
}

// CreateAndAwaitDeployment creates/updates the given Deployment and waits for
// it to be ready.
func CreateAndAwaitDeployment(client clientset.Interface, desired *appsv1.Deployment) error {
	actual, err := client.AppsV1().Deployments(desired.Namespace).Create(desired)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrapf(err, "failed to create Deployment %q", desired.Name)
		}
		Logf("Deployment %q already exists. Updating.", desired.Name)
		actual, err = client.AppsV1().Deployments(desired.Namespace).Update(desired)
		if err != nil {
			return errors.Wrapf(err, "updating Deployment %q", desired.Name)
		}
	} else {
		Logf("Created Deployment %q in namespace %q", actual.Name, actual.Namespace)
	}

	return wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
		actual, err := client.AppsV1().Deployments(actual.Namespace).Get(actual.Name, metav1.GetOptions{})
		if err != nil {
			return false, errors.Wrap(err, "waiting for Deployment to be ready")
		}
		if actual.Status.Replicas != 0 && actual.Status.Replicas == actual.Status.ReadyReplicas {
			return true, nil
		}
		Logf("%s Deployment not yet ready (replicas=%d, readyReplicas=%d). Waiting...",
			actual.Name, actual.Status.Replicas, actual.Status.ReadyReplicas)
		return false, nil
	})
}
