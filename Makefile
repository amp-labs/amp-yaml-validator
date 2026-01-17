.PHONY: help
.DEFAULT_GOAL := help

# Show available targets
help:
	@echo "amp-yaml-validator - Available targets:"
	@echo ""
	@echo "  test            Run all tests"
	@echo "  test-pretty     Run tests with pretty output (gotestsum)"
	@echo ""
	@echo "  build           Build the CLI tool"
	@echo "  build-all       Build for all platforms (linux, darwin, windows)"
	@echo ""
	@echo "  lint            Run linters without auto-fix"
	@echo "  fix             Run formatters and linters with auto-fix"
	@echo "  fix/sort        Run fix with sorted output"
	@echo "  fix-markdown    Fix markdown files"
	@echo "  format          Alias for 'fix'"
	@echo ""
	@echo "  help            Show this help message"

# ====================
# Formatting & linting
# ====================
.PHONY: lint
lint:
	golangci-lint config verify && \
	golangci-lint run -c .golangci.yml --max-issues-per-linter 0 --max-same-issues 0 && \
	typos --config .typos.toml

# Run a few autoformatters and print out unfixable errors
# PRE-REQUISITES: install linters, see https://ampersand.slab.com/posts/engineering-onboarding-guide-environment-set-up-9v73t3l8#huik9-install-linters
# If you're curious, run `golangci-lint help linters` to see which linters have auto-fix enabled by golangci-lint.
# For ones that do not have auto-fix enabled by golangci-lint (e.g. wsl and gci), we add the fix commands manually to this list.
# For the wsl CLI, we manually run it against select repos, since it does not read from .golangci.yml and therefore cannot ignore directories.
.PHONY: fix
fix:
	wsl --fix ./... && \
		gci write . && \
		golangci-lint run -c .golangci.yml --fix

.PHONY: fix-markdown
fix-markdown:
	markdownlint --fix .

.PHONY: fix/sort
fix/sort:
	make fix | grep "" | sort

# Alias for fix
.PHONY: format
format: fix


.PHONY: test
test:
	go test -v ./...

.PHONY: test-pretty
test-pretty:
	RUNNING_ENV=test go run gotest.tools/gotestsum@latest -- -v -gcflags="all=-N -l" ./...

# ====================
# Build
# ====================
.PHONY: build
build:
	go build -o amp-yaml-validator ./cmd/amp-yaml-validator

.PHONY: build-all
build-all:
	GOOS=linux GOARCH=amd64 go build -o amp-yaml-validator-linux-amd64 ./cmd/amp-yaml-validator
	GOOS=darwin GOARCH=amd64 go build -o amp-yaml-validator-darwin-amd64 ./cmd/amp-yaml-validator
	GOOS=darwin GOARCH=arm64 go build -o amp-yaml-validator-darwin-arm64 ./cmd/amp-yaml-validator
	GOOS=windows GOARCH=amd64 go build -o amp-yaml-validator-windows-amd64.exe ./cmd/amp-yaml-validator
