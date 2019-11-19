package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/datascience"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createDataScienceClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := datascience.NewDataScienceClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientActivateModel(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "ActivateModel")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ActivateModel is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "ActivateModel", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "ActivateModel")
	assert.NoError(t, err)

	type ActivateModelRequestInfo struct {
		ContainerId string
		Request     datascience.ActivateModelRequest
	}

	var requests []ActivateModelRequestInfo
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

			response, err := c.ActivateModel(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientActivateNotebookSession(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "ActivateNotebookSession")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ActivateNotebookSession is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "ActivateNotebookSession", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "ActivateNotebookSession")
	assert.NoError(t, err)

	type ActivateNotebookSessionRequestInfo struct {
		ContainerId string
		Request     datascience.ActivateNotebookSessionRequest
	}

	var requests []ActivateNotebookSessionRequestInfo
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

			response, err := c.ActivateNotebookSession(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientCancelWorkRequest(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "CancelWorkRequest")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CancelWorkRequest is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "CancelWorkRequest", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "CancelWorkRequest")
	assert.NoError(t, err)

	type CancelWorkRequestRequestInfo struct {
		ContainerId string
		Request     datascience.CancelWorkRequestRequest
	}

	var requests []CancelWorkRequestRequestInfo
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

			response, err := c.CancelWorkRequest(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientChangeModelCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "ChangeModelCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeModelCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "ChangeModelCompartment", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "ChangeModelCompartment")
	assert.NoError(t, err)

	type ChangeModelCompartmentRequestInfo struct {
		ContainerId string
		Request     datascience.ChangeModelCompartmentRequest
	}

	var requests []ChangeModelCompartmentRequestInfo
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

			response, err := c.ChangeModelCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientChangeNotebookSessionCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "ChangeNotebookSessionCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeNotebookSessionCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "ChangeNotebookSessionCompartment", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "ChangeNotebookSessionCompartment")
	assert.NoError(t, err)

	type ChangeNotebookSessionCompartmentRequestInfo struct {
		ContainerId string
		Request     datascience.ChangeNotebookSessionCompartmentRequest
	}

	var requests []ChangeNotebookSessionCompartmentRequestInfo
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

			response, err := c.ChangeNotebookSessionCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientChangeProjectCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "ChangeProjectCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeProjectCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "ChangeProjectCompartment", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "ChangeProjectCompartment")
	assert.NoError(t, err)

	type ChangeProjectCompartmentRequestInfo struct {
		ContainerId string
		Request     datascience.ChangeProjectCompartmentRequest
	}

	var requests []ChangeProjectCompartmentRequestInfo
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

			response, err := c.ChangeProjectCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientCreateModel(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "CreateModel")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateModel is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "CreateModel", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "CreateModel")
	assert.NoError(t, err)

	type CreateModelRequestInfo struct {
		ContainerId string
		Request     datascience.CreateModelRequest
	}

	var requests []CreateModelRequestInfo
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

			response, err := c.CreateModel(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientCreateModelArtifact(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "CreateModelArtifact")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateModelArtifact is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "CreateModelArtifact", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "CreateModelArtifact")
	assert.NoError(t, err)

	type CreateModelArtifactRequestInfo struct {
		ContainerId string
		Request     datascience.CreateModelArtifactRequest
	}

	var requests []CreateModelArtifactRequestInfo
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

			response, err := c.CreateModelArtifact(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientCreateModelProvenance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "CreateModelProvenance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateModelProvenance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "CreateModelProvenance", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "CreateModelProvenance")
	assert.NoError(t, err)

	type CreateModelProvenanceRequestInfo struct {
		ContainerId string
		Request     datascience.CreateModelProvenanceRequest
	}

	var requests []CreateModelProvenanceRequestInfo
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

			response, err := c.CreateModelProvenance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientCreateNotebookSession(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "CreateNotebookSession")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateNotebookSession is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "CreateNotebookSession", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "CreateNotebookSession")
	assert.NoError(t, err)

	type CreateNotebookSessionRequestInfo struct {
		ContainerId string
		Request     datascience.CreateNotebookSessionRequest
	}

	var requests []CreateNotebookSessionRequestInfo
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

			response, err := c.CreateNotebookSession(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientCreateProject(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "CreateProject")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateProject is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "CreateProject", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "CreateProject")
	assert.NoError(t, err)

	type CreateProjectRequestInfo struct {
		ContainerId string
		Request     datascience.CreateProjectRequest
	}

	var requests []CreateProjectRequestInfo
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

			response, err := c.CreateProject(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientDeactivateModel(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "DeactivateModel")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeactivateModel is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "DeactivateModel", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "DeactivateModel")
	assert.NoError(t, err)

	type DeactivateModelRequestInfo struct {
		ContainerId string
		Request     datascience.DeactivateModelRequest
	}

	var requests []DeactivateModelRequestInfo
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

			response, err := c.DeactivateModel(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientDeactivateNotebookSession(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "DeactivateNotebookSession")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeactivateNotebookSession is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "DeactivateNotebookSession", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "DeactivateNotebookSession")
	assert.NoError(t, err)

	type DeactivateNotebookSessionRequestInfo struct {
		ContainerId string
		Request     datascience.DeactivateNotebookSessionRequest
	}

	var requests []DeactivateNotebookSessionRequestInfo
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

			response, err := c.DeactivateNotebookSession(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientDeleteModel(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "DeleteModel")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteModel is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "DeleteModel", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "DeleteModel")
	assert.NoError(t, err)

	type DeleteModelRequestInfo struct {
		ContainerId string
		Request     datascience.DeleteModelRequest
	}

	var requests []DeleteModelRequestInfo
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

			response, err := c.DeleteModel(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientDeleteNotebookSession(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "DeleteNotebookSession")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteNotebookSession is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "DeleteNotebookSession", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "DeleteNotebookSession")
	assert.NoError(t, err)

	type DeleteNotebookSessionRequestInfo struct {
		ContainerId string
		Request     datascience.DeleteNotebookSessionRequest
	}

	var requests []DeleteNotebookSessionRequestInfo
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

			response, err := c.DeleteNotebookSession(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientDeleteProject(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "DeleteProject")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteProject is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "DeleteProject", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "DeleteProject")
	assert.NoError(t, err)

	type DeleteProjectRequestInfo struct {
		ContainerId string
		Request     datascience.DeleteProjectRequest
	}

	var requests []DeleteProjectRequestInfo
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

			response, err := c.DeleteProject(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientGetModel(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "GetModel")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetModel is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "GetModel", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "GetModel")
	assert.NoError(t, err)

	type GetModelRequestInfo struct {
		ContainerId string
		Request     datascience.GetModelRequest
	}

	var requests []GetModelRequestInfo
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

			response, err := c.GetModel(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientGetModelArtifactContent(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "GetModelArtifactContent")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetModelArtifactContent is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "GetModelArtifactContent", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "GetModelArtifactContent")
	assert.NoError(t, err)

	type GetModelArtifactContentRequestInfo struct {
		ContainerId string
		Request     datascience.GetModelArtifactContentRequest
	}

	var requests []GetModelArtifactContentRequestInfo
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

			response, err := c.GetModelArtifactContent(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientGetModelProvenance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "GetModelProvenance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetModelProvenance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "GetModelProvenance", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "GetModelProvenance")
	assert.NoError(t, err)

	type GetModelProvenanceRequestInfo struct {
		ContainerId string
		Request     datascience.GetModelProvenanceRequest
	}

	var requests []GetModelProvenanceRequestInfo
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

			response, err := c.GetModelProvenance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientGetNotebookSession(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "GetNotebookSession")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetNotebookSession is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "GetNotebookSession", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "GetNotebookSession")
	assert.NoError(t, err)

	type GetNotebookSessionRequestInfo struct {
		ContainerId string
		Request     datascience.GetNotebookSessionRequest
	}

	var requests []GetNotebookSessionRequestInfo
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

			response, err := c.GetNotebookSession(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientGetProject(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "GetProject")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetProject is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "GetProject", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "GetProject")
	assert.NoError(t, err)

	type GetProjectRequestInfo struct {
		ContainerId string
		Request     datascience.GetProjectRequest
	}

	var requests []GetProjectRequestInfo
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

			response, err := c.GetProject(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientGetWorkRequest(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "GetWorkRequest")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetWorkRequest is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "GetWorkRequest", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "GetWorkRequest")
	assert.NoError(t, err)

	type GetWorkRequestRequestInfo struct {
		ContainerId string
		Request     datascience.GetWorkRequestRequest
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

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientHeadModelArtifact(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "HeadModelArtifact")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("HeadModelArtifact is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "HeadModelArtifact", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "HeadModelArtifact")
	assert.NoError(t, err)

	type HeadModelArtifactRequestInfo struct {
		ContainerId string
		Request     datascience.HeadModelArtifactRequest
	}

	var requests []HeadModelArtifactRequestInfo
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

			response, err := c.HeadModelArtifact(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientListModels(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "ListModels")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListModels is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "ListModels", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "ListModels")
	assert.NoError(t, err)

	type ListModelsRequestInfo struct {
		ContainerId string
		Request     datascience.ListModelsRequest
	}

	var requests []ListModelsRequestInfo
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
				r := req.(*datascience.ListModelsRequest)
				return c.ListModels(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datascience.ListModelsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datascience.ListModelsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientListNotebookSessionShapes(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "ListNotebookSessionShapes")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListNotebookSessionShapes is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "ListNotebookSessionShapes", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "ListNotebookSessionShapes")
	assert.NoError(t, err)

	type ListNotebookSessionShapesRequestInfo struct {
		ContainerId string
		Request     datascience.ListNotebookSessionShapesRequest
	}

	var requests []ListNotebookSessionShapesRequestInfo
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
				r := req.(*datascience.ListNotebookSessionShapesRequest)
				return c.ListNotebookSessionShapes(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datascience.ListNotebookSessionShapesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datascience.ListNotebookSessionShapesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientListNotebookSessions(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "ListNotebookSessions")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListNotebookSessions is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "ListNotebookSessions", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "ListNotebookSessions")
	assert.NoError(t, err)

	type ListNotebookSessionsRequestInfo struct {
		ContainerId string
		Request     datascience.ListNotebookSessionsRequest
	}

	var requests []ListNotebookSessionsRequestInfo
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
				r := req.(*datascience.ListNotebookSessionsRequest)
				return c.ListNotebookSessions(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datascience.ListNotebookSessionsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datascience.ListNotebookSessionsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientListProjects(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "ListProjects")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListProjects is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "ListProjects", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "ListProjects")
	assert.NoError(t, err)

	type ListProjectsRequestInfo struct {
		ContainerId string
		Request     datascience.ListProjectsRequest
	}

	var requests []ListProjectsRequestInfo
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
				r := req.(*datascience.ListProjectsRequest)
				return c.ListProjects(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datascience.ListProjectsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datascience.ListProjectsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientListWorkRequestErrors(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "ListWorkRequestErrors")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequestErrors is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "ListWorkRequestErrors", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "ListWorkRequestErrors")
	assert.NoError(t, err)

	type ListWorkRequestErrorsRequestInfo struct {
		ContainerId string
		Request     datascience.ListWorkRequestErrorsRequest
	}

	var requests []ListWorkRequestErrorsRequestInfo
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

			response, err := c.ListWorkRequestErrors(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientListWorkRequestLogs(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "ListWorkRequestLogs")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequestLogs is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "ListWorkRequestLogs", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "ListWorkRequestLogs")
	assert.NoError(t, err)

	type ListWorkRequestLogsRequestInfo struct {
		ContainerId string
		Request     datascience.ListWorkRequestLogsRequest
	}

	var requests []ListWorkRequestLogsRequestInfo
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

			response, err := c.ListWorkRequestLogs(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientListWorkRequests(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "ListWorkRequests")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequests is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "ListWorkRequests", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "ListWorkRequests")
	assert.NoError(t, err)

	type ListWorkRequestsRequestInfo struct {
		ContainerId string
		Request     datascience.ListWorkRequestsRequest
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
				r := req.(*datascience.ListWorkRequestsRequest)
				return c.ListWorkRequests(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datascience.ListWorkRequestsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datascience.ListWorkRequestsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientUpdateModel(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "UpdateModel")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateModel is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "UpdateModel", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "UpdateModel")
	assert.NoError(t, err)

	type UpdateModelRequestInfo struct {
		ContainerId string
		Request     datascience.UpdateModelRequest
	}

	var requests []UpdateModelRequestInfo
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

			response, err := c.UpdateModel(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientUpdateModelProvenance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "UpdateModelProvenance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateModelProvenance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "UpdateModelProvenance", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "UpdateModelProvenance")
	assert.NoError(t, err)

	type UpdateModelProvenanceRequestInfo struct {
		ContainerId string
		Request     datascience.UpdateModelProvenanceRequest
	}

	var requests []UpdateModelProvenanceRequestInfo
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

			response, err := c.UpdateModelProvenance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientUpdateNotebookSession(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "UpdateNotebookSession")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateNotebookSession is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "UpdateNotebookSession", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "UpdateNotebookSession")
	assert.NoError(t, err)

	type UpdateNotebookSessionRequestInfo struct {
		ContainerId string
		Request     datascience.UpdateNotebookSessionRequest
	}

	var requests []UpdateNotebookSessionRequestInfo
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

			response, err := c.UpdateNotebookSession(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datascience_grp@oracle.com" jiraProject="ODSC" opsJiraProject="ODSC"
func TestDataScienceClientUpdateProject(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datascience", "UpdateProject")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateProject is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datascience", "DataScience", "UpdateProject", createDataScienceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datascience.DataScienceClient)

	body, err := testClient.getRequests("datascience", "UpdateProject")
	assert.NoError(t, err)

	type UpdateProjectRequestInfo struct {
		ContainerId string
		Request     datascience.UpdateProjectRequest
	}

	var requests []UpdateProjectRequestInfo
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

			response, err := c.UpdateProject(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
