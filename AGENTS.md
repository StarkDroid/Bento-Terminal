# Bentro

Bentro is a full-screen terminal workspace for one project directory. It previews files, lists the tree and the current branch's commits, and runs detected build commands. The first supported terminal is macOS Terminal. The UI draws its own lagoon palette and does not read the user's theme.

When you change packages, focus order, theme tokens, keybindings, or the install flow, update this file in the same change.

## Package map

- `cmd/bentro` — process entry. Uses the working directory, or one optional directory argument, as the project root.
- `internal/ui` — Bubble Tea model, bento layout, theme, focus, keys, and rendering. Scroll position for the preview and the build output is a Bubbles viewport. The running indicator is a Bubbles spinner. Bindings are Bubbles keys.
- `internal/git` — parses `git status --porcelain=v2 --branch` and `git log` lines (`hash`, unix time, subject, tab-separated). `Load` shells out to git. Untracked paths count as unstaged.
- `internal/files` — lazy directory tree and a bounded text preview. Dot entries are hidden. `node_modules`, `vendor`, `target`, `__pycache__`, `dist`, `.next`, `.venv`, and `venv` are skipped. Symlinks are not followed.
- `internal/build` — `Inspect` reads project files and returns automatic and manual command lists. `Start` runs a command and cancels it by signaling the process group. No shell is inserted unless the command itself is a shell.
- `scripts/install.sh` — curl entrypoint. Release asset first, then `go build` from a local checkout or a clone.

There is no `internal/install`. The installer is shell only.

## Build and test

```sh
go test ./...
make build
make install
make darwin-arm64
make darwin-amd64
```

Go 1.23 or newer. Dependencies are Go modules. `make darwin-arm64` and `make darwin-amd64` set `GOOS=darwin` and write `dist/bentro_darwin_arm64` and `dist/bentro_darwin_amd64`, which are the installer's release asset names.

Tests cover command detection fixtures in `internal/build/testdata`, git status and log fixtures in `internal/git/testdata`, focus cycling, layout size, empty views, file preview, and a real `echo` run. A git binary is required for `internal/git` load tests; those tests skip when it is missing.

## UI conventions

Focus order is Main, Files, History, Build. The app starts on Files. Tab moves Files → History → Build → Main. Shift-Tab reverses.

The focused pane uses border `#3DDECE` and a bold title. Unfocused panes use border `#24555C`. A selected row uses fill `#1B5E5A` when its pane is focused and `#14343C` when it is not. Borders are not the selection.

Theme tokens:

| Token | Hex | Use |
| --- | --- | --- |
| ocean | `#071E26` | screen and gaps |
| panel | `#0C2832` | pane fill |
| status | `#04161C` | status bar |
| sand | `#F4E4C4` | text |
| sand dim | `#B7A88A` | hints, unfocused titles |
| lagoon | `#1F8F86` | mode label, secondary accent |
| lagoon bright | `#3DDECE` | focus border, focused title |
| coral | `#FF6F61` | dirty counts, failures |
| palm | `#7DCE82` | success, detected systems |

Every pane sets foreground and background. Do not rely on the terminal's default colors. Stay with Unicode that Menlo and SF Mono provide: `⎇ ↑ ↓ ● ▶ ✓ ✗ ▸ ▾ ≡` and rounded box drawing. Do not require a Nerd Font, and do not use braille spinner frames.

Build mode defaults to automatic. `a` switches to manual. `c` cancels the process group. `r` reruns the last argv.

## Install

`scripts/install.sh` accepts `--with-zsh` and `--no-zsh`. The default and `--no-zsh` install the binary only, to `~/.local/bin` unless `BENTRO_INSTALL_DIR` is set. `--with-zsh` appends a marked PATH export and a `bentro` function to `~/.zshrc` and does not edit the prompt. Both flags together is an error.

The script prefers `https://github.com/$BENTRO_REPO/releases/latest/download/bentro_<os>_<arch>` (`BENTRO_REPO` defaults to `trishiraj/bentro`). If that asset is not a Mach-O or ELF binary, it builds with Go. That fallback is the only path that needs a toolchain. Say so in the README if the asset names or the fallback change.
