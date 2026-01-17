<script lang="ts">
	type DataPoint = {
		label: string;
		value: number;
		average: number;
	};

	let { data }: { data: DataPoint[] } = $props();

	let rawMax = $derived(Math.max(...data.map((d) => Math.max(d.value, d.average)), 1));
	let maxHours = $derived(Math.ceil(rawMax / 60));
	let maxValue = $derived(maxHours * 60);
	let yTicks = $derived(
		Array.from({ length: maxHours + 1 }, (_, i) => maxHours - i)
	);
</script>

<div class="chart-container">
	<div class="chart-area">
		<div class="y-axis">
			{#each yTicks as hour}
				<div class="y-tick">
					<span class="y-label">{hour}h</span>
					<span class="tick-mark"></span>
				</div>
			{/each}
		</div>
		<div class="chart">
			{#each data as point}
				{@const valuePercent = (point.value / maxValue) * 100}
				{@const avgPercent = (point.average / maxValue) * 100}
				<div class="bar-group">
					<div class="bars">
						<div class="bar bar-value" style="height: {valuePercent}%"></div>
						<div class="bar bar-average" style="height: {avgPercent}%"></div>
					</div>
					<span class="label">{point.label}</span>
				</div>
			{/each}
		</div>
	</div>

	<div class="legend">
		<span class="legend-item"><span class="legend-box legend-value"></span> This week</span>
		<span class="legend-item"><span class="legend-box legend-avg"></span> Historical avg</span>
	</div>
</div>

<style>
	.chart-container {
		width: 100%;
		height: 100%;
		display: flex;
		flex-direction: column;
	}

	.chart-area {
		flex: 1;
		display: flex;
		min-height: 0;
	}

	.y-axis {
		display: flex;
		flex-direction: column;
		justify-content: space-between;
		padding: 1rem 0;
	}

	.y-tick {
		display: flex;
		align-items: center;
		gap: 4px;
	}

	.y-label {
		font-size: 11px;
		color: #6b7280;
		text-align: right;
		min-width: 32px;
	}

	.tick-mark {
		width: 6px;
		height: 1px;
		background: #d1d5db;
	}

	.chart {
		flex: 1;
		display: flex;
		align-items: flex-end;
		gap: 1rem;
		padding: 1rem 0;
		min-height: 0;
		border-left: 1px solid #e5e7eb;
		padding-left: 1rem;
	}

	.bar-group {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		height: 100%;
	}

	.bars {
		flex: 1;
		width: 100%;
		display: flex;
		align-items: flex-end;
		justify-content: center;
		gap: 4px;
	}

	.bar {
		width: 40%;
		max-width: 32px;
		border-radius: 4px 4px 0 0;
		transition: opacity 0.15s ease;
	}

	.bar-value {
		background: #3b82f6;
	}

	.bar-average {
		background: #e5e7eb;
	}

	.bar-group:hover .bar {
		opacity: 0.8;
	}

	.label {
		margin-top: 0.5rem;
		font-size: 12px;
		color: #6b7280;
	}

	.legend {
		display: flex;
		gap: 1.5rem;
		justify-content: center;
		padding-top: 1rem;
		border-top: 1px solid #e5e7eb;
		font-size: 12px;
		color: #6b7280;
	}

	.legend-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.legend-box {
		width: 12px;
		height: 12px;
		border-radius: 2px;
	}

	.legend-value {
		background: #3b82f6;
	}

	.legend-avg {
		background: #e5e7eb;
	}
</style>
