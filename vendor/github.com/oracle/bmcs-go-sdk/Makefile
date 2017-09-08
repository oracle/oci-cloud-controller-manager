GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)

default: fmt build

deps:
	go get github.com/kardianos/govendor

fmt:
	gofmt -w $(GOFMT_FILES)

build: test
	@go build github.com/MustWin/baremetal-sdk-go
	@go build github.com/MustWin/baremetal-sdk-go/cmd/obmc-container-clean

test:
	@DEBUG=true go test -v github.com/MustWin/baremetal-sdk-go

acceptance_test:
	@TEST=true go test -v github.com/MustWin/baremetal-sdk-go/acceptance-test

record_acceptance_test:
	@go test -v -timeout 120m -tags "recording all" github.com/MustWin/baremetal-sdk-go/acceptance-test

.PHONY: build test regression_test acceptance_test record_acceptance_test
