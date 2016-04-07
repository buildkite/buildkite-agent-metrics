
VERSION=$(shell git describe --tags --candidates=1 --dirty --always)
FLAGS=-X main.Version=$(VERSION)
BIN=build/buildkite-metrics

build: $(BIN)

clean:
	-rm build/*

build/buildkite-cloudwatch-metrics:
	-mkdir -p build/
	which glide || go get github.com/Masterminds/glide
	glide install
	go build -o $(BIN) -ldflags="$(FLAGS)" .