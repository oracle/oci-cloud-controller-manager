#!/bin/bash

# Copyright 2018 Oracle and/or its affiliates. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


# Utility functions for managing the version of the CCM running in a Kubernetes
# cluster.
# 
# Functions for deploying the specified version of the CCM in a K8s cluster that 
# that already has a DaemonSet deployed CCM. The version should be defined by the  
# environment variable VERSION.
# 
# Functions for managing a soft (no-threadsafe) lock on the CCM daemonset to 
# minimise it being upgraded/downgraded during integration tests.
#
# Kubectl configured to point to the required target cluster is required on the 
# host machine.


# The root name of the CCM daemon-set and pods.
CCM_NAME="oci-cloud-controller-manager"
# The name of an annotation that can be used to 'lock' the CCM from upgrade/
# downgrade operations. 
CCM_LOCK_LABEL="ccm-deployment-lock"


# Kubernetes Cluster CCM Functions ********************************************
#

function get-k8s-api-server() {
    local res=$(cat $KUBECONFIG | grep 'server:' | awk '{print $2}')
    echo "${res}" 
}

function get-k8s-master() {
    local res=$(kubectl get nodes | grep master | head -n 1 | awk '{print $1}')
    echo "${res}" 
}

function get-ccm-ds-image-version() {
    local res=$(kubectl -n kube-system get ds "${CCM_NAME}" -o=jsonpath="{.spec.template.spec.containers[0].image}")
    echo "${res}"
}

function get-ccm-ds-image() {
    local res=$(get-ccm-ds-image-version | cut -d':' -f 1)
    echo "${res}"
}

function get-ccm-ds-version() {
    local res=$(get-ccm-ds-image-version | cut -d':' -f 2)
    echo "${res}"
}

function get-ccm-ds-json() {
    local res=$(kubectl -n kube-system get ds "${CCM_NAME}" -ojson)
    echo "${res}"
}

function get-ccm-pod-name() {
    local name=$(kubectl -n kube-system get pods | grep oci-cloud-controller-manager | awk '{print $1}')
    echo "${name}"
}

function get-ccm-pod-image-version() {
    local name=$(get-ccm-pod-name)
    local ready=$(kubectl -n kube-system get pod ${name} -o=jsonpath='{.status.containerStatuses[0].image}')
    echo "${ready}"
}

function get-ccm-pod-version() {
    local res=$(get-ccm-pod-image-version | cut -d':' -f 2)
    echo "${res}"
}

function get-ccm-pod-ready() {
    local name=$(get-ccm-pod-name)
    local ready=$(kubectl -n kube-system get pod ${name} -o=jsonpath='{.status.containerStatuses[0].ready}')
    echo "${ready}"
}

function is-ccm-pod-version-ready() {
    local version=$1
    local pod_vsn=$(get-ccm-pod-version)
    if [ "${version}" = "${pod_vsn}" ]; then
        echo $(get-ccm-pod-ready)
    else
        echo "false"
    fi
}

# Wait for the specified ccm pod version to be ready.
function wait-for-ccm-pod-version-ready() {
    local version=$1
    local duration=${2:-60}
    local sleep=${3:-10}
    local timeout=$(($(date +%s) + $duration))
    while [ $(date +%s) -lt $timeout ]; do
        if [ $(is-ccm-pod-version-ready ${version}) = 'true' ]; then
            return 0
        fi
        sleep ${sleep}
    done
    echo "Failed to wait for pod version '${version}' to be ready."
    exit 1 
}

# Kubernetes Manifest CCM Functions *******************************************
#

function get-ccm-manifest-image-version() {
    local manifest=$1
    local res=$(cat "${manifest}" | grep image | awk '{print $2}')
    echo "${res}"
}

function get-ccm-manifest-image() {
    local manifest=$1
    local res=$(get-ccm-manifest-image-version ${manifest} | cut -d':' -f 1)
    echo "${res}"
}

function get-ccm-manifest-version() {
    local manifest=$1
    local res=$(get-ccm-manifest-image-version ${manifest} | cut -d':' -f 2)
    echo "${res}"
}

# Deployment Lock Functions ***************************************************
#

# NB: The date is used to help auto-release a lock that has been placed.
function lock-ccm-deployment() {
    kubectl -n kube-system annotate ds "${CCM_NAME}" "${CCM_LOCK_LABEL}"=$(date +%s)
}

function unlock-ccm-deployment() {
    kubectl -n kube-system annotate ds "${CCM_NAME}" "${CCM_LOCK_LABEL}-"
}

function get-ccm-deployment-lock() {
    local res=$(kubectl -n kube-system get ds ${CCM_NAME} -ojsonpath="{.metadata.annotations.${CCM_LOCK_LABEL}}")
    echo "${res}"
}

function is-ccm-deployment-locked() {
    local res=$(kubectl -n kube-system get ds ${CCM_NAME} -ojsonpath="{.metadata.annotations.${CCM_LOCK_LABEL}}")
    if [ -z "${res}" ]; then
        echo "false"
    else
        echo "true"
    fi
}

# If the CCM has a lock older than hour; the automatically remove it.
function auto-release-lock() {
    local locked=$(get-ccm-deployment-lock)
    if [ ! -z "${locked}" ]; then
        local timeout=$((${locked} + 3600)) 
        local now=$(date +%s)
        if [ $now -gt $timeout ]; then
           unlock-ccm-deployment 
        fi
    fi
}

