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

# default seclist *************************************************************

function get_default_seclist() {
    get_seclist "${DEFAULT_SECLIST_ID}"
}

function default_seclist_egress_rules() {
   read -r -d '' rules <<EOF
[
    {
    "destination": "0.0.0.0/0",
    "icmp-options": null,
    "is-stateless": null,
    "protocol": "all",
    "tcp-options": null,
    "udp-options": null
    }
]
EOF
    echo $rules
}

function default_seclist_ingress_rules() {
   read -r -d '' rules <<EOF
[
    {
    "icmp-options": null,
    "is-stateless": null,
    "protocol": "6",
    "source": "0.0.0.0/0",
    "tcp-options": {
        "destination-port-range": {
        "max": 22,
        "min": 22
        },
        "source-port-range": null
    },
    "udp-options": null
    },
    {
    "icmp-options": {
        "code": 4,
        "type": 3
    },
    "is-stateless": null,
    "protocol": "1",
    "source": "0.0.0.0/0",
    "tcp-options": null,
    "udp-options": null
    },
    {
    "icmp-options": {
        "code": null,
        "type": 3
    },
    "is-stateless": null,
    "protocol": "1",
    "source": "10.0.0.0/16",
    "tcp-options": null,
    "udp-options": null
    }
]
EOF
    echo $rules
}

function default_seclist_reset() {
    if [ -z "${DEFAULT_SECLIST_ID}" ]; then
        echo "No \$DEFAULT_SECLIST_ID was configured for the operation. Please set carefully."
        exit 1
    fi
    oci network security-list update \
        --security-list-id "${DEFAULT_SECLIST_ID}" \
        --egress-security-rules "$(default_seclist_egress_rules)" \
        --ingress-security-rules "$(default_seclist_ingress_rules)" \
        --config-file "${OCI_CONFIG}"
}

# etcd seclist ****************************************************************

function get_etcd_seclist() {
    get_seclist "${ETCD_SECLIST_ID}" 
}

function etcd_seclist_egress_rules() {
   read -r -d '' rules <<EOF
[
    {
    "destination": "0.0.0.0/0",
    "icmp-options": null,
    "is-stateless": false,
    "protocol": "all",
    "tcp-options": null,
    "udp-options": null
    }
]
EOF
    echo $rules
}

function etcd_seclist_ingress_rules() {
   read -r -d '' rules <<EOF
[
    {
    "icmp-options": {
        "code": 4,
        "type": 3
    },
    "is-stateless": false,
    "protocol": "1",
    "source": "0.0.0.0/0",
    "tcp-options": null,
    "udp-options": null
    },
    {
    "icmp-options": {
        "code": 4,
        "type": 3
    },
    "is-stateless": false,
    "protocol": "1",
    "source": "10.0.0.0/16",
    "tcp-options": null,
    "udp-options": null
    },
    {
    "icmp-options": null,
    "is-stateless": false,
    "protocol": "6",
    "source": "129.144.0.0/12",
    "tcp-options": null,
    "udp-options": null
    },
    {
    "icmp-options": null,
    "is-stateless": false,
    "protocol": "6",
    "source": "129.213.0.0/16",
    "tcp-options": null,
    "udp-options": null
    },
    {
    "icmp-options": null,
    "is-stateless": false,
    "protocol": "6",
    "source": "10.0.0.0/16",
    "tcp-options": null,
    "udp-options": null
    },
    {
    "icmp-options": null,
    "is-stateless": false,
    "protocol": "6",
    "source": "0.0.0.0/0",
    "tcp-options": {
        "destination-port-range": {
        "max": 22,
        "min": 22
        },
        "source-port-range": null
    },
    "udp-options": null
    },
    {
    "icmp-options": null,
    "is-stateless": false,
    "protocol": "6",
    "source": "10.0.0.0/16",
    "tcp-options": {
        "destination-port-range": {
        "max": 2380,
        "min": 2379
        },
        "source-port-range": null
    },
    "udp-options": null
    }
]
EOF
    echo $rules
}

