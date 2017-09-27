REGISTRY := registry.oracledx.com/skeppare
PKG := github.com/oracle/oci-cloud-controller-manager
BIN := oci-cloud-controller-manager
IMAGE := $(REGISTRY)/$(BIN)
VERSION := $(shell git describe --always --dirty)

GOOS ?= linux
ARCH ?= amd64

SRC_DIRS := cmd pkg # directories which hold app source (not vendored)

RETURN_CODE := $(shell sed --version >/dev/null 2>&1; echo $$?)
ifeq ($(RETURN_CODE),1)
    SED_INPLACE = -i ''
else
    SED_INPLACE = -i
endif

.PHONY: all
all: build

.PHONY: gofmt
gofmt:
	@./hack/check-gofmt.sh ${SRC_DIRS}

.PHONY: golint
golint:
	@./hack/check-golint.sh ${SRC_DIRS}

.PHONY: govet
govet:
	@./hack/check-govet.sh ${SRC_DIRS}

.PHONY: build-dirs
build-dirs:
	@mkdir -p dist/

.PHONY: build
build: build-dirs
	@GOOS=${GOOS} GOARCH=${ARCH} go build     \
	    -i                                    \
	    -o dist/oci-cloud-controller-manager  \
	    -installsuffix "static"               \
	    -ldflags "-X main.version=${VERSION}" \
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
	@rm -rf dist/*

.PHONY: deploy
deploy:
	kubectl -n kube-system set image ds/${BIN} ${BIN}=${IMAGE}:${VERSION}

.PHONY: run-dev
run-dev:
	@go run cmd/$(BIN)/main.go                    \
	    --kubeconfig=${KUBECONFIG}                \
	    --v=4                                     \
	    --cloud-config=$$(pwd)/cloud-provider.cfg \
	    --cluster-cidr=10.244.0.0/16              \
	    --cloud-provider=external                 \
	    -v=4
