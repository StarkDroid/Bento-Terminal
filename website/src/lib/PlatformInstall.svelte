<script>
	import { onMount } from 'svelte';
	import { base } from '$app/paths';
	import CopyCommand from '$lib/CopyCommand.svelte';
	import { trackEvent } from '$lib/analytics';

	const installSh =
		'curl -fsSL https://raw.githubusercontent.com/StarkDroid/Bento-Terminal/main/scripts/install.sh | sh';
	const installPs1 =
		'irm https://raw.githubusercontent.com/StarkDroid/Bento-Terminal/main/scripts/install.ps1 | iex';
	const installGo =
		'go install github.com/StarkDroid/Bento-Terminal/cmd/bento@latest';

	const platforms = [
		{
			id: 'macos',
			label: 'macOS',
			prompt: '$',
			command: installSh,
			event: 'copy-hero-macos',
			note: 'macOS'
		},
		{
			id: 'linux',
			label: 'Linux',
			prompt: '$',
			command: installSh,
			event: 'copy-hero-linux',
			note: 'Linux'
		},
		{
			id: 'windows',
			label: 'Windows',
			prompt: 'PS>',
			command: installPs1,
			event: 'copy-hero-windows',
			note: 'Windows PowerShell'
		},
		{
			id: 'npm',
			label: 'npm',
			prompt: '$',
			command: 'npx bento-terminal',
			event: 'copy-hero-npm',
			note: 'npm / npx'
		},
		{
			id: 'go',
			label: 'Go',
			prompt: '$',
			command: installGo,
			event: 'copy-hero-go',
			note: 'Any platform with Go 1.23+'
		}
	];

	let platform = $state('macos');
	let mobile = $state(false);
	let known = $state(true);

	const current = $derived(
		platforms.find((item) => item.id === platform) ?? platforms[0]
	);

	function choose(id) {
		platform = id;
		trackEvent(`install-platform-${id}`);
	}

	onMount(() => {
		const ua = navigator.userAgent;
		const isIPad = /Macintosh/.test(ua) && navigator.maxTouchPoints > 1;
		const isIOS = /iPhone|iPad|iPod/.test(ua) || isIPad;
		const isAndroid = /Android/.test(ua);
		const isWindows = /Windows/.test(ua);
		const isMac = /Macintosh|Mac OS X/.test(ua) && !isIPad;
		const isLinux = /Linux|X11|CrOS/.test(ua) && !isAndroid;
		const isMobileUA = isIOS || isAndroid || /Mobile/.test(ua);
		const narrow = window.matchMedia('(max-width: 760px)');

		const apply = () => {
			mobile = isMobileUA || narrow.matches;
			if (isWindows) {
				platform = 'windows';
			} else if (isMac) {
				platform = 'macos';
			} else if (isAndroid) {
				platform = 'linux';
			} else if (isLinux) {
				platform = 'linux';
			} else if (isIOS) {
				platform = 'macos';
			} else {
				platform = 'macos';
				known = false;
			}
		};

		apply();
		narrow.addEventListener('change', apply);
		return () => narrow.removeEventListener('change', apply);
	});
</script>

<div class="install">
	{#if mobile || !known}
		<div class="platforms" role="group" aria-label="Choose your install method">
			{#each platforms as item (item.id)}
				<button
					type="button"
					class:active={item.id === platform}
					aria-pressed={item.id === platform}
					onclick={() => choose(item.id)}
				>
					{item.label}
				</button>
			{/each}
		</div>
	{/if}

	<CopyCommand command={current.command} prompt={current.prompt} event={current.event} />

	<p class="hint">
		{current.note}.
		{#if mobile}
			<a href="{base}/docs/#install">More install options</a>
		{:else}
			<a href="{base}/docs/#install">Homebrew, Scoop, Go, npm and more</a>
			· <a href="{base}/docs/">Docs</a>
		{/if}
	</p>
</div>

<style>
	.platforms {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		margin-bottom: 10px;
	}

	.platforms button {
		transition:
			background-color 140ms ease,
			border-color 140ms ease,
			color 140ms ease;
		border: 1px solid var(--line);
		border-radius: 7px;
		background: var(--panel);
		color: var(--sand-dim);
		font: inherit;
		font-size: 0.78rem;
		padding: 6px 12px;
		cursor: pointer;
	}

	.platforms button:hover {
		border-color: var(--moss);
		color: var(--sand);
	}

	.platforms button.active {
		border-color: var(--bright);
		background: var(--status);
		color: var(--bright);
	}

	@media (prefers-reduced-motion: reduce) {
		.platforms button {
			transition: none;
		}
	}
</style>
