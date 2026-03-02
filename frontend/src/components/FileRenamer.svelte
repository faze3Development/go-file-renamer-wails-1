<script lang="ts">
  import { Activity, AlertCircle, CheckCircle, Clock } from 'lucide-svelte';
  import type { StatsPayload, LogEntryPayload } from '../types/events';

  export let isWatching: boolean = false;
  export let isBusy: boolean = false;
  export let stats: StatsPayload = { scanned: 0, renamed: 0, skipped: 0, errors: 0 };
  export let logs: LogEntryPayload[] = [];

  function getMessage(log: LogEntryPayload) {
    return log.message ?? '';
  }

  function getLevel(log: LogEntryPayload) {
    const sev = log.severity ?? 'info';
    return String(sev).toUpperCase();
  }

  function getTimestamp(log: LogEntryPayload) {
    return log.timestamp ?? new Date().toISOString();
  }

  $: statusMessage = isWatching
    ? "Watching for file changes..."
    : isBusy
    ? "Processing..."
    : "Ready to start watching";

  $: statusIcon = isWatching
    ? Activity
    : isBusy
    ? Clock
    : CheckCircle;

  $: statusColor = isWatching
    ? "var(--accent-primary)"
    : isBusy
    ? "var(--warning-color)"
    : "var(--success-color)";
</script>

