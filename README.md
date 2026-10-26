# Bento terminal

**A full-screen terminal workspace for one project directory.**

Bento terminal previews files with syntax highlighting, browses the tree and the current branch's commits, and runs detected build commands — all inside a set of island-themed bento boxes. It is written in Go with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and is designed for the macOS Terminal first (Menlo or SF Mono, no Nerd Font).

[![CI](https://img.shields.io/github/actions/workflow/status/StarkDroid/Bento-Terminal/ci.yml?branch=main&style=flat-square&label=build&color=3DDECE)](https://github.com/StarkDroid/Bento-Terminal/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/StarkDroid/Bento-Terminal?style=flat-square&color=1F8F86)](https://github.com/StarkDroid/Bento-Terminal/releases/latest)
[![Go version](https://img.shields.io/github/go-mod/go-version/StarkDroid/Bento-Terminal?style=flat-square&color=3DDECE)](go.mod)
[![License](https://img.shields.io/github/license/StarkDroid/Bento-Terminal?style=flat-square&color=7DCE82)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux-071E26?style=flat-square)](https://github.com/StarkDroid/Bento-Terminal)
[![Go Report Card](https://goreportcard.com/badge/github.com/StarkDroid/Bento-Terminal?style=flat-square)](https://goreportcard.com/report/github.com/StarkDroid/Bento-Terminal)
[![Stars](https://img.shields.io/github/stars/StarkDroid/Bento-Terminal?style=flat-square&color=FF6F61)](https://github.com/StarkDroid/Bento-Terminal/stargazers)
[![Issues](https://img.shields.io/github/issues/StarkDroid/Bento-Terminal?style=flat-square&color=E9C46A)](https://github.com/StarkDroid/Bento-Terminal/issues)
[![PRs welcome](https://img.shields.io/badge/PRs-welcome-3DDECE?style=flat-square)](https://github.com/StarkDroid/Bento-Terminal/pulls)

## Why Bento

- **One window, one project.** Open `bento` in a directory and the whole project is in view. No tabs to manage.
- **Code-aware preview.** Files are detected by name or contents and highlighted with a palette that matches the app. Go, Python, and the common brace languages also get foldable functions and classes.
- **Your shortcuts, in your box.** Every box has a footer with its own keys. Press `?` in any box to see the full list and run a command from it.
- **Builds without setup.** Common project files become a run list. Automatic mode finds tests, builds, and lint commands; manual mode keeps a short, familiar set.
- **Fun to look at.** Rounded islands, a lagoon palette, and no dependence on your terminal theme.

## Layout

```
╭──────────────────────────────╮ ╭──────────────────╮
│ project                      │ │ ≡ Files          │
│ ⎇ branch                     │ │ ▾ src            │
│ Go · Make                    │ │   ≡ main.go      │
│                              │ ├──────────────────┤
│  12 │ selected file preview  │ │ ⎇ History        │
│                              │ │ abc1234 subject  │
│ ↑↓←→ move · [ ] fold · ? keys │ │ ↑↓ move · ? keys │
╰──────────────────────────────╯ ╰──────────────────╯

╭──────────────────────────────────────────────────╮
│ ▶ Build  auto                                    │
│ ▶ npm test                                       │
│ ──────────────────────────────────────────────── │
│ $ npm test                                       │
│ ↑↓ pick · enter run · a mode · ? keys            │
╰──────────────────────────────────────────────────╯
 ⎇ main  ↑1 ↓0  ● staged 0  ● unstaged 1  stash 0
```

- **Main.** The selected file, with line numbers, syntax highlighting, a character cursor, and fold markers. With no selection it shows an island home: project name, branch, and detected build system.
- **Files.** The project tree. Selection follows the cursor. Click a directory to open or close it; click a file to preview it.
- **History.** Commits on the current branch: short hash, subject, and a relative or short date. Enter or click previews the commit in Main.
- **Build.** Run commands in automatic or manual mode, plus the streaming output of the selected run.
- **Status.** Branch, ahead/behind, staged and unstaged counts (untracked files included), and the stash count.

Icons use Unicode that macOS Terminal can draw with Menlo or SF Mono: `⎇` branch, `↑` `↓` ahead/behind, `●` dirty, `▶` run, `✓` `✗` result, `▸` `▾` fold, `≡` files, `▌` cursor.

Each box has its own focus and selection. The focused box has a bright lagoon border, a brighter title, and a footer that lists its shortcuts. The selected row is a filled highlight, separate from the border.

Bento terminal needs a window of at least **72 × 22** cells.

### Empty states

- Not a git repository: the history box and the status bar say so; the home screen omits a branch name.
- Empty tree: the files box says the directory is empty.
- No detected commands: the build list says none were found. If the other mode has commands, it says to press `a`.
- Build idle: the output area says to select a command and press enter.
- Very small window: a single panel asks you to enlarge the terminal.

## Keys

Focus starts on the file tree. `tab` moves Files → History → Build → Main, then back to Files.

Every box shows its own shortcuts in the bottom row. Press `?` in a box to open **its** command list, then move with `↑` `↓` and run with `enter` — or press a listed shortcut directly. `?` or `esc` closes the list. Build commands work from any box, but they are listed under Build.

| Key | Action |
| --- | --- |
| `tab` / `shift+tab` | Next or previous box |
| `↑` `↓` or `k` `j` | Move the selection or the preview caret by a line |
| `←` `→` or `h` `l` | Move the preview caret by a character; close or open a directory in Files |
| `[` / `]` | Fold or unfold a function or class in Main |
| `pgup` / `pgdown` | Page the focused list, preview, or build output |
| `enter` | Open a file, expand a directory, preview a commit, run a command, or toggle a fold |
| `s` | Turn syntax highlighting on or off in Main |
| `a` | Toggle automatic and manual build mode |
| `r` | Rerun the last command |
| `c` | Cancel the running command |
| `?` | Command list for the focused box |
| `,` | Settings and about (version, links) |
| `esc` | Close the command list or settings |
| `q` or `ctrl+c` | Quit |

### Mouse

Bento terminal enables mouse reporting on start.

- Click a box to focus it.
- Click a file to preview it; click a directory to open or close it; click a commit to preview it.
- Click a line in Main to put the caret there. Clicking a foldable line (one with `▸` or `▾`) folds or unfolds that function or class.
- The scroll wheel scrolls the box under the pointer.

### Settings

Press `,` in any box for **Settings · About**. It shows the BENTO wordmark, the version of the running binary, a one-line description, and the repository link. `esc` or `,` closes it. `bento -v` prints the same version.

## Code preview

The preview shows at most 256 KB of a text file. When a language is detected, the title shows its name (for example `≡ main.go · Go`) and the code is highlighted. `s` turns highlighting off and the title says `Go off`; folding still works.

Highlighting is capped at 8,000 lines per file. Larger files stay plain so scrolling remains quick. Detection uses the file name first, then the content (including `#!` lines), through [Chroma](https://github.com/alecthomas/chroma).

Foldable blocks are found for:

- Go: `func` and `type ... struct` / `interface`
- Python: `def`, `async def`, and `class`, by indentation
- JavaScript, JSX, TypeScript, Java, C, C++, C#, Rust, Swift, Kotlin, Scala, Dart, PHP, Groovy, Solidity, Zig, Bash, Zsh: declaration blocks, by matched braces

Fold markers (`▾` open, `▸` closed) appear in the left margin. Folding a block hides its body and shows a `… N lines` summary.

## Build commands

Automatic mode is the default. It reads:

- `package.json` scripts, as `npm run <name>`
- `Makefile` targets, as `make <target>`
- `gradlew` tasks parsed from `build.gradle` or `build.gradle.kts`
- `go.mod`: `go test ./...`, `go build ./...`, `go vet ./...`
- `Cargo.toml`: `cargo test`, `cargo build`, `cargo check`, and `cargo run` when the package has a binary
- `pyproject.toml`: `python -m pytest`, and `python -m build` when a build system is declared

Manual mode keeps a shorter list: `npm start`, `npm test`, and `npm run build` when those scripts exist; `make`, plus `make test` and `make build` when those targets exist; `./gradlew test`, `build`, and `check`; `go test` and `go build`; `cargo test` and `cargo build`; `python -m pytest`.

`bento` with no arguments uses the working directory as the project root. `bento /path/to/project` opens that directory.

## Install

The end-user install is a shell script. It does not use Homebrew or npm. A published release binary is the path that does not need Go. The script looks for a release asset named `bento_<os>_<arch>` — `bento_darwin_arm64`, `bento_darwin_amd64`, `bento_linux_arm64`, or `bento_linux_amd64`. When that asset is missing, it builds from source and requires `git` and `go` on `PATH`.

Binary only, the default:

```sh
curl -fsSL https://raw.githubusercontent.com/StarkDroid/Bento-Terminal/main/scripts/install.sh | sh
```

The same binary-only install, stated explicitly:

```sh
curl -fsSL https://raw.githubusercontent.com/StarkDroid/Bento-Terminal/main/scripts/install.sh | sh -s -- --no-zsh
```

Install the binary and append a small snippet to `~/.zshrc`:

```sh
curl -fsSL https://raw.githubusercontent.com/StarkDroid/Bento-Terminal/main/scripts/install.sh | sh -s -- --with-zsh
```

The snippet adds `~/.local/bin` (or `BENTO_INSTALL_DIR`) to `PATH` and defines a `bento` function. It does not replace your theme or prompt.

The binary is installed to `~/.local/bin/bento`. Override the directory with `BENTO_INSTALL_DIR`, and the GitHub repo with `BENTO_REPO` (default `StarkDroid/Bento-Terminal`) or `BENTO_REPO_URL`.

Passing both `--with-zsh` and `--no-zsh` is an error. Running the script again with `--with-zsh` does not duplicate the snippet.

### From a clone

Contributors can build without the release asset:

```sh
go install ./cmd/bento
make install
```

`make install` copies `bento` to `~/.local/bin`. `PREFIX` changes that root (`PREFIX=$HOME/.local` by default, binary at `$PREFIX/bin/bento`).

`go install github.com/StarkDroid/Bento-Terminal/cmd/bento@latest` works once the module is published.

### macOS release binaries

Cross-compile from any platform with Go:

```sh
make darwin-arm64
make darwin-amd64
```

That writes `dist/bento_darwin_arm64` and `dist/bento_darwin_amd64`. The same commands are:

```sh
GOOS=darwin GOARCH=arm64 go build -o dist/bento_darwin_arm64 ./cmd/bento
GOOS=darwin GOARCH=amd64 go build -o dist/bento_darwin_amd64 ./cmd/bento
```

Publish those files as GitHub release assets to make `curl | sh` install without a toolchain.

## Develop

Bento terminal is Go, and the module version is Go 1.23 or newer. Tests do not call macOS-only APIs, so `go test ./...` runs on Linux and in CI.

```sh
go test ./...
make build
./bento
```

`bento -h` and `bento -v` print usage and the version. The UI lives in `internal/ui`; language detection, highlighting, and folding live in `internal/syntax`.

Please read [AGENTS.md](AGENTS.md) before changing layout, focus order, theme tokens, or keybindings.

## Contributing

Issues and pull requests are welcome. Small, focused changes are easiest to review.

1. Fork the repository and create a branch.
2. Run `go vet ./...` and `go test ./...`.
3. Add a test when you change behavior.
4. Update the README or AGENTS.md when keys, layout, packages, or the install flow change.

## License

MIT. Copyright 2026 Trishiraj. See [LICENSE](LICENSE).
