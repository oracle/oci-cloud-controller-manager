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
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/pborman/uuid"
	"k8s.io/api/core/v1"
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
		id = string(uuid.NewUUID())
	}

	Logf("Election id: %q", id)

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
