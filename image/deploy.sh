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

set -o errexit
set -o pipefail

VENDOR=oracle
DRIVER=oci

driver_dir="/flexmnt/$VENDOR${VENDOR:+"~"}${DRIVER}"

LOG_FILE="$driver_dir/oci_flexvolume_driver.log"

config_file_name="config.yaml"
kubeconfig_file_name="kubeconfig"
config_tmp_dir="/tmp"
kubeconfig_tmp_dir="/tmp2"

CONFIG_FILE="$config_tmp_dir/$config_file_name"

KUBECONFIG_FILE="$kubeconfig_tmp_dir/$kubeconfig_file_name"

if [ ! -d "$driver_dir" ]; then
  mkdir "$driver_dir"
fi

cp "/$DRIVER" "$driver_dir/.$DRIVER"
mv -f "$driver_dir/.$DRIVER" "$driver_dir/$DRIVER"

if [ -f "$CONFIG_FILE" ]; then
  cp  "$CONFIG_FILE"  "$driver_dir/$config_file_name"
fi

if [ -f "$KUBECONFIG_FILE" ]; then
  cp  "$KUBECONFIG_FILE"  "$driver_dir/$kubeconfig_file_name"
fi

while : ; do
  touch $LOG_FILE
  tail -f $LOG_FILE
done
