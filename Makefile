# This Source Code Form is subject to the terms of the Mozilla Public
# License, version 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org/MPL/2.0/.

NAME := pistol
VERSION := v0.1.1 (2020-04-17) (breaking release, see README)
# version := $(word 1, $(VERSION))

build:
	go build -ldflags "-X 'main.Version=$(VERSION)'" ./cmd/pistol

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

deps:
	go get github.com/c4milo/github-release
	go get github.com/mitchellh/gox

changelog:
	@latest_tag=$$(git describe --tags `git rev-list --tags --max-count=1`); \
	comparison="$$latest_tag..HEAD"; \
	if [ -z "$$latest_tag" ]; then comparison=""; fi; \
	git --no-pager log $$comparison --oneline --no-merges

.PHONY: build install changelog
