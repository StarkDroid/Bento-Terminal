# Packaging

Bento can be installed from GitHub Releases today and is prepared for the
package managers below. Every manifest is rendered from the release's
`SHA256SUMS`, so a release never ships a stale hash.

| Source | Command | Prepared here | Published by |
| --- | --- | --- | --- |
| npm / npx | `npx bento-terminal` | `packaging/npm` | Release workflow when `NPM_TOKEN` is set |
| Homebrew | `brew install StarkDroid/tap/bento` | `packaging/homebrew` | Tap update (manual or `TAP_TOKEN`) |
| Scoop | `scoop install bento` | `packaging/scoop` | Bucket update (manual or `SCOOP_TOKEN`) |
| WinGet | `winget install StarkDroid.Bento` | `packaging/winget` | PR to `microsoft/winget-pkgs` |
| Chocolatey | `choco install bento-terminal` | `packaging/chocolatey` | `choco push` with an API key |
| Arch / AUR | `yay -S bento-terminal-bin` | `packaging/aur` | Push to the AUR |
| Debian / RPM / Alpine | `.deb`, `.rpm`, `.apk` attachments | `packaging/nfpm` | Release workflow |
| Go | `go install github.com/StarkDroid/Bento-Terminal/cmd/bento@latest` | module | Go proxy |
| Direct binary | GitHub release asset | `scripts/install.sh`, `scripts/install.ps1` | Release workflow |

## Render the manifests

The release workflow renders them automatically, and the same command works
locally after downloading a release's `SHA256SUMS`:

```sh
curl -fsSL \
  https://github.com/StarkDroid/Bento-Terminal/releases/download/v0.7.0/SHA256SUMS \
  -o SHA256SUMS

python3 packaging/render.py --version 0.7.0 --sums SHA256SUMS --out /tmp/bento-packages
```

The output tree contains:

```text
homebrew/bento.rb
scoop/bento.json
winget/manifests/s/StarkDroid/Bento/<version>/*.yaml
chocolatey/bento-terminal.nuspec
chocolatey/tools/*.ps1
aur/PKGBUILD
aur/.SRCINFO
nfpm/amd64.yaml
nfpm/arm64.yaml
```

`scripts/release.sh` is unchanged: it still publishes the six raw binaries and
`SHA256SUMS`. The package manifests point at those assets.

## npm

The package is `packaging/npm`. It ships a small launcher and downloads the
native binary during `postinstall`, verifying it against `SHA256SUMS`.

The release workflow sets `package.json` to the release version and publishes
with provenance when the repository secret `NPM_TOKEN` exists. Without the
secret the job skips publishing and writes a note to the summary.

One-time setup:

1. `npm login` as the account that owns the `bento-terminal` name.
2. Create an Automation token with publish access. A granular token also works
   if it has the 2FA bypass setting; a classic token still asks for an OTP.
3. Add it to the repository as `NPM_TOKEN`.

Local dry run:

```sh
cd packaging/npm
npm pack --dry-run
BENTO_VERSION=0.7.0 node scripts/postinstall.js
BENTO_VERSION=0.7.0 node bin/bento.js --version
```

The launcher also works when `npm install --ignore-scripts` is used: it
downloads the binary on first run. `BENTO_BASE_URL` points it at a mirror or a
local release directory.

## Homebrew

The formula is `packaging/homebrew/bento.rb` after rendering. It installs the
raw release binary for the host.

