# Bento terminal

Bento terminal is a full-screen terminal workspace for one project directory. It previews and edits files with syntax highlighting and folding, lists the tree, the current branch's commits, and its working-tree changes, runs detected build commands, and runs your own commands in a shell. The supported terminals are macOS Terminal and Windows Terminal. The UI draws its own lagoon palette and does not read the user's theme.

Every box draws its own shortcut footer. `?` opens that box's command list, and a listed command can be run with enter or by pressing its shortcut. When you change packages, commands, focus order, theme tokens, keybindings, or the install flow, update this file in the same change.

## Package map

- `cmd/bento` — process entry. Uses the working directory, or one optional directory argument, as the project root. Starts Bubble Tea with the alternate screen and mouse cell motion.
- `internal/ui` — Bubble Tea model, bento layout, theme, focus, keys, and rendering. Scroll position for the preview, the build output, and the terminal is a Bubbles viewport. The running indicator is a Bubbles spinner. Bindings are Bubbles keys.
  - `commands.go` — one action table per box. Actions carry their binding and a pane-specific description, and `runAction` is the single dispatcher. This table drives both key handling and the `?` command list.
  - `palette.go` — the per-box command list: open/close, cursor movement, enter to run, direct shortcut to run.
  - `settings.go` — the settings overlay: wordmark, version, about, and the `github`/`commands`/`quit`/`close` chips.
  - `browser.go` — hands the project URL to the default browser.
  - `codeview.go` — renders the file preview: gutter, fold markers, caret row and block, highlighted spans, and collapsed-block summaries.
  - `preview.go` — line numbering, wrapping, and diff colouring helpers for plain previews.
  - `paint.go` — piece strings, left/right joins, tab chips, titles, rows, and panel frames.
  - `output.go` — sanitises captured process output before it reaches a viewport.
  - `editor.go` — the preview editor: open file, buffer edits, save, and the new-file prompt.
  - `terminal.go` — the Terminal tab: prompt input, history, run dispatch, and terminal scrollback.
  - `view.go` — pane bodies, the History and Build tab titles, footers, and overlays.
- `internal/syntax` — Chroma-based language detection, highlighting with the lagoon palette, and foldable block discovery. Go and Python use dedicated strategies; the common brace languages share a balanced-brace scanner. Highlighting is capped at 8,000 lines.
- `internal/git` — parses `git status --porcelain=v2 --branch` and `git log` lines (`hash`, unix time, subject, tab-separated). `Load` shells out to git. Untracked paths count as unstaged. `Status.Files` keeps one `FileStatus` per changed path, and `Diff` returns the staged and unstaged patch, with a synthesised new-file patch for untracked files.
- `internal/files` — lazy directory tree and a bounded text preview. `Create` and `Save` resolve project-relative paths and reject anything outside the project; `Tree.Refresh` re-reads loaded folders while keeping them open. Dot entries are hidden. `node_modules`, `vendor`, `target`, `__pycache__`, `dist`, `.next`, `.venv`, and `venv` are skipped. Symlinks are not followed.
- `internal/build` — `Inspect` reads project files and returns automatic and manual command lists. `Start` runs a command and cancels it with the process group (`run_unix.go`: `Setpgid` and `SIGTERM`; `run_windows.go`: `CREATE_NEW_PROCESS_GROUP` and `taskkill /T`). No shell is inserted unless the command itself is a shell; `ShellCommand` wraps a typed line for the Terminal tab through `$SHELL -c` on Unix and `%COMSPEC% /C` on Windows.
- `scripts/install.sh` — macOS and Linux entrypoint. Release asset first, then `go build` from a local checkout or a clone; `--from-source` skips the download.
- `scripts/install.ps1` — Windows entrypoint. The same order: the `.exe` release asset for the machine's architecture, then a build from the checkout this file lives in or a clone (`-FromSource`, `-Source`). Installs `bento.exe` into `%LOCALAPPDATA%\bento` and adds that folder to the user PATH (`-NoPath` skips it), so a new Windows Terminal finds `bento`.
- `scripts/release.sh` — the release. Tests, builds the six assets, writes `SHA256SUMS`, tags, and publishes the GitHub release from the merged PRs' changelogs.
- `website` — the GitHub Pages site: SvelteKit prerendered to static HTML. It is a separate Node project and the Go build never touches it.

