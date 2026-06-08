<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '$helpers/cn';

  type AlertType = 'success' | 'error' | 'warning' | 'info';

  export let type: AlertType = 'info';
  export let dismissible = false;
  export let visible = true;
  export let className = '';

  const dispatch = createEventDispatcher();

  const typeClasses: Record<AlertType, string> = {
    success: 'bg-green-50 text-green-800 border-green-200',
    error: 'bg-red-50 text-red-800 border-red-200',
    warning: 'bg-yellow-50 text-yellow-800 border-yellow-200',
    info: 'bg-blue-50 text-blue-800 border-blue-200'
  };

  const iconPaths: Record<AlertType, string> = {
    success:
      'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
    error:
      'M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z',
    warning:
      'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z',
    info:
      'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z'
  };

  function dismiss() {
    visible = false;
    dispatch('close');
  }
</script>

{#if visible}
  <div
    class={cn('border rounded-md p-4 flex items-start', typeClasses[type], className)}
    role="alert"
    {...$$restProps}
  >
    <svg class="h-5 w-5 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
      <path stroke-linecap="round" stroke-linejoin="round" d={iconPaths[type]} />
    </svg>
    <div class="ml-3 flex-1">
      <slot />
    </div>
    {#if dismissible}
      <button on:click={dismiss} class="ml-4 flex-shrink-0 opacity-60 hover:opacity-100" aria-label="Dismiss">
        <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    {/if}
  </div>
{/if}
