package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/datacatalog"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createDataCatalogClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := datacatalog.NewDataCatalogClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientCreateAttribute(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "CreateAttribute")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateAttribute is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "CreateAttribute", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "CreateAttribute")
	assert.NoError(t, err)

	type CreateAttributeRequestInfo struct {
		ContainerId string
		Request     datacatalog.CreateAttributeRequest
	}

	var requests []CreateAttributeRequestInfo
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

			response, err := c.CreateAttribute(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientCreateAttributeTag(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "CreateAttributeTag")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateAttributeTag is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "CreateAttributeTag", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "CreateAttributeTag")
	assert.NoError(t, err)

	type CreateAttributeTagRequestInfo struct {
		ContainerId string
		Request     datacatalog.CreateAttributeTagRequest
	}

	var requests []CreateAttributeTagRequestInfo
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

			response, err := c.CreateAttributeTag(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientCreateCatalog(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "CreateCatalog")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateCatalog is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "CreateCatalog", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "CreateCatalog")
	assert.NoError(t, err)

	type CreateCatalogRequestInfo struct {
		ContainerId string
		Request     datacatalog.CreateCatalogRequest
	}

	var requests []CreateCatalogRequestInfo
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

			response, err := c.CreateCatalog(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientCreateConnection(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "CreateConnection")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateConnection is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "CreateConnection", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "CreateConnection")
	assert.NoError(t, err)

	type CreateConnectionRequestInfo struct {
		ContainerId string
		Request     datacatalog.CreateConnectionRequest
	}

	var requests []CreateConnectionRequestInfo
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

			response, err := c.CreateConnection(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientCreateDataAsset(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "CreateDataAsset")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateDataAsset is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "CreateDataAsset", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "CreateDataAsset")
	assert.NoError(t, err)

	type CreateDataAssetRequestInfo struct {
		ContainerId string
		Request     datacatalog.CreateDataAssetRequest
	}

	var requests []CreateDataAssetRequestInfo
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

			response, err := c.CreateDataAsset(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientCreateDataAssetTag(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "CreateDataAssetTag")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateDataAssetTag is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "CreateDataAssetTag", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "CreateDataAssetTag")
	assert.NoError(t, err)

	type CreateDataAssetTagRequestInfo struct {
		ContainerId string
		Request     datacatalog.CreateDataAssetTagRequest
	}

	var requests []CreateDataAssetTagRequestInfo
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

			response, err := c.CreateDataAssetTag(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientCreateEntity(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "CreateEntity")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateEntity is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "CreateEntity", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "CreateEntity")
	assert.NoError(t, err)

	type CreateEntityRequestInfo struct {
		ContainerId string
		Request     datacatalog.CreateEntityRequest
	}

	var requests []CreateEntityRequestInfo
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

			response, err := c.CreateEntity(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientCreateEntityTag(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "CreateEntityTag")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateEntityTag is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "CreateEntityTag", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "CreateEntityTag")
	assert.NoError(t, err)

	type CreateEntityTagRequestInfo struct {
		ContainerId string
		Request     datacatalog.CreateEntityTagRequest
	}

	var requests []CreateEntityTagRequestInfo
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

			response, err := c.CreateEntityTag(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientCreateFolder(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "CreateFolder")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateFolder is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "CreateFolder", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "CreateFolder")
	assert.NoError(t, err)

	type CreateFolderRequestInfo struct {
		ContainerId string
		Request     datacatalog.CreateFolderRequest
	}

	var requests []CreateFolderRequestInfo
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

			response, err := c.CreateFolder(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientCreateFolderTag(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "CreateFolderTag")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateFolderTag is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "CreateFolderTag", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "CreateFolderTag")
	assert.NoError(t, err)

	type CreateFolderTagRequestInfo struct {
		ContainerId string
		Request     datacatalog.CreateFolderTagRequest
	}

	var requests []CreateFolderTagRequestInfo
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

			response, err := c.CreateFolderTag(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientCreateGlossary(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "CreateGlossary")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateGlossary is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "CreateGlossary", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "CreateGlossary")
	assert.NoError(t, err)

	type CreateGlossaryRequestInfo struct {
		ContainerId string
		Request     datacatalog.CreateGlossaryRequest
	}

	var requests []CreateGlossaryRequestInfo
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

			response, err := c.CreateGlossary(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientCreateJob(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "CreateJob")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateJob is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "CreateJob", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "CreateJob")
	assert.NoError(t, err)

	type CreateJobRequestInfo struct {
		ContainerId string
		Request     datacatalog.CreateJobRequest
	}

	var requests []CreateJobRequestInfo
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

			response, err := c.CreateJob(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientCreateJobDefinition(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "CreateJobDefinition")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateJobDefinition is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "CreateJobDefinition", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "CreateJobDefinition")
	assert.NoError(t, err)

	type CreateJobDefinitionRequestInfo struct {
		ContainerId string
		Request     datacatalog.CreateJobDefinitionRequest
	}

	var requests []CreateJobDefinitionRequestInfo
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

			response, err := c.CreateJobDefinition(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientCreateJobExecution(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "CreateJobExecution")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateJobExecution is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "CreateJobExecution", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "CreateJobExecution")
	assert.NoError(t, err)

	type CreateJobExecutionRequestInfo struct {
		ContainerId string
		Request     datacatalog.CreateJobExecutionRequest
	}

	var requests []CreateJobExecutionRequestInfo
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

			response, err := c.CreateJobExecution(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientCreateTerm(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "CreateTerm")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateTerm is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "CreateTerm", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "CreateTerm")
	assert.NoError(t, err)

	type CreateTermRequestInfo struct {
		ContainerId string
		Request     datacatalog.CreateTermRequest
	}

	var requests []CreateTermRequestInfo
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

			response, err := c.CreateTerm(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientCreateTermRelationship(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "CreateTermRelationship")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateTermRelationship is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "CreateTermRelationship", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "CreateTermRelationship")
	assert.NoError(t, err)

	type CreateTermRelationshipRequestInfo struct {
		ContainerId string
		Request     datacatalog.CreateTermRelationshipRequest
	}

	var requests []CreateTermRelationshipRequestInfo
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

			response, err := c.CreateTermRelationship(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientDeleteAttribute(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "DeleteAttribute")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteAttribute is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "DeleteAttribute", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "DeleteAttribute")
	assert.NoError(t, err)

	type DeleteAttributeRequestInfo struct {
		ContainerId string
		Request     datacatalog.DeleteAttributeRequest
	}

	var requests []DeleteAttributeRequestInfo
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

			response, err := c.DeleteAttribute(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientDeleteAttributeTag(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "DeleteAttributeTag")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteAttributeTag is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "DeleteAttributeTag", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "DeleteAttributeTag")
	assert.NoError(t, err)

	type DeleteAttributeTagRequestInfo struct {
		ContainerId string
		Request     datacatalog.DeleteAttributeTagRequest
	}

	var requests []DeleteAttributeTagRequestInfo
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

			response, err := c.DeleteAttributeTag(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientDeleteCatalog(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "DeleteCatalog")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteCatalog is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "DeleteCatalog", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "DeleteCatalog")
	assert.NoError(t, err)

	type DeleteCatalogRequestInfo struct {
		ContainerId string
		Request     datacatalog.DeleteCatalogRequest
	}

	var requests []DeleteCatalogRequestInfo
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

			response, err := c.DeleteCatalog(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientDeleteConnection(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "DeleteConnection")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteConnection is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "DeleteConnection", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "DeleteConnection")
	assert.NoError(t, err)

	type DeleteConnectionRequestInfo struct {
		ContainerId string
		Request     datacatalog.DeleteConnectionRequest
	}

	var requests []DeleteConnectionRequestInfo
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

			response, err := c.DeleteConnection(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientDeleteDataAsset(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "DeleteDataAsset")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteDataAsset is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "DeleteDataAsset", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "DeleteDataAsset")
	assert.NoError(t, err)

	type DeleteDataAssetRequestInfo struct {
		ContainerId string
		Request     datacatalog.DeleteDataAssetRequest
	}

	var requests []DeleteDataAssetRequestInfo
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

			response, err := c.DeleteDataAsset(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientDeleteDataAssetTag(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "DeleteDataAssetTag")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteDataAssetTag is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "DeleteDataAssetTag", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "DeleteDataAssetTag")
	assert.NoError(t, err)

	type DeleteDataAssetTagRequestInfo struct {
		ContainerId string
		Request     datacatalog.DeleteDataAssetTagRequest
	}

	var requests []DeleteDataAssetTagRequestInfo
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

			response, err := c.DeleteDataAssetTag(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientDeleteEntity(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "DeleteEntity")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteEntity is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "DeleteEntity", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "DeleteEntity")
	assert.NoError(t, err)

	type DeleteEntityRequestInfo struct {
		ContainerId string
		Request     datacatalog.DeleteEntityRequest
	}

	var requests []DeleteEntityRequestInfo
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

			response, err := c.DeleteEntity(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientDeleteEntityTag(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "DeleteEntityTag")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteEntityTag is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "DeleteEntityTag", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "DeleteEntityTag")
	assert.NoError(t, err)

	type DeleteEntityTagRequestInfo struct {
		ContainerId string
		Request     datacatalog.DeleteEntityTagRequest
	}

	var requests []DeleteEntityTagRequestInfo
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

			response, err := c.DeleteEntityTag(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientDeleteFolder(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "DeleteFolder")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteFolder is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "DeleteFolder", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "DeleteFolder")
	assert.NoError(t, err)

	type DeleteFolderRequestInfo struct {
		ContainerId string
		Request     datacatalog.DeleteFolderRequest
	}

	var requests []DeleteFolderRequestInfo
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

			response, err := c.DeleteFolder(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientDeleteFolderTag(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "DeleteFolderTag")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteFolderTag is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "DeleteFolderTag", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "DeleteFolderTag")
	assert.NoError(t, err)

	type DeleteFolderTagRequestInfo struct {
		ContainerId string
		Request     datacatalog.DeleteFolderTagRequest
	}

	var requests []DeleteFolderTagRequestInfo
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

			response, err := c.DeleteFolderTag(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientDeleteGlossary(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "DeleteGlossary")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteGlossary is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "DeleteGlossary", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "DeleteGlossary")
	assert.NoError(t, err)

	type DeleteGlossaryRequestInfo struct {
		ContainerId string
		Request     datacatalog.DeleteGlossaryRequest
	}

	var requests []DeleteGlossaryRequestInfo
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

			response, err := c.DeleteGlossary(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientDeleteJob(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "DeleteJob")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteJob is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "DeleteJob", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "DeleteJob")
	assert.NoError(t, err)

	type DeleteJobRequestInfo struct {
		ContainerId string
		Request     datacatalog.DeleteJobRequest
	}

	var requests []DeleteJobRequestInfo
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

			response, err := c.DeleteJob(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientDeleteJobDefinition(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "DeleteJobDefinition")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteJobDefinition is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "DeleteJobDefinition", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "DeleteJobDefinition")
	assert.NoError(t, err)

	type DeleteJobDefinitionRequestInfo struct {
		ContainerId string
		Request     datacatalog.DeleteJobDefinitionRequest
	}

	var requests []DeleteJobDefinitionRequestInfo
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

			response, err := c.DeleteJobDefinition(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientDeleteTerm(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "DeleteTerm")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteTerm is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "DeleteTerm", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "DeleteTerm")
	assert.NoError(t, err)

	type DeleteTermRequestInfo struct {
		ContainerId string
		Request     datacatalog.DeleteTermRequest
	}

	var requests []DeleteTermRequestInfo
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

			response, err := c.DeleteTerm(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientDeleteTermRelationship(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "DeleteTermRelationship")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteTermRelationship is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "DeleteTermRelationship", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "DeleteTermRelationship")
	assert.NoError(t, err)

	type DeleteTermRelationshipRequestInfo struct {
		ContainerId string
		Request     datacatalog.DeleteTermRelationshipRequest
	}

	var requests []DeleteTermRelationshipRequestInfo
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

			response, err := c.DeleteTermRelationship(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientExpandTreeForGlossary(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ExpandTreeForGlossary")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ExpandTreeForGlossary is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ExpandTreeForGlossary", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ExpandTreeForGlossary")
	assert.NoError(t, err)

	type ExpandTreeForGlossaryRequestInfo struct {
		ContainerId string
		Request     datacatalog.ExpandTreeForGlossaryRequest
	}

	var requests []ExpandTreeForGlossaryRequestInfo
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

			response, err := c.ExpandTreeForGlossary(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientExportGlossary(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ExportGlossary")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ExportGlossary is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ExportGlossary", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ExportGlossary")
	assert.NoError(t, err)

	type ExportGlossaryRequestInfo struct {
		ContainerId string
		Request     datacatalog.ExportGlossaryRequest
	}

	var requests []ExportGlossaryRequestInfo
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

			response, err := c.ExportGlossary(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientGetAttribute(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "GetAttribute")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetAttribute is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "GetAttribute", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "GetAttribute")
	assert.NoError(t, err)

	type GetAttributeRequestInfo struct {
		ContainerId string
		Request     datacatalog.GetAttributeRequest
	}

	var requests []GetAttributeRequestInfo
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

			response, err := c.GetAttribute(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientGetAttributeTag(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "GetAttributeTag")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetAttributeTag is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "GetAttributeTag", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "GetAttributeTag")
	assert.NoError(t, err)

	type GetAttributeTagRequestInfo struct {
		ContainerId string
		Request     datacatalog.GetAttributeTagRequest
	}

	var requests []GetAttributeTagRequestInfo
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

			response, err := c.GetAttributeTag(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientGetCatalog(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "GetCatalog")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetCatalog is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "GetCatalog", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "GetCatalog")
	assert.NoError(t, err)

	type GetCatalogRequestInfo struct {
		ContainerId string
		Request     datacatalog.GetCatalogRequest
	}

	var requests []GetCatalogRequestInfo
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

			response, err := c.GetCatalog(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientGetConnection(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "GetConnection")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetConnection is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "GetConnection", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "GetConnection")
	assert.NoError(t, err)

	type GetConnectionRequestInfo struct {
		ContainerId string
		Request     datacatalog.GetConnectionRequest
	}

	var requests []GetConnectionRequestInfo
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

			response, err := c.GetConnection(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientGetDataAsset(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "GetDataAsset")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetDataAsset is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "GetDataAsset", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "GetDataAsset")
	assert.NoError(t, err)

	type GetDataAssetRequestInfo struct {
		ContainerId string
		Request     datacatalog.GetDataAssetRequest
	}

	var requests []GetDataAssetRequestInfo
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

			response, err := c.GetDataAsset(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientGetDataAssetTag(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "GetDataAssetTag")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetDataAssetTag is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "GetDataAssetTag", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "GetDataAssetTag")
	assert.NoError(t, err)

	type GetDataAssetTagRequestInfo struct {
		ContainerId string
		Request     datacatalog.GetDataAssetTagRequest
	}

	var requests []GetDataAssetTagRequestInfo
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

			response, err := c.GetDataAssetTag(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientGetEntity(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "GetEntity")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetEntity is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "GetEntity", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "GetEntity")
	assert.NoError(t, err)

	type GetEntityRequestInfo struct {
		ContainerId string
		Request     datacatalog.GetEntityRequest
	}

	var requests []GetEntityRequestInfo
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

			response, err := c.GetEntity(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientGetEntityTag(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "GetEntityTag")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetEntityTag is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "GetEntityTag", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "GetEntityTag")
	assert.NoError(t, err)

	type GetEntityTagRequestInfo struct {
		ContainerId string
		Request     datacatalog.GetEntityTagRequest
	}

	var requests []GetEntityTagRequestInfo
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

			response, err := c.GetEntityTag(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientGetFolder(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "GetFolder")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetFolder is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "GetFolder", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "GetFolder")
	assert.NoError(t, err)

	type GetFolderRequestInfo struct {
		ContainerId string
		Request     datacatalog.GetFolderRequest
	}

	var requests []GetFolderRequestInfo
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

			response, err := c.GetFolder(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientGetFolderTag(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "GetFolderTag")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetFolderTag is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "GetFolderTag", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "GetFolderTag")
	assert.NoError(t, err)

	type GetFolderTagRequestInfo struct {
		ContainerId string
		Request     datacatalog.GetFolderTagRequest
	}

	var requests []GetFolderTagRequestInfo
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

			response, err := c.GetFolderTag(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientGetGlossary(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "GetGlossary")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetGlossary is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "GetGlossary", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "GetGlossary")
	assert.NoError(t, err)

	type GetGlossaryRequestInfo struct {
		ContainerId string
		Request     datacatalog.GetGlossaryRequest
	}

	var requests []GetGlossaryRequestInfo
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

			response, err := c.GetGlossary(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientGetJob(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "GetJob")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetJob is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "GetJob", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "GetJob")
	assert.NoError(t, err)

	type GetJobRequestInfo struct {
		ContainerId string
		Request     datacatalog.GetJobRequest
	}

	var requests []GetJobRequestInfo
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

			response, err := c.GetJob(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientGetJobDefinition(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "GetJobDefinition")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetJobDefinition is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "GetJobDefinition", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "GetJobDefinition")
	assert.NoError(t, err)

	type GetJobDefinitionRequestInfo struct {
		ContainerId string
		Request     datacatalog.GetJobDefinitionRequest
	}

	var requests []GetJobDefinitionRequestInfo
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

			response, err := c.GetJobDefinition(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientGetJobExecution(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "GetJobExecution")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetJobExecution is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "GetJobExecution", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "GetJobExecution")
	assert.NoError(t, err)

	type GetJobExecutionRequestInfo struct {
		ContainerId string
		Request     datacatalog.GetJobExecutionRequest
	}

	var requests []GetJobExecutionRequestInfo
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

			response, err := c.GetJobExecution(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientGetJobLog(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "GetJobLog")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetJobLog is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "GetJobLog", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "GetJobLog")
	assert.NoError(t, err)

	type GetJobLogRequestInfo struct {
		ContainerId string
		Request     datacatalog.GetJobLogRequest
	}

	var requests []GetJobLogRequestInfo
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

			response, err := c.GetJobLog(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientGetJobMetrics(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "GetJobMetrics")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetJobMetrics is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "GetJobMetrics", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "GetJobMetrics")
	assert.NoError(t, err)

	type GetJobMetricsRequestInfo struct {
		ContainerId string
		Request     datacatalog.GetJobMetricsRequest
	}

	var requests []GetJobMetricsRequestInfo
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

			response, err := c.GetJobMetrics(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientGetTerm(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "GetTerm")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetTerm is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "GetTerm", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "GetTerm")
	assert.NoError(t, err)

	type GetTermRequestInfo struct {
		ContainerId string
		Request     datacatalog.GetTermRequest
	}

	var requests []GetTermRequestInfo
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

			response, err := c.GetTerm(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientGetTermRelationship(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "GetTermRelationship")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetTermRelationship is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "GetTermRelationship", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "GetTermRelationship")
	assert.NoError(t, err)

	type GetTermRelationshipRequestInfo struct {
		ContainerId string
		Request     datacatalog.GetTermRelationshipRequest
	}

	var requests []GetTermRelationshipRequestInfo
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

			response, err := c.GetTermRelationship(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientGetType(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "GetType")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetType is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "GetType", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "GetType")
	assert.NoError(t, err)

	type GetTypeRequestInfo struct {
		ContainerId string
		Request     datacatalog.GetTypeRequest
	}

	var requests []GetTypeRequestInfo
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

			response, err := c.GetType(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientGetWorkRequest(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "GetWorkRequest")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetWorkRequest is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "GetWorkRequest", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "GetWorkRequest")
	assert.NoError(t, err)

	type GetWorkRequestRequestInfo struct {
		ContainerId string
		Request     datacatalog.GetWorkRequestRequest
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

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientImportConnection(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ImportConnection")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ImportConnection is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ImportConnection", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ImportConnection")
	assert.NoError(t, err)

	type ImportConnectionRequestInfo struct {
		ContainerId string
		Request     datacatalog.ImportConnectionRequest
	}

	var requests []ImportConnectionRequestInfo
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

			response, err := c.ImportConnection(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientImportGlossary(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ImportGlossary")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ImportGlossary is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ImportGlossary", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ImportGlossary")
	assert.NoError(t, err)

	type ImportGlossaryRequestInfo struct {
		ContainerId string
		Request     datacatalog.ImportGlossaryRequest
	}

	var requests []ImportGlossaryRequestInfo
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

			response, err := c.ImportGlossary(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListAttributeTags(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListAttributeTags")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListAttributeTags is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListAttributeTags", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListAttributeTags")
	assert.NoError(t, err)

	type ListAttributeTagsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListAttributeTagsRequest
	}

	var requests []ListAttributeTagsRequestInfo
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
				r := req.(*datacatalog.ListAttributeTagsRequest)
				return c.ListAttributeTags(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListAttributeTagsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListAttributeTagsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListAttributes(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListAttributes")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListAttributes is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListAttributes", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListAttributes")
	assert.NoError(t, err)

	type ListAttributesRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListAttributesRequest
	}

	var requests []ListAttributesRequestInfo
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
				r := req.(*datacatalog.ListAttributesRequest)
				return c.ListAttributes(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListAttributesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListAttributesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListCatalogPermissions(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListCatalogPermissions")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListCatalogPermissions is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListCatalogPermissions", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListCatalogPermissions")
	assert.NoError(t, err)

	type ListCatalogPermissionsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListCatalogPermissionsRequest
	}

	var requests []ListCatalogPermissionsRequestInfo
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
				r := req.(*datacatalog.ListCatalogPermissionsRequest)
				return c.ListCatalogPermissions(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListCatalogPermissionsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListCatalogPermissionsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListCatalogs(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListCatalogs")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListCatalogs is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListCatalogs", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListCatalogs")
	assert.NoError(t, err)

	type ListCatalogsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListCatalogsRequest
	}

	var requests []ListCatalogsRequestInfo
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
				r := req.(*datacatalog.ListCatalogsRequest)
				return c.ListCatalogs(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListCatalogsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListCatalogsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListConnections(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListConnections")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListConnections is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListConnections", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListConnections")
	assert.NoError(t, err)

	type ListConnectionsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListConnectionsRequest
	}

	var requests []ListConnectionsRequestInfo
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
				r := req.(*datacatalog.ListConnectionsRequest)
				return c.ListConnections(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListConnectionsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListConnectionsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListDataAssetPermissions(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListDataAssetPermissions")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListDataAssetPermissions is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListDataAssetPermissions", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListDataAssetPermissions")
	assert.NoError(t, err)

	type ListDataAssetPermissionsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListDataAssetPermissionsRequest
	}

	var requests []ListDataAssetPermissionsRequestInfo
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
				r := req.(*datacatalog.ListDataAssetPermissionsRequest)
				return c.ListDataAssetPermissions(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListDataAssetPermissionsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListDataAssetPermissionsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListDataAssetTags(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListDataAssetTags")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListDataAssetTags is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListDataAssetTags", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListDataAssetTags")
	assert.NoError(t, err)

	type ListDataAssetTagsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListDataAssetTagsRequest
	}

	var requests []ListDataAssetTagsRequestInfo
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
				r := req.(*datacatalog.ListDataAssetTagsRequest)
				return c.ListDataAssetTags(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListDataAssetTagsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListDataAssetTagsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListDataAssets(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListDataAssets")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListDataAssets is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListDataAssets", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListDataAssets")
	assert.NoError(t, err)

	type ListDataAssetsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListDataAssetsRequest
	}

	var requests []ListDataAssetsRequestInfo
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
				r := req.(*datacatalog.ListDataAssetsRequest)
				return c.ListDataAssets(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListDataAssetsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListDataAssetsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListEntities(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListEntities")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListEntities is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListEntities", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListEntities")
	assert.NoError(t, err)

	type ListEntitiesRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListEntitiesRequest
	}

	var requests []ListEntitiesRequestInfo
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
				r := req.(*datacatalog.ListEntitiesRequest)
				return c.ListEntities(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListEntitiesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListEntitiesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListEntityTags(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListEntityTags")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListEntityTags is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListEntityTags", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListEntityTags")
	assert.NoError(t, err)

	type ListEntityTagsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListEntityTagsRequest
	}

	var requests []ListEntityTagsRequestInfo
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
				r := req.(*datacatalog.ListEntityTagsRequest)
				return c.ListEntityTags(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListEntityTagsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListEntityTagsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListFolderTags(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListFolderTags")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListFolderTags is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListFolderTags", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListFolderTags")
	assert.NoError(t, err)

	type ListFolderTagsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListFolderTagsRequest
	}

	var requests []ListFolderTagsRequestInfo
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
				r := req.(*datacatalog.ListFolderTagsRequest)
				return c.ListFolderTags(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListFolderTagsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListFolderTagsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListFolders(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListFolders")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListFolders is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListFolders", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListFolders")
	assert.NoError(t, err)

	type ListFoldersRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListFoldersRequest
	}

	var requests []ListFoldersRequestInfo
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
				r := req.(*datacatalog.ListFoldersRequest)
				return c.ListFolders(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListFoldersResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListFoldersResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListGlossaries(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListGlossaries")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListGlossaries is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListGlossaries", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListGlossaries")
	assert.NoError(t, err)

	type ListGlossariesRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListGlossariesRequest
	}

	var requests []ListGlossariesRequestInfo
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
				r := req.(*datacatalog.ListGlossariesRequest)
				return c.ListGlossaries(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListGlossariesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListGlossariesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListGlossaryPermissions(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListGlossaryPermissions")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListGlossaryPermissions is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListGlossaryPermissions", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListGlossaryPermissions")
	assert.NoError(t, err)

	type ListGlossaryPermissionsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListGlossaryPermissionsRequest
	}

	var requests []ListGlossaryPermissionsRequestInfo
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
				r := req.(*datacatalog.ListGlossaryPermissionsRequest)
				return c.ListGlossaryPermissions(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListGlossaryPermissionsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListGlossaryPermissionsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListGlossaryTerms(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListGlossaryTerms")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListGlossaryTerms is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListGlossaryTerms", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListGlossaryTerms")
	assert.NoError(t, err)

	type ListGlossaryTermsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListGlossaryTermsRequest
	}

	var requests []ListGlossaryTermsRequestInfo
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
				r := req.(*datacatalog.ListGlossaryTermsRequest)
				return c.ListGlossaryTerms(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListGlossaryTermsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListGlossaryTermsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListJobDefinitionPermissions(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListJobDefinitionPermissions")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListJobDefinitionPermissions is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListJobDefinitionPermissions", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListJobDefinitionPermissions")
	assert.NoError(t, err)

	type ListJobDefinitionPermissionsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListJobDefinitionPermissionsRequest
	}

	var requests []ListJobDefinitionPermissionsRequestInfo
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
				r := req.(*datacatalog.ListJobDefinitionPermissionsRequest)
				return c.ListJobDefinitionPermissions(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListJobDefinitionPermissionsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListJobDefinitionPermissionsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListJobDefinitions(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListJobDefinitions")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListJobDefinitions is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListJobDefinitions", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListJobDefinitions")
	assert.NoError(t, err)

	type ListJobDefinitionsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListJobDefinitionsRequest
	}

	var requests []ListJobDefinitionsRequestInfo
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
				r := req.(*datacatalog.ListJobDefinitionsRequest)
				return c.ListJobDefinitions(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListJobDefinitionsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListJobDefinitionsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListJobExecutions(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListJobExecutions")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListJobExecutions is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListJobExecutions", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListJobExecutions")
	assert.NoError(t, err)

	type ListJobExecutionsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListJobExecutionsRequest
	}

	var requests []ListJobExecutionsRequestInfo
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
				r := req.(*datacatalog.ListJobExecutionsRequest)
				return c.ListJobExecutions(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListJobExecutionsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListJobExecutionsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListJobLogs(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListJobLogs")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListJobLogs is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListJobLogs", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListJobLogs")
	assert.NoError(t, err)

	type ListJobLogsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListJobLogsRequest
	}

	var requests []ListJobLogsRequestInfo
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
				r := req.(*datacatalog.ListJobLogsRequest)
				return c.ListJobLogs(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListJobLogsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListJobLogsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListJobMetrics(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListJobMetrics")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListJobMetrics is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListJobMetrics", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListJobMetrics")
	assert.NoError(t, err)

	type ListJobMetricsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListJobMetricsRequest
	}

	var requests []ListJobMetricsRequestInfo
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
				r := req.(*datacatalog.ListJobMetricsRequest)
				return c.ListJobMetrics(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListJobMetricsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListJobMetricsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListJobs(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListJobs")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListJobs is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListJobs", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListJobs")
	assert.NoError(t, err)

	type ListJobsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListJobsRequest
	}

	var requests []ListJobsRequestInfo
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
				r := req.(*datacatalog.ListJobsRequest)
				return c.ListJobs(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListJobsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListJobsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListTags(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListTags")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListTags is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListTags", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListTags")
	assert.NoError(t, err)

	type ListTagsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListTagsRequest
	}

	var requests []ListTagsRequestInfo
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
				r := req.(*datacatalog.ListTagsRequest)
				return c.ListTags(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListTagsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListTagsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListTermRelationships(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListTermRelationships")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListTermRelationships is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListTermRelationships", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListTermRelationships")
	assert.NoError(t, err)

	type ListTermRelationshipsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListTermRelationshipsRequest
	}

	var requests []ListTermRelationshipsRequestInfo
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
				r := req.(*datacatalog.ListTermRelationshipsRequest)
				return c.ListTermRelationships(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListTermRelationshipsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListTermRelationshipsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListTypes(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListTypes")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListTypes is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListTypes", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListTypes")
	assert.NoError(t, err)

	type ListTypesRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListTypesRequest
	}

	var requests []ListTypesRequestInfo
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
				r := req.(*datacatalog.ListTypesRequest)
				return c.ListTypes(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListTypesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListTypesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListWorkRequestErrors(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListWorkRequestErrors")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequestErrors is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListWorkRequestErrors", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListWorkRequestErrors")
	assert.NoError(t, err)

	type ListWorkRequestErrorsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListWorkRequestErrorsRequest
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
				r := req.(*datacatalog.ListWorkRequestErrorsRequest)
				return c.ListWorkRequestErrors(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListWorkRequestErrorsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListWorkRequestErrorsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListWorkRequestLogs(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListWorkRequestLogs")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequestLogs is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListWorkRequestLogs", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListWorkRequestLogs")
	assert.NoError(t, err)

	type ListWorkRequestLogsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListWorkRequestLogsRequest
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
				r := req.(*datacatalog.ListWorkRequestLogsRequest)
				return c.ListWorkRequestLogs(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListWorkRequestLogsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListWorkRequestLogsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientListWorkRequests(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ListWorkRequests")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequests is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ListWorkRequests", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ListWorkRequests")
	assert.NoError(t, err)

	type ListWorkRequestsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ListWorkRequestsRequest
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
				r := req.(*datacatalog.ListWorkRequestsRequest)
				return c.ListWorkRequests(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ListWorkRequestsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ListWorkRequestsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientObjectStats(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ObjectStats")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ObjectStats is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ObjectStats", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ObjectStats")
	assert.NoError(t, err)

	type ObjectStatsRequestInfo struct {
		ContainerId string
		Request     datacatalog.ObjectStatsRequest
	}

	var requests []ObjectStatsRequestInfo
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
				r := req.(*datacatalog.ObjectStatsRequest)
				return c.ObjectStats(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.ObjectStatsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.ObjectStatsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientParseConnection(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ParseConnection")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ParseConnection is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ParseConnection", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ParseConnection")
	assert.NoError(t, err)

	type ParseConnectionRequestInfo struct {
		ContainerId string
		Request     datacatalog.ParseConnectionRequest
	}

	var requests []ParseConnectionRequestInfo
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

			response, err := c.ParseConnection(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientSearchCriteria(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "SearchCriteria")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("SearchCriteria is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "SearchCriteria", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "SearchCriteria")
	assert.NoError(t, err)

	type SearchCriteriaRequestInfo struct {
		ContainerId string
		Request     datacatalog.SearchCriteriaRequest
	}

	var requests []SearchCriteriaRequestInfo
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
				r := req.(*datacatalog.SearchCriteriaRequest)
				return c.SearchCriteria(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.SearchCriteriaResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.SearchCriteriaResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientTestConnection(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "TestConnection")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("TestConnection is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "TestConnection", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "TestConnection")
	assert.NoError(t, err)

	type TestConnectionRequestInfo struct {
		ContainerId string
		Request     datacatalog.TestConnectionRequest
	}

	var requests []TestConnectionRequestInfo
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

			response, err := c.TestConnection(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientUpdateAttribute(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "UpdateAttribute")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateAttribute is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "UpdateAttribute", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "UpdateAttribute")
	assert.NoError(t, err)

	type UpdateAttributeRequestInfo struct {
		ContainerId string
		Request     datacatalog.UpdateAttributeRequest
	}

	var requests []UpdateAttributeRequestInfo
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

			response, err := c.UpdateAttribute(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientUpdateCatalog(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "UpdateCatalog")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateCatalog is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "UpdateCatalog", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "UpdateCatalog")
	assert.NoError(t, err)

	type UpdateCatalogRequestInfo struct {
		ContainerId string
		Request     datacatalog.UpdateCatalogRequest
	}

	var requests []UpdateCatalogRequestInfo
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

			response, err := c.UpdateCatalog(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientUpdateConnection(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "UpdateConnection")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateConnection is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "UpdateConnection", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "UpdateConnection")
	assert.NoError(t, err)

	type UpdateConnectionRequestInfo struct {
		ContainerId string
		Request     datacatalog.UpdateConnectionRequest
	}

	var requests []UpdateConnectionRequestInfo
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

			response, err := c.UpdateConnection(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientUpdateDataAsset(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "UpdateDataAsset")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateDataAsset is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "UpdateDataAsset", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "UpdateDataAsset")
	assert.NoError(t, err)

	type UpdateDataAssetRequestInfo struct {
		ContainerId string
		Request     datacatalog.UpdateDataAssetRequest
	}

	var requests []UpdateDataAssetRequestInfo
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

			response, err := c.UpdateDataAsset(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientUpdateEntity(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "UpdateEntity")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateEntity is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "UpdateEntity", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "UpdateEntity")
	assert.NoError(t, err)

	type UpdateEntityRequestInfo struct {
		ContainerId string
		Request     datacatalog.UpdateEntityRequest
	}

	var requests []UpdateEntityRequestInfo
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

			response, err := c.UpdateEntity(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientUpdateFolder(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "UpdateFolder")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateFolder is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "UpdateFolder", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "UpdateFolder")
	assert.NoError(t, err)

	type UpdateFolderRequestInfo struct {
		ContainerId string
		Request     datacatalog.UpdateFolderRequest
	}

	var requests []UpdateFolderRequestInfo
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

			response, err := c.UpdateFolder(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientUpdateGlossary(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "UpdateGlossary")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateGlossary is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "UpdateGlossary", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "UpdateGlossary")
	assert.NoError(t, err)

	type UpdateGlossaryRequestInfo struct {
		ContainerId string
		Request     datacatalog.UpdateGlossaryRequest
	}

	var requests []UpdateGlossaryRequestInfo
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

			response, err := c.UpdateGlossary(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientUpdateJob(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "UpdateJob")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateJob is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "UpdateJob", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "UpdateJob")
	assert.NoError(t, err)

	type UpdateJobRequestInfo struct {
		ContainerId string
		Request     datacatalog.UpdateJobRequest
	}

	var requests []UpdateJobRequestInfo
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

			response, err := c.UpdateJob(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientUpdateJobDefinition(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "UpdateJobDefinition")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateJobDefinition is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "UpdateJobDefinition", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "UpdateJobDefinition")
	assert.NoError(t, err)

	type UpdateJobDefinitionRequestInfo struct {
		ContainerId string
		Request     datacatalog.UpdateJobDefinitionRequest
	}

	var requests []UpdateJobDefinitionRequestInfo
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

			response, err := c.UpdateJobDefinition(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientUpdateTerm(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "UpdateTerm")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateTerm is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "UpdateTerm", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "UpdateTerm")
	assert.NoError(t, err)

	type UpdateTermRequestInfo struct {
		ContainerId string
		Request     datacatalog.UpdateTermRequest
	}

	var requests []UpdateTermRequestInfo
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

			response, err := c.UpdateTerm(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientUpdateTermRelationship(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "UpdateTermRelationship")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateTermRelationship is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "UpdateTermRelationship", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "UpdateTermRelationship")
	assert.NoError(t, err)

	type UpdateTermRelationshipRequestInfo struct {
		ContainerId string
		Request     datacatalog.UpdateTermRelationshipRequest
	}

	var requests []UpdateTermRelationshipRequestInfo
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

			response, err := c.UpdateTermRelationship(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientUploadCredentials(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "UploadCredentials")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UploadCredentials is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "UploadCredentials", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "UploadCredentials")
	assert.NoError(t, err)

	type UploadCredentialsRequestInfo struct {
		ContainerId string
		Request     datacatalog.UploadCredentialsRequest
	}

	var requests []UploadCredentialsRequestInfo
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

			response, err := c.UploadCredentials(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientUsers(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "Users")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("Users is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "Users", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "Users")
	assert.NoError(t, err)

	type UsersRequestInfo struct {
		ContainerId string
		Request     datacatalog.UsersRequest
	}

	var requests []UsersRequestInfo
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
				r := req.(*datacatalog.UsersRequest)
				return c.Users(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]datacatalog.UsersResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(datacatalog.UsersResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="datacatalog_ww_grp@oracle.com" jiraProject="DCAT" opsJiraProject="ADCS"
func TestDataCatalogClientValidateConnection(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("datacatalog", "ValidateConnection")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ValidateConnection is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("datacatalog", "DataCatalog", "ValidateConnection", createDataCatalogClientWithProvider)
	assert.NoError(t, err)
	c := cc.(datacatalog.DataCatalogClient)

	body, err := testClient.getRequests("datacatalog", "ValidateConnection")
	assert.NoError(t, err)

	type ValidateConnectionRequestInfo struct {
		ContainerId string
		Request     datacatalog.ValidateConnectionRequest
	}

	var requests []ValidateConnectionRequestInfo
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

			response, err := c.ValidateConnection(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
