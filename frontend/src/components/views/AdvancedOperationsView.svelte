<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import AdvancedOperationsPanel from "./AdvancedOperationsPanel.svelte";
  import type { ErrorPayload } from "../../lib/errorBus";
  import type { Config } from '../../../wailsjs/go/models';

  export let config: Config;

  const dispatch = createEventDispatcher<{
    error: ErrorPayload;
  }>();

  let panelError = null;

  function handlePanelError(event) {
    panelError = event?.detail || null;
    dispatch("error", event.detail);
  }

  function clearPanelError() {
    panelError = null;
  }
</script>

<div class="operations-view">
  <div class="panel-wrapper">
    {#if panelError}
      <div class="error-banner">
        <div>
          <strong>{panelError?.context || "Advanced operations"}</strong>
          <span>{panelError?.message || "An unexpected error occurred."}</span>
        </div>
        <button type="button" on:click={clearPanelError}>Dismiss</button>
      </div>
    {/if}
    <AdvancedOperationsPanel {config} on:error={handlePanelError} />
  </div>
</div>

<style>
  .operations-view {
    padding: 1.5rem;
    overflow-y: auto;
    height: 100%;
    box-sizing: border-box;
  }

  .panel-wrapper {
    max-width: 1200px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
  }

  .error-banner {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
    padding: 0.75rem 1rem;
    border: 1px solid rgba(239, 68, 68, 0.3);
    border-radius: 10px;
    background: rgba(239, 68, 68, 0.12);
    color: var(--text-primary);
  }

  .error-banner strong {
    display: block;
    font-size: 0.85rem;
    margin-bottom: 0.2rem;
  }

  .error-banner span {
    font-size: 0.8rem;
    color: var(--text-secondary);
  }

  .error-banner button {
    background: none;
    border: 1px solid rgba(239, 68, 68, 0.4);
    border-radius: 8px;
    color: var(--text-primary);
    padding: 0.4rem 0.75rem;
    cursor: pointer;
    font-size: 0.75rem;
  }

  .error-banner button:hover {
    background: rgba(239, 68, 68, 0.08);
  }

  @media (max-width: 768px) {
    .operations-view {
      padding: 1rem;
    }
  }
</style>