<div class="file-renamer">
  <!-- Status Header -->
  <div class="renamer-header">
    <div class="status-indicator">
      <svelte:component this={statusIcon} size={20} style="color: {statusColor}" />
      <span class="status-text">{statusMessage}</span>
    </div>
  </div>

  <!-- Statistics -->
  <div class="renamer-stats">
    <div class="stat-item">
      <div class="stat-value">{stats.scanned}</div>
      <div class="stat-label">Files Scanned</div>
    </div>
    <div class="stat-item">
      <div class="stat-value">{stats.renamed}</div>
      <div class="stat-label">Files Renamed</div>
    </div>
    <div class="stat-item">
      <div class="stat-value">{stats.skipped}</div>
      <div class="stat-label">Files Skipped</div>
    </div>
    <div class="stat-item">
      <div class="stat-value">{stats.errors}</div>
      <div class="stat-label">Errors</div>
    </div>
  </div>

  <!-- Progress Indicator -->
  {#if isBusy}
    <div class="progress-indicator">
      <div class="progress-bar">
        <div class="progress-fill" style="width: {stats.scanned > 0 ? Math.min((stats.renamed / stats.scanned) * 100, 100) : 0}%"></div>
      </div>
      <span class="progress-text">
        {stats.renamed} of {stats.scanned} files processed
      </span>
    </div>
  {/if}

  <!-- Activity Log -->
  <div class="log-section">
    <div class="log-header">
      <h4>Activity Log</h4>
      <p>Real-time operation details</p>
    </div>
    <div class="log-viewer">
      <div class="log-container">
        {#each logs as log, i (i)}
          <div
            class="log-entry"
            class:log-error={getLevel(log) === "ERROR"}
            class:log-warn={getLevel(log) === "WARN"}
            class:log-info={getLevel(log) === "INFO"}
            class:log-debug={getLevel(log) === "DEBUG"}
          >
            <span class="log-timestamp">{new Date(getTimestamp(log)).toLocaleTimeString()}</span>
            <span class="log-level-badge">{getLevel(log)}</span>
            <span class="log-message">{getMessage(log)}</span>
          </div>
        {/each}
        {#if logs.length === 0}
          <div class="log-placeholder">
            <div class="placeholder-icon">◊</div>
            <p>Activity logs will appear here when you start watching...</p>
          </div>
        {/if}
      </div>
    </div>
  </div>
</div>

<style>
  .file-renamer {
    background: var(--card-bg);
    border: 1px solid var(--border-color);
    border-radius: 10px;
    padding: 1.25rem;
    margin: 1rem 1.5rem;
    display: flex;
    flex-direction: column;
    max-height: calc(100vh - 200px);
    overflow: hidden;
  }

  .renamer-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1rem;
    flex-wrap: wrap;
    gap: 0.75rem;
    flex-shrink: 0;
  }

  .status-indicator {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .status-text {
    font-weight: 500;
    color: var(--text-primary);
    font-size: 0.9rem;
  }

  .renamer-stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(110px, 1fr));
    gap: 0.75rem;
    margin-bottom: 1rem;
    flex-shrink: 0;
  }

  .stat-item {
    text-align: center;
    padding: 0.75rem;
    background: rgba(255, 255, 255, 0.04);
    border-radius: 8px;
    border: 1px solid var(--border-color);
  }

  .stat-value {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--text-primary);
    line-height: 1;
  }

  .stat-label {
    font-size: 0.75rem;
    color: var(--text-secondary);
    margin-top: 0.25rem;
  }

  .progress-indicator {
    margin: 0.75rem 0;
    flex-shrink: 0;
  }

  .progress-bar {
    width: 100%;
    height: 6px;
    background: rgba(255, 255, 255, 0.08);
    border-radius: 3px;
    overflow: hidden;
    margin-bottom: 0.4rem;
  }

  .progress-fill {
    height: 100%;
    background: linear-gradient(90deg, var(--accent-primary), var(--accent-secondary, var(--accent-primary)));
    border-radius: 3px;
    transition: width 0.3s ease;
  }

  .progress-text {
    font-size: 0.8rem;
    color: var(--text-secondary);
    text-align: center;
  }

  /* Log Section */
  .log-section {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-height: 0;
    overflow: hidden;
  }

  .log-header {
    margin-bottom: 0.75rem;
    flex-shrink: 0;
  }

  .log-header h4 {
    font-size: 1rem;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0 0 0.25rem 0;
  }

  .log-header p {
    font-size: 0.8rem;
    color: var(--text-secondary);
    margin: 0;
  }

  .log-viewer {
    flex: 1;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    min-height: 0;
  }

  .log-container {
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
    padding: 0.5rem;
  }

  .log-entry {
    display: flex;
    align-items: flex-start;
    gap: 0.5rem;
    padding: 0.4rem;
    border-radius: 4px;
    margin-bottom: 0.25rem;
    font-size: 0.8rem;
    background: rgba(255, 255, 255, 0.02);
  }

  .log-entry:last-child {
    margin-bottom: 0;
  }

  .log-timestamp {
    color: var(--text-secondary);
    font-family: 'Courier New', monospace;
    font-size: 0.75rem;
    min-width: 65px;
    flex-shrink: 0;
  }

  .log-level-badge {
    padding: 0.1rem 0.3rem;
    border-radius: 3px;
    font-size: 0.7rem;
    font-weight: 600;
    text-transform: uppercase;
    min-width: 45px;
    text-align: center;
    flex-shrink: 0;
  }

  .log-entry.log-error .log-level-badge {
    background: var(--error-color);
    color: white;
  }

  .log-entry.log-warn .log-level-badge {
    background: var(--warning-color);
    color: white;
  }

  .log-entry.log-info .log-level-badge {
    background: var(--accent-primary);
    color: white;
  }

  .log-entry.log-debug .log-level-badge {
    background: var(--text-secondary);
    color: white;
  }

  .log-message {
    color: var(--text-primary);
    flex: 1;
    word-break: break-word;
    min-width: 0;
  }

  .log-placeholder {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 2rem;
    text-align: center;
    color: var(--text-secondary);
    min-height: 120px;
  }

  .placeholder-icon {
    font-size: 2rem;
    margin-bottom: 0.5rem;
    opacity: 0.4;
  }

  /* Mobile responsiveness */
  @media (max-width: 768px) {
    .file-renamer {
      padding: 1rem;
      margin: 1rem;
    }

    .renamer-header {
      flex-direction: column;
      align-items: stretch;
    }

    .renamer-stats {
      grid-template-columns: repeat(2, 1fr);
    }

    .log-entry {
      flex-wrap: wrap;
    }

    .log-timestamp,
    .log-level-badge {
      min-width: auto;
    }
  }
</style>