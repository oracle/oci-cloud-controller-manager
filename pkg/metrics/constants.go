// Copyright 2020 Oracle and/or its affiliates. All rights reserved.
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

package metrics

const (
	// LBProvision is the OCI metric suffix for LB provision
	LBProvision = "LB_PROVISION"
	// LBUpdate is the OCI metric suffix for LB update
	LBUpdate = "LB_UPDATE"
	// LBDelete is the OCI metric suffix for LB delete
	LBDelete = "LB_DELETE"

	// NLBProvision is the OCI metric suffix for NLB provision
	NLBProvision = "NLB_PROVISION"
	// NLBUpdate is the OCI metric suffix for NLB update
	NLBUpdate = "NLB_UPDATE"
	// NLBDelete is the OCI metric suffix for NLB delete
	NLBDelete = "NLB_DELETE"

	// PVProvision is the OCI metric suffix for PV provision
	PVProvision = "PV_PROVISION"
	// PVAttach is the OCI metric suffix for PV attach
	PVAttach = "PV_ATTACH"
	// PVDetach is the OCI metric suffix for PV detach
	PVDetach = "PV_DETACH"
	// PVDelete is the OCI metric suffix for PV delete
	PVDelete = "PV_DELETE"
	// PVExpand is the OCI metric suffix for PV Expand
	PVExpand = "PV_EXPAND"
	// PVClone is the OCI metric for PV Clone
	PVClone = "PV_CLONE"

	// FSSProvision is the OCI metric suffix for FSS provision
	FSSProvision = "FSS_PROVISION"
	// FSSDelete is the OCI metric suffix for FSS delete
	FSSDelete = "FSS_DELETE"

	// MTProvision is the OCI metric suffix for Mount Target provision
	MTProvision = "MT_PROVISION"
	// MTDelete is the OCI metric suffix for Mount Target delete
	MTDelete = "MT_DELETE"

	// ExportProvision is the OCI metric suffix for Export provision
	ExportProvision = "EXP_PROVISION"
	// ExportDelete is the OCI metric suffix for Export delete
	ExportDelete = "EXP_DELETE"

	// BlockSnapshotProvision is the OCI metric suffix for Block Volume Snapshot Provision
	BlockSnapshotProvision = "BSNAP_PROVISION"
	// BlockSnapshotDelete is the OCI metric suffix for Block Volume Snapshot Delete
	BlockSnapshotDelete = "BSNAP_DELETE"
	// BlockSnapshotRestore is the OCI metric suffix for Block Volume Snapshot Restore
	BlockSnapshotRestore = "BSNAP_RESTORE"

	// FssAllProvision is the OCI metric suffix for FSS end to end provision
	FssAllProvision = "FSS_ALL_PROVISION"

	// FssAllDelete is the OCI metric suffix for FSS end to end deletion
	FssAllDelete = "FSS_ALL_DELETE"

	ResourceOCIDDimension     = "resourceOCID"
	ComponentDimension        = "component"
	BackendSetsCountDimension = "backendSetsCount"
)