The tap is already published at
[`StarkDroid/homebrew-tap`](https://github.com/StarkDroid/homebrew-tap). Users run:

```sh
brew install StarkDroid/tap/bento
```

To recreate the tap from scratch:

```sh
gh repo create StarkDroid/homebrew-tap --public \
  --description "Homebrew tap for Bento terminal"

git clone git@github.com:StarkDroid/homebrew-tap.git
mkdir -p homebrew-tap/Formula
cp /tmp/bento-packages/homebrew/bento.rb homebrew-tap/Formula/bento.rb
cd homebrew-tap
git add Formula/bento.rb
git commit -m "bento 0.7.0"
git push
```

`brew style` the rendered formula before pushing. The tap lives at
`StarkDroid/homebrew-tap`; the release workflow updates `Formula/bento.rb`
automatically when the repository secret `TAP_TOKEN` is set.

## Scoop

The manifest is `packaging/scoop/bento.json` after rendering. It uses the
architecture-specific `bin` entries so the downloaded `.exe` becomes `bento`.

The bucket is already published at
[`StarkDroid/scoop-bucket`](https://github.com/StarkDroid/scoop-bucket). Users run:

```sh
scoop bucket add starkdroid https://github.com/StarkDroid/scoop-bucket
scoop install bento
```

To recreate the bucket from scratch:

```sh
gh repo create StarkDroid/scoop-bucket --public \
  --description "Scoop bucket for Bento terminal"

git clone git@github.com:StarkDroid/scoop-bucket.git
mkdir -p scoop-bucket/bucket
cp /tmp/bento-packages/scoop/bento.json scoop-bucket/bucket/bento.json
cd scoop-bucket
git add bucket/bento.json
git commit -m "bento: add 0.7.0"
git push
```

The release workflow updates `bucket/bento.json` automatically when the
repository secret `SCOOP_TOKEN` is set.

## WinGet

The manifests are under `packaging/winget` after rendering. They use the
portable installer type, so WinGet places `bento` on PATH without an installer.

Submit them with [`wingetcreate`](https://github.com/microsoft/winget-create):

```sh
wingetcreate submit \
  /tmp/bento-packages/winget/manifests/s/StarkDroid/Bento/0.7.0 \
  --token "$GITHUB_TOKEN"
```

That opens a pull request against `microsoft/winget-pkgs`. The first submission
is manual; later versions can be submitted by the same command.

## Chocolatey

The package is `packaging/chocolatey` after rendering. It downloads the
architecture-specific portable `.exe`, verifies its checksum, and creates a
`bento` shim.

```sh
cd /tmp/bento-packages/chocolatey
choco pack
choco push bento-terminal.0.7.0.nupkg \
  --source https://push.chocolatey.org/ \
  --api-key "$CHOCO_API_KEY"
```

The Chocolatey package id is `bento-terminal`.

## Arch / AUR

The AUR package is `bento-terminal-bin`, built from the release binaries. Push
`PKGBUILD` and `.SRCINFO` to the AUR repository:

```sh
git clone ssh://aur@aur.archlinux.org/bento-terminal-bin.git
cp /tmp/bento-packages/aur/PKGBUILD /tmp/bento-packages/aur/.SRCINFO bento-terminal-bin/
cd bento-terminal-bin
git add PKGBUILD .SRCINFO
git commit -m "bento-terminal-bin 0.7.0"
git push
```

Regenerate `.SRCINFO` with `makepkg --printsrcinfo > .SRCINFO` on Arch before
pushing; the renderer writes the equivalent fields.

## Linux packages

The `Packages` workflow builds `.deb`, `.rpm`, and `.apk` for amd64 and arm64
with [nfpm](https://nfpm.goreleaser.com), then attaches them to the GitHub
release. It runs after every release. To build locally:

```sh
go install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.43.0
python3 packaging/render.py --version 0.7.0 --sums SHA256SUMS --out /tmp/bento-packages
mkdir -p dist
curl -fsSL \
  https://github.com/StarkDroid/Bento-Terminal/releases/download/v0.7.0/bento_linux_amd64 \
  -o dist/bento_linux_amd64

for format in deb rpm apk; do
  nfpm package --config /tmp/bento-packages/nfpm/amd64.yaml \
    --packager "$format" --target dist
done
```

## Go

The module path is already installable:

```sh
go install github.com/StarkDroid/Bento-Terminal/cmd/bento@latest
```

This reports `dev` as the version because Go installs the module source rather
than a tagged release build. That is expected.

## Notes

- Never commit a token. The workflows read `NPM_TOKEN`, `TAP_TOKEN`,
  `SCOOP_TOKEN`, `CHOCO_API_KEY`, and `AUR_SSH_KEY` from repository secrets.
- The release workflow only publishes npm and Linux packages automatically.
  Homebrew, Scoop, AUR, Chocolatey, and WinGet need the one-time repository or
  account setup above.
- Update this file when a new source is added or a package id changes.
