#!/usr/bin/env bash

if ./pistol --config tests/config tests/exit-code.bz2 > /dev/null 2>&1; then
	tput setaf 1
	echo "exit code was not non-zero when testing a non real command"
	exit 1
else
	tput setaf 2
	echo "exit code was not zero when testing a non real command"
fi

if ./pistol --config tests/config tests/inputs > /dev/null 2>&1; then
	tput setaf 1
	echo "exit code was not non-zero when testing a real command with invalid arguments"
	exit 1
else
	tput setaf 2
	echo "exit code was not zero when testing a real command with invalid arguments"
fi
tput sgr0
exit 0
