# docker2nix

[![CI](https://github.com/UnstoppableMango/docker2nix/actions/workflows/ci.yml/badge.svg)](https://github.com/UnstoppableMango/docker2nix/actions/workflows/ci.yml)
[![Last commit](https://img.shields.io/github/last-commit/UnstoppableMango/docker2nix)](https://github.com/UnstoppableMango/docker2nix/commits/main)
[![Go Report Card](https://goreportcard.com/badge/github.com/unstoppablemango/docker2nix)](https://goreportcard.com/report/github.com/unstoppablemango/docker2nix)
[![Go Reference](https://pkg.go.dev/badge/github.com/unstoppablemango/docker2nix.svg)](https://pkg.go.dev/github.com/unstoppablemango/docker2nix)
[![codecov](https://codecov.io/gh/UnstoppableMango/docker2nix/graph/badge.svg)](https://codecov.io/gh/UnstoppableMango/docker2nix)
[![GitHub release](https://img.shields.io/github/v/release/UnstoppableMango/docker2nix)](https://github.com/UnstoppableMango/docker2nix/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Convert Dockerfiles to Nix expressions for [nixpkgs `dockerTools`](https://nixos.org/manual/nixpkgs/stable/#ssec-pkgs-dockerTools-buildLayeredImage) or [nix2container](https://github.com/nlewo/nix2container).

> [!WARNING]
> Very WIP and, at this point, mostly vibed

## Usage

```sh
# From stdin
cat Dockerfile | docker2nix

# From file
docker2nix ./Dockerfile

# nix2container output
docker2nix --format nix2container ./Dockerfile
```

### Formats

| Flag value | Output function |
| --- | --- |
| `docker-tools` (default) | `dockerTools.buildLayeredImage { ... }` |
| `nix2container` | `nix2container.buildImage { ... }` |

### Example

Input:

```dockerfile
FROM ubuntu:24.04
ENV FOO=bar
CMD ["/app/main"]
```

Output (`docker-tools`):

```nix
dockerTools.buildLayeredImage {
  name = "ubuntu";
  tag = "24.04";
  config = {
    Env = [
      "FOO=bar";
    ];
    Cmd = [
      "/app/main";
    ];
  };
}
```

Multi-stage Dockerfiles are supported. Named stages become `let` bindings in the output.

## Installation

```sh
nix build github:UnstoppableMango/docker2nix
```

Or run directly:

```sh
nix run github:UnstoppableMango/docker2nix -- ./Dockerfile
```

## Development

Requires [Nix](https://nixos.org/) with flakes enabled. With [direnv](https://direnv.net/):

```sh
direnv allow
```

Or manually:

```sh
nix develop
```

### Commands

```sh
# Build
nix build .#

# Test
ginkgo run -r

# Format
nix fmt

# Lint
nix flake check

# Regenerate protobuf
buf generate

# Sync go.sum + nix/gomod2nix.toml
make tidy
```

## Supported Dockerfile Instructions

| Instruction | Status |
| ------------------ | ------------------------ |
| `FROM` | Supported |
| `ENV` | Supported |
| `CMD` | Supported |
| `RUN` | Parsed, not yet rendered |
| `COPY` | Parsed, not yet rendered |
| Multi-stage (`AS`) | Supported |

## License

See [LICENSE](LICENSE).
