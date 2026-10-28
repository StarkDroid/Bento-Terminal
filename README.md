![Bento terminal — one project, multi-pane](assets/banner.png)

Bento terminal is a full-screen workspace for one project, in the terminal: preview and edit files, read the tree, commits and working-tree changes, run builds, and open a shell — all in one keyboard-first bento layout. It is written in Go with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and runs in the macOS Terminal or Windows Terminal (Menlo, SF Mono, Cascadia Mono; no Nerd Font).

[![CI](https://img.shields.io/github/actions/workflow/status/StarkDroid/Bento-Terminal/ci.yml?branch=main&style=flat-square&label=build)](https://github.com/StarkDroid/Bento-Terminal/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/StarkDroid/Bento-Terminal?style=flat-square)](https://github.com/StarkDroid/Bento-Terminal/releases/latest)
[![Go version](https://img.shields.io/github/go-mod/go-version/StarkDroid/Bento-Terminal?style=flat-square)](go.mod)
[![License](https://img.shields.io/github/license/StarkDroid/Bento-Terminal?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows?style=flat-square)](https://github.com/StarkDroid/Bento-Terminal)
[![Go Report Card](https://goreportcard.com/badge/github.com/StarkDroid/Bento-Terminal?style=flat-square)](https://goreportcard.com/report/github.com/StarkDroid/Bento-Terminal)
[![Stars](https://img.shields.io/github/stars/StarkDroid/Bento-Terminal?style=flat-square)](https://github.com/StarkDroid/Bento-Terminal/stargazers)
[![Issues](https://img.shields.io/github/issues/StarkDroid/Bento-Terminal?style=flat-square)](https://github.com/StarkDroid/Bento-Terminal/issues)
[![PRs welcome](https://img.shields.io/badge/PRs-welcome?style=flat-square)](https://github.com/StarkDroid/Bento-Terminal/pulls)

**Site** [starkdroid.github.io/Bento-Terminal](https://starkdroid.github.io/Bento-Terminal/) · **Docs** [install, keys, layout](https://starkdroid.github.io/Bento-Terminal/docs/) · **Releases** [latest](https://github.com/StarkDroid/Bento-Terminal/releases/latest)

## Install

macOS and Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/StarkDroid/Bento-Terminal/main/scripts/install.sh | sh
```

Windows, in Windows Terminal:

```powershell
irm https://raw.githubusercontent.com/StarkDroid/Bento-Terminal/main/scripts/install.ps1 | iex
```

The Unix installer puts the binary in `~/.local/bin`; `--with-zsh` also adds that folder to PATH. The Windows installer puts `bento.exe` in `%LOCALAPPDATA%\bento` and adds it to your user PATH, so a new Windows Terminal finds `bento`. Flags, install directories, and the source-build path are in the [install docs](https://starkdroid.github.io/Bento-Terminal/docs/#install).

## What it does

- **Preview and edit.** Highlighting, folding, and a save in place.
- **Every change.** Modified, staged, untracked, and conflicted paths, with patches.
- **Your shell.** Any command, run inside the project.
- **Builds without setup.** Detected tests, builds, and lint commands.

Every build ships the same panes, keys, and behaviour. Only the shell the Terminal tab runs (`$SHELL` on macOS and Linux, `%COMSPEC%` on Windows) and the installer differ per platform.

Editing, changes, the terminal, build detection, shortcuts, mouse, and layout are in the [docs](https://starkdroid.github.io/Bento-Terminal/docs/).

## Develop

macOS and Linux:

```sh
go test ./...
make build
./bento
sh scripts/install.sh --from-source   # install the checkout, not the release
```

Windows, in Windows Terminal:

```powershell
go test ./...
go build -o bento.exe ./cmd/bento
.\bento.exe
.\scripts\install.ps1 -FromSource     # install the build so `bento` works
```

`go run ./cmd/bento` is the quickest loop on any platform; it runs from a temp directory and touches neither PATH nor the repository. Cross-compile with `make darwin-arm64`, `make darwin-amd64`, `make linux-arm64`, `make linux-amd64`, `make windows-amd64`, and `make windows-arm64`; without `make`, use `go build -o dist\bento_windows_amd64.exe ./cmd/bento`.

Read [AGENTS.md](AGENTS.md) before changing layout, focus order, theme tokens, or keybindings. The site lives in `website/`; [its README](website/README.md) covers the build. Issues and pull requests are welcome.

## License

MIT. Copyright 2026 Trishiraj. See [LICENSE](LICENSE).
