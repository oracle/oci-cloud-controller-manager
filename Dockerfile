
FROM iad.ocir.io/odx-oke/oke/golang-buildbox:1.12.4 as builder

ENV SRC /go/src/github.com/oracle/oci-cloud-controller-manager

RUN mkdir -p /go/bin $SRC
ADD . $SRC
WORKDIR $SRC

RUN make build

FROM oraclelinux:7-slim

COPY --from=0 /go/src/github.com/oracle/oci-cloud-controller-manager/dist/oci-cloud-controller-manager /usr/local/bin/
COPY --from=0 /go/src/github.com/oracle/oci-cloud-controller-manager/dist/oci-flexvolume-driver /usr/local/bin/
COPY --from=0 /go/src/github.com/oracle/oci-cloud-controller-manager/dist/oci-volume-provisioner /usr/local/bin/
