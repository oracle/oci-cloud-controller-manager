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

ifeq "$(CI_IMAGE_REGISTRY)" ""
    CI_IMAGE_REGISTRY   ?= "iad.ocir.io/oracle"
else
    CI_IMAGE_REGISTRY   ?= ${CI_IMAGE_REGISTRY}
endif
IMAGE ?= $(CI_IMAGE_REGISTRY)/oci-kube-ci

ifndef VERSION
  $(error VERSION is undefined)
endif

.PHONY: build
build:
	docker build \
	  --build-arg=http_proxy \
	  --build-arg=https_proxy \
	  -f Dockerfile \
	  -t ${IMAGE}:${VERSION} \
	  .

.PHONY: push
push: build
	docker login --username="${ci_docker_username}" --password="${ci_docker_password}" $(CI_IMAGE_REGISTRY)
	docker push $(IMAGE):$(VERSION)

