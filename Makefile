# This Source Code Form is subject to the terms of the Mozilla Public
# License, version 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org/MPL/2.0/.

NAME := pistol
VERSION := v$(shell cat VERSION)-git
ifdef MAGIC_DB
	# Set by flake.nix
	MAGIC_DB := $(MAGIC_DB)
else
	MAGIC_DB := /usr/share/misc/magic.mgc
endif

.PHONY: pistol
pistol: build

.PHONY: build
# Cross platform build command, with the VERSION embedded to the executable -
# in contrary to all nix related builds
build:
	go build -ldflags "-X 'main.Version=$(VERSION)'" ./cmd/pistol

THIS_MAKEFILE_PATH:=$(word $(words $(MAKEFILE_LIST)),$(MAKEFILE_LIST))
THIS_DIR:=$(shell cd $(dir $(THIS_MAKEFILE_PATH));pwd)
# https://stackoverflow.com/a/76119094/4935114
COLOUR_GREEN='\033[0;32m'
COLOUR_RED='\033[0;31m'
COLOUR_BLUE='\033[0;34m'
COLOUR_YELLOW='\033[1;33m'
COLOUR_CYAN='\033[1;36m'
COLOUR_PURPLE='\033[1;35m'
COLOUR_WHITE='\033[1;37m'
COLOUR_RESET='\033[0m'

# https://stackoverflow.com/a/5810179/4935114
ifeq (, $(shell which jq)$(shell which nix))
$(warning "No jq and/or nix executables in PATH, cannot get info from flake.nix")
else
# Interestingly, builtins.currentSystem is undefined for `nix repl` and a few
# other nix commands. This example is from
# https://nix.dev/manual/nix/stable/language/builtin-constants#builtins-currentSystem
NIX_CURRENT_SYSTEM=$(shell nix-instantiate \
	--eval \
	--expr builtins.currentSystem \
	--json |\
	jq --raw-output . \
)
NIX_ATTRIBUTES=$(shell nix search \
	$(THIS_DIR) \
	pistol-static-linux \
	--json |\
	jq --raw-output 'keys | .[]' |\
	sed 's/^packages.$(NIX_CURRENT_SYSTEM).//g' \
)
NIX_TARGETS=$(foreach attr, $(NIX_ATTRIBUTES), releaseAssets/$(attr))

V_MAJOR=$(shell cut -d. -f1 VERSION)
V_MINOR=$(shell cut -d. -f2 VERSION)
V_PATCH=$(shell cut -d. -f3 VERSION)
NEXT_VERSION:=$(V_MAJOR).$(V_MINOR).$(shell echo $$(($(V_PATCH)+1)))
version_ok=$(strip $(shell \
	for version_part_idx in 1 2 3; do \
		version_part=$$(echo $(NEXT_VERSION) | cut -d. -f$$version_part_idx); \
		if test "$$version_part" -eq "$$version_part" 2> /dev/null; then \
			continue; \
		else \
			echo error: semver part $$version_part_idx of \
				version $(NEXT_VERSION) is \'$$version_part\' which is not \
				an integer; \
			break; \
		fi; \
	done \
))
ifneq (, $(version_ok))
$(error $(version_ok))
endif

TESTS_INPUTS=$(wildcard $(THIS_DIR)/tests/inputs/*)
TESTS_OUTPUTS_CURRENT=$(foreach input, \
	$(TESTS_INPUTS), \
	$(THIS_DIR)/tests/outputs/$(notdir $(input)).current \
)

check-git-clean:
	@git diff-index --quiet HEAD || ( \
		echo -e $(COLOUR_RED)Git directory is dirty, Cannot commit a new \
		VERSION file and use Nix to compile from a clean checkout.\
		$(COLOUR_RESET); exit 2)

# This below 2 wildcard checkes essentially mean: No matter how old the
# new{VersionFile,Tag} files, consider these targets as updated if the files
# exist. Useful when debugging these phases.
ifeq (,$(wildcard newVersionFile))
newVersionFile: VERSION check-git-clean
	@echo -e $(COLOUR_CYAN)❯ Updating version: \
		$(COLOUR_WHITE)$(V_MAJOR).$(V_MINOR).$(V_PATCH)$(COLOUR_RESET) \
		"->" \
		$(COLOUR_WHITE)$(NEXT_VERSION)$(COLOUR_RESET)
	@echo $(NEXT_VERSION) > VERSION
	git add VERSION
	git commit -m "Bump version to $(NEXT_VERSION)"
	@touch newVersionFile
endif
ifeq (,$(wildcard newTag))
newTag: newVersionFile
	git tag -a -m v$(NEXT_VERSION) v$(NEXT_VERSION)
	git push
	git push origin --tags v$(NEXT_VERSION)
	@touch newTag
endif

# Nix is smarter then gnumake in deciding whether a target is already available
# in the /nix/store cache or not
.PHONY: $(NIX_TARGETS)
$(NIX_TARGETS):
	@mkdir -p releaseAssets
	ln -sf $$(nix build \
		--print-build-logs \
		--no-link \
		--print-out-paths \
		.\#$(@F) \
	)/bin/pistol "$@"
	ldd "$@" 2>&1 | grep -q 'not a dynamic executable'

release: pistol.1 newTag $(NIX_TARGETS)
	gh release create v$(NEXT_VERSION) --generate-notes pistol.1 $(NIX_TARGETS)
	$(MAKE) cleanReleaseTemps

cleanReleaseTemps:
	rm -f newTag newVersionFile $(NIX_TARGETS)
endif

# Manpage
pistol.1: README.adoc
	asciidoctor -b manpage -d manpage README.adoc

manpage: pistol.1

.PHONY: install
install:
	go install -ldflags "-X 'main.Version=$(VERSION)'" ./cmd/pistol

.PHONY: $(THIS_DIR)/tests/outputs/%.current
# requires: bat (https://github.com/sharkdp/bat), elinks . Both of them are
# added to the flake.nix.
$(THIS_DIR)/tests/outputs/%.current: pistol tests/outputs/%.expected tests/inputs/%
	@echo testing input file $*
	@./pistol \
		--config $(THIS_DIR)/tests/config \
		$(THIS_DIR)/tests/inputs/$* \
		-- \
		$$(if [[ -f $(THIS_DIR)/tests/args/$*.txt ]]; then \
			cat $(THIS_DIR)/tests/args/$*.txt; \
		else \
			echo ""; \
		fi) \
		2>&1 | bat --decorations=never --show-all --color=always \
		> $@
	@diff --report-identical-files $@ $(THIS_DIR)/tests/outputs/$*.expected
	@rm $@

.PHONY: test
test: $(TESTS_OUTPUTS_CURRENT)
	@$(THIS_DIR)/tests/exit-code.sh
