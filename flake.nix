# nix-direnv cache busting line: sha256-+C7gqtx+Bf6gVbBn687AbKlj//S75D7IbXmG5Qiub74=

{
  description = "trash tools";
  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, utils }@inputs:
    utils.lib.eachSystem [ "x86_64-linux" "aarch64-linux" "aarch64-darwin" ]
    (system:
      let
        graft = pkgs: pkg:
          pkg.override { buildGoModule = pkgs.buildGo122Module; };
        pkgs = import nixpkgs {
          inherit system;
          overlays = [
            (final: prev: {
              go = prev.go_1_22;
              go-tools = graft prev prev.go-tools;
              gotools = graft prev prev.gotools;
              gopls = graft prev prev.gopls;
            })
          ];
        };
        vendorHash = pkgs.lib.fileContents ./.go.mod.sri;
        version = "${self.sourceInfo.lastModifiedDate}";

        s3serve = pkgs.buildGo122Module {
          pname = "s3serve";
          inherit version vendorHash;
          src = ./.;
          subPackages = [ "cmd/s3serve" ];
        };
      in {
        packages = { s3serve = s3serve; };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            go-tools
            golangci-lint
            nixfmt
          ];

          GOAMD64 = "v3";
        };
      });
}
