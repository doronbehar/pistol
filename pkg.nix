{ lib
, stdenv
, buildGoApplication
, version, src
, file
, buildPackages
, removeReferencesTo
, installShellFiles
, asciidoctor
}:

buildGoApplication {
  pname = "pistol";
  inherit version src;
  pwd = ./.;
  modules = ./gomod2nix.toml;

  subPackages = [ "cmd/pistol" ];
  ldflags = [ "-s" "-w" "-X main.Version=${version}" ];

  doCheck = false;
  buildInputs = [
    file
  ];
  nativeBuildInputs = [
    installShellFiles
    asciidoctor
  ] ++ lib.optionals stdenv.hostPlatform.isStatic [
    removeReferencesTo
  ];
  NIX_LDFLAGS = lib.optionalString stdenv.hostPlatform.isStatic "-lz";
  tags = lib.optionals stdenv.hostPlatform.isStatic [
    "EMBED_MAGIC_DB"
  ];
  preBuild = lib.optionalString stdenv.hostPlatform.isStatic ''
    cp ${file}/share/misc/magic.mgc ./cmd/pistol/magic.mgc
  '';
  postInstall = ''
    asciidoctor -b manpage -d manpage README.adoc
    installManPage pistol.1
  '';
  postFixup = lib.optionalString stdenv.hostPlatform.isStatic ''
    # Remove unnecessary references to zlib.
    rm -r $out/nix-support
    # Remove more unnecessary references which I don't know the source of
    # which. I guess they are due to features of some go modules I don't
    # use.
    remove-references-to -t ${buildPackages.mailcap} $out/bin/pistol
    remove-references-to -t ${buildPackages.iana-etc} $out/bin/pistol
    remove-references-to -t ${buildPackages.tzdata} $out/bin/pistol
    remove-references-to -t ${file} $out/bin/pistol
  '';
  meta = {
    description = "General purpose file previewer designed for Ranger, Lf to make scope.sh redundant";
    homepage = "https://github.com/doronbehar/pistol";
    license = lib.licenses.mit;
  };
}
