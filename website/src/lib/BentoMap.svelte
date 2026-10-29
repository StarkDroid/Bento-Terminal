<script>
	// The pane grid follows internal/ui/layout.go: Main on the top left, Files
	// and History stacked on the right, Build across the bottom. The window adds
	// a macOS Terminal title bar and an even gutter around and between panes.
	let { pane = '' } = $props();

	const panes = [
		{ id: 'main', x: 3.5, y: 8.5, w: 72, h: 18, label: 'Main' },
		{ id: 'files', x: 78.5, y: 8.5, w: 38, h: 10, label: 'Files' },
		{ id: 'history', x: 78.5, y: 21.5, w: 38, h: 5, label: 'History' },
		{ id: 'build', x: 3.5, y: 29.5, w: 113, h: 9, label: 'Build' }
	];

	const lit = $derived(panes.find((box) => box.id === pane));
</script>

<svg
	class="bento-map"
	viewBox="0 0 120 42"
	role="img"
	aria-label={lit ? `Bento layout with the ${lit.label} pane highlighted` : 'Bento layout'}
>
	<rect class="window" x="0.5" y="0.5" width="119" height="41" rx="2.5" />
	<path
		class="titlebar"
		d="M3,0.5 H117 A2.5,2.5 0 0 1 119.5,3 V5.5 H0.5 V3 A2.5,2.5 0 0 1 3,0.5 Z"
	/>
	<line class="divider" x1="0.5" y1="5.5" x2="119.5" y2="5.5" />
	<circle class="dot red" cx="4.5" cy="3" r="0.8" />
	<circle class="dot yellow" cx="7.3" cy="3" r="0.8" />
	<circle class="dot green" cx="10.1" cy="3" r="0.8" />
	{#each panes as box}
		<rect
			class="pane"
			class:lit={pane === box.id}
			x={box.x}
			y={box.y}
			width={box.w}
			height={box.h}
			rx="1"
		/>
	{/each}
</svg>
