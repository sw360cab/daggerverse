#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 2 ]; then
  echo "Usage: $(basename "$0") <package> <version>"
  echo "Example: $(basename "$0") go.opentelemetry.io/otel/sdk v1.43.0"
  exit 1
fi

pkg=$1
ver=$2

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
daggerverse_dir="$script_dir/../daggerverse"

count=0
while IFS= read -r m; do
  dir=$(dirname "$m")
  (cd "$dir" && go mod edit -require="$pkg@$ver" && go mod tidy 2>&1 | sed "s|^|[$dir] |") &
  count=$((count + 1))
done < <(find "$daggerverse_dir" -name go.mod -not -path '*/node_modules/*')
wait

echo "Updated $count module(s) to $pkg@$ver"
exit 0
