package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/nosql"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createNosqlClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := nosql.NewNosqlClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="andc_ops_ww_grp@oracle.com" jiraProject="NOSQL" opsJiraProject="NOSQL"
func TestNosqlClientChangeTableCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("nosql", "ChangeTableCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeTableCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("nosql", "Nosql", "ChangeTableCompartment", createNosqlClientWithProvider)
	assert.NoError(t, err)
	c := cc.(nosql.NosqlClient)

	body, err := testClient.getRequests("nosql", "ChangeTableCompartment")
	assert.NoError(t, err)

	type ChangeTableCompartmentRequestInfo struct {
		ContainerId string
		Request     nosql.ChangeTableCompartmentRequest
	}

	var requests []ChangeTableCompartmentRequestInfo
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

			response, err := c.ChangeTableCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="andc_ops_ww_grp@oracle.com" jiraProject="NOSQL" opsJiraProject="NOSQL"
func TestNosqlClientCreateIndex(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("nosql", "CreateIndex")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateIndex is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("nosql", "Nosql", "CreateIndex", createNosqlClientWithProvider)
	assert.NoError(t, err)
	c := cc.(nosql.NosqlClient)

	body, err := testClient.getRequests("nosql", "CreateIndex")
	assert.NoError(t, err)

	type CreateIndexRequestInfo struct {
		ContainerId string
		Request     nosql.CreateIndexRequest
	}

	var requests []CreateIndexRequestInfo
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

			response, err := c.CreateIndex(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="andc_ops_ww_grp@oracle.com" jiraProject="NOSQL" opsJiraProject="NOSQL"
func TestNosqlClientCreateTable(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("nosql", "CreateTable")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateTable is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("nosql", "Nosql", "CreateTable", createNosqlClientWithProvider)
	assert.NoError(t, err)
	c := cc.(nosql.NosqlClient)

	body, err := testClient.getRequests("nosql", "CreateTable")
	assert.NoError(t, err)

	type CreateTableRequestInfo struct {
		ContainerId string
		Request     nosql.CreateTableRequest
	}

	var requests []CreateTableRequestInfo
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

			response, err := c.CreateTable(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="andc_ops_ww_grp@oracle.com" jiraProject="NOSQL" opsJiraProject="NOSQL"
func TestNosqlClientDeleteIndex(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("nosql", "DeleteIndex")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteIndex is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("nosql", "Nosql", "DeleteIndex", createNosqlClientWithProvider)
	assert.NoError(t, err)
	c := cc.(nosql.NosqlClient)

	body, err := testClient.getRequests("nosql", "DeleteIndex")
	assert.NoError(t, err)

	type DeleteIndexRequestInfo struct {
		ContainerId string
		Request     nosql.DeleteIndexRequest
	}

	var requests []DeleteIndexRequestInfo
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

			response, err := c.DeleteIndex(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="andc_ops_ww_grp@oracle.com" jiraProject="NOSQL" opsJiraProject="NOSQL"
func TestNosqlClientDeleteRow(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("nosql", "DeleteRow")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteRow is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("nosql", "Nosql", "DeleteRow", createNosqlClientWithProvider)
	assert.NoError(t, err)
	c := cc.(nosql.NosqlClient)

	body, err := testClient.getRequests("nosql", "DeleteRow")
	assert.NoError(t, err)

	type DeleteRowRequestInfo struct {
		ContainerId string
		Request     nosql.DeleteRowRequest
	}

	var requests []DeleteRowRequestInfo
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

			response, err := c.DeleteRow(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="andc_ops_ww_grp@oracle.com" jiraProject="NOSQL" opsJiraProject="NOSQL"
func TestNosqlClientDeleteTable(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("nosql", "DeleteTable")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteTable is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("nosql", "Nosql", "DeleteTable", createNosqlClientWithProvider)
	assert.NoError(t, err)
	c := cc.(nosql.NosqlClient)

	body, err := testClient.getRequests("nosql", "DeleteTable")
	assert.NoError(t, err)

	type DeleteTableRequestInfo struct {
		ContainerId string
		Request     nosql.DeleteTableRequest
	}

	var requests []DeleteTableRequestInfo
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

			response, err := c.DeleteTable(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="andc_ops_ww_grp@oracle.com" jiraProject="NOSQL" opsJiraProject="NOSQL"
func TestNosqlClientDeleteWorkRequest(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("nosql", "DeleteWorkRequest")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteWorkRequest is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("nosql", "Nosql", "DeleteWorkRequest", createNosqlClientWithProvider)
	assert.NoError(t, err)
	c := cc.(nosql.NosqlClient)

	body, err := testClient.getRequests("nosql", "DeleteWorkRequest")
	assert.NoError(t, err)

	type DeleteWorkRequestRequestInfo struct {
		ContainerId string
		Request     nosql.DeleteWorkRequestRequest
	}

	var requests []DeleteWorkRequestRequestInfo
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

			response, err := c.DeleteWorkRequest(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="andc_ops_ww_grp@oracle.com" jiraProject="NOSQL" opsJiraProject="NOSQL"
func TestNosqlClientGetIndex(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("nosql", "GetIndex")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetIndex is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("nosql", "Nosql", "GetIndex", createNosqlClientWithProvider)
	assert.NoError(t, err)
	c := cc.(nosql.NosqlClient)

	body, err := testClient.getRequests("nosql", "GetIndex")
	assert.NoError(t, err)

	type GetIndexRequestInfo struct {
		ContainerId string
		Request     nosql.GetIndexRequest
	}

	var requests []GetIndexRequestInfo
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

			response, err := c.GetIndex(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="andc_ops_ww_grp@oracle.com" jiraProject="NOSQL" opsJiraProject="NOSQL"
func TestNosqlClientGetRow(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("nosql", "GetRow")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetRow is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("nosql", "Nosql", "GetRow", createNosqlClientWithProvider)
	assert.NoError(t, err)
	c := cc.(nosql.NosqlClient)

	body, err := testClient.getRequests("nosql", "GetRow")
	assert.NoError(t, err)

	type GetRowRequestInfo struct {
		ContainerId string
		Request     nosql.GetRowRequest
	}

	var requests []GetRowRequestInfo
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

			response, err := c.GetRow(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="andc_ops_ww_grp@oracle.com" jiraProject="NOSQL" opsJiraProject="NOSQL"
func TestNosqlClientGetTable(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("nosql", "GetTable")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetTable is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("nosql", "Nosql", "GetTable", createNosqlClientWithProvider)
	assert.NoError(t, err)
	c := cc.(nosql.NosqlClient)

	body, err := testClient.getRequests("nosql", "GetTable")
	assert.NoError(t, err)

	type GetTableRequestInfo struct {
		ContainerId string
		Request     nosql.GetTableRequest
	}

	var requests []GetTableRequestInfo
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

			response, err := c.GetTable(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="andc_ops_ww_grp@oracle.com" jiraProject="NOSQL" opsJiraProject="NOSQL"
func TestNosqlClientGetWorkRequest(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("nosql", "GetWorkRequest")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetWorkRequest is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("nosql", "Nosql", "GetWorkRequest", createNosqlClientWithProvider)
	assert.NoError(t, err)
	c := cc.(nosql.NosqlClient)

	body, err := testClient.getRequests("nosql", "GetWorkRequest")
	assert.NoError(t, err)

	type GetWorkRequestRequestInfo struct {
		ContainerId string
		Request     nosql.GetWorkRequestRequest
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

// IssueRoutingInfo tag="default" email="andc_ops_ww_grp@oracle.com" jiraProject="NOSQL" opsJiraProject="NOSQL"
func TestNosqlClientListIndexes(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("nosql", "ListIndexes")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListIndexes is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("nosql", "Nosql", "ListIndexes", createNosqlClientWithProvider)
	assert.NoError(t, err)
	c := cc.(nosql.NosqlClient)

	body, err := testClient.getRequests("nosql", "ListIndexes")
	assert.NoError(t, err)

	type ListIndexesRequestInfo struct {
		ContainerId string
		Request     nosql.ListIndexesRequest
	}

	var requests []ListIndexesRequestInfo
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
				r := req.(*nosql.ListIndexesRequest)
				return c.ListIndexes(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]nosql.ListIndexesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(nosql.ListIndexesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="andc_ops_ww_grp@oracle.com" jiraProject="NOSQL" opsJiraProject="NOSQL"
func TestNosqlClientListTableUsage(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("nosql", "ListTableUsage")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListTableUsage is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("nosql", "Nosql", "ListTableUsage", createNosqlClientWithProvider)
	assert.NoError(t, err)
	c := cc.(nosql.NosqlClient)

	body, err := testClient.getRequests("nosql", "ListTableUsage")
	assert.NoError(t, err)

	type ListTableUsageRequestInfo struct {
		ContainerId string
		Request     nosql.ListTableUsageRequest
	}

	var requests []ListTableUsageRequestInfo
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
				r := req.(*nosql.ListTableUsageRequest)
				return c.ListTableUsage(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]nosql.ListTableUsageResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(nosql.ListTableUsageResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="andc_ops_ww_grp@oracle.com" jiraProject="NOSQL" opsJiraProject="NOSQL"
func TestNosqlClientListTables(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("nosql", "ListTables")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListTables is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("nosql", "Nosql", "ListTables", createNosqlClientWithProvider)
	assert.NoError(t, err)
	c := cc.(nosql.NosqlClient)

	body, err := testClient.getRequests("nosql", "ListTables")
	assert.NoError(t, err)

	type ListTablesRequestInfo struct {
		ContainerId string
		Request     nosql.ListTablesRequest
	}

	var requests []ListTablesRequestInfo
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
				r := req.(*nosql.ListTablesRequest)
				return c.ListTables(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]nosql.ListTablesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(nosql.ListTablesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="andc_ops_ww_grp@oracle.com" jiraProject="NOSQL" opsJiraProject="NOSQL"
func TestNosqlClientListWorkRequestErrors(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("nosql", "ListWorkRequestErrors")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequestErrors is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("nosql", "Nosql", "ListWorkRequestErrors", createNosqlClientWithProvider)
	assert.NoError(t, err)
	c := cc.(nosql.NosqlClient)

	body, err := testClient.getRequests("nosql", "ListWorkRequestErrors")
	assert.NoError(t, err)

	type ListWorkRequestErrorsRequestInfo struct {
		ContainerId string
		Request     nosql.ListWorkRequestErrorsRequest
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
				r := req.(*nosql.ListWorkRequestErrorsRequest)
				return c.ListWorkRequestErrors(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]nosql.ListWorkRequestErrorsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(nosql.ListWorkRequestErrorsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="andc_ops_ww_grp@oracle.com" jiraProject="NOSQL" opsJiraProject="NOSQL"
func TestNosqlClientListWorkRequestLogs(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("nosql", "ListWorkRequestLogs")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequestLogs is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("nosql", "Nosql", "ListWorkRequestLogs", createNosqlClientWithProvider)
	assert.NoError(t, err)
	c := cc.(nosql.NosqlClient)

	body, err := testClient.getRequests("nosql", "ListWorkRequestLogs")
	assert.NoError(t, err)

	type ListWorkRequestLogsRequestInfo struct {
		ContainerId string
		Request     nosql.ListWorkRequestLogsRequest
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
				r := req.(*nosql.ListWorkRequestLogsRequest)
				return c.ListWorkRequestLogs(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]nosql.ListWorkRequestLogsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(nosql.ListWorkRequestLogsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="andc_ops_ww_grp@oracle.com" jiraProject="NOSQL" opsJiraProject="NOSQL"
func TestNosqlClientListWorkRequests(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("nosql", "ListWorkRequests")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequests is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("nosql", "Nosql", "ListWorkRequests", createNosqlClientWithProvider)
	assert.NoError(t, err)
	c := cc.(nosql.NosqlClient)

	body, err := testClient.getRequests("nosql", "ListWorkRequests")
	assert.NoError(t, err)

	type ListWorkRequestsRequestInfo struct {
		ContainerId string
		Request     nosql.ListWorkRequestsRequest
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
				r := req.(*nosql.ListWorkRequestsRequest)
				return c.ListWorkRequests(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]nosql.ListWorkRequestsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(nosql.ListWorkRequestsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="andc_ops_ww_grp@oracle.com" jiraProject="NOSQL" opsJiraProject="NOSQL"
func TestNosqlClientPrepareStatement(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("nosql", "PrepareStatement")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("PrepareStatement is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("nosql", "Nosql", "PrepareStatement", createNosqlClientWithProvider)
	assert.NoError(t, err)
	c := cc.(nosql.NosqlClient)

	body, err := testClient.getRequests("nosql", "PrepareStatement")
	assert.NoError(t, err)

	type PrepareStatementRequestInfo struct {
		ContainerId string
		Request     nosql.PrepareStatementRequest
	}

	var requests []PrepareStatementRequestInfo
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

			response, err := c.PrepareStatement(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="andc_ops_ww_grp@oracle.com" jiraProject="NOSQL" opsJiraProject="NOSQL"
func TestNosqlClientQuery(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("nosql", "Query")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("Query is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("nosql", "Nosql", "Query", createNosqlClientWithProvider)
	assert.NoError(t, err)
	c := cc.(nosql.NosqlClient)

	body, err := testClient.getRequests("nosql", "Query")
	assert.NoError(t, err)

	type QueryRequestInfo struct {
		ContainerId string
		Request     nosql.QueryRequest
	}

	var requests []QueryRequestInfo
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
				r := req.(*nosql.QueryRequest)
				return c.Query(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]nosql.QueryResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(nosql.QueryResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="andc_ops_ww_grp@oracle.com" jiraProject="NOSQL" opsJiraProject="NOSQL"
func TestNosqlClientSummarizeStatement(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("nosql", "SummarizeStatement")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("SummarizeStatement is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("nosql", "Nosql", "SummarizeStatement", createNosqlClientWithProvider)
	assert.NoError(t, err)
	c := cc.(nosql.NosqlClient)

	body, err := testClient.getRequests("nosql", "SummarizeStatement")
	assert.NoError(t, err)

	type SummarizeStatementRequestInfo struct {
		ContainerId string
		Request     nosql.SummarizeStatementRequest
	}

	var requests []SummarizeStatementRequestInfo
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

			response, err := c.SummarizeStatement(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="andc_ops_ww_grp@oracle.com" jiraProject="NOSQL" opsJiraProject="NOSQL"
func TestNosqlClientUpdateRow(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("nosql", "UpdateRow")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateRow is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("nosql", "Nosql", "UpdateRow", createNosqlClientWithProvider)
	assert.NoError(t, err)
	c := cc.(nosql.NosqlClient)

	body, err := testClient.getRequests("nosql", "UpdateRow")
	assert.NoError(t, err)

	type UpdateRowRequestInfo struct {
		ContainerId string
		Request     nosql.UpdateRowRequest
	}

	var requests []UpdateRowRequestInfo
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

			response, err := c.UpdateRow(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="andc_ops_ww_grp@oracle.com" jiraProject="NOSQL" opsJiraProject="NOSQL"
func TestNosqlClientUpdateTable(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("nosql", "UpdateTable")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateTable is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("nosql", "Nosql", "UpdateTable", createNosqlClientWithProvider)
	assert.NoError(t, err)
	c := cc.(nosql.NosqlClient)

	body, err := testClient.getRequests("nosql", "UpdateTable")
	assert.NoError(t, err)

	type UpdateTableRequestInfo struct {
		ContainerId string
		Request     nosql.UpdateTableRequest
	}

	var requests []UpdateTableRequestInfo
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

			response, err := c.UpdateTable(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
