
VERSION=$(shell git describe --tags --candidates=1 --dirty --always)
FLAGS=-X main.Version=$(VERSION)
BIN=build/buildkite-metrics-$(shell uname -s)-$(shell uname -m)-$(VERSION)
LATEST=build/buildkite-metrics-$(shell uname -s)-$(shell uname -m)

build: $(BIN)

clean:
	-rm build/*

$(BIN): main.go
	-mkdir -p build/
	go build -o $(BIN) -ldflags="$(FLAGS)" .
	cp -a $(BIN) $(LATEST)

upload: build
	aws s3 sync --acl public-read build s3://buildkite-metrics/