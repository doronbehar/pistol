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

  inputs.gomod2nix-flake.url = "github:tweag/gomod2nix";

  outputs = { self
    , nixpkgs
    , flake-utils
    , flake-compat
    , gomod2nix-flake
  }:
  flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs {
        overlays = [
          gomod2nix-flake.overlay
        ];
        inherit system;
        config = {};
      };
      pistol = pkgs.buildGoApplication rec {
        pname = "pistol";
        version = "${pkgs.lib.strings.removeSuffix "\n" (builtins.readFile ./VERSION)}-flake";
        src = builtins.filterSource
          (path: type: type != "directory" || baseNameOf path != ".git")
          ./.;
        modules = ./gomod2nix.toml;
        inherit (pkgs.pistol)
          meta
          passthru
          subPackages
          buildInputs
        ;
        buildFlagsArray = [ "-ldflags=-s -w -X main.Version=${version}" ];
      };
    in rec {
      devShell = pkgs.mkShell {
        buildInputs = [
          pkgs.file
          pkgs.elinks
          pkgs.gomod2nix
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
