include Makefile.include

# Which architecture to build - see $(ALL_ARCH) for options.
ARCH ?= amd64


ifdef GITLAB_CI
    # Gitlab CI doesn't allocate a TTY
    DOCKER_OPS_INTERACTIVE = -t
ifdef CI_COMMIT_TAG
    VERSION := ${CI_COMMIT_TAG}
else
    VERSION := ${CI_COMMIT_SHA}
endif
else
    VERSION := ${USER}-$(shell  date +%Y%m%d%H%M%S)
    DOCKER_OPS_INTERACTIVE = -ti
endif

BASEIMAGE = oraclelinux:7.3

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
    DOCKER_VOLUME_CONSISTENCY_CONFIG = :delegated
else
    DOCKER_VOLUME_CONSISTENCY_CONFIG =
endif

RETURN_CODE := $(shell sed --version >/dev/null 2>&1; echo $$?)
ifeq ($(RETURN_CODE),1)
    SED_INPLACE = -i ''
else
    SED_INPLACE = -i
endif

###
### These variables should not need tweaking.
###

SRC_DIRS := cmd pkg # directories which hold app source (not vendored)

IMAGE := $(REGISTRY)/$(BIN)

BUILD_IMAGE ?= golang:1.9-alpine

# If you want to build all binaries, see the 'all-build' rule.
# If you want to build all containers, see the 'all-container' rule.
# If you want to build AND push all containers, see the 'all-push' rule.
.PHONY: all
all: build

build: dist/bin/$(ARCH)/$(BIN)

