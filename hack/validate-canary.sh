#!/bin/bash

# Copyright 2017 Oracle and/or its affiliates. All rights reserved.
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


# A small script to validate the CCM 'Canary' test image works as expected. 
#
# https://confluence.oci.oraclecorp.com/display/BRISTOL/OKE+Canary+Test+Image+Contract


# Helper Functions ************************************************************
#

function get-pod-status() {
    local res=$(kubectl get pod oci-cloud-controller-manager-canary --show-all -ojsonpath="{.status.phase}" 2> /dev/null)
    echo "${res}"
}

# Wait for the CCM canary test pod to reach the specified state within the timeout period. 
function wait-for-canary-pod-state() {
    local state=${1:-"Running"}
    local duration=${2:-60}
    local sleep=${3:-0.5}
    local timeout=$(($(date +%s) + $duration))
    while [ $(date +%s) -lt $timeout ]; do
        local current=$(get-pod-status) 
        echo "waiting for pod oci-cloud-controller-manager-canary state '${state}', currently '${current}'."
        if [ "${current}" = "${state}" ]; then
            return 0
        fi
        sleep ${sleep}
    done
    echo "Failed to wait for oci-cloud-controller-manager-canary state: '${state}'."
    exit 1
} 

# Clean up the CCM canary test pods and associated manifest resources.
function clean-canary() {
    local hack_dir=$(cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd)
    local dist_dir=$(dirname "${hack_dir}")/dist
    local canary_manifest=${dist_dir}/validate-canary-pod.yaml
    local status=$(get-pod-status) # 'Completed' if finished.
    if [ ! -z "${status}" ]; then
        kubectl delete pod oci-cloud-controller-manager-canary
        rm -f ${canary_manifest} 
    fi
}

function get-logs() {
    local type=$1
    kubectl logs oci-cloud-controller-manager-canary -c oci-cloud-controller-manager-canary-${type}
}

function get-test-runner-logs() {
    get-logs test-runner
}

function get-test-reporter-logs() {
    get-logs test-reporter
}

function ensure-cluster-docker-pull-secrets() {
    kubectl create secret docker-registry ocir \
        --docker-server="${OCIREGISTRY}" \
        --docker-username="${OCIRUSERNAME}" \
        --docker-password="${OCIRPASSWORD}" \
        --docker-email="user@example.com"
}

# Shell into the specified canary image via Docker. Useful for debugging the container.
# NB: May have proxy issues for some tests.
function local-docker-mode() {
    local image="iad.ocir.io/oracle/oci-cloud-controller-manager-canary"
    local version="${VERSION}"
    local cid=$(docker run -d -e KUBECONFIG_VAR=$(cat ${KUBECONFIG} | openssl enc -base64 -A) ${image}:${version})
    docker exec -it ${cid} /bin/bash
}


# Test Functions **************************************************************
#

function generate-canary-manifest() {
    local hack_dir=$(cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd)
    local dist_dir=$(dirname "${hack_dir}")/dist
    local canary_manifest=${dist_dir}/validate-canary-pod.yaml
    local version=${VERSION}
    rm -f "${canary_manifest}" 
    cat > "${canary_manifest}" <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: oci-cloud-controller-manager-canary
spec:
  containers:
  - name: oci-cloud-controller-manager-canary-test-runner
    image: iad.ocir.io/oracle/oci-cloud-controller-manager-canary:${version}
    command: ["/bin/bash"]
    args: ["-ec", "make canary"]
    env:
      - name: METRICS_FILE
        value: /metrics/output.json
      - name: KUBECONFIG_VAR
        value: $(cat ${KUBECONFIG} | openssl enc -base64 -A)
    volumeMounts:
    - mountPath: /metrics
      name: metrics-volume
 
  - name: oci-cloud-controller-manager-canary-test-reporter
    image: iad.ocir.io/oracle/oci-cloud-controller-manager-ci-e2e:1.0.1
    command: ["/bin/bash"]
    args: ["-ec", "touch \$METRICS_FILE; while [ -z \$(cat \$METRICS_FILE | grep 'end_time' | cut -d':' -f 1) ]; do sleep 1;  done; cat \$METRICS_FILE"]
    env:
      - name: METRICS_FILE
        value: /metrics/output.json
    volumeMounts:
    - mountPath: /metrics
      name: metrics-volume
  
  imagePullSecrets:
  - name: ocir
 
  volumes:
  - name: metrics-volume
    emptyDir: {}
  restartPolicy: Never
EOF
}

function deploy-canary() {
    local hack_dir=$(cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd)
    local dist_dir=$(dirname "${hack_dir}")/dist
    local canary_manifest=${dist_dir}/validate-canary-pod.yaml
    kubectl apply -f ${canary_manifest} > /dev/null
    wait-for-canary-pod-state
}

function run() {
    # Start a new canary.
    clean-canary
    generate-canary-manifest
    deploy-canary
    # Tail the logs of the reporter to block until it completes. The report only logs the result file.
    res=$(kubectl logs -f oci-cloud-controller-manager-canary -c oci-cloud-controller-manager-canary-test-reporter)
    # Display the results.
    echo "${res}"
    # Grep the log to return an error code.
    error=$(echo "${res}" | grep 'end_time' | cut -d':' -f 1)
    if [ -z ${error} ]; then
        exit 1
    else
        exit 0
    fi
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
if [ -z "${VERSION}" ]; then
    echo "The VERSION must be set"
    exit 1
fi

# If provided, execute the specified function.
if [ ! -z "$1" ]; then
    $1
    exit "$?"
else
    run
fi
