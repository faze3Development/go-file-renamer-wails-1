<script lang="ts">
  import { FileCog, Activity, Cpu, Wrench } from 'lucide-svelte';

  export let currentView: string = 'configuration';
</script>

<header class="main-header">
  <div class="header-content">
    {#if currentView === "configuration"}
      <h1><FileCog size={32} />Intelligent File Renaming</h1>
      <p class="header-subtitle">
        Automate your file organization with smart pattern matching and
        renaming
      </p>
    {:else if currentView === "monitoring"}
      <h1><Activity size={32} />Live Monitoring</h1>
      <p class="header-subtitle">
        Real-time statistics and activity logs
      </p>
    {:else if currentView === "advanced"}
      <h1><Cpu size={32} />Advanced Options</h1>
      <p class="header-subtitle">
        Fine-tune the behavior and processing settings
      </p>
    {:else if currentView === "advancedOperations"}
      <h1><Wrench size={32} />Advanced Operations</h1>
      <p class="header-subtitle">
        Run ad-hoc batches through the processing pipeline
      </p>
    {/if}
  </div>
</header>

<style>
  .main-header {
    padding: 2rem 2rem 1rem;
    border-bottom: 1px solid var(--border-color);
    background: rgba(255, 255, 255, 0.02);
    backdrop-filter: blur(10px);
  }

  .header-content h1 {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    letter-spacing: -0.5px;
    margin: 0 0 0.5rem;
    font-size: 2.25rem;
    font-weight: 700;
    background: linear-gradient(
      135deg,
      var(--accent-primary),
      var(--accent-secondary, var(--accent-primary))
    );
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .header-subtitle {
    margin: 0;
    color: var(--text-secondary);
    font-size: 1.1rem;
    font-weight: 400;
  }
</style>