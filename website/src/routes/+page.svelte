<script>
	import { base } from '$app/paths';
	import CopyCommand from '$lib/CopyCommand.svelte';

	const install =
		'curl -fsSL https://raw.githubusercontent.com/StarkDroid/Bento-Terminal/main/scripts/install.sh | sh';
</script>

<svelte:head>
	<title>Bento terminal — a terminal workspace for one project</title>
	<meta
		name="description"
		content="A full-screen terminal workspace for one project: preview and edit files, review changes, run builds, and use a shell in one bento layout. Go, macOS and Linux."
	/>
	<link rel="canonical" href="https://starkdroid.github.io/Bento-Terminal/" />
	<meta property="og:type" content="website" />
	<meta property="og:site_name" content="Bento terminal" />
	<meta property="og:title" content="Bento terminal — a terminal workspace for one project" />
	<meta
		property="og:description"
		content="Preview and edit files, review changes, run builds, and use a shell in one bento layout."
	/>
	<meta property="og:url" content="https://starkdroid.github.io/Bento-Terminal/" />
	<meta name="twitter:card" content="summary" />
	<meta name="twitter:title" content="Bento terminal" />
	<meta name="twitter:description" content="A terminal workspace for one project." />
</svelte:head>

<main id="content">
	<section class="hero" aria-labelledby="title">
		<p class="eyebrow">terminal workspace</p>
		<h1 id="title">
			A terminal workspace for one project.<span class="caret" aria-hidden="true"></span>
		</h1>
		<p class="lede">
			Preview and edit files, read the tree, commits and working-tree changes, run builds, and open a
			shell — in the directory you are already in.
		</p>

		<div class="install">
			<CopyCommand command={install} />
			<p class="hint">
				Binary only. <a href="{base}/docs/#install">Install options</a> ·
				<a href="{base}/docs/">Docs</a>
			</p>
		</div>
	</section>

	<section aria-label="What it does">
		<h2 class="sr-only">What it does</h2>
		<dl class="features">
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
