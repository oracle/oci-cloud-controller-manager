package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createComputeManagementClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := core.NewComputeManagementClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientAttachLoadBalancer(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "AttachLoadBalancer")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("AttachLoadBalancer is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "AttachLoadBalancer", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "AttachLoadBalancer")
	assert.NoError(t, err)

	type AttachLoadBalancerRequestInfo struct {
		ContainerId string
		Request     core.AttachLoadBalancerRequest
	}

	var requests []AttachLoadBalancerRequestInfo
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

			response, err := c.AttachLoadBalancer(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientChangeClusterNetworkCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "ChangeClusterNetworkCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeClusterNetworkCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "ChangeClusterNetworkCompartment", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "ChangeClusterNetworkCompartment")
	assert.NoError(t, err)

	type ChangeClusterNetworkCompartmentRequestInfo struct {
		ContainerId string
		Request     core.ChangeClusterNetworkCompartmentRequest
	}

	var requests []ChangeClusterNetworkCompartmentRequestInfo
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

			response, err := c.ChangeClusterNetworkCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientChangeInstanceConfigurationCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "ChangeInstanceConfigurationCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeInstanceConfigurationCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "ChangeInstanceConfigurationCompartment", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "ChangeInstanceConfigurationCompartment")
	assert.NoError(t, err)

	type ChangeInstanceConfigurationCompartmentRequestInfo struct {
		ContainerId string
		Request     core.ChangeInstanceConfigurationCompartmentRequest
	}

	var requests []ChangeInstanceConfigurationCompartmentRequestInfo
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

			response, err := c.ChangeInstanceConfigurationCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientChangeInstancePoolCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "ChangeInstancePoolCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeInstancePoolCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "ChangeInstancePoolCompartment", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "ChangeInstancePoolCompartment")
	assert.NoError(t, err)

	type ChangeInstancePoolCompartmentRequestInfo struct {
		ContainerId string
		Request     core.ChangeInstancePoolCompartmentRequest
	}

	var requests []ChangeInstancePoolCompartmentRequestInfo
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

			response, err := c.ChangeInstancePoolCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientCreateClusterNetwork(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "CreateClusterNetwork")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateClusterNetwork is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "CreateClusterNetwork", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "CreateClusterNetwork")
	assert.NoError(t, err)

	type CreateClusterNetworkRequestInfo struct {
		ContainerId string
		Request     core.CreateClusterNetworkRequest
	}

	var requests []CreateClusterNetworkRequestInfo
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

			response, err := c.CreateClusterNetwork(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientCreateInstanceConfiguration(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "CreateInstanceConfiguration")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateInstanceConfiguration is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "CreateInstanceConfiguration", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "CreateInstanceConfiguration")
	assert.NoError(t, err)

	type CreateInstanceConfigurationRequestInfo struct {
		ContainerId string
		Request     core.CreateInstanceConfigurationRequest
	}

	var requests []CreateInstanceConfigurationRequestInfo
	var pr []map[string]interface{}
	err = json.Unmarshal([]byte(body), &pr)
	assert.NoError(t, err)
	requests = make([]CreateInstanceConfigurationRequestInfo, len(pr))
	polymorphicRequestInfo := map[string]PolymorphicRequestUnmarshallingInfo{}
	polymorphicRequestInfo["CreateInstanceConfigurationBase"] =
		PolymorphicRequestUnmarshallingInfo{
			DiscriminatorName: "source",
			DiscriminatorValuesAndTypes: map[string]interface{}{
				"NONE":     &core.CreateInstanceConfigurationDetails{},
				"INSTANCE": &core.CreateInstanceConfigurationFromInstanceDetails{},
			},
		}

	for i, ppr := range pr {
		conditionalStructCopy(ppr, &requests[i], polymorphicRequestInfo, testClient.Log)
	}

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.CreateInstanceConfiguration(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientCreateInstancePool(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "CreateInstancePool")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateInstancePool is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "CreateInstancePool", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "CreateInstancePool")
	assert.NoError(t, err)

	type CreateInstancePoolRequestInfo struct {
		ContainerId string
		Request     core.CreateInstancePoolRequest
	}

	var requests []CreateInstancePoolRequestInfo
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

			response, err := c.CreateInstancePool(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientDeleteInstanceConfiguration(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "DeleteInstanceConfiguration")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteInstanceConfiguration is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "DeleteInstanceConfiguration", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "DeleteInstanceConfiguration")
	assert.NoError(t, err)

	type DeleteInstanceConfigurationRequestInfo struct {
		ContainerId string
		Request     core.DeleteInstanceConfigurationRequest
	}

	var requests []DeleteInstanceConfigurationRequestInfo
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

			response, err := c.DeleteInstanceConfiguration(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientDetachLoadBalancer(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "DetachLoadBalancer")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DetachLoadBalancer is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "DetachLoadBalancer", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "DetachLoadBalancer")
	assert.NoError(t, err)

	type DetachLoadBalancerRequestInfo struct {
		ContainerId string
		Request     core.DetachLoadBalancerRequest
	}

	var requests []DetachLoadBalancerRequestInfo
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

			response, err := c.DetachLoadBalancer(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientGetClusterNetwork(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "GetClusterNetwork")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetClusterNetwork is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "GetClusterNetwork", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "GetClusterNetwork")
	assert.NoError(t, err)

	type GetClusterNetworkRequestInfo struct {
		ContainerId string
		Request     core.GetClusterNetworkRequest
	}

	var requests []GetClusterNetworkRequestInfo
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

			response, err := c.GetClusterNetwork(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientGetInstanceConfiguration(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "GetInstanceConfiguration")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetInstanceConfiguration is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "GetInstanceConfiguration", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "GetInstanceConfiguration")
	assert.NoError(t, err)

	type GetInstanceConfigurationRequestInfo struct {
		ContainerId string
		Request     core.GetInstanceConfigurationRequest
	}

	var requests []GetInstanceConfigurationRequestInfo
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

			response, err := c.GetInstanceConfiguration(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientGetInstancePool(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "GetInstancePool")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetInstancePool is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "GetInstancePool", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "GetInstancePool")
	assert.NoError(t, err)

	type GetInstancePoolRequestInfo struct {
		ContainerId string
		Request     core.GetInstancePoolRequest
	}

	var requests []GetInstancePoolRequestInfo
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

			response, err := c.GetInstancePool(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientGetInstancePoolLoadBalancerAttachment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "GetInstancePoolLoadBalancerAttachment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetInstancePoolLoadBalancerAttachment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "GetInstancePoolLoadBalancerAttachment", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "GetInstancePoolLoadBalancerAttachment")
	assert.NoError(t, err)

	type GetInstancePoolLoadBalancerAttachmentRequestInfo struct {
		ContainerId string
		Request     core.GetInstancePoolLoadBalancerAttachmentRequest
	}

	var requests []GetInstancePoolLoadBalancerAttachmentRequestInfo
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

			response, err := c.GetInstancePoolLoadBalancerAttachment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientLaunchInstanceConfiguration(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "LaunchInstanceConfiguration")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("LaunchInstanceConfiguration is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "LaunchInstanceConfiguration", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "LaunchInstanceConfiguration")
	assert.NoError(t, err)

	type LaunchInstanceConfigurationRequestInfo struct {
		ContainerId string
		Request     core.LaunchInstanceConfigurationRequest
	}

	var requests []LaunchInstanceConfigurationRequestInfo
	var pr []map[string]interface{}
	err = json.Unmarshal([]byte(body), &pr)
	assert.NoError(t, err)
	requests = make([]LaunchInstanceConfigurationRequestInfo, len(pr))
	polymorphicRequestInfo := map[string]PolymorphicRequestUnmarshallingInfo{}
	polymorphicRequestInfo["InstanceConfigurationInstanceDetails"] =
		PolymorphicRequestUnmarshallingInfo{
			DiscriminatorName: "instanceType",
			DiscriminatorValuesAndTypes: map[string]interface{}{
				"compute": &core.ComputeInstanceDetails{},
			},
		}

	for i, ppr := range pr {
		conditionalStructCopy(ppr, &requests[i], polymorphicRequestInfo, testClient.Log)
	}

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.LaunchInstanceConfiguration(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientListClusterNetworkInstances(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "ListClusterNetworkInstances")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListClusterNetworkInstances is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "ListClusterNetworkInstances", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "ListClusterNetworkInstances")
	assert.NoError(t, err)

	type ListClusterNetworkInstancesRequestInfo struct {
		ContainerId string
		Request     core.ListClusterNetworkInstancesRequest
	}

	var requests []ListClusterNetworkInstancesRequestInfo
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
				r := req.(*core.ListClusterNetworkInstancesRequest)
				return c.ListClusterNetworkInstances(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]core.ListClusterNetworkInstancesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(core.ListClusterNetworkInstancesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientListClusterNetworks(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "ListClusterNetworks")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListClusterNetworks is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "ListClusterNetworks", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "ListClusterNetworks")
	assert.NoError(t, err)

	type ListClusterNetworksRequestInfo struct {
		ContainerId string
		Request     core.ListClusterNetworksRequest
	}

	var requests []ListClusterNetworksRequestInfo
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
				r := req.(*core.ListClusterNetworksRequest)
				return c.ListClusterNetworks(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]core.ListClusterNetworksResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(core.ListClusterNetworksResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientListInstanceConfigurations(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "ListInstanceConfigurations")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListInstanceConfigurations is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "ListInstanceConfigurations", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "ListInstanceConfigurations")
	assert.NoError(t, err)

	type ListInstanceConfigurationsRequestInfo struct {
		ContainerId string
		Request     core.ListInstanceConfigurationsRequest
	}

	var requests []ListInstanceConfigurationsRequestInfo
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
				r := req.(*core.ListInstanceConfigurationsRequest)
				return c.ListInstanceConfigurations(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]core.ListInstanceConfigurationsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(core.ListInstanceConfigurationsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientListInstancePoolInstances(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "ListInstancePoolInstances")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListInstancePoolInstances is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "ListInstancePoolInstances", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "ListInstancePoolInstances")
	assert.NoError(t, err)

	type ListInstancePoolInstancesRequestInfo struct {
		ContainerId string
		Request     core.ListInstancePoolInstancesRequest
	}

	var requests []ListInstancePoolInstancesRequestInfo
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
				r := req.(*core.ListInstancePoolInstancesRequest)
				return c.ListInstancePoolInstances(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]core.ListInstancePoolInstancesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(core.ListInstancePoolInstancesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientListInstancePools(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "ListInstancePools")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListInstancePools is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "ListInstancePools", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "ListInstancePools")
	assert.NoError(t, err)

	type ListInstancePoolsRequestInfo struct {
		ContainerId string
		Request     core.ListInstancePoolsRequest
	}

	var requests []ListInstancePoolsRequestInfo
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
				r := req.(*core.ListInstancePoolsRequest)
				return c.ListInstancePools(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]core.ListInstancePoolsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(core.ListInstancePoolsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientResetInstancePool(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "ResetInstancePool")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ResetInstancePool is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "ResetInstancePool", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "ResetInstancePool")
	assert.NoError(t, err)

	type ResetInstancePoolRequestInfo struct {
		ContainerId string
		Request     core.ResetInstancePoolRequest
	}

	var requests []ResetInstancePoolRequestInfo
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

			response, err := c.ResetInstancePool(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientSoftresetInstancePool(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "SoftresetInstancePool")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("SoftresetInstancePool is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "SoftresetInstancePool", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "SoftresetInstancePool")
	assert.NoError(t, err)

	type SoftresetInstancePoolRequestInfo struct {
		ContainerId string
		Request     core.SoftresetInstancePoolRequest
	}

	var requests []SoftresetInstancePoolRequestInfo
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

			response, err := c.SoftresetInstancePool(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientStartInstancePool(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "StartInstancePool")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("StartInstancePool is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "StartInstancePool", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "StartInstancePool")
	assert.NoError(t, err)

	type StartInstancePoolRequestInfo struct {
		ContainerId string
		Request     core.StartInstancePoolRequest
	}

	var requests []StartInstancePoolRequestInfo
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

			response, err := c.StartInstancePool(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientStopInstancePool(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "StopInstancePool")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("StopInstancePool is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "StopInstancePool", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "StopInstancePool")
	assert.NoError(t, err)

	type StopInstancePoolRequestInfo struct {
		ContainerId string
		Request     core.StopInstancePoolRequest
	}

	var requests []StopInstancePoolRequestInfo
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

			response, err := c.StopInstancePool(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientTerminateClusterNetwork(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "TerminateClusterNetwork")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("TerminateClusterNetwork is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "TerminateClusterNetwork", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "TerminateClusterNetwork")
	assert.NoError(t, err)

	type TerminateClusterNetworkRequestInfo struct {
		ContainerId string
		Request     core.TerminateClusterNetworkRequest
	}

	var requests []TerminateClusterNetworkRequestInfo
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

			response, err := c.TerminateClusterNetwork(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientTerminateInstancePool(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "TerminateInstancePool")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("TerminateInstancePool is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "TerminateInstancePool", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "TerminateInstancePool")
	assert.NoError(t, err)

	type TerminateInstancePoolRequestInfo struct {
		ContainerId string
		Request     core.TerminateInstancePoolRequest
	}

	var requests []TerminateInstancePoolRequestInfo
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

			response, err := c.TerminateInstancePool(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientUpdateClusterNetwork(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "UpdateClusterNetwork")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateClusterNetwork is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "UpdateClusterNetwork", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "UpdateClusterNetwork")
	assert.NoError(t, err)

	type UpdateClusterNetworkRequestInfo struct {
		ContainerId string
		Request     core.UpdateClusterNetworkRequest
	}

	var requests []UpdateClusterNetworkRequestInfo
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

			response, err := c.UpdateClusterNetwork(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientUpdateInstanceConfiguration(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "UpdateInstanceConfiguration")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateInstanceConfiguration is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "UpdateInstanceConfiguration", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "UpdateInstanceConfiguration")
	assert.NoError(t, err)

	type UpdateInstanceConfigurationRequestInfo struct {
		ContainerId string
		Request     core.UpdateInstanceConfigurationRequest
	}

	var requests []UpdateInstanceConfigurationRequestInfo
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

			response, err := c.UpdateInstanceConfiguration(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="computeManagement" email="instance_dev_us_grp@oracle.com" jiraProject="CIM" opsJiraProject="IPA"
func TestComputeManagementClientUpdateInstancePool(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "UpdateInstancePool")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateInstancePool is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "ComputeManagement", "UpdateInstancePool", createComputeManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.ComputeManagementClient)

	body, err := testClient.getRequests("core", "UpdateInstancePool")
	assert.NoError(t, err)

	type UpdateInstancePoolRequestInfo struct {
		ContainerId string
		Request     core.UpdateInstancePoolRequest
	}

	var requests []UpdateInstancePoolRequestInfo
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

			response, err := c.UpdateInstancePool(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
