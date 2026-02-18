<script lang="ts">
  type DataPoint = {
    label: string;
    value: number;
  };

  let { data, weeklyAvg }: { data: DataPoint[]; weeklyAvg: number } = $props();

  let rawMax = $derived(Math.max(...data.map((d) => d.value), 1));
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
          {@const tickPosition = (hour / maxHours) * 100}
          <div class="y-tick" style="bottom: {tickPosition}%">
            <span class="y-label">{hour}h</span>
            <span class="tick-mark"></span>
          </div>
        {/each}
      </div>

      <!-- Average line wrapper - spans both columns -->
      <div class="avg-line-wrapper" style="height: {weeklyAvgPercent}%">
        <div class="avg-line">
          <span class="avg-label">avg</span>
        </div>
      </div>

      <div class="bars-area">
        <!-- Horizontal grid lines -->
        <div class="grid-container">
          {#each yTicks as hour}
            {@const gridLineHeight = (hour / maxHours) * 100}
            <div class="grid-line-horizontal" style="bottom: {gridLineHeight}%"></div>
          {/each}
        </div>

        <!-- Bar columns with vertical grid separators -->
        {#each data as point, index}
          {@const valuePercent = (point.value / maxValue) * 100}
          <div class="bar-column">
            <div class="bar bar-value" style="height: {valuePercent}%"></div>
          </div>
          {#if index < data.length - 1}
            <div class="grid-separator">
              <div class="grid-line-vertical"></div>
            </div>
          {/if}
        {/each}
      </div>
    </div>
    <div class="labels">
      {#each data as point, index}
        <span class="label">{point.label}</span>
        {#if index < data.length - 1}
          <div class="label-separator"></div>
        {/if}
      {/each}
    </div>
  </div>

  <!-- Legend temporarily hidden - will be re-implemented later
  <div class="legend">
    <span class="legend-item"><span class="legend-box legend-value"></span> This week</span>
    <span class="legend-item"><span class="legend-box legend-avg"></span> Historical avg</span>
  </div>
  -->
</div>

<style>
  .chart-container {
    /* CSS Custom Properties */
    --chart-gap: 1rem;
    --chart-padding: 1rem;
    --bar-gap: 4px;

    --color-primary: var(--accent-color, #3b82f6);
    --color-average: var(--border-color, #e5e7eb);
    --color-avg-line: var(--avg-line-color, #f97316);
    --color-border: var(--border-color, #e5e7eb);
    --color-text-secondary: var(--text-secondary, #6b7280);

    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  .chart-area {
    flex: 1;
    display: grid;
    grid-template-columns: 1fr auto;  /* Chart takes rest, Y-axis auto-sizes */
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
    grid-column: 2;
    grid-row: 1;
    position: relative;
    /* No padding - matches chart space exactly */
  }

  .y-tick {
    position: absolute;
    right: 0;
    display: flex;
    align-items: center;
    gap: 0.25rem;
    flex-direction: row-reverse;  /* Tick mark on left, label on right */
    transform: translateY(50%);  /* Center the tick on its position */
  }

  .y-label {
    font-size: var(--font-size-small);
    color: var(--color-text-secondary);
    text-align: left;
  }

  .tick-mark {
    width: 6px;
    height: 1px;
    background: var(--color-border);
  }

  .bars-area {
    grid-column: 1;
    grid-row: 1;
    display: flex;
    align-items: flex-end;
    position: relative;
    padding-right: var(--chart-padding);  /* Only right padding for spacing before border */
    border-left: 1px solid var(--color-border);
    border-right: 1px solid var(--color-border);
    min-height: 0;
  }

  .grid-container {
    position: absolute;
    left: 0;
    right: 0;
    top: 0;
    bottom: 0;
    pointer-events: none;
    z-index: 0;
  }

  .grid-line-horizontal {
    position: absolute;
    left: 0;
    right: 0;
    border-top: 1px dotted var(--chart-grid, #d1d5db);
  }

  .grid-separator {
    width: var(--chart-gap);
    height: 100%;
    position: relative;
    flex-shrink: 0;
  }

  .grid-line-vertical {
    position: absolute;
    left: 50%;
    top: 0;
    bottom: calc(-0.5rem - 1.5em);  /* Extend down through labels (margin-top + label height) */
    border-left: 1px dotted var(--chart-grid, #d1d5db);
  }

  .avg-line-wrapper {
    grid-column: 1 / 2;  /* Span both bars area and y-axis */
    grid-row: 1;
    position: relative;
    pointer-events: none;
    align-self: end;  /* Align to bottom of grid cell */
  }

  .avg-line {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    border-top: 2px dotted var(--color-avg-line);
    display: flex;
    align-items: center;
  }

  .avg-label {
    position: absolute;
    /* left: 70%;  */
    left: calc(100% + 0.2rem);
    font-size: var(--font-size-small);
    color: var(--color-avg-line);
    /* background: var(--card-bg, #f5f5f5); */
    padding: 0 0;
    /* transform: translate(-50%, -50%); */
    transition: background-color 0.3s ease;
  }

  .bar-column {
    flex: 1;
    height: 100%;
    display: flex;
    align-items: flex-end;
    justify-content: center;
    gap: var(--bar-gap);
    position: relative;
    z-index: 1;
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
    grid-column: 1;
    grid-row: 2;
    display: flex;
    margin-top: 0.5rem;
    padding-right: var(--chart-padding);  /* Matches bars-area right padding */
  }

  .label {
    flex: 1;
    text-align: center;
    font-size: var(--font-size-small);
    color: var(--color-text-secondary);
  }

  .label-separator {
    width: var(--chart-gap);
    flex-shrink: 0;
  }
/* 
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
  } */
</style>
