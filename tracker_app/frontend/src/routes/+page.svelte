<script lang="ts">
  import BarChart from "$lib/BarChart.svelte";
  import TopSitesList from "$lib/TopSitesList.svelte";
  import { GetAggregations } from "../../wailsjs/go/main/App.js";
  import { onMount } from "svelte";

  // Apply theme class to document
  $effect(() => {
    if (isDarkMode) {
      document.documentElement.classList.add('dark-mode');
    } else {
      document.documentElement.classList.remove('dark-mode');
    }
  });

  // Handle panel resize dragging
  function handleMouseDown(e: MouseEvent) {
    isDragging = true;
    e.preventDefault();
  }

  function handleMouseMove(e: MouseEvent) {
    if (!isDragging) return;
    const newWidth = Math.max(200, Math.min(600, e.clientX));
    leftPanelWidth = newWidth;
  }

  function handleMouseUp() {
    isDragging = false;
  }

  $effect(() => {
    if (isDragging) {
      document.addEventListener('mousemove', handleMouseMove);
      document.addEventListener('mouseup', handleMouseUp);
      return () => {
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
      };
    }
  });

  type DataPoint = {
    label: string;
    value: number;
  };

  type TopSite = {
    name: string;
    duration: number;
  };

  let weekData: DataPoint[] = $state([]);
  let topSites: TopSite[] = $state([]);
  let isLoading = $state(true);
  let daysElapsed = $state(7); // Number of days that have occurred in the current week
  let activeView = $state("Daily"); // Track active navigation item
  let isDarkMode = $state(false); // Track theme mode
  let weekOverWeekChange = $state<number | null>(null); // Percentage change from last week
  let leftPanelWidth = $state(325); // Width of left panel in pixels
  let isDragging = $state(false);

  onMount(async () => {
    await fetchWeekData();
  });

  async function fetchWeekData() {
    const today = new Date();
    const dayOfWeek = today.getDay(); // 0 = Sunday, 1 = Monday, etc.

    // Calculate days since Monday
    const daysSinceMonday = dayOfWeek === 0 ? 6 : dayOfWeek - 1;
    // Calculate number of days elapsed (including today)
    daysElapsed = daysSinceMonday + 1;

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

    // Calculate last week's date range
    const lastWeekMonday = new Date(monday);
    lastWeekMonday.setDate(monday.getDate() - 7);
    const lastWeekDates: Date[] = [];
    for (let i = 0; i < 7; i++) {
      const date = new Date(lastWeekMonday);
      date.setDate(lastWeekMonday.getDate() + i);
      lastWeekDates.push(date);
    }
    const lastWeekDateIds = lastWeekDates.map(date => {
      const year = date.getFullYear();
      const month = date.getMonth() + 1;
      const day = date.getDate();
      return year * 10000 + month * 100 + day;
    });
    const lastWeekStartDate = lastWeekDateIds[0].toString();
    const lastWeekEndDate = lastWeekDateIds[6].toString();

    try {
      // Fetch data for both this week and last week
      const [dateAggregations, detailedAggregations, lastWeekAggregations] = await Promise.all([
        GetAggregations(["date"], { start_date: startDate, end_date: endDate }),
        GetAggregations(["url", "name"], { start_date: startDate, end_date: endDate }),
        GetAggregations(["date"], { start_date: lastWeekStartDate, end_date: lastWeekEndDate })
      ]);

      // Process bar chart data (by date)
      const durationMap = new Map<number, number>();
      for (const agg of dateAggregations) {
        const dateId = agg.groupers.date as number;
        durationMap.set(dateId, agg.duration);
      }

      const dayLabels = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
      weekData = dateIds.map((dateId, index) => {
        const durationSeconds = durationMap.get(dateId) || 0;
        const durationMinutes = Math.round(durationSeconds / 60);

        return {
          label: dayLabels[index],
          value: durationMinutes
        };
      });

      // Process top sites data (by url/name)
      const siteMap = new Map<string, number>();
      for (const agg of detailedAggregations) {
        const url = agg.groupers.url as string;
        const name = agg.groupers.name as string;
        const identifier = url && url.trim() !== "" ? url : name;

        if (identifier) {
          const currentDuration = siteMap.get(identifier) || 0;
          siteMap.set(identifier, currentDuration + agg.duration);
        }
      }

      topSites = Array.from(siteMap.entries())
        .map(([name, duration]) => ({ name, duration }))
        .sort((a, b) => b.duration - a.duration)
        .slice(0, 10);

      // Calculate week-over-week change
      const lastWeekDurationMap = new Map<number, number>();
      for (const agg of lastWeekAggregations) {
        const dateId = agg.groupers.date as number;
        lastWeekDurationMap.set(dateId, agg.duration);
      }

      // Calculate last week's total and average (in minutes)
      const lastWeekTotalSeconds = lastWeekDateIds.reduce((sum, dateId) => {
        return sum + (lastWeekDurationMap.get(dateId) || 0);
      }, 0);
      const lastWeekAvgMinutes = lastWeekTotalSeconds / (60 * 7); // 7 days in last week

      // Calculate this week's average (in minutes)
      const thisWeekTotalMinutes = weekData.reduce((sum, d) => sum + d.value, 0);
      const thisWeekAvgMinutes = thisWeekTotalMinutes / daysElapsed;

      // Calculate percentage change
      if (lastWeekAvgMinutes > 0) {
        weekOverWeekChange = ((thisWeekAvgMinutes - lastWeekAvgMinutes) / lastWeekAvgMinutes) * 100;
      } else if (thisWeekAvgMinutes > 0) {
        // When coming from zero, treat as infinite increase (show as "new")
        weekOverWeekChange = Infinity;
      } else {
        // Both weeks are zero
        weekOverWeekChange = null;
      }

    } catch (error) {
      console.error("Failed to fetch aggregations:", error);
      // Fall back to empty data
      weekData = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"].map(label => ({
        label,
        value: 0
      }));
      topSites = [];
    } finally {
      isLoading = false;
    }
  }

  let dailyAvg = $derived(
    weekData.length > 0 && daysElapsed > 0
      ? weekData.reduce((sum, d) => sum + d.value, 0) / daysElapsed
      : 0
  );

  let dailyAvgFormatted = $derived(() => {
    const hours = Math.floor(dailyAvg / 60);
    const minutes = Math.round(dailyAvg % 60);
    return { hours, minutes };
  });
