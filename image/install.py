#!/usr/bin/env python

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

import argparse
from string import Template
import subprocess
import os.path
import os
from shutil import copyfile
import base64
import select

parser = argparse.ArgumentParser()
parser.add_argument("-c", "--cloud-config", dest="config", default="/etc/oci/cloud-provider.yaml")
parser.add_argument("-d", "--driver-mount", dest="driver_mount", default="/flexmnt")

options = parser.parse_args()

VENDOR = "oracle"
DRIVER = "oci"
DRIVER_EXEC_PATH= "/usr/local/bin/oci-flexvolume-driver"

DRIVER_DIRECTORY = "{}/{}~{}".format(options.driver_mount, VENDOR, DRIVER)

LOG_FILE = "{}/oci_flexvolume_driver.log".format(DRIVER_DIRECTORY)


def create_driver_directory():
    if not os.path.isdir(DRIVER_DIRECTORY):
        os.mkdir(DRIVER_DIRECTORY)

def copy_driver_binary():
    #Copy executable atomically
    copyfile(DRIVER_EXEC_PATH, "{}/.{}".format(DRIVER_DIRECTORY, DRIVER))
    os.rename("{}/.{}".format(DRIVER_DIRECTORY, DRIVER), "{}/{}".format(DRIVER_DIRECTORY, DRIVER))
    os.chmod("{}/{}".format(DRIVER_DIRECTORY, DRIVER), 0755)

def generate_kubeconfig():
    script_path = os.path.abspath(os.path.dirname(__file__))
    template_path = os.path.join(script_path, "kubeconfig.yml.template")
    with open(template_path, "r") as template_file, open("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt", "r") as ca_file, open("/var/run/secrets/kubernetes.io/serviceaccount/token", "r") as token_file: 
        template = Template(template_file.read())
        result = template.substitute({
            "ca" : base64.b64encode(ca_file.read()),
            "token" : token_file.read(),
            "server" : "https://{}:{}".format(os.getenv("KUBERNETES_SERVICE_HOST", "0.0.0.0"), os.getenv("KUBERNETES_SERVICE_PORT", "443"))
        })
    with open("{}/kubeconfig".format(DRIVER_DIRECTORY),"w+") as kubeconfig:
        kubeconfig.write(result)

def create_log():
    with open(LOG_FILE, "w+") as log:
        log.write("---OCI FLEXVOLUME DRIVER---\n")

def tail_log():
    log_process = subprocess.Popen(['tail', '-F', LOG_FILE], stdout=subprocess.PIPE,stderr=subprocess.PIPE)
    while True:
        print log_process.stdout.readline()

def copy_config_to_driver_dir():
    if os.path.isfile(options.config):
        copyfile(options.config, "{}/config.yaml".format(DRIVER_DIRECTORY))
    else:
        with open(LOG_FILE, "w+") as log:
            log.write("Could not copy configuration from {}. Assuming worker node\n".format(options.config))

create_driver_directory()
copy_driver_binary()
create_log()
copy_config_to_driver_dir()
generate_kubeconfig()
tail_log()
