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
    url = "github:doronbehar/gomod2nix/go-stdenv";
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
          gomod2nix.overlay
        ];
      };
      inherit (gitignore.lib) gitignoreSource;
      # https://discourse.nixos.org/t/passing-git-commit-hash-and-tag-to-build-with-flakes/11355/2
      version_rev = if (self ? rev) then (builtins.substring 0 8 self.rev) else "dirty";
      version = "${pkgs.lib.fileContents ./VERSION}-${version_rev}-flake";
      # Create the buildGoApplication variants
      inherit (pkgs) buildGoApplication;
      buildGoApplicationStatic = buildGoApplication.override {
        stdenv = pkgs.pkgsStatic.stdenv;
      };
      # Used also in the devShell
      MAGIC_DB = "${pkgs.pkgsStatic.file}/share/misc/magic.mgc";
      # arguments used in many derivation arguments calls
      common-drv-args = {
        pname = "pistol";
        inherit version;
        src = pkgs.lib.cleanSourceWith {
          # Ignore many files that gitignoreSource doesn't ignore, see:
          # https://github.com/hercules-ci/gitignore.nix/issues/9#issuecomment-635458762
          filter = path: type:
          ! (builtins.any (r: (builtins.match r (builtins.baseNameOf path)) != null) [
            # Nix files
            "flake.nix"
            "flake.lock"
            "default.nix"
            "shell.nix"
            ".envrc"
            # Evaluated but not used for the build itself
            "gomod2nix.toml"
            "VERSION"
            "bump-version.sh"
            # CI files
            "renovate.json5"
            # Git files
            ".gitignore"
            ".git"
          ])
          ;
          src = gitignoreSource ./.;
        };
        buildFlagsArray = ''
          -ldflags=
          -X main.Version=${version}
        '';
        modules = ./gomod2nix.toml;
        inherit (pkgs.pistol)
          nativeBuildInputs
          subPackages
          postBuild
          meta
        ;
        CGO_ENABLED = 1;
      };
      common-static-drv-args = (common-drv-args // {
        nativeBuildInputs = common-drv-args.nativeBuildInputs ++ [
          pkgs.removeReferencesTo
        ];
        # From some reason even though zlib is static we need this, but it
        # doesn't create a real reference to zlib.
        NIX_LDFLAGS = "-lz";
        preBuild = ''
          cp ${MAGIC_DB} ./cmd/pistol/magic.mgc
        '';
        buildFlags = ''
          -tags EMBED_MAGIC_DB
        '';
        postFixup = ''
          # Remove unnecessary references to zlib.
          rm -r $out/nix-support
          # Remove more unnecessary references which I don't know the source of
          # which. I guess they are due to features of some go modules I don't
          # use.
          remove-references-to -t ${pkgs.mailcap} $out/bin/pistol
          remove-references-to -t ${pkgs.iana-etc} $out/bin/pistol
          remove-references-to -t ${pkgs.tzdata} $out/bin/pistol
        '';
      });
      pistol = buildGoApplication (common-drv-args // {
        inherit (pkgs.pistol) buildInputs;
      });
      pistol-static = buildGoApplicationStatic (common-static-drv-args // {
        buildInputs = [
          pkgs.pkgsStatic.file
          pkgs.pkgsStatic.zlib
        ];
        postFixup = common-static-drv-args.postFixup + ''
          remove-references-to -t ${pkgs.pkgsStatic.file} $out/bin/pistol
        '';
      });
    in rec {
      devShell = pkgs.mkShell {
        inherit (pistol) buildInputs;
        nativeBuildInputs = pistol.nativeBuildInputs ++ [
          # For make check
          pkgs.elinks
          pkgs.gomod2nix
        ];
        inherit MAGIC_DB;
      };
      packages = {
        inherit
          pistol
          pistol-static
        ;
      };
      defaultPackage = pistol;
      apps.pistol = {
        type = "app";
        program = "${pistol}/bin/pistol";
      };
      defaultApp = apps.pistol;
    }
  );
}
