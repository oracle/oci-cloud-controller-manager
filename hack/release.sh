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

RELEASE=${RELEASE:=0.1.0}

IMAGE=wcr.io/oracle/oci-cloud-controller-manager

SHA=${SHA:=$(git rev-parse --short=8 HEAD)}

function do_release() {
    if git rev-parse "$RELEASE" >/dev/null 2>&1; then
        echo "Tag $RELEASE already exists. Doing nothing."
        exit 1
    fi

    echo "Creating new release $RELEASE for SHA $SHA"

    git tag -a "$RELEASE" -m "Release version: $RELEASE"
    git push --tags

    docker pull $IMAGE:$SHA
    docker tag $IMAGE:$SHA $IMAGE:$RELEASE
    docker push $IMAGE:$RELEASE
}

read -r -p "Are you sure you want to release ${SHA} as ${RELEASE}? [y/N] " response
case "$response" in
    [yY][eE][sS]|[yY])
        do_release
        ;;
    *)
        exit
        ;;
esac
