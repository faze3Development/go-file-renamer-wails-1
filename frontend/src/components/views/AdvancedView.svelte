<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import Toggle from "../shared/Toggle.svelte";
  import type { Config } from '../../../wailsjs/go/models';

  export let config: Config;
  export let isWatching: boolean;

  const dispatch = createEventDispatcher<{
    updateConfig: { key: string; value: any };
  }>();

  function updateConfig(key: string, value: any) {
    dispatch("updateConfig", { key, value });
  }

  let preview = "";
  $: {
    try {
      const now = new Date();
      const year = now.getFullYear().toString();
      const month = String(now.getMonth() + 1).padStart(2, "0");
      const day = String(now.getDate()).padStart(2, "0");
      const hour = String(now.getHours()).padStart(2, "0");
      const minute = String(now.getMinutes()).padStart(2, "0");
      const second = String(now.getSeconds()).padStart(2, "0");
      const date = now.toISOString().slice(0, 10);
      const time = now.toTimeString().slice(0, 5).replace(":", "-");
      const datetime = now
        .toISOString()
        .slice(0, 19)
        .replace(/:/g, "-")
        .replace("T", "_");
      const unix = Math.floor(Date.now() / 1000).toString();
      const unixmilli = Date.now().toString();

      if (config.NamerID === "datetime") {
        preview = datetime;
      } else if (config.NamerID === "custom_datetime") {
        preview = config.DateTimeFormat.replace("2006", year)
          .replace("01", month)
          .replace("02", day)
          .replace("15", hour)
          .replace("04", minute)
          .replace("05", second);
      } else if (config.NamerID === "template") {
        preview = config.TemplateString.replace("{original}", "example")
          .replace("{date}", date)
          .replace("{time}", time)
          .replace("{datetime}", datetime)
          .replace("{year}", year)
          .replace("{month}", month)
          .replace("{day}", day)
          .replace("{hour}", hour)
          .replace("{minute}", minute)
          .replace("{second}", second)
          .replace("{unix}", unix)
          .replace("{unixmilli}", unixmilli)
          .replace("{count}", "1")
          .replace("{count:2}", "01")
          .replace("{count:3}", "001")
          .replace("{count:4}", "0001");
      } else {
        preview = "example.jpg";
      }
    } catch (e) {
      console.error("Error generating filename preview:", e);
      preview = "Error generating preview";
    }
  }
</script>

<div class="advanced-section">
  <!-- Processing Options -->
  <div class="config-card">
    <div class="card-header">
      <h3>📁 Directory Processing</h3>
      <p>Configure how files and directories are processed</p>
    </div>
    <div class="card-content">
      <div class="option-toggle">
        <Toggle
          checked={config.Recursive}
          disabled={isWatching}
          on:change={(e) => updateConfig("Recursive", e.detail.checked)}
          ariaLabel="Watch Recursively"
        />
        <span class="toggle-label">
          <strong>Watch Recursively</strong>
          <small>Monitor subdirectories and their contents</small>
        </span>
      </div>
    </div>
  </div>

  <!-- Execution Options -->
  <div class="config-card">
    <div class="card-header">
      <h3>⚙️ Execution Mode</h3>
      <p>Control how file operations are performed</p>
    </div>
    <div class="card-content">
      <div class="options-grid">
        <div class="option-toggle">
          <Toggle
            checked={config.DryRun}
            disabled={isWatching}
            on:change={(e) => updateConfig("DryRun", e.detail.checked)}
            ariaLabel="Dry Run Mode"
          />
          <span class="toggle-label">
            <strong>Dry Run Mode</strong>
            <small>Preview changes without actually renaming files</small>
          </span>
        </div>
        <div class="option-toggle">
          <Toggle
            checked={config.NoInitialScan}
            disabled={isWatching}
            on:change={(e) => updateConfig("NoInitialScan", e.detail.checked)}
            ariaLabel="Skip Initial Scan"
          />
          <span class="toggle-label">
            <strong>Skip Initial Scan</strong>
            <small>Only process new files, ignore existing ones</small>
          </span>
        </div>
      </div>
    </div>
  </div>

  <!-- Date/Time Preview -->
  <div class="config-card">
    <div class="card-header">
      <h3>📅 Date & Time Preview</h3>
      <p>Preview how your date/time naming will look</p>
    </div>
    <div class="card-content">
      <div class="datetime-preview">
        <div class="preview-item">
          <label for="filename-preview">Sample Filename:</label>
          <div class="filename-preview" id="filename-preview">
            {preview}
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Performance Tuning -->
  <div class="config-card">
    <div class="card-header">
      <h3>🚀 Performance Settings</h3>
      <p>Fine-tune processing behavior and timing</p>
    </div>
    <div class="card-content">
      <div class="advanced-form-grid">
        <div class="form-group">
          <label for="settle-time">File Settle Time (ms)</label>
          <input
            id="settle-time"
            type="number"
            bind:value={config.Settle}
            disabled={isWatching}
            min="0"
            max="5000"
            step="50"
          />
          <div class="help-text">
            Time to wait for file operations to complete
          </div>
        </div>
        <div class="form-group">
          <label for="settle-timeout">Settle Timeout (seconds)</label>
          <input
            id="settle-timeout"
            type="number"
            bind:value={config.SettleTimeout}
            disabled={isWatching}
            min="1"
            max="30"
          />
          <div class="help-text">Maximum wait time for file stability</div>
        </div>
        <div class="form-group">
          <label for="retry-count">Retry Attempts</label>
          <input
            id="retry-count"
            type="number"
            bind:value={config.Retries}
            disabled={isWatching}
            min="1"
            max="10"
          />
          <div class="help-text">
            Number of retry attempts for failed operations
          </div>
        </div>
      </div>
    </div>
  </div>