There is no `internal/install`. The installers are shell and PowerShell only.

## Platforms

macOS Terminal and Windows Terminal are the supported terminals; Linux runs the same binary as macOS. Every user-facing feature works on every platform: panes, focus, keys, mouse, editing, diffs, build detection, and the Terminal tab are not allowed to be macOS-only or Windows-only, and the docs and README describe one feature set for all platforms.

The only platform differences are the ones the OS forces:

- The Terminal tab runs the user's shell: `$SHELL -c`, falling back to `/bin/sh`, on Unix; `%COMSPEC% /C`, falling back to `cmd.exe`, on Windows. Keep that behind `run_unix.go` and `run_windows.go` so the rest of `internal/build` stays shared.
- Cancelling a run ends the process group: `Setpgid` and `SIGTERM` on Unix; `CREATE_NEW_PROCESS_GROUP` and `taskkill /T` on Windows.
- The installer: `scripts/install.sh` on macOS and Linux, `scripts/install.ps1` on Windows. The two must stay in step — a flag or fallback on one needs its equivalent on the other.

Test a platform change on the platform whose file you touched, and leave the other to the CI matrix; the shared code still runs through `go vet ./...` on every platform. When you add a feature, add it to `internal/ui` once rather than branching on `runtime.GOOS`.

## Build and test

```sh
go test ./...
make build
make install
make darwin-arm64
make darwin-amd64
make linux-arm64
make linux-amd64
make windows-amd64
make windows-arm64
```

Windows, in Windows Terminal:

```powershell
go test ./...
go build -o bento.exe ./cmd/bento
.\bento.exe
.\scripts\install.ps1 -FromSource   # install the build, then `bento` works
```

`go run ./cmd/bento [dir]` is the no-install loop on every platform: it compiles to a temp directory and leaves PATH and the repository alone. `--from-source` and `-FromSource` install the checkout itself, so a branch can be tried before a release exists. `make` is not assumed on Windows; the `go build -o dist\bento_windows_amd64.exe ./cmd/bento` form produces the same release asset.

Go 1.23 or newer. Dependencies are Go modules. The platform targets set `GOOS` and `GOARCH` and write `dist/bento_<os>_<arch>`, plus `.exe` on Windows; those are the installers' release asset names.

Tests cover command detection fixtures in `internal/build/testdata`, git status and log fixtures in `internal/git/testdata`, per-file status and diffs, create/save and tree refresh, language detection and highlighting, Go/Python/brace folding, caret navigation, settings, focus cycling, layout size, empty views, file preview, editor keys and save, the new-file prompt, the Changes tab and diff preview, the terminal tab and a real shell run, captured output sanitising, tab clicks, per-box footers and command lists, keyboard and mouse folding, and a real `echo` build. Process tests live in `run_unix_test.go` and `run_windows_test.go`; the two output tests that use `printf` and `clear` skip on Windows. A git binary is required for `internal/git` load tests; those tests skip when it is missing.

CI runs gofmt, `go vet`, `go test`, and a build on Ubuntu, macOS, and Windows. The Windows job builds `bento.exe` and uploads it as the `bento-windows-amd64` artifact, so a platform change can be tested before a release.

`scripts/release.sh` is the only thing that publishes a release; nothing runs on merge. CI (`.github/workflows/ci.yml`) only checks gofmt, `go vet`, `go test`, and a build on Ubuntu, macOS, and Windows.

## Website

`website/` is the project site: a SvelteKit app prerendered to static HTML and deployed to GitHub Pages by `.github/workflows/pages.yml`. It is a separate Node project, so the Go build and tests never touch it.

```sh
cd website
npm ci
npm run dev      # local preview
npm run build    # writes website/build
```

