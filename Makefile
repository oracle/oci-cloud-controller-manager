include Makefile.include

GOOS ?= linux
ARCH ?= amd64

VERSION := $(shell git describe --always --long --dirty)

RETURN_CODE := $(shell sed --version >/dev/null 2>&1; echo $$?)
ifeq ($(RETURN_CODE),1)
    SED_INPLACE = -i ''
else
    SED_INPLACE = -i
endif

SRC_DIRS := cmd pkg # directories which hold app source (not vendored)

IMAGE := $(REGISTRY)/$(BIN)

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
	@mkdir -p dist/bin/

.PHONY: build
build: build-dirs
	@GOOS=${GOOS} GOARCH=${ARCH} go build     \
	    -i                                    \
	    -o dist/oci-cloud-controller-manager  \
	    -installsuffix "static"               \
	    -ldflags "-X main.version=${VERSION}" \
	    ./cmd/oci-cloud-controller-manager

.PHONY: dist
dist: build-dirs
	@cp -a manifests/* dist
	@sed ${SED_INPLACE}                                            \
	    's#${IMAGE}:[0-9]\+.[0-9]\+.[0-9]\+#${IMAGE}:${VERSION}#g' \
	    dist/oci-cloud-controller-manager.yaml

.PHONY: test
test: build-dirs
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