.PHONY: dist
dist: build-dirs
	@echo "Building manifests"
	@cp -a manifests/* dist
	@find dist/ -type f -name '*.yaml' -exec sed ${SED_INPLACE} 's#{{VERSION}}#${VERSION}#g' {} +
	@find dist/ -type f -name '*.yaml' -exec  sed ${SED_INPLACE} 's#{{REGISTRY}}#${REGISTRY}#g' {} +

dist/bin/$(ARCH)/$(BIN): build-dirs
	echo "building: $@"
	@docker run                                                                                                \
	    ${DOCKER_OPS_INTERACTIVE}                                                                              \
	    --rm                                                                                                   \
	    -u $$(id -u):$$(id -g)                                                                                 \
	    -v "$$(pwd)/.go:/go$(DOCKER_VOLUME_CONSISTENCY_CONFIG)"                                                \
	    -v "$$(pwd):/go/src/$(PKG)$(DOCKER_VOLUME_CONSISTENCY_CONFIG)"                                         \
	    -v "$$(pwd)/dist/bin/$(ARCH):/go/bin"                                                                  \
	    -v "$$(pwd)/dist/bin/$(ARCH):/go/bin/linux_$(ARCH)$(DOCKER_VOLUME_CONSISTENCY_CONFIG)"                 \
	    -v "$$(pwd)/.go/std/$(ARCH):/usr/local/go/pkg/linux_$(ARCH)_static$(DOCKER_VOLUME_CONSISTENCY_CONFIG)" \
	    -w /go/src/$(PKG)                                                                                      \
	    $(BUILD_IMAGE)                                                                                         \
	    /bin/sh -x -c "                                                                                        \
	        ARCH=$(ARCH)                                                                                       \
	        VERSION=$(VERSION)                                                                                 \
	        PKG=$(PKG)                                                                                         \
	        ./hack/build.sh                                                                                    \
	    "

.PHONY: shell
# Example: make shell CMD="-c 'date > datefile'"
shell: build-dirs
	@echo "launching a shell in the containerized build environment"
	@docker run                                                             \
	    ${DOCKER_OPS_INTERACTIVE}                                           \
	    --rm                                                                \
	    -u $$(id -u):$$(id -g)                                              \
	    -v "$$(pwd)/.go:/go"                                                \
	    -v "$$(pwd):/go/src/$(PKG)"                                         \
	    -v "$$(pwd)/dist/bin/$(ARCH):/go/bin"                               \
	    -v "$$(pwd)/dist/bin/$(ARCH):/go/bin/linux_$(ARCH)"                 \
	    -v "$$(pwd)/.go/std/$(ARCH):/usr/local/go/pkg/linux_$(ARCH)_static" \
	    -w /go/src/$(PKG)                                                   \
	    $(BUILD_IMAGE)                                                      \
	    /bin/sh $(CMD)

DOTFILE_IMAGE = $(subst :,_,$(subst /,_,$(IMAGE))-$(VERSION))

.PHONY: container
container: .container-$(DOTFILE_IMAGE) container-name

.container-$(DOTFILE_IMAGE): dist/bin/$(ARCH)/$(BIN) Dockerfile.in dist
	@sed \
	    -e 's|ARG_BIN|$(BIN)|g' \
	    -e 's|ARG_ARCH|$(ARCH)|g' \
	    -e 's|ARG_FROM|$(BASEIMAGE)|g' \
	    Dockerfile.in > .dockerfile-$(ARCH)
	docker build -t $(IMAGE):$(VERSION) -f .dockerfile-$(ARCH) .
	docker images -q $(IMAGE):$(VERSION) > $@

.PHONY: container-name
container-name:
	@echo "container: $(IMAGE):$(VERSION)"

.PHONY: push
push: .push-$(DOTFILE_IMAGE) push-name

.push-$(DOTFILE_IMAGE): .container-$(DOTFILE_IMAGE)
	@docker login -u '$(DOCKER_REGISTRY_USERNAME)' -p '$(DOCKER_REGISTRY_PASSWORD)' $(REGISTRY)
ifeq ($(findstring gcr.io,$(REGISTRY)),gcr.io)
	@gcloud docker -- push $(IMAGE):$(VERSION)
else
	@docker push $(IMAGE):$(VERSION)
endif
	@docker images -q $(IMAGE):$(VERSION) > $@

.PHONY: push-name
push-name:
	@echo "pushed: $(IMAGE):$(VERSION)"

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: test
test: build-dirs
	@docker run                                                             \
	    ${DOCKER_OPS_INTERACTIVE}                                           \
	    --rm                                                                \
	    -u $$(id -u):$$(id -g)                                              \
	    -v "$$(pwd)/.go:/go"                                                \
	    -v "$$(pwd):/go/src/$(PKG)"                                         \
	    -v "$$(pwd)/dist/bin/$(ARCH):/go/bin"                               \
	    -v "$$(pwd)/.go/std/$(ARCH):/usr/local/go/pkg/linux_$(ARCH)_static" \
	    -w /go/src/$(PKG)                                                   \
	    $(BUILD_IMAGE)                                                      \
	    /bin/sh -c "./hack/test.sh $(SRC_DIRS)"

.PHONY: e2e
e2e:
	echo "TODO: running e2e tests..."
	#@go test -v ./test/e2e/ --kubeconfig=${KUBECONFIG}

.PHONY: build-dirs
build-dirs:
	@mkdir -p dist/bin/$(ARCH)
	@mkdir -p .go/src/$(PKG) .go/pkg .go/bin .go/std/$(ARCH)

.PHONY: clean
clean: container-clean bin-clean

.PHONY: container-clean
container-clean:
	rm -rf .container-* .dockerfile-* .push-* dist

.PHONY: bin-clean
bin-clean:
	rm -rf .go dist/bin

.PHONY: deploy
deploy: push
	kubectl -n bmcs set image ds/${BIN}-ds ${BIN}=${IMAGE}:${VERSION}

.PHONY: run-dev
run-dev:
	@go run cmd/$(BIN)/main.go                    \
	    --kubeconfig=${KUBECONFIG}                \
	    --v=4                                     \
	    --cloud-config=$$(pwd)/cloud-provider.cfg \
	    --cluster-cidr=10.244.0.0/16              \
	    --cloud-provider=external                 \
	    -v=4
