GO        ?= go
GOMOD2NIX ?= gomod2nix
GINKGO    ?= ginkgo

GO_SRC ?= $(shell find . -name '*.go')

build:
	nix build .#

container:
	nix build .#container

test:
	$(GINKGO) run -r

cover: coverprofile.out

coverprofile.out: ${GO_SRC}
	$(GINKGO) run -r --cover

update:
	nix flake update

check:
	nix flake check

lint:
	golangci-lint run ./...

format fmt:
	nix fmt

tidy: go.sum nix/gomod2nix.toml

go.sum: go.mod ${GO_SRC}
	$(GO) mod tidy

nix/gomod2nix.toml: go.sum ${GO_SRC}
	$(GOMOD2NIX) generate --dir ${CURDIR} --outdir ${@D}

.PHONY: test
