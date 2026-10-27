<script>
	import { onMount } from 'svelte';
	import { base } from '$app/paths';
	import CopyCommand from '$lib/CopyCommand.svelte';

	const install =
		'curl -fsSL https://raw.githubusercontent.com/StarkDroid/Bento-Terminal/main/scripts/install.sh | sh';

	const sections = [
		{ id: 'install', label: 'Install' },
		{ id: 'releases', label: 'Releases' },
		{ id: 'shortcuts', label: 'Shortcuts' },
		{ id: 'layout', label: 'Layout' },
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
		content="Install Bento terminal, check the keyboard shortcuts and window layout, and see how releases are published."
	/>
	<link rel="canonical" href="https://starkdroid.github.io/Bento-Terminal/docs/" />
	<meta property="og:type" content="article" />
	<meta property="og:title" content="Docs — Bento terminal" />
	<meta
		property="og:description"
		content="Install Bento terminal, check the shortcuts and layout, and see how releases work."
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
			aria-label="Documentation sections"
			aria-orientation="vertical"
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
			<section
				class="panel"
				role="tabpanel"
				id="install"
				aria-labelledby="tab-install"
				hidden={active !== 'install'}
			>
				<h2>Install</h2>
				<p>
					One command. It downloads the release binary for your OS and architecture into
					<code>~/.local/bin</code>.
				</p>
				<CopyCommand command={install} />
				<h3>Options</h3>
				<ul>
					<li>
						<code>--with-zsh</code> adds the PATH export and a <code>bento</code> function to
						<code>~/.zshrc</code>, without touching your prompt.
					</li>
					<li><code>--no-zsh</code> is the default: the binary only.</li>
					<li><code>BENTO_INSTALL_DIR</code> changes the install directory.</li>
					<li>
						From a clone, use <code>go install ./cmd/bento</code> or <code>make install</code>.
					</li>
				</ul>
			</section>

			<section
				class="panel"
				role="tabpanel"
				id="releases"
				aria-labelledby="tab-releases"
				hidden={active !== 'releases'}
			>
				<h2>Releases</h2>
				<p>
					A release is a tagged <code>vX.Y.Z</code> GitHub release with a binary for each
					platform and a checksum file. The installer picks the right asset; only when one is
					missing does it build from source with Go.
				</p>
				<pre>bento_darwin_arm64
bento_darwin_amd64
bento_linux_arm64
bento_linux_amd64
SHA256SUMS</pre>
				<p>
					Notes are the changelog from each merged pull request, so a release page reads as a
					short list of what changed. Maintainers publish with <code>scripts/release.sh</code>;
					merging a pull request on its own publishes nothing.
				</p>
				<p>
					<a href="https://github.com/StarkDroid/Bento-Terminal/releases/latest">Latest release →</a>
				</p>
			</section>

			<section
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
								<tr><th><kbd>,</kbd></th><td>Settings and version</td></tr>
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
					<kbd>tab</kbd> still moves between boxes. While editing, printable keys type.
				</p>
			</section>

			<section
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
				<p class="muted">
					Focus starts on Files. The window needs at least 72 × 22 cells; below that it asks
					for a larger terminal.
				</p>
			</section>

			<section
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
					It targets the macOS Terminal first, with Menlo or SF Mono: rounded boxes, no Nerd Font,
					no braille spinners. Linux uses the same build. MIT licensed, contributions welcome.
				</p>
				<p>
					<a href="https://github.com/StarkDroid/Bento-Terminal">Source on GitHub →</a>
				</p>
			</section>
		</div>
	</div>
</main>
