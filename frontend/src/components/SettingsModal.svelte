<script lang="ts">
  import { createEventDispatcher, onMount } from "svelte";
  import {
    Save,
    Trash2,
    X,
    Palette,
    UserSquare,
    Layout,
    Sparkles,
    Terminal,
    Leaf,
  } from "lucide-svelte";
  import Toggle from "./shared/Toggle.svelte";
  import {
    settings,
    themes,
    LOG_RETENTION_DEFAULT,
    LOG_RETENTION_SOFT_MAX,
    LOG_RETENTION_HARD_MAX,
  } from "../stores";

  const dispatch = createEventDispatcher<{
    close: null;
    updateTheme: string;
    selectProfile: string;
    saveProfile: null;
    deleteProfile: null;
    resetSettings: null;
  }>();

  let modalEl;

  export let profiles: Record<string, any> = {};
  export let selectedProfile: string = "";
  export let isWatching: boolean = false;

  const verbosityOptions = [
    {
      key: "global",
      title: "Verbose Backend Events",
      description: "Forward debug-level telemetry from all backend modules.",
    },
    {
      key: "watcher",
      title: "Watcher Diagnostics",
      description: "Include filesystem watcher state changes and errors.",
    },
    {
      key: "advancedOperations",
      title: "Advanced Operations Diagnostics",
      description: "Include rename pipeline insights and advanced ops traces.",
    },
  ];

  function close() {
    dispatch("close");
  }

  function updateTheme(themeName) {
    dispatch("updateTheme", themeName);
  }

  function handleProfileSelect(event) {
    dispatch("selectProfile", event.target.value);
  }

  function handleSaveProfile() {
    dispatch("saveProfile");
  }

  function handleDeleteProfile() {
    dispatch("deleteProfile");
  }

  function toggleCompactMode() {
    settings.updateSetting("compactMode", !$settings.compactMode);
  }

  function handleRetentionChange(event) {
    const next = Number(event.currentTarget.value);
    settings.updateSetting("logRetentionLimit", next);
  }

  function toggleBackendVerbosity(key, next) {
    // accept explicit boolean from Toggle component (or fallback to toggling)
    if (typeof next === 'boolean') {
      settings.updateBackendVerbosity(key, !!next);
    } else {
      settings.updateBackendVerbosity(key, !($settings.backendVerbosity?.[key] ?? false));
    }
  }

  function resetSettings() {
    dispatch("resetSettings");
  }

  onMount(() => {
    if (modalEl && typeof modalEl.focus === "function") {
      modalEl.focus();
    }
  });

  $: retentionLimit = $settings.logRetentionLimit ?? LOG_RETENTION_DEFAULT;
  $: retentionWarning = retentionLimit > LOG_RETENTION_SOFT_MAX;
  $: backendVerbosity = $settings.backendVerbosity ?? {};
</script>

<button
  class="modal-overlay"
  on:click={(e) => {
    if (e.target === e.currentTarget) close();
  }}
  aria-label="Close settings modal"
  type="button"
