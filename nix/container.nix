{
  docker2nix,
  lib,
  nix2container,
  version,
}:
nix2container.buildImage {
  name = "docker2nix";
  tag = version;
  config.entrypoint = [ (lib.getExe docker2nix) ];
}
