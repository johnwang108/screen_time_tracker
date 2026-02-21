<script lang="ts">
  import { page } from '$app/stores';
  import BarChart from "$lib/BarChart.svelte";
  import TopListSection from "$lib/TopListSection.svelte";
  import { GetAggregations } from "../../../wailsjs/go/main/App.js";
  import type { Aggregation, DataPoint } from "$lib/utils";

  let mode = $derived($page.params.mode as 'daily' | 'weekly');

  let dateAggregations: Aggregation[] = $state([]);
  let siteAggregations: Aggregation[] = $state([]);
  let categoryAggregations: Aggregation[] = $state([]);
  let isLoading = $state(true);

  // Daily-only state
  let daysElapsed = $state(7);
  let weekOverWeekChange = $state<number | null>(null);

  function toDateId(date: Date): number {
    return date.getFullYear() * 10000 + (date.getMonth() + 1) * 100 + date.getDate();
  }

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
      dateIds.push(toDateId(date));
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

    // 4 weeks, oldest first
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

  function transformWeekly(aggregations: Aggregation[]): DataPoint[] {
    return buildWeeklyBars(aggregations).bars;
  }

  async function fetchData() {
    isLoading = true;
    dateAggregations = [];
    siteAggregations = [];
    categoryAggregations = [];
    weekOverWeekChange = null;

    const today = new Date();
    const dayOfWeek = today.getDay();
    const daysSinceMonday = dayOfWeek === 0 ? 6 : dayOfWeek - 1;
    const currentMonday = new Date(today);
    currentMonday.setDate(today.getDate() - daysSinceMonday);

    try {
      if (mode === 'daily') {
        daysElapsed = daysSinceMonday + 1;

        const weekDates: Date[] = [];
        for (let i = 0; i < 7; i++) {
          const date = new Date(currentMonday);
          date.setDate(currentMonday.getDate() + i);
          weekDates.push(date);
        }
        const dateIds = weekDates.map(toDateId);
        const startDate = dateIds[0].toString();
        const endDate = dateIds[6].toString();

        const lastWeekMonday = new Date(currentMonday);
        lastWeekMonday.setDate(currentMonday.getDate() - 7);
        const lastWeekStart = toDateId(lastWeekMonday).toString();
        const lastWeekEnd = toDateId(new Date(lastWeekMonday.getTime() + 6 * 86400000)).toString();

        const [dateAggs, siteAggs, catAggs, lastWeekAggs] = await Promise.all([
          GetAggregations(["date", "category"], { start_date: startDate, end_date: endDate }),
          GetAggregations(["url", "exe_path"], { start_date: startDate, end_date: endDate }),
          GetAggregations(["category"], { start_date: startDate, end_date: endDate }),
          GetAggregations(["date"], { start_date: lastWeekStart, end_date: lastWeekEnd })
        ]);

        dateAggregations = dateAggs;
        siteAggregations = siteAggs;
        categoryAggregations = catAggs;

        const lastWeekAvgMinutes = lastWeekAggs.reduce((s: number, a: Aggregation) => s + a.duration, 0) / (60 * 7);
        const thisWeekAvgMinutes = dateAggs.reduce((s: number, a: Aggregation) => s + a.duration, 0) / (60 * daysElapsed);

        if (lastWeekAvgMinutes > 0) {
          weekOverWeekChange = ((thisWeekAvgMinutes - lastWeekAvgMinutes) / lastWeekAvgMinutes) * 100;
        } else if (thisWeekAvgMinutes > 0) {
          weekOverWeekChange = Infinity;
        }
      } else {
        const startMonday = new Date(currentMonday);
        startMonday.setDate(currentMonday.getDate() - 21);
        const startDate = toDateId(startMonday).toString();
        const endDate = toDateId(today).toString();

        const [dateAggs, siteAggs, catAggs] = await Promise.all([
          GetAggregations(["date", "category"], { start_date: startDate, end_date: endDate }),
          GetAggregations(["url", "exe_path"], { start_date: startDate, end_date: endDate }),
          GetAggregations(["category"], { start_date: startDate, end_date: endDate })
        ]);

        dateAggregations = dateAggs;
        siteAggregations = siteAggs;
        categoryAggregations = catAggs;
      }
    } catch (error) {
      console.error("Failed to fetch aggregations:", error);
    } finally {
      isLoading = false;
    }
  }

  $effect(() => {
    // re-runs whenever mode changes
    mode;
    fetchData();
  });

  // Daily derived
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

  // Weekly derived
  let weeklyResult = $derived.by(() => buildWeeklyBars(dateAggregations));

  let weekdayAvgFormatted = $derived(() => {
    const avg = weeklyResult.weekdayAvg;
    return { hours: Math.floor(avg / 60), minutes: Math.round(avg % 60) };
  });

  let weekendAvgFormatted = $derived(() => {
    const avg = weeklyResult.weekendAvg;
    return { hours: Math.floor(avg / 60), minutes: Math.round(avg % 60) };
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
      {#if mode === 'daily'}
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
      {:else}
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
      {/if}
    </div>
    {#if isLoading}
      <p class="loading">Loading...</p>
    {:else if dateAggregations.length > 0}
      <BarChart
        aggregations={dateAggregations}
        transform={mode === 'daily' ? transformDaily : transformWeekly}
        weeklyAvg={mode === 'daily' ? dailyAvg : chartAvg}
      />
    {:else}
      <p class="no-data">No data available</p>
    {/if}
  </div>
</div>

{#if !isLoading && siteAggregations.length > 0}
  <TopListSection
    heading={mode === 'daily' ? 'Top This Week' : 'Top (Past 4 Weeks)'}
    {siteAggregations}
    {categoryAggregations}
  />
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

  .weekly-averages {
    display: flex;
    gap: 2.5rem;
  }

  .avg-group {
    display: flex;
    flex-direction: column;
  }
</style>
