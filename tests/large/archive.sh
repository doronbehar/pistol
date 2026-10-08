#!/usr/bin/env bash
# Tests that pistol doesn't use much memory when previewing large compressed
# archives, either because they contain a huge amount of files, or a few huge
# files. See
# https://github.com/doronbehar/pistol/issues/186
#
# tests/inputs/large.amounts.tar.gz and tests/inputs/large.files.tar.gz are
# expected to exist, the Makefile takes care of that.
#
# The memory usage is sampled using ps, so the peak is approximate.

set -euo pipefail

THIS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# pistol itself uses about 30 MiB when it previews a tiny file.
MAX_RSS_MIB=256

for input in large.{amounts,files}.tar.gz; do
	# An empty config makes sure we test pistol's internal writers, and not an
	# external command.
	"$THIS_DIR/../../pistol" \
		--config /dev/null \
		"$THIS_DIR/../inputs/$input" \
		> /dev/null &
	pid=$!
	peak_kib=0
	# Sample the current resident set size (in KiB) of pistol every once in a
	# while, and keep the maximal one. Once pistol exits ps prints nothing, or 0
	# if it wasn't waited for yet, and then the loop ends.
	while rss_kib=$(ps -o rss= -p "$pid") && [[ "$rss_kib" -gt 0 ]]; do
		if [[ "$rss_kib" -gt "$peak_kib" ]]; then
			peak_kib=$rss_kib
		fi
		if [[ "$peak_kib" -gt $((MAX_RSS_MIB * 1024)) ]]; then
			# No need to wait for it to use even more memory.
			kill "$pid"
			wait "$pid" || true
			tput setaf 1
			echo "$input: pistol used more than $MAX_RSS_MIB MiB of memory ($((peak_kib / 1024)) MiB so far)"
			tput sgr0
			exit 1
		fi
		sleep 0.01
	done
	if ! wait "$pid"; then
		tput setaf 1
		echo "$input: pistol failed"
		tput sgr0
		exit 1
	fi
	tput setaf 2
	echo "$input: pistol used up to $((peak_kib / 1024)) MiB of memory, less than $MAX_RSS_MIB MiB"
	tput sgr0
done
