#!/bin/bash

##################################################################################################
# This template can be used to tweak the environment variables needed to run the E2E tests locally #
# Default behavior:
# Runs test on an existing cluster in dev0-iad

# To run the tests:
# 1. Change the FOCUS variable here to specify the subset of E2E tests to run
# 2. Set CLUSTER_KUBECONFIG and CLOUD_CONFIG if needed
# 3. run 'source existing-cluster-dev0-env-template.sh' to set the variables
# 4. run 'make run-ccm-e2e-tests-local`
##################################################################################################

# The test suites to run (can replace or add tags)
export FOCUS="\[test1\]"

# The test suites to skip (can replace or add tags)
export FOCUS_SKIP=""

# Run E2Es in parallel. # of ginkgo "nodes" will be decided automatically based on CPU cores.
export ENABLE_PARALLEL_RUN=true

# This variable tells the test not to install oci cli and wipe out your .oci/config
export LOCAL_RUN=1
export TC_BUILD=0

# This allows you to use your existing cluster
export ENABLE_CREATE_CLUSTER=false

# Set path to kubeconfig of existing cluster if it does not exist in default path. Defaults to $HOME/.kube/config_*
export CLUSTER_KUBECONFIG=/Users/ypgohoka/.e2e_ccm_csi/oss-1-31.kubeconfig

# Set path to cloud_config of existing cluster if it does not exist in default path. Defaults to $HOME/cloudconfig_*
export CLOUD_CONFIG=/Users/ypgohoka/go/src/github.com/oracle/oci-cloud-controller-manager/manifests/provider-config-example.yaml


export IMAGE_PULL_REPO="iad.ocir.io/okedev/e2e-tests/"
export ADLOCATION="Ddfp:US-ASHBURN-AD-2"

#KMS key for CMEK testing
export CMEK_KMS_KEY="ocid1.key.oc1.iad.b5r7iu5xaagy2.abuwcljsizwczmdfnitnxaruykluz3p6kwprasd2l7ebfvbkxbytvuumg4cq"

#NSG Network security group created in cluster's VCN
export NSG_OCIDS="ocid1.networksecuritygroup.oc1.iad.aaaaaaaarqaak4vupqsxits6crgsxu5p65eh7p422iry6qttvafn5jvhsmva,ocid1.networksecuritygroup.oc1.iad.aaaaaaaaie4b3quurf3b5sgwz7lwmygii65k3yvlhkyakqacu74xowe3763q"

# NSG Network security group created in cluster's VCN for backend management, this NSG will have to be attached to the nodes manually for tests to pass
export BACKEND_NSG_OCIDS="ocid1.networksecuritygroup.oc1.iad.aaaaaaaaie4b3quurf3b5sgwz7lwmygii65k3yvlhkyakqacu74xowe3763q"

#Reserved IP created in e2e test compartment
export RESERVED_IP="169.155.149.109"

#Architecture to run tests on
export ARCHITECTURE_AMD="AMD"
export ARCHITECTURE_ARM="ARM"

#Focus the tests : ARM, AMD or BOTH
export SCOPE="AMD"

# For debugging the tests in existing cluster, do not turn it off by default.
# Optional
# export DELETE_NAMESPACE=false

# FSS volume handle
# format is FileSystemOCID:serverIP:path
export FSS_VOLUME_HANDLE="ocid1.filesystem.oc1.iad.aaaaaaaaaacdndlxnfqwillqojxwiotjmfsc2ylefuyqaaaa:10.0.73.199:/oss-test"
export FSS_VOLUME_HANDLE_ARM="ocid1.filesystem.oc1.iad.aaaaaaaaaacdndlxnfqwillqojxwiotjmfsc2ylefuyqaaaa:10.0.73.199:/oss-test"

export MNT_TARGET_ID="ocid1.mounttarget.oc1.iad.aaaaacvippzjdfiynfqwillqojxwiotjmfsc2ylefuyqaaaa"
export MNT_TARGET_SUBNET_ID="ocid1.subnet.oc1.iad.aaaaaaaafujcpvwdn3s2liqwrilolm7jmxkwq35zieo7zk4medjtqxjac7cq"
export MNT_TARGET_COMPARTMENT_ID="ocid1.compartment.oc1..aaaaaaaaee2fxlf36idmiqlyvnyhkh2oquz5loogbmzat73hnnqhu2c3352a"

export STATIC_SNAPSHOT_COMPARTMENT_ID="ocid1.compartment.oc1..aaaaaaaaee2fxlf36idmiqlyvnyhkh2oquz5loogbmzat73hnnqhu2c3352a"

# For SKE node, node_info, node_lifecycle controller tests against PDE
# To setup PDE and point your localhost:25000 to the PDE CP API refer: Refer: https://bitbucket.oci.oraclecorp.com/projects/OKE/repos/oke-control-plane/browse/personal-environments/README.md
# export CE_ENDPOINT_OVERRIDE="http://localhost:25000"

# Whether to run UHP E2Es or not, requires Volume Management Plugin enabled on the node and 16+ cores
# Check the following doc for the exact requirements:
# https://docs.oracle.com/en-us/iaas/Content/Block/Concepts/blockvolumeperformance.htm#shapes_block_details
export RUN_UHP_E2E="false"
