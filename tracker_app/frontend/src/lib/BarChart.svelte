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
  let yTicks = $derived(Array.from({ length: maxHours + 1 }, (_, i) => maxHours - i));
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
    /* CSS Custom Properties */
    --chart-gap: 1rem;
    --chart-padding: 1rem;
    --bar-gap: 4px;

    --color-primary: #3b82f6;
    --color-average: #e5e7eb;
    --color-avg-line: #f97316;
    --color-border: #e5e7eb;
    --color-text-secondary: #6b7280;

    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  .chart-area {
    flex: 1;
    display: grid;
    grid-template-columns: auto 1fr;  /* Y-axis auto-sizes, chart takes rest */
    grid-template-rows: 1fr auto;     /* Bars take space, labels at bottom */
    column-gap: var(--chart-gap);
    padding: var(--chart-padding) 0;  /* Padding moved here - outside the chart space */
    min-height: 0;
  }

  .bars-row {
    /* Remove from grid flow - children become direct grid items */
    display: contents;
  }

  .y-axis {
    grid-column: 1;
    grid-row: 1;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    /* No padding - matches chart space exactly */
  }

  .y-tick {
    display: flex;
    align-items: center;
    gap: 0.25rem;
  }

  .y-label {
    font-size: var(--font-size-small);
    color: var(--color-text-secondary);
    text-align: right;
  }

  .tick-mark {
    width: 6px;
    height: 1px;
    background: var(--color-border);
  }

  .bars-area {
    grid-column: 2;
    grid-row: 1;
    display: flex;
    align-items: flex-end;
    gap: var(--chart-gap);
    position: relative;
    padding-left: var(--chart-padding);  /* Only left padding for spacing after border */
    border-left: 1px solid var(--color-border);
    min-height: 0;
  }

  .avg-line-wrapper {
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    pointer-events: none;
    /* Now references the same height as bars (no padding offset) */
  }

  .avg-line {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    border-top: 2px dotted var(--color-avg-line);
    display: flex;
    justify-content: flex-end;
    align-items: center;
  }

  .avg-label {
    font-size: var(--font-size-small);
    color: var(--color-avg-line);
    background: #ffffff;
    padding: 0 0.25rem;
    transform: translateY(-50%);
  }

  .bar-column {
    flex: 1;
    height: 100%;
    display: flex;
    align-items: flex-end;
    justify-content: center;
    gap: var(--bar-gap);
  }

  .bar {
    width: 40%;
    max-width: 2rem;
    border-radius: 4px 4px 0 0;
    transition: opacity 0.15s ease;
  }

  .bar-value {
    background: var(--color-primary);
  }

  .bar-average {
    background: var(--color-average);
  }

  .bar-column:hover .bar {
    opacity: 0.8;
  }

  .labels {
    /* Auto-aligns with bars - same grid column */
    grid-column: 2;
    grid-row: 2;
    display: flex;
    gap: var(--chart-gap);
    margin-top: 0.5rem;
    padding-left: var(--chart-padding);  /* Matches bars-area left padding */
  }

  .label {
    flex: 1;
    text-align: center;
    font-size: var(--font-size-small);
    color: var(--color-text-secondary);
  }

  .legend {
    display: flex;
    gap: 1.5rem;
    justify-content: center;
    padding-top: var(--chart-padding);
    border-top: 1px solid var(--color-border);
    font-size: var(--font-size-small);
    color: var(--color-text-secondary);
  }

  .legend-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .legend-box {
    width: 0.75rem;
    height: 0.75rem;
    border-radius: 2px;
  }

  .legend-value {
    background: var(--color-primary);
  }

  .legend-avg {
    background: var(--color-average);
  }
</style>