function etcd_seclist_reset() {
    if [ -z "${ETCD_SECLIST_ID}" ]; then
        echo "No \$ETCD_SECLIST_ID was configured for the operation. Please set carefully."
        exit 1
    fi
    oci network security-list update \
        --security-list-id "${ETCD_SECLIST_ID}" \
        --egress-security-rules "$(etcd_seclist_egress_rules)" \
        --ingress-security-rules "$(etcd_seclist_ingress_rules)" \
        --config-file "${OCI_CONFIG}"
}

# k8s_ccm seclist *************************************************************

function get_k8s_ccm_seclist() {
    get_seclist "${K8S_CCM_SECLIST_ID}" 
}

function k8s_ccm_seclist_egress_rules() {
   read -r -d '' rules <<EOF
[
    {
    "destination": "0.0.0.0/0",
    "icmp-options": null,
    "is-stateless": false,
    "protocol": "all",
    "tcp-options": null,
    "udp-options": null
    }
]
EOF
    echo $rules
}

function k8s_ccm_seclist_ingress_rules() {
   read -r -d '' rules <<EOF
[]
EOF
    echo $rules
}

function k8s_ccm_seclist_reset() {
    if [ -z "${K8S_CCM_SECLIST_ID}" ]; then
        echo "No \$K8S_CCM_SECLIST_ID was configured for the operation. Please set carefully."
        exit 1
    fi
    oci network security-list update \
        --security-list-id "${K8S_CCM_SECLIST_ID}" \
        --egress-security-rules "$(k8s_ccm_seclist_egress_rules)" \
        --ingress-security-rules "$(k8s_ccm_seclist_ingress_rules)" \
        --config-file "${OCI_CONFIG}"
}

# k8s_master seclist **********************************************************

function get_k8s_master_seclist() {
    get_seclist "${K8S_MASTER_SECLIST_ID}" 
}

function k8s_master_seclist_egress_rules() {
   read -r -d '' rules <<EOF
[
    {
    "destination": "0.0.0.0/0",
    "icmp-options": null,
    "is-stateless": false,
    "protocol": "all",
    "tcp-options": null,
    "udp-options": null
    }
]
EOF
    echo $rules
}

function k8s_master_seclist_ingress_rules() {
   read -r -d '' rules <<EOF
[
    {
    "icmp-options": {
        "code": 4,
        "type": 3
    },
    "is-stateless": false,
    "protocol": "1",
    "source": "0.0.0.0/0",
    "tcp-options": null,
    "udp-options": null
    },
    {
    "icmp-options": {
        "code": 4,
        "type": 3
    },
    "is-stateless": false,
    "protocol": "1",
    "source": "10.0.0.0/16",
    "tcp-options": null,
    "udp-options": null
    },
    {
    "icmp-options": null,
    "is-stateless": false,
    "protocol": "6",
    "source": "129.144.0.0/12",
    "tcp-options": null,
    "udp-options": null
    },
    {
    "icmp-options": null,
    "is-stateless": false,
    "protocol": "6",
    "source": "129.213.0.0/16",
    "tcp-options": null,
    "udp-options": null
    },
    {
    "icmp-options": null,
    "is-stateless": false,
    "protocol": "all",
    "source": "10.0.0.0/16",
    "tcp-options": null,
    "udp-options": null
    },
    {
    "icmp-options": null,
    "is-stateless": false,
    "protocol": "6",
    "source": "0.0.0.0/0",
    "tcp-options": {
        "destination-port-range": {
        "max": 22,
        "min": 22
        },
        "source-port-range": null
    },
    "udp-options": null
    },
    {
    "icmp-options": null,
    "is-stateless": false,
    "protocol": "6",
    "source": "10.0.0.0/16",
    "tcp-options": {
        "destination-port-range": {
        "max": 8080,
        "min": 8080
        },
        "source-port-range": null
    },
    "udp-options": null
    },
    {
    "icmp-options": null,
    "is-stateless": false,
    "protocol": "6",
    "source": "0.0.0.0/0",
    "tcp-options": {
        "destination-port-range": {
        "max": 443,
        "min": 443
        },
        "source-port-range": null
    },
    "udp-options": null
    },
    {
    "icmp-options": null,
    "is-stateless": false,
    "protocol": "6",
    "source": "10.0.0.0/16",
    "tcp-options": {
        "destination-port-range": {
        "max": 32767,
        "min": 30000
        },
        "source-port-range": null
    },
    "udp-options": null
    }
]
EOF
    echo $rules
}

