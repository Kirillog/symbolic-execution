{
  description = "symbolic-execution-course dev environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.go
            pkgs.gcc
            pkgs.z3
            pkgs.pkg-config
            pkgs.gopls
            pkgs.delve
          ];

          env = {
            CGO_ENABLED = "1";
            GOTOOLCHAIN = "auto";
          };

          shellHook = ''
            export CGO_CFLAGS="-I${pkgs.z3.dev}/include $CGO_CFLAGS"
            export CGO_LDFLAGS="-L${pkgs.z3.lib}/lib $CGO_LDFLAGS"
          '';
        };
      });
}
