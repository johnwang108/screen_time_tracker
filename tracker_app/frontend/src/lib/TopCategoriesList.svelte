<script lang="ts">
  import type { Aggregation } from "$lib/utils";
  import { formatDuration } from "$lib/utils";

  let { aggregations }: { aggregations: Aggregation[] } = $props();

  // Sum durations per category and sort descending; uncategorized items grouped as "Other" at the end
  let categories = $derived.by(() => {
    const catMap = new Map<string, number>();
    let otherDuration = 0;
    for (const agg of aggregations) {
      const category = agg.groupers.category as string;
      if (category && category !== "Other") {
        catMap.set(category, (catMap.get(category) || 0) + agg.duration);
      } else {
        otherDuration += agg.duration;
      }
    }
    const sorted = Array.from(catMap.entries())
      .map(([name, duration]) => ({ name, duration, isOther: false }));
    if (otherDuration > 0) {
      sorted.push({ name: "Other", duration: otherDuration, isOther: true });
    }
    return sorted.sort((a, b) => b.duration - a.duration);
  });

  let maxDuration = $derived(categories.length > 0 ? categories[0].duration : 1);
</script>

<div class="top-categories">
  <div class="categories-list">
    {#each categories as cat}
      <div class="category-entry">
        <div class="category-info">
          <span class="category-name">{cat.name}</span>
          <span class="category-duration">{formatDuration(cat.duration)}</span>
        </div>
        <div class="progress-bar-container">
          <div
            class="progress-bar"
            style="width: {(cat.duration / maxDuration) * 100}%"
          ></div>
        </div>
      </div>
    {/each}
    {#if categories.length === 0}
      <p class="no-data">No category data available</p>
    {/if}
  </div>
</div>

<style>
  .top-categories {
    flex-shrink: 0;
  }

  .categories-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .category-entry {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .category-info {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .category-name {
    font-size: 0.875rem;
    color: var(--text-primary, #374151);
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
    margin-right: 0.5rem;
  }

  .category-duration {
    font-size: 0.875rem;
    color: var(--text-secondary, #6b7280);
    font-weight: 400;
    white-space: nowrap;
  }

  .progress-bar-container {
    height: 6px;
    background: var(--hover-bg, #f3f4f6);
    border-radius: 3px;
    overflow: hidden;
  }

  .progress-bar {
    height: 100%;
    background: var(--accent-bg, rgba(59, 130, 246, 0.4));
    border-radius: 3px;
    transition: width 0.3s ease;
  }

  .no-data {
    font-size: 0.875rem;
    color: var(--text-secondary);
    margin: 0;
  }
</style>
