{
  description = "A Nix flake";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    systems.url = "github:nix-systems/default";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.inputs.systems.follows = "systems";
    };

    nix2container = {
      url = "github:nlewo/nix2container";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    mangonix = {
      url = "github:UnstoppableMango/nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.systems.follows = "systems";
      inputs.flake-parts.follows = "flake-parts";
      inputs.gomod2nix.follows = "gomod2nix";
      inputs.treefmt-nix.follows = "treefmt-nix";
      inputs.nix2container.follows = "nix2container";
    };
  };

  outputs =
    inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import inputs.systems;
      imports = with inputs; [ treefmt-nix.flakeModule ];

      perSystem =
        { pkgs, system, ... }:
        let
          version = "0.0.1";
          docker2nix = pkgs.callPackage ./nix { inherit version; };
        in
        {
          _module.args.pkgs = import inputs.nixpkgs {
            inherit system;
            overlays = with inputs; [ gomod2nix.overlays.default ];
          };

          packages.default = docker2nix;

          packages.container = pkgs.callPackage ./nix/container.nix {
            inherit docker2nix version;
            inherit (inputs.nix2container.packages.${system}) nix2container;
          };

          packages.registries-conf = pkgs.callPackage ./nix/registries-conf.nix { };

          checks.golangci-lint = pkgs.buildGoApplication {
            name = "docker2nix-lint";
            src = pkgs.lib.cleanSource ./.;
            modules = ./nix/gomod2nix.toml;
            nativeBuildInputs = [ pkgs.golangci-lint ];
            buildPhase = ''
              HOME=$TMPDIR golangci-lint run ./...
            '';
            installPhase = "touch $out";
            doCheck = false;
          };

          devShells.default = pkgs.mkShellNoCC {
            packages = with pkgs; [
              buf
              direnv
              go
              gomod2nix
              golangci-lint
              gopls
              ginkgo
              gnumake
              nixfmt
              protoc-gen-go
            ];

            GO = "${pkgs.go}/bin/go";
            GOMOD2NIX = "${pkgs.gomod2nix}/bin/gomod2nix";
          };

          treefmt.programs = {
            actionlint.enable = true;
            gofmt.enable = true;
            jsonfmt.enable = true;
            mdformat.enable = true;
            nixfmt.enable = true;
          };
        };
    };
}
