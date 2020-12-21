# default to vendor mod, since our minimal supported version of Go is
# 1.11
GOFLAGS ?= "-mod=vendor"
GO111MODULE ?= "on"

all: build-restd

build-%:
	cd cmd/$* ; \
	export GO111MODULE=$(GO111MODULE) ; \
	go build $(GOFLAGS) -ldflags "-X main.Version=$(shell git describe --tags --always --long --dirty)"

lint:
	GO111MODULE=off go get -u golang.org/x/lint/golint
	$(shell go env GOPATH)/bin/golint -set_exit_status $(shell go list $(GOFLAGS) ./...)

.PHONY: build lint
