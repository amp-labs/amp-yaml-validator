.PHONY: all
all:
	echo "No default target."

# ====================
# Formatting & linting
# ====================
.PHONY: lint
lint:
	golangci-lint run -c .golangci.yml

# Run a few autoformatters and print out unfixable errors
# PRE-REQUISITES: install linters, see https://ampersand.slab.com/posts/engineering-onboarding-guide-environment-set-up-9v73t3l8#huik9-install-linters
# If you're curious, run `golangci-lint help linters` to see which linters have auto-fix enabled by golangci-lint.
# For ones that do not have auto-fix enabled by golangci-lint (e.g. wsl and gci), we add the fix commands manually to this list.
# For the wsl CLI, we manually run it against select repos, since it does not read from .golangci.yml and therefore cannot ignore directories.
.PHONY: fix
fix:
	wsl --allow-cuddle-declarations --allow-trailing-comment --fix ./... && \
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
