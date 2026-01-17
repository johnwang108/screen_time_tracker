<script lang="ts">
	type DataPoint = {
		label: string;
		value: number;
		average: number;
	};

	let { data, weeklyAvg }: { data: DataPoint[]; weeklyAvg: number } = $props();

	let rawMax = $derived(Math.max(...data.map((d) => Math.max(d.value, d.average)), 1));
	let maxHours = $derived(Math.ceil(rawMax / 60));
	let maxValue = $derived(maxHours * 60);
	let yTicks = $derived(
		Array.from({ length: maxHours + 1 }, (_, i) => maxHours - i)
	);
	let weeklyAvgPercent = $derived((weeklyAvg / maxValue) * 100);
</script>

<div class="chart-container">
	<div class="chart-area">
		<div class="bars-row">
			<div class="y-axis">
				{#each yTicks as hour}
					<div class="y-tick">
						<span class="y-label">{hour}h</span>
						<span class="tick-mark"></span>
					</div>
				{/each}
			</div>
			<div class="bars-area">
				<div class="avg-line-wrapper" style="height: {weeklyAvgPercent}%">
					<div class="avg-line"><span class="avg-label">avg</span></div>
				</div>
				{#each data as point}
					{@const valuePercent = (point.value / maxValue) * 100}
					{@const avgPercent = (point.average / maxValue) * 100}
					<div class="bar-column">
						<div class="bar bar-value" style="height: {valuePercent}%"></div>
						<div class="bar bar-average" style="height: {avgPercent}%"></div>
					</div>
				{/each}
			</div>
		</div>
		<div class="labels">
			{#each data as point}
				<span class="label">{point.label}</span>
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
		flex-direction: column;
		min-height: 0;
	}

	.bars-row {
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
		font-size: var(--font-size-small);
		color: #6b7280;
		text-align: right;
		min-width: 32px;
	}

	.tick-mark {
		width: 6px;
		height: 1px;
		background: #d1d5db;
	}

	.bars-area {
		flex: 1;
		display: flex;
		align-items: flex-end;
		gap: 1rem;
		position: relative;
		padding: 1rem 0;
		border-left: 1px solid #e5e7eb;
		padding-left: 1rem;
	}

	.avg-line-wrapper {
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0;
		pointer-events: none;
	}

	.avg-line {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		border-top: 2px dotted #f97316;
		display: flex;
		justify-content: flex-end;
		align-items: center;
	}

	.avg-label {
		font-size: var(--font-size-small);
		color: #f97316;
		background: #ffffff;
		padding: 0 4px;
		transform: translateY(-50%);
	}

	.bar-column {
		flex: 1;
		height: 100%;
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

	.bar-column:hover .bar {
		opacity: 0.8;
	}

	.labels {
		display: flex;
		gap: 1rem;
		padding-top: 0.5rem;
		margin-left: 50px;
		padding-left: 1rem;
	}

	.label {
		flex: 1;
		text-align: center;
		font-size: var(--font-size-small);
		color: #6b7280;
	}

	.legend {
		display: flex;
		gap: 1.5rem;
		justify-content: center;
		padding-top: 1rem;
		border-top: 1px solid #e5e7eb;
		font-size: var(--font-size-small);
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
