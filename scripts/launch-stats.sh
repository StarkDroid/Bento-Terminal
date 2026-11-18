#!/usr/bin/env bash
# Print public launch metrics. No app telemetry is involved: these are GitHub,
# npm, and Chocolatey counts, plus GitHub's own repository traffic.
#
#   scripts/launch-stats.sh
#
# Needs gh (logged in), curl, and jq. xmllint is optional.

set -euo pipefail

repo="StarkDroid/Bento-Terminal"
npm_package="bento-terminal"
choco_package="bento-terminal"

have() { command -v "$1" >/dev/null 2>&1; }
have gh || { echo "gh is required" >&2; exit 1; }
have curl || { echo "curl is required" >&2; exit 1; }
have jq || { echo "jq is required" >&2; exit 1; }

echo "Bento terminal — launch stats"
echo "============================"

echo
echo "GitHub"
gh api "repos/$repo" --jq '"stars \(.stargazers_count)  forks \(.forks_count)  watchers \(.subscribers_count)  issues \(.open_issues_count)"'

tag=$(gh release view --repo "$repo" --json tagName --jq .tagName 2>/dev/null || true)
if [[ -n "$tag" ]]; then
	echo
	echo "Latest release: $tag"
	gh release view "$tag" --repo "$repo" --json assets \
		--jq '.assets[] | "  \(.name)  \(.downloadCount)"' | sort
fi

echo
echo "GitHub traffic (last 14 days)"
for endpoint in views clones; do
	if data=$(gh api "repos/$repo/traffic/$endpoint" 2>/dev/null); then
		printf '  %-6s %s\n' "$endpoint" "$(jq -r '"\(.count) total, \(.uniques) unique"' <<<"$data")"
	else
		printf '  %-6s unavailable (needs repo push access)\n' "$endpoint"
	fi
done

echo
echo "npm"
if npm_json=$(curl -fsSL "https://api.npmjs.org/downloads/point/last-month/$npm_package" 2>/dev/null); then
	jq -r '"  \(.downloads) downloads in the last month"' <<<"$npm_json"
else
	echo "  not published yet"
fi

echo
echo "Chocolatey"
choco_version="${tag#v}"
if [[ -n "$choco_version" ]] && choco_xml=$(curl -fsSL "https://community.chocolatey.org/api/v2/Packages(Id='$choco_package',Version='$choco_version')" 2>/dev/null); then
	if have xmllint; then
		downloads=$(xmllint --xpath 'string(//*[local-name()="DownloadCount"])' <<<"$choco_xml" 2>/dev/null || true)
	else
		downloads=$(grep -o '<d:DownloadCount[^>]*>[^<]*' <<<"$choco_xml" | head -1 | sed 's/.*>//' || true)
	fi
	echo "  ${choco_version}: ${downloads:-0} downloads"
else
	echo "  no published ${choco_version:-} package found"
fi

echo
echo "Website"
echo "  GoatCounter dashboard: https://<yourcode>.goatcounter.com"
echo "  Set GOATCOUNTER_CODE as a repository variable to enable the script."
echo "  Track install-copy events from the dashboard; the app has no telemetry."
