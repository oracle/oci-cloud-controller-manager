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

REGISTRY := wcr.io/oracle
PKG := github.com/oracle/oci-cloud-controller-manager
BIN := oci-cloud-controller-manager
IMAGE := $(REGISTRY)/$(BIN)


BUILD := $(shell git describe --always --dirty)
# Allow overriding for release versions
# Else just equal the build (git hash)
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
	@./hack/check-gofmt.sh ${SRC_DIRS}

.PHONY: golint
golint:
	@./hack/check-golint.sh ${SRC_DIRS}

.PHONY: govet
govet:
	@./hack/check-govet.sh ${SRC_DIRS}

.PHONY: check
check: gofmt govet golint

.PHONY: build-dirs
build-dirs:
	@mkdir -p dist/

.PHONY: build
build: build-dirs manifests
	@GOOS=${GOOS} GOARCH=${ARCH} go build -v  \
	    -i                                    \
	    -o dist/oci-cloud-controller-manager  \
	    -installsuffix "static"               \
	    -ldflags "-X main.version=${VERSION} -X main.build=${BUILD}" \
	    ./cmd/oci-cloud-controller-manager

.PHONY: manifests
manifests: build-dirs
	@cp -a manifests/* dist
	@sed ${SED_INPLACE}                                            \
	    's#${IMAGE}:[0-9]\+.[0-9]\+.[0-9]\+#${IMAGE}:${VERSION}#g' \
	    dist/oci-cloud-controller-manager.yaml

.PHONY: test
test:
	@./hack/test.sh $(SRC_DIRS)

.PHONY: clean
clean:
	@rm -rf dist

.PHONY: deploy
deploy:
	kubectl -n kube-system set image ds/${BIN} ${BIN}=${IMAGE}:${VERSION}

.PHONY: run-dev
run-dev: build
	dist/oci-cloud-controller-manager             \
	    --kubeconfig=${KUBECONFIG}                \
	    --cloud-config=${CLOUD_PROVIDER_CFG}      \
	    --cluster-cidr=10.244.0.0/16              \
	    --cloud-provider=external                 \
	    -v=4

.PHONY: version
version:
	@echo ${VERSION}
