#!/usr/bin/env bash
# Pull shared docs/dotfiles from the language-agnostic template remote.
# Usage: ./scripts/sync-from-template.sh
# Env: TEMPLATE_REMOTE (default: template), TEMPLATE_REF (default: main)

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

REMOTE="${TEMPLATE_REMOTE:-template}"
REF="${TEMPLATE_REF:-main}"
PATHS_FILE="${TEMPLATE_SYNC_PATHS:-template-sync.paths}"

if ! git remote get-url "$REMOTE" >/dev/null 2>&1; then
  echo "Remote '$REMOTE' is not configured." >&2
  echo "Add it with:" >&2
  echo "  git remote add template https://github.com/Nooshu/govuk-frontend-example.git" >&2
  exit 1
fi

if [[ ! -f "$PATHS_FILE" ]]; then
  echo "Missing paths file: $PATHS_FILE" >&2
  exit 1
fi

# ASCII "...": macOS bash 3.2 treats a unicode ellipsis as part of $REMOTE under `set -u`.
echo "Fetching $REMOTE..."
git fetch "$REMOTE" "$REF"

echo "Checking out shared paths from $REMOTE/$REF..."
while IFS= read -r path || [[ -n "${path:-}" ]]; do
  # Skip blank lines and comments
  [[ -z "${path// }" ]] && continue
  [[ "$path" =~ ^[[:space:]]*# ]] && continue
  if git cat-file -e "$REMOTE/$REF:$path" 2>/dev/null; then
    git checkout "$REMOTE/$REF" -- "$path"
    echo "  + $path"
  else
    echo "  ! skip (missing on template): $path" >&2
  fi
done < "$PATHS_FILE"

echo
echo "Shared paths are staged from $REMOTE/$REF."
echo "Review with: git status && git diff --cached"
echo "Commit when ready, e.g.:"
echo "  git commit -m \"chore: sync shared paths from template\""
