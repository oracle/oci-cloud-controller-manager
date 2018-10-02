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


# should be possible to create and mutate a Service type:LoadBalancer"

# Functions *******************************************************************
#

function run_e2e_tests() {
    echo "Running e2e tests..."
    # Just run the canary for now... we are limite don loadbalancers.
    ginkgo -v -progress \
        test/e2e \
        -- --kubeconfig=${KUBECONFIG} --cloud-config=${CLOUDCONFIG} --delete-namespace=true
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

if [ -z "${CLOUDCONFIG}" ]; then
    if [ -z "${CLOUDCONFIG_VAR}" ]; then
        echo "CLOUDCONFIG or CLOUDCONFIG_VAR must be set"
        exit 1
    else
        # NB: Wercker environment variables are base64 encoded.
        echo "$CLOUDCONFIG_VAR" | openssl enc -base64 -d -A > /tmp/cloudconfig
        export CLOUDCONFIG=/tmp/cloudconfig
    fi
fi

run_e2e_tests

exit $?
