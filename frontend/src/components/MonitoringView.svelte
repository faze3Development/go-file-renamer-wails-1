<script lang="ts">
  import { BookText } from 'lucide-svelte';
  import type { StatsPayload, LogEntryPayload } from '../types/events';

  export let stats: StatsPayload;
  // Canonical structured logs for monitoring view
  export let logs: LogEntryPayload[] = [];
  export let logContainer: HTMLElement;
</script>

<div class="monitoring-section">
  <!-- Stats Panel -->
  <div class="stats-section">
    <div class="stats-header">
      <h3>Live Statistics</h3>
      <p>Real-time monitoring data</p>
    </div>
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon">◉</div>
        <div class="stat-content">
          <div class="stat-value">{stats.scanned}</div>
          <div class="stat-label">Files Scanned</div>
        </div>
      </div>
      <div class="stat-card success">
        <div class="stat-icon">◆</div>
        <div class="stat-content">
          <div class="stat-value">{stats.renamed}</div>
          <div class="stat-label">Files Renamed</div>
        </div>
      </div>
      <div class="stat-card warning">
        <div class="stat-icon">◇</div>
        <div class="stat-content">
          <div class="stat-value">{stats.skipped}</div>
          <div class="stat-label">Files Skipped</div>
        </div>
      </div>
      <div class="stat-card error">
        <div class="stat-icon">◈</div>
        <div class="stat-content">
          <div class="stat-value">{stats.errors}</div>
          <div class="stat-label">Errors</div>
        </div>
      </div>
    </div>
  </div>

  <!-- Log Viewer -->
  <div class="log-section">
    <div class="log-header">
      <h3><BookText size={20} />Activity Log</h3>
      <p>Real-time operation details</p>
    </div>
    <div class="log-viewer">
      <div class="log-container" bind:this={logContainer}>
        {#each logs as log, i (i)}
            <div
              class="log-entry"
              class:log-error={(log.severity ?? 'info').toString().toUpperCase() === "ERROR"}
              class:log-warn={(log.severity ?? 'info').toString().toUpperCase() === "WARN"}
              class:log-info={(log.severity ?? 'info').toString().toUpperCase() === "INFO"}
              class:log-debug={(log.severity ?? 'info').toString().toUpperCase() === "DEBUG"}
            >
              <span class="log-timestamp">{new Date(log.timestamp).toLocaleTimeString()}</span>
              <span class="log-level-badge">{(log.severity ?? 'info').toString().toUpperCase()}</span>
              <span class="log-message">{log.message}</span>
            </div>
        {/each}
        {#if logs.length === 0}
          <div class="log-placeholder">
            <div class="placeholder-icon">◊</div>
            <p>
              Activity logs will appear here when you start
              watching...
            </p>
          </div>
        {/if}
      </div>
    </div>
  </div>
</div>

<style>
  .monitoring-section {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    padding: 1.5rem;
    height: 100%;
    overflow: hidden;
    width: 100%;
    max-width: 1100px;
    margin: 0 auto;
    box-sizing: border-box;
  }

  .stats-section {
    flex-shrink: 0;
  }

  .stats-header h3 {
    margin: 0 0 0.5rem;
    font-size: 1.15rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .stats-header p {
    margin: 0;
    color: var(--text-secondary);
    font-size: 0.85rem;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 0.75rem;
    margin-top: 1rem;
  }

  .stat-card {
    background: linear-gradient(
      135deg,
      var(--card-bg),
      rgba(22, 33, 62, 0.8) 100%
    );
    border: 1px solid var(--border-color);
    border-radius: 10px;
    padding: 1rem;
    display: flex;
    align-items: center;
    gap: 0.75rem;
    transition: all 0.3s ease;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  }

  .stat-card:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  }

  .stat-card.success {
    border-color: var(--success-color);
    background: linear-gradient(
      135deg,
      rgba(16, 185, 129, 0.08),
      rgba(22, 33, 62, 0.8) 100%
    );
  }

  .stat-card.warning {
    border-color: var(--warning-color);
    background: linear-gradient(
      135deg,
      rgba(245, 158, 11, 0.08),
      rgba(22, 33, 62, 0.8) 100%
    );
  }

  .stat-card.error {
    border-color: var(--error-color);
    background: linear-gradient(
      135deg,
      rgba(239, 68, 68, 0.08),
      rgba(22, 33, 62, 0.8) 100%
    );
  }

  .stat-icon {
    font-size: 1.5rem;
    opacity: 0.7;
  }

  .stat-card.success .stat-icon {
    color: var(--success-color);
  }

  .stat-card.warning .stat-icon {
    color: var(--warning-color);
  }

  .stat-card.error .stat-icon {
    color: var(--error-color);
  }

  .stat-content {
    flex: 1;
  }

  .stat-value {
    font-size: 1.75rem;
    font-weight: 700;
    color: var(--text-primary);
    line-height: 1;
  }

  .stat-label {
    font-size: 0.8rem;
    color: var(--text-secondary);
    margin-top: 0.25rem;
  }

  .log-section {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-height: 0;
    overflow: hidden;
  }

  .log-header h3 {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin: 0 0 0.5rem;
    font-size: 1.15rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .log-header p {
    margin: 0;
    color: var(--text-secondary);
    font-size: 0.85rem;
  }

  .log-viewer {
    flex: 1;
    background: linear-gradient(
      135deg,
      var(--card-bg),
      rgba(22, 33, 62, 0.8) 100%
    );
    border: 1px solid var(--border-color);
    border-radius: 10px;
    margin-top: 0.75rem;
    display: flex;
    flex-direction: column;
    min-height: 0;
    overflow: hidden;
  }

  .log-container {
    flex: 1;
    overflow-y: auto;
    padding: 0.75rem;
    background: var(--card-bg);
    border-radius: 10px;
  }

  .log-entry {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.4rem 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 0.8rem;
  }

  .log-entry:last-child {
    border-bottom: none;
  }

  .log-timestamp {
    color: var(--text-muted);
    font-size: 0.75rem;
    min-width: 70px;
    flex-shrink: 0;
  }

  .log-level-badge {
    padding: 0.15rem 0.4rem;
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
    color: black;
  }

  .log-entry.log-info .log-level-badge {
    background: var(--accent-primary);
    color: white;
  }

  .log-entry.log-debug .log-level-badge {
    background: var(--text-muted);
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
    height: 150px;
    color: var(--text-muted);
    text-align: center;
  }

  .placeholder-icon {
    font-size: 2.5rem;
    margin-bottom: 0.75rem;
    opacity: 0.4;
  }

  .log-placeholder p {
    margin: 0;
    font-size: 0.85rem;
  }

  /* Mobile responsiveness */
  @media (max-width: 768px) {
    .monitoring-section {
      padding: 1rem;
    }

    .stats-grid {
      grid-template-columns: 1fr 1fr;
    }

    .stat-card {
      padding: 0.75rem;
    }

    .stat-value {
      font-size: 1.5rem;
    }

    .log-container {
      padding: 0.5rem;
    }

    .log-entry {
      flex-wrap: wrap;
      gap: 0.25rem;
    }

    .log-timestamp,
    .log-level-badge {
      min-width: auto;
    }
  }
</style>