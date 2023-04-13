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
  inputs.gomod2nix = {
    # For static compilation I need: https://github.com/tweag/gomod2nix/pull/24
    url = "github:nix-community/gomod2nix";
    inputs.nixpkgs.follows = "nixpkgs";
    inputs.utils.follows = "flake-utils";
  };
  inputs.gitignore = {
    url = "github:hercules-ci/gitignore.nix";
    inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs = { self
    , nixpkgs
    , flake-utils
    , flake-compat
    , gitignore
    , gomod2nix
  }:
  flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs {
        inherit system;
        overlays = [
          gomod2nix.overlays.default
        ];
      };
      inherit (gitignore.lib) gitignoreFilterWith;
      # https://discourse.nixos.org/t/passing-git-commit-hash-and-tag-to-build-with-flakes/11355/2
      version_rev = if (self ? rev) then (builtins.substring 0 8 self.rev) else "dirty";
      version = "${pkgs.lib.fileContents ./VERSION}-${version_rev}-flake";
      # Used also in the devShell
      MAGIC_DB = "${pkgs.pkgsStatic.file}/share/misc/magic.mgc";
      src = pkgs.lib.cleanSourceWith {
        filter = gitignoreFilterWith {
          basePath = ./.;
          extraRules = ''
            flake*
            *.nix
            ./azure-pipelines.yml
            .envrc
            # Evaluated but not used for the build itself
            VERSION
            "bump-version.sh"
            # CI files
            renovate.json5
            # Git files
            .gitignore
          '';
        };
        src = ./.;
      };
      pkgArgs = {
        inherit version src;
      };
      pistol = pkgs.callPackage ./pkg.nix pkgArgs;
      pistol-static-native = pkgs.pkgsStatic.callPackage ./pkg.nix pkgArgs;
      pistol-static-x86_64 = pkgs.pkgsCross.gnu64.pkgsStatic.callPackage ./pkg.nix pkgArgs;
      pistol-static-aarch64 = pkgs.pkgsCross.aarch64-multiplatform-musl.pkgsStatic.callPackage ./pkg.nix pkgArgs;
      pistol-static-armv7l = pkgs.pkgsCross.armv7l-hf-multiplatform.pkgsStatic.callPackage ./pkg.nix pkgArgs;
      pistol-static-arm = pkgs.pkgsCross.arm-embedded.pkgsStatic.callPackage ./pkg.nix pkgArgs;
    in {
      devShell = pkgs.mkShell {
        nativeBuildInputs = [
          pkgs.file
          # For make check
          pkgs.elinks
          pkgs.gomod2nix
        ];
        inherit MAGIC_DB;
      };
      packages = {
        inherit
          pistol
          pistol-static-native
          pistol-static-x86_64
          pistol-static-aarch64
          pistol-static-armv7l
          pistol-static-arm
        ;
      };
      defaultPackage = pistol;
      apps.pistol = {
        type = "app";
        program = "${pistol}/bin/pistol";
      };
      defaultApp = self.apps.${system}.pistol;
    }
  );
}
