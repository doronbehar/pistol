#!/usr/bin/env bash

# Thanks goes to @pete-otaqui for the initial gist:
# https://gist.github.com/pete-otaqui/4188238
#
# Original version modified by Marek Suscak
#
# works with a file called VERSION in the current directory,
# the contents of which should be a semantic version number
# such as "1.2.3" or even "1.2.3-beta+001.ab"

# this script will display the current version, automatically
# suggest a "minor" version update, and ask for input to use
# the suggestion, or a newly entered value.

# once the new version number is determined, the script will
# pull a list of changes from git history, prepend this to
# a file called CHANGELOG.md (under the title of the new version
# number), give user a chance to review and update the changelist
# manually if needed and create a GIT tag.

set -eu

NOW="$(date +'%B %d, %Y')"
RED="\033[1;31m"
GREEN="\033[0;32m"
YELLOW="\033[1;33m"
BLUE="\033[1;34m"
PURPLE="\033[1;35m"
CYAN="\033[1;36m"
WHITE="\033[1;37m"
RESET="\033[0m"

QUESTION_FLAG="${GREEN}?"
WARNING_FLAG="${YELLOW}!"
NOTICE_FLAG="${CYAN}❯"

ADJUSTMENTS_MSG="${QUESTION_FLAG} ${CYAN}Now you can make adjustments to ${WHITE}CHANGELOG.md${CYAN}. Then press enter to continue."
PUSHING_MSG="${NOTICE_FLAG} Pushing new version to the ${WHITE}origin${CYAN}..."

if [ -f VERSION ]; then
    BASE_STRING=`cat VERSION`
    BASE_LIST=(`echo $BASE_STRING | tr '.' ' '`)
    V_MAJOR=${BASE_LIST[0]}
    V_MINOR=${BASE_LIST[1]}
    V_PATCH=${BASE_LIST[2]}
    echo -e "${NOTICE_FLAG} Current version: ${WHITE}$BASE_STRING"
    SUGGESTED_VERSION="$V_MAJOR.$V_MINOR.$((V_PATCH + 1))"
    echo -ne "${QUESTION_FLAG} ${CYAN}Enter a version number [${WHITE}$SUGGESTED_VERSION${CYAN}]: "
    read INPUT_STRING
    if [ "$INPUT_STRING" = "" ]; then
        INPUT_STRING=$SUGGESTED_VERSION
    fi
    echo -e "${NOTICE_FLAG} Will set new version to be ${WHITE}$INPUT_STRING"
    echo $INPUT_STRING > VERSION
    git add VERSION
    git commit -m "Bump version to ${INPUT_STRING}."
    git push
    git tag -a -m "v$INPUT_STRING" "v$INPUT_STRING"
    git push origin --tags "v$INPUT_STRING"
    if git status --short | grep -q '*'; then
        echo -e "${WARNING_FLAG} Git directory is dirty, refusing to compile pistol-static." >&2
        exit 2
    fi
    rm -f result
    nix search . pistol-static --json | jq --raw-output 'keys | .[]' | cut -d'.' -f3 | while read target; do
        rm -f result-"$target"
        nix build --print-build-logs --print-out-paths ".#$target" --out-link result-"$target"
        echo -e "${NOTICE_FLAG} Checking that the produced executable is not a dynamically linked"
        ldd ./result-"$target"/bin/pistol 2>&1 | grep 'not a dynamic executable'
        echo -e "${NOTICE_FLAG} Checking that the produced executable has the version string compiled into it"
    done
    # Test the only executable that we can run that it has a good --version output
    ./result-pistol-static-linux-x86_64/bin/pistol --version | grep $INPUT_STRING
    gh release create v$INPUT_STRING --generate-notes \
        ./result-pistol-static-linux-x86_64/share/man/man1/pistol.1.gz
    nix search . pistol-static --json | jq --raw-output 'keys | .[]' | cut -d'.' -f3 | while read target; do
        ln -s result-"$target"/bin/pistol "$target"
        gh release upload v$INPUT_STRING "$target"
        rm "$target"
    done
else
    echo -e "${WARNING_FLAG} Could not find a VERSION file." >&2
    exit 1
fi

echo -e "${NOTICE_FLAG} Finished."
