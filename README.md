# nillinter

A Go static analysis tool that flags comparisons of slice values to `nil`, suggesting the use of `len(s) == 0` or `len(s) != 0` instead for clearer and more correct emptiness checks.

## Overview

`nillinter` detects when slices are compared to `nil` and suggests replacing these comparisons with length checks. This is a style preference that can help make code more explicit about checking for empty slices rather than nil slices.

**Flags:**

- `s == nil` → suggests `len(s) == 0`
- `s != nil` → suggests `len(s) != 0`

## Installation

### Standalone

```bash
go install github.com/amitaifrey/nillinter/cmd/nillinter@latest
```

### As a golangci-lint Plugin

`nillinter` can be used as a plugin with [golangci-lint](https://golangci-lint.run/). The plugin is automatically registered when the module is imported.

## Usage

### Standalone

Run `nillinter` on your Go code:

```bash
nillinter ./...
```

### With golangci-lint

Add `nillinter` to your `.golangci.yml` configuration:

```yaml
linters:
  enable:
    - nillinter
```

## Examples

### Before

```go
var s []string

if s == nil {
    // handle empty case
}

if s != nil {
    // handle non-empty case
}
```

### After

```go
var s []string

if len(s) == 0 {
    // handle empty case
}

if len(s) != 0 {
    // handle non-empty case
}
```

## Ignoring Specific Lines

If you need to keep a nil comparison for a specific reason, you can use the ignore directive:

```go
//nillinter:ignore
if s == nil {
    // this line will not be flagged
}
```

## Note

This linter represents a style opinion. In Go, `nil` slices and empty slices behave similarly in most contexts, but they are technically different. If your codebase relies on distinguishing between `nil` and empty slices, you may want to disable this linter or use the ignore directive for specific cases.

## License

See [LICENSE](LICENSE) file for details.
