<script lang="ts">
  export let variant: 'primary' | 'secondary' | 'ghost' | 'danger' = 'primary';
  export let disabled: boolean = false;
  export let type: 'button' | 'submit' | 'reset' = 'button';
  export let ariaLabel: string = '';
  export let className: string = '';
</script>

<button
  class={"btn " + variant + (className ? ` ${className}` : '')}
  {disabled}
  {type}
  aria-label={ariaLabel || undefined}
>
  <slot />
</button>

<style>
  .btn {
    padding: 0.65rem 1.1rem;
    border-radius: 10px;
    font-weight: 600;
    cursor: pointer;
    border: none;
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
  }

  .btn:disabled { opacity: 0.6; cursor: not-allowed; }

  .primary {
    background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary));
    color: white;
  }

  .secondary {
    background: rgba(255,255,255,0.06);
    color: var(--text-secondary);
    border: 1px solid var(--border-color);
  }

  .ghost {
    background: transparent;
    color: var(--text-secondary);
    border: 1px solid rgba(255,255,255,0.12);
  }

  .danger {
    background: rgba(239,68,68,0.15);
    color: var(--error-color);
    border: 1px solid var(--error-color);
  }
</style>
