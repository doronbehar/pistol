# This Source Code Form is subject to the terms of the Mozilla Public
# License, version 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org/MPL/2.0/.

NAME := pistol
VERSION := v$(shell cat VERSION)-git
ifdef MAGIC_DB
	MAGIC_DB := $(MAGIC_DB)
else
	MAGIC_DB := /usr/share/misc/magic.mgc
endif

pistol: build

build:
	go build -ldflags "-X 'main.Version=$(VERSION)'" ./cmd/pistol

build-static:
	nix build -L ".#pistol-static"
	ldd ./result/bin/pistol 2>&1 | grep -q 'not a dynamic executable'

release:
	./bump-version.sh

# Manpage
pistol.1: README.adoc
	asciidoctor -b manpage -d manpage README.adoc

manpage: pistol.1

install:
	go install -ldflags "-X 'main.Version=$(VERSION)'" ./cmd/pistol

# requires: bat (https://github.com/sharkdp/bat), elinks
test: pistol
	@echo -------------------
	@echo fpath
	@echo -------------------
	@./pistol --config tests/config tests/fpath-no-sh
	@tput sgr0
	@echo -------------------
	@echo fpath + sh:
	@echo -------------------
	@./pistol --config tests/config tests/fpath-with-sh
	@tput sgr0
	@echo -------------------
	@echo mimetype
	@echo -------------------
	@./pistol --config tests/config tests/mimetype-no-sh
	@tput sgr0
	@echo -------------------
	@echo mimetype + sh:
	@echo -------------------
	@./pistol --config tests/config tests/mimetype-with-sh
	@tput sgr0
	@echo -------------------
	@echo application/json \(issue '#'34\):
	@echo -------------------
	@./pistol --config tests/config tests/34.json
	@tput sgr0
	@echo -------------------
	@echo exit code \(issue '#'52\):
	@echo -------------------
	@./tests/exit-code.sh
	@tput sgr0
	@echo -------------------
	@echo ./tests/VERSION.bz2 should appear along with license of bz2
	@echo -------------------
	@./pistol --config tests/config tests/VERSION.bz2 -- -v -L
	@tput sgr0
	@echo -------------------
	@echo ./tests/renovate.json5.bz2 should appear without a license of bz2
	@echo or verbosity, although the arguments are passed to pistol
	@echo -------------------
	@./pistol --config tests/config tests/renovate.json5.bz2 -- -v -L
	@echo -------------------
	@echo Checks substitution of multiple pistol-extra arguments without
	@echo a space between them \(issue 56\). The output should be:
	@echo
	@echo "     tests/multi-extra AxB"
	@echo
	@echo -------------------
	@./pistol --config tests/config tests/multi-extra A B

.PHONY: build install changelog
