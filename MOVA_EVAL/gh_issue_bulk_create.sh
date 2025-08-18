#!/usr/bin/env bash
# Usage: REPO=owner/name ./gh_issue_bulk_create.sh mova_issues.jsonl
set -euo pipefail
FILE="${1:-mova_issues.jsonl}"
: "${REPO:?Set REPO=owner/name}"
while IFS= read -r line; do
  title=$(jq -r '.title' <<<"$line")
  body=$(jq -r '.body' <<<"$line")
  labels=$(jq -r '.labels | join(",")' <<<"$line")
  gh issue create --repo "$REPO" --title "$title" --body "$body" --label "$labels"
done < "$FILE"
echo "Done."