</script>

<div class="container" style="grid-template-columns: {leftPanelWidth}px 4px 1fr;">
  <div class="left-panel">
    <!-- Logo -->
    <div class="logo">
      <div class="logo-dot"></div>
      Truth
    </div>

    <!-- Navigation -->
    <nav class="nav-section">
      <div class="nav-section-label">Aggregations</div>
      <div
        class="nav-item"
        class:active={activeView === "Daily"}
        onclick={() => activeView = "Daily"}
      >
        <span class="nav-icon">◉</span>
        Daily
      </div>
    </nav>

    <nav class="nav-section">
      <div class="nav-section-label">Configurations</div>
      <!-- Empty for now -->
    </nav>

    <!-- Theme toggle at bottom -->
    <div class="theme-toggle" onclick={() => isDarkMode = !isDarkMode}>
      <span class="theme-icon">{isDarkMode ? '☀' : '🌙'}</span>
      <span class="theme-label">{isDarkMode ? 'Light Mode' : 'Dark Mode'}</span>
    </div>
  </div>

  <!-- Draggable divider -->
  <div class="panel-divider" onmousedown={handleMouseDown}></div>

  <div class="right-panel">
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
        {:else if weekData.length > 0}
          <BarChart data={weekData} weeklyAvg={dailyAvg} />
        {:else}
          <p class="no-data">No data available</p>
        {/if}
      </div>
    </div>

    {#if !isLoading && topSites.length > 0}
      <div class="section-wrapper">
        <h2 class="section-heading">Top Sites & Apps This Week</h2>
        <div class="content-card">
          <TopSitesList sites={topSites} />
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  :global(html) {
    /* Light mode (default) */
    --bg-primary: #ffffff;
    --bg-secondary: #fafafa;
    --bg-tertiary: #f9fafb;
    --border-color: #e5e7eb;
    --text-primary: #111827;
    --text-secondary: #6b7280;
    --text-tertiary: #9ca3af;
    --accent-color: #3b82f6;
    --accent-bg: rgba(59, 130, 246, 0.1);
    --hover-bg: #f3f4f6;
    --nav-hover: #f3f4f6;
    --chart-grid: #9ca3af;
    --avg-line-color: #f97316;
    --card-bg: #f5f5f5;
    --card-border: #e0e0e0;
    --card-shadow: rgba(0, 0, 0, 0.05);

    --font-size-heading: 1.25rem;
    --font-size-small: 12px;
  }

  :global(html.dark-mode) {
    /* Dark mode */
    --bg-primary: #171717;
    --bg-secondary: #0F0F0F;
    --bg-tertiary: #1E1E1E;
    --border-color: #2A2A2A;
    --text-primary: #ECECEC;
    --text-secondary: #888888;
    --text-tertiary: #555555;
    --accent-color: #3b82f6;
    --accent-bg: rgba(59, 130, 246, 0.12);
    --hover-bg: #262626;
    --nav-hover: #1E1E1E;
    --chart-grid: #555555;
    --avg-line-color: #f97316;
    --card-bg: #1E1E1E;
    --card-border: #2A2A2A;
    --card-shadow: rgba(0, 0, 0, 0.3);
  }

  :global(html, body) {
    margin: 0;
    padding: 0;
    height: 100%;
    overflow: hidden;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen, Ubuntu, sans-serif;
    background: var(--bg-tertiary);
    color: var(--text-primary);
    transition: background-color 0.3s ease, color 0.3s ease;
  }

  .container {
    display: grid;
    grid-template-columns: 325px 4px 1fr;  /* Left panel, divider, right panel */
    height: 100vh;
    overflow: hidden;
  }

  .panel-divider {
    background: var(--border-color);
    cursor: col-resize;
    transition: background-color 0.2s ease;
    position: relative;
  }

  .panel-divider:hover {
    background: var(--accent-color);
  }

  .panel-divider:active {
    background: var(--accent-color);
  }

  .left-panel {
    padding: 28px 20px;
    background: var(--bg-secondary);
    border-right: 1px solid var(--border-color);
    display: flex;
    flex-direction: column;
    gap: 32px;
    transition: background-color 0.3s ease, border-color 0.3s ease;
  }

  /* Navigation Logo */
  .logo {
    font-family: serif;
    font-size: 22px;
    letter-spacing: -0.02em;
    color: var(--text-primary);
    display: flex;
    align-items: center;
    gap: 10px;
    font-weight: 400;
  }

  .logo-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--accent-color);
    box-shadow: 0 0 12px rgba(59, 130, 246, 0.4);
  }

  /* Navigation Section */
  .nav-section {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .nav-section-label {
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--text-tertiary);
    margin-bottom: 8px;
    padding: 0 12px;
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 9px 12px;
    border-radius: 10px;
    font-size: 13.5px;
    font-weight: 400;
    color: var(--text-secondary);
    cursor: pointer;
    transition: all 0.2s ease;
    user-select: none;
  }

  .nav-item:hover {
    background: var(--nav-hover);
    color: var(--text-primary);
  }

  .nav-item.active {
    background: var(--accent-bg);
    color: var(--accent-color);
    font-weight: 500;
  }

  .nav-icon {
    width: 18px;
    height: 18px;
    display: flex;
    align-items: center;
    justify-content: center;
    opacity: 0.5;
    font-size: 14px;
  }

  .nav-item.active .nav-icon {
    opacity: 1;
  }

  /* Theme Toggle */
  .theme-toggle {
    margin-top: auto;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 9px 12px;
    border-radius: 10px;
    font-size: 13.5px;
    color: var(--text-secondary);
    cursor: pointer;
    transition: all 0.2s ease;
    user-select: none;
  }

  .theme-toggle:hover {
    background: var(--nav-hover);
    color: var(--text-primary);
  }

  .theme-icon {
    width: 18px;
    height: 18px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 14px;
  }

  .right-panel {
    padding: 2rem;
    background: var(--bg-primary);
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    overflow-y: auto;
    transition: background-color 0.3s ease;
  }

  .content-card {
    background: var(--card-bg);
    border: 1px solid var(--card-border);
    border-radius: 16px;
    padding: 2rem;
    box-shadow:
      0 1px 3px var(--card-shadow),
      inset 0 1px 0 rgba(255, 255, 255, 0.05);
    transition: all 0.3s ease;
  }

  h2 {
    margin: 0 0 1rem 0;
    font-size: var(--font-size-heading);
    font-weight: 600;
    color: var(--text-primary);
    flex-shrink: 0;
    text-align: center;
  }

  .chart-header {
    margin-bottom: 0;
    flex-shrink: 0;
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
  }

  .chart-header-left {
    display: flex;
    flex-direction: column;
  }

  .chart-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--text-secondary);
    margin-bottom: 0.25rem;
    text-align: left;
  }

  .chart-value {
    font-size: 2rem;
    font-weight: 600;
    color: var(--text-primary);
    text-align: left;
    line-height: 1.2;
  }

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

  .section-wrapper {
    display: flex;
    flex-direction: column;
  }

  .section-heading {
    margin: 0 0 0.3rem 0;
    font-size: 1.5rem;
    font-weight: 600;
    color: var(--text-primary);
    text-align: left;
  }

  .chart-wrapper {
    height: 400px;
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
  }

  .loading,
  .no-data {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-secondary);
    font-size: 0.875rem;
  }
</style>
