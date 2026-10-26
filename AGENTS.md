# Bento terminal

Bento terminal is a full-screen terminal workspace for one project directory. It previews files with syntax highlighting and folding, lists the tree and the current branch's commits, and runs detected build commands. The first supported terminal is macOS Terminal. The UI draws its own lagoon palette and does not read the user's theme.

Every box draws its own shortcut footer. `?` opens that box's command list, and a listed command can be run with enter or by pressing its shortcut. When you change packages, commands, focus order, theme tokens, keybindings, or the install flow, update this file in the same change.

## Package map

- `cmd/bento` — process entry. Uses the working directory, or one optional directory argument, as the project root. Starts Bubble Tea with the alternate screen and mouse cell motion.
- `internal/ui` — Bubble Tea model, bento layout, theme, focus, keys, and rendering. Scroll position for the preview and the build output is a Bubbles viewport. The running indicator is a Bubbles spinner. Bindings are Bubbles keys.
  - `commands.go` — one action table per box. Actions carry their binding and a pane-specific description, and `runAction` is the single dispatcher. This table drives both key handling and the `?` command list.
  - `palette.go` — the per-box command list: open/close, cursor movement, enter to run, direct shortcut to run.
  - `settings.go` — the settings overlay: wordmark, version, and about.
  - `codeview.go` — renders the file preview: gutter, fold markers, caret row and block, highlighted spans, and collapsed-block summaries.
  - `preview.go` — line numbering and wrapping helpers for plain previews.
- `internal/syntax` — Chroma-based language detection, highlighting with the lagoon palette, and foldable block discovery. Go and Python use dedicated strategies; the common brace languages share a balanced-brace scanner. Highlighting is capped at 8,000 lines.
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
make linux-arm64
make linux-amd64
```

Go 1.23 or newer. Dependencies are Go modules. The platform targets set `GOOS` and `GOARCH` and write `dist/bento_<os>_<arch>`, which are the installer's release asset names.

Tests cover command detection fixtures in `internal/build/testdata`, git status and log fixtures in `internal/git/testdata`, language detection and highlighting, Go/Python/brace folding, caret navigation, settings, focus cycling, layout size, empty views, file preview, per-box footers and command lists, keyboard and mouse folding, and a real `echo` run. A git binary is required for `internal/git` load tests; those tests skip when it is missing.

CI runs gofmt, `go vet`, `go test`, and `make build` on Ubuntu and macOS.

`.github/workflows/release.yml` runs on every push to `main`. It finds the merged pull request, reads `Release:` and the `## Changelog` bullets from the PR body, collects notes from every PR merged since the last release, runs the tests, builds `bento_<os>_<arch>` for macOS and Linux with `main.version` set, writes `SHA256SUMS`, and creates the GitHub release. The PR template asks for that format; commit subjects are only a fallback. Bump the version in `cmd/bento/main.go` and `Makefile` in the same change.

## UI conventions

The minimum window is **72 × 22** cells; smaller windows show a single "make the terminal larger" panel.

Focus order is Main, Files, History, Build. The app starts on Files. Tab moves Files → History → Build → Main. Shift-Tab reverses. Changing focus re-renders the preview so the cursor row only highlights while Main is focused.

The focused pane uses border `#3DDECE` and a bold title. Unfocused panes use border `#24555C`. A selected row uses fill `#1B5E5A` when its pane is focused and `#14343C` when it is not. Borders are not the selection.

### Footers and command lists

Each box reserves its last body row for a footer of its own shortcuts. Footer hints are:

| Box | Hints |
| --- | --- |
| Main, highlighted file | `↑↓←→ move · [ ] fold · ? keys` |
| Main, file with no detected language | `↑↓←→ move · ? keys` |
| Main, commit | `↑↓ scroll · ? keys` |
| Main, home | `? keys · tab panes` |
| Files | `↑↓ · enter open · ? keys` |
| History | `↑↓ · enter show · ? keys` |
| Build | `↑↓ pick · enter run · a mode · r rerun · c cancel · ? keys` |

`?` opens the command list for the focused box (`? Keys · Files`). The list shows that box's commands and then a Global section. `↑` `↓` (or `k` `j`) move, `pgup`/`pgdown` jump, `enter` runs the selected row, and pressing a listed shortcut runs it directly and closes the list. `?` and `esc` close it. While the list is open, focus does not move, and mouse events are ignored.

