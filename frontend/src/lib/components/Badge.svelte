<script lang="ts">
  import { cn } from '$helpers/cn';

  type Variant = 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' | 'gray' | 'purple';
  type Size = 'sm' | 'md';

  export let variant: Variant = 'default';
  export let size: Size = 'sm';
  export let className = '';

  const variantClasses: Record<Variant, string> = {
    default: 'bg-gray-100 text-gray-800',
    primary: 'bg-primary-100 text-primary-800',
    success: 'bg-green-100 text-green-800',
    warning: 'bg-yellow-100 text-yellow-800',
    danger: 'bg-red-100 text-red-800',
    info: 'bg-blue-100 text-blue-800',
    gray: 'bg-gray-100 text-gray-800',
    purple: 'bg-purple-100 text-purple-800'
  };

  const sizeClasses: Record<Size, string> = {
    sm: 'px-2 py-0.5 text-xs',
    md: 'px-2.5 py-1 text-sm'
  };
</script>

<span
  class={cn('inline-flex items-center font-medium rounded-full', variantClasses[variant], sizeClasses[size], className)}
  {...$$restProps}
>
  <slot />
</span>
