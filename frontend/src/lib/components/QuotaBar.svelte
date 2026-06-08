<script lang="ts">
  import { cn } from '$helpers/cn';

  export let used: number;
  export let total: number;
  export let label = '';
  export let showPercentage = true;
  export let className = '';

  $: percentage = total > 0 ? Math.min((used / total) * 100, 100) : 0;

  $: barColor =
    percentage >= 100
      ? 'bg-red-500'
      : percentage >= 80
        ? 'bg-yellow-500'
        : 'bg-green-500';

  $: textColor =
    percentage >= 100
      ? 'text-red-700'
      : percentage >= 80
        ? 'text-yellow-700'
        : 'text-green-700';

  function formatNumber(num: number): string {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M';
    }
    if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'K';
    }
    return num.toString();
  }
</script>

<div class={cn('w-full', className)} {...$$restProps}>
  {#if label}
    <div class="flex justify-between items-center mb-1">
      <span class="text-sm font-medium text-gray-700">{label}</span>
      <span class={cn('text-sm font-medium', textColor)}>
        {formatNumber(used)} / {formatNumber(total)}
        {#if showPercentage}
          <span class="text-gray-500 ml-1">({percentage.toFixed(0)}%)</span>
        {/if}
      </span>
    </div>
  {/if}
  <div class="w-full bg-gray-200 rounded-full h-2 overflow-hidden">
    <div
      class={cn('h-full rounded-full transition-all duration-300', barColor)}
      style="width: {percentage}%"
      role="progressbar"
      aria-valuenow={used}
      aria-valuemin={0}
      aria-valuemax={total}
    />
  </div>
</div>