Build commands (`a`, `r`, `c`) are listed under Build but work from any box. `q` quits even while the list is open.

`,` opens the settings overlay from any box. It draws the BENTO wordmark, the version passed to `ui.New` by `cmd/bento`, and a short about. `esc` or `,` closes it. Mouse events are ignored while it is open.

To add a command: add the binding to `keyMap`, add an action to the right `paneActions` case (and `buildActions` if it must work globally), and handle its id in `runAction`. Update the footer hints if the command is important enough to show there.

### Code preview

The preview is capped at 256 KB. The title is ` ≡ name` plus `  ·  Language` when a language is detected. `s` toggles highlighting and the title then says `Language off`; folding still works with highlighting off.

The Main pane has a source caret. `↑` `↓` move it by line, `←` `→` move it by character (wrapping at line ends), `pgup`/`pgdown` page it, and the viewport follows. `[` folds the innermost function or class at the caret, `]` unfolds it, and `enter` toggles it. A folded block keeps its opening line with a `▸` marker and a `… N lines` summary; open blocks use `▾`. The caret row uses fill `#14343C`, a `▌` marker when the line is not a fold start, and a lagoon-bright block over the character at the caret.

Commit previews have no gutter, cursor, highlighting, or folding; they scroll as plain wrapped text.

### Mouse

Mouse reporting is on (`tea.WithMouseCellMotion`). A left click focuses the box under the pointer and moves its selection; clicking a directory toggles it, clicking a file opens the preview, clicking a commit previews it. In Main, a click puts the caret at the start of that source line and, when the line opens a foldable block, toggles that block. The wheel scrolls the box under the pointer by three rows (Build scrolls its output).

`paneAt` in `layout.go` maps screen cells to boxes; body-top helpers keep click rows aligned with the rendered content.

Theme tokens:

| Token | Hex | Use |
| --- | --- | --- |
| ocean | `#071E26` | screen and gaps |
| panel | `#0C2832` | pane fill |
| status | `#04161C` | status bar |
| sand | `#F4E4C4` | text |
| sand dim | `#B7A88A` | hints, unfocused titles |
| lagoon | `#1F8F86` | mode label, secondary accent, focused footer |
| lagoon bright | `#3DDECE` | focus border, focused title, keywords |
| coral | `#FF6F61` | dirty counts, failures, tags |
| palm | `#7DCE82` | success, detected systems, function and class names |
| sun | `#E9C46A` | strings, decorators, collapsed fold marker |
| amber | `#F4A261` | numbers |
| moss | `#5F7D7B` | comments, gutter |
| reef | `#5FB3C4` | types and builtins |

Every pane sets foreground and background. Do not rely on the terminal's default colors. The syntax palette lives in `internal/syntax/highlight.go` and mirrors these tokens. Stay with Unicode that Menlo and SF Mono provide: `⎇ ↑ ↓ ● ▶ ✓ ✗ ▸ ▾ ≡ ▌ · …` and rounded box drawing. Do not require a Nerd Font, do not use braille spinner frames, and do not use `⏎` in hints (write `enter`).

Build mode defaults to automatic. `a` switches to manual. `c` cancels the process group. `r` reruns the last argv.

### Empty states

- Not a git repository: the history box and the status bar say so. The home screen omits a branch name.
- Empty tree: the files box says the directory is empty.
- No detected commands: the build list says none were found. If the other mode has commands, it says to press `a`.
- Build idle: the output area says to select a command and press enter.
- No language detected in a file: no language badge, no fold markers, and `s` does nothing.

## Install

`scripts/install.sh` accepts `--with-zsh` and `--no-zsh`. The default and `--no-zsh` install the binary only, to `~/.local/bin` unless `BENTO_INSTALL_DIR` is set. `--with-zsh` appends a marked PATH export and a `bento` function to `~/.zshrc` and does not edit the prompt. Both flags together is an error.

The script prefers `https://github.com/$BENTO_REPO/releases/latest/download/bento_<os>_<arch>` (`BENTO_REPO` defaults to `StarkDroid/Bento-Terminal`). If that asset is not a Mach-O or ELF binary, it builds with Go. That fallback is the only path that needs a toolchain. Say so in the README if the asset names or the fallback change.
