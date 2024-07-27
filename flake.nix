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
            ### Nix related
            flake*
            *.nix
            .envrc
            .direnv
            ### Makefile related files
            ./Makefile
            "bump-version.sh"
            # built by go build or simply with `make`
            ./pistol
            ./pistol.1
            ./README.html
            # Evaluated here in this flake.nix but not used for the build itself
            VERSION
            ### CI files
            renovate.json5
            ### Git files
            .gitignore
          '';
        };
        src = ./.;
      };
      pkgArgs = {
        inherit version src;
      };
      pistol = pkgs.callPackage ./pkg.nix pkgArgs;
      pistol-static-linux-native = pkgs.pkgsStatic.callPackage ./pkg.nix pkgArgs;
      pistol-static-linux-x86_64 = pkgs.pkgsCross.gnu64.pkgsStatic.callPackage ./pkg.nix pkgArgs;
      pistol-static-linux-aarch64 = pkgs.pkgsCross.aarch64-multiplatform-musl.pkgsStatic.callPackage ./pkg.nix pkgArgs;
      pistol-static-linux-armv7l = pkgs.pkgsCross.armv7l-hf-multiplatform.pkgsStatic.callPackage ./pkg.nix pkgArgs;
      pistol-static-linux-arm = pkgs.pkgsCross.arm-embedded.pkgsStatic.callPackage ./pkg.nix pkgArgs;
    in {
      devShell = pkgs.mkShell {
        nativeBuildInputs = [
          pkgs.file
          # For make check
          pkgs.elinks
          pkgs.gomod2nix
        ];
        # Only useful if I need to play with static compilation out side of
        # Nix, mostly it is never used.
        inherit MAGIC_DB;
      };
      packages = {
        inherit
          pistol
          pistol-static-linux-native
          pistol-static-linux-x86_64
          pistol-static-linux-aarch64
          pistol-static-linux-armv7l
          #pistol-static-linux-arm # Currently broken
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