function k8s_master_seclist_reset() {
    if [ -z "${K8S_MASTER_SECLIST_ID}" ]; then
        echo "No \$K8S_MASTER_SECLIST_ID was configured for the operation. Please set carefully."
        exit 1
    fi
    oci network security-list update \
        --security-list-id "${K8S_MASTER_SECLIST_ID}" \
        --egress-security-rules "$(k8s_master_seclist_egress_rules)" \
        --ingress-security-rules "$(k8s_master_seclist_ingress_rules)" \
        --config-file "${OCI_CONFIG}"
}

# k8s_worker secList **********************************************************

function get_k8s_worker_seclist() {
    get_seclist "${K8S_WORKER_SECLIST_ID}" 
}

function k8s_worker_seclist_egress_rules() {
   read -r -d '' rules <<EOF
[
    {
    "destination": "0.0.0.0/0",
    "icmp-options": null,
    "is-stateless": false,
    "protocol": "all",
    "tcp-options": null,
    "udp-options": null
    }
]
EOF
    echo $rules
}

function k8s_worker_seclist_ingress_rules() {
   read -r -d '' rules <<EOF
[
    {
    "icmp-options": {
        "code": 4,
        "type": 3
    },
    "is-stateless": false,
    "protocol": "1",
    "source": "0.0.0.0/0",
    "tcp-options": null,
    "udp-options": null
    },
    {
    "icmp-options": {
        "code": 4,
        "type": 3
    },
    "is-stateless": false,
    "protocol": "1",
    "source": "10.0.0.0/16",
    "tcp-options": null,
    "udp-options": null
    },
    {
    "icmp-options": null,
    "is-stateless": false,
    "protocol": "6",
    "source": "129.144.0.0/12",
    "tcp-options": null,
    "udp-options": null
    },
    {
    "icmp-options": null,
    "is-stateless": false,
    "protocol": "6",
    "source": "129.213.0.0/16",
    "tcp-options": null,
    "udp-options": null
    },
    {
    "icmp-options": null,
    "is-stateless": false,
    "protocol": "all",
    "source": "10.0.0.0/16",
    "tcp-options": null,
    "udp-options": null
    },
    {
    "icmp-options": null,
    "is-stateless": false,
    "protocol": "6",
    "source": "0.0.0.0/0",
    "tcp-options": {
        "destination-port-range": {
        "max": 22,
        "min": 22
        },
        "source-port-range": null
    },
    "udp-options": null
    },
    {
    "icmp-options": null,
    "is-stateless": false,
    "protocol": "6",
    "source": "0.0.0.0/0",
    "tcp-options": {
        "destination-port-range": {
        "max": 32767,
        "min": 30000
        },
        "source-port-range": null
    },
    "udp-options": null
    }
]
EOF
    echo $rules
}

function k8s_worker_seclist_reset() {
    if [ -z "${K8S_WORKER_SECLIST_ID}" ]; then
        echo "No \$K8S_WORKER_SECLIST_ID was configured for the operation. Please set carefully."
        exit 1
    fi
    oci network security-list update \
        --security-list-id "${K8S_WORKER_SECLIST_ID}" \
        --egress-security-rules "$(k8s_worker_seclist_egress_rules)" \
        --ingress-security-rules "$(k8s_worker_seclist_ingress_rules)" \
        --config-file "${OCI_CONFIG}"
}

# support ***********************************************************************

function get_seclist() {
    local seclist_id=$1
    oci network security-list get \
        --security-list-id "${seclist_id}" \
        --config-file "${OCI_CONFIG}" 
}

# Reset the default seclists between the loadbalancer and the nodes.
# 
function loadbalancer_tunnel_seclist_reset() {
    k8s_ccm_seclist_reset
    k8s_worker_seclist_reset
}

# Reset all seclists. 
# 
function all_seclist_reset() {
    default_seclist_reset
    etcd_seclist_reset
    k8s_ccm_seclist_reset
    k8s_master_seclist_reset
    k8s_worker_seclist_reset
}

# main ************************************************************************

if [ -z "${OCI_CONFIG}" ]; then
    echo "No \$OCI_CONFIG was configured for the target tenancy."
    exit 1
fi

$@

