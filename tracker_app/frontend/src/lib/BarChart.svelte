<script lang="ts">
  import type { Aggregation, DataPoint } from "$lib/utils";

  let { aggregations, transform, weeklyAvg }: {
    aggregations: Aggregation[];
    transform: (aggregations: Aggregation[]) => DataPoint[];
    weeklyAvg: number;
  } = $props();

  const OTHER_COLOR = '#9ca3af';

  let hoveredCategory: string | null = $state(null);

  let data = $derived(transform(aggregations));

  // Group flat DataPoints by label, preserving order of first appearance
  let bars = $derived.by(() => {
    const labelOrder: string[] = [];
    const labelMap = new Map<string, { category: string; value: number }[]>();

    for (const point of data) {
      if (!labelMap.has(point.label)) {
        labelOrder.push(point.label);
        labelMap.set(point.label, []);
      }
      if (point.category && point.value > 0) {
        labelMap.get(point.label)!.push({ category: point.category, value: point.value });
      }
    }

    return labelOrder.map(label => {
      const segments = (labelMap.get(label) || []).sort((a, b) => {
        if (a.category === 'Other') return 1;
        if (b.category === 'Other') return -1;
        return a.category.localeCompare(b.category);
      });
      return {
        label,
        segments,
        total: segments.reduce((s, seg) => s + seg.value, 0)
      };
    });
  });

  // Unique categories sorted alphabetically, "Other" always last
  let allCategories = $derived.by(() => {
    const catSet = new Set<string>();
    for (const bar of bars) {
      for (const seg of bar.segments) {
        catSet.add(seg.category);
      }
    }
    return Array.from(catSet)
      .filter(c => c !== 'Other')
      .sort((a, b) => a.localeCompare(b))
      .concat(catSet.has('Other') ? ['Other'] : []);
  });

  function categoryColor(name: string): string {
    let hash = 0;
    for (let i = 0; i < name.length; i++) {
      hash = (hash * 31 + name.charCodeAt(i)) >>> 0;
    }
    const hue = (hash * 137.508) % 360;
    return `hsl(${hue.toFixed(1)}, 65%, 52%)`;
  }

  let colorMap = $derived.by(() => {
    const map = new Map<string, string>();
    for (const cat of allCategories) {
      map.set(cat, cat === 'Other' ? OTHER_COLOR : categoryColor(cat));
    }
    return map;
  });

  let rawMax = $derived(Math.max(...bars.map(b => b.total), 1));
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

        <!-- Bar columns -->
        {#each bars as bar, index}
          {@const valuePercent = (bar.total / maxValue) * 100}
          <div class="bar-column">
            <div class="bar-stack" style="height: {valuePercent}%">
              {#each bar.segments as seg}
                {@const segPercent = bar.total > 0 ? (seg.value / bar.total) * 100 : 0}
                <div
                  class="segment"
                  class:dimmed={hoveredCategory !== null && hoveredCategory !== seg.category}
                  style="height: {segPercent}%; background: {colorMap.get(seg.category) ?? OTHER_COLOR}"
                  onmouseenter={() => hoveredCategory = seg.category}
                  onmouseleave={() => hoveredCategory = null}
                ></div>
              {/each}
            </div>
          </div>
          {#if index < bars.length - 1}
            <div class="grid-separator">
              <div class="grid-line-vertical"></div>
            </div>
          {/if}
        {/each}
      </div>
    </div>

    <div class="labels">
      {#each bars as bar, index}
        <span class="label">{bar.label}</span>
        {#if index < bars.length - 1}
          <div class="label-separator"></div>
        {/if}
      {/each}
    </div>
  </div>

  <div class="legend">
    {#each allCategories as cat}
      <span
        class="legend-item"
        class:dimmed={hoveredCategory !== null && hoveredCategory !== cat}
        onmouseenter={() => hoveredCategory = cat}
        onmouseleave={() => hoveredCategory = null}
        role="button"
        tabindex="0"
      >
        <span class="legend-box" style="background: {colorMap.get(cat)}"></span>
        {cat}
      </span>
    {/each}
  </div>
</div>

<style>
  .chart-container {
    /* CSS Custom Properties */
    --chart-gap: 1rem;
    --chart-padding: 1rem;
    --bar-gap: 4px;

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
    grid-template-columns: 1fr auto;
    grid-template-rows: 1fr auto;
    column-gap: var(--chart-gap);
    padding: var(--chart-padding) 0;
    min-height: 0;
  }

  .bars-row {
    display: contents;
  }

  .y-axis {
    grid-column: 2;
    grid-row: 1;
    position: relative;
  }

  .y-tick {
    position: absolute;
    right: 0;
    display: flex;
    align-items: center;
    gap: 0.25rem;
    flex-direction: row-reverse;
    transform: translateY(50%);
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
    padding-right: var(--chart-padding);
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
    bottom: calc(-0.5rem - 1.5em);
    border-left: 1px dotted var(--chart-grid, #d1d5db);
  }

  .avg-line-wrapper {
    grid-column: 1 / 2;
    grid-row: 1;
    position: relative;
    pointer-events: none;
    align-self: end;
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
    left: calc(100% + 0.2rem);
    font-size: var(--font-size-small);
    color: var(--color-avg-line);
    padding: 0 0;
    transition: background-color 0.3s ease;
  }

  .bar-column {
    flex: 1;
    height: 100%;
    display: flex;
    align-items: flex-end;
    justify-content: center;
    position: relative;
    z-index: 1;
  }

  .bar-stack {
    width: 40%;
    max-width: 2rem;
    display: flex;
    flex-direction: column;
    border-radius: 4px 4px 0 0;
    overflow: hidden;
  }

  .segment {
    width: 100%;
    flex-shrink: 0;
    transition: opacity 0.15s ease;
  }

  .segment.dimmed {
    opacity: 0.15;
  }

  .labels {
    grid-column: 1;
    grid-row: 2;
    display: flex;
    margin-top: 0.5rem;
    padding-right: var(--chart-padding);
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

  .legend {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem 1.25rem;
    justify-content: center;
    padding-top: var(--chart-padding);
    border-top: 1px solid var(--color-border);
    font-size: var(--font-size-small);
    color: var(--color-text-secondary);
  }

  .legend-item {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    cursor: default;
    transition: opacity 0.15s ease;
  }

  .legend-item.dimmed {
    opacity: 0.3;
  }

  .legend-box {
    width: 0.65rem;
    height: 0.65rem;
    border-radius: 2px;
    flex-shrink: 0;
  }
</style>
