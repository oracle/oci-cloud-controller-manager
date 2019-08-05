PKG_NAME=httpsigner
PKG=oracle.com/oci/$(PKG_NAME)
WORKING_DIR=/go/src/$(PKG)
DOCKER_REPO=odo-docker-local.artifactory.oci.oraclecorp.com

# Transition (To be removed when Jenkins jobs and GHE hooks get disabled)
DYN_REPO=artifacts.dynguts.net:5001
COVERAGE_IMAGE=$(DYN_REPO)/gocov:1.1.1-dyn-glide-0.10.2-glide-0.12.3-golang-1.8.3

# Docker Images
PLEX_CI_IMAGE=$(DOCKER_REPO)/plex-go-ci:golang-1.10.8

# Docker Options
WORKING_DIR_OPTS=-v $(CURDIR):$(WORKING_DIR) -w $(WORKING_DIR) -e WORKDIR=$(WORKING_DIR)

VENDOR_OPTS=$(WORKING_DIR_OPTS) $(PLEX_CI_IMAGE)

COVERAGE_OPTS=-v $(HOME)/.netrc:/root/.netrc:ro $(WORKING_DIR_OPTS) $(COVERAGE_IMAGE)

JSON_LOG_OPTS=--log-driver=json-file --log-opt max-size=1m

LINTER_OPTS=$(WORKING_DIR_OPTS) $(JSON_LOG_OPTS) $(PLEX_CI_IMAGE)

LINTER_CMD=gometalinter \
		   --deadline=120s \
		   --enable-gc \
		   --disable-all \
		   --enable=vet \
		   --enable=vetshadow \
		   --enable=golint \
		   --enable=ineffassign \
		   --enable=goconst \
		   --enable=gofmt \
		   --enable=misspell \
		   --exclude=/usr/local/go/src \
		   --vendor \
		   --tests \
		   ./...


# Pull dependencies, run tests, and lint for library
.PHONY: all
all: vendor unittest lint

# Install necessary host-system libraries
.PHONY: bootstrap
bootstrap:
	# Install linting tools
	go get -u github.com/golang/lint/golint

# run unittests and output a coverage report to display in the default browser using
# the standard gotest coverage reporting (NOTE: not used for jenkins job, but useful for local coverage testing)
.PHONY: go-coverage
go-coverage:
	docker run --rm --name $(PKG_NAME)_unittest $(VENDOR_OPTS) \
		go test $$(glide novendor) -coverprofile=coverage.cov \
			&& go tool cover -html=coverage.cov

# run coverage report for cobertura plugin
.PHONY: coverage
coverage:
	docker run --rm $(COVERAGE_OPTS)

# Runs the godoc server and open it in the default browser
.PHONY: godoc
godoc:
	open "http://localhost:6060/pkg/${PKG}/" && \
		godoc -http :6060

# run gometalinter against the source
.PHONY: lint
lint:
	docker run --rm $(LINTER_OPTS) $(LINTER_CMD)

# recipe for prb jobs. effectively, build the binary, run the unittests, report coverage and
# lint the code
.PHONY: prb
prb: vendor unittest-report-xml coverage lint

### Testing
.PHONY: unittest
unittest:
	docker run --rm --name $(PKG_NAME)_unittest $(VENDOR_OPTS) \
		go test $$(glide novendor) -cover

# run unittests and generate jenkins compatible xml report
.PHONY: unittest-report-xml
unittest-report-xml:
	docker run --rm --name $(PKG_NAME)_unittest --entrypoint='sh' $(VENDOR_OPTS) \
		-c 'go get -u github.com/jstemmer/go-junit-report && \
		go test -v $$(glide novendor) | go-junit-report > unittest.xml'

# Install glide dependencies
.PHONY: vendor
vendor:
	docker run --rm $(VENDOR_OPTS) glide install


### TeamCity

.PHONY: prb-tc
prb-tc: vendor coverage-tc lint

# Run unit tests with 'go tool' coverage report
.PHONY: unittest-tc
unittest-tc:
	docker run --rm --name $(PKG_NAME)_unittest $(VENDOR_OPTS) \
		go test -v $$(glide novendor) -coverprofile=coverage.out

# Report Coverage Metrics for TeamCity
# "##teamcity[buildStatisticValue key='CodeCoverageS' value='«percentage»']
.PHONY: coverage-tc
coverage-tc: unittest-tc
	docker run --rm --name $(PKG_NAME)_coverage $(VENDOR_OPTS) go tool cover -func=coverage.out
	@echo "##teamcity[buildStatisticValue key='CodeCoverageS' value='$(shell docker \
		run --rm --name $(PKG_NAME)_coverage $(VENDOR_OPTS) go tool cover -func=coverage.out | tail -1 | grep -Eo "[0-9]+\.[0-9]+")']"


### debugging targets

.PHONY: build_debug
build_debug:
	docker run -it --rm $(VENDOR_OPTS) /bin/bash

.PHONY: build_env
build_env:
	docker run --rm $(VENDOR_OPTS) /usr/bin/env
