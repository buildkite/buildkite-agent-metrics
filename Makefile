
VERSION=$(shell git fetch --tags && git describe --tags --candidates=1 --dirty --always)
FLAGS=-X main.Version=$(VERSION)
BIN=build/buildkite-metrics-$(shell uname -s)-$(shell uname -m)-$(VERSION)
LAMBDA_ZIP=$(BIN)-lambda.zip
SRC=$(shell find . -name '*.go')

build: $(BIN)

build-lambda: $(LAMBDA_ZIP)

clean:
	-rm -f build/

$(BIN): $(SRC)
	-mkdir -p build/
	go build -o $(BIN) -ldflags="$(FLAGS)" .

$(LAMBDA_ZIP): $(SRC)
	docker run --rm -v $(GOPATH):/go -v $(PWD):/tmp eawsy/aws-lambda-go
	mv handler.zip $(LAMBDA_ZIP)

upload:
	aws s3 sync --acl public-read build s3://buildkite-metrics/
