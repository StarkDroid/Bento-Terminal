<script>
	import { onMount } from 'svelte';
	import '../app.css';
	import { page } from '$app/stores';
	import Header from '$lib/Header.svelte';
	import Footer from '$lib/Footer.svelte';

	let { children } = $props();

	// Move a soft light with a fine pointer. CSS keeps a fixed glow for touch,
	// for reduced motion, and for the prerendered page before hydration.
	onMount(() => {
		if (!matchMedia('(pointer: fine)').matches) return;
		if (matchMedia('(prefers-reduced-motion: reduce)').matches) return;

		let frame = 0;
		let x = 0;
		let y = 0;

		const apply = () => {
			frame = 0;
			document.documentElement.style.setProperty('--mx', `${x}px`);
			document.documentElement.style.setProperty('--my', `${y}px`);
		};

		const move = (event) => {
			x = event.clientX;
			y = event.clientY;
			if (!frame) frame = requestAnimationFrame(apply);
		};

		window.addEventListener('pointermove', move, { passive: true });
		return () => {
			window.removeEventListener('pointermove', move);
			if (frame) cancelAnimationFrame(frame);
		};
	});
</script>

<a class="skip-link" href="#content">Skip to content</a>

<div class="spotlight" aria-hidden="true"></div>

<div class="wrap">
	<Header path={$page.url.pathname} />
	{@render children()}
	<Footer />
</div>
