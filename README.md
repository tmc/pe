# pe

pe is a collection of tools for working with prompt engineering concepts, files, and tools.

## Tools

* [pfutil](cmd/pfutil): operations on promptfoo configuration files
  * `vet`: validate promptfoo configuration files
  * `fmt`: format promptfoo configuration files
  * `convert`: convert promptfoo configuration files between formats (yaml, json)
  * `run`: execute promptfoo configuration files against LLM providers (with mock implementations)
* Scripts:
  * `scripts/test-convert.sh`: simple test for format conversion roundtrips
  * `scripts/test-semantic-equality.sh`: comprehensive test for data integrity during format conversions


## Specifications

* [prompt_engineering.proto](./prompt_engineering.proto): Protobuf definitions for prompt
  engineering concepts.

## Download/Install

The easiest way to install is to run `go install github.com/tmc/pe/cmd/...@latest`.
You can also clone the repository and run `go install ./cmd/...`.

## Contribute

Contributions to pe are welcome. Before submitting a pull request, please make
sure tests pass:

```
# Run Go tests
go test ./...

# Test the pfutil tool's format conversion functionality
./scripts/test-semantic-equality.sh
```

## Report Issues / Send Patches

https://github.com/tmc/pe/issues.
