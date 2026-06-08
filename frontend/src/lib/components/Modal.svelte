<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';

  export let show = false;
  export let title = '';
  export let size: 'sm' | 'md' | 'lg' | 'xl' = 'md';
  export let closeOnOverlay = true;
  export let closeOnEsc = true;

  const sizeClasses: Record<string, string> = {
    sm: 'max-w-sm',
    md: 'max-w-md',
    lg: 'max-w-lg',
    xl: 'max-w-xl'
  };

  const dispatch = createEventDispatcher();

  function handleOverlayClick() {
    if (closeOnOverlay) {
      closeModal();
    }
  }

  function handleOverlayKeyDown(e: KeyboardEvent) {
    if (closeOnOverlay && (e.key === 'Enter' || e.key === ' ')) {
      e.preventDefault();
      closeModal();
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (closeOnEsc && e.key === 'Escape' && show) {
      closeModal();
    }
  }

  function closeModal() {
    show = false;
    dispatch('close');
  }

  onMount(() => {
    document.addEventListener('keydown', handleKeyDown);
  });

  onDestroy(() => {
    document.removeEventListener('keydown', handleKeyDown);
  });

  $: if (show && typeof document !== 'undefined') {
    document.body.style.overflow = 'hidden';
  } else if (typeof document !== 'undefined') {
    document.body.style.overflow = '';
  }
</script>

{#if show}
  <div class="fixed inset-0 z-50 flex items-center justify-center">
    <div
      class="absolute inset-0 bg-black bg-opacity-50 transition-opacity cursor-pointer"
      role="button"
      tabindex="0"
      aria-label="Close modal"
      on:click={handleOverlayClick}
      on:keydown={handleOverlayKeyDown}
    />
    <div
      class="relative bg-white rounded-lg shadow-xl w-full {sizeClasses[size]} mx-4 transform transition-all"
      role="dialog"
      aria-modal="true"
      aria-labelledby="modal-title"
    >
      <div class="flex items-center justify-between p-4 border-b border-gray-200">
        <h3 id="modal-title" class="text-lg font-semibold text-gray-900">
          {#if title}
            {title}
          {:else}
            <slot name="title" />
          {/if}
        </h3>
        <button
          on:click={closeModal}
          class="text-gray-400 hover:text-gray-600 transition-colors"
          aria-label="Close"
        >
          <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
      <div class="p-4">
        <slot name="content">
          <slot />
        </slot>
      </div>
      <div class="flex justify-end gap-2 p-4 border-t border-gray-200">
        <slot name="footer" />
      </div>
    </div>
  </div>
{/if}
