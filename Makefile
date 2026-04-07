GO ?= go
GOLANGCI_LINT ?= golangci-lint
BIN_DIR ?= .bin
BINARY ?= $(BIN_DIR)/codex-switcher
SMOKE_DIR ?= .tmp/smoke

.PHONY: help fmt test vet lint build install smoke clean

help:
	@printf "Targets:\n"
	@printf "  make fmt      Run gofmt on Go sources\n"
	@printf "  make test     Run go test ./...\n"
	@printf "  make vet      Run go vet ./...\n"
	@printf "  make lint     Run golangci-lint if available\n"
	@printf "  make build    Build $(BINARY)\n"
	@printf "  make install  Build codex-switcher into $$HOME/.local/bin\n"
	@printf "  make smoke    Run a local CLI smoke test in $(SMOKE_DIR)\n"
	@printf "  make clean    Remove local build and smoke artifacts\n"

fmt:
	@find . -name '*.go' -not -path './.git/*' -print0 | xargs -0 gofmt -w

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

lint:
	@if command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then \
		$(GOLANGCI_LINT) run ./...; \
	else \
		printf "golangci-lint not found in PATH\n" >&2; \
		exit 1; \
	fi

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BINARY) ./cmd/codex-switcher

install: build
	@mkdir -p $$HOME/.local/bin
	install -m 755 $(BINARY) $$HOME/.local/bin/codex-switcher

smoke: build
	@rm -rf $(SMOKE_DIR)
	@mkdir -p $(SMOKE_DIR)
	@printf '{"email":"smoke@example.com","tokens":{"access":"abc"}}\n' > $(SMOKE_DIR)/auth.json
	$(BINARY) --config-dir $(SMOKE_DIR)/config --auth-file $(SMOKE_DIR)/auth.json save-current
	$(BINARY) --config-dir $(SMOKE_DIR)/config --auth-file $(SMOKE_DIR)/auth.json status
	$(BINARY) --config-dir $(SMOKE_DIR)/config --auth-file $(SMOKE_DIR)/auth.json list
	$(BINARY) --config-dir $(SMOKE_DIR)/config --auth-file $(SMOKE_DIR)/auth.json add --save-current

clean:
	rm -rf $(BIN_DIR) $(SMOKE_DIR)
