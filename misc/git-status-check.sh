#!/usr/bin/env bash

# SPDX-FileCopyrightText: 2025 Dominik Wombacher <dominik@wombacher.cc>
#
# SPDX-License-Identifier: Apache-2.0

# git-status-check.sh
#
# Recursively finds all git repositories under a given directory and reports
# any that have local work which hasn't been committed or pushed yet.
#
# Detected issues per repo:
#   - Unstaged changes       (modified tracked files, not yet staged)
#   - Staged but uncommitted (staged changes with no commit)
#   - Untracked files        (new files not added to git)
#   - Commits ahead of origin (local commits not yet pushed)
#   - Commits behind origin  (upstream commits not yet pulled)
#
# The ahead/behind check is fetch-free and uses cached remote tracking refs,
# so it won't reflect very recent upstream changes but is fast.
#
# Usage:
#   git-status-check.sh [path]
#
#   path  Directory to search under (default: current working directory)
#
# Examples:
#   ./git-status-check.sh
#   ./git-status-check.sh ~/code

set -euo pipefail

SEARCH_ROOT="${1:-$(pwd)}"

if [[ ! -d "$SEARCH_ROOT" ]]; then
    echo "Error: '$SEARCH_ROOT' is not a directory" >&2
    exit 1
fi

found_issues=0

while IFS= read -r -d '' git_dir; do
    repo_dir="${git_dir%/.git}"
    repo_name="$(basename "$repo_dir")"

    issues=()

    # Check for uncommitted changes (staged or unstaged)
    if ! git -C "$repo_dir" diff --quiet 2>/dev/null; then
        issues+=("unstaged changes")
    fi
    if ! git -C "$repo_dir" diff --cached --quiet 2>/dev/null; then
        issues+=("staged but uncommitted changes")
    fi

    # Check for untracked files
    untracked="$(git -C "$repo_dir" ls-files --others --exclude-standard 2>/dev/null)"
    if [[ -n "$untracked" ]]; then
        issues+=("untracked files")
    fi

    # Check ahead/behind without fetching (uses cached remote tracking refs)
    branch="$(git -C "$repo_dir" symbolic-ref --short HEAD 2>/dev/null || echo "")"
    if [[ -n "$branch" ]]; then
        upstream="$(git -C "$repo_dir" rev-parse --abbrev-ref "@{u}" 2>/dev/null || echo "")"
        if [[ -n "$upstream" ]]; then
            ahead="$(git -C "$repo_dir" rev-list --count "@{u}..HEAD" 2>/dev/null || echo 0)"
            behind="$(git -C "$repo_dir" rev-list --count "HEAD..@{u}" 2>/dev/null || echo 0)"
            [[ "$ahead" -gt 0 ]] && issues+=("${ahead} commit(s) ahead of origin (not pushed)")
            [[ "$behind" -gt 0 ]] && issues+=("${behind} commit(s) behind origin (need pull)")
        fi
    fi

    if [[ ${#issues[@]} -gt 0 ]]; then
        found_issues=1
        echo "repo:   $repo_name"
        echo "path:   $repo_dir"
        for issue in "${issues[@]}"; do
            echo "issue:  $issue"
        done
        echo ""
    fi
done < <(find "$SEARCH_ROOT" -type d -name ".git" -print0 2>/dev/null)

if [[ $found_issues -eq 0 ]]; then
    echo "All repos are clean."
fi
