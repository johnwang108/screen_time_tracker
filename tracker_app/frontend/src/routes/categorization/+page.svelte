<script lang="ts">
  import { GetAggregations, GetCategories, SetItemCategory, CreateCategory, ReorderCategories } from "../../../wailsjs/go/main/App.js";
  import { onMount, tick } from "svelte";
  import type { Aggregation } from "$lib/utils";
  import { formatDuration } from "$lib/utils";

  type CategoryItems = {
    sites: string[];
    apps: string[];
  };

  type UncategorizedItem = {
    identifier: string;
    isApp: boolean;
    duration: number;
    displayName: string;
  };

  let categoriesMap: Record<string, CategoryItems> = $state({});
  let categoryOrder: string[] = $state([]);
  let allItemAggregations: Aggregation[] = $state([]);
  let isLoading = $state(true);

  // Item drag-and-drop state
  let dragOverCategory: string | null = $state(null);
  let dragOverUncategorized = $state(false);

  // Category reorder drag state
  let reorderDragIndex: number | null = $state(null);
  let reorderOverIndex: number | null = $state(null);

  // New category state
  let newCategoryName = $state("");
  let validationError = $state("");
  let categoryInputEl: HTMLInputElement | undefined = $state(undefined);
  let inputFocused = $state(false);
  let hideShort = $state(false);

  // Panel resize state
  let rightPanelWidth = $state(38);  // percentage
  let isResizing = $state(false);
  let containerEl: HTMLDivElement | undefined = $state(undefined);

  function handleResizeStart(e: MouseEvent) {
    isResizing = true;
    e.preventDefault();
  }

  function handleResizeMove(e: MouseEvent) {
    if (!isResizing || !containerEl) return;
    const rect = containerEl.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const pct = 100 - (x / rect.width) * 100;
    rightPanelWidth = Math.max(20, Math.min(60, pct));
  }

  function handleResizeEnd() {
    isResizing = false;
  }

  $effect(() => {
    if (isResizing) {
      document.addEventListener("mousemove", handleResizeMove);
      document.addEventListener("mouseup", handleResizeEnd);
      return () => {
        document.removeEventListener("mousemove", handleResizeMove);
        document.removeEventListener("mouseup", handleResizeEnd);
      };
    }
  });

  onMount(async () => {
    await fetchData();
  });

  async function fetchData() {
    try {
      const [catResponse, allItemAggs] = await Promise.all([
        GetCategories(),
        GetAggregations(["url", "exe_path", "category"], {})
      ]);
      categoriesMap = catResponse.categories;
      categoryOrder = catResponse.order;
      allItemAggregations = allItemAggs;
    } catch (error) {
      console.error("Failed to fetch categorization data:", error);
    } finally {
      isLoading = false;
    }
  }

  let durationMap = $derived.by(() => {
    const map = new Map<string, number>();
    for (const agg of allItemAggregations) {
      const key = (agg.groupers.url as string) || (agg.groupers.exe_path as string);
      if (key) map.set(key, (map.get(key) || 0) + agg.duration);
    }
    return map;
  });

  let uncategorizedItems: UncategorizedItem[] = $derived.by(() => {
    const itemMap = new Map<string, UncategorizedItem>();
    for (const agg of allItemAggregations) {
      const category = agg.groupers.category as string;
      if (category && category !== "Other") continue;

      const url = agg.groupers.url as string;
      const exePath = agg.groupers.exe_path as string;

      if (url) {
        const existing = itemMap.get(url);
        if (existing) {
          existing.duration += agg.duration;
        } else {
          itemMap.set(url, {
            identifier: url,
            isApp: false,
            duration: agg.duration,
            displayName: extractDisplayName(url, false)
          });
        }
      } else if (exePath) {
        const existing = itemMap.get(exePath);
        if (existing) {
          existing.duration += agg.duration;
        } else {
          itemMap.set(exePath, {
            identifier: exePath,
            isApp: true,
            duration: agg.duration,
            displayName: extractDisplayName(exePath, true)
          });
        }
      }
    }

    return Array.from(itemMap.values()).sort((a, b) => b.duration - a.duration);
  });

  function extractDisplayName(identifier: string, isApp: boolean): string {
    if (isApp) {
      const parts = identifier.replace(/\\/g, "/").split("/");
      const filename = parts[parts.length - 1];
      return filename.replace(/\.exe$/i, "");
    }
    return identifier;
  }

  function validateCategoryName(name: string): string | null {
    const trimmed = name.trim();
    if (!trimmed) return "Category name cannot be empty";
    if (trimmed.length > 30) return "Name must be 30 characters or less";
    if (categoriesMap[trimmed]) return `"${trimmed}" already exists`;
    return null;
  }

  function handleInput() {
    if (validationError) validationError = "";
  }

  // --- Item drag-and-drop ---

  function handleItemDragStart(e: DragEvent, identifier: string, isApp: boolean) {
    e.dataTransfer?.setData("application/json", JSON.stringify({ type: "item", identifier, isApp }));
  }

  async function handleItemDrop(e: DragEvent, targetCategory: string) {
    e.preventDefault();
    dragOverCategory = null;
    const data = e.dataTransfer?.getData("application/json");
    if (!data) return;
    const parsed = JSON.parse(data);
    if (parsed.type !== "item") return;

    try {
      await SetItemCategory(parsed.identifier, targetCategory, parsed.isApp);
      await fetchData();
    } catch (error) {
      console.error("Failed to set item category:", error);
    }
  }

  async function handleDropUncategorize(e: DragEvent) {
    e.preventDefault();
    dragOverUncategorized = false;
    const data = e.dataTransfer?.getData("application/json");
    if (!data) return;
    const parsed = JSON.parse(data);
    if (parsed.type !== "item") return;

    try {
      await SetItemCategory(parsed.identifier, "", parsed.isApp);
      await fetchData();
    } catch (error) {
      console.error("Failed to uncategorize item:", error);
    }
  }

  function handleCategoryDragOver(e: DragEvent, categoryName: string) {
    e.preventDefault();
    // Only highlight for item drops, not category reorders
    const data = e.dataTransfer?.types;
    if (data?.includes("application/json")) {
      dragOverCategory = categoryName;
    }
  }

  // --- Category reorder drag-and-drop ---

  function handleReorderDragStart(e: DragEvent, index: number) {
    reorderDragIndex = index;
    e.dataTransfer?.setData("text/plain", index.toString());
    e.dataTransfer!.effectAllowed = "move";
  }

  function handleReorderDragOver(e: DragEvent, index: number) {
    if (reorderDragIndex === null) return;
    e.preventDefault();
    e.dataTransfer!.dropEffect = "move";
    reorderOverIndex = index;
  }

  async function handleReorderDrop(e: DragEvent, dropIndex: number) {
    e.preventDefault();
    if (reorderDragIndex === null || reorderDragIndex === dropIndex) {
      reorderDragIndex = null;
      reorderOverIndex = null;
      return;
    }

    // Reorder the array
    const newOrder = [...categoryOrder];
    const [moved] = newOrder.splice(reorderDragIndex, 1);
    newOrder.splice(dropIndex, 0, moved);

    categoryOrder = newOrder;
    reorderDragIndex = null;
    reorderOverIndex = null;

    try {
      await ReorderCategories(newOrder);
    } catch (error) {
      console.error("Failed to save category order:", error);
      await fetchData();
    }
  }

  function handleReorderDragEnd() {
    reorderDragIndex = null;
    reorderOverIndex = null;
  }

  // --- New category ---

  async function handleCreateCategory() {
    const error = validateCategoryName(newCategoryName);
    if (error) {
      validationError = error;
      return;
    }

    try {
      await CreateCategory(newCategoryName.trim());
      newCategoryName = "";
      validationError = "";
      await fetchData();
    } catch (error) {
      validationError = "Failed to create category";
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Enter") {
      handleCreateCategory();
    } else if (e.key === "Escape") {
      newCategoryName = "";
      validationError = "";
      categoryInputEl?.blur();
    }
  }
