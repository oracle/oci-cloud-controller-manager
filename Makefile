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
IMAGE ?= iad.ocir.io/spinnaker/cloud-provider-oci

BUILD := $(shell git describe --exact-match 2> /dev/null || git describe --match=$(git rev-parse --short=8 HEAD) --always --dirty --abbrev=8)
# Allow overriding for release versions else just equal the build (git hash)
VERSION ?= ${BUILD}

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

.PHONY: oci-cloud-controller-manager
oci-cloud-controller-manager: build-dirs
	@GOOS=$(GOOS) GOARCH=$(ARCH) go build     \
	    -i                                    \
	    -o dist/oci-cloud-controller-manager  \
	    -installsuffix "static"               \
	    -ldflags "-X main.version=$(VERSION) -X main.build=$(BUILD)" \
	    ./cmd/oci-cloud-controller-manager

.PHONY: oci-volume-provisioner
oci-volume-provisioner: build-dirs
	@GOOS=$(GOOS) GOARCH=$(ARCH) CGO_ENABLED=0 go build                                    \
	-i                                                                                     \
	-o dist/oci-volume-provisioner                                                         \
	-ldflags="-s -w -X main.version=${VERSION} -X main.build=${BUILD} -extldflags -static" \
	./cmd/oci-volume-provisioner


.PHONY: oci-flexvolume-driver
oci-flexvolume-driver: build-dirs
	@GOOS=$(GOOS) GOARCH=$(ARCH) CGO_ENABLED=0 go build                    \
	    -i                                                                 \
	    -o dist/oci-flexvolume-driver                                      \
	    -ldflags="-s -w -X main.version=$(VERSION) -X main.build=$(BUILD)" \
	    ./cmd/oci-flexvolume-driver/

.PHONY: build
build: oci-cloud-controller-manager oci-volume-provisioner oci-flexvolume-driver

.PHONY: manifests
manifests: build-dirs
	@cp -a manifests/**/*.yaml dist
	@sed $(SED_INPLACE)                                            \
	    's#${IMAGE}:[0-9]\+.[0-9]\+.[0-9]\+#${IMAGE}:${VERSION}#g' \
	    dist/oci-cloud-controller-manager.yaml

.PHONY: test
test:
	@./hack/test.sh $(SRC_DIRS)

# Deploys the current version to a specified cluster.
# Requires binary, manifests, images to be built and pushed. Requires $KUBECONFIG set.
.PHONY: upgrade
upgrade:
	# Upgrade the current CCM to the specified version
	@./hack/deploy.sh deploy-build-version-ccm

.PHONY: rollback
rollback:
	@./hack/deploy.sh delete-ccm-ds

.PHONY: e2e
e2e:
	@./hack/test-e2e.sh

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

.PHONY: run-dev
run-dev: build
	@dist/oci-cloud-controller-manager          \
	    --kubeconfig=$(KUBECONFIG)              \
	    --cloud-config=$(CLOUD_PROVIDER_CFG)    \
	    --cluster-cidr=10.244.0.0/16            \
	    --leader-elect-resource-lock=configmaps \
	    --cloud-provider=oci                    \
	    -v=4

.PHONY: version
version:
	@echo $(VERSION)
