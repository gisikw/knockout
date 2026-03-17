{
  description = "knockout — git-native task tracker and agent pipeline";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = if (self ? shortRev) then self.shortRev else "dev";
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "ko";
          inherit version;
          src = ./.;
          vendorHash = null; # deps vendored or pure Go
          ldflags = [ "-X" "main.version=${version}" ];

          # Tests run via `just test` in CI; testscript needs
          # git and a writable HOME which the sandbox doesn't provide.
          doCheck = false;

          postInstall = ''
            mv $out/bin/knockout $out/bin/ko
          '';
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [ go gopls just ];

          shellHook = ''
            echo "knockout — just --list for recipes"
          '';
        };
      }
    );
}
