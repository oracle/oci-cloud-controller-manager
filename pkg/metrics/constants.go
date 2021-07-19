package metrics

const (
	// LBProvisionFailure is the OCI metric suffix for
	// LB provision failure
	LBProvisionFailure = "LB_PROVISION_FAILURE"
	// LBUpdateFailure is the OCI metric suffix for
	// LB update failure
	LBUpdateFailure = "LB_UPDATE_FAILURE"
	// LBDeleteFailure is the OCI metric suffix for
	// LB delete failure
	LBDeleteFailure = "LB_DELETE_FAILURE"

	// LBProvisionSuccess is the OCI metric suffix for
	// LB provision success
	LBProvisionSuccess = "LB_PROVISION_SUCCESS"
	// LBUpdateSuccess is the OCI metric suffix for
	// LB update success
	LBUpdateSuccess = "LB_UPDATE_SUCCESS"
	// LBDeleteSuccess is the OCI metric suffix for
	// LB delete success
	LBDeleteSuccess = "LB_DELETE_SUCCESS"
)

const (
	// PVProvisionFailure is the metric for PV provision failures
	PVProvisionFailure = "PV_PROVISION_FAILURE"
	// PVAttachFailure is the metric for PV attach failures
	PVAttachFailure = "PV_ATTACH_FAILURE"
	// PVDetachFailure is the metric for PV detach failure
	PVDetachFailure = "PV_DETACH_FAILURE"
	// PVDeleteFailure is the metric for PV delete failure
	PVDeleteFailure = "PV_DELETE_FAILURE"

	// PVProvisionSuccess is the metric used to track the time
	// taken for the provision operation
	PVProvisionSuccess = "PV_PROVISION_SUCCESS"
	// PVAttachSuccess is the metric used to track the time
	// taken for the provision operation
	PVAttachSuccess = "PV_ATTACH_SUCCESS"
	// PVDetachSuccess is the metric used to track the time
	// taken for the provision operation
	PVDetachSuccess = "PV_DETACH_SUCCESS"
	// PVDeleteSuccess is the metric used to track the time
	// taken for the provision operation
	PVDeleteSuccess = "PV_DELETE_SUCCESS"
)
