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

FROM oraclelinux:7-slim

RUN yum install oracle-softwarecollection-release-el7 -y \
    ca-certificates \
    gcc \
    git \
    jq \
    make \
    openssl \
    pwgen \
    python3 \
    wget \
    python-yaml \
    unzip && \
    yum clean all && rm -rf /var/cache/yum

RUN wget https://bootstrap.pypa.io/pip/3.6/get-pip.py
RUN python3 get-pip.py

# Install golang environment
RUN curl https://storage.googleapis.com/golang/go1.15.12.linux-amd64.tar.gz -O && \
    mkdir /tools && \
    tar xzf go1.15.12.linux-amd64.tar.gz -C /tools && \
    rm go1.15.12.linux-amd64.tar.gz && \
    mkdir -p /go/bin

ENV PATH=/tools/go/bin:/go/bin:/tools/linux-amd64:$PATH \
    GOPATH=/go \
    GOROOT=/tools/go

# Install the kubectl client
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.19.0/bin/linux/amd64/kubectl && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/local/bin/kubectl

# Install OCI client
RUN pip install \
    oci \
    requests[security]
