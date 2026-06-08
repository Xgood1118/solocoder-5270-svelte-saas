<script context="module" lang="ts">
  export interface Column<T = unknown> {
    key: string;
    title: string;
    sortable?: boolean;
    width?: string;
    className?: string;
    render?: (row: T, index: number) => unknown;
  }

  export interface PaginationInfo {
    page: number;
    perPage: number;
    total: number;
    totalPages: number;
  }
</script>

<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '$helpers/cn';

  export let columns: Column[] = [];
  export let data: unknown[] = [];
  export let loading = false;
  export let pagination: PaginationInfo | null = null;
  export let sortKey: string | null = null;
  export let sortDir: 'asc' | 'desc' = 'asc';
  export let className = '';
  export let emptyText = 'No data available';

  const dispatch = createEventDispatcher();

  function handleSort(key: string) {
    if (sortKey === key) {
      sortDir = sortDir === 'asc' ? 'desc' : 'asc';
    } else {
      sortKey = key;
      sortDir = 'asc';
    }
    dispatch('sort', { key: sortKey, dir: sortDir });
  }

  function handlePageChange(page: number) {
    dispatch('pageChange', page);
  }

  function getCellValue(row: unknown, column: Column, index: number): unknown {
    if (column.render) {
      return column.render(row, index);
    }
    return (row as Record<string, unknown>)[column.key];
  }

  function goToPrevPage() {
    if (pagination && pagination.page > 1) {
      handlePageChange(pagination.page - 1);
    }
  }

  function goToNextPage() {
    if (pagination && pagination.page < pagination.totalPages) {
      handlePageChange(pagination.page + 1);
    }
  }

  function formatValue(value: unknown): string {
    if (value === null || value === undefined) {
      return '-';
    }
    if (typeof value === 'object') {
      return '';
    }
    return String(value);
  }
</script>

<div class={cn('flex flex-col', className)} {...$$restProps}>
  <div class="-my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
    <div class="py-2 align-middle inline-block min-w-full sm:px-6 lg:px-8">
      <div class="shadow overflow-hidden border-b border-gray-200 sm:rounded-lg bg-white">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              {#each columns as column (column.key)}
                <th
                  scope="col"
                  class={cn(
                    'px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider',
                    column.sortable ? 'cursor-pointer select-none hover:bg-gray-100' : '',
                    column.className
                  )}
                  style={column.width ? `width: ${column.width}` : undefined}
                  on:click={() => column.sortable && handleSort(column.key)}
                >
                  <div class="flex items-center gap-1">
                    {column.title}
                    {#if column.sortable}
                      <span class="text-gray-400">
                        {#if sortKey === column.key}
                          {#if sortDir === 'asc'}
                            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
                            </svg>
                          {:else}
                            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                            </svg>
                          {/if}
                        {:else}
                          <svg class="w-4 h-4 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4" />
                          </svg>
                        {/if}
                      </span>
                    {/if}
                  </div>
                </th>
              {/each}
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            {#if loading}
              <tr>
                <td colspan={columns.length} class="px-6 py-12 text-center">
                  <div class="flex flex-col items-center justify-center gap-3">
                    <svg class="animate-spin h-8 w-8 text-primary-600" fill="none" viewBox="0 0 24 24">
                      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                      <path
                        class="opacity-75"
                        fill="currentColor"
                        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                      />
                    </svg>
                    <span class="text-sm text-gray-500">Loading...</span>
                  </div>
                </td>
              </tr>
            {:else if data.length === 0}
              <tr>
                <td colspan={columns.length} class="px-6 py-12 text-center">
                  <div class="flex flex-col items-center justify-center gap-2">
                    <svg class="h-12 w-12 text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
                    </svg>
                    <span class="text-sm text-gray-500">{emptyText}</span>
                  </div>
                </td>
              </tr>
            {:else}
              {#each data as row, rowIndex (rowIndex)}
                <tr class="hover:bg-gray-50">
                  {#each columns as column (column.key)}
                    {@const cellValue = getCellValue(row, column, rowIndex)}
                    <td class={cn('px-6 py-4 whitespace-nowrap text-sm text-gray-900', column.className)}>
                      {#if typeof cellValue === 'object' && cellValue !== null}
                        <slot name="cell" {row} {column} {rowIndex} value={cellValue} />
                      {:else}
                        {formatValue(cellValue)}
                      {/if}
                    </td>
                  {/each}
                </tr>
              {/each}
            {/if}
          </tbody>
        </table>
      </div>
    </div>
  </div>

  {#if pagination && pagination.totalPages > 0}
    <div class="flex items-center justify-between px-4 py-3 bg-white border-t border-gray-200 sm:px-6">
      <div class="flex-1 flex justify-between sm:hidden">
        <button
          type="button"
          on:click={goToPrevPage}
          disabled={pagination.page <= 1}
          class="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Previous
        </button>
        <button
          type="button"
          on:click={goToNextPage}
          disabled={pagination.page >= pagination.totalPages}
          class="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Next
        </button>
      </div>
      <div class="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
        <div>
          <p class="text-sm text-gray-700">
            Showing
            <span class="font-medium"> {(pagination.page - 1) * pagination.perPage + 1} </span>
            to
            <span class="font-medium"> {Math.min(pagination.page * pagination.perPage, pagination.total)} </span>
            of
            <span class="font-medium"> {pagination.total} </span>
            results
          </p>
        </div>
        <div>
          <nav class="relative z-0 inline-flex rounded-md shadow-sm -space-x-px" aria-label="Pagination">
            <button
              type="button"
              on:click={goToPrevPage}
              disabled={pagination.page <= 1}
              class="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              aria-label="Previous"
            >
              <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
              </svg>
            </button>

            {#each Array.from({ length: pagination.totalPages }, (_, i) => i + 1) as pageNum}
              {#if pageNum === 1 || pageNum === pagination.totalPages || (pageNum >= pagination.page - 1 && pageNum <= pagination.page + 1)}
                <button
                  type="button"
                  on:click={() => handlePageChange(pageNum)}
                  class={cn(
                    'relative inline-flex items-center px-4 py-2 border text-sm font-medium',
                    pagination.page === pageNum
                      ? 'z-10 bg-primary-600 border-primary-600 text-white'
                      : 'bg-white border-gray-300 text-gray-500 hover:bg-gray-50'
                  )}
                >
                  {pageNum}
                </button>
              {:else if pageNum === pagination.page - 2 || pageNum === pagination.page + 2}
                <span class="relative inline-flex items-center px-4 py-2 border border-gray-300 bg-white text-sm font-medium text-gray-700">
                  ...
                </span>
              {/if}
            {/each}

            <button
              type="button"
              on:click={goToNextPage}
              disabled={pagination.page >= pagination.totalPages}
              class="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              aria-label="Next"
            >
              <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
            </button>
          </nav>
        </div>
      </div>
    </div>
  {/if}
</div>
