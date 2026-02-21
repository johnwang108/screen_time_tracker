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

    // dateId -> category -> seconds
    const dateMap = new Map<number, Map<string, number>>();
    for (const agg of aggregations) {
      const dateId = agg.groupers.date as number;
      const category = (agg.groupers.category as string) || "Other";
      if (!dateMap.has(dateId)) dateMap.set(dateId, new Map());
      const catMap = dateMap.get(dateId)!;
      catMap.set(category, (catMap.get(category) || 0) + agg.duration);
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

      const wdCats = new Map<string, number>();
      let wdDays = 0;
      const weCats = new Map<string, number>();
      let weDays = 0;

      for (let d = 0; d < 7; d++) {
        const date = new Date(monday);
        date.setDate(monday.getDate() + d);
        if (date > today) continue;

        const dateId = toDateId(date);
        const catMap = dateMap.get(dateId);
        const dow = date.getDay();

        if (dow === 0 || dow === 6) {
          weDays++;
          if (catMap) {
            for (const [cat, dur] of catMap) {
              weCats.set(cat, (weCats.get(cat) || 0) + dur);
            }
          }
        } else {
          wdDays++;
          if (catMap) {
            for (const [cat, dur] of catMap) {
              wdCats.set(cat, (wdCats.get(cat) || 0) + dur);
            }
          }
        }
      }

      if (wdDays > 0) {
        const wdTotal = Array.from(wdCats.values()).reduce((s, v) => s + v, 0);
        totalWeekdaySeconds += wdTotal;
        totalWeekdayDays += wdDays;
        if (wdCats.size > 0) {
          for (const [category, totalSec] of wdCats) {
            bars.push({ label: `${weekLabel} WD`, value: Math.round(totalSec / (60 * wdDays)), category });
          }
        } else {
          bars.push({ label: `${weekLabel} WD`, value: 0 });
        }
      }

      if (weDays > 0) {
        const weTotal = Array.from(weCats.values()).reduce((s, v) => s + v, 0);
        totalWeekendSeconds += weTotal;
        totalWeekendDays += weDays;
        if (weCats.size > 0) {
          for (const [category, totalSec] of weCats) {
            bars.push({ label: `${weekLabel} WE`, value: Math.round(totalSec / (60 * weDays)), category });
          }
        } else {
          bars.push({ label: `${weekLabel} WE`, value: 0 });
        }
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
      const [dateAggs, siteAggs, catAggs] = await Promise.all([
        GetAggregations(["date", "category"], { start_date: startDate, end_date: endDate }),
        GetAggregations(["url", "exe_path"], { start_date: startDate, end_date: endDate }),
        GetAggregations(["category"], { start_date: startDate, end_date: endDate })
      ]);

      dateAggregations = dateAggs;
      siteAggregations = siteAggs;
      categoryAggregations = catAggs;
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

  let chartAvg = $derived.by(() => {
    if (weeklyResult.bars.length === 0) return 0;
    const totals = new Map<string, number>();
    for (const bar of weeklyResult.bars) {
      totals.set(bar.label, (totals.get(bar.label) || 0) + bar.value);
    }
    const vals = Array.from(totals.values());
    return vals.reduce((s, v) => s + v, 0) / vals.length;
  });
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
  <TopListSection heading="Top (Past 4 Weeks)" {siteAggregations} {categoryAggregations} />
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
