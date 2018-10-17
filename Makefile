# Makefile based on https://sahilm.com/makefiles-for-golang/

BINARY := buildkite-agent-metrics
VERSION ?= vlatest
HANDLER ?= handler.handle
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
	rm -f $(HANDLER)
	ln release/$(BINARY)-$(VERSION)-linux-amd64 $(HANDLER)
	zip release/$@-$(VERSION).zip $(HANDLER)
	rm $(HANDLER)

.PHONY: release
release: darwin linux $(BINARY)-lambda