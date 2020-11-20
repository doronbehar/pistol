{
  description = "General purpose file previewer designed for Ranger, Lf to make scope.sh redundant";

  # To make user overrides of the nixpkgs flake not take effect
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  # https://nixos.wiki/wiki/Flakes#Using_flakes_project_from_a_legacy_Nix
  inputs.flake-compat = {
    url = "github:edolstra/flake-compat";
    flake = false;
  };

  outputs = { self, nixpkgs, flake-utils, flake-compat }:
    flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          pistol = pkgs.pistol.overrideAttrs(oldAttrs: rec {
            version = "${builtins.readFile ./VERSION}-flake";
            buildFlagsArray = [
              "-ldflags=-s -w -X main.Version=${version}"
            ];
          });
        in rec {
          devShell = pkgs.mkShell {
            buildInputs = pistol.buildInputs ++ [
              pkgs.elinks
            ];
          };
          packages.pistol = pistol;
          defaultPackage = pistol;
          apps.pistol = {
            type = "app";
            program = "${pistol}/bin/pistol";
          };
          defaultApp = apps.pistol;
        }
      );
}
