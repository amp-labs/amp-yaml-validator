# OpenAPI Definitions

This package contains Go types generated from OpenAPI specs defined in <https://github.com/amp-labs/openapi>.

## Generating Go Types

Generated Go types live in `*.gen.go` files in this directory.

If you haven't yet, you'll first need to install oapi-codegen:

```shell
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
```

To regenerate the Go structs, run:

```shell
make gen
```

## Updating types

- Make changes to the OpenAPI specs in the `openapi` repo.
- Run `make gen/<yaml>` in this repo to generate Go types.
- Create a PR in this repo with the changes.
