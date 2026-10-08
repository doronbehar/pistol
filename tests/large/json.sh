#!/usr/bin/env bash
# Tests that pistol doesn't print JSON files larger than $PISTOL_CHROMA_SIZE
# bytes, as they can't be parsed when truncated. Instead, it should print a
# message about their size. See
# https://github.com/doronbehar/pistol/issues/186
#
# tests/inputs/large.json is expected to exist, the Makefile takes care of that.
# NOTE: It has to be smaller than 7 MiB, otherwise libmagic doesn't detect it
# as JSON.

set -euo pipefail

THIS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Reading the whole file takes a long time, so this is a generous limit.
TIMEOUT=30

for chroma_size in 1000 100000 1000000; do
	# An empty config makes sure we test pistol's internal writers, and not an
	# external command.
	output=$(
		PISTOL_CHROMA_SIZE=$chroma_size timeout "$TIMEOUT" "$THIS_DIR/../../pistol" \
			--config /dev/null \
			"$THIS_DIR/../inputs/large.json"
	)
	if [[ "$output" != "JSON file larger then "* ]]; then
		tput setaf 1
		echo "large.json: PISTOL_CHROMA_SIZE=$chroma_size: pistol didn't print the JSON size message"
		tput sgr0
		exit 1
	fi
	tput setaf 2
	echo "large.json: PISTOL_CHROMA_SIZE=$chroma_size: pistol printed: $output"
	tput sgr0
done
