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

RUN yum install --enablerepo=ol7_developer_EPEL -y \
    ca-certificates \
    gcc \
    git \
    jq \
    make \
    openssl \
    pwgen \
    python \
    python-pip \
    python-yaml \
    unzip && \
    yum clean all && rm -rf /var/cache/yum

# Install golang environment
RUN curl https://storage.googleapis.com/golang/go1.11.1.linux-amd64.tar.gz -O && \
    mkdir /tools && \
    tar xzf go1.11.1.linux-amd64.tar.gz -C /tools && \
    rm go1.11.1.linux-amd64.tar.gz && \
    mkdir -p /go/bin

ENV PATH=/tools/go/bin:/go/bin:/tools/linux-amd64:$PATH \
    GOPATH=/go \
    GOROOT=/tools/go

# Install the kubectl client
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.11.0/bin/linux/amd64/kubectl && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/local/bin/kubectl

# Install Ginkgo
RUN go get -u github.com/onsi/ginkgo/ginkgo && \
    go get -u github.com/onsi/gomega/...

# Install golint
RUN go get -u golang.org/x/lint/golint

# Install Terraform
RUN curl https://releases.hashicorp.com/terraform/0.10.7/terraform_0.10.7_linux_amd64.zip -LO && \
    unzip terraform_0.10.7_linux_amd64.zip && \
    mv terraform /usr/bin && \
    rm -f terraform terraform_0.10.7_linux_amd64.zip

# Installs the OCI terraform provider
RUN curl -LO https://github.com/oracle/terraform-provider-oci/releases/download/2.0.2/linux.tar.gz && \
    tar -xvf linux.tar.gz -C / && \
    echo "providers { oci = \"/linux_amd64/terraform-provider-oci_v2.0.2\" }" > ~/.terraformrc && \
    rm -f linux.tar.gz

# Install OCI client
RUN pip install \
    oci \
    requests[security]
