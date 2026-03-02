<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { CircleAlert, RefreshCw, X } from 'lucide-svelte';

  export let errorMessage: string = '';
  export let onRetry: (() => void) | null = null;
  export let onDismiss: (() => void) | null = null;

  const dispatch = createEventDispatcher<{
    retry: null;
    dismiss: null;
  }>();

  function handleRetry() {
    if (onRetry) {
      onRetry();
    } else {
      dispatch('retry');
    }
  }

  function handleDismiss() {
    if (onDismiss) {
      onDismiss();
    } else {
      dispatch('dismiss');
    }
  }
</script>

<div class="error-boundary">
  <div class="error-content">
    <div class="error-icon">
      <CircleAlert size={48} />
    </div>
    <h2>Something went wrong</h2>
    <p class="error-message">{errorMessage}</p>
    <div class="error-actions">
      <button class="retry-btn" on:click={handleRetry}>
        <RefreshCw size={16} />
        Retry
      </button>
      <button class="secondary-btn" on:click={handleDismiss}>
        <X size={16} />
        Dismiss
      </button>
    </div>
  </div>
</div>

<style>
  .error-boundary {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100vh;
    background: linear-gradient(
      135deg,
      var(--primary-bg) 0%,
      var(--secondary-bg) 100%
    );
    padding: 2rem;
  }

  .error-content {
    text-align: center;
    max-width: 500px;
    background: linear-gradient(
      135deg,
      var(--card-bg),
      rgba(22, 33, 62, 0.8) 100%
    );
    border: 1px solid var(--border-color);
    border-radius: var(--border-radius);
    padding: 3rem 2rem;
    backdrop-filter: blur(20px);
    box-shadow: var(--shadow);
  }

  .error-icon {
    color: var(--error-color);
    margin-bottom: 1.5rem;
  }

  .error-content h2 {
    margin: 0 0 1rem;
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--text-primary);
  }

  .error-message {
    margin: 0 0 2rem;
    color: var(--text-secondary);
    font-size: 0.9rem;
    line-height: 1.5;
  }

  .error-actions {
    display: flex;
    gap: 1rem;
    justify-content: center;
  }

  .retry-btn {
    padding: 0.75rem 1.5rem;
    background: linear-gradient(
      135deg,
      var(--accent-primary),
      var(--accent-secondary, var(--accent-primary))
    );
    border: none;
    border-radius: 0 50px 50px 0;
    color: white;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .retry-btn:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 20px rgba(139, 92, 246, 0.4);
  }

  .secondary-btn {
    padding: 0.75rem 1.5rem;
    background: rgba(255, 255, 255, 0.1);
    border: 1px solid var(--border-color);
    border-radius: 50px 0 0 50px;
    color: var(--text-secondary);
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .secondary-btn:hover {
    background: rgba(255, 255, 255, 0.15);
    border-color: var(--accent-primary);
  }
</style>