# bento-terminal

A full-screen terminal workspace for one project directory: preview and edit
files, read the tree, commits and working-tree changes, run builds, open a
shell, and run coding agents beside the preview.

This package installs the native Bento terminal binary for macOS, Linux, or
Windows and exposes the `bento` command.

## Run without installing

```sh
npx bento-terminal
```

Open a specific project:

```sh
npx bento-terminal ~/Projects/my-app
```

## Install globally

```sh
npm install -g bento-terminal
bento
```

Then run `bento` inside any project. With no argument it uses the current
directory; `bento <dir>` opens another one.

## How it works

The install downloads `bento_<os>_<arch>` from the matching
[Bento terminal release](https://github.com/StarkDroid/Bento-Terminal/releases)
and verifies it against the release's `SHA256SUMS`. The downloaded binary lives
inside the package; the `bento` command starts it with inherited stdio so the
TUI works normally.

Supported hosts: macOS arm64/x64, Linux arm64/x64, Windows arm64/x64.

If `npm install --ignore-scripts` is used, the launcher downloads the binary on
first run instead.

## Environment

| Variable | Purpose |
| --- | --- |
| `BENTO_VERSION` | Download a release other than the package version. |
| `BENTO_BASE_URL` | Use a mirror or a local release directory. |
| `BENTO_SKIP_DOWNLOAD` | Skip the postinstall download. |
| `BENTO_FORCE_DOWNLOAD` | Replace an already installed binary. |

## Other install routes

- Homebrew: `brew install StarkDroid/tap/bento`
- Scoop: `scoop install bento`
- WinGet: `winget install StarkDroid.Bento`
- Chocolatey: `choco install bento-terminal`
- Arch: `yay -S bento-terminal-bin`
- Go: `go install github.com/StarkDroid/Bento-Terminal/cmd/bento@latest`
- Direct release binary: [GitHub releases](https://github.com/StarkDroid/Bento-Terminal/releases/latest)

## License

MIT. See [LICENSE](LICENSE).
