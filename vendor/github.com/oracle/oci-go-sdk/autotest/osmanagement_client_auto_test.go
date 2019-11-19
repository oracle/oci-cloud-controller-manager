package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/osmanagement"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createOsManagementClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := osmanagement.NewOsManagementClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientAddPackagesToSoftwareSource(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "AddPackagesToSoftwareSource")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("AddPackagesToSoftwareSource is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "AddPackagesToSoftwareSource", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "AddPackagesToSoftwareSource")
	assert.NoError(t, err)

	type AddPackagesToSoftwareSourceRequestInfo struct {
		ContainerId string
		Request     osmanagement.AddPackagesToSoftwareSourceRequest
	}

	var requests []AddPackagesToSoftwareSourceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.AddPackagesToSoftwareSource(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientAttachChildSoftwareSourceToManagedInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "AttachChildSoftwareSourceToManagedInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("AttachChildSoftwareSourceToManagedInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "AttachChildSoftwareSourceToManagedInstance", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "AttachChildSoftwareSourceToManagedInstance")
	assert.NoError(t, err)

	type AttachChildSoftwareSourceToManagedInstanceRequestInfo struct {
		ContainerId string
		Request     osmanagement.AttachChildSoftwareSourceToManagedInstanceRequest
	}

	var requests []AttachChildSoftwareSourceToManagedInstanceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.AttachChildSoftwareSourceToManagedInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientAttachManagedInstanceToManagedInstanceGroup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "AttachManagedInstanceToManagedInstanceGroup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("AttachManagedInstanceToManagedInstanceGroup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "AttachManagedInstanceToManagedInstanceGroup", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "AttachManagedInstanceToManagedInstanceGroup")
	assert.NoError(t, err)

	type AttachManagedInstanceToManagedInstanceGroupRequestInfo struct {
		ContainerId string
		Request     osmanagement.AttachManagedInstanceToManagedInstanceGroupRequest
	}

	var requests []AttachManagedInstanceToManagedInstanceGroupRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.AttachManagedInstanceToManagedInstanceGroup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientAttachParentSoftwareSourceToManagedInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "AttachParentSoftwareSourceToManagedInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("AttachParentSoftwareSourceToManagedInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "AttachParentSoftwareSourceToManagedInstance", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "AttachParentSoftwareSourceToManagedInstance")
	assert.NoError(t, err)

	type AttachParentSoftwareSourceToManagedInstanceRequestInfo struct {
		ContainerId string
		Request     osmanagement.AttachParentSoftwareSourceToManagedInstanceRequest
	}

	var requests []AttachParentSoftwareSourceToManagedInstanceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.AttachParentSoftwareSourceToManagedInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientChangeManagedInstanceGroupCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "ChangeManagedInstanceGroupCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeManagedInstanceGroupCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "ChangeManagedInstanceGroupCompartment", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "ChangeManagedInstanceGroupCompartment")
	assert.NoError(t, err)

	type ChangeManagedInstanceGroupCompartmentRequestInfo struct {
		ContainerId string
		Request     osmanagement.ChangeManagedInstanceGroupCompartmentRequest
	}

	var requests []ChangeManagedInstanceGroupCompartmentRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.ChangeManagedInstanceGroupCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientChangeScheduledJobCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "ChangeScheduledJobCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeScheduledJobCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "ChangeScheduledJobCompartment", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "ChangeScheduledJobCompartment")
	assert.NoError(t, err)

	type ChangeScheduledJobCompartmentRequestInfo struct {
		ContainerId string
		Request     osmanagement.ChangeScheduledJobCompartmentRequest
	}

	var requests []ChangeScheduledJobCompartmentRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.ChangeScheduledJobCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientChangeSoftwareSourceCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "ChangeSoftwareSourceCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeSoftwareSourceCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "ChangeSoftwareSourceCompartment", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "ChangeSoftwareSourceCompartment")
	assert.NoError(t, err)

	type ChangeSoftwareSourceCompartmentRequestInfo struct {
		ContainerId string
		Request     osmanagement.ChangeSoftwareSourceCompartmentRequest
	}

	var requests []ChangeSoftwareSourceCompartmentRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.ChangeSoftwareSourceCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientCreateManagedInstanceGroup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "CreateManagedInstanceGroup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateManagedInstanceGroup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "CreateManagedInstanceGroup", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "CreateManagedInstanceGroup")
	assert.NoError(t, err)

	type CreateManagedInstanceGroupRequestInfo struct {
		ContainerId string
		Request     osmanagement.CreateManagedInstanceGroupRequest
	}

	var requests []CreateManagedInstanceGroupRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.CreateManagedInstanceGroup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientCreateScheduledJob(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "CreateScheduledJob")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateScheduledJob is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "CreateScheduledJob", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "CreateScheduledJob")
	assert.NoError(t, err)

	type CreateScheduledJobRequestInfo struct {
		ContainerId string
		Request     osmanagement.CreateScheduledJobRequest
	}

	var requests []CreateScheduledJobRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.CreateScheduledJob(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientCreateSoftwareSource(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "CreateSoftwareSource")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateSoftwareSource is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "CreateSoftwareSource", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "CreateSoftwareSource")
	assert.NoError(t, err)

	type CreateSoftwareSourceRequestInfo struct {
		ContainerId string
		Request     osmanagement.CreateSoftwareSourceRequest
	}

	var requests []CreateSoftwareSourceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.CreateSoftwareSource(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientDeleteManagedInstanceGroup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "DeleteManagedInstanceGroup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteManagedInstanceGroup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "DeleteManagedInstanceGroup", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "DeleteManagedInstanceGroup")
	assert.NoError(t, err)

	type DeleteManagedInstanceGroupRequestInfo struct {
		ContainerId string
		Request     osmanagement.DeleteManagedInstanceGroupRequest
	}

	var requests []DeleteManagedInstanceGroupRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.DeleteManagedInstanceGroup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientDeleteScheduledJob(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "DeleteScheduledJob")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteScheduledJob is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "DeleteScheduledJob", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "DeleteScheduledJob")
	assert.NoError(t, err)

	type DeleteScheduledJobRequestInfo struct {
		ContainerId string
		Request     osmanagement.DeleteScheduledJobRequest
	}

	var requests []DeleteScheduledJobRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.DeleteScheduledJob(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientDeleteSoftwareSource(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "DeleteSoftwareSource")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteSoftwareSource is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "DeleteSoftwareSource", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "DeleteSoftwareSource")
	assert.NoError(t, err)

	type DeleteSoftwareSourceRequestInfo struct {
		ContainerId string
		Request     osmanagement.DeleteSoftwareSourceRequest
	}

	var requests []DeleteSoftwareSourceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.DeleteSoftwareSource(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientDetachChildSoftwareSourceFromManagedInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "DetachChildSoftwareSourceFromManagedInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DetachChildSoftwareSourceFromManagedInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "DetachChildSoftwareSourceFromManagedInstance", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "DetachChildSoftwareSourceFromManagedInstance")
	assert.NoError(t, err)

	type DetachChildSoftwareSourceFromManagedInstanceRequestInfo struct {
		ContainerId string
		Request     osmanagement.DetachChildSoftwareSourceFromManagedInstanceRequest
	}

	var requests []DetachChildSoftwareSourceFromManagedInstanceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.DetachChildSoftwareSourceFromManagedInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientDetachManagedInstanceFromManagedInstanceGroup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "DetachManagedInstanceFromManagedInstanceGroup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DetachManagedInstanceFromManagedInstanceGroup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "DetachManagedInstanceFromManagedInstanceGroup", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "DetachManagedInstanceFromManagedInstanceGroup")
	assert.NoError(t, err)

	type DetachManagedInstanceFromManagedInstanceGroupRequestInfo struct {
		ContainerId string
		Request     osmanagement.DetachManagedInstanceFromManagedInstanceGroupRequest
	}

	var requests []DetachManagedInstanceFromManagedInstanceGroupRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.DetachManagedInstanceFromManagedInstanceGroup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientDetachParentSoftwareSourceFromManagedInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "DetachParentSoftwareSourceFromManagedInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DetachParentSoftwareSourceFromManagedInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "DetachParentSoftwareSourceFromManagedInstance", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "DetachParentSoftwareSourceFromManagedInstance")
	assert.NoError(t, err)

	type DetachParentSoftwareSourceFromManagedInstanceRequestInfo struct {
		ContainerId string
		Request     osmanagement.DetachParentSoftwareSourceFromManagedInstanceRequest
	}

	var requests []DetachParentSoftwareSourceFromManagedInstanceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.DetachParentSoftwareSourceFromManagedInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientGetErratum(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "GetErratum")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetErratum is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "GetErratum", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "GetErratum")
	assert.NoError(t, err)

	type GetErratumRequestInfo struct {
		ContainerId string
		Request     osmanagement.GetErratumRequest
	}

	var requests []GetErratumRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.GetErratum(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientGetManagedInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "GetManagedInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetManagedInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "GetManagedInstance", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "GetManagedInstance")
	assert.NoError(t, err)

	type GetManagedInstanceRequestInfo struct {
		ContainerId string
		Request     osmanagement.GetManagedInstanceRequest
	}

	var requests []GetManagedInstanceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.GetManagedInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientGetManagedInstanceGroup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "GetManagedInstanceGroup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetManagedInstanceGroup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "GetManagedInstanceGroup", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "GetManagedInstanceGroup")
	assert.NoError(t, err)

	type GetManagedInstanceGroupRequestInfo struct {
		ContainerId string
		Request     osmanagement.GetManagedInstanceGroupRequest
	}

	var requests []GetManagedInstanceGroupRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.GetManagedInstanceGroup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientGetScheduledJob(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "GetScheduledJob")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetScheduledJob is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "GetScheduledJob", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "GetScheduledJob")
	assert.NoError(t, err)

	type GetScheduledJobRequestInfo struct {
		ContainerId string
		Request     osmanagement.GetScheduledJobRequest
	}

	var requests []GetScheduledJobRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.GetScheduledJob(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientGetSoftwarePackage(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "GetSoftwarePackage")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetSoftwarePackage is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "GetSoftwarePackage", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "GetSoftwarePackage")
	assert.NoError(t, err)

	type GetSoftwarePackageRequestInfo struct {
		ContainerId string
		Request     osmanagement.GetSoftwarePackageRequest
	}

	var requests []GetSoftwarePackageRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.GetSoftwarePackage(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientGetSoftwareSource(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "GetSoftwareSource")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetSoftwareSource is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "GetSoftwareSource", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "GetSoftwareSource")
	assert.NoError(t, err)

	type GetSoftwareSourceRequestInfo struct {
		ContainerId string
		Request     osmanagement.GetSoftwareSourceRequest
	}

	var requests []GetSoftwareSourceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.GetSoftwareSource(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientGetWorkRequest(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "GetWorkRequest")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetWorkRequest is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "GetWorkRequest", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "GetWorkRequest")
	assert.NoError(t, err)

	type GetWorkRequestRequestInfo struct {
		ContainerId string
		Request     osmanagement.GetWorkRequestRequest
	}

	var requests []GetWorkRequestRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.GetWorkRequest(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientInstallAllPackageUpdatesOnManagedInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "InstallAllPackageUpdatesOnManagedInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("InstallAllPackageUpdatesOnManagedInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "InstallAllPackageUpdatesOnManagedInstance", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "InstallAllPackageUpdatesOnManagedInstance")
	assert.NoError(t, err)

	type InstallAllPackageUpdatesOnManagedInstanceRequestInfo struct {
		ContainerId string
		Request     osmanagement.InstallAllPackageUpdatesOnManagedInstanceRequest
	}

	var requests []InstallAllPackageUpdatesOnManagedInstanceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.InstallAllPackageUpdatesOnManagedInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientInstallPackageOnManagedInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "InstallPackageOnManagedInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("InstallPackageOnManagedInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "InstallPackageOnManagedInstance", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "InstallPackageOnManagedInstance")
	assert.NoError(t, err)

	type InstallPackageOnManagedInstanceRequestInfo struct {
		ContainerId string
		Request     osmanagement.InstallPackageOnManagedInstanceRequest
	}

	var requests []InstallPackageOnManagedInstanceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.InstallPackageOnManagedInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientInstallPackageUpdateOnManagedInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "InstallPackageUpdateOnManagedInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("InstallPackageUpdateOnManagedInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "InstallPackageUpdateOnManagedInstance", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "InstallPackageUpdateOnManagedInstance")
	assert.NoError(t, err)

	type InstallPackageUpdateOnManagedInstanceRequestInfo struct {
		ContainerId string
		Request     osmanagement.InstallPackageUpdateOnManagedInstanceRequest
	}

	var requests []InstallPackageUpdateOnManagedInstanceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.InstallPackageUpdateOnManagedInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientListAvailablePackagesForManagedInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "ListAvailablePackagesForManagedInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListAvailablePackagesForManagedInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "ListAvailablePackagesForManagedInstance", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "ListAvailablePackagesForManagedInstance")
	assert.NoError(t, err)

	type ListAvailablePackagesForManagedInstanceRequestInfo struct {
		ContainerId string
		Request     osmanagement.ListAvailablePackagesForManagedInstanceRequest
	}

	var requests []ListAvailablePackagesForManagedInstanceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*osmanagement.ListAvailablePackagesForManagedInstanceRequest)
				return c.ListAvailablePackagesForManagedInstance(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]osmanagement.ListAvailablePackagesForManagedInstanceResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(osmanagement.ListAvailablePackagesForManagedInstanceResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientListAvailableSoftwareSourcesForManagedInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "ListAvailableSoftwareSourcesForManagedInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListAvailableSoftwareSourcesForManagedInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "ListAvailableSoftwareSourcesForManagedInstance", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "ListAvailableSoftwareSourcesForManagedInstance")
	assert.NoError(t, err)

	type ListAvailableSoftwareSourcesForManagedInstanceRequestInfo struct {
		ContainerId string
		Request     osmanagement.ListAvailableSoftwareSourcesForManagedInstanceRequest
	}

	var requests []ListAvailableSoftwareSourcesForManagedInstanceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*osmanagement.ListAvailableSoftwareSourcesForManagedInstanceRequest)
				return c.ListAvailableSoftwareSourcesForManagedInstance(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]osmanagement.ListAvailableSoftwareSourcesForManagedInstanceResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(osmanagement.ListAvailableSoftwareSourcesForManagedInstanceResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientListAvailableUpdatesForManagedInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "ListAvailableUpdatesForManagedInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListAvailableUpdatesForManagedInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "ListAvailableUpdatesForManagedInstance", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "ListAvailableUpdatesForManagedInstance")
	assert.NoError(t, err)

	type ListAvailableUpdatesForManagedInstanceRequestInfo struct {
		ContainerId string
		Request     osmanagement.ListAvailableUpdatesForManagedInstanceRequest
	}

	var requests []ListAvailableUpdatesForManagedInstanceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*osmanagement.ListAvailableUpdatesForManagedInstanceRequest)
				return c.ListAvailableUpdatesForManagedInstance(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]osmanagement.ListAvailableUpdatesForManagedInstanceResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(osmanagement.ListAvailableUpdatesForManagedInstanceResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientListManagedInstanceGroups(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "ListManagedInstanceGroups")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListManagedInstanceGroups is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "ListManagedInstanceGroups", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "ListManagedInstanceGroups")
	assert.NoError(t, err)

	type ListManagedInstanceGroupsRequestInfo struct {
		ContainerId string
		Request     osmanagement.ListManagedInstanceGroupsRequest
	}

	var requests []ListManagedInstanceGroupsRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*osmanagement.ListManagedInstanceGroupsRequest)
				return c.ListManagedInstanceGroups(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]osmanagement.ListManagedInstanceGroupsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(osmanagement.ListManagedInstanceGroupsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientListManagedInstances(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "ListManagedInstances")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListManagedInstances is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "ListManagedInstances", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "ListManagedInstances")
	assert.NoError(t, err)

	type ListManagedInstancesRequestInfo struct {
		ContainerId string
		Request     osmanagement.ListManagedInstancesRequest
	}

	var requests []ListManagedInstancesRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*osmanagement.ListManagedInstancesRequest)
				return c.ListManagedInstances(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]osmanagement.ListManagedInstancesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(osmanagement.ListManagedInstancesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientListPackagesInstalledOnManagedInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "ListPackagesInstalledOnManagedInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListPackagesInstalledOnManagedInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "ListPackagesInstalledOnManagedInstance", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "ListPackagesInstalledOnManagedInstance")
	assert.NoError(t, err)

	type ListPackagesInstalledOnManagedInstanceRequestInfo struct {
		ContainerId string
		Request     osmanagement.ListPackagesInstalledOnManagedInstanceRequest
	}

	var requests []ListPackagesInstalledOnManagedInstanceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*osmanagement.ListPackagesInstalledOnManagedInstanceRequest)
				return c.ListPackagesInstalledOnManagedInstance(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]osmanagement.ListPackagesInstalledOnManagedInstanceResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(osmanagement.ListPackagesInstalledOnManagedInstanceResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientListScheduledJobs(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "ListScheduledJobs")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListScheduledJobs is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "ListScheduledJobs", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "ListScheduledJobs")
	assert.NoError(t, err)

	type ListScheduledJobsRequestInfo struct {
		ContainerId string
		Request     osmanagement.ListScheduledJobsRequest
	}

	var requests []ListScheduledJobsRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*osmanagement.ListScheduledJobsRequest)
				return c.ListScheduledJobs(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]osmanagement.ListScheduledJobsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(osmanagement.ListScheduledJobsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientListSoftwareSourcePackages(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "ListSoftwareSourcePackages")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListSoftwareSourcePackages is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "ListSoftwareSourcePackages", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "ListSoftwareSourcePackages")
	assert.NoError(t, err)

	type ListSoftwareSourcePackagesRequestInfo struct {
		ContainerId string
		Request     osmanagement.ListSoftwareSourcePackagesRequest
	}

	var requests []ListSoftwareSourcePackagesRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*osmanagement.ListSoftwareSourcePackagesRequest)
				return c.ListSoftwareSourcePackages(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]osmanagement.ListSoftwareSourcePackagesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(osmanagement.ListSoftwareSourcePackagesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientListSoftwareSources(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "ListSoftwareSources")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListSoftwareSources is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "ListSoftwareSources", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "ListSoftwareSources")
	assert.NoError(t, err)

	type ListSoftwareSourcesRequestInfo struct {
		ContainerId string
		Request     osmanagement.ListSoftwareSourcesRequest
	}

	var requests []ListSoftwareSourcesRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*osmanagement.ListSoftwareSourcesRequest)
				return c.ListSoftwareSources(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]osmanagement.ListSoftwareSourcesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(osmanagement.ListSoftwareSourcesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientListUpcomingScheduledJobs(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "ListUpcomingScheduledJobs")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListUpcomingScheduledJobs is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "ListUpcomingScheduledJobs", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "ListUpcomingScheduledJobs")
	assert.NoError(t, err)

	type ListUpcomingScheduledJobsRequestInfo struct {
		ContainerId string
		Request     osmanagement.ListUpcomingScheduledJobsRequest
	}

	var requests []ListUpcomingScheduledJobsRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*osmanagement.ListUpcomingScheduledJobsRequest)
				return c.ListUpcomingScheduledJobs(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]osmanagement.ListUpcomingScheduledJobsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(osmanagement.ListUpcomingScheduledJobsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientListWorkRequestErrors(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "ListWorkRequestErrors")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequestErrors is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "ListWorkRequestErrors", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "ListWorkRequestErrors")
	assert.NoError(t, err)

	type ListWorkRequestErrorsRequestInfo struct {
		ContainerId string
		Request     osmanagement.ListWorkRequestErrorsRequest
	}

	var requests []ListWorkRequestErrorsRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*osmanagement.ListWorkRequestErrorsRequest)
				return c.ListWorkRequestErrors(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]osmanagement.ListWorkRequestErrorsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(osmanagement.ListWorkRequestErrorsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientListWorkRequestLogs(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "ListWorkRequestLogs")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequestLogs is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "ListWorkRequestLogs", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "ListWorkRequestLogs")
	assert.NoError(t, err)

	type ListWorkRequestLogsRequestInfo struct {
		ContainerId string
		Request     osmanagement.ListWorkRequestLogsRequest
	}

	var requests []ListWorkRequestLogsRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*osmanagement.ListWorkRequestLogsRequest)
				return c.ListWorkRequestLogs(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]osmanagement.ListWorkRequestLogsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(osmanagement.ListWorkRequestLogsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientListWorkRequests(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "ListWorkRequests")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequests is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "ListWorkRequests", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "ListWorkRequests")
	assert.NoError(t, err)

	type ListWorkRequestsRequestInfo struct {
		ContainerId string
		Request     osmanagement.ListWorkRequestsRequest
	}

	var requests []ListWorkRequestsRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*osmanagement.ListWorkRequestsRequest)
				return c.ListWorkRequests(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]osmanagement.ListWorkRequestsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(osmanagement.ListWorkRequestsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientRemovePackageFromManagedInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "RemovePackageFromManagedInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("RemovePackageFromManagedInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "RemovePackageFromManagedInstance", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "RemovePackageFromManagedInstance")
	assert.NoError(t, err)

	type RemovePackageFromManagedInstanceRequestInfo struct {
		ContainerId string
		Request     osmanagement.RemovePackageFromManagedInstanceRequest
	}

	var requests []RemovePackageFromManagedInstanceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.RemovePackageFromManagedInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientRemovePackagesFromSoftwareSource(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "RemovePackagesFromSoftwareSource")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("RemovePackagesFromSoftwareSource is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "RemovePackagesFromSoftwareSource", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "RemovePackagesFromSoftwareSource")
	assert.NoError(t, err)

	type RemovePackagesFromSoftwareSourceRequestInfo struct {
		ContainerId string
		Request     osmanagement.RemovePackagesFromSoftwareSourceRequest
	}

	var requests []RemovePackagesFromSoftwareSourceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.RemovePackagesFromSoftwareSource(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientRunScheduledJobNow(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "RunScheduledJobNow")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("RunScheduledJobNow is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "RunScheduledJobNow", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "RunScheduledJobNow")
	assert.NoError(t, err)

	type RunScheduledJobNowRequestInfo struct {
		ContainerId string
		Request     osmanagement.RunScheduledJobNowRequest
	}

	var requests []RunScheduledJobNowRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.RunScheduledJobNow(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientSearchSoftwarePackages(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "SearchSoftwarePackages")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("SearchSoftwarePackages is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "SearchSoftwarePackages", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "SearchSoftwarePackages")
	assert.NoError(t, err)

	type SearchSoftwarePackagesRequestInfo struct {
		ContainerId string
		Request     osmanagement.SearchSoftwarePackagesRequest
	}

	var requests []SearchSoftwarePackagesRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*osmanagement.SearchSoftwarePackagesRequest)
				return c.SearchSoftwarePackages(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]osmanagement.SearchSoftwarePackagesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(osmanagement.SearchSoftwarePackagesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientSkipNextScheduledJobExecution(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "SkipNextScheduledJobExecution")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("SkipNextScheduledJobExecution is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "SkipNextScheduledJobExecution", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "SkipNextScheduledJobExecution")
	assert.NoError(t, err)

	type SkipNextScheduledJobExecutionRequestInfo struct {
		ContainerId string
		Request     osmanagement.SkipNextScheduledJobExecutionRequest
	}

	var requests []SkipNextScheduledJobExecutionRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.SkipNextScheduledJobExecution(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientUpdateManagedInstanceGroup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "UpdateManagedInstanceGroup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateManagedInstanceGroup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "UpdateManagedInstanceGroup", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "UpdateManagedInstanceGroup")
	assert.NoError(t, err)

	type UpdateManagedInstanceGroupRequestInfo struct {
		ContainerId string
		Request     osmanagement.UpdateManagedInstanceGroupRequest
	}

	var requests []UpdateManagedInstanceGroupRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.UpdateManagedInstanceGroup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientUpdateScheduledJob(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "UpdateScheduledJob")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateScheduledJob is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "UpdateScheduledJob", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "UpdateScheduledJob")
	assert.NoError(t, err)

	type UpdateScheduledJobRequestInfo struct {
		ContainerId string
		Request     osmanagement.UpdateScheduledJobRequest
	}

	var requests []UpdateScheduledJobRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.UpdateScheduledJob(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_osms_us_grp@oracle.com" jiraProject="OSMS" opsJiraProject="OSMS"
func TestOsManagementClientUpdateSoftwareSource(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("osmanagement", "UpdateSoftwareSource")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateSoftwareSource is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("osmanagement", "OsManagement", "UpdateSoftwareSource", createOsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(osmanagement.OsManagementClient)

	body, err := testClient.getRequests("osmanagement", "UpdateSoftwareSource")
	assert.NoError(t, err)

	type UpdateSoftwareSourceRequestInfo struct {
		ContainerId string
		Request     osmanagement.UpdateSoftwareSourceRequest
	}

	var requests []UpdateSoftwareSourceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.UpdateSoftwareSource(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
