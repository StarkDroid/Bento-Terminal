<script>
	import { base } from '$app/paths';

	let { path = '/' } = $props();

	const docs = `${base}/docs/`;
	let onDocs = $derived(path === docs || path === docs.slice(0, -1));
</script>

<header class="site-header">
	<a class="brand" href="{base}/" aria-label="Bento terminal home">
		<span class="dot" aria-hidden="true"></span> bento
	</a>
	<nav class="site-nav" aria-label="Main">
		<a href="{base}/docs/" aria-current={onDocs ? 'page' : undefined}>Docs</a>
		<a href="https://github.com/StarkDroid/Bento-Terminal">GitHub</a>
		<a href="https://github.com/StarkDroid/Bento-Terminal/releases/latest">Releases</a>
	</nav>
</header>
