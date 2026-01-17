<script lang="ts">
	import BarChart from '$lib/BarChart.svelte';

	// Mock data: daily screen time (minutes) vs historical average
	const weekData = [
		{ label: 'Mon', value: 320, average: 280 },
		{ label: 'Tue', value: 290, average: 280 },
		{ label: 'Wed', value: 350, average: 280 },
		{ label: 'Thu', value: 260, average: 280 },
		{ label: 'Fri', value: 310, average: 280 },
		{ label: 'Sat', value: 180, average: 200 },
		{ label: 'Sun', value: 150, average: 200 }
	];

	const dailyAvg = weekData.reduce((sum, d) => sum + d.value, 0) / weekData.length;
	const dailyAvgHours = (dailyAvg / 60).toFixed(1);
</script>

<div class="container">
	<div class="left-panel">
		<div class="chart-wrapper">
			<h2>Daily Average: {dailyAvgHours}h</h2>
			<BarChart data={weekData} weeklyAvg={dailyAvg} />
		</div>
	</div>
	<div class="right-panel">
		<!-- Reserved for future content -->
	</div>
</div>

<style>
	:global(html, body) {
		margin: 0;
		padding: 0;
		height: 100%;
		overflow: hidden;
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
		background: #f9fafb;

		--font-size-heading: 1.25rem;
		--font-size-small: 12px;
	}

	.container {
		display: grid;
		grid-template-columns: 1fr 1fr;
		height: 100vh;
		overflow: hidden;
	}

	.left-panel {
		padding: 2rem;
		background: #ffffff;
		border-right: 1px solid #e5e7eb;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}

	.right-panel {
		padding: 2rem;
		background: #f9fafb;
	}

	h2 {
		margin: 0 0 1rem 0;
		font-size: var(--font-size-heading);
		font-weight: 600;
		color: #111827;
		flex-shrink: 0;
		text-align: center;
	}

	.chart-wrapper {
		height: 400px;
		display: flex;
		flex-direction: column;
	}
</style>
