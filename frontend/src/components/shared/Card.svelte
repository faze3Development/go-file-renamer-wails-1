<script lang="ts">
  export let compact: boolean = false;
  export let fullWidth: boolean = false;
</script>

<div class="card {compact ? 'compact' : ''} {fullWidth ? 'full-width' : ''}">
  <slot name="header"></slot>
  <slot></slot>
</div>

<style>
.card {
  background: linear-gradient(135deg, var(--card-bg), rgba(22, 33, 62, 0.8) 100%);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 1rem;
  backdrop-filter: blur(10px);
  transition: all 0.3s ease;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  position: relative;
}
.card.compact {
  padding: 0.75rem 1rem;
}
.card:hover {
  border-color: var(--accent-primary);
  box-shadow: 0 4px 20px rgba(139, 92, 246, 0.15);
  transform: translateY(-1px);
}
.card.full-width {
  grid-column: 1 / -1;
}
</style>
