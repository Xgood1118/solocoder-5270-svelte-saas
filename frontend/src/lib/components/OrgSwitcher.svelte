<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { orgs, currentOrg, currentOrgId, orgStore } from '$stores/org';
  import type { Org } from '$types';
  import { cn } from '$helpers/cn';

  export let className = '';

  const dispatch = createEventDispatcher();

  let isOpen = false;
  let dropdownRef: HTMLDivElement | null = null;

  function toggleDropdown() {
    isOpen = !isOpen;
  }

  function selectOrg(org: Org) {
    orgStore.setCurrentOrgId(org.id);
    isOpen = false;
    dispatch('change', org);
  }

  function handleCreateOrg() {
    isOpen = false;
    dispatch('createOrg');
  }

  function handleClickOutside(e: MouseEvent) {
    if (dropdownRef && !dropdownRef.contains(e.target as Node)) {
      isOpen = false;
    }
  }

  onMount(() => {
    document.addEventListener('mousedown', handleClickOutside);
  });

  onDestroy(() => {
    document.removeEventListener('mousedown', handleClickOutside);
  });

  function getOrgInitials(name: string): string {
    return name
      .split(' ')
      .map((w) => w[0])
      .join('')
      .toUpperCase()
      .slice(0, 2);
  }
</script>

<div class={cn('relative', className)} bind:this={dropdownRef}>
  <button
    type="button"
    on:click={toggleDropdown}
    class="w-full flex items-center justify-between gap-2 px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
    aria-haspopup="listbox"
    aria-expanded={isOpen}
  >
    <div class="flex items-center gap-2 min-w-0">
      <div class="h-7 w-7 rounded-md bg-primary-100 flex items-center justify-center flex-shrink-0">
        <span class="text-xs font-semibold text-primary-700">
          {#if $currentOrg}
            {getOrgInitials($currentOrg.name)}
          {:else}
            ?
          {/if}
        </span>
      </div>
      <span class="truncate">
        {#if $currentOrg}
          {$currentOrg.name}
        {:else}
          Select organization
        {/if}
      </span>
    </div>
    <svg class="h-4 w-4 text-gray-400 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
    </svg>
  </button>

  {#if isOpen}
    <div class="absolute z-10 w-full mt-1 bg-white border border-gray-200 rounded-md shadow-lg max-h-60 overflow-auto">
      <ul class="py-1" role="listbox">
        {#each $orgs as org (org.id)}
          <li>
            <button
              type="button"
              on:click={() => selectOrg(org)}
              class={cn(
                'w-full flex items-center gap-2 px-3 py-2 text-sm text-left',
                $currentOrgId === org.id
                  ? 'bg-primary-50 text-primary-700'
                  : 'text-gray-700 hover:bg-gray-50'
              )}
              role="option"
              aria-selected={$currentOrgId === org.id}
            >
              <div class="h-7 w-7 rounded-md bg-gray-100 flex items-center justify-center flex-shrink-0">
                <span class="text-xs font-semibold text-gray-600">
                  {getOrgInitials(org.name)}
                </span>
              </div>
              <span class="truncate flex-1">{org.name}</span>
              {#if $currentOrgId === org.id}
                <svg class="h-4 w-4 text-primary-600 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                </svg>
              {/if}
            </button>
          </li>
        {/each}
      </ul>
      <div class="border-t border-gray-200 py-1">
        <button
          type="button"
          on:click={handleCreateOrg}
          class="w-full flex items-center gap-2 px-3 py-2 text-sm text-gray-700 hover:bg-gray-50"
        >
          <svg class="h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
          </svg>
          Create new organization
        </button>
      </div>
    </div>
  {/if}
</div>
