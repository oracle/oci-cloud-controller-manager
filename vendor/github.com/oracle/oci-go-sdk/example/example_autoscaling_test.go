// Copyright (c) 2019, Oracle and/or its affiliates. All rights reserved.
//
// Example code for Autoscaling Services API
//
//

/**
 * This class provides an example of how you can create an AutoScalingConfiguration and use that with a InstancePools. It will:
 * <ul>
 * <li>Create the InstanceConfiguration</li>
 * <li>Create a pool based off that configuration.</li>
 * <li>Create an AutoScalingConfiguration for that pool.</li>
 * <li>Clean everything up.</li>
 * </ul>
 */

package example

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/oracle/oci-go-sdk/autoscaling"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-go-sdk/example/helpers"
)

// Example to showcase autoscaling configuration creation and eventual teardown
func ExampleCreateAndDeleteAutoscalingConfiguration() {
	AutoscalingParseEnvironmentVariables()

	ctx := context.Background()

	computeMgmtClient, err := core.NewComputeManagementClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(err)

	autoscalingClient, err := autoscaling.NewAutoScalingClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(err)

	createInstanceConfigurationResponse, _ := createInstanceConfiguration(ctx, computeMgmtClient, imageId, compartmentId)
	fmt.Println("Instance configuration created")

	instanceConfiguration := createInstanceConfigurationResponse.InstanceConfiguration

	instancePool, _ := createInstancePool(ctx, computeMgmtClient, *instanceConfiguration.Id, subnetId, ad, compartmentId)
	fmt.Println("Instance pool created")

	autoscalingConfiguration, _ := createAutoscalingConfiguration(ctx, autoscalingClient, *instancePool.Id, compartmentId)

	fmt.Println("Autoscaling configuration created")

	// clean up resources
	defer func() {
		deleteAutoscalingConfiguration(ctx, autoscalingClient, *autoscalingConfiguration.Id)
		fmt.Println("Deleted Autoscaling Configuration")

		terminateInstancePool(ctx, computeMgmtClient, *instancePool.Id)
		fmt.Println("Terminated Instance Pool")

		deleteInstanceConfiguration(ctx, computeMgmtClient, *instanceConfiguration.Id)
		fmt.Println("Deleted Instance Configuration")
	}()

	// Output:
	// Instance configuration created
	// Instance pool created
	// Autoscaling configuration created
	// Deleted Autoscaling Configuration
	// Terminated Instance Pool
	// Deleted Instance Configuration
}

// Usage printing
func AutoscalingUsage() {
	log.Printf("Please set the following environment variables to run Autoscaling sample")
	log.Printf(" ")
	log.Printf("   IMAGE_ID       # Required: Image Id to use")
	log.Printf("   COMPARTMENT_ID    # Required: Compartment Id to use")
	log.Printf("   AD          # Required: AD to use")
	log.Printf("   SUBNET_ID   # Required: Subnet to use")
	log.Printf(" ")
	os.Exit(1)
}

// Args parser
func AutoscalingParseEnvironmentVariables() {

	imageId = os.Getenv("IMAGE_ID")
	compartmentId = os.Getenv("COMPARTMENT_ID")
	ad = os.Getenv("AD")
	subnetId = os.Getenv("SUBNET_ID")

	if imageId == "" ||
		compartmentId == "" ||
		ad == "" ||
		subnetId == "" {
		AutoscalingUsage()
	}

	log.Printf("IMAGE_ID     : %s", imageId)
	log.Printf("COMPARTMENT_ID  : %s", compartmentId)
	log.Printf("AD     : %s", ad)
	log.Printf("SUBNET_ID  : %s", subnetId)
}

// helper method to create an autoscaling configuration
func createAutoscalingConfiguration(ctx context.Context, client autoscaling.AutoScalingClient,
	instancePoolId string, compartmentId string) (response autoscaling.CreateAutoScalingConfigurationResponse, err error) {

	displayName := "Autoscaling Example"

	scaleInThreshold := 30
	scaleInChange := -1
	scaleOutThreshold := 70
	scaleOutChange := 1
	capInit := 2
	capMax := 3
	capMin := 1

	resource := autoscaling.InstancePoolResource{
		Id: &instancePoolId,
	}

	capacity := autoscaling.Capacity{
		Initial: &capInit,
		Max:     &capMax,
		Min:     &capMin,
	}

	// scale in params
	lowerBound := autoscaling.Threshold{
		Operator: autoscaling.ThresholdOperatorLt,
		Value:    &scaleInThreshold,
	}

	scaleInAction := autoscaling.Action{
		Type:  autoscaling.ActionTypeChangeCountBy,
		Value: &scaleInChange,
	}

	scaleInMetric := autoscaling.Metric{
		Threshold:  &lowerBound,
		MetricType: autoscaling.MetricMetricTypeCpuUtilization,
	}

	scaleInRule := autoscaling.CreateConditionDetails{
		Action: &scaleInAction,
		Metric: &scaleInMetric,
	}

	// scale out params
	upperBound := autoscaling.Threshold{
		Operator: autoscaling.ThresholdOperatorGt,
		Value:    &scaleOutThreshold,
	}

	scaleOutAction := autoscaling.Action{
		Type:  autoscaling.ActionTypeChangeCountBy,
		Value: &scaleOutChange,
	}

	scaleOutMetric := autoscaling.Metric{
		Threshold:  &upperBound,
		MetricType: autoscaling.MetricMetricTypeCpuUtilization,
	}

	scaleOutRule := autoscaling.CreateConditionDetails{
		Action: &scaleOutAction,
		Metric: &scaleOutMetric,
	}

	// defining the threshold policy
	policy := autoscaling.CreateThresholdPolicyDetails{
		Capacity: &capacity,
		Rules: []autoscaling.CreateConditionDetails{
			scaleInRule,
			scaleOutRule,
		},
	}

	// defining the autoscaling configuration
	createAutoscalingConfigurationDetails := autoscaling.CreateAutoScalingConfigurationDetails{
		DisplayName:   &displayName,
		CompartmentId: &compartmentId,
		Resource:      &resource,
		Policies:      []autoscaling.CreateAutoScalingPolicyDetails{&policy},
	}

	req := autoscaling.CreateAutoScalingConfigurationRequest{
		CreateAutoScalingConfigurationDetails: createAutoscalingConfigurationDetails,
	}

	response, err = client.CreateAutoScalingConfiguration(ctx, req)
	helpers.FatalIfError(err)

	return
}

// helper method to delete an instance configuration
func deleteAutoscalingConfiguration(ctx context.Context, client autoscaling.AutoScalingClient,
	autoscalingConfigurationId string) (response autoscaling.DeleteAutoScalingConfigurationResponse, err error) {

	req := autoscaling.DeleteAutoScalingConfigurationRequest{
		AutoScalingConfigurationId: &autoscalingConfigurationId,
	}

	response, err = client.DeleteAutoScalingConfiguration(ctx, req)
	helpers.FatalIfError(err)

	return
}
