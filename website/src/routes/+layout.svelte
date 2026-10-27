<script>
	import '../app.css';
	import { page } from '$app/stores';
	import Header from '$lib/Header.svelte';
	import Footer from '$lib/Footer.svelte';

	let { children } = $props();
</script>

<a class="skip-link" href="#content">Skip to content</a>

<div class="wrap">
	<Header path={$page.url.pathname} />
	{@render children()}
	<Footer />
</div>
