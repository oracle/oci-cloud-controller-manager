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


PKG := github.com/oracle/oci-cloud-controller-manager

ifeq "$(CI_IMAGE_REGISTRY)" ""
    CI_IMAGE_REGISTRY   ?= iad.ocir.io/oracle
else
    CI_IMAGE_REGISTRY   ?= ${CI_IMAGE_REGISTRY}
endif

ifeq "$(OSS_REGISTRY)" ""
    OSS_REGISTRY   ?= iad.ocir.io/oracle
else
    OSS_REGISTRY   ?= ${OSS_REGISTRY}
endif
IMAGE ?= $(OSS_REGISTRY)/cloud-provider-oci
COMPONENT ?= oci-cloud-controller-manager oci-volume-provisioner oci-flexvolume-driver oci-csi-controller-driver oci-csi-node-driver

ifeq "$(VERSION)" ""
    BUILD := $(shell git describe --exact-match 2> /dev/null || git describe --match=$(git rev-parse --short=8 HEAD) --always --dirty --abbrev=8)
    # Allow overriding for release versions else just equal the build (git hash)
    VERSION ?= ${BUILD}
else
    VERSION   ?= ${VERSION}
endif


GOOS ?= linux
ARCH ?= amd64

SRC_DIRS := cmd pkg # directories which hold app source (not vendored)

# Allows overriding where the CCM should look for the cloud provider config
# when running via make run-dev.
CLOUD_PROVIDER_CFG ?= $$(pwd)/cloud-provider.yaml

RETURN_CODE := $(shell sed --version >/dev/null 2>&1; echo $$?)
ifeq ($(RETURN_CODE),1)
    SED_INPLACE = -i ''
else
    SED_INPLACE = -i
endif

.PHONY: all
all: check test build

.PHONY: gofmt
gofmt:
	@./hack/check-gofmt.sh $(SRC_DIRS)

.PHONY: golint
golint:
	@./hack/check-golint.sh $(SRC_DIRS)

.PHONY: govet
govet:
	@./hack/check-govet.sh $(SRC_DIRS)

.PHONY: check
check: gofmt govet golint

.PHONY: build-dirs
build-dirs:
	@mkdir -p dist/

.PHONY: build
build: build-dirs
	@for component in $(COMPONENT); do \
		GOOS=$(GOOS) GOARCH=$(ARCH) CGO_ENABLED=0 go build -o dist/$$component -ldflags "-X main.version=$(VERSION) -X main.build=$(BUILD)" ./cmd/$$component ; \
    done

.PHONY: manifests
manifests: build-dirs
	@cp -a manifests/**/*.yaml dist
	@sed $(SED_INPLACE)                         \
	  's#${IMAGE}:latest#${IMAGE}:${VERSION}#g' \
	  dist/*.yaml

.PHONY: vendor
vendor:
	@GO111MODULE=on go mod vendor -v

.PHONY: test
test:
	@./hack/test.sh $(SRC_DIRS)

# Run the canary tests - in single run mode.
.PHONY: canary-run-once
canary-run-once:
	@./hack/test-canary.sh run-once

# Run the canary tests - in monitor (infinite loop) mode.
.PHONY: canary-monitor
canary-monitor:
	@./hack/test-canary.sh monitor

# Validate the generated canary test image. Runs test once
# and monitors from sidecar.
.PHONY: validate-canary
validate-canary:
	@./hack/validate-canary.sh run

.PHONY: clean
clean:
	@rm -rf dist

.PHONY: run-ccm-dev
run-ccm-dev:
	@go run cmd/oci-cloud-controller-manager/main.go  \
	  --kubeconfig=$(KUBECONFIG)                      \
	  --cloud-config=$(CLOUD_PROVIDER_CFG)            \
	  --cluster-cidr=10.244.0.0/16                    \
	  --leader-elect-resource-lock=configmaps         \
	  --cloud-provider=oci                            \
	  -v=4

.PHONY: run-volume-provisioner-dev
run-volume-provisioner-dev:
	@NODE_NAME=$(shell hostname)                      \
	CONFIG_YAML_FILENAME=cloud-provider.yaml          \
	go run cmd/oci-volume-provisioner/main.go         \
	    --kubeconfig=$(KUBECONFIG)                    \
	    -v=4

.PHONY: image
BUILD_ARGS = --build-arg CI_IMAGE_REGISTRY="$(CI_IMAGE_REGISTRY)" --build-arg COMPONENT="$(COMPONENT)"
image:
	docker  build $(BUILD_ARGS) \
		-t $(IMAGE):$(VERSION) .

.PHONY: push
push: image
	docker login --username="${oss_docker_username}" --password="${oss_docker_password}" $(OSS_REGISTRY)
	docker push $(IMAGE):$(VERSION)

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: build-local
build-local: build

.PHONY: run-ccm-e2e-tests-local
run-ccm-e2e-tests-local:
	./hack/run_e2e_test.sh

