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

# A small script to run the CCM ginkgo 'Canary' e2e tests, and, generate the
# defined canary test response file.
#
# https://confluence.oci.oraclecorp.com/display/BRISTOL/OKE+Canary+Test+Image+Contract

# Functions *******************************************************************
#

function now() {
    echo $(date +"%Y-%m-%d-%H%M%S")
}

# Run the e2e [Canary] tests to produce a gingko result log.
function run_canary_tests() {
    echo "Running canary tests ..."
    ginkgo -v -progress -noColor=true \
        -focus "\[Canary\]" \
        test/e2e/cloud-controller-manager \
        -- --kubeconfig=${KUBECONFIG} --delete-namespace=true \
        2>&1 | tee "${TEST_LOG}"
}

# Extract a {PASSED|FAILED|UNKNOWN} response from a Gingko test log based
# on the specified 'test_matcher'.
function extract_result() {
    local test_matcher=$1
    local result=$(cat "${TEST_LOG}" | grep "${test_matcher}" | tail -n 1 | cut -d ' ' -f 1)
    if [ "${result}" = [Fail] ]; then
        echo "0"
    else
        local passed=$(tail -n 1 "${TEST_LOG}")
        if [ "${passed}" = 'Test Suite Passed' ]; then
            echo "1"
        else
            echo "0"
        fi
    fi
}

# Initialise the result file.
function init_results() {
    local metrics_dir="$(dirname ${METRICS_FILE})"
    mkdir -p "${metrics_dir}"
    echo "Initialising result file: ${METRICS_FILE}"
    cat > "${METRICS_FILE}" <<EOF
{
    "start_time": "${START}"
}
EOF
}

# A set of test_matcher strings that must match the appropriate gingko test
# descriptions. These are used to extract the required test results.
CREATE_LB_TEST="\[It\] should be possible to create and mutate a Service type:LoadBalancer \[Canary\]"
# Creates a JSON result file for the specified [Canary] tests to be extracted.
function create_results() {
    local metrics_dir="$(dirname ${METRICS_FILE})"
    mkdir -p "${metrics_dir}"
    echo "Creating result file: ${METRICS_FILE}"
    cat > "${METRICS_FILE}" <<EOF
{
    "start_time": "${START}",
    "create_lb": "$(extract_result ${CREATE_LB_TEST})",
    "end_time": "$(now)"
}
EOF
}

# Run the tests once and extract the results.
function run-once() {
    START=$(now)
    init_results
    cat "${METRICS_FILE}"
    run_canary_tests
    if [ ! -z "${METRICS_FILE}" ]; then
        create_results
        cat "${METRICS_FILE}"
    fi
}

# Helper function to clean up log and json files.
function clean() {
    echo "ensuring fresh \$START."
    unset START
    echo "ensuring fresh ${TEST_LOG} file."
    rm -f "${TEST_LOG}"
    echo "ensuring fresh ${METRICS_FILE} result file."
    rm -f  "${METRICS_FILE}"
    echo "ensuring all 'cm-e2e-tests' namespaces are terminated."
    local res=$(kubectl get ns | grep 'cm-e2e-tests-' | awk '{print $1}')
    if [ ! -z "${res}" ]; then
        echo ${res} | xargs kubectl delete ns 2> /dev/null
    fi
}

# Run the tests in loop with the specified wait period.
function monitor() {
    local period=${1:-$MONITOR_PERIOD}
    while true;
    do
        clean && run-once
        echo "Sleeping for ${period} before next run..."
        sleep "${period}"
    done
}

# Main ************************************************************************
#

# Handle mandatory KUBECONFIG requirement.
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

# If not specified, default mandatory 'metrics file' location.
if [ -z "${METRICS_FILE}" ]; then
    export METRICS_FILE=/tmp/ccm-canary-metrics.json
fi

# If not specified, default mandatory 'monitor period' in seconds.
if [ -z "${MONITOR_PERIOD}" ]; then
    export MONITOR_PERIOD=30
fi

# Set up directory for filesystem test log. The success of the test
# is extracted from this log.
if [ -z "${TEST_DIR}" ]; then
    TEST_DIR="/tmp"
fi
mkdir -p "${TEST_DIR}"
TEST_LOG="${TEST_DIR}/oci-ccm-canary-test.log"

if [ ! -z "$1" ]; then
    # If provided, execute the specified function with args.
    # e.g. run-once, monitor, clean, etc.
    $@
else
    # Otherwise, run the monitor
    monitor
fi

exit $?
