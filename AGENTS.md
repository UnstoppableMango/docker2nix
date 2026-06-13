# AGENTS.md

This file provides guidance to AI agents when working with code in this repository.

## Commands

```bash
# Build
nix build .#

# Test (all)
ginkgo run -r

# Test (single package)
ginkgo run ./pkg/

# Test (single spec, by name pattern)
ginkgo run -r --focus "pattern"

# Format
nix fmt

# Lint/check
nix flake check

# Regenerate protobuf
buf generate

# Tidy (go.sum + nix/gomod2nix.toml)
make tidy
```

## Architecture

Dockerfile-to-Nix expression converter. CLI built with Cobra; core logic in `pkg/`.

**Data flow:** CLI (`cmd/root.go`) → `pkg.Generate(GenerateRequest)` → `GenerateResponse.Nix`

**Protobuf-driven types:** `GenerateRequest` and `GenerateResponse` are defined in `proto/docker2nix/v1alpha1/generate.proto`, generated into `pkg/docker2nix/v1alpha1/generate.pb.go`. Type aliases live in `pkg/types.go`.

**Output formats:** controlled by `Format` field on `GenerateRequest` (and `--format` CLI flag):

- `docker-tools` (default) — renders `dockerTools.buildLayeredImage`
- `nix2container` — renders `nix2container.buildImage`

**Core function:** `pkg/generate.go:Generate()` — parses a Dockerfile and dispatches to `renderNix` or `renderNix2Container` based on `Format`; extend these as additional Dockerfile instructions are supported.

**Tests:**

- Unit: `pkg/generate_test.go` — tests `Generate()` directly with Dockerfile strings
- E2E: `test/e2e/e2e_test.go` — tests the compiled binary via stdin/file input using `gexec`
- Both suites use Ginkgo v2 + Gomega

**Nix build:** `nix/default.nix` uses `buildGoApplication` (gomod2nix). After adding Go dependencies, run `make tidy` to sync `nix/gomod2nix.toml`.

**Dev shell:** `.envrc` + `flake.nix` provides buf, go, gopls, ginkgo, gnumake, nixfmt, protoc-gen-go.

## Code Style

**Go version:** This project uses Go 1.26.3. Use the current Go language spec; do not assume pre-1.26 behavior. For example, `new(value)` creates a pointer to a copy of `value` and is valid syntax.

**Proto builder pattern.** Assign the builder to a variable named `req`, then call `.Build()` inline at the call site.

```go
// Bad
req := docker2nix.GenerateRequest_builder{Dockerfile: new("FROM ubuntu:24.04\n")}.Build()
Generate(ctx, req)

// Good
req := docker2nix.GenerateRequest_builder{Dockerfile: new("FROM ubuntu:24.04\n")}
Generate(ctx, req.Build())
```
