# Generate Go types from OpenAPI spec.

ifndef OPENAPI_COMMIT_ID
$(error OPENAPI_COMMIT_ID is not set. Please set it to a valid commit hash or "main" for the latest version)
endif
CONFIG_YAML = https://raw.githubusercontent.com/amp-labs/openapi/$(OPENAPI_COMMIT_ID)/config/config.yaml
MANIFEST_YAML = https://raw.githubusercontent.com/amp-labs/openapi/$(OPENAPI_COMMIT_ID)/manifest/manifest.yaml

.PHONY: gen
gen: gen/config gen/manifest
	echo "{\"commit_id\": \"${OPENAPI_COMMIT_ID}\"}" > commit.json

.PHONY: gen/config
gen/config:
	oapi-codegen --config=oapi.config.yaml --o=config.gen.go $(CONFIG_YAML)

.PHONY: gen/manifest
gen/manifest:
	oapi-codegen --config=oapi.config.yaml --o=manifest.gen.go $(MANIFEST_YAML)

.PHONY: all
all:
	echo "No default target."

.PHONY: clean
clean:
	echo "No clean target."

.PHONY: test
test:
	echo "No test target."