>
  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
  <div
    class="settings-modal"
    bind:this={modalEl}
    on:keydown={(e) => e.key === "Escape" && close()}
    role="dialog"
    aria-labelledby="settings-title"
    aria-modal="true"
    tabindex="-1"
  >
    <div class="modal-header">
      <h2 id="settings-title">⚛ Settings</h2>
      <button class="close-btn" on:click={close}>
        <X size={20} />
      </button>
    </div>

    <div class="modal-content">
      <!-- Theme Selection -->
      <div class="settings-section">
        <h3><Palette size={18} /> Theme</h3>
        <p class="section-description">Choose your preferred color scheme</p>
        <div class="theme-grid">
          {#each Object.entries(themes) as [themeKey, theme]}
            <button
              class="theme-option"
              class:active={$settings.theme === themeKey}
              on:click={() => updateTheme(themeKey)}
            >
              <div class="theme-icon">
                {#if themeKey === "default"}
                  <Sparkles size={20} />
                {:else if themeKey === "cyberpunk"}
                  <Terminal size={20} />
                {:else if themeKey === "forest"}
                  <Leaf size={20} />
                {:else}
                  <div
                    class="theme-preview"
                    style="background: {theme.colors['--accent-primary']};"
                  />
                {/if}
              </div>
              <div class="theme-info">
                <strong>{theme.name}</strong>
                <small>{theme.description}</small>
              </div>
            </button>
          {/each}
        </div>
      </div>

      <!-- Logging Controls -->
      <div class="settings-section">
        <h3><Layout size={18} /> Logging</h3>
        <p class="section-description">
          Control in-memory retention, backend verbosity, and export guardrails
        </p>
        <div class="logging-options">
          <label for="log-retention" class="logging-label">
            <strong>Log Retention Limit</strong>
            <small>
              Keep up to {retentionLimit.toLocaleString()} entries (default {LOG_RETENTION_DEFAULT.toLocaleString()}, soft max {LOG_RETENTION_SOFT_MAX.toLocaleString()}, hard max {LOG_RETENTION_HARD_MAX.toLocaleString()}).
            </small>
          </label>
          <input
            id="log-retention"
            class="logging-input"
            type="number"
            min="100"
            max={LOG_RETENTION_HARD_MAX}
            step="100"
            value={retentionLimit}
            on:change={handleRetentionChange}
            on:blur={handleRetentionChange}
          />
          {#if retentionWarning}
            <p class="logging-warning">
              Values above {LOG_RETENTION_SOFT_MAX.toLocaleString()} may affect performance; the app will enforce a hard ceiling of {LOG_RETENTION_HARD_MAX.toLocaleString()} entries.
            </p>
          {/if}
        </div>
        <div class="logging-verbosity">
          <span class="logging-subheading">Backend Verbosity</span>
          {#each verbosityOptions as option}
            <div class="logging-toggle">
              <Toggle
                checked={!!backendVerbosity[option.key]}
                on:change={(e) => toggleBackendVerbosity(option.key, e.detail.checked)}
                ariaLabel={option.title}
              />
              <span class="logging-toggle-copy">
                <strong>{option.title}</strong>
                <small>{option.description}</small>
              </span>
            </div>
          {/each}
        </div>
      </div>

      <!-- Profile Management -->
      <div class="settings-section">
        <h3><UserSquare size={18} /> Profile Management</h3>
        <p class="section-description">Save and load configuration presets</p>
        <div class="profile-card">
          <div class="profile-selector-wrapper">
            <div class="profile-selector">
              <label for="profile-select">Current Profile</label>
              <select
                id="profile-select"
                class="profile-dropdown"
                value={selectedProfile}
                on:change={handleProfileSelect}
                disabled={isWatching}
              >
                <option value="">Select Profile</option>
                {#each Object.keys(profiles || {}) as profileName}
                  <option value={profileName}>{profileName}</option>
                {/each}
              </select>
            </div>
            <div class="profile-buttons">
              <button
                class="secondary-btn"
                on:click={handleSaveProfile}
                disabled={isWatching}
              >
                <Save size={16} />
                Save Current
              </button>
              <button
                class="danger-btn"
                on:click={handleDeleteProfile}
                disabled={isWatching || !selectedProfile}
              >
                <Trash2 size={16} />
                Delete Selected
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- UI Options -->
      <div class="settings-section">
        <h3><Layout size={18} /> Interface</h3>
        <p class="section-description">
          Customize what's visible in your workspace
        </p>
        <div class="settings-options">
          <div class="setting-toggle">
            <Toggle
              checked={$settings.compactMode}
              on:change={(e) => settings.updateSetting('compactMode', e.detail.checked)}
              ariaLabel="Compact Mode"
            />
            <span class="toggle-content">
              <strong>Compact Mode</strong>
              <small>Reduce spacing for more content</small>
            </span>
          </div>
        </div>
      </div>

      <!-- Actions -->
      <div class="settings-section">
        <div class="settings-actions">
          <button class="secondary-btn" on:click={resetSettings}>
            Reset to Defaults
          </button>
          <button class="primary-btn" on:click={close}> Done </button>
        </div>
      </div>
    </div>
  </div>
</button>

<style>
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(8px);
    z-index: 10000;
    display: flex;
    align-items: center;
    justify-content: center;
    animation: fadeIn 0.2s ease;
    border: none;
    padding: 0;
    width: 100%;
    text-align: inherit;
    cursor: default;
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }

  .settings-modal {
    background: var(--secondary-bg);
    border: 1px solid var(--border-color);
    border-radius: 24px;
    width: 90%;
    max-width: 700px;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
    box-shadow: 0 20px 50px rgba(0, 0, 0, 0.4);
    animation: slideUp 0.3s ease;
  }

  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(20px) scale(0.95);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1.5rem 2rem;
    border-bottom: 1px solid var(--border-color);
    flex-shrink: 0;
  }

  .modal-header h2 {
    margin: 0;
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--text-primary);
  }

  .close-btn {
    background: rgba(255, 255, 255, 0.1);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-secondary);
    cursor: pointer;
  }

  .close-btn:hover {
    background: rgba(255, 255, 255, 0.15);
    color: var(--text-primary);
  }

  .modal-content {
    padding: 2rem;
    overflow-y: auto;
  }

  .settings-section {
    margin-bottom: 2.5rem;
  }

  .settings-section:last-child {
    margin-bottom: 0;
  }

  .settings-section h3 {
    margin: 0 0 0.5rem;
    font-size: 1.2rem;
    font-weight: 600;
    color: var(--text-primary);
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .section-description {
    margin: 0 0 1.5rem;
    color: var(--text-muted);
    font-size: 0.9rem;
  }

  /* Theme Selection */
  .theme-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 1rem;
  }

  .theme-option {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1rem;
    background: var(--card-bg);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    cursor: pointer;
    text-align: left;
  }

  .theme-option:hover {
    background: rgba(255, 255, 255, 0.08);
    border-color: var(--accent-primary, var(--accent-purple));
  }

  .theme-option.active {
    border-color: var(--accent-primary, var(--accent-purple));
    background: rgba(139, 92, 246, 0.1);
    box-shadow: 0 0 0 2px var(--accent-primary, var(--accent-purple));
  }

  .theme-preview {
    width: 100%;
    height: 100%;
  }

  .theme-icon {
    width: 40px;
    height: 40px;
    border-radius: 12px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(255, 255, 255, 0.1);
    color: var(--text-secondary);
  }

  .theme-option.active .theme-icon {
    background: var(--accent-primary, var(--accent-purple));
    color: white;
  }

  .theme-info {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .theme-info strong {
    color: var(--text-primary);
    font-weight: 600;
  }

  .theme-info small {
    color: var(--text-secondary);
    font-size: 0.75rem;
  }

  /* Profile Management */
  .profile-card {
    background: var(--card-bg);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    padding: 1rem;
    transition: all 0.2s ease;
  }

  .profile-card:hover {
    border-color: var(--accent-primary, var(--accent-purple));
  }

  .profile-selector-wrapper {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 1rem;
  }

  .profile-selector {
    display: flex;
    flex-direction: column;
    flex-grow: 1;
  }

  .profile-buttons {
    display: flex;
    gap: 0.75rem;
  }

  .profile-selector label {
    font-size: 0.8rem;
    font-weight: 500;
    color: var(--text-secondary);
    margin-bottom: 0.5rem;
  }

  .profile-dropdown {
    width: 100%;
    padding: 0.75rem 1rem;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid var(--border-color);
    border-radius: 10px;
    color: var(--text-primary);
    font-size: 0.875rem;
    transition: all 0.2s ease;
    -webkit-appearance: none;
    appearance: none;
    position: relative;
    z-index: 1;
  }

  .profile-dropdown:focus {
    outline: none;
    border-color: var(--accent-primary, var(--accent-purple));
    box-shadow: 0 0 0 2px rgba(139, 92, 246, 0.2);
  }

  /* Ensure dropdown options are readable: dark background and high-contrast text */
  .profile-dropdown option {
    background: var(--card-bg);
    color: var(--text-primary);
  }

  /* Hover/selection styling for the options where supported */
  .profile-dropdown option:hover,
  .profile-dropdown option:checked {
    background: rgba(139, 92, 246, 0.08);
    color: var(--text-primary);
  }

  /* Provide a visible caret for the select (non-intrusive) */
  .profile-selector {
    position: relative;
  }

  .profile-selector::after {
    content: "\25BE"; /* downward caret */
    position: absolute;
    right: 12px;
    top: 50%;
    transform: translateY(-50%);
    pointer-events: none;
    color: var(--text-secondary);
    font-size: 0.85rem;
  }

  /* UI Options */
  .settings-options {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .logging-options {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    background: var(--card-bg);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    padding: 1rem;
  }

  .logging-label {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    color: var(--text-secondary);
  }

  .logging-input {
    width: 180px;
    padding: 0.5rem 0.75rem;
    border-radius: 8px;
    border: 1px solid var(--border-color);
    background: var(--primary-bg);
    color: var(--text-primary);
    font-size: 0.95rem;
  }

  .logging-input:focus {
    outline: none;
    border-color: var(--accent-primary);
    box-shadow: 0 0 0 2px rgba(139, 92, 246, 0.2);
  }

  .logging-warning {
    margin: 0;
    font-size: 0.8rem;
    color: var(--warning-color);
  }

  .logging-verbosity {
    margin-top: 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .logging-subheading {
    font-size: 0.85rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--text-muted);
  }

  .logging-toggle {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    padding: 0.75rem 1rem;
    background: var(--card-bg);
    border: 1px solid var(--border-color);
    border-radius: 10px;
    cursor: pointer;
  }

  .logging-toggle-copy {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .logging-toggle-copy strong {
    color: var(--text-primary);
    font-size: 0.95rem;
  }

  .logging-toggle-copy small {
    color: var(--text-secondary);
    font-size: 0.8rem;
  }

  .setting-toggle {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1rem;
    background: var(--card-bg);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    cursor: pointer;
  }

  .setting-toggle:hover {
    background: rgba(255, 255, 255, 0.08);
    border-color: var(--accent-primary, var(--accent-purple));
  }

  /* Toggle visuals are provided by the shared Toggle component; local slider/input rules removed. */

  .toggle-content {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .toggle-content strong {
    color: var(--text-primary);
    font-weight: 600;
  }

  .toggle-content small {
    color: var(--text-secondary);
    font-size: 0.75rem;
  }

  /* Actions */
  .settings-actions {
    display: flex;
    gap: 1rem;
    justify-content: flex-end;
    margin-top: 1rem;
    padding-top: 1.5rem;
    border-top: 1px solid var(--border-color);
  }

  .secondary-btn,
  .primary-btn,
  .danger-btn {
    padding: 0.75rem 1.5rem;
    border-radius: 12px;
    font-weight: 600;
    cursor: pointer;
    border: none;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .secondary-btn {
    background: rgba(255, 255, 255, 0.1);
    color: var(--text-secondary);
    border: 1px solid var(--border-color);
  }

  .secondary-btn:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.15);
    color: var(--text-primary);
  }

  .primary-btn {
    background: linear-gradient(
      135deg,
      var(--accent-primary),
      var(--accent-secondary, var(--accent-primary))
    );
    color: white;
  }

  .primary-btn:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 4px 20px rgba(139, 92, 246, 0.4);
  }

  .danger-btn {
    background: rgba(239, 68, 68, 0.15);
    color: var(--error-color);
    border: 1px solid var(--error-color);
  }

  .danger-btn:hover:not(:disabled) {
    background: rgba(239, 68, 68, 0.25);
    color: #fca5a5;
  }
</style>