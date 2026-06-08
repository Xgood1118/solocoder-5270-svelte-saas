<script lang="ts">
  import { cn } from '$helpers/cn';

  type Variant = 'primary' | 'secondary' | 'danger' | 'ghost';
  type Size = 'sm' | 'md' | 'lg';

  export let variant: Variant = 'primary';
  export let size: Size = 'md';
  export let loading = false;
  export let disabled = false;
  export let type: 'button' | 'submit' | 'reset' = 'button';
  export let className = '';

  const baseClasses =
    'inline-flex items-center justify-center font-medium rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2';

  const variantClasses: Record<Variant, string> = {
    primary:
      'bg-primary-600 text-white hover:bg-primary-700 focus:ring-primary-500 disabled:bg-primary-300',
    secondary:
      'bg-gray-100 text-gray-700 hover:bg-gray-200 focus:ring-gray-500 disabled:bg-gray-50 disabled:text-gray-400',
    danger:
      'bg-red-600 text-white hover:bg-red-700 focus:ring-red-500 disabled:bg-red-300',
    ghost:
      'text-gray-700 hover:bg-gray-100 focus:ring-gray-500 disabled:text-gray-400'
  };

  const sizeClasses: Record<Size, string> = {
    sm: 'px-3 py-1.5 text-sm',
    md: 'px-4 py-2 text-sm',
    lg: 'px-6 py-3 text-base'
  };
</script>

<button
  {type}
  disabled={disabled || loading}
  class={cn(baseClasses, variantClasses[variant], sizeClasses[size], className)}
  {...$$restProps}
>
  {#if loading}
    <svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
      <path
        class="opacity-75"
        fill="currentColor"
        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
      />
    </svg>
  {/if}
  <slot />
</button>
