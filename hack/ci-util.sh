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

# Utility functions for working with the CCM's Wercker CI.

# Functions *******************************************************************
#

function _check_var() {
    local name=$1
    if [ -z ${!name} ]; then
        echo "WARNING: '""$name""' is required."
    fi
}

# Check the required environnment variables are set when generating a 
# 'cloud-provider.yaml' file. 
function _check_vars() {    
    _check_var OCI_REGION
    _check_var OCI_TENANCY
    _check_var OCI_COMPARTMENT
    _check_var OCI_USER
    _check_var FINGERPRINT
    _check_var PRIVATE_KEY
    _check_var OCI_SUBNET_01 
    _check_var OCI_SUBNET_02
}

# Generate a 'cloud-provider.yaml' file for use in CCM deployment and 
# e2e testing. 
function generate-cloud-provider-config() {
    local file=${1:-"./cloud-provider.yaml"}
    _check_vars 
    cat > $file <<EOF
auth:
  region: $OCI_REGION
  tenancy: $OCI_TENANCY
  compartment: $OCI_COMPARTMENT
  user: $OCI_USER
  key: |
    $PRIVATE_KEY
  fingerprint: $FINGERPRINT
loadBalancer:
  disableSecurityListManagement: false
  subnet1: $OCI_SUBNET_01
  subnet2: $OCI_SUBNET_02
EOF
}

# The Wercker CI platform requires configuration files are base64 encoded.
function base64_encode() {
    local input=$1
    local output=${2:-encoded}
    cat $input | openssl enc -base64 -A > $output
}

function base64_decode() {
    local input=$1
    local output=${2:-decoded}
    cat $input | openssl enc -base64 -d -A > $output
}

# If provided, execute the specified function.
if [ ! -z "$1" ]; then
    $1
else
    generate-cloud-provider-config
fi
