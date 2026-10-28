<script>
	import { onMount } from 'svelte';
	import { base } from '$app/paths';
	import CopyCommand from '$lib/CopyCommand.svelte';

	const install =
		'curl -fsSL https://raw.githubusercontent.com/StarkDroid/Bento-Terminal/main/scripts/install.sh | sh';
	const installWindows =
		'irm https://raw.githubusercontent.com/StarkDroid/Bento-Terminal/main/scripts/install.ps1 | iex';

	const sections = [
		{ id: 'install', label: 'Install' },
		{ id: 'shortcuts', label: 'Shortcuts' },
		{ id: 'editing', label: 'Editing' },
		{ id: 'changes', label: 'Changes' },
		{ id: 'terminal', label: 'Terminal' },
		{ id: 'builds', label: 'Builds' },
		{ id: 'layout', label: 'Layout' },
		{ id: 'releases', label: 'Releases' },
		{ id: 'about', label: 'About' }
	];

	let active = $state('install');

	onMount(() => {
		const pick = () => {
			const id = location.hash.slice(1);
			if (sections.some((section) => section.id === id)) active = id;
		};
		pick();
		window.addEventListener('hashchange', pick);
		return () => window.removeEventListener('hashchange', pick);
	});

	function select(id) {
		active = id;
		history.replaceState(null, '', `#${id}`);
	}

	function onKeydown(event) {
		const index = sections.findIndex((section) => section.id === active);
		let next = index;
		if (event.key === 'ArrowDown' || event.key === 'ArrowRight') {
			next = (index + 1) % sections.length;
		} else if (event.key === 'ArrowUp' || event.key === 'ArrowLeft') {
			next = (index - 1 + sections.length) % sections.length;
		} else if (event.key === 'Home') {
			next = 0;
		} else if (event.key === 'End') {
			next = sections.length - 1;
		} else {
			return;
		}
		event.preventDefault();
		select(sections[next].id);
		document.getElementById(`tab-${sections[next].id}`)?.focus();
	}

	const layout = `╭──────────────────────────╮ ╭────────────────────╮
│ ≡ main.go  + new file    │ │ ≡ Files            │
│  1 │ package main        │ │ ▾ internal         │
│  2 │                     │ │   ≡ ui             │
│  3 │ func main() {       │ ├────────────────────┤
│  4 │     run()           │ │ History  Changes   │
│  5 │ }                   │ │ ↑1 ↓0 · ● 2        │
│ ↑↓←→ · e edit · ? keys   │ │ ↑↓ · ←→ tab        │
╰──────────────────────────╯ ╰────────────────────╯
╭─────────────────────────────────────────────────╮
│ ▶ Build  Terminal  auto                         │
│ $ go test ./...                                 │
│ ✓ done                                          │
│ ↑↓ pick · enter run · a · r · c · ←→ · ?        │
╰─────────────────────────────────────────────────╯
 ⎇ main  ↑1 ↓0  ● staged 0  ● unstaged 2`;
</script>

<svelte:head>
	<title>Docs — Bento terminal</title>
	<meta
		name="description"
		content="Install, shortcuts, editing, changes, the terminal, builds, layout, and releases for Bento terminal."
	/>
	<link rel="canonical" href="https://starkdroid.github.io/Bento-Terminal/docs/" />
	<meta property="og:type" content="article" />
	<meta property="og:title" content="Docs — Bento terminal" />
	<meta
		property="og:description"
		content="Every part of Bento terminal, one panel at a time: install, shortcuts, editing, changes, terminal, builds, layout, releases."
	/>
	<meta property="og:url" content="https://starkdroid.github.io/Bento-Terminal/docs/" />
	<meta name="twitter:card" content="summary" />
</svelte:head>

