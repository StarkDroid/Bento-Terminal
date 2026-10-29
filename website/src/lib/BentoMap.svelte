<script>
	// The pane grid follows internal/ui/layout.go: Main on the top left, Files
	// and History stacked on the right, Build across the bottom. Terminal cells
	// are about twice as tall as they are wide, so each row is two units high;
	// the window adds a macOS Terminal title bar and an even gutter.
	let { pane = '' } = $props();

	const panes = [
		{ id: 'main', x: 3, y: 10, w: 79, h: 41, label: 'Main' },
		{ id: 'files', x: 85, y: 10, w: 38, h: 22, label: 'Files' },
		{ id: 'history', x: 85, y: 35, w: 38, h: 16, label: 'History' },
		{ id: 'build', x: 3, y: 54, w: 120, h: 20, label: 'Build' }
	];

	const lit = $derived(panes.find((box) => box.id === pane));
</script>

<svg
	class="bento-map"
	viewBox="0 0 126 78"
	role="img"
	aria-label={lit ? `Bento layout with the ${lit.label} pane highlighted` : 'Bento layout'}
>
	<rect class="window" x="0.5" y="0.5" width="125" height="77" rx="3" />
	<path
		class="titlebar"
		d="M3.5,0.5 H122.5 A3,3 0 0 1 125.5,3.5 V7 H0.5 V3.5 A3,3 0 0 1 3.5,0.5 Z"
	/>
	<line class="divider" x1="0.5" y1="7" x2="125.5" y2="7" />
	<circle class="dot red" cx="4.6" cy="3.75" r="0.9" />
	<circle class="dot yellow" cx="7.4" cy="3.75" r="0.9" />
	<circle class="dot green" cx="10.2" cy="3.75" r="0.9" />
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
