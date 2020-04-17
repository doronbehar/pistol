# This Source Code Form is subject to the terms of the Mozilla Public
# License, version 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org/MPL/2.0/.

NAME := pistol
VERSION := v0.1 (2020-04-17) (breaking release, see README)
version := $(word 1, $(VERSION))

build:
	go build -ldflags "-X 'main.VERSION=$(VERSION)'" ./cmd/pistol

install:
	go install -ldflags "-X 'main.VERSION=$(VERSION)'" ./cmd/pistol

deps:
	go get github.com/c4milo/github-release
	go get github.com/mitchellh/gox

release:
	@latest_tag=$$(git describe --tags `git rev-list --tags --max-count=1`); \
	comparison="$$latest_tag..HEAD"; \
	if [ -z "$$latest_tag" ]; then comparison=""; fi; \
	changelog=$$(git log $$comparison --oneline --no-merges); \
	echo github-release doronbehar/$(NAME) $(version) "$$(git rev-parse --abbrev-ref HEAD)" "**Changelog**<br/>$$changelog" 'dist/*'; \
	echo git pull

.PHONY: build compile install deps dist release
