{
  buildGoApplication,
  ginkgo,
  lib,
  nix,
  version,
}:

buildGoApplication {
  pname = "docker2nix";
  inherit version;

  src = lib.cleanSource ../.;
  modules = ./gomod2nix.toml;

  nativeCheckInputs = [
    ginkgo
    nix
  ];

  checkPhase = ''
    export NIX_STATE_DIR=$(mktemp -d)
    ginkgo run -r
  '';

  meta = {
    description = "Convert Dockerfiles to Nix expressions";
    homepage = "https://github.com/UnstoppableMango/docker2nix";
    license = lib.licenses.mit;
    maintainers = with lib.maintainers; [ ];
    mainProgram = "docker2nix";
  };
}
