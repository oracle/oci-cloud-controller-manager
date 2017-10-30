#!/bin/bash

# Edit this to perform a new release
RELEASE=0.1.0

if git rev-parse "$RELEASE" >/dev/null 2>&1; then
    echo "Tag $RELEASE already exists. Doing nothing"
else
    SHA=$(git rev-parse --short=8 HEAD)
    echo "Creating new release $RELEASE for SHA $SHA"
    git tag -a "$RELEASE" -m "Release version: $RELEASE"
    git push --tags
    docker pull wcr.io/oracle/oci-cloud-controller-manager:$SHA
    docker tag wcr.io/oracle/oci-cloud-controller-manager:$SHA wcr.io/oracle/oci-cloud-controller-manager:$RELEASE
    docker push wcr.io/oracle/oci-cloud-controller-manager:$RELEASE
fi
