<svelte:head>
	<title>Privacy — Bento terminal</title>
	<meta
		name="description"
		content="Bento terminal has no app telemetry. The website uses privacy-friendly GoatCounter analytics without cookies."
	/>
	<link rel="canonical" href="https://starkdroid.github.io/Bento-Terminal/privacy/" />
</svelte:head>

<main id="content">
	<div class="docs-head">
		<h1>Privacy</h1>
		<p>No app telemetry. Privacy-friendly website analytics. No cookies.</p>
	</div>

	<div class="legal">
		<div class="panel">
			<h2>The app</h2>
			<p>
				Bento terminal does not send usage data, project paths, file contents, terminal output, or
				agent conversations anywhere. There is no account and no telemetry.
			</p>
			<p>
				On startup, Bento asks the public GitHub releases API whether a newer version exists.
				The request contains no account, project, file, path, or usage data. It only reads the
				latest release tag. Set <code>BENTO_NO_UPDATE_CHECK=1</code> to skip the check.
			</p>
			<p>
				Recent-project state is stored locally in your platform config directory under
				<code>bento/state.json</code>. It contains project paths and the file, caret, scroll, and
				focus you left in each one. Deleting that file clears the state.
			</p>

			<h2>The website</h2>
			<p>
				This site uses <a href="https://www.goatcounter.com/">GoatCounter</a> for aggregate page
				views and install-copy events. GoatCounter does not use cookies, does not store IP
				addresses, and does not track people across sites. It records aggregate page paths,
				referrers, browser/OS, screen size, and country.
			</p>
			<p>
				The analytics script is only present when the site is built with a GoatCounter site
				code. Forks and local builds without that code have no analytics.
			</p>
			<p>
				You can block the script with any content blocker; the site works without it.
			</p>

			<h2>Package downloads</h2>
			<p>
				GitHub, npm, Chocolatey, Homebrew, and Scoop publish aggregate download or install
				counts. Those are platform-level counts, not data sent by Bento.
			</p>

			<h2>Contact</h2>
			<p>
				Questions or removal requests:
				<a href="https://github.com/StarkDroid/Bento-Terminal/issues">open an issue</a>.
			</p>
			<p class="muted">Last updated: November 19, 2026.</p>
		</div>
	</div>
</main>
