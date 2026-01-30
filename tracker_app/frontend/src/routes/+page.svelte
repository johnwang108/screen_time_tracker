<script lang="ts">
  import BarChart from "$lib/BarChart.svelte";
  import { GetAggregations } from "../../wailsjs/go/main/App.js";
  import { onMount } from "svelte";

  type DataPoint = {
    label: string;
    value: number;
    average: number;
  };

  let weekData: DataPoint[] = $state([]);
  let isLoading = $state(true);

  onMount(async () => {
    await fetchWeekData();
  });

  async function fetchWeekData() {
    const today = new Date();
    const dayOfWeek = today.getDay(); // 0 = Sunday, 1 = Monday, etc.

    // Calculate days since Monday
    const daysSinceMonday = dayOfWeek === 0 ? 6 : dayOfWeek - 1;

    // Get Monday of this week
    const monday = new Date(today);
    monday.setDate(today.getDate() - daysSinceMonday);

    // Create array of dates for this week (Mon-Sun)
    const weekDates: Date[] = [];
    for (let i = 0; i < 7; i++) {
      const date = new Date(monday);
      date.setDate(monday.getDate() + i);
      weekDates.push(date);
    }

    // Convert dates to date_ids (YYYYMMDD format)
    const dateIds = weekDates.map(date => {
      const year = date.getFullYear();
      const month = date.getMonth() + 1;
      const day = date.getDate();
      return year * 10000 + month * 100 + day;
    });

    // Fetch aggregations for the week
    const startDate = dateIds[0].toString();
    const endDate = dateIds[6].toString();

    try {
      const aggregations = await GetAggregations(
        ["date"],
        {
          start_date: startDate,
          end_date: endDate
        }
      );

      // Create a map of date_id to duration
      const durationMap = new Map<number, number>();
      for (const agg of aggregations) {
        const dateId = agg.groupers.date as number;
        durationMap.set(dateId, agg.duration);
      }

      // Build weekData array
      const dayLabels = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
      weekData = dateIds.map((dateId, index) => {
        const durationSeconds = durationMap.get(dateId) || 0;
        const durationMinutes = Math.round(durationSeconds / 60);

        return {
          label: dayLabels[index],
          value: durationMinutes,
          average: 0 // Ignoring average for now
        };
      });

    } catch (error) {
      console.error("Failed to fetch aggregations:", error);
      // Fall back to empty data
      weekData = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"].map(label => ({
        label,
        value: 0,
        average: 0
      }));
    } finally {
      isLoading = false;
    }
  }

  let dailyAvg = $derived(
    weekData.length > 0
      ? weekData.reduce((sum, d) => sum + d.value, 0) / weekData.length
      : 0
  );

  let dailyAvgHours = $derived((dailyAvg / 60).toFixed(1));
</script>

<div class="container">
  <div class="left-panel">
    <div class="chart-wrapper">
      <h2>Daily Average: {dailyAvgHours}h</h2>
      {#if isLoading}
        <p class="loading">Loading...</p>
      {:else if weekData.length > 0}
        <BarChart data={weekData} weeklyAvg={dailyAvg} />
      {:else}
        <p class="no-data">No data available</p>
      {/if}
    </div>
  </div>
  <div class="right-panel">
    <!-- Reserved for future content -->
  </div>
</div>

<style>
  :global(html, body) {
    margin: 0;
    padding: 0;
    height: 100%;
    overflow: hidden;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen, Ubuntu, sans-serif;
    background: #f9fafb;

    --font-size-heading: 1.25rem;
    --font-size-small: 12px;
  }

  .container {
    display: grid;
    grid-template-columns: 1fr 1fr;
    height: 100vh;
    overflow: hidden;
  }

  .left-panel {
    padding: 2rem;
    background: #ffffff;
    border-right: 1px solid #e5e7eb;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .right-panel {
    padding: 2rem;
    background: #f9fafb;
  }

  h2 {
    margin: 0 0 1rem 0;
    font-size: var(--font-size-heading);
    font-weight: 600;
    color: #111827;
    flex-shrink: 0;
    text-align: center;
  }

  .chart-wrapper {
    height: 400px;
    display: flex;
    flex-direction: column;
  }

  .loading,
  .no-data {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #6b7280;
    font-size: 0.875rem;
  }
</style>
