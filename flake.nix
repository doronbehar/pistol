{
  description = "General purpose file previewer designed for Ranger, Lf to make scope.sh redundant";

  # To make user overrides of the nixpkgs flake not take effect
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    # https://nixos.wiki/wiki/Flakes#Using_flakes_project_from_a_legacy_Nix
    flake-compat = {
      url = "github:edolstra/flake-compat";
      flake = false;
    };
    gomod2nix-flake = {
      url = "github:tweag/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

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
      nativeBuildInputs = [
        pkgs.gomod2nix
      ];
      checkInputs = [
        pkgs.elinks
        pkgs.bat
        # For tput
        pkgs.ncurses
      ];
      buildInputs = [
        pkgs.file
      ];
      pistol = pkgs.buildGoApplication rec {
        pname = "pistol";
        version = "${pkgs.lib.strings.removeSuffix "\n" (builtins.readFile ./VERSION)}-flake";
        src = builtins.filterSource
          (path: type: type != "directory" || baseNameOf path != ".git")
          ./.;
        modules = ./gomod2nix.toml;
        inherit (pkgs.pistol) meta;
        subPackages = [ "cmd/pistol" ];
        # We don't do a check in every build (yet) so we don't need to inherit
        # checkInputs as of yet.
        inherit buildInputs;
        CGO_ENABLED = "1";
        buildFlagsArray = [ "-ldflags=-s -w -X main.Version=${version}" ];
      };
      pistol-static = pistol.overrideAttrs(oldAttrs: {
        pname = "pistol-static";
        buildInputs = buildInputs ++ [
          pkgs.glibc.static
          pkgs.pkgsStatic.file
        ];
        CFLAGS = "-I${pkgs.glibc.dev}/include";
        LDFLAGS = "-I${pkgs.glibc}/lib";
      });
    in {
      devShell = pkgs.mkShell {
        inherit buildInputs nativeBuildInputs checkInputs;
      };
      packages = {
        inherit pistol pistol-static;
      };
      defaultPackage = pistol;
      apps.pistol = {
        type = "app";
        program = "${pistol}/bin/pistol";
      };
      defaultApp = self.apps.pistol;
    }
  );
}