</div>

<style>
  .advanced-section {
    padding: 1.5rem;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 1rem;
    min-height: 0;
    height: 100%;
  }

  .advanced-form-grid {
    display: grid;
    /* allow tighter wrapping so inputs don't overflow on narrow windows */
    grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
    gap: 0.75rem;
  }

  /* Date/Time Preview Styles */
  .datetime-preview {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .preview-item {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }

  .preview-item label {
    font-size: 0.85rem;
    font-weight: 500;
    color: var(--text-primary);
  }

  .filename-preview {
    padding: 0.65rem 0.85rem;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    color: var(--text-primary);
    font-family: "Monaco", "Menlo", "Ubuntu Mono", monospace;
    font-size: 0.85rem;
    word-break: break-all;
  }

  .config-card {
    background: linear-gradient(
      135deg,
      var(--card-bg) 0%,
      rgba(22, 33, 62, 0.8) 100%
    );
    border: 1px solid var(--border-color);
    border-radius: 10px;
    padding: 1rem;
    backdrop-filter: blur(10px);
    transition: all 0.3s ease;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  }

  .config-card:hover {
    border-color: rgba(139, 92, 246, 0.6);
    box-shadow: 0 4px 20px rgba(139, 92, 246, 0.15);
    transform: translateY(-1px);
  }

  .card-header h3 {
    margin: 0 0 0.35rem;
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--text-primary);
    display: flex;
    align-items: center;
    gap: 0.4rem;
  }

  .card-header p {
    margin: 0;
    color: var(--text-muted);
    font-size: 0.75rem;
  }

  .card-content {
    margin-top: 0.75rem;
  }

  .form-group {
    margin-bottom: 0.75rem;
  }

  .form-group:last-child {
    margin-bottom: 0;
  }

  .form-group label {
    display: block;
    margin-bottom: 0.4rem;
    font-size: 0.85rem;
    font-weight: 500;
    color: var(--text-primary);
  }

  .form-group input {
    width: 100%;
    box-sizing: border-box;
    min-width: 0; /* allow grid items to shrink on small viewports */
    padding: 0.65rem;
    background: var(--card-bg);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    color: var(--text-primary);
    font-size: 0.85rem;
  }

  .form-group input:focus {
    outline: none;
    border-color: var(--accent-primary);
    box-shadow: 0 0 0 3px rgba(139, 92, 246, 0.1);
  }

  .help-text {
    margin-top: 0.3rem;
    font-size: 0.7rem;
    color: var(--text-muted);
    line-height: 1.3;
  }

  .options-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 0.75rem;
  }

  .option-toggle {
    position: relative;
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem;
    border-radius: 8px;
    border: 1px solid rgba(255, 255, 255, 0.08);
    background: rgba(0, 0, 0, 0.25);
    cursor: pointer;
    transition: all 0.2s ease;
    overflow: visible;
  }

  .option-toggle:hover {
    border-color: rgba(139, 92, 246, 0.4);
    background: rgba(139, 92, 246, 0.06);
  }

  /* Toggle visuals are provided by the shared Toggle component; local input/slider rules removed. */

  .toggle-label {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    flex: 1;
  }

  .toggle-label strong {
    color: var(--text-primary);
    font-size: 0.9rem;
  }

  .toggle-label small {
    color: var(--text-muted);
    font-size: 0.75rem;
  }

  @media (max-width: 768px) {
    .advanced-section {
      padding: 1rem;
    }

    .advanced-form-grid {
      grid-template-columns: 1fr;
    }
  }
</style>