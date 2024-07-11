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
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	fssAddressSuffix               = "-fss.sock"
	fssVolumeNameAppendedPrefix    = "-fss"
	CrossNamespaceVolumeDataSource = "CrossNamespaceVolumeDataSource"
	VolumeAttributesClass          = "VolumeAttributesClass"
)

// CSIOptions structure which contains flag values
type CSIOptions struct {
	Master                      string
	Kubeconfig                  string
	CsiAddress                  string
	Endpoint                    string
	FssCsiAddress               string
	FssEndpoint                 string
	VolumeNamePrefix            string
	FssVolumeNamePrefix         string
	VolumeNameUUIDLength        int
	ShowVersion                 bool
	RetryIntervalStart          time.Duration
	RetryIntervalMax            time.Duration
	WorkerThreads               uint
	OperationTimeout            time.Duration
	EnableLeaderElection        bool
	LeaderElectionType          string
	LeaderElectionNamespace     string
	StrictTopology              bool
	Resync                      time.Duration
	Timeout                     time.Duration
	FeatureGates                map[string]bool
	FinalizerThreads            uint
	MetricsAddress              string
	MetricsPath                 string
	ExtraCreateMetadata         bool
	ReconcileSync               time.Duration
	EnableResizer               bool
	GroupSnapshotNamePrefix     string
	GroupSnapshotNameUUIDLength int
}

// NewCSIOptions initializes the flag
func NewCSIOptions() *CSIOptions {
	csioptions := CSIOptions{
		Master:                      *flag.String("master", "", "kube master"),
		Kubeconfig:                  *flag.String("kubeconfig", "", "cluster kube config"),
		CsiAddress:                  *flag.String("csi-address", "/run/csi/socket", "Address of the CSI BV driver socket."),
		Endpoint:                    *flag.String("csi-endpoint", "unix://tmp/csi.sock", "CSI BV endpoint"),
		FssCsiAddress:               *flag.String("fss-csi-address", "/run/fss/socket", "Address of the CSI FSS driver socket."),
		FssEndpoint:                 *flag.String("fss-csi-endpoint", "unix://tmp/csi-fss.sock", "CSI FSS endpoint"),
		VolumeNamePrefix:            *flag.String("csi-volume-name-prefix", "pvc", "Prefix to apply to the name of a created volume."),
		FssVolumeNamePrefix:         *flag.String("fss-csi-volume-name-prefix", "pvc", "Prefix to apply to the name of a volume created for FSS."),
		VolumeNameUUIDLength:        *flag.Int("csi-volume-name-uuid-length", -1, "Truncates generated UUID of a created volume to this length. Defaults behavior is to NOT truncate."),
		ShowVersion:                 *flag.Bool("csi-version", false, "Show version."),
		RetryIntervalStart:          *flag.Duration("csi-retry-interval-start", time.Second, "Initial retry interval of failed provisioning or deletion. It doubles with each failure, up to retry-interval-max."),
		RetryIntervalMax:            *flag.Duration("csi-retry-interval-max", 5*time.Minute, "Maximum retry interval of failed provisioning or deletion."),
		WorkerThreads:               *flag.Uint("csi-worker-threads", 100, "Number of provisioner worker threads, in other words nr. of simultaneous CSI calls."),
		OperationTimeout:            *flag.Duration("csi-op-timeout", 10*time.Second, "Timeout for waiting for creation or deletion of a volume"),
		EnableLeaderElection:        *flag.Bool("csi-enable-leader-election", false, "Enables leader election. If leader election is enabled, additional RBAC rules are required. Please refer to the Kubernetes CSI documentation for instructions on setting up these RBAC rules."),
		LeaderElectionType:          *flag.String("csi-leader-election-type", "endpoints", "the type of leader election, options are 'endpoints' (default) or 'leases' (strongly recommended). The 'endpoints' option is deprecated in favor of 'leases'."),
		LeaderElectionNamespace:     *flag.String("csi-leader-election-namespace", "", "Namespace where the leader election resource lives. Defaults to the pod namespace if not set."),
		StrictTopology:              *flag.Bool("csi-strict-topology", false, "Passes only selected node topology to CreateVolume Request, unlike default behavior of passing aggregated cluster topologies that match with topology keys of the selected node."),
		Resync:                      *flag.Duration("csi-resync", 10*time.Minute, "Resync interval of the controller."),
		Timeout:                     *flag.Duration("csi-timeout", 15*time.Second, "Timeout for waiting for attaching or detaching the volume."),
		FinalizerThreads:            *flag.Uint("cloning-protection-threads", 1, "Number of simultaniously running threads, handling cloning finalizer removal"),
		MetricsAddress:              *flag.String("metrics-address", "", "The TCP network address where the prometheus metrics endpoint will listen (example: `:8080`). The default is empty string, which means metrics endpoint is disabled."),
		MetricsPath:                 *flag.String("metrics-path", "/metrics", "The HTTP path where prometheus metrics will be exposed. Default is `/metrics`."),
		ExtraCreateMetadata:         *flag.Bool("extra-create-metadata", false, "If set, add pv/pvc metadata to plugin create requests as parameters."),
		ReconcileSync:               *flag.Duration("reconcile-sync", 1*time.Minute, "Resync interval of the VolumeAttachment reconciler."),
		EnableResizer:               *flag.Bool("csi-bv-expansion-enabled", false, "Enables go routine csi-resizer."),
		GroupSnapshotNamePrefix:     *flag.String("groupsnapshot-name-prefix", "groupsnapshot", "Prefix to apply to the name of a created group snapshot"),
		GroupSnapshotNameUUIDLength: *flag.Int("groupsnapshot-name-uuid-length", -1, "Length in characters for the generated uuid of a created group snapshot. Defaults behavior is to NOT truncate."),
	}
	return &csioptions
}

// GetFssAddress returns the fssAddress based on csiAddress
func GetFssAddress(csiAddress, defaultAddress string) string {
	logger := zap.L().Sugar()
	address := strings.Split(csiAddress, ".sock")
	if len(address) != 2 || !strings.HasSuffix(csiAddress, ".sock") {
		logger.Errorf("failed to parse csi-address : %s. Defaulting to : %s", csiAddress, defaultAddress)
		return defaultAddress
	}
	fssAddress := address[0] + fssAddressSuffix
	return fssAddress
}

// GetFssVolumeNamePrefix returns the fssVolumeNamePrefix based on csiVolumeNamePrefix
func GetFssVolumeNamePrefix(csiVolumeNamePrefix string) string {
	return csiVolumeNamePrefix + fssVolumeNameAppendedPrefix
}

// UpdateFeatureGates add CrossNamespaceVolumeDataSource (default value false) to featureGate if not present
// add VolumeAttributesClass (default value false) to featureGate if not present

func UpdateFeatureGates(featureGate map[string]bool) map[string]bool {
	//If key does not exist
	if featureGate != nil && !featureGate[CrossNamespaceVolumeDataSource] {
		featureGate[CrossNamespaceVolumeDataSource] = false
	}
	if featureGate != nil && !featureGate[VolumeAttributesClass] {
		featureGate[VolumeAttributesClass] = false
	}
	return featureGate
}
