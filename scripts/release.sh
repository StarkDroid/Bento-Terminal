#!/usr/bin/env bash
#
# Publish a release from the pull requests merged since the last one.
#
# A push to main publishes automatically through .github/workflows/release.yml,
# which runs this script with --auto. Run it by hand when you want a version the
# automation cannot infer:
#
#   scripts/release.sh --auto        # version from the merged pull requests
#   scripts/release.sh 0.5.0         # or name the version yourself
#   scripts/release.sh --auto --dry-run
#
# The script runs the tests, builds the release binaries, tags the current
# commit, and publishes a GitHub release whose notes are the "## Changelog"
# bullets of every pull request merged since the last release.
#
# It needs git, go, gh, and make.

set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$root"

usage() {
	cat <<'EOF'
Publish a release from the pull requests merged since the last one.

A push to main publishes automatically through .github/workflows/release.yml.
Run this by hand when you want a version the automation cannot infer:

  scripts/release.sh --auto          # version from the merged pull requests
  scripts/release.sh 0.5.0           # or name the version yourself
  scripts/release.sh --auto --dry-run
  scripts/release.sh --since v0.4.0  # release from an earlier tag

--auto reads each merged pull request since the last tag: a major, minor, or
patch label wins, then a "!" title or a BREAKING CHANGE footer, then the
conventional type (feat bumps the minor, anything else the patch). While the
major is 0 a breaking change bumps the minor. A "Release: X.Y.Z" line on a
pull request overrides the computed version.

The script runs the tests, builds the release binaries, tags the current
commit, and publishes a GitHub release whose notes are the "## Changelog"
bullets of every pull request merged since the base tag.

It needs git, go, gh, and make.
EOF
}

die() {
	printf 'release: %s\n' "$*" >&2
	exit 1
}

version=""
dry_run=false
auto=false
since=""
while [[ $# -gt 0 ]]; do
	case $1 in
	--auto)
		auto=true
		;;
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

for tool in git gh; do
	command -v "$tool" >/dev/null 2>&1 || die "$tool is required"
done
if ! $dry_run; then
	for tool in go make; do
		command -v "$tool" >/dev/null 2>&1 || die "$tool is required"
	done
fi

[[ -z $(git status --porcelain) ]] || die "commit or stash the working tree first"
if ! $dry_run; then
	[[ $(git rev-parse --abbrev-ref HEAD) == main ]] || die "run this on main"
fi

git fetch --tags --quiet origin
if ! $dry_run; then
	remote=$(git rev-parse origin/main)
	[[ $(git rev-parse HEAD) == "$remote" ]] || die "main is not pushed; pull first"
fi

latest=$(git tag --list 'v*' --sort=-v:refname | head -n 1)
[[ -n $latest ]] || latest=""
if [[ -n $since ]]; then
	git rev-parse -q --verify "refs/tags/$since" >/dev/null || die "$since is not a tag"
	latest=$since
fi

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

if [[ -z $numbers ]]; then
	if $auto; then
		printf 'release: nothing merged since %s; no release\n' "${latest:-the first release}"
		exit 0
	fi
	die "nothing has merged since ${latest:-the first release}"
fi

# bump_rank orders the semver bumps so a batch takes the strongest one.
bump_rank() {
	case $1 in
	major) printf '3\n' ;;
	minor) printf '2\n' ;;
	*) printf '1\n' ;;
	esac
}

# bump_for_title reads the conventional type of a pull request title.
bump_for_title() {
	case $1 in
	*!:*) printf 'major\n'; return ;;
	esac
	case $1 in
	feat:* | feat\(*\):*) printf 'minor\n'; return ;;
	esac
	printf 'patch\n'
}

# next_version bumps the latest tag. Before 1.0 a breaking change bumps the
# minor instead of the major.
next_version() {
	current=${latest#v}
	[[ -n $current ]] || current=0.0.0
	IFS=. read -r cur_major cur_minor cur_patch <<<"$current"
	case $1 in
	major)
		if [[ ${cur_major:-0} -ge 1 ]]; then
			cur_major=$((cur_major + 1))
			cur_minor=0
			cur_patch=0
		else
			cur_minor=$((cur_minor + 1))
			cur_patch=0
		fi
		;;
	minor)
		cur_minor=$((cur_minor + 1))
		cur_patch=0
		;;
	*)
		cur_patch=$((cur_patch + 1))
		;;
	esac
	printf '%s.%s.%s\n' "$cur_major" "$cur_minor" "$cur_patch"
}

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

