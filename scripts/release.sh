#!/usr/bin/env bash
#
# Publish a release from the pull requests merged since the last one.
#
# Merging a pull request publishes nothing. Merge it yourself, then run this
# from main:
#
#   scripts/release.sh            # version from the merged PR's Release: line
#   scripts/release.sh 0.5.0      # or name the version yourself
#   scripts/release.sh --dry-run  # print the version and notes, change nothing
#
# The script runs the tests, builds the release binaries, tags the current
# commit, and publishes a GitHub release whose notes are the "## Changelog"
# bullets of every pull request merged since the last release.
#
# It needs git, go, and the GitHub CLI signed in.

set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$root"

usage() {
	cat <<'EOF'
Publish a release from the pull requests merged since the last one.

Merging a pull request publishes nothing. Merge it yourself, then run this
from main:

  scripts/release.sh                 # version from the PR's Release: line
  scripts/release.sh 0.5.0           # or name the version yourself
  scripts/release.sh --dry-run       # print the version and notes, do nothing
  scripts/release.sh --since v0.4.0  # release from an earlier tag

The script runs the tests, builds the release binaries, tags the current
commit, and publishes a GitHub release whose notes are the "## Changelog"
bullets of every pull request merged since the base tag.

It needs git, go, and the GitHub CLI signed in.
EOF
}

die() {
	printf 'release: %s\n' "$*" >&2
	exit 1
}

