<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { Folder, FolderOpen, Eye, FileText, Cog, Play, Square, Info } from 'lucide-svelte';
  import type { Config } from '../../wailsjs/go/models';
  import type { PatternInfo } from '../../wailsjs/go/models';
  import type { Info as NamerInfo } from '../../wailsjs/go/models';
  import type { Info as ActionInfo } from '../../wailsjs/go/models';

  export let config: Config;
  export let watchPathDisplay: string;
  export let selectedPatternID: string;
  export let availablePatterns: PatternInfo[] = [];
  export let availableNamers: NamerInfo[] = [];
  export let availableActions: ActionInfo[] = [];
  export let isWatching: boolean = false;
  export let isBusy: boolean = false;

  const dispatch = createEventDispatcher<{
    selectDirectory: null;
    patternSelect: any;
    selectActionDirectory: null;
    requestStart: null;
    requestStop: null;
    openAdvancedOperations: null;
    openMonitoring: null;
  }>();

  let showPatternInfo = false;
  let showNamingInfo = false;
  let showActionInfo = false;
  const TEMPLATE_PLACEHOLDER = "{original}-{date}";

  $: statusTone = isWatching ? "active" : isBusy ? "busy" : "idle";
  $: statusMessage = isWatching
    ? "Watching for file changes..."
    : isBusy
    ? "Processing pending actions..."
    : "Watcher idle";

  function handleSelectDirectory() {
    dispatch('selectDirectory');
  }

  function handlePatternSelect(event) {
    dispatch('patternSelect', event);
  }

  function handleSelectActionDirectory() {
    dispatch('selectActionDirectory');
  }

  function handleStart() {
    dispatch('requestStart');
  }

  function handleStop() {
    dispatch('requestStop');
  }

  function handleOpenAdvancedOperations() {
    dispatch('openAdvancedOperations');
  }

  function handleOpenMonitoring() {
    dispatch('openMonitoring');
  }
</script>

