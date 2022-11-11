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
	PVDelete= "PV_DELETE"
	// PVExpand is the OCI metric suffix for PV Expand
	PVExpand = "PV_EXPAND"

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

	ResourceOCIDDimension     = "resourceOCID"
	ComponentDimension        = "component"
	BackendSetsCountDimension = "backendSetsCount"
)
