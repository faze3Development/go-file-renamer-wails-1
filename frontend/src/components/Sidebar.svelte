<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { FileText, Activity, Layers, Wrench, Settings } from "lucide-svelte";

  export let currentView: string = "configuration";
  export let isWatching: boolean = false;

  const dispatch = createEventDispatcher<{
    switchView: string;
    openSettings: null;
  }>();

  function switchView(view) {
    dispatch("switchView", view);
  }

  function openSettings() {
    dispatch("openSettings");
  }
</script>

<aside class="sidebar">
  <div class="sidebar-header">
    <div class="app-logo">
      <div class="logo-icon">⧉</div>
      <h2>File Renamer</h2>
    </div>
    <div class="status-badge" class:watching={isWatching}>
      <div class="status-indicator" />
      {isWatching ? "Active" : "Idle"}
    </div>
  </div>

  <nav class="sidebar-nav">
    <div class="nav-section">
      <h3>Navigation</h3>
      <div
        class="nav-item"
        class:active={currentView === "configuration"}
        on:click={() => switchView("configuration")}
        on:keydown={(e) => e.key === "Enter" && switchView("configuration")}
        role="button"
        tabindex="0"
      >
        <FileText size={18} />
        <span>File Renaming</span>
      </div>
      <div
        class="nav-item"
        class:active={currentView === "monitoring"}
        on:click={() => switchView("monitoring")}
        on:keydown={(e) => e.key === "Enter" && switchView("monitoring")}
        role="button"
        tabindex="0"
      >
        <Activity size={18} />
        <span>Live Monitoring</span>
        {#if isWatching}
          <span class="nav-badge">Live</span>
        {/if}
      </div>
      <div
        class="nav-item"
        class:active={currentView === "advanced"}
        on:click={() => switchView("advanced")}
        on:keydown={(e) => e.key === "Enter" && switchView("advanced")}
        role="button"
        tabindex="0"
      >
        <Layers size={18} />
        <span>Advanced Options</span>
      </div>
      <div
        class="nav-item"
        class:active={currentView === "advancedOperations"}
        on:click={() => switchView("advancedOperations")}
        on:keydown={(e) =>
          e.key === "Enter" && switchView("advancedOperations")}
        role="button"
        tabindex="0"
      >
        <Wrench size={18} />
        <span>Advanced Operations</span>
      </div>
      <div
        class="nav-item"
        on:click={openSettings}
        on:keydown={(e) => e.key === "Enter" && openSettings()}
        role="button"
        tabindex="0"
      >
        <Settings size={18} />
        <span>Settings</span>
      </div>
    </div>
  </nav>
</aside>

<style>
  /* Sidebar - Refined for Desktop */
  .sidebar {
    width: 280px;
    background: linear-gradient(180deg, var(--sidebar-bg) 0%, #0f0f1a 100%);
    border-right: 1px solid var(--border-color);
    display: flex;
    flex-direction: column;
    padding: 0;
    backdrop-filter: blur(20px);
    box-shadow: 4px 0 20px rgba(0, 0, 0, 0.2);
    z-index: 10;
  }

  .sidebar-header {
    padding: 1.5rem 1.25rem 1rem;
    border-bottom: 1px solid var(--border-color);
    background: rgba(255, 255, 255, 0.02);
  }

  .app-logo {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 1rem;
  }

  .logo-icon {
    font-size: 1.4rem;
    background: linear-gradient(
      135deg,
      var(--accent-purple),
      var(--accent-pink)
    );
    padding: 0.5rem;
    border-radius: var(--border-radius);
    box-shadow: 0 4px 20px rgba(139, 92, 246, 0.4);
    backdrop-filter: blur(10px);
  }

  .app-logo h2 {
    margin: 0;
    font-size: 1.2rem;
    font-weight: 700;
    background: linear-gradient(
      135deg,
      var(--accent-purple),
      var(--accent-pink)
    );
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .status-badge {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 1rem;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 20px;
    font-size: 0.875rem;
    font-weight: 500;
    border: 1px solid var(--border-color);
  }

  .status-indicator {
    width: 8px;
    height: 8px;
    background: var(--text-muted);
    border-radius: 50%;
    transition: background-color 0.3s ease;
  }

  .status-badge.watching .status-indicator {
    background: var(--success-color);
    box-shadow: 0 0 10px rgba(16, 185, 129, 0.5);
  }

  /* Navigation */
  .sidebar-nav {
    flex-grow: 1;
    padding: 1.5rem 1rem;
    overflow-y: auto;
  }

  .nav-section {
    margin-bottom: 2rem;
  }

  .nav-section h3 {
    margin: 0 0 1rem;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--text-muted);
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem 1rem;
    border-radius: 12px;
    cursor: pointer;
    transition: all 0.2s ease;
    color: var(--text-secondary);
  }

  .nav-item:hover {
    background: rgba(255, 255, 255, 0.08);
    color: var(--text-primary);
  }

  .nav-item.active {
    background: linear-gradient(
      135deg,
      rgba(139, 92, 246, 0.2),
      rgba(236, 72, 153, 0.1)
    );
    color: var(--text-primary);
    border: 1px solid rgba(139, 92, 246, 0.3);
  }

  .nav-badge {
    background: linear-gradient(135deg, var(--success-color), #059669);
    color: white;
    font-size: 0.6rem;
    font-weight: 600;
    padding: 0.125rem 0.5rem;
    border-radius: 10px;
    margin-left: auto;
    box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);
    animation: pulse 2s infinite;
  }

  @keyframes pulse {
    0%,
    100% {
      opacity: 1;
      transform: scale(1);
    }
    50% {
      opacity: 0.8;
      transform: scale(0.95);
    }
  }
</style>