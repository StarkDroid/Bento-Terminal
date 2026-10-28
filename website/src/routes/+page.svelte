<script>
	import { base } from '$app/paths';
	import { onMount } from 'svelte';
	import CopyCommand from '$lib/CopyCommand.svelte';

	const unixInstall =
		'curl -fsSL https://raw.githubusercontent.com/StarkDroid/Bento-Terminal/main/scripts/install.sh | sh';
	const windowsInstall =
		'irm https://raw.githubusercontent.com/StarkDroid/Bento-Terminal/main/scripts/install.ps1 | iex';

	// The page is prerendered with the curl line; Windows swaps to PowerShell
	// after hydration, because install.sh only handles macOS and Linux.
	let install = $state(unixInstall);
	let prompt = $state('$');

	onMount(() => {
		if (/windows/i.test(navigator.userAgent)) {
			install = windowsInstall;
			prompt = 'PS>';
		}
	});

	// The glow follows the pointer across a card, so each one lights where the
	// cursor is instead of in the middle.
	function glow(event) {
		const card = event.target instanceof Element ? event.target.closest('.feature') : null;
		if (!card) return;
		const rect = card.getBoundingClientRect();
		card.style.setProperty('--cx', `${event.clientX - rect.left}px`);
		card.style.setProperty('--cy', `${event.clientY - rect.top}px`);
	}
</script>

<svelte:head>
	<title>Bento terminal — one project, multi-pane</title>
	<meta
		name="description"
		content="Preview and edit files, review changes, run builds, and use a shell for the project in front of you — one bento layout, no tabs. Go, macOS, Linux and Windows."
	/>
	<link rel="canonical" href="https://starkdroid.github.io/Bento-Terminal/" />
	<meta property="og:type" content="website" />
	<meta property="og:site_name" content="Bento terminal" />
	<meta property="og:title" content="Bento terminal — one project, multi-pane" />
	<meta
		property="og:description"
		content="Preview, edit, review changes, run builds, and open a shell without leaving the directory."
	/>
	<meta property="og:url" content="https://starkdroid.github.io/Bento-Terminal/" />
	<meta name="twitter:card" content="summary" />
	<meta name="twitter:title" content="Bento terminal" />
	<meta name="twitter:description" content="One project. Multi-pane." />
</svelte:head>

<main id="content">
	<section class="hero" aria-labelledby="title">
		<p class="eyebrow">terminal workspace</p>
		<h1 id="title">
			One project.<br /><em>Multi-pane.</em><span class="caret" aria-hidden="true"></span>
		</h1>
		<p class="lede">
			Preview and edit files, read the tree, commits and working-tree changes, run builds, and open a
			shell — without leaving the directory you are in.
		</p>

		<div class="install">
			<CopyCommand command={install} {prompt} />
			<p class="hint">
				macOS, Linux and Windows. <a href="{base}/docs/#install">Install options</a> ·
				<a href="{base}/docs/">Docs</a>
			</p>
		</div>
	</section>

	<section aria-label="What it does">
		<h2 class="sr-only">What it does</h2>
		<dl class="features" onpointermove={glow}>
			<div class="feature">
				<dt>Preview &amp; edit</dt>
				<dd>Syntax highlighting, folding, and an editor that saves in place.</dd>
			</div>
			<div class="feature">
				<dt>Changes</dt>
				<dd>Every modified, staged, untracked, and conflicted path, with its patch.</dd>
			</div>
			<div class="feature">
				<dt>Terminal</dt>
				<dd>Any command, run through your own shell inside the project.</dd>
			</div>
			<div class="feature">
				<dt>Builds</dt>
				<dd>Detected test, build, and lint commands, one key away.</dd>
			</div>
		</dl>
	</section>
</main>
