<script lang="ts">
  import BarChart from "$lib/BarChart.svelte";
  import TopSitesList from "$lib/TopSitesList.svelte";
  import { GetAggregations } from "../../../wailsjs/go/main/App.js";
  import { onMount } from "svelte";

  type Aggregation = {
    groupers: Record<string, any>;
    duration: number;
  };

  type DataPoint = {
    label: string;
    value: number;
  };

  let dateAggregations: Aggregation[] = $state([]);
  let siteAggregations: Aggregation[] = $state([]);
  let isLoading = $state(true);

  onMount(async () => {
    await fetchWeeklyData();
  });

  function toDateId(date: Date): number {
    return date.getFullYear() * 10000 + (date.getMonth() + 1) * 100 + date.getDate();
  }

  // Builds up to 8 bars (weekday/weekend pair per week) with amortized values,
  // and computes overall weekday/weekend averages across all 4 weeks.
  function buildWeeklyBars(aggregations: Aggregation[]): {
    bars: DataPoint[];
    weekdayAvg: number;
    weekendAvg: number;
  } {
    const today = new Date();
    const dayOfWeek = today.getDay();
    const daysSinceMonday = dayOfWeek === 0 ? 6 : dayOfWeek - 1;
    const currentMonday = new Date(today);
    currentMonday.setDate(today.getDate() - daysSinceMonday);

    const durationMap = new Map<number, number>();
    for (const agg of aggregations) {
      durationMap.set(agg.groupers.date as number, agg.duration);
    }

    const bars: DataPoint[] = [];
    let totalWeekdaySeconds = 0;
    let totalWeekdayDays = 0;
    let totalWeekendSeconds = 0;
    let totalWeekendDays = 0;

    // 4 weeks, oldest first (w=3 is 3 weeks ago, w=0 is current week)
    for (let w = 3; w >= 0; w--) {
      const monday = new Date(currentMonday);
      monday.setDate(currentMonday.getDate() - w * 7);
      const weekLabel = `${monday.getMonth() + 1}/${monday.getDate()}`;

      let weekdaySeconds = 0;
      let weekdayDaysElapsed = 0;
      let weekendSeconds = 0;
      let weekendDaysElapsed = 0;

      for (let d = 0; d < 7; d++) {
        const date = new Date(monday);
        date.setDate(monday.getDate() + d);

        // Skip future dates
        if (date > today) continue;

        const dateId = toDateId(date);
        const duration = durationMap.get(dateId) || 0;
        const dow = date.getDay();

        if (dow === 0 || dow === 6) {
          weekendSeconds += duration;
          weekendDaysElapsed++;
        } else {
          weekdaySeconds += duration;
          weekdayDaysElapsed++;
        }
      }

      if (weekdayDaysElapsed > 0) {
        bars.push({
          label: `${weekLabel} WD`,
          value: Math.round(weekdaySeconds / (60 * weekdayDaysElapsed))
        });
        totalWeekdaySeconds += weekdaySeconds;
        totalWeekdayDays += weekdayDaysElapsed;
      }

      if (weekendDaysElapsed > 0) {
        bars.push({
          label: `${weekLabel} WE`,
          value: Math.round(weekendSeconds / (60 * weekendDaysElapsed))
        });
        totalWeekendSeconds += weekendSeconds;
        totalWeekendDays += weekendDaysElapsed;
      }
    }

    return {
      bars,
      weekdayAvg: totalWeekdayDays > 0 ? totalWeekdaySeconds / (60 * totalWeekdayDays) : 0,
      weekendAvg: totalWeekendDays > 0 ? totalWeekendSeconds / (60 * totalWeekendDays) : 0
    };
  }

  // Transform for BarChart: extracts just the bars from buildWeeklyBars
  function transformWeekly(aggregations: Aggregation[]): DataPoint[] {
    return buildWeeklyBars(aggregations).bars;
  }

  async function fetchWeeklyData() {
    const today = new Date();
    const dayOfWeek = today.getDay();
    const daysSinceMonday = dayOfWeek === 0 ? 6 : dayOfWeek - 1;
    const currentMonday = new Date(today);
    currentMonday.setDate(today.getDate() - daysSinceMonday);

    // 4 weeks ago Monday
    const startMonday = new Date(currentMonday);
    startMonday.setDate(currentMonday.getDate() - 21);

    const startDate = toDateId(startMonday).toString();
    const endDate = toDateId(today).toString();

    try {
      const [dateAggs, siteAggs] = await Promise.all([
        GetAggregations(["date"], { start_date: startDate, end_date: endDate }),
        GetAggregations(["url", "exe_path"], { start_date: startDate, end_date: endDate })
      ]);

      dateAggregations = dateAggs;
      siteAggregations = siteAggs;
    } catch (error) {
      console.error("Failed to fetch weekly aggregations:", error);
      dateAggregations = [];
      siteAggregations = [];
    } finally {
      isLoading = false;
    }
  }

  let weeklyResult = $derived.by(() => buildWeeklyBars(dateAggregations));

  let weekdayAvgFormatted = $derived(() => {
    const avg = weeklyResult.weekdayAvg;
    const hours = Math.floor(avg / 60);
    const minutes = Math.round(avg % 60);
    return { hours, minutes };
  });

  let weekendAvgFormatted = $derived(() => {
    const avg = weeklyResult.weekendAvg;
    const hours = Math.floor(avg / 60);
    const minutes = Math.round(avg % 60);
    return { hours, minutes };
  });

  let chartAvg = $derived(
    weeklyResult.bars.length > 0
      ? weeklyResult.bars.reduce((sum, b) => sum + b.value, 0) / weeklyResult.bars.length
      : 0
  );
</script>

<div class="content-card">
  <div class="chart-wrapper">
    <div class="chart-header">
      <div class="weekly-averages">
        <div class="avg-group">
          <div class="chart-label">Weekday Daily Average</div>
          <div class="chart-value">
            {weekdayAvgFormatted().hours}h {weekdayAvgFormatted().minutes}m
          </div>
        </div>
        <div class="avg-group">
          <div class="chart-label">Weekend Daily Average</div>
          <div class="chart-value">
            {weekendAvgFormatted().hours}h {weekendAvgFormatted().minutes}m
          </div>
        </div>
      </div>
    </div>
    {#if isLoading}
      <p class="loading">Loading...</p>
    {:else if dateAggregations.length > 0}
      <BarChart aggregations={dateAggregations} transform={transformWeekly} weeklyAvg={chartAvg} />
    {:else}
      <p class="no-data">No data available</p>
    {/if}
  </div>
</div>

{#if !isLoading && siteAggregations.length > 0}
  <div class="section-wrapper">
    <h2 class="section-heading">Top Sites & Apps (Past 4 Weeks)</h2>
    <div class="content-card">
      <TopSitesList aggregations={siteAggregations} />
    </div>
  </div>
{/if}

<style>
  .weekly-averages {
    display: flex;
    gap: 2.5rem;
  }

  .avg-group {
    display: flex;
    flex-direction: column;
  }
</style>