<div class="config-section">
  <!-- Directory Selection Card -->
  <div class="config-card directory-selection">
    <div class="card-header">
      <h3><Folder size={16} />Watch Directory</h3>
      <p>Select or drag & drop a folder to monitor</p>
    </div>
    <div class="card-content">
      <div class="directory-input">
        <input
          id="watch-path"
          type="text"
          readonly
          bind:value={watchPathDisplay}
          placeholder="No directory selected..."
          class="directory-path"
        />
        <button
          class="browse-btn"
          on:click={handleSelectDirectory}
        >
          <FolderOpen size={16} />
          Browse
        </button>
        {#if !isWatching}
          <button
            class="start-btn"
            on:click={handleStart}
            disabled={isBusy}
          >
            <Play size={16} />
            Start Watching
          </button>
        {:else}
          <button
            class="stop-btn"
            on:click={handleStop}
            disabled={isBusy}
          >
            <Square size={16} />
            Stop Watching
          </button>
        {/if}
      </div>
      <div class="status-row">
        <span class={`status-indicator ${statusTone}`}>{statusMessage}</span>
        <button class="status-link" type="button" on:click={handleOpenMonitoring}>
          View monitoring
        </button>
      </div>
    </div>
  </div>

  <!-- Pattern & Naming Cards -->
  <div class="pattern-naming-grid">
    <div class="config-card compact">
      <div class="card-header">
        <h3><Eye size={16} />Pattern Matching</h3>
        <button 
          class="info-btn" 
          on:mouseenter={() => showPatternInfo = true}
          on:mouseleave={() => showPatternInfo = false}
          aria-label="Show additional information about pattern matching"
        >
          <Info size={14} />
        </button>
      </div>
      {#if showPatternInfo}
        <div class="info-tooltip">Define which files to rename using pattern matching</div>
      {/if}
      <div class="card-content">
        <div class="form-group compact">
          <select
            id="pattern-id"
            on:change={handlePatternSelect}
            bind:value={selectedPatternID}
          >
            {#if availablePatterns.length === 0}
              <option value="">Loading patterns...</option>
            {:else}
              {#each availablePatterns as pattern}
                <option value={pattern.id} title={pattern.description}
                  >{pattern.name}</option
                >
              {/each}
              <option value="custom">Custom Regex</option>
            {/if}
          </select>
        </div>
        {#if selectedPatternID === "custom"}
          <div class="form-group compact">
            <input
              id="name-pattern"
              type="text"
              bind:value={config.NamePattern}
              placeholder="Custom regex pattern"
            />
          </div>
        {/if}
      </div>
    </div>

    <div class="config-card compact">
      <div class="card-header">
        <h3><FileText size={16} />Naming Scheme</h3>
        <button 
          class="info-btn" 
          on:mouseenter={() => showNamingInfo = true}
          on:mouseleave={() => showNamingInfo = false}
          aria-label="Show additional information about naming schemes"
        >
          <Info size={14} />
        </button>
      </div>
      {#if showNamingInfo}
        <div class="info-tooltip">Choose how to rename matched files</div>
      {/if}
      <div class="card-content">
        <div class="form-group compact">
          <select
            id="namer-id"
            bind:value={config.NamerID}
          >
            {#if availableNamers.length === 0}
              <option value="">Loading naming methods...</option>
            {:else}
              {#each availableNamers as namer}
                <option value={namer.id} title={namer.description}
                  >{namer.name}</option
                >
              {/each}
            {/if}
          </select>
        </div>
        {#if config.NamerID === "random"}
          <div class="form-group compact">
            <input
              id="name-length"
              type="number"
              bind:value={config.RandomLength}
              placeholder="Random name length"
            />
          </div>
        {:else if config.NamerID === "template"}
          <div class="form-group compact">
            <input
              id="template-string"
              type="text"
              bind:value={config.TemplateString}
              placeholder={`e.g. ${TEMPLATE_PLACEHOLDER}`}
            />
            <div class="help-text compact">
              {'{'}original}, {'{'}date}, {'{'}time}, {'{'}count}
            </div>
          </div>
        {:else if config.NamerID === "custom_datetime" || config.NamerID === "sequential_datetime"}
          <div class="form-group compact">
            <input
              id="datetime-format"
              type="text"
              bind:value={config.DateTimeFormat}
              placeholder="2006-01-02_15-04-05"
            />
            <div class="help-text compact">
              Go format: 2006=year, 01=month, 02=day
            </div>
          </div>
        {/if}
      </div>
    </div>
  </div>

  <!-- Post-Rename Actions Card -->
  <div class="config-card compact full-width">
    <div class="card-header">
      <h3><Cog size={16} />Post-Rename Actions</h3>
      <button 
        class="info-btn" 
        on:mouseenter={() => showActionInfo = true}
        on:mouseleave={() => showActionInfo = false}
        aria-label="Show additional information about post-rename actions"
      >
        <Info size={14} />
      </button>
    </div>
    {#if showActionInfo}
      <div class="info-tooltip">What to do after renaming files</div>
    {/if}
    <div class="card-content">
      <div class="form-group compact">
        <select
          id="action-id"
          bind:value={config.ActionID}
        >
          {#if availableActions.length === 0}
            <option value="">Loading actions...</option>
          {:else}
            {#each availableActions as action}
              <option value={action.id} title={action.description}
                >{action.name}</option
              >
            {/each}
          {/if}
        </select>
      </div>
      {#if config.ActionID === "move" || config.ActionID === "copy"}
        <div class="form-group compact">
          <div class="directory-input">
            <input
              id="action-dest"
              type="text"
              readonly
              placeholder="Select destination..."
              value={config.ActionConfig?.destinationPath || ""}
              class="directory-path"
            />
            <button
              class="browse-btn"
              on:click={handleSelectActionDirectory}
            >
              <FolderOpen size={16} />
              Browse
            </button>
          </div>
        </div>
      {:else if config.ActionID === "advanced_operations"}
        <div class="action-info-card compact">
          <button
            class="secondary-btn"
            on:click={handleOpenAdvancedOperations}
            disabled={isWatching}
          >
            Manage Advanced Operations
          </button>
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .config-section {
    padding: 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 1rem;
    overflow-y: auto;
    height: 100%;
    box-sizing: border-box;
    width: 100%;
    max-width: 1100px;
    margin: 0 auto;
    min-height: 0;
  }

  .config-card {
    background: linear-gradient(
      135deg,
      var(--card-bg),
      rgba(22, 33, 62, 0.8) 100%
    );
    border: 1px solid var(--border-color);
    border-radius: 12px;
    padding: 1rem;
    backdrop-filter: blur(10px);
    transition: all 0.3s ease;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
    position: relative;
  }

  .config-card.compact {
    padding: 0.75rem 1rem;
  }

  .config-card:hover {
    border-color: var(--accent-primary);
    box-shadow: 0 4px 20px rgba(139, 92, 246, 0.15);
    transform: translateY(-1px);
  }

  .config-card.full-width {
    grid-column: 1 / -1;
  }

  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 0.75rem;
  }

  .card-header h3 {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin: 0;
    font-size: 1rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .info-btn {
    background: transparent;
    border: 1px solid var(--border-color);
    border-radius: 50%;
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    color: var(--text-secondary);
    transition: all 0.2s ease;
    padding: 0;
  }

  .info-btn:hover {
    background: rgba(139, 92, 246, 0.1);
    border-color: var(--accent-primary);
    color: var(--accent-primary);
  }

  .info-tooltip {
    position: absolute;
    top: 3rem;
    right: 1rem;
    background: var(--card-bg);
    border: 1px solid var(--accent-primary);
    border-radius: 8px;
    padding: 0.5rem 0.75rem;
    font-size: 0.85rem;
    color: var(--text-secondary);
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
    z-index: 100;
    max-width: 250px;
    animation: fadeIn 0.2s ease;
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
      transform: translateY(-5px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .card-content {
    margin-top: 0.5rem;
  }

  .directory-selection {
    border: 2px dashed var(--border-color);
    background: rgba(139, 92, 246, 0.05);
  }

  .directory-selection:hover {
    border-color: var(--accent-primary);
    background: rgba(139, 92, 246, 0.1);
  }

  .directory-input {
    display: flex;
    gap: 0.5rem;
    align-items: stretch;
  }

  .status-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 0.75rem;
  }

  .status-indicator {
    font-size: 0.8rem;
    color: var(--text-secondary);
  }

  .status-indicator.active {
    color: var(--accent-primary);
  }

  .status-indicator.busy {
    color: var(--warning-color);
  }

  .status-link {
    background: none;
    border: none;
    color: var(--accent-primary);
    font-size: 0.8rem;
    cursor: pointer;
    padding: 0;
  }

  .status-link:hover {
    text-decoration: underline;
  }

  .directory-path {
    flex: 1;
    padding: 0.65rem;
    background: var(--card-bg);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    color: var(--text-primary);
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 0.85rem;
  }

  .directory-path:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .browse-btn {
    padding: 0.65rem 1rem;
    background: var(--accent-primary);
    border: none;
    border-radius: 8px;
    color: white;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-weight: 500;
    transition: all 0.2s ease;
    white-space: nowrap;
  }

  .browse-btn:hover:not(:disabled) {
    background: var(--accent-secondary);
    transform: translateY(-1px);
  }

  .browse-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .action-info-card {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 0.75rem;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid var(--border-color);
    border-radius: 8px;
  }

  .action-info-card.compact {
    padding: 0.5rem;
  }

  .secondary-btn {
    padding: 0.5rem 1rem;
    border-radius: 8px;
    border: 1px solid rgba(139, 92, 246, 0.6);
    background: transparent;
    color: var(--text-primary);
    font-weight: 500;
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    cursor: pointer;
    transition: all 0.2s ease;
    font-size: 0.9rem;
  }

  .secondary-btn:hover {
    background: rgba(139, 92, 246, 0.1);
    transform: translateY(-1px);
  }

  .secondary-btn:focus {
    outline: none;
    box-shadow: 0 0 0 3px rgba(139, 92, 246, 0.2);
  }

  .secondary-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
    background: transparent;
  }

  .pattern-naming-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 1rem;
  }

  .form-group {
    margin-bottom: 0.75rem;
  }

  .form-group.compact {
    margin-bottom: 0.5rem;
  }

  .form-group:last-child {
    margin-bottom: 0;
  }

  .form-group input,
  .form-group select {
    width: 100%;
    padding: 0.65rem;
    background: var(--card-bg);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    color: var(--text-primary);
    font-size: 0.9rem;
    transition: border-color 0.2s ease;
  }

  .form-group input:focus,
  .form-group select:focus {
    outline: none;
    border-color: var(--accent-primary);
    box-shadow: 0 0 0 3px rgba(139, 92, 246, 0.1);
  }

  .form-group input:disabled,
  .form-group select:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .help-text {
    margin-top: 0.4rem;
    font-size: 0.75rem;
    color: var(--text-muted);
    line-height: 1.3;
  }

  .help-text.compact {
    font-size: 0.7rem;
    margin-top: 0.25rem;
  }

  .start-btn,
  .stop-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.65rem 1rem;
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
    border: none;
    transition: all 0.2s ease;
    font-size: 0.9rem;
    white-space: nowrap;
  }

  .start-btn {
    background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary, var(--accent-primary)));
    color: white;
  }

  .start-btn:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 4px 20px rgba(139, 92, 246, 0.4);
  }

  .stop-btn {
    background: rgba(239, 68, 68, 0.15);
    color: var(--error-color);
    border: 1px solid var(--error-color);
  }

  .stop-btn:hover:not(:disabled) {
    background: rgba(239, 68, 68, 0.25);
  }

  .start-btn:disabled,
  .stop-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
    box-shadow: none;
  }

  /* Responsive */
  @media (max-width: 768px) {
    .config-section {
      padding: 1rem;
    }

    .pattern-naming-grid {
      grid-template-columns: 1fr;
    }

    .directory-input {
      flex-wrap: wrap;
    }

    .browse-btn,
    .start-btn,
    .stop-btn {
      flex: 1;
      justify-content: center;
    }
  }
</style>