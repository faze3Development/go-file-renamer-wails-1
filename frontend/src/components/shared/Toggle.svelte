<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  export let checked: boolean = false;
  export let disabled: boolean = false;
  // Make `id` and `ariaLabel` have safe defaults so Svelte/TS treat them as optional
  export let id: string = "";
  export let ariaLabel: string = "";
  export let size: 'sm' | 'md' | 'lg' = 'md';

  const dispatch = createEventDispatcher<{ change: { checked: boolean } }>();

  function handleChange(e: Event) {
    const input = e.currentTarget as HTMLInputElement;
    dispatch('change', { checked: input.checked });
  }
</script>

<label class="toggle" aria-label={ariaLabel} aria-disabled={disabled}>
  <input
    type="checkbox"
    role="switch"
    {id}
    {disabled}
    checked={checked}
    on:change={handleChange}
  />
  <span class="slider" data-size={size} aria-hidden="true"></span>
</label>

<style>
  .toggle {
    display: inline-flex;
    align-items: center;
    gap: .5rem;
    cursor: pointer;
  }

  .toggle input {
    position: absolute;
    opacity: 0;
    width: 0;
    height: 0;
  }

  .slider {
    display: inline-block;
    /* subtle muted track when off */
    background: rgba(255,255,255,0.06);
    border-radius: 999px;
    transition: background .2s ease;
    position: relative;
  }

  .slider::after {
    content: "";
    position: absolute;
    top: 50%;
    left: 4px;
    transform: translateY(-50%);
    width: 16px;
    height: 16px;
    background: white;
    border-radius: 50%;
    transition: transform .18s ease;
  }

  .slider[data-size="sm"] { width: 34px; height: 16px; --knob-offset: 18px; }
  .slider[data-size="sm"]::after { width: 12px; height: 12px; left: 2px; }

  .slider[data-size="md"] { width: 42px; height: 22px; --knob-offset: 22px; }
  .slider[data-size="md"]::after { width: 16px; height: 16px; left: 3px; }

  .slider[data-size="lg"] { width: 56px; height: 28px; --knob-offset: 30px; }
  .slider[data-size="lg"]::after { width: 22px; height: 22px; left: 4px; }

  /* move knob and emphasize track when checked */
  input:checked + .slider {
    background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary, var(--accent-primary)));
    box-shadow: 0 6px 18px rgba(139,92,246,0.12), inset 0 -2px 6px rgba(0,0,0,0.08);
  }

  input:checked + .slider::after {
    transform: translateY(-50%) translateX(var(--knob-offset));
    box-shadow: 0 2px 8px rgba(0,0,0,0.2);
  }

  /* disabled look */
  input:disabled + .slider { opacity: 0.6; cursor: not-allowed; }
</style>
