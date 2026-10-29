<script>
	// The pane proportions come from internal/ui/layout.go at a 120 x 32 window:
	// Main fills the top left, Files and History stack on the right, and Build
	// spans the bottom above the status row.
	let { pane = '' } = $props();

	const panes = [
		{ id: 'main', x: 0, y: 0, w: 79, h: 20, label: 'Main' },
		{ id: 'files', x: 80, y: 0, w: 40, h: 11, label: 'Files' },
		{ id: 'history', x: 80, y: 12, w: 40, h: 8, label: 'History' },
		{ id: 'build', x: 0, y: 21, w: 120, h: 10, label: 'Build' }
	];

	const lit = $derived(panes.find((box) => box.id === pane));
</script>

<svg
	class="bento-map"
	viewBox="0 0 120 32"
	role="img"
	aria-label={lit ? `Bento layout with the ${lit.label} pane highlighted` : 'Bento layout'}
>
	<rect class="screen" x="0.5" y="0.5" width="119" height="31" rx="2" />
	{#each panes as box}
		<rect
			class="pane"
			class:lit={pane === box.id}
			x={box.x + 0.5}
			y={box.y + 0.5}
			width={box.w - 1}
			height={box.h - 1}
			rx="2"
		/>
	{/each}
	<rect class="status" x="1" y="31" width="118" height="0.8" rx="0.4" />
</svg>
