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
    url = "github:tweag/gomod2nix";
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
      pistol = pkgs.buildGoApplication {
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
            # CI files
            "renovate.json5"
            # Git files
            ".gitignore"
            ".git"
          ])
          ;
          src = gitignoreSource ./.;
        };
        buildFlagsArray = [
          "-ldflags=-s -w -X main.Version=${version}"
        ];
        modules = ./gomod2nix.toml;
        inherit (pkgs.pistol)
          nativeBuildInputs
          buildInputs
          subPackages
          postBuild
          meta
        ;
        CGO_ENABLED = 1;
      };
    in rec {
      devShell = pkgs.mkShell {
        inherit (pistol) buildInputs;
        nativeBuildInputs = pistol.nativeBuildInputs ++ [
          # For make check
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
