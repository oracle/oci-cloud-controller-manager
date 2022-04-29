#!/bin/bash

##################################################################################################
# This template can be used to tweak the environment variables needed to run the E2E tests locally #
# Default behavior:
# Runs test on an existing cluster
# Note: All variables that have comment as "# Mandatory" need to be filled with appropriate values for the tests to run correctly.

# To run the tests:
# 1. Change the FOCUS valiable here to specify the subset of E2E tests to run
# 2. Set CLUSTER_KUBECONFIG and CLOUD_CONFIG if needed
# 3. run 'source existing-standalone-cluster-env-template.sh' to set the variables
# 4. run 'make run-ccm-e2e-tests-local`
##################################################################################################

# The test suites to run (can replace or add tags)
export FOCUS="\[cloudprovider\]"

# Scope can be ARM / AMD / BOTH
# Mandatory
export SCOPE="BOTH"

# A Reserved IP in your compartment for testing LB creation with Reserved IP
# Create a public reserved IP in your compartment using the following link:
# https://docs.oracle.com/en-us/iaas/Content/Network/Tasks/managingpublicIPs.htm#console-reserved
# Set the public reserved IP in the following env-variable:
# Mandatory
export RESERVED_IP=""

# Set path to kubeconfig of existing cluster if it does not exist in default path. Defaults to $HOME/.kube/config.
# Mandatory
export CLUSTER_KUBECONFIG=$HOME/.kube/config

# Set path to cloud_config of existing cluster if it does not exist in default path. Defaults to $HOME/cloudconfig.
# Mandatory
export CLOUD_CONFIG=$HOME/cloudconfig

# ADLOCATION example is IqDk:US-ASHBURN-AD-1
# Mandatory
export ADLOCATION=""

# KMS key for CMEK testing
# CMEK KEY example "ocid1.key.relm.region.bb..cc.aaa...aa"
# Mandatory
export CMEK_KMS_KEY=""

# NSG Network security group created in cluster's VCN
# CCM E2E tests require two NSGs to run successfully. Please create two NSGs in the cluster's VCN and set NSG_OCIDS
# NSG_OCIDS example ocid1.networksecuritygroup.relm.region.aa...aa,ocid1.networksecuritygroup.relm.region.aa...aa
# Mandatory
export NSG_OCIDS=""

# FSS VOLUME HANDLE in the format filesystem_ocid:mountTargetIP:export_path
# Make sure fss volume handle is in the same subnet as your nodes
# Create a file system, file export path and mount target in your VCN by following
# https://docs.oracle.com/en-us/iaas/Content/File/Tasks/creatingfilesystems.htm#Using_the_Console
# And setup your network for the file system by following:
# https://docs.oracle.com/en-us/iaas/Content/File/Tasks/securitylistsfilestorage.htm
# Mandatory
export FSS_VOLUME_HANDLE=""

# For debugging the tests in existing cluster, do not turn it off by default.
# Optional
# export DELETE_NAMESPACE=false

# By default, public images are used. But if your Cluster's environment cannot access above public images then below option can be used to specify an accessible repo.
# Optional
# export IMAGE_PULL_REPO="accessiblerepo.com/repo/path/"
