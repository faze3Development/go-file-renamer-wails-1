<script lang="ts">
  export let title: string = 'Loading Application';
  export let message: string = 'Initializing components and loading data...';
</script>

<div class="loading-boundary">
  <div class="loading-content">
    <div class="loading-spinner"></div>
    <h2>{title}</h2>
    <p>{message}</p>
  </div>
</div>

<style>
  .loading-boundary {
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

  .loading-content {
    text-align: center;
    max-width: 400px;
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

  .loading-spinner {
    width: 48px;
    height: 48px;
    border: 3px solid var(--border-color);
    border-top: 3px solid var(--accent-primary);
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin: 0 auto 1.5rem;
  }

  @keyframes spin {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }

  .loading-content h2 {
    margin: 0 0 0.5rem;
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .loading-content p {
    margin: 0;
    color: var(--text-secondary);
    font-size: 0.9rem;
  }
</style>