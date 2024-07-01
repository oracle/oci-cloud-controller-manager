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

ARG CI_IMAGE_REGISTRY

FROM golang:1.21.5 as builder

ARG COMPONENT

ENV SRC /go/src/github.com/oracle/oci-cloud-controller-manager

ENV GOPATH /go/
RUN mkdir -p /go/bin $SRC
ADD . $SRC
WORKDIR $SRC

RUN COMPONENT=${COMPONENT} make clean build

FROM oraclelinux:8-slim

COPY --from=0 /go/src/github.com/oracle/oci-cloud-controller-manager/dist/* /usr/local/bin/
COPY --from=0 /go/src/github.com/oracle/oci-cloud-controller-manager/image/* /usr/local/bin/

RUN microdnf -y install util-linux e2fsprogs xfsprogs python2 && \
    microdnf update && \
    microdnf clean all

COPY scripts/encrypt-mount /sbin/encrypt-mount
COPY scripts/encrypt-umount /sbin/encrypt-umount
COPY scripts/rpm-host /sbin/rpm-host
COPY scripts/chroot-bash /sbin/chroot-bash
RUN chmod 755 /sbin/encrypt-mount
RUN chmod 755 /sbin/encrypt-umount
RUN chmod 755 /sbin/rpm-host
RUN chmod 755 /sbin/chroot-bash

COPY --from=0 /go/src/github.com/oracle/oci-cloud-controller-manager/dist/* /usr/local/bin/