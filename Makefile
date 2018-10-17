# Makefile based on https://sahilm.com/makefiles-for-golang/

BINARY := buildkite-metrics
VERSION ?= vlatest
PLATFORMS := linux darwin
os = $(word 1, $@)

.PHONY: clean
clean:
	rm -f release/*

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	mkdir -p release
	GOOS=$(os) GOARCH=amd64 go build -o release/$(BINARY)-$(VERSION)-$(os)-amd64

.PHONY: $(BINARY)-lambda
$(BINARY)-lambda: linux
	rm -f buildkite-agent
	ln release/$(BINARY)-$(VERSION)-linux-amd64 buildkite-agent
	zip release/$@-$(VERSION).zip buildkite-agent
	rm buildkite-agent

.PHONY: release
release: darwin linux $(BINARY)-lambda