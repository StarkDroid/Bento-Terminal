# Bentro

Bentro is an open-source terminal tool that feels like a small IDE. It opens the project in the current directory as a full-screen workspace: a text preview, a directory tree, commit history, and a build window.

The interface is an island-themed bento layout. Panels are separate rounded islands with gaps between them. Bentro paints its own colors, so the macOS Terminal theme does not change the UI. The first platform is macOS Terminal with its default font (Menlo or SF Mono). Bentro does not need a Nerd Font.

## Layout

```
╭──────────────────────────────╮ ╭──────────────────╮
│ project                      │ │ ≡ Files          │
│ ⎇ branch                     │ │ ▾ src            │
│ Go · Make                    │ │   ≡ main.go      │
│                              │ ├──────────────────┤
│  12 │ selected file preview  │ │ ⎇ History        │
│                              │ │ abc1234 subject  │
╰──────────────────────────────╯ ╰──────────────────╯

╭──────────────────────────────────────────────────╮
│ ▶ Build  auto                                    │
│ ▶ npm test                                       │
│ ──────────────────────────────────────────────── │
│ $ npm test                                       │
╰──────────────────────────────────────────────────╯
 ⎇ main  ↑1 ↓0  ● staged 0  ● unstaged 1  stash 0
```

- **Main pane.** A viewer for the selected text file, with line numbers. When nothing is selected, it shows an island home: project name, branch, and detected build system.
- **Files.** The project tree. Scroll and select. Enter opens a file. Enter, or the left and right arrows, expands or collapses a directory.
- **History.** Commits on the current branch: short hash, subject, and a relative or short date. Enter previews that commit in the main pane.
- **Build.** A list of run commands and the streaming output of the selected run. Automatic mode detects commands from the project. Manual mode lists the primary commands for the ecosystems that are present. A running command can be cancelled.
- **Status.** Branch name, ahead/behind versus upstream, staged count, unstaged count (untracked files included), and stash count.

Icons use Unicode that macOS Terminal can draw with Menlo or SF Mono: `⎇` branch, `↑` `↓` ahead/behind, `●` dirty, `▶` run, `✓` `✗` result, `▸` `▾` tree, `≡` files.

Each pane has its own focus and selection. The focused pane has a bright lagoon border and a brighter title. Other panes use a muted border. The selected row is a filled highlight, separate from the border.

### Empty states

- Not a git repository: the history pane and the status bar say so. The home screen omits a branch name.
- Empty tree: the files pane says the directory is empty.
- No detected commands: the build list says none were found. If the other mode has commands, it says to press `a`.
- Build idle: the output area says to select a command and press enter.

`bentro` with no arguments uses the working directory as the project root. `bentro /path/to/project` opens that directory.

## Keys

Focus starts on the file tree. Tab moves Files → History → Build → Main, then back to Files.

| Key | Action |
| --- | --- |
| `tab` / `shift+tab` | Next or previous pane |
| `↑` `↓` or `k` `j` | Move the selection, or scroll the focused preview |
| `←` `→` or `h` `l` | Collapse or expand a directory |
| `pgup` / `pgdown` | Page the focused list, preview, or build output |
| `enter` | Open a file, expand a directory, preview a commit, or run the selected command |
| `a` | Toggle automatic and manual build mode |
| `r` | Rerun the last command |
| `c` | Cancel the running command |
| `?` | Toggle help |
| `esc` | Close help |
| `q` or `ctrl+c` | Quit |

Automatic mode is the default. It reads:

- `package.json` scripts, as `npm run <name>`
- `Makefile` targets, as `make <target>`
- `gradlew` tasks parsed from `build.gradle` or `build.gradle.kts`
- `go.mod`: `go test ./...`, `go build ./...`, `go vet ./...`
- `Cargo.toml`: `cargo test`, `cargo build`, `cargo check`, and `cargo run` when the package has a binary
- `pyproject.toml`: `python -m pytest`, and `python -m build` when a build system is declared

Manual mode keeps a shorter list: `npm start`, `npm test`, and `npm run build` when those scripts exist; `make`, plus `make test` and `make build` when those targets exist; `./gradlew test`, `build`, and `check`; `go test` and `go build`; `cargo test` and `cargo build`; `python -m pytest`.

## Install

The end-user install is a shell script. It does not use Homebrew or npm. A published release binary is the path that does not need Go. The script looks for a release asset named `bentro_darwin_arm64` or `bentro_darwin_amd64` (and the matching Linux name if you run it there). When that asset is missing, it builds from source and requires `git` and `go` on `PATH`.

Binary only, the default:

```sh
curl -fsSL https://raw.githubusercontent.com/trishiraj/bentro/main/scripts/install.sh | sh
```

The same binary-only install, stated explicitly:

```sh
curl -fsSL https://raw.githubusercontent.com/trishiraj/bentro/main/scripts/install.sh | sh -s -- --no-zsh
```

Install the binary and append a small snippet to `~/.zshrc`. The snippet adds `~/.local/bin` (or `BENTRO_INSTALL_DIR`) to `PATH` and defines a `bentro` function. It does not replace your theme or prompt:

```sh
curl -fsSL https://raw.githubusercontent.com/trishiraj/bentro/main/scripts/install.sh | sh -s -- --with-zsh
```

The binary is installed to `~/.local/bin/bentro`. Override the directory with `BENTRO_INSTALL_DIR`, and the GitHub repo with `BENTRO_REPO` (default `trishiraj/bentro`) or `BENTRO_REPO_URL`.

Passing both `--with-zsh` and `--no-zsh` is an error. Running the script again with `--with-zsh` does not duplicate the snippet.

### From a clone

Contributors can build without the release asset:

```sh
go install ./cmd/bentro
make install
```

`make install` copies `bentro` to `~/.local/bin`. `PREFIX` changes that root (`PREFIX=$HOME/.local` by default, binary at `$PREFIX/bin/bentro`).

`go install github.com/trishiraj/bentro/cmd/bentro@latest` works once the module is published.

### macOS release binaries

Cross-compile from any platform with Go:

```sh
make darwin-arm64
make darwin-amd64
```

That writes `dist/bentro_darwin_arm64` and `dist/bentro_darwin_amd64`. The same commands are:

```sh
GOOS=darwin GOARCH=arm64 go build -o dist/bentro_darwin_arm64 ./cmd/bentro
GOOS=darwin GOARCH=amd64 go build -o dist/bentro_darwin_amd64 ./cmd/bentro
```

Publish those files as GitHub release assets to make `curl | sh` install without a toolchain.

## Develop

Bentro is Go. Tests do not call macOS-only APIs, so `go test ./...` runs on Linux.

```sh
go test ./...
make build
./bentro
```

`bentro -h` and `bentro -v` print usage and the version.

## License

MIT. Copyright 2026 Trishiraj. See [LICENSE](LICENSE).
