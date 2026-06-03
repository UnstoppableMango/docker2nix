{
  buildGoApplication,
  ginkgo,
  lib,
  version,
}:
buildGoApplication {
  pname = "docker2nix";
  inherit version;

  src = lib.cleanSource ../.;
  modules = ./gomod2nix.toml;

  nativeCheckInputs = [ ginkgo ];

  checkPhase = ''
    ginkgo run -r
  '';
}
