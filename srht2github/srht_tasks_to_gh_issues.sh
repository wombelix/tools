#!/bin/bash

# SPDX-FileCopyrightText: 2025 Dominik Wombacher <dominik@wombacher.cc>
#
# SPDX-License-Identifier: Apache-2.0

# One-shot script to migrate my tasks from todo.sr.ht to github issues

# Usage: ./srht_tasks_to_gh_issues.sh tasks.json owner/repo

while IFS= read -r ticket; do
  title=$(echo "$ticket" | jq -r '.subject')
  body=$(echo "$ticket" | jq -r '.body // ""')
  status=$(echo "$ticket" | jq -r '.status')
  id=$(echo "$ticket" | jq -r '.id')

  full_body="$body

---

*Migrated from https://todo.sr.ht/~wombelix/tasks/$id*"

  echo "Creating: $title"

  issue=$(gh issue create -R "$2" -t "$title" -b "$full_body" --label "migrated srht task")

  if [ "$status" = "RESOLVED" ]; then
    num=$(echo "$issue" | grep -oE '[0-9]+$')
    gh issue close "$num" -R "$2"
  fi
done < <(jq -c '.tickets[]' "$1")
