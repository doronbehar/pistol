#!/usr/bin/env bash
# Tests that pistol doesn't read (and hence print) more than $PISTOL_CHROMA_SIZE
# bytes of a large text file, either plain or compressed, for a few different
# values of it. See
# https://github.com/doronbehar/pistol/issues/186
#
# tests/inputs/large.log and tests/inputs/large.log.gz are expected to exist, the
# Makefile takes care of that.

set -euo pipefail

THIS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Reading the whole file takes minutes, so this is a generous limit.
TIMEOUT=30

for input in large.log{,.gz}; do
	for chroma_size in 1000 100000 1000000 10000000; do
		# EPOCHREALTIME is in seconds with a microseconds fraction. Without the
		# dot it's an integer of microseconds, which bash can do math with.
		start=${EPOCHREALTIME/./}
		# An empty config makes sure we test pistol's internal writers, and not
		# an external command. chroma adds ANSI escape sequences to the text,
		# which shouldn't be counted.
		size=$(
			PISTOL_CHROMA_SIZE=$chroma_size timeout "$TIMEOUT" "$THIS_DIR/../../pistol" \
				--config /dev/null \
				"$THIS_DIR/../inputs/$input" |
				sed 's/\x1b\[[0-9;]*m//g' |
				wc --bytes
		)
		elapsed_us=$((${EPOCHREALTIME/./} - start))
		# The input files are way larger than all sizes tested, so we expect to
		# get exactly PISTOL_CHROMA_SIZE bytes.
		if [[ "$size" -ne "$chroma_size" ]]; then
			tput setaf 1
			echo "$input: pistol printed $size bytes, expected PISTOL_CHROMA_SIZE=$chroma_size"
			tput sgr0
			exit 1
		fi
		tput setaf 2
		# Pad to at least 7 digits, so that there's always a digit before the
		# 6 digits of the fraction.
		printf -v elapsed '%07d' "$elapsed_us"
		echo "$input: PISTOL_CHROMA_SIZE=$chroma_size: pistol printed $size bytes, in ${elapsed:0:-6}.${elapsed: -6}s"
		tput sgr0
	done
done