`BASE_PATH` sets the path prefix; the workflow passes the value from `actions/configure-pages`, so the same build works on the project page or a custom domain. The pages are `/` and `/docs/`. The docs page keeps all nine sections in the prerendered HTML and switches panels with tabs, so it stays short and stays crawlable: install, shortcuts, editing, changes, terminal, builds, layout, releases, about. Copy lives in the Svelte components, not Markdown: keep the landing page to one screen and do not duplicate whole README sections.

The landing page reacts to a fine pointer: `+layout.svelte` moves the fixed `.spotlight`, and a handler on the feature grid sets `--cx`/`--cy` so each card lights under the cursor. Both use CSS-variable fallbacks, are set from `onMount` so prerendering stays static, and are switched off for coarse pointers and reduced motion.

The README banner is `assets/banner.png`, a 2560 × 640 render (2× of 1280 × 320) of `assets/banner.html` made with headless Chrome. It uses the site's mark, palette, and mono font; update the HTML and re-render when the mark, name, or tagline changes.

Site changes go through a pull request and never cut a release. The Pages workflow builds the site on any pull request that touches `website/` and only deploys from `main`.

## Workflow

Never commit a feature straight to `main`, and never merge a pull request yourself. Make a branch, commit there in logical chunks, push it, and open a pull request — for every change, however small — then leave the merge to the user. Only a pull request lands on `main`, and `main` stays green because CI runs on the PR.

Never publish a release while delivering a feature. Releases are a separate step the user starts after merging; see Releases.

Every pull request body carries a user-facing `## Changelog`, and that changelog is kept current as the branch grows: when you add or change something, update the PR body in the same pass. It becomes the release notes. Add a `Release: <version>` line only when the PR is the release PR. Keep the author yours; do not add a tool or assistant as a co-author.

Every commit is GPG-signed: `commit.gpgsign` is on and the key must be the one uploaded to the maintainer's GitHub account, or GitHub will not mark the commit verified. Never pass `--no-gpg-sign`, never wrap `gpg.program`, and never rewrite a commit unsigned. If signing hangs, run `gpgconf --launch gpg-agent` and retry; confirm with `git log -1 --show-signature` (`%G?` is `G`) before finishing.

## Releases

`scripts/release.sh` (or `make release V=<version>`) publishes; merging never does. The user merges the release pull request, then runs the script on `main`:

1. It takes the version from the merged pull request's `Release:` line, or from an argument, and stops if the tag or release already exists.
2. It collects the `## Changelog` bullets of every pull request merged since the last release — falling back to a cleaned-up title — and writes the notes with a compare link.
3. It runs `go test ./...`, builds `bento_<os>_<arch>` for macOS, Linux, and Windows (`bento_windows_<arch>.exe`) with `main.version` set, and writes `SHA256SUMS`.
4. It tags `v<version>`, pushes the tag, and creates the GitHub release with the notes and assets.

`--dry-run` prints the version and notes and changes nothing; `--since <tag>` sets the base tag. The script does not commit, so the release pull request must bump the version in `cmd/bento/main.go` and `Makefile`; the script warns when the source and the release version differ.

## UI conventions

The minimum window is **72 × 22** cells; smaller windows show a single "make the terminal larger" panel. macOS Terminal and Windows Terminal 1.x or newer are the targets; the legacy Windows console (conhost) is not.

Focus order is Main, Files, History, Build. The app starts on Files. Tab moves Files → History → Build → Main. Shift-Tab reverses. Changing focus re-renders the preview so the cursor row only highlights while Main is focused.

The focused pane uses border `#3DDECE` and a bold title. Unfocused panes use border `#24555C`. A selected row uses fill `#1B5E5A` when its pane is focused and `#14343C` when it is not; tabs reuse those two fills, so the lighter one always means "this one". Borders are not the selection.

### Footers and command lists

Each box reserves its last body row for a footer of its own shortcuts. Footer hints are:

