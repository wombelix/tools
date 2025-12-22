#!/bin/bash

# SPDX-FileCopyrightText: 2025 Dominik Wombacher <dominik@wombacher.cc>
#
# SPDX-License-Identifier: Apache-2.0

# One-shot script to migrate my tasks from todo.sr.ht to github issues

# Usage: ./bulk_close_srht_tasks.sh tasks.json owner/repo tracker

TRACKER="$3"

# Get GitHub issues
gh_issues=$(GH_PAGER="" gh issue list -R "$2" --state all --json number,title --jq 'map({(.title): .number}) | add')

# Process each ticket
jq -c '.tickets[]' "$1" | while IFS= read -r ticket; do
  title=$(echo "$ticket" | jq -r '.subject')
  id=$(echo "$ticket" | jq -r '.id')
  status=$(echo "$ticket" | jq -r '.status')

  # Get corresponding GitHub issue number
  gh_num=$(echo "$gh_issues" | jq -r --arg title "$title" '.[$title] // empty')

  if [ -z "$gh_num" ]; then
    echo "Skipping #$id - no matching GitHub issue"
    continue
  fi

  gh_url="https://github.com/$2/issues/$gh_num"
  comment="Migrated to GitHub: $gh_url"

  echo "Closing todo.sr.ht/$TRACKER #$id with reference to: $gh_url"

  # Add comment and close if not already resolved
  if [ "$status" != "RESOLVED" ]; then
    echo "$comment" | hut todo ticket comment "$id" -t "$TRACKER" --stdin -s RESOLVED -r CLOSED
  else
    echo "$comment" | hut todo ticket comment "$id" -t "$TRACKER" --stdin
  fi
done
