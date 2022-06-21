#!/bin/bash

# Copyright 2020 Oracle and/or its affiliates. All rights reserved.
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

function check-env () {
    if [ -z "$2" ]; then
        echo "$1 must be defined"
        exit 1
    fi
}

check-env "CLUSTER_KUBECONFIG"    $CLUSTER_KUBECONFIG
check-env "CLOUD_CONFIG"          $CLOUD_CONFIG
check-env "ADLOCATION"      $ADLOCATION
check-env "NSG_OCIDS"       $NSG_OCIDS
check-env "FSS_VOLUME_HANDLE"       $FSS_VOLUME_HANDLE

function set_image_pull_repo_and_delete_namespace_flag () {
    if [ -z "$IMAGE_PULL_REPO" ]; then
        IMAGE_PULL_REPO=""
    fi

    DELETE_NAMESPACE=${DELETE_NAMESPACE:-"true"}
}

function run_e2e_tests_existing_cluster() {
    ginkgo -v -progress --trace "${FOCUS_OPT}" "${FOCUS_FP_OPT}"  \
        test/e2e/cloud-provider-oci -- \
        --cluster-kubeconfig=${CLUSTER_KUBECONFIG} \
        --cloud-config=${CLOUD_CONFIG} \
        --adlocation=${ADLOCATION} \
        --delete-namespace=${DELETE_NAMESPACE} \
        --image-pull-repo=${IMAGE_PULL_REPO} \
        --cmek-kms-key=${CMEK_KMS_KEY} \
        --nsg-ocids=${NSG_OCIDS} \
        --reserved-ip=${RESERVED_IP} \
        --architecture=${ARCHITECTURE} \
        --volume-handle=${FSS_VOLUME_HANDLE}
    retval=$?
    return $retval
}

function set_focus () {
    # The FOCUS environment variable can be set with a regex to tun selected tests
    # e.g. export FOCUS="\[cloudprovider\]"
    export FOCUS_OPT=""
    export FOCUS_FP_OPT=""
    if [ ! -z "${FOCUS}" ]; then
        # Because we tag our test descriptions with tags that are surrounded
        # by square brackets, we have to escape the brackets when we set the
        # FOCUS variable to match on a bracket rather than have it interpreted
        # as a regex character class. The check below looks to see if the FOCUS
        # has square brackets which aren't yet escaped and fixes them if needed.
        re1='^\[.+\]$' # [ccm]
        if [[ "${FOCUS}" =~ $re1 ]]; then
            echo -E "Escaping square brackes in ${FOCUS} to work as a regex match."
            FOCUS=$(echo $FOCUS|sed -e 's/\[/\\[/g' -e 's/\]/\\]/g')
            echo -E "Modified FOCUS value to: ${FOCUS}"
        fi

        echo "Running focused tests: ${FOCUS}"
        FOCUS_OPT="-focus=${FOCUS}"

        # The FILES environment variable can be defined to interpret the regex as a
        # set of files.
        # e.g. export FILES="true"
        if [[ ! -z "${FILES}" && "${FILES}" == "true" ]]; then
            echo "Running focused test regex as filepath expression."
            FOCUS_FP_OPT="-regexScansFilePath=${FILES}"
        fi
    fi
}

echo "CLUSTER_KUBECONFIG is ${CLUSTER_KUBECONFIG}"
echo "CLOUD_CONFIG is ${CLOUD_CONFIG}"

function run_tests () {
    set_image_pull_repo_and_delete_namespace_flag
    set_focus
    # run the ginko test framework for existing cluster
    # run ARM tests
    if [[ "$SCOPE" == "BOTH" || "$SCOPE" == "ARM" ]]; then
        run_e2e_tests_existing_cluster
        retval_arm=$?
    fi
    # run AMD tests
    if [[ "$SCOPE" == "BOTH" || "$SCOPE" == "AMD" ]]; then
        run_e2e_tests_existing_cluster
        retval_amd=$?
    fi

    RED='\033[0;31m'
    NC='\033[0m' # No Color
    if [[ "$SCOPE" == "BOTH" ]]; then
        if [[ $retval_amd == 0 && $retval_arm == 0 ]]; then
            printf "ARM and AMD tests are Successful!"
            return $retval_amd
        fi

        if [[ $retval_amd != 0 ]]; then
            printf "${RED}AMD Failed${NC}"
            return $retval_amd
        fi

        if [[ $retval_arm != 0 ]]; then
            printf "${RED}ARM Failed${NC}"
            return $retval_arm
        fi
    fi

    if [[ "$SCOPE" == "ARM" ]]; then
        if [[ $retval_arm != 0 ]]; then
            printf "${RED}ARM Failed${NC}"
            return $retval_arm
        else
            echo "ARM tests are Successful"
            return $retval_arm
        fi
    fi

    if [[ "$SCOPE" == "AMD" ]]; then
        if [[ $retval_amd != 0 ]]; then
            printf "${RED}AMD Failed${NC}"
            return $retval_amd
        else
            echo "AMD tests are Successful"
            return $retval_amd
        fi
    fi
}

run_tests