| Box | Hints |
| --- | --- |
| Main, editing | `↑↓←→ move · ctrl+s save · esc done` |
| Main, highlighted file | `↑↓←→ move · [ ] fold · e edit · ? keys` |
| Main, file with no detected language | `↑↓←→ move · e edit · ? keys` |
| Main, commit or patch | `↑↓ scroll · ? keys` |
| Main, home | `n new · ? keys · tab panes` |
| Files | `↑↓ · enter open · ? keys` |
| History | `↑↓ · ←→ tab · ? keys` |
| Build | `↑↓ pick · enter run · a mode · r rerun · ←→ tab · ? keys` |
| Terminal | `enter run · ctrl+c stop · esc back · ? keys` |

`?` opens the command list for the focused box (`? Keys · Files`). The list shows that box's commands and then a Global section. `↑` `↓` (or `k` `j`) move, `pgup`/`pgdown` jump, `enter` runs the selected row, and pressing a listed shortcut runs it directly and closes the list. `?` and `esc` close it. While the list is open, focus does not move, and mouse events are ignored.

Build commands (`a`, `r`, `c`) are listed under Build but work from any box. `q` quits even while the list is open.

`,` opens the settings overlay from any box. It draws the BENTO wordmark, the version passed to `ui.New` by `cmd/bento`, a short about, and a row of chips: `github` opens the repository in the default browser, `? commands` opens the command list, `q quit` quits, and `esc close` closes. `esc` or `,` also closes it. A left click on a chip runs it; other mouse events are ignored while it is open.

To add a command: add the binding to `keyMap`, add an action to the right `paneActions` case (and `buildActions` if it must work globally), and handle its id in `runAction`. Update the footer hints if the command is important enough to show there.

### Tabs

Tabs are chips drawn by `chip` and `tabColors` in `paint.go`: one cell of fill on each side of the label, the active tab on `#1B5E5A` with bold lagoon-bright text while its box is focused (sand when not), and inactive tabs on `#14343C` with sand-dim text. The chip is what separates a tab from the title bar; do not paint tabs on the panel colour.

History draws its title as `⎇  History   Changes` and Build as `▶  Build   Terminal`. `←` `→` (and `h` `l`) switch tabs while that box is focused, and a click on a tab label switches too. Click zones are in inner-title cells: History has a 3-cell `⎇` prefix, then a 9-cell ` History `, a 1-cell gap, and a 9-cell ` Changes `; Build has a 3-cell state glyph, then a 7-cell ` Build `, a 1-cell gap, and a 10-cell ` Terminal `. Keep `historyTabAt` and `buildTabAt` in `update.go` in step with those widths.

Main is a two-tab strip in the same style: the open file's tab (`fileTab`, always the active chip, with the language badge and save marker inside it) and the `+ new file` tab (`newFileTab`, inactive) directly after it with no gap, the way a browser keeps its new-tab button beside the last tab. The new-file tab is hidden while editing. Its click zone moves with the file name, so `newFileTabAt` measures `fileTab().width` and `newFileTab(false).width` instead of hard-coding cells; a long name that pushes the tab past the right edge leaves the clipped part dead.

Each tab keeps its own cursor and scroll: commits and changes, commands and terminal output. Switching tabs does not run anything.

### Code preview

The preview is capped at 256 KB. The title is the file's tab: ` ≡ name`, then `  ·  Language` when a language is detected, then the save marker and ` edit`. The `+ new file` tab follows it. `s` toggles highlighting and the tab then says `Language off`; folding still works with highlighting off.

The Main pane has a source caret. `↑` `↓` move it by line, `←` `→` move it by character (wrapping at line ends), `pgup`/`pgdown` page it, and the viewport follows. `[` folds the innermost function or class at the caret, `]` unfolds it, and `enter` toggles it. A folded block keeps its opening line with a `▸` marker and a `… N lines` summary; open blocks use `▾`. The caret row uses fill `#14343C`, a `▌` marker when the line is not a fold start, and a lagoon-bright block over the character at the caret.

