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

package csioptions

import (
	"flag"
	"time"
)

type CSIOptions struct {
	Master string
	Kubeconfig string
	CsiAddress string
	Endpoint string
	VolumeNamePrefix string
	VolumeNameUUIDLength int
	ShowVersion bool
	RetryIntervalStart time.Duration
	RetryIntervalMax time.Duration
	WorkerThreads uint
	OperationTimeout time.Duration
	EnableLeaderElection bool
	LeaderElectionType string
	LeaderElectionNamespace string
	StrictTopology bool
	Resync time.Duration
	Timeout time.Duration
	NodeID string
	LogLevel  string
}

func NewCSIOptions() (*CSIOptions) {
	csioptions := CSIOptions{
		Master: *flag.String("master", "", "kube master"),
		Kubeconfig: *flag.String("kubeconfig", "", "cluster kube config"),
		CsiAddress: *flag.String("csi-address", "/run/csi/socket", "Address of the CSI driver socket."),
		Endpoint: *flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint"),
		VolumeNamePrefix: *flag.String("volume-name-prefix", "pvc", "Prefix to apply to the name of a created volume."),
		VolumeNameUUIDLength: *flag.Int("volume-name-uuid-length", -1, "Truncates generated UUID of a created volume to this length. Defaults behavior is to NOT truncate."),
		ShowVersion: *flag.Bool("version", false, "Show version."),
		RetryIntervalStart: *flag.Duration("retry-interval-start", time.Second, "Initial retry interval of failed provisioning or deletion. It doubles with each failure, up to retry-interval-max."),
		RetryIntervalMax: *flag.Duration("retry-interval-max", 5*time.Minute, "Maximum retry interval of failed provisioning or deletion."),
		WorkerThreads: *flag.Uint("worker-threads", 100, "Number of provisioner worker threads, in other words nr. of simultaneous CSI calls."),
		OperationTimeout: *flag.Duration("op-timeout", 10*time.Second, "Timeout for waiting for creation or deletion of a volume"),
		EnableLeaderElection: *flag.Bool("enable-leader-election", false, "Enables leader election. If leader election is enabled, additional RBAC rules are required. Please refer to the Kubernetes CSI documentation for instructions on setting up these RBAC rules."),
		LeaderElectionType: *flag.String("leader-election-type", "endpoints", "the type of leader election, options are 'endpoints' (default) or 'leases' (strongly recommended). The 'endpoints' option is deprecated in favor of 'leases'."),
		LeaderElectionNamespace: *flag.String("leader-election-namespace", "", "Namespace where the leader election resource lives. Defaults to the pod namespace if not set."),
		StrictTopology: *flag.Bool("strict-topology", false, "Passes only selected node topology to CreateVolume Request, unlike default behavior of passing aggregated cluster topologies that match with topology keys of the selected node."),
		Resync: *flag.Duration("resync", 10*time.Minute, "Resync interval of the controller."),
		Timeout: *flag.Duration("timeout", 15*time.Second, "Timeout for waiting for attaching or detaching the volume."),
		NodeID: *flag.String("nodeid", "", "node id"),
		LogLevel: *flag.String("loglevel", "info", "log level"),
	}
	return &csioptions
}
