<script lang="ts">
  type TopSite = {
    name: string;
    duration: number; // in seconds
  };

  type Props = {
    sites: TopSite[];
  };

  let { sites }: Props = $props();

  // Format duration from seconds to "Xh Ym" format
  function formatDuration(seconds: number): string {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);

    if (hours > 0) {
      return `${hours}h ${minutes}m`;
    }
    return `${minutes}m`;
  }

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