Commit and patch previews have no gutter, cursor, highlighting, or folding; they scroll as plain text. Patches are coloured: `+` palm, `-` coral, `@@` reef, and file headers moss.

### Editing

`e` edits an open text file. The buffer lives in `previewEditLines`, not `previewRaw`, because `fileLines` trims trailing newlines and the editor must be able to append a final line; `sourceLines()` returns the buffer while editing and the preview text otherwise. Folding blocks are dropped while editing and rebuilt on exit.

Printable keys type, `enter` splits, `backspace`/`delete` delete and join lines, `home`/`end` jump within the line, `tab` inserts four spaces, and arrows/page keys move the caret. `ctrl+s` saves, `esc` leaves the editor and keeps the buffer in the preview. The title shows `●` for unsaved changes, `✓` after a save, and `save failed` when the write fails. Highlighting is refreshed after every edit and save.

Saving joins the buffer with newlines and adds a trailing newline when the buffer is not empty. Writes go through `files.Save`, which keeps the path inside the project. A successful save refreshes the tree in place and reloads git.

`n` or `+` opens the new-file prompt over the grid: type a project-relative path, `enter` creates it (parent directories included), and `esc` cancels. A path that already exists or climbs out of the project keeps the prompt open with an error. A created file opens in the editor, and its name decides the language, so saving a `.go` or `.py` file highlights it.

### Changes tab

`git.Status.Files` lists every changed path in git's order. The Changes tab shows the porcelain columns (`M `, ` M`, `MM`, `A `, `??`, `R `, and unmerged codes) and renders renames as `old → new`. The title carries the change count.

`enter` or a click opens the selected path in Main as a `±` patch. `git.Diff` joins the staged and unstaged patches and synthesises a new-file patch for an untracked file. A clean tree shows "Working tree clean".

### Terminal tab

The terminal prompt always has the keyboard while the tab is visible: printable keys type (including `q`, `?`, and `,`), `tab`/`shift+tab` still cycle focus, `↑`/`↓` walk the command history, `←`/`→` move the prompt caret, and `home`/`end` jump. `enter` runs the line through `build.ShellCommand`, which uses `$SHELL -c` and falls back to `/bin/sh`. `esc` clears the prompt, then returns to the Build tab. `ctrl+c` stops a running process, or clears the prompt when nothing runs. The command list and settings open on `?` and `,` only while the prompt is empty.

Terminal output has its own viewport and 2,000-line scrollback, separate from the Build output. `runTarget` selects which buffer receives lines and which state records the result; starting any run stops the previous one. The terminal output height is `Build.H - 5` (panel body minus the prompt and footer).

Captured output is sanitised by `sanitizeLine` in `output.go` before it reaches either viewport: escape sequences are dropped, a carriage return keeps only its last segment, tabs become four spaces, and a line that held nothing but control sequences is skipped. A command still thinks it owns a terminal, so `clear` would otherwise send `\x1b[2J` into the scrollback and wipe the whole interface. Never append raw process output to a viewport, and do not try to keep colour codes: they would fight the pane background. Full-screen programs are not supported; the tab is for one-shot commands.

### Mouse

Mouse reporting is on (`tea.WithMouseCellMotion`). A left click focuses the box under the pointer and moves its selection; clicking a directory toggles it, clicking a file opens the preview, clicking a commit or a changed path previews it, and clicking a tab label switches tabs. In Main's title, clicking the `+ new file` tab opens the new-file prompt (but not while editing). In Main, a click puts the caret at the start of that source line and, when the line opens a foldable block, toggles that block. The wheel scrolls the box under the pointer by three rows (Build scrolls the command output or the terminal scrollback). In the settings overlay, a click on a chip runs that chip; every other click is ignored.

`paneAt` in `layout.go` maps screen cells to boxes; body-top helpers keep click rows aligned with the rendered content.

Theme tokens:

