GOPATH := $(shell pwd)/.gopath
STAGING_PATH = $(shell pwd)/.gopath/src/github.com/wolfeidau
DEPS = $(go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)

all: deps test

deps:
	mkdir -p ${STAGING_PATH}
	ln -s $(shell pwd) ${STAGING_PATH}/go-buildkite || true
	cd ${STAGING_PATH}/go-buildkite
	go get -d -v ./...

test: deps
	cd ${STAGING_PATH}/go-buildkite
	go test -timeout=3s -v ./...

docker:
	docker run --rm -v "$(shell pwd)":/go-buildkite -w /go-buildkite golang:1.5 make

clean:
	rm -rf $(shell pwd)/.gopath || true

.PHONY: all deps test clean docker
