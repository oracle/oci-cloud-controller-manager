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

check-env "CLUSTER_KUBECONFIG"        $CLUSTER_KUBECONFIG
check-env "CLOUD_CONFIG"              $CLOUD_CONFIG
check-env "ADLOCATION"                $ADLOCATION
check-env "NSG_OCIDS"                 $NSG_OCIDS
check-env "BACKEND_NSG_OCIDS"         $BACKEND_NSG_OCIDS
check-env "FSS_VOLUME_HANDLE"         $FSS_VOLUME_HANDLE
check-env "MNT_TARGET_ID"             $MNT_TARGET_ID
check-env "MNT_TARGET_SUBNET_ID"      $MNT_TARGET_SUBNET_ID
check-env "MNT_TARGET_COMPARTMENT_ID" $MNT_TARGET_COMPARTMENT_ID
check-env "ENABLE_PARALLEL_RUN"       $ENABLE_PARALLEL_RUN
check-env "RUN_UHP_E2E"               $RUN_UHP_E2E


function set_image_pull_repo_and_delete_namespace_flag () {
    if [ -z "$IMAGE_PULL_REPO" ]; then
        IMAGE_PULL_REPO=""
    fi

    DELETE_NAMESPACE=${DELETE_NAMESPACE:-"true"}
}

function run_e2e_tests_existing_cluster() {
    if [[ -z "${E2E_NODE_COUNT}" ]]; then
        E2E_NODE_COUNT=1
    fi

    if [ "$ENABLE_PARALLEL_RUN" == "true" ] || [ "$ENABLE_PARALLEL_RUN" == "TRUE" ]; then
        ginkgo -v -p -progress --trace "${FOCUS_OPT}" "${FOCUS_SKIP_OPT}" "${FOCUS_FP_OPT}"  \
                    test/e2e/cloud-provider-oci -- \
                    --cluster-kubeconfig=${CLUSTER_KUBECONFIG} \
                    --cloud-config=${CLOUD_CONFIG} \
                    --adlocation=${ADLOCATION} \
                    --delete-namespace=${DELETE_NAMESPACE} \
                    --image-pull-repo=${IMAGE_PULL_REPO} \
                    --cmek-kms-key=${CMEK_KMS_KEY} \
                    --mnt-target-id=${MNT_TARGET_ID} \
                    --mnt-target-subnet-id=${MNT_TARGET_SUBNET_ID} \
                    --mnt-target-compartment-id=${MNT_TARGET_COMPARTMENT_ID} \
                    --nsg-ocids=${NSG_OCIDS} \
                    --backend-nsg-ocids=${BACKEND_NSG_OCIDS} \
                    --reserved-ip=${RESERVED_IP} \
                    --architecture=${ARCHITECTURE} \
                    --volume-handle=${FSS_VOLUME_HANDLE} \
                    --lustre-volume-handle=${LUSTRE_VOLUME_HANDLE} \
                    --static-snapshot-compartment-id=${STATIC_SNAPSHOT_COMPARTMENT_ID} \
                    --enable-parallel-run=${ENABLE_PARALLEL_RUN} \
                    --run-uhp-e2e=${RUN_UHP_E2E} \
                    --add-oke-system-tags="false" \
                    --maxpodspernode=${MAX_PODS_PER_NODE}
    else
        ginkgo -v -progress --trace -nodes=${E2E_NODE_COUNT} "${FOCUS_OPT}" "${FOCUS_SKIP_OPT}" "${FOCUS_FP_OPT}"  \
            ginkgo -v -p -progress --trace "${FOCUS_OPT}" "${FOCUS_FP_OPT}"  \
                    test/e2e/cloud-provider-oci -- \
                    --cluster-kubeconfig=${CLUSTER_KUBECONFIG} \
                    --cloud-config=${CLOUD_CONFIG} \
                    --adlocation=${ADLOCATION} \
                    --delete-namespace=${DELETE_NAMESPACE} \
                    --image-pull-repo=${IMAGE_PULL_REPO} \
                    --cmek-kms-key=${CMEK_KMS_KEY} \
                    --mnt-target-id=${MNT_TARGET_ID} \
                    --mnt-target-subnet-id=${MNT_TARGET_SUBNET_ID} \
                    --mnt-target-compartment-id=${MNT_TARGET_COMPARTMENT_ID} \
                    --nsg-ocids=${NSG_OCIDS} \
                    --backend-nsg-ocids=${BACKEND_NSG_OCIDS} \
                    --reserved-ip=${RESERVED_IP} \
                    --architecture=${ARCHITECTURE} \
                    --volume-handle=${FSS_VOLUME_HANDLE} \
                    --lustre-volume-handle=${LUSTRE_VOLUME_HANDLE} \
                    --static-snapshot-compartment-id=${STATIC_SNAPSHOT_COMPARTMENT_ID} \
                    --enable-parallel-run=${ENABLE_PARALLEL_RUN} \
                    --run-uhp-e2e=${RUN_UHP_E2E} \
                    --add-oke-system-tags="false" \
                    --maxpodspernode=${MAX_PODS_PER_NODE}
    fi
    retval=$?
    return $retval
}

function set_focus () {
    # The FOCUS environment variable can be set with a regex to tun selected tests
    # e.g. export FOCUS="\[cloudprovider\]"
    # e.g. export FILES="true" && export FOCUS="\[fss_\]" would run E2Es from both fss_dynamic.go and fss_static.go (FOCUS used for file regex instead)
    # e.g. export FOCUS="\[cloudprovider\]" && export FOCUS_SKIP="\[node-update\]" would run all E2Es except ones that have "\[node-update\]" FOCUS.
    export FOCUS_OPT=""
    export FOCUS_FP_OPT=""
    export FOCUS_SKIP_OPT=""
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
            FOCUS_FP_OPT="-regexScansFilePath=${FOCUS}"
        fi
    fi

    # e.g. export FOCUS_SKIP="\[node-update\]" would run all E2Es except ones that have "\[node-update\]" FOCUS.
    # If FOCUS is set as well, all E2Es with FOCUS will run except ones that are covered by SKIP_FOCUS
    if [ ! -z "${FOCUS_SKIP}" ]; then
        # Same for skipping tests with certain FOCUS.
        re1='^\[.+\]$' # [ccm]
        if [[ "${FOCUS_SKIP}" =~ $re1 ]]; then
            echo -E "Escaping square brackes in ${FOCUS_SKIP} to work as a regex match."
            FOCUS_SKIP=$(echo $FOCUS_SKIP|sed -e 's/\[/\\[/g' -e 's/\]/\\]/g')
            echo -E "Modified FOCUS_SKIP value to: ${FOCUS_SKIP}"
        fi

        echo "Skipping focused tests: ${FOCUS_SKIP}"
        FOCUS_SKIP_OPT="-skip=${FOCUS_SKIP}"
    fi
}

echo "CLUSTER_KUBECONFIG is ${CLUSTER_KUBECONFIG}"
echo "CLOUD_CONFIG is ${CLOUD_CONFIG}"
echo "MNT_TARGET_ID is ${MNT_TARGET_ID}"
echo "MNT_TARGET_SUBNET_ID is ${MNT_TARGET_SUBNET_ID}"
echo "MNT_TARGET_COMPARTMENT_ID is ${MNT_TARGET_COMPARTMENT_ID}"

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
