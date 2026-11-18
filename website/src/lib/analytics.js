// Website analytics are intentionally small and privacy-friendly. The app
// itself has no telemetry; this only counts page views and install-copy events.
export function trackEvent(name) {
	if (typeof window === 'undefined') return;
	const counter = window.goatcounter;
	if (counter && typeof counter.count === 'function') {
		counter.count({ path: name, event: true });
	}
}
