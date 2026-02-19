<script lang="ts">
  import { page } from '$app/stores';

  let { children } = $props();

  let isDarkMode = $state(false);
  let leftPanelWidth = $state(325);
  let isDragging = $state(false);

  $effect(() => {
    if (isDarkMode) {
      document.documentElement.classList.add('dark-mode');
    } else {
      document.documentElement.classList.remove('dark-mode');
    }
  });

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
      <a
        class="nav-item"
        class:active={$page.url.pathname === '/daily' || $page.url.pathname === '/'}
        href="/daily"
      >
        <span class="nav-icon">&#9673;</span>
        Daily
      </a>
      <a
        class="nav-item"
        class:active={$page.url.pathname === '/weekly'}
        href="/weekly"
      >
        <span class="nav-icon">&#9673;</span>
        Weekly
      </a>
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
    {@render children()}
  </div>
</div>

<style>
  :global(html) {
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
    grid-template-columns: 325px 4px 1fr;
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

  /* Logo */
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

  /* Navigation */
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
    text-decoration: none;
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

  /* Shared content styles for route pages */
  .right-panel :global(.content-card) {
    background: var(--card-bg);
    border: 1px solid var(--card-border);
    border-radius: 16px;
    padding: 2rem;
    box-shadow:
      0 1px 3px var(--card-shadow),
      inset 0 1px 0 rgba(255, 255, 255, 0.05);
    transition: all 0.3s ease;
  }

  .right-panel :global(h2) {
    margin: 0 0 1rem 0;
    font-size: var(--font-size-heading);
    font-weight: 600;
    color: var(--text-primary);
    flex-shrink: 0;
    text-align: center;
  }

  .right-panel :global(.chart-header) {
    margin-bottom: 0;
    flex-shrink: 0;
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
  }

  .right-panel :global(.chart-header-left) {
    display: flex;
    flex-direction: column;
  }

  .right-panel :global(.chart-label) {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--text-secondary);
    margin-bottom: 0.25rem;
    text-align: left;
  }

  .right-panel :global(.chart-value) {
    font-size: 2rem;
    font-weight: 600;
    color: var(--text-primary);
    text-align: left;
    line-height: 1.2;
  }

  .right-panel :global(.section-wrapper) {
    display: flex;
    flex-direction: column;
  }

  .right-panel :global(.section-heading) {
    margin: 0 0 0.3rem 0;
    font-size: 1.5rem;
    font-weight: 600;
    color: var(--text-primary);
    text-align: left;
  }

  .right-panel :global(.chart-wrapper) {
    height: 400px;
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
  }

  .right-panel :global(.loading),
  .right-panel :global(.no-data) {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-secondary);
    font-size: 0.875rem;
  }
</style>