<main id="content">
	<div class="docs-head">
		<h1>Docs</h1>
		<p>Install it, learn the keys, and see how the boxes fit together.</p>
	</div>

	<div class="docs-grid">
		<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
		<div
			class="docs-nav"
			role="tablist"
			tabindex="-1"
			aria-label="Documentation sections"
			onkeydown={onKeydown}
		>
			{#each sections as section}
				<button
					type="button"
					role="tab"
					id="tab-{section.id}"
					aria-controls={section.id}
					aria-selected={active === section.id}
					tabindex={active === section.id ? 0 : -1}
					onclick={() => select(section.id)}
				>
					{section.label}
				</button>
			{/each}
		</div>

		<div>
			<div
				class="panel"
				role="tabpanel"
				id="install"
				aria-labelledby="tab-install"
				hidden={active !== 'install'}
			>
				<h2>Install</h2>
				<p>
					Install is the published binary; development is a build from a checkout. The features are
					the same in both, on every platform: only the shell the Terminal tab runs and the
					installer differ.
				</p>

				<h3>macOS and Linux</h3>
				<p>
					One command. It downloads the release binary for your OS and architecture into
					<code>~/.local/bin</code>. Add <code>--with-zsh</code> to put that folder on PATH.
				</p>
				<CopyCommand command={install} />

				<h3>Windows</h3>
				<p>
					Run this in Windows Terminal. It downloads <code>bento.exe</code> into
					<code>%LOCALAPPDATA%\bento</code>, adds that folder to your user PATH, and checks that
					the binary runs. Open a new Windows Terminal and type <code>bento</code>.
				</p>
				<CopyCommand command={installWindows} prompt="PS>" />

				<h3>Develop</h3>
				<p>macOS and Linux:</p>
				<pre>go test ./...
make build
./bento
sh scripts/install.sh --from-source</pre>
				<p>Windows, in Windows Terminal:</p>
				<pre>go test ./...
go build -o bento.exe ./cmd/bento
.\bento.exe
.\scripts\install.ps1 -FromSource</pre>
				<p>
					<code>go run ./cmd/bento</code> is the quickest loop anywhere: it compiles to a temp
					directory, so PATH and the checkout stay untouched. <code>--from-source</code> and
					<code>-FromSource</code> install the checkout itself, which is how a branch is tried
					before a release exists. <code>go install ./cmd/bento</code> also works, but writes to
					<code>go env GOPATH</code>'s <code>bin</code>, which may not be on PATH.
				</p>

				<h3>Options</h3>
				<ul>
					<li>
						<code>--with-zsh</code> adds the PATH export and a <code>bento</code> function to
						<code>~/.zshrc</code>, without touching your prompt.
					</li>
					<li><code>--no-zsh</code> is the default: the binary only.</li>
					<li>
						<code>--from-source</code> on macOS and Linux, <code>-FromSource</code> on Windows:
						build the checkout instead of downloading a release.
					</li>
					<li>
						<code>BENTO_INSTALL_DIR</code> changes the install directory; on Windows, so does
						<code>-InstallDir</code>.
					</li>
					<li>
						<code>-NoPath</code> installs <code>bento.exe</code> without editing your user PATH.
					</li>
					<li>
						<code>BENTO_REPO</code> points the installer at a fork;
						<code>BENTO_REPO_URL</code> changes the clone URL the source build uses.
					</li>
				</ul>
			</div>

			<div
				class="panel"
				role="tabpanel"
				id="shortcuts"
				aria-labelledby="tab-shortcuts"
				hidden={active !== 'shortcuts'}
			>
				<h2>Shortcuts</h2>
				<div class="key-grid">
					<div>
						<h3>Anywhere</h3>
						<table>
							<tbody>
								<tr><th><kbd>tab</kbd> <kbd>shift+tab</kbd></th><td>Next or previous box</td></tr>
								<tr><th><kbd>?</kbd></th><td>Commands for this box</td></tr>
								<tr><th><kbd>,</kbd></th><td>Settings, version, and links</td></tr>
								<tr><th><kbd>q</kbd></th><td>Quit</td></tr>
							</tbody>
						</table>
						<h3>Main</h3>
						<table>
							<tbody>
								<tr><th><kbd>↑</kbd> <kbd>↓</kbd> <kbd>←</kbd> <kbd>→</kbd></th><td>Move the caret</td></tr>
								<tr><th><kbd>pgup</kbd> <kbd>pgdown</kbd></th><td>Page the preview</td></tr>
								<tr><th><kbd>[</kbd> <kbd>]</kbd></th><td>Fold or unfold a block</td></tr>
								<tr><th><kbd>e</kbd></th><td>Edit; <kbd>ctrl+s</kbd> saves</td></tr>
								<tr><th><kbd>n</kbd> or <kbd>+</kbd></th><td>New file</td></tr>
								<tr><th><kbd>s</kbd></th><td>Highlighting on or off</td></tr>
							</tbody>
						</table>
					</div>
					<div>
						<h3>Files</h3>
						<table>
							<tbody>
								<tr><th><kbd>enter</kbd></th><td>Open a file or folder</td></tr>
							</tbody>
						</table>
						<h3>History</h3>
						<table>
							<tbody>
								<tr><th><kbd>←</kbd> <kbd>→</kbd></th><td>History or Changes</td></tr>
								<tr><th><kbd>enter</kbd></th><td>Show the commit or patch</td></tr>
							</tbody>
						</table>
						<h3>Build</h3>
						<table>
							<tbody>
								<tr><th><kbd>enter</kbd></th><td>Run the selected command</td></tr>
								<tr><th><kbd>a</kbd></th><td>Automatic or manual list</td></tr>
								<tr><th><kbd>r</kbd> <kbd>c</kbd></th><td>Rerun or cancel</td></tr>
								<tr><th><kbd>←</kbd> <kbd>→</kbd></th><td>Build or Terminal</td></tr>
							</tbody>
						</table>
						<h3>Terminal</h3>
						<table>
							<tbody>
								<tr><th><kbd>enter</kbd></th><td>Run the line</td></tr>
								<tr><th><kbd>ctrl+c</kbd></th><td>Stop the command</td></tr>
								<tr><th><kbd>↑</kbd> <kbd>↓</kbd></th><td>Earlier commands</td></tr>
								<tr><th><kbd>esc</kbd></th><td>Clear, then back to Build</td></tr>
							</tbody>
						</table>
					</div>
				</div>
				<p class="muted">
					In the Terminal tab every printable key goes to the prompt, including <kbd>q</kbd>;
					<kbd>tab</kbd> still moves between boxes. While editing, printable keys type. The about
					overlay from <kbd>,</kbd> also carries chips for the project on GitHub, the command
					list, quit, and close; a click runs one.
				</p>
			</div>

			<div
				class="panel"
				role="tabpanel"
				id="editing"
				aria-labelledby="tab-editing"
				hidden={active !== 'editing'}
			>
				<h2>Editing</h2>
				<p>
					<kbd>e</kbd> edits the open text file and <kbd>ctrl+s</kbd> saves it. Arrows move the
					caret, <kbd>enter</kbd> splits a line, <kbd>backspace</kbd> and <kbd>delete</kbd> remove
					and join lines, <kbd>home</kbd> and <kbd>end</kbd> jump within a line, and
					<kbd>tab</kbd> inserts four spaces. Printable keys type.
				</p>
				<p>
					The file's tab shows <code>●</code> while the buffer differs from the file,
					<code>✓</code> after a save, and <code>save failed</code> when the write fails.
					<kbd>esc</kbd> leaves the editor and keeps unsaved text in the preview, where
					<kbd>ctrl+s</kbd> still saves it.
				</p>
				<p>
					The <code>+ new file</code> tab, or <kbd>n</kbd>, opens a path prompt and puts the new
					file straight into the editor; parent directories are created. Every write stays
					inside the project. A binary file, a commit, and a patch are read-only.
				</p>
			</div>

			<div
				class="panel"
				role="tabpanel"
				id="changes"
				aria-labelledby="tab-changes"
				hidden={active !== 'changes'}
			>
				<h2>Changes</h2>
				<p>
					History's <strong>Changes</strong> tab lists every path git reports, in order, with its
					two porcelain columns. <kbd>enter</kbd> or a click opens the patch in Main, and the tab
					title carries the number of changed paths.
				</p>
				<pre>M   staged             M   unstaged
MM  staged and unstaged
A   added              ??  untracked
R   renamed, old → new
UU  unmerged</pre>
				<p>
					Staged and unstaged changes are shown together, and an untracked file gets a generated
					new-file patch. A clean tree says so.
				</p>
			</div>

			<div
				class="panel"
				role="tabpanel"
				id="terminal"
				aria-labelledby="tab-terminal"
				hidden={active !== 'terminal'}
			>
				<h2>Terminal</h2>
				<p>
					Build's <strong>Terminal</strong> tab runs a typed line through your shell inside the
					project root: <code>$SHELL -c</code>, falling back to <code>/bin/sh</code>, on macOS and
					Linux; <code>%COMSPEC% /C</code>, falling back to <code>cmd.exe</code>, on Windows.
					<kbd>enter</kbd> runs, <kbd>↑</kbd> <kbd>↓</kbd> recall earlier lines, <kbd>esc</kbd>
					clears the prompt and then returns to Build, and <kbd>ctrl+c</kbd> stops a running command
					or clears the prompt.
				</p>
				<p>
					Output is captured as text: colour, cursor, and erase sequences are dropped so a command
					like <code>clear</code> cannot touch Bento's own screen. The scrollback is separate from
					the Build output, and starting a command stops the previous one.
				</p>
				<p>
					Full-screen programs such as <code>vim</code> or <code>top</code> are not supported; the
					tab is for one-shot commands. Pipes, globs, and builtins work because the line goes
					through your shell.
				</p>
			</div>

			<div
				class="panel"
				role="tabpanel"
				id="builds"
				aria-labelledby="tab-builds"
				hidden={active !== 'builds'}
			>
				<h2>Builds</h2>
				<p>Automatic mode reads the project and offers:</p>
				<ul>
					<li><code>package.json</code> scripts, <code>Makefile</code> targets, and <code>gradlew</code> tasks</li>
					<li><code>go.mod</code>: test, build, and vet; <code>Cargo.toml</code>: test, build, check, and run</li>
					<li><code>pyproject.toml</code>: <code>python -m pytest</code>, and a build when one is declared</li>
				</ul>
				<p>
					<kbd>a</kbd> switches to a shorter manual list, <kbd>enter</kbd> runs the selected
					command, <kbd>r</kbd> reruns it, and <kbd>c</kbd> cancels the process group. The output
					streams under the list.
				</p>
				<p>
					Detected commands run directly, without a shell, so pipes, globs, and builtins belong in
					the Terminal tab.
				</p>
			</div>

			<div
				class="panel"
				role="tabpanel"
				id="layout"
				aria-labelledby="tab-layout"
				hidden={active !== 'layout'}
			>
				<h2>Layout</h2>
				<pre class="diagram">{layout}</pre>
				<ul>
					<li><strong>Main:</strong> the open file, commit, or patch, with the source caret.</li>
					<li><strong>Files:</strong> the project tree, with hidden and vendor folders skipped.</li>
					<li><strong>History:</strong> commits on the branch and the working-tree changes, as tabs.</li>
					<li><strong>Build:</strong> detected commands and the terminal, as tabs.</li>
				</ul>
				<p>
					Click a box or a tab label to focus and switch; a click in Main puts the caret on that
					line and toggles a fold marker. The wheel scrolls the box under the pointer. Focus starts
					on Files, <kbd>tab</kbd> cycles the boxes, and the window needs at least 72 × 22 cells.
				</p>
				<p class="muted">
					Missing git, an empty tree, no detected commands, and a clean working tree each say so
					in place.
				</p>
			</div>

			<div
				class="panel"
				role="tabpanel"
				id="releases"
				aria-labelledby="tab-releases"
				hidden={active !== 'releases'}
			>
				<h2>Releases</h2>
				<p>
					A release is a tagged <code>vX.Y.Z</code> GitHub release with a binary for each
					platform and a checksum file. The installer picks the right asset and checks it before
					replacing your copy; only when one is missing, or when you ask for a source build, does
					it build with Go.
				</p>
				<pre>bento_darwin_arm64
bento_darwin_amd64
bento_linux_arm64
bento_linux_amd64
bento_windows_amd64.exe
bento_windows_arm64.exe
SHA256SUMS</pre>
				<p>
					Notes are the changelog from each merged pull request, so a release page reads as a
					short list of what changed. Maintainers publish with <code>scripts/release.sh</code>;
					merging a pull request on its own publishes nothing.
				</p>
				<p>
					<a href="https://github.com/StarkDroid/Bento-Terminal/releases/latest">Latest release →</a>
				</p>
			</div>

			<div
				class="panel"
				role="tabpanel"
				id="about"
				aria-labelledby="tab-about"
				hidden={active !== 'about'}
			>
				<h2>About</h2>
				<p>
					Bento is a single Go binary built on <a
						href="https://github.com/charmbracelet/bubbletea">Bubble Tea</a
					>, <a href="https://github.com/charmbracelet/lipgloss">Lip Gloss</a>, and <a
						href="https://github.com/alecthomas/chroma">Chroma</a
					>. It draws its own lagoon palette, so your terminal theme does not change how it looks.
				</p>
				<p>
					It runs in the macOS Terminal and Windows Terminal, with Menlo or SF Mono on macOS and
					Cascadia Mono or Consolas on Windows: rounded boxes, no Nerd Font, no braille spinners.
					Linux uses the same build. Every platform gets the same panes, keys, and features; only
					the shell the Terminal tab runs and the installer differ. MIT licensed, contributions
					welcome.
				</p>
				<p>
					<a href="https://github.com/StarkDroid/Bento-Terminal">Source on GitHub →</a>
				</p>
			</div>
		</div>
	</div>
</main>
