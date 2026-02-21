<script lang="ts">
  import TopSitesList from "$lib/TopSitesList.svelte";
  import TopCategoriesList from "$lib/TopCategoriesList.svelte";
  import type { Aggregation } from "$lib/utils";

  let { siteAggregations, categoryAggregations, heading }: {
    siteAggregations: Aggregation[];
    categoryAggregations: Aggregation[];
    heading: string;
  } = $props();

  let activeListTab: "sites" | "categories" = $state("sites");
</script>

<div class="section-wrapper">
  <div class="list-header">
    <h2 class="section-heading">{heading}</h2>
    <div class="list-tabs">
      <button class="list-tab" class:active={activeListTab === "sites"} onclick={() => activeListTab = "sites"}>Sites & Apps</button>
      <button class="list-tab" class:active={activeListTab === "categories"} onclick={() => activeListTab = "categories"}>Categories</button>
    </div>
  </div>
  <div class="content-card">
    {#if activeListTab === "sites"}
      <TopSitesList aggregations={siteAggregations} />
    {:else}
      <TopCategoriesList aggregations={categoryAggregations} />
    {/if}
  </div>
</div>

<style>
  .list-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 0.3rem;
  }

  .list-tabs {
    display: flex;
    gap: 0;
    border-bottom: 1px solid var(--border-color);
  }

  .list-tab {
    font-size: 0.8rem;
    font-weight: 500;
    padding: 0.35rem 0.85rem;
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
    transition: color 0.15s ease, box-shadow 0.15s ease;
  }

  .list-tab:hover {
    color: var(--text-secondary);
  }

  .list-tab.active {
    color: var(--accent-color);
    font-weight: 600;
    box-shadow: inset 0 -2px 0 var(--accent-color);
  }
</style>
