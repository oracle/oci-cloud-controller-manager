
FROM iad.ocir.io/odx-oke/oke/golang-buildbox:1.12.7-fips as builder

ARG COMPONENT

ENV SRC /go/src/github.com/oracle/oci-cloud-controller-manager

ENV GOPATH /go/
RUN mkdir -p /go/bin $SRC
ADD . $SRC
WORKDIR $SRC

RUN COMPONENT=${COMPONENT} make clean build

FROM oraclelinux:7-slim

COPY --from=0 /go/src/github.com/oracle/oci-cloud-controller-manager/dist/* /usr/local/bin/

RUN yum install -y iscsi-initiator-utils-6.2.0.874-10.0.5.el7 \
 && yum install -y e2fsprogs \
 && yum clean all
