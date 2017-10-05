FROM oraclelinux:7-slim

COPY dist/oci-cloud-controller-manager /

USER nobody:nobody

ENTRYPOINT [ "/oci-cloud-controller-manager" ]
