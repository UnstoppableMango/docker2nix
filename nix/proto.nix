{
  a2b,
  protoc-gen-go,
}:
a2b.buf.generate {
  name = "docker2nix-proto";
  src = ../proto;
  template = ../buf.gen.yaml;
  env = {
    nativeBuildInputs = [ protoc-gen-go ];
  };
}