version=""
dry_run=false
since=""
while [[ $# -gt 0 ]]; do
	case $1 in
	--dry-run)
		dry_run=true
		;;
	--since)
		[[ $# -ge 2 ]] || die "--since needs a tag"
		since=$2
		shift
		;;
	--since=*)
		since=${1#--since=}
		;;
	-h | --help)
		usage
		exit 0
		;;
	-*)
		die "unknown option $1"
		;;
	*)
		[[ -z $version ]] || die "give one version"
		version=${1#v}
		;;
	esac
	shift
done

for tool in git go gh; do
	command -v "$tool" >/dev/null 2>&1 || die "$tool is required"
done

[[ -z $(git status --porcelain) ]] || die "commit or stash the working tree first"
[[ $(git rev-parse --abbrev-ref HEAD) == main ]] || die "run this on main"

git fetch --tags --quiet origin
remote=$(git rev-parse origin/main)
[[ $(git rev-parse HEAD) == "$remote" ]] || die "main is not pushed; pull first"

latest=$(git tag --list 'v*' --sort=-v:refname | head -n 1)
[[ -n $latest ]] || latest=""
if [[ -n $since ]]; then
	git rev-parse -q --verify "refs/tags/$since" >/dev/null || die "$since is not a tag"
	latest=$since
fi

# release_version_from_prs prints the Release: line of the pull request that
# produced HEAD, then of the most recently merged pull requests.
release_version_from_prs() {
	{
		gh api "repos/{owner}/{repo}/commits/$(git rev-parse HEAD)/pulls" \
			--jq '.[] | select(.merged_at) | .body' 2>/dev/null || true
		gh pr list --state merged --base main --limit 10 \
			--json body --jq '.[] | .body' 2>/dev/null || true
	} |
		sed -n 's/^[[:space:]]*[Rr]elease:[[:space:]]*[vV]\{0,1\}\([0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*\)[[:space:]]*$/\1/p' |
		head -n 1
}

if [[ -z $version ]]; then
	version=$(release_version_from_prs)
fi
[[ -n $version ]] || die "no Release: line on the merged pull request; pass a version, for example scripts/release.sh 0.5.0"
[[ $version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "$version is not a version like 0.5.0"

tag="v$version"
if git rev-parse -q --verify "refs/tags/$tag" >/dev/null; then
	die "$tag already exists"
fi
if gh release view "$tag" >/dev/null 2>&1; then
	die "release $tag already exists"
fi

repo=$(gh repo view --json nameWithOwner --jq .nameWithOwner)

range="HEAD"
[[ -n $latest ]] && range="$latest..HEAD"
numbers=$(
	{
		git log "$range" --first-parent --merges --format=%s
		git log "$range" --first-parent --no-merges --format=%s
	} |
		sed -n -e 's/^Merge pull request #\([0-9]\{1,\}\).*/\1/p' \
			-e 's/.*(#\([0-9]\{1,\}\))$/\1/p' |
		awk 'NF && !seen[$0]++'
)
[[ -n $numbers ]] || die "nothing has merged since ${latest:-the first release}"

# extract_changelog prints the bullets under "## Changelog" and stops at the
# next heading, so comments and other sections stay out of the notes.
extract_changelog() {
	awk '
		BEGIN { inlog = 0 }
		tolower($0) ~ /^##[[:space:]]*changelog[[:space:]]*$/ { inlog = 1; next }
		inlog && /^##[[:space:]]/ { inlog = 0 }
		inlog && /^[-*][[:space:]]+[^[:space:]]/ {
			line = $0
			sub(/^[-*][[:space:]]+/, "", line)
			print line
		}
	'
}

bullets=$(mktemp)
notes=$(mktemp)
cleanup() {
	rm -f "$bullets" "$bullets.clean" "$notes"
}
trap cleanup EXIT

for number in $numbers; do
	body=$(gh pr view "$number" --json body --jq '.body // ""' 2>/dev/null || true)
	title=$(gh pr view "$number" --json title --jq '.title // ""' 2>/dev/null || true)
	from_body=$(printf '%s\n' "$body" | extract_changelog)
	if [[ -n $from_body ]]; then
		printf '%s\n' "$from_body" >>"$bullets"
		continue
	fi
	cleaned=$(printf '%s\n' "$title" |
		sed -E 's/^(feat|fix|perf|refactor|docs|test|ci|build|chore|style)(\([^)]*\))?!?:[[:space:]]*//')
	[[ -n $cleaned ]] && printf '%s\n' "$cleaned" >>"$bullets"
done

awk 'NF && !seen[tolower($0)]++' "$bullets" >"$bullets.clean"
[[ -s $bullets.clean ]] || printf 'Maintenance and small fixes.\n' >"$bullets.clean"

{
	printf "## What's new in %s\n\n" "$tag"
	sed 's/^/- /' "$bullets.clean"
	if [[ -n $latest ]]; then
		printf '\n**Full changelog**: https://github.com/%s/compare/%s...%s\n' "$repo" "$latest" "$tag"
	fi
} >"$notes"

source_version=$(sed -n 's/^var version = "\(.*\)"$/\1/p' cmd/bento/main.go)
if [[ $source_version != "$version" ]]; then
	printf 'release: note: cmd/bento/main.go says %s; bump it in the release pull request so the source matches %s\n' \
		"${source_version:-unknown}" "$version" >&2
fi

printf 'release: %s since %s, %d pull request(s)\n' "$tag" "${latest:-the first release}" "$(printf '%s\n' "$numbers" | wc -l | tr -d ' ')"

if $dry_run; then
	printf '\n%s\n' "$(cat "$notes")"
	printf 'release: dry run; nothing was tagged or published\n'
	exit 0
fi

printf 'release: running the tests\n'
go test ./...

printf 'release: building the binaries\n'
rm -rf dist
make darwin-arm64 darwin-amd64 linux-arm64 linux-amd64 windows-amd64 windows-arm64 VERSION="$version" >/dev/null

if command -v sha256sum >/dev/null 2>&1; then
	sum=(sha256sum)
else
	sum=(shasum -a 256)
fi
(
	cd dist
	"${sum[@]}" bento_darwin_arm64 bento_darwin_amd64 bento_linux_arm64 bento_linux_amd64 \
		bento_windows_amd64.exe bento_windows_arm64.exe >SHA256SUMS
)

printf 'release: tagging %s\n' "$tag"
git tag -a "$tag" -m "Bento terminal $tag"
git push origin "$tag"

printf 'release: publishing %s\n' "$tag"
gh release create "$tag" dist/bento_darwin_arm64 dist/bento_darwin_amd64 \
	dist/bento_linux_arm64 dist/bento_linux_amd64 \
	dist/bento_windows_amd64.exe dist/bento_windows_arm64.exe \
	dist/SHA256SUMS \
	--title "Bento terminal $tag" \
	--notes-file "$notes" \
	--verify-tag >/dev/null

gh release view "$tag" --json url --jq .url
