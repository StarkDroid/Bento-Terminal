<script>
	let { command, prompt = '$' } = $props();

	let copied = $state(false);
	let timer;

	async function copy() {
		try {
			await navigator.clipboard.writeText(command);
		} catch {
			const area = document.createElement('textarea');
			area.value = command;
			area.setAttribute('readonly', '');
			area.style.position = 'fixed';
			area.style.opacity = '0';
			document.body.append(area);
			area.select();
			document.execCommand('copy');
			area.remove();
		}
		copied = true;
		clearTimeout(timer);
		timer = setTimeout(() => (copied = false), 1600);
	}
</script>

<div class="cmd">
	<code><span class="prompt" aria-hidden="true">{prompt}</span>{command}</code>
	<button type="button" onclick={copy} aria-label={copied ? 'Copied' : 'Copy command'}>
		{copied ? 'copied' : 'copy'}
	</button>
</div>
<p class="sr-only" aria-live="polite">{copied ? 'Command copied' : ''}</p>
