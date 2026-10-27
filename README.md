# Bento terminal

**A full-screen terminal workspace for one project directory.**

Bento terminal previews and edits files with syntax highlighting, browses the tree, the current branch's commits, and its working-tree changes, and runs detected build commands or your own shell commands — all inside a set of island-themed bento boxes. It is written in Go with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and is designed for the macOS Terminal first (Menlo or SF Mono, no Nerd Font).

[![CI](https://img.shields.io/github/actions/workflow/status/StarkDroid/Bento-Terminal/ci.yml?branch=main&style=flat-square&label=build&color=3DDECE)](https://github.com/StarkDroid/Bento-Terminal/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/StarkDroid/Bento-Terminal?style=flat-square&color=1F8F86)](https://github.com/StarkDroid/Bento-Terminal/releases/latest)
[![Go version](https://img.shields.io/github/go-mod/go-version/StarkDroid/Bento-Terminal?style=flat-square&color=3DDECE)](go.mod)
[![License](https://img.shields.io/github/license/StarkDroid/Bento-Terminal?style=flat-square&color=7DCE82)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux-071E26?style=flat-square)](https://github.com/StarkDroid/Bento-Terminal)
[![Go Report Card](https://goreportcard.com/badge/github.com/StarkDroid/Bento-Terminal?style=flat-square)](https://goreportcard.com/report/github.com/StarkDroid/Bento-Terminal)

**Site:** [starkdroid.github.io/Bento-Terminal](https://starkdroid.github.io/Bento-Terminal/) · **Docs:** [install, shortcuts, layout](https://starkdroid.github.io/Bento-Terminal/docs/)
[![Stars](https://img.shields.io/github/stars/StarkDroid/Bento-Terminal?style=flat-square&color=FF6F61)](https://github.com/StarkDroid/Bento-Terminal/stargazers)
[![Issues](https://img.shields.io/github/issues/StarkDroid/Bento-Terminal?style=flat-square&color=E9C46A)](https://github.com/StarkDroid/Bento-Terminal/issues)
[![PRs welcome](https://img.shields.io/badge/PRs-welcome-3DDECE?style=flat-square)](https://github.com/StarkDroid/Bento-Terminal/pulls)

## Why Bento

- **One window, one project.** Open `bento` in a directory and the whole project is in view. No tabs to manage.
- **Code-aware preview.** Files are detected by name or contents and highlighted with a palette that matches the app. Go, Python, and the common brace languages also get foldable functions and classes.
- **Edit without leaving.** `e` turns the preview into an editor, `ctrl+s` saves, and the `+ new file` tab or `n` creates a new file in the project.
- **Changes at a glance.** History has a **Changes** tab that lists every modified, staged, untracked, and conflicted path and shows its patch.
- **A terminal in the box.** Build has a **Terminal** tab that runs anything you type through your own shell, inside the project.
- **Your shortcuts, in your box.** Every box has a footer with its own keys. Press `?` in any box to see the full list and run a command from it.
- **Builds without setup.** Common project files become a run list. Automatic mode finds tests, builds, and lint commands; manual mode keeps a short, familiar set.
- **Fun to look at.** Rounded islands, a lagoon palette, and no dependence on your terminal theme.

## Layout

```
╭──────────────────────────────╮ ╭──────────────────────╮
│ ≡ main.go  ·  Go  + new file │ │ ≡ Files              │
│  1 │ package main            │ │ ▾ src                │
│  2 │                         │ │   ≡ main.go          │
│  3 │ func main() {           │ ├──────────────────────┤
│  4 │     print("hi")         │ │ ⎇ History Changes    │
│  5 │ }                       │ │ abc1234 subject      │
│                              │ │ abc1233 older        │
│ ↑↓←→ move · e edit · ? keys  │ │ ↑↓ · ←→ tab · ? keys │
╰──────────────────────────────╯ ╰──────────────────────╯

╭──────────────────────────────────────────────────╮
│ ▶ Build  Terminal  auto                          │
│ ▶ npm test                                       │
│ ──────────────────────────────────────────────── │
│ $ npm test                                       │
│ ↑↓ pick · enter run · ←→ tab · ? keys            │
╰──────────────────────────────────────────────────╯
 ⎇ main  ↑1 ↓0  ● staged 0  ● unstaged 1  stash 0
```

- **Main.** The selected file, with line numbers, syntax highlighting, a character cursor, and fold markers. The file name sits in a tab, and the `+ new file` tab right beside it, or `n`, creates a new file. `e` edits the file and `ctrl+s` saves it. With no selection it shows an island home: project name, branch, and detected build system.
- **Files.** The project tree. Selection follows the cursor. Click a directory to open or close it; click a file to preview it.
- **History.** Two tabs. **History** lists commits on the current branch, and **Changes** lists every staged, unstaged, untracked, and unmerged path. Enter or click shows a commit or a patch in Main.
- **Build.** Two tabs. **Build** runs detected commands in automatic or manual mode and streams their output. **Terminal** runs what you type through your shell and keeps its own scrollback and command history.
- **Status.** Branch, ahead/behind, staged and unstaged counts (untracked files included), and the stash count.

Icons use Unicode that macOS Terminal can draw with Menlo or SF Mono: `⎇` branch, `↑` `↓` ahead/behind, `●` dirty, `▶` run, `✓` `✗` result, `▸` `▾` fold, `≡` files, `±` patch, `▌` cursor, `+` new file.

Tabs are chips on the title bar. The active tab has the lighter fill and brighter text; the other tabs sit darker beside it. Tabs switch with `←` `→` or a click, and the open file and the `+ new file` tab sit side by side like browser tabs.

Each box has its own focus and selection. The focused box has a bright lagoon border, a brighter title, and a footer that lists its shortcuts. The selected row is a filled highlight, separate from the border.

Bento terminal needs a window of at least **72 × 22** cells.

### Empty states

- Not a git repository: the history box and the status bar say so; the home screen omits a branch name.
- Empty tree: the files box says the directory is empty.
- Clean working tree: the Changes tab says so.
- No detected commands: the build list says none were found. If the other mode has commands, it says to press `a`.
- Build idle: the output area says to select a command and press enter.
- Terminal idle: the prompt says to type a command and press enter.
- Very small window: a single panel asks you to enlarge the terminal.

## Keys

Focus starts on the file tree. `tab` moves Files → History → Build → Main, then back to Files.

Every box shows its own shortcuts in the bottom row. Press `?` in a box to open **its** command list, then move with `↑` `↓` and run with `enter` — or press a listed shortcut directly. `?` or `esc` closes the list. Build commands work from any box, but they are listed under Build.

| Key | Action |
| --- | --- |
| `tab` / `shift+tab` | Next or previous box |
| `↑` `↓` or `k` `j` | Move the selection or the preview caret by a line |
| `←` `→` or `h` `l` | Move the preview caret by a character; close or open a directory in Files; switch History and Build tabs |
| `[` / `]` | Fold or unfold a function or class in Main |
| `pgup` / `pgdown` | Page the focused list, preview, or output |
| `enter` | Open a file, expand a directory, show a commit or patch, run a command, split a line while editing, create the typed file, or toggle a fold |
| `e` | Edit the open file |
| `ctrl+s` | Save the file while editing |
| `n` or `+` | Create a new file (path relative to the project) |
| `s` | Turn syntax highlighting on or off in Main |
| `a` | Toggle automatic and manual build mode |
| `r` | Rerun the last build command |
| `c` | Cancel the running command |
| `?` | Command list for the focused box |
| `,` | Settings and about (version, links) |
| `esc` | Close the command list or settings; leave the editor; clear the terminal prompt, then return to the Build tab |
| `ctrl+c` | Stop a running command; in the terminal, stop it or clear the prompt |
| `q` or `ctrl+c` | Quit |

While the **Terminal** tab is focused, every printable key goes to the prompt, including `q`, `?`, and `,`; the command list and settings open only when the prompt is empty. `tab` still moves focus. While the editor is open, printable keys type, so `[` `]` `e` `n` `s` `q` are text; `esc` leaves the editor.

### Mouse

Bento terminal enables mouse reporting on start.

- Click a box to focus it, and click a tab label to switch tabs.
- In Main's title, click `+ new file` to create a file.
- Click a file to preview it; click a directory to open or close it; click a commit or a changed path to preview it.
- Click a line in Main to put the caret there. Clicking a foldable line (one with `▸` or `▾`) folds or unfolds that function or class.
- The scroll wheel scrolls the box under the pointer, including the terminal scrollback.

### Settings

Press `,` in any box for **Settings · About**. It shows the BENTO wordmark, the version of the running binary, a one-line description, and the repository link. `esc` or `,` closes it. `bento -v` prints the same version.

## Code preview

The preview shows at most 256 KB of a text file. The file's name is a tab, with the `+ new file` tab beside it. When a language is detected, the tab shows its name too (for example `≡ main.go · Go`) and the code is highlighted. `s` turns highlighting off and the tab says `Go off`; folding still works.

Highlighting is capped at 8,000 lines per file. Larger files stay plain so scrolling remains quick. Detection uses the file name first, then the content (including `#!` lines), through [Chroma](https://github.com/alecthomas/chroma).

Foldable blocks are found for:

- Go: `func` and `type ... struct` / `interface`
- Python: `def`, `async def`, and `class`, by indentation
- JavaScript, JSX, TypeScript, Java, C, C++, C#, Rust, Swift, Kotlin, Scala, Dart, PHP, Groovy, Solidity, Zig, Bash, Zsh: declaration blocks, by matched braces

Fold markers (`▾` open, `▸` closed) appear in the left margin. Folding a block hides its body and shows a `… N lines` summary.

### Editing

`e` edits the open text file, and the title says `edit`. Arrow keys move the caret, `enter` splits a line, `backspace` and `delete` remove characters and join lines, `home` and `end` jump within the line, and `tab` inserts four spaces. Printable keys type.

`ctrl+s` saves. The title shows `●` while the buffer differs from the file, `✓` after a save, and `save failed` when the write fails. `esc` leaves the editor and keeps unsaved text in the preview, where `ctrl+s` still saves it. Highlighting and folding come back when the editor closes. A binary file, a commit, and a patch are read-only.

Files are always written inside the project. A new path creates its parent directories; a path that climbs out of the project is rejected. New files open straight in the editor, and saving a `.go`, `.py`, or any other detected name highlights it like any other file.

## Changes

History's **Changes** tab lists every path git reports, in git's order, with its two porcelain columns:

- `M ` staged, ` M` unstaged, `MM` both
- `A ` added, `??` untracked
- `R ` renamed, shown as `old → new`
- `UU` and other unmerged codes

Enter or click opens the patch in Main under a `±` title. Staged and unstaged changes are shown together, and an untracked file gets a generated new-file patch. A clean tree says the working tree is clean. The tab title carries the number of changed paths.

## Terminal

Build's **Terminal** tab runs whatever you type through `$SHELL -c`, falling back to `/bin/sh`, inside the project root. Enter runs, `esc` clears the prompt (and then returns to the Build tab), and `ctrl+c` stops a running command or clears the prompt. `↑` `↓` recall earlier commands, and the wheel scrolls the output.

The terminal keeps its own scrollback, separate from the Build output. Starting a command stops the previous one.

Output is captured as text, so a command cannot clear or repaint Bento. Colour, cursor, and erase sequences are dropped before they reach the scrollback, a progress-bar redraw keeps only its last line, and a command that writes nothing but control sequences (`clear`) leaves no trace. Full-screen programs such as `vim` or `top` are not supported; the tab is for one-shot commands, and Build is the place for project commands whose output you want to read.

## Build commands

Automatic mode is the default. It reads:

- `package.json` scripts, as `npm run <name>`
- `Makefile` targets, as `make <target>`
- `gradlew` tasks parsed from `build.gradle` or `build.gradle.kts`
- `go.mod`: `go test ./...`, `go build ./...`, `go vet ./...`
- `Cargo.toml`: `cargo test`, `cargo build`, `cargo check`, and `cargo run` when the package has a binary
- `pyproject.toml`: `python -m pytest`, and `python -m build` when a build system is declared

Manual mode keeps a shorter list: `npm start`, `npm test`, and `npm run build` when those scripts exist; `make`, plus `make test` and `make build` when those targets exist; `./gradlew test`, `build`, and `check`; `go test` and `go build`; `cargo test` and `cargo build`; `python -m pytest`.

Detected commands run directly, without a shell, so pipes, globs, and builtins belong in the Terminal tab.

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

### macOS and Linux release binaries

Cross-compile from any platform with Go:

```sh
make darwin-arm64
make darwin-amd64
make linux-arm64
make linux-amd64
```

That writes `dist/bento_darwin_arm64`, `dist/bento_darwin_amd64`, `dist/bento_linux_arm64`, and `dist/bento_linux_amd64`. The same commands are:

```sh
GOOS=darwin GOARCH=arm64 go build -o dist/bento_darwin_arm64 ./cmd/bento
GOOS=darwin GOARCH=amd64 go build -o dist/bento_darwin_amd64 ./cmd/bento
GOOS=linux GOARCH=arm64 go build -o dist/bento_linux_arm64 ./cmd/bento
GOOS=linux GOARCH=amd64 go build -o dist/bento_linux_amd64 ./cmd/bento
```

The release script publishes those four files, plus `SHA256SUMS`, so `curl | sh` installs without a toolchain.

## Releasing

Every change lands on `main` through a pull request, and merging publishes nothing. Each PR body carries a changelog written for users:

```markdown
Release: 0.5.0

## Changelog

- The preview now follows the arrow keys and folds with [ and ].
- Settings shows the running version.
```

Keep those bullets current as the branch grows; every PR merged since the last release contributes to the next set of notes. Leave `Release:` blank on work that should not cut a release.

After merging the release pull request yourself, run the release script on `main`:

```sh
scripts/release.sh                 # version from the merged PR's Release: line
make release V=0.5.0               # or name the version yourself
scripts/release.sh --dry-run       # print the notes and stop
scripts/release.sh --since v0.4.0  # release from an earlier tag
```

It collects the `## Changelog` bullets from every PR merged since the last release, runs the tests, builds the four binaries with the version in the build metadata, tags `v0.5.0`, and publishes the release with `SHA256SUMS`. A PR without a changelog falls back to its cleaned-up title.

The script needs `git`, `go`, and the [GitHub CLI](https://cli.github.com) signed in. It refuses to run on a dirty tree, off `main`, or for a tag that already exists, and it prints a note when `cmd/bento/main.go` still carries a different version from the one being released.

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
5. Fill in the user-facing `## Changelog` in the PR body, and add `Release:` only when the PR should cut a release. Merging publishes nothing; the release step runs `scripts/release.sh` afterwards.

## License

MIT. Copyright 2026 Trishiraj. See [LICENSE](LICENSE).
