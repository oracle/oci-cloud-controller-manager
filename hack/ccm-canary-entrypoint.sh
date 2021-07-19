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

echo "\$METRICS_FILE: ${METRICS_FILE}"
echo "\$MONITOR_PERIOD: ${MONITOR_PERIOD}"

# For OCI usage canary mode is the default
if [ -z "${CANARY_MODE}" ]; then
    export CANARY_MODE="monitor"
fi

pushd "${GOPATH}/src/github.com/oracle/oci-cloud-controller-manager"
./hack/test-canary.sh ${CANARY_MODE}
popd
