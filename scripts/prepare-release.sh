#!/usr/bin/env bash
# Create a release commit, tag it, and push it.
set -euo pipefail

usage() {
	printf 'Usage: %s vMAJOR.MINOR.PATCH\n' "${0##*/}" >&2
	exit 2
}

fail() {
	printf 'Release preparation failed: %s\n' "$1" >&2
	exit 1
}

if [[ $# -ne 1 || ! $1 =~ ^v[0-9]+\.[0-9]+\.[0-9]+([-.][0-9A-Za-z.-]+)?$ ]]; then
	usage
fi

version=$1
repository_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$repository_root"

git rev-parse --is-inside-work-tree >/dev/null || fail 'repository root is not a Git work tree'

if [[ -n $(git status --porcelain --untracked-files=all) ]]; then
	git status --short >&2
	fail 'working tree must be clean'
fi

if git rev-parse --verify --quiet "refs/tags/$version" >/dev/null; then
	fail "tag $version already exists"
fi

if ! awk '
	$0 == "## Unreleased" { in_unreleased = 1; next }
	in_unreleased && /^## / { exit }
	in_unreleased && $0 !~ /^[[:space:]]*$/ { exit 1 }
	END { if (!in_unreleased) exit 1 }
' CHANGELOG.md; then
	fail 'CHANGELOG.md must have an empty ## Unreleased section'
fi

if ! grep -Eq "^## \\[$version\\]( |$)" CHANGELOG.md; then
	fail "CHANGELOG.md does not contain a release heading for $version"
fi

printf 'VERSION=%s\n' "$version" > .version
go run ./cmd/goht generate --force

git add -A
if git diff --cached --quiet; then
	fail 'version update and regeneration produced no changes'
fi

git commit -m "chore(release): $version"
git tag -a "$version" -m "Release $version"

branch=$(git branch --show-current)
[[ -n $branch ]] || fail 'cannot push a release from a detached HEAD'
git push origin "$branch" --follow-tags
