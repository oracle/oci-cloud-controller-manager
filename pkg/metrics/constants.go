package metrics

const (
	// LBProvision is the OCI metric suffix for LB provision
	LBProvision = "LB_PROVISION"
	// LBUpdate is the OCI metric suffix for LB update
	LBUpdate = "LB_UPDATE"
	// LBDelete is the OCI metric suffix for LB delete
	LBDelete = "LB_DELETE"

	// PVProvision is the OCI metric suffix for PV provision
	PVProvision = "PV_PROVISION"
	// PVAttach is the OCI metric suffix for PV attach
	PVAttach = "PV_ATTACH"
	// PVDetach is the OCI metric suffix for PV detach
	PVDetach = "PV_DETACH"
	// PVDelete is the OCI metric suffix for PV delete
	PVDelete= "PV_DELETE"
	// PVProvision is the OCI metric suffix for PV provision
	PVExpand = "PV_EXPAND"

    ResourceOCIDDimension = "resourceOCID"
	ComponentDimension = "component"
	BackendSetsCountDimension = "backendSetsCount"
)
