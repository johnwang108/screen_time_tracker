<script lang="ts">
  import BarChart from "$lib/BarChart.svelte";
  import TopListSection from "$lib/TopListSection.svelte";
  import { GetAggregations } from "../../../wailsjs/go/main/App.js";
  import { onMount } from "svelte";
  import type { Aggregation, DataPoint } from "$lib/utils";

  let dateAggregations: Aggregation[] = $state([]);
  let siteAggregations: Aggregation[] = $state([]);
  let categoryAggregations: Aggregation[] = $state([]);
  let isLoading = $state(true);
  let daysElapsed = $state(7);
  let weekOverWeekChange = $state<number | null>(null);

  onMount(async () => {
    await fetchWeekData();
  });

  // Transform for BarChart: one DataPoint per (date × category) for stacked rendering
  function transformDaily(aggregations: Aggregation[]): DataPoint[] {
    const today = new Date();
    const dayOfWeek = today.getDay();
    const daysSinceMonday = dayOfWeek === 0 ? 6 : dayOfWeek - 1;
    const monday = new Date(today);
    monday.setDate(today.getDate() - daysSinceMonday);

    const dateIds: number[] = [];
    for (let i = 0; i < 7; i++) {
      const date = new Date(monday);
      date.setDate(monday.getDate() + i);
      dateIds.push(date.getFullYear() * 10000 + (date.getMonth() + 1) * 100 + date.getDate());
    }

    // dateId -> category -> seconds
    const dateMap = new Map<number, Map<string, number>>();
    for (const agg of aggregations) {
      const dateId = agg.groupers.date as number;
      const category = (agg.groupers.category as string) || "Other";
      if (!dateMap.has(dateId)) dateMap.set(dateId, new Map());
      const catMap = dateMap.get(dateId)!;
      catMap.set(category, (catMap.get(category) || 0) + agg.duration);
    }

    const dayLabels = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
    const result: DataPoint[] = [];
    for (let i = 0; i < 7; i++) {
      const label = dayLabels[i];
      const catMap = dateMap.get(dateIds[i]);
      if (catMap && catMap.size > 0) {
        for (const [category, duration] of catMap) {
          result.push({ label, value: Math.round(duration / 60), category });
        }
      } else {
        result.push({ label, value: 0 });
      }
    }
    return result;
  }

  async function fetchWeekData() {
    const today = new Date();
    const dayOfWeek = today.getDay();
    const daysSinceMonday = dayOfWeek === 0 ? 6 : dayOfWeek - 1;
    daysElapsed = daysSinceMonday + 1;

    const monday = new Date(today);
    monday.setDate(today.getDate() - daysSinceMonday);

    const weekDates: Date[] = [];
    for (let i = 0; i < 7; i++) {
      const date = new Date(monday);
      date.setDate(monday.getDate() + i);
      weekDates.push(date);
    }

    const dateIds = weekDates.map(date =>
      date.getFullYear() * 10000 + (date.getMonth() + 1) * 100 + date.getDate()
    );

    const startDate = dateIds[0].toString();
    const endDate = dateIds[6].toString();

    // Last week date range for week-over-week comparison
    const lastWeekMonday = new Date(monday);
    lastWeekMonday.setDate(monday.getDate() - 7);
    const lastWeekDateIds: number[] = [];
    for (let i = 0; i < 7; i++) {
      const date = new Date(lastWeekMonday);
      date.setDate(lastWeekMonday.getDate() + i);
      lastWeekDateIds.push(date.getFullYear() * 10000 + (date.getMonth() + 1) * 100 + date.getDate());
    }
    const lastWeekStartDate = lastWeekDateIds[0].toString();
    const lastWeekEndDate = lastWeekDateIds[6].toString();

    try {
      const [dateAggs, siteAggs, catAggs, lastWeekAggs] = await Promise.all([
        GetAggregations(["date", "category"], { start_date: startDate, end_date: endDate }),
        GetAggregations(["url", "exe_path"], { start_date: startDate, end_date: endDate }),
        GetAggregations(["category"], { start_date: startDate, end_date: endDate }),
        GetAggregations(["date"], { start_date: lastWeekStartDate, end_date: lastWeekEndDate })
      ]);

      dateAggregations = dateAggs;
      siteAggregations = siteAggs;
      categoryAggregations = catAggs;

      // Week-over-week change
      const lastWeekTotalSeconds = lastWeekAggs.reduce((sum: number, agg: Aggregation) => sum + agg.duration, 0);
      const lastWeekAvgMinutes = lastWeekTotalSeconds / (60 * 7);

      const thisWeekTotalSeconds = dateAggs.reduce((sum: number, agg: Aggregation) => sum + agg.duration, 0);
      const thisWeekAvgMinutes = thisWeekTotalSeconds / (60 * daysElapsed);

      if (lastWeekAvgMinutes > 0) {
        weekOverWeekChange = ((thisWeekAvgMinutes - lastWeekAvgMinutes) / lastWeekAvgMinutes) * 100;
      } else if (thisWeekAvgMinutes > 0) {
        weekOverWeekChange = Infinity;
      } else {
        weekOverWeekChange = null;
      }
    } catch (error) {
      console.error("Failed to fetch aggregations:", error);
      dateAggregations = [];
      siteAggregations = [];
    } finally {
      isLoading = false;
    }
  }

  let dailyAvg = $derived(
    dateAggregations.length > 0 && daysElapsed > 0
      ? dateAggregations.reduce((sum, agg) => sum + agg.duration, 0) / (60 * daysElapsed)
      : 0
  );

  let dailyAvgFormatted = $derived(() => {
    const hours = Math.floor(dailyAvg / 60);
    const minutes = Math.round(dailyAvg % 60);
    return { hours, minutes };
  });
</script>

<div class="content-card">
  <div class="chart-wrapper">
    <div class="chart-header">
      <div class="chart-header-left">
        <div class="chart-label">Daily Average</div>
        <div class="chart-value">
          {dailyAvgFormatted().hours}h {dailyAvgFormatted().minutes}m
        </div>
      </div>
      {#if weekOverWeekChange !== null}
        <div class="week-over-week" class:positive={weekOverWeekChange > 0} class:negative={weekOverWeekChange < 0}>
          {#if weekOverWeekChange === Infinity}
            ↑ New data this week
          {:else}
            {weekOverWeekChange > 0 ? '↑' : '↓'} {Math.abs(weekOverWeekChange).toFixed(1)}% from last week
          {/if}
        </div>
      {/if}
    </div>
    {#if isLoading}
      <p class="loading">Loading...</p>
    {:else if dateAggregations.length > 0}
      <BarChart aggregations={dateAggregations} transform={transformDaily} weeklyAvg={dailyAvg} />
    {:else}
      <p class="no-data">No data available</p>
    {/if}
  </div>
</div>

{#if !isLoading && siteAggregations.length > 0}
  <TopListSection heading="Top This Week" {siteAggregations} {categoryAggregations} />
{/if}

<style>
  .week-over-week {
    font-size: 0.875rem;
    font-weight: 500;
    padding: 0.25rem 0.75rem;
    border-radius: 6px;
    white-space: nowrap;
  }

  .week-over-week.positive {
    color: #059669;
    background: rgba(5, 150, 105, 0.1);
  }

  .week-over-week.negative {
    color: #dc2626;
    background: rgba(220, 38, 38, 0.1);
  }

</style>