bump=""
explicit=""

for number in $numbers; do
	body=$(gh pr view "$number" --json body --jq '.body // ""' 2>/dev/null || true)
	title=$(gh pr view "$number" --json title --jq '.title // ""' 2>/dev/null || true)
	labels=$(gh pr view "$number" --json labels --jq '.labels[].name' 2>/dev/null | tr '[:upper:]' '[:lower:]' || true)

	from_body=$(printf '%s\n' "$body" | extract_changelog)
	if [[ -n $from_body ]]; then
		printf '%s\n' "$from_body" >>"$bullets"
	else
		cleaned=$(printf '%s\n' "$title" |
			sed -E 's/^(feat|fix|perf|refactor|docs|test|ci|build|chore|style)(\([^)]*\))?!?:[[:space:]]*//')
		[[ -n $cleaned ]] && printf '%s\n' "$cleaned" >>"$bullets"
	fi

	pr_bump=""
	if printf '%s\n' "$labels" | grep -qx 'major'; then
		pr_bump=major
	elif printf '%s\n' "$labels" | grep -qx 'minor'; then
		pr_bump=minor
	elif printf '%s\n' "$labels" | grep -qx 'patch'; then
		pr_bump=patch
	fi
	if [[ -z ${pr_bump:-} ]]; then
		pr_bump=$(bump_for_title "$title")
		if printf '%s\n' "$body" | grep -qi 'BREAKING[ -]CHANGE'; then
			pr_bump=major
		fi
	fi
	if [[ $(bump_rank "$pr_bump") -gt $(bump_rank "${bump:-patch}") ]]; then
		bump=$pr_bump
	fi

	rel=$(printf '%s\n' "$body" |
		sed -n 's/^[[:space:]]*[Rr]elease:[[:space:]]*[vV]\{0,1\}\([0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*\)[[:space:]]*$/\1/p' |
		head -n 1)
	if [[ -n $rel ]]; then
		if [[ -n $explicit && $explicit != "$rel" ]]; then
			die "the merged pull requests ask for both $explicit and $rel; name the version"
		fi
		explicit=$rel
	fi
done

if [[ -z $version ]]; then
	if [[ -n $explicit ]]; then
		version=$explicit
	elif $auto; then
		[[ -n $bump ]] || bump=patch
		version=$(next_version "$bump")
	else
		die "no Release: line on the merged pull requests; pass a version like 0.5.0, or use --auto"
	fi
fi
[[ $version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "$version is not a version like 0.5.0"

tag="v$version"
if git rev-parse -q --verify "refs/tags/$tag" >/dev/null; then
	if $auto; then
		printf 'release: %s already exists; no release\n' "$tag"
		exit 0
	fi
	die "$tag already exists"
fi
if gh release view "$tag" >/dev/null 2>&1; then
	if $auto; then
		printf 'release: %s already exists; no release\n' "$tag"
		exit 0
	fi
	die "release $tag already exists"
fi

repo=$(gh repo view --json nameWithOwner --jq .nameWithOwner)

awk 'NF && !seen[tolower($0)]++' "$bullets" >"$bullets.clean"
[[ -s $bullets.clean ]] || printf 'Maintenance and small fixes.\n' >"$bullets.clean"

{
	printf "## What's new in %s\n\n" "$tag"
	sed 's/^/- /' "$bullets.clean"
	if [[ -n $latest ]]; then
		printf '\n**Full changelog**: https://github.com/%s/compare/%s...%s\n' "$repo" "$latest" "$tag"
	fi
} >"$notes"

if $auto; then
	printf 'release: %s since %s, %d pull request(s), %s bump\n' \
		"$tag" "${latest:-the first release}" "$(printf '%s\n' "$numbers" | wc -l | tr -d ' ')" "$bump"
else
	printf 'release: %s since %s, %d pull request(s)\n' \
		"$tag" "${latest:-the first release}" "$(printf '%s\n' "$numbers" | wc -l | tr -d ' ')"
fi

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