| Token | Hex | Use |
| --- | --- | --- |
| ocean | `#071E26` | screen and gaps |
| panel | `#0C2832` | pane fill |
| status | `#04161C` | status bar |
| sand | `#F4E4C4` | text |
| sand dim | `#B7A88A` | hints, inactive tabs, unfocused titles |
| lagoon | `#1F8F86` | mode label, secondary accent, focused footer |
| lagoon bright | `#3DDECE` | focus border, active tab text, keywords |
| coral | `#FF6F61` | dirty counts, failures, tags |
| palm | `#7DCE82` | success, detected systems, function and class names |
| sun | `#E9C46A` | strings, decorators, collapsed fold marker |
| amber | `#F4A261` | numbers |
| moss | `#5F7D7B` | comments, gutter |
| reef | `#5FB3C4` | types and builtins |
| select | `#1B5E5A` | active tab chip, selected row in the focused pane |
| select dim | `#14343C` | inactive tab chip, selected row in an unfocused pane |

Every pane sets foreground and background. Do not rely on the terminal's default colors. The syntax palette lives in `internal/syntax/highlight.go` and mirrors these tokens. Stay with Unicode that Menlo and SF Mono (macOS) and Cascadia Mono or Consolas (Windows) provide: `⎇ ↑ ↓ ● ▶ ✓ ✗ ▸ ▾ ≡ ± ▌ + · …` and rounded box drawing. Do not require a Nerd Font, do not use braille spinner frames, and do not use `⏎` in hints (write `enter`).

Build mode defaults to automatic. `a` switches to manual. `c` cancels the command and its children. `r` reruns the last argv.

### Empty states

- Not a git repository: the history box and the status bar say so. The home screen omits a branch name.
- Empty tree: the files box says the directory is empty.
- Clean working tree: the Changes tab says the working tree is clean.
- No detected commands: the build list says none were found. If the other mode has commands, it says to press `a`.
- Build idle: the output area says to select a command and press enter.
- Terminal idle: the prompt area says to type a command and press enter.
- No language detected in a file: no language badge, no fold markers, and `s` does nothing.

## Install

Installation and development are separate per platform. Installation is the published binary; development is a build from a checkout.

`scripts/install.sh` is the macOS and Linux installer. It accepts `--with-zsh`, `--no-zsh`, and `--from-source`. The default and `--no-zsh` install the binary only, to `~/.local/bin` unless `BENTO_INSTALL_DIR` is set, and print the path to run. `--with-zsh` appends a marked PATH export and a `bento` function to `~/.zshrc` and does not edit the prompt. Both zsh flags together is an error. `--from-source` builds the checkout the script lives in instead of downloading the release, then falls back to a clone.

`scripts/install.ps1` is the Windows path. It takes `-InstallDir`, `-FromSource`, `-Source`, and `-NoPath`. It downloads `bento_windows_<arch>.exe` for the machine (arm64 or amd64) into `%LOCALAPPDATA%\bento` unless `-InstallDir` or `BENTO_INSTALL_DIR` is set, checks that the file runs, then appends that folder to the user PATH as a registry write that keeps `%VAR%` entries unexpanded and broadcasts the environment change, unless `-NoPath`. `-FromSource` or `-Source` builds the checkout instead; that is the flag a branch dev uses, because the release does not contain the branch. `BENTO_REPO` overrides the release repository and `BENTO_REPO_URL` the clone URL. Open a new Windows Terminal afterwards; the binary is unsigned, so SmartScreen may ask once.

Both scripts prefer `https://github.com/$BENTO_REPO/releases/latest/download/` and fall back to `go build` from a local checkout or a shallow clone. They validate the download before replacing the installed binary (Mach-O or ELF magic in `install.sh`, `MZ` in `install.ps1`). That fallback is the only path that needs a toolchain (Go 1.23+). Keep the two in step: a new flag, directory rule, or fallback on one needs its equivalent on the other, and the README and docs page must state the per-platform install and develop commands. Say so in the README if the asset names or the fallback change.

For a branch without installing anything, use `go run ./cmd/bento [dir]`, `go build -o bento.exe ./cmd/bento`, or `make windows-amd64`. A Windows dev loop lives in Build and test.
