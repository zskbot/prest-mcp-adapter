#!/usr/bin/env bash
# Fail if overall statement coverage is below COVERAGE_MIN (default 80).
set -euo pipefail

MIN="${COVERAGE_MIN:-80}"
GO="${GO:-go}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

profile="${COVERAGE_PROFILE:-coverage.out}"

"$GO" test -timeout 30s -count=1 -covermode=atomic -coverprofile="$profile" ./...

total="$("$GO" tool cover -func="$profile" | awk '/^total:/{print $NF; exit}')"
if [[ -z "$total" ]]; then
	echo "FAIL could not determine overall coverage from $profile" >&2
	exit 1
fi

num="${total%%%}"
echo "overall coverage ${num}%"

if awk -v n="$num" -v min="$MIN" 'BEGIN { exit !(n+0 < min+0) }'; then
	echo "FAIL overall coverage ${num}% < ${MIN}%" >&2
	exit 1
fi

echo "OK   overall coverage ${num}% >= ${MIN}%"