# Wait for the CCM to have no lock present.
function wait-for-ccm-deployment-permitted() {
    local duration=${1:-3600}
    local sleep=${2:-60}
    local timeout=$(($(date +%s) + $duration))
    while [ $(date +%s) -lt $timeout ]; do
        auto-release-lock
        if [ $(is-ccm-deployment-locked) = 'false' ]; then
            return 0
        fi
        sleep ${sleep}
    done
    echo "Failed to wait for ccm to finish running existing ci pipeline tests."
    exit 1 
}

# Obtain the deployment lock and ensure that all test namespaces have been cleaned up.
function obtain-ccm-deployment-lock() {
    wait-for-ccm-deployment-permitted
    lock-ccm-deployment
    ensure-clean-e2e-test-namespace
}

# Release the deployment lock and ensure that all test namespaces have been cleaned up.
function release-ccm-deployment-lock() {
    ensure-clean-e2e-test-namespace
    unlock-ccm-deployment
}

# Test clean-up Functions *****************************************************
#

function ensure-clean-e2e-test-namespace() {
    local res=$(kubectl get pods --all-namespaces | grep 'cm-e2e-tests-' | awk '{print $1}')
    if [ ! -z ${res} ]; then
        cat ${res} | xargs kubectl delete ns
    fi
}

# Deploy CCM Functions ********************************************************
#

# Upgrade an already deployed CCM to the specified $VERSION.
function deploy-build-version-ccm() {
    local hack_dir=$(cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd)
    local dist_dir=$(dirname "${hack_dir}")/dist
    local build_version_manifest="${dist_dir}/oci-cloud-controller-manager.yaml"
    local rollback_manifest="${dist_dir}/oci-cloud-controller-manager-rollback.yaml"

    local version=$(cat ${dist_dir}/VERSION.txt)
    local build_version_image=$(get-ccm-manifest-image ${build_version_manifest}) 
    local build_version=$(get-ccm-manifest-version ${build_version_manifest})
    local rollback_image=$(get-ccm-ds-image) 
    local rollback_version=$(get-ccm-ds-version)

    # Wait for there to be no lock on CCM deployment; then take the lock. 
    # NB: Not threadsafe, but, better then nothing...
    obtain-ccm-deployment-lock

    # Generate a rollback CCM daemon-set manifest.
    sed s#${rollback_image}:.*#${rollback_image}:${rollback_version}#g < ${build_version_manifest} > ${rollback_manifest}
    
    # Apply the build daemon-set manifest.
    echo "deploying test '${version}' CCM '${build_version_image}:${build_version}' to cluster '$(get-k8s-master)'."
    kubectl apply -f ${build_version_manifest}
    
    # Wait for CCM to be ready...
    wait-for-ccm-pod-version-ready "${build_version}" 
    
    # Display Info
    echo "currently deployed CCM daemon-set version: $(get-ccm-ds-image-version)"
    echo "currently deployed CCM pod version       : $(get-ccm-pod-image-version)"
    echo "currently deployed CCM pod ready state   : $(get-ccm-pod-ready)"
}

# Rollback to the CCM version the cluster originally used before it was upgraded.
function rollback-original-ccm() {
    local hack_dir=$(cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd)
    local dist_dir=$(dirname "${hack_dir}")/dist
    local build_version_manifest="${dist_dir}/oci-cloud-controller-manager.yaml"
    local rollback_manifest="${dist_dir}/oci-cloud-controller-manager-rollback.yaml"
    local rollback_image=$(get-ccm-manifest-image ${rollback_manifest}) 
    local rollback_version=$(get-ccm-manifest-version ${rollback_manifest})

    # Check the rollback manifest exists.
    if [ ! -f ${rollback_manifest} ]; then
        echo "the rollback manifest '${rollback_manifest}' did not exist."
        exit 1
    fi
    
    # Apply original CCM daemon-set manifest.
    echo "rolling back CCM '${rollback_image}:${rollback_version}' to cluster '$(get-k8s-master)'."
    kubectl apply -f ${rollback_manifest}
    
    # Wait for CCM to be ready after rollback...
    wait-for-ccm-pod-version-ready "${rollback_version}" 
    
    # Display Info
    echo "currently deployed CCM daemon-set version: $(get-ccm-ds-image-version)"
    echo "currently deployed CCM pod version       : $(get-ccm-pod-image-version)"
    echo "currently deployed CCM pod ready state   : $(get-ccm-pod-ready)"

    # Release the lock on the CCM deployment mechanism.
    release-ccm-deployment-lock
}


# Main ************************************************************************
#

if [ -z "${KUBECONFIG}" ]; then
    if [ -z "${KUBECONFIG_VAR}" ]; then
        echo "KUBECONFIG or KUBECONFIG_VAR must be set"
        exit 1
    else
        # NB: Wercker environment variables are base64 encoded.
        echo "$KUBECONFIG_VAR" | openssl enc -base64 -d -A > /tmp/kubeconfig
        export KUBECONFIG=/tmp/kubeconfig
    fi
fi

# If provided, execute the specified function.
if [ ! -z "$1" ]; then
    $1
fi

exit $?