</script>

{#if isLoading}
  <p class="loading">Loading...</p>
{:else}
  <div class="categorization-layout" bind:this={containerEl} style="grid-template-columns: 1fr 4px {rightPanelWidth}%;">
    <!-- Left panel: Category groups -->
    <div class="categories-panel">
      <div class="categories-panel-header">
        <h2 class="panel-heading">Categories</h2>
        <button class="toggle-short" class:active={hideShort} onclick={() => hideShort = !hideShort}>
          &lt;1h
        </button>
      </div>

      {#each categoryOrder as categoryName, index}
        {@const items = categoriesMap[categoryName]}
        {#if items}
          {@const allItems = [
            ...items.sites.map(id => ({ id, isApp: false })),
            ...items.apps.map(id => ({ id, isApp: true }))
          ].sort((a, b) => (durationMap.get(b.id) ?? 0) - (durationMap.get(a.id) ?? 0))
            .filter(item => !hideShort || (durationMap.get(item.id) ?? 0) >= 3600)}
          <div
            class="category-group"
            class:drag-over={dragOverCategory === categoryName}
            class:reorder-over={reorderOverIndex === index && reorderDragIndex !== index}
            class:reorder-dragging={reorderDragIndex === index}
            ondragover={(e) => {
              handleCategoryDragOver(e, categoryName);
              handleReorderDragOver(e, index);
            }}
            ondragenter={() => dragOverCategory = categoryName}
            ondragleave={() => { if (dragOverCategory === categoryName) dragOverCategory = null; }}
            ondrop={(e) => {
              if (reorderDragIndex !== null) {
                handleReorderDrop(e, index);
              } else {
                handleItemDrop(e, categoryName);
              }
            }}
            role="region"
            aria-label="{categoryName} category"
          >
            <div
              class="category-header"
              draggable="true"
              ondragstart={(e) => handleReorderDragStart(e, index)}
              ondragend={handleReorderDragEnd}
            >
              <span class="drag-handle">&#8942;&#8942;</span>
              <span class="category-title">{categoryName}</span>
            </div>
            <div class="category-items">
              {#if allItems.length === 0}
                <div class="empty-hint">Drop items here</div>
              {/if}
              {#each allItems as item}
                <div
                  class="category-item"
                  draggable="true"
                  ondragstart={(e) => handleItemDragStart(e, item.id, item.isApp)}
                  role="button"
                  tabindex="0"
                >
                  <div class="item-icon" class:app={item.isApp} class:site={!item.isApp}>
                    <span class="icon-label">{formatDuration(durationMap.get(item.id) ?? 0, true)}</span>
                  </div>
                  <span class="item-name">{extractDisplayName(item.id, item.isApp)}</span>
                </div>
              {/each}
            </div>
          </div>
        {/if}
      {/each}
    </div>

    <!-- Resize divider -->
    <div class="panel-divider" onmousedown={handleResizeStart}></div>

    <!-- Right panel: New Category bar + Uncategorized list -->
    <div class="right-side">
      <!-- New Category bar -->
      <div class="new-category-bar" class:focused={inputFocused}>
        <div class="new-category-row">
          <input
            type="text"
            placeholder="New category..."
            bind:value={newCategoryName}
            bind:this={categoryInputEl}
            oninput={handleInput}
            onkeydown={handleKeydown}
            onfocus={() => inputFocused = true}
            onblur={() => inputFocused = false}
            class:input-error={validationError !== ""}
          />
          <button
            class="btn-create"
            onclick={handleCreateCategory}
            disabled={newCategoryName.trim() === ""}
          >
            Create
          </button>
        </div>
        {#if validationError}
          <span class="validation-message">{validationError}</span>
        {/if}
      </div>

      <!-- Uncategorized section -->
      <div
        class="uncategorized-section"
        class:drag-over={dragOverUncategorized}
        ondragover={(e) => e.preventDefault()}
        ondragenter={() => dragOverUncategorized = true}
        ondragleave={() => dragOverUncategorized = false}
        ondrop={handleDropUncategorize}
      >
        <div class="uncategorized-header">
          <h2 class="panel-heading">Uncategorized</h2>
          {#if uncategorizedItems.length > 0}
            <span class="item-count">{uncategorizedItems.length}</span>
          {/if}
        </div>

        <div class="category-items">
          {#each uncategorizedItems.filter(item => !hideShort || item.duration >= 3600) as item}
            <div
              class="category-item"
              draggable="true"
              ondragstart={(e) => handleItemDragStart(e, item.identifier, item.isApp)}
              role="button"
              tabindex="0"
            >
              <div class="item-icon" class:app={item.isApp} class:site={!item.isApp}>
                <span class="icon-label">{formatDuration(item.duration, true)}</span>
              </div>
              <span class="item-name">{item.displayName}</span>
            </div>
          {/each}
          {#if uncategorizedItems.length === 0}
            <div class="empty-hint centered">All items are categorized</div>
          {/if}
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  .categorization-layout {
    display: grid;
    height: 100%;
    min-height: 0;
  }

  .panel-divider {
    background: var(--border-color);
    cursor: col-resize;
    border-radius: 2px;
    transition: background-color 0.2s ease;
  }

  .panel-divider:hover,
  .panel-divider:active {
    background: var(--accent-color);
  }

  .panel-heading {
    margin: 0;
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--text-primary);
    text-align: left;
  }

  /* ── Categories panel ── */
  .categories-panel {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    overflow-y: auto;
  }

  .category-group {
    background: var(--card-bg);
    border: 2px solid var(--card-border);
    border-radius: 12px;
    padding: 1rem;
    transition: border-color 0.2s ease, background-color 0.2s ease, opacity 0.2s ease, transform 0.15s ease;
  }

  .category-group.drag-over {
    border-color: var(--accent-color);
    background: var(--accent-bg);
  }

  .category-group.reorder-over {
    border-color: var(--accent-color);
    box-shadow: 0 -2px 0 0 var(--accent-color);
  }

  .category-group.reorder-dragging {
    opacity: 0.4;
  }

  .category-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.75rem;
    cursor: grab;
    user-select: none;
  }

  .category-header:active {
    cursor: grabbing;
  }

  .drag-handle {
    font-size: 11px;
    color: var(--text-tertiary);
    letter-spacing: -2px;
    line-height: 1;
    opacity: 0.5;
    transition: opacity 0.15s ease;
  }

  .category-header:hover .drag-handle {
    opacity: 1;
  }

  .category-title {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .category-items {
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem;
  }

  .category-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.35rem;
    cursor: grab;
    padding: 0.4rem;
    border-radius: 8px;
    transition: background-color 0.15s ease;
    max-width: 80px;
  }

  .category-item:hover {
    background: var(--hover-bg);
  }

  .category-item:active {
    cursor: grabbing;
  }

  .item-icon {
    width: 44px;
    height: 44px;
    border-radius: 10px;
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .item-icon.app {
    background: rgba(59, 130, 246, 0.15);
    border-color: rgba(59, 130, 246, 0.3);
  }

  .item-icon.site {
    background: rgba(16, 185, 129, 0.15);
    border-color: rgba(16, 185, 129, 0.3);
  }

  .icon-label {
    font-size: 0.6rem;
    font-weight: 600;
    color: var(--text-tertiary);
  }

  .item-name {
    font-size: 11px;
    color: var(--text-secondary);
    text-align: center;
    word-break: break-all;
    line-height: 1.2;
  }

  .empty-hint {
    font-size: 0.8rem;
    color: var(--text-tertiary);
    font-style: italic;
    padding: 0.5rem 0;
  }

  .empty-hint.centered {
    text-align: center;
    padding: 2rem 0;
  }

  /* ── Right side ── */
  .right-side {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    min-height: 0;
  }

  /* ── New category bar ── */
  .new-category-bar {
    background: var(--card-bg);
    border: 1px solid var(--card-border);
    border-radius: 10px;
    padding: 0.6rem 0.75rem;
    flex-shrink: 0;
    transition: border-color 0.2s ease, box-shadow 0.2s ease;
  }

  .new-category-bar.focused {
    border-color: var(--accent-color);
    box-shadow: 0 0 0 2px var(--accent-bg);
  }

  .new-category-row {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .new-category-row input {
    flex: 1;
    font-size: 0.825rem;
    padding: 0.4rem 0;
    border: none;
    background: transparent;
    color: var(--text-primary);
    outline: none;
  }

  .new-category-row input::placeholder {
    color: var(--text-tertiary);
  }

  .new-category-row input.input-error {
    color: #ef4444;
  }

  .btn-create {
    font-size: 0.75rem;
    font-weight: 500;
    padding: 0.35rem 0.75rem;
    border: none;
    border-radius: 6px;
    background: var(--accent-color);
    color: white;
    cursor: pointer;
    transition: opacity 0.15s ease;
    white-space: nowrap;
  }

  .btn-create:hover:not(:disabled) {
    opacity: 0.85;
  }

  .btn-create:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  .validation-message {
    display: block;
    font-size: 0.7rem;
    color: #ef4444;
    margin-top: 0.35rem;
  }

  /* ── Uncategorized section ── */
  .uncategorized-section {
    flex: 1;
    min-height: 0;
    background: var(--card-bg);
    border: 1px solid var(--card-border);
    border-radius: 12px;
    padding: 1rem;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    transition: border-color 0.2s ease, background-color 0.2s ease;
  }

  .uncategorized-section.drag-over {
    border-color: var(--accent-color);
    background: var(--accent-bg);
  }

  .categories-panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .uncategorized-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-shrink: 0;
  }

  .toggle-short {
    font-size: 0.65rem;
    font-weight: 600;
    padding: 2px 7px;
    border-radius: 10px;
    border: 1px solid var(--border-color);
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .toggle-short:hover {
    color: var(--text-secondary);
    border-color: var(--text-secondary);
  }

  .toggle-short.active {
    background: var(--accent-bg);
    color: var(--accent-color);
    border-color: var(--accent-color);
  }

  .item-count {
    font-size: 0.65rem;
    font-weight: 700;
    background: var(--accent-bg);
    color: var(--accent-color);
    padding: 1px 7px;
    border-radius: 10px;
  }

</style>
