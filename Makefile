
VERSION=$(shell git describe --tags --candidates=1 --dirty --always)
FLAGS=-X main.Version=$(VERSION)
BIN=build/buildkite-metrics

build: $(BIN)

clean:
	-rm build/*

$(BIN):
	-mkdir -p build/
	go build -o $(BIN) -ldflags="$(FLAGS)" .