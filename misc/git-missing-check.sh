#!/usr/bin/env bash

# SPDX-FileCopyrightText: 2026 Dominik Wombacher <dominik@wombacher.cc>
#
# SPDX-License-Identifier: Apache-2.0

# git-missing-check.sh
#
# Checks the first-level subdirectories under a given path and reports any
# that do not have a git repository initialized (no .git directory present).
#
# Only the immediate children of the search root are examined — no recursive
# nesting is assumed or supported.
#
# Usage:
#   git-missing-check.sh [path]
#
#   path  Directory to inspect (default: current working directory)
#
# Examples:
#   ./git-missing-check.sh
#   ./git-missing-check.sh ~/code

set -euo pipefail

SEARCH_ROOT="${1:-$(pwd)}"

if [[ ! -d "$SEARCH_ROOT" ]]; then
    echo "Error: '$SEARCH_ROOT' is not a directory" >&2
    exit 1
fi

found=0

for dir in "$SEARCH_ROOT"/*/; do
    [[ -d "$dir" ]] || continue
    if [[ ! -d "${dir}.git" ]]; then
        found=1
        echo "no git repo: $dir"
    fi
done

if [[ $found -eq 0 ]]; then
    echo "All directories have a git repository."
fi
