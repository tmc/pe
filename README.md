# pe

pe is a collection of tools for working with prompt engineering concepts, files, and tools.

## Tools

* [pfutil](cmd/pfutil): operations on promptfoo configuration files


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
go test ./...
```

## Report Issues / Send Patches

https://github.com/tmc/pe/issues.
