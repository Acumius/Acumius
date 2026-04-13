GO ?= go
GO_PACKAGES := ./...
GOCACHE ?= $(CURDIR)/.gocache

.PHONY: fmt fmt-check lint test check

fmt:
	GOCACHE=$(GOCACHE) $(GO) fmt $(GO_PACKAGES)

fmt-check:
	@files="$$(find . -type f -name '*.go' -not -path './vendor/*' -not -path './.gocache/*')"; \
	if [ -z "$$files" ]; then \
		echo "No Go files to check."; \
		exit 0; \
	fi; \
	unformatted="$$(gofmt -l $$files)"; \
	if [ -n "$$unformatted" ]; then \
		echo "Unformatted files detected:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi; \
	echo "Formatting check passed."

lint:
	GOCACHE=$(GOCACHE) $(GO) vet $(GO_PACKAGES)

test:
	GOCACHE=$(GOCACHE) $(GO) test $(GO_PACKAGES)

check: fmt-check lint test
