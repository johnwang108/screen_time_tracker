<script lang="ts">
  import type { Aggregation } from "$lib/utils";
  import { formatDuration } from "$lib/utils";

  let { aggregations }: { aggregations: Aggregation[] } = $props();

  // Extract app name from exe path (e.g. "C:\...\Code.exe" -> "Code")
  function extractAppName(exePath: string): string {
    const segments = exePath.split(/[/\\]/);
    const filename = segments[segments.length - 1] || exePath;
    return filename.replace(/\.exe$/i, '');
  }

  // Group by url for websites, exe_path for apps; sum durations, sort, take top 10
  let sites = $derived.by(() => {
    const siteMap = new Map<string, number>();
    for (const agg of aggregations) {
      const url = agg.groupers.url as string;
      const exePath = agg.groupers.exe_path as string;
      const identifier = url && url.trim() !== ""
        ? url
        : extractAppName(exePath || "Unknown");
      if (identifier) {
        const current = siteMap.get(identifier) || 0;
        siteMap.set(identifier, current + agg.duration);
      }
    }
    return Array.from(siteMap.entries())
      .map(([name, duration]) => ({ name, duration }))
      .sort((a, b) => b.duration - a.duration)
      .slice(0, 10);
  });

  // Get max duration for scaling bars
  let maxDuration = $derived(
    sites.length > 0 ? sites[0].duration : 1
  );
</script>

<div class="top-sites">
  <div class="sites-list">
    {#each sites as site}
      <div class="site-entry">
        <div class="site-info">
          <span class="site-name">{site.name}</span>
          <span class="site-duration">{formatDuration(site.duration)}</span>
        </div>
        <div class="progress-bar-container">
          <div
            class="progress-bar"
            style="width: {(site.duration / maxDuration) * 100}%"
          ></div>
        </div>
      </div>
    {/each}
  </div>
</div>

<style>
  .top-sites {
    flex-shrink: 0;
  }

  .sites-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .site-entry {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .site-info {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .site-name {
    font-size: 0.875rem;
    color: var(--text-primary, #374151);
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
    margin-right: 0.5rem;
  }

  .site-duration {
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
</style>
