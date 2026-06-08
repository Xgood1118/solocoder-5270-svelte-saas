<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { orgStore } from '$stores/org';
  import { authStore } from '$stores/auth';
  import Modal from '$components/Modal.svelte';
  import Badge from '$components/Badge.svelte';

  let showCreateModal = false;
  let newOrgName = '';
  let newOrgSlug = '';
  let createError = '';
  let isCreating = false;

  function getPlanBadgeVariant(plan: string): 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' | 'gray' | 'purple' {
    switch (plan) {
      case 'enterprise':
        return 'purple';
      case 'team':
      case 'pro':
        return 'primary';
      default:
        return 'default';
    }
  }

  function formatPlanName(plan: string): string {
    return plan.charAt(0).toUpperCase() + plan.slice(1);
  }

  async function handleCreateOrg() {
    if (!newOrgName.trim()) {
      createError = 'Organization name is required';
      return;
    }

    isCreating = true;
    createError = '';

    try {
      const org = await orgStore.createOrg({ 
        name: newOrgName.trim(), 
        slug: newOrgSlug.trim() || undefined 
      });
      orgStore.setCurrentOrgId(org.id);
      showCreateModal = false;
      newOrgName = '';
      newOrgSlug = '';
      await goto('/dashboard');
    } catch (err) {
      createError = err instanceof Error ? err.message : 'Failed to create organization';
    } finally {
      isCreating = false;
    }
  }

  async function selectOrg(orgId: string) {
    orgStore.setCurrentOrgId(orgId);
    await goto('/dashboard');
  }

  function handleModalClose() {
    showCreateModal = false;
    createError = '';
    newOrgName = '';
    newOrgSlug = '';
  }

  onMount(async () => {
    if (!$authStore.isAuthenticated) {
      await goto('/login');
      return;
    }
    if ($orgStore.orgs.length === 0) {
      await orgStore.fetchOrgs();
    }
  });
</script>

<div class="min-h-screen bg-gray-50">
  <div class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
    <div class="mb-8 flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Organizations</h1>
        <p class="mt-1 text-sm text-gray-500">Manage your organizations and switch between them.</p>
      </div>
      <button class="btn-primary" on:click={() => (showCreateModal = true)}>
        <svg class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        Create Organization
      </button>
    </div>

    {#if $orgStore.isLoading}
      <div class="flex justify-center py-16">
        <div class="text-center">
          <svg class="animate-spin h-8 w-8 text-primary-600 mx-auto" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          <p class="mt-4 text-gray-500">Loading organizations...</p>
        </div>
      </div>
    {:else if $orgStore.error}
      <div class="rounded-lg bg-red-50 p-6 border border-red-200">
        <p class="text-red-700">{$orgStore.error}</p>
        <button class="btn-secondary mt-4" on:click={() => orgStore.fetchOrgs()}>
          Retry
        </button>
      </div>
    {:else if $orgStore.orgs.length === 0}
      <div class="text-center py-16">
        <div class="mx-auto h-16 w-16 rounded-full bg-gray-100 flex items-center justify-center">
          <svg class="h-8 w-8 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
          </svg>
        </div>
        <h2 class="mt-4 text-lg font-medium text-gray-900">No organizations yet</h2>
        <p class="mt-2 text-gray-500 max-w-md mx-auto">Create your first organization to get started. You'll be the owner and can invite team members.</p>
        <button class="btn-primary mt-6" on:click={() => (showCreateModal = true)}>
          Create Organization
        </button>
      </div>
    {:else}
      <div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {#each $orgStore.orgs as org (org.id)}
          <div
            class="card cursor-pointer p-6 transition-all hover:shadow-md hover:border-primary-300 group"
            on:click={() => selectOrg(org.id)}
          >
            <div class="flex items-start justify-between">
              <div class="flex items-center space-x-3">
                <div class="flex h-12 w-12 items-center justify-center rounded-lg bg-primary-100 text-primary-700 font-bold text-lg group-hover:bg-primary-200 transition-colors">
                  {org.name.charAt(0).toUpperCase()}
                </div>
                <div>
                  <h3 class="font-semibold text-gray-900 group-hover:text-primary-700 transition-colors">{org.name}</h3>
                  <p class="text-sm text-gray-500">{org.slug}</p>
                </div>
              </div>
            </div>
            <div class="mt-4 flex items-center justify-between">
              <Badge variant={getPlanBadgeVariant(org.plan)}>
                {formatPlanName(org.plan)} Plan
              </Badge>
              <Badge variant="success">
                Owner
              </Badge>
            </div>
            <div class="mt-4 pt-4 border-t border-gray-100">
              <p class="text-xs text-gray-400">
                Created {new Date(org.createdAt).toLocaleDateString()}
              </p>
            </div>
          </div>
        {/each}

        <div
          class="card p-6 border-2 border-dashed border-gray-300 hover:border-primary-400 cursor-pointer transition-colors flex flex-col items-center justify-center min-h-[160px]"
          on:click={() => (showCreateModal = true)}
        >
          <div class="h-12 w-12 rounded-full bg-gray-100 flex items-center justify-center">
            <svg class="h-6 w-6 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
          </div>
          <p class="mt-3 text-sm font-medium text-gray-700">Create new organization</p>
        </div>
      </div>
    {/if}
  </div>

  <Modal title="Create Organization" bind:show={showCreateModal} on:close={handleModalClose}>
    <form on:submit|preventDefault={handleCreateOrg} class="space-y-4">
      {#if createError}
        <div class="rounded-md bg-red-50 p-3 border border-red-200">
          <p class="text-sm text-red-700">{createError}</p>
        </div>
      {/if}
      <div>
        <label for="orgName" class="block text-sm font-medium text-gray-700">Organization name</label>
        <input
          id="orgName"
          type="text"
          required
          class="input-field mt-1"
          bind:value={newOrgName}
          placeholder="Acme Inc."
          autofocus
        />
      </div>
      <div>
        <label for="orgSlug" class="block text-sm font-medium text-gray-700">
          Slug
          <span class="text-gray-400 font-normal">(optional)</span>
        </label>
        <input
          id="orgSlug"
          type="text"
          class="input-field mt-1"
          bind:value={newOrgSlug}
          placeholder="acme"
        />
        <p class="mt-1 text-xs text-gray-500">A unique identifier for your organization</p>
      </div>
    </form>
    <div slot="footer" class="flex justify-end gap-2">
      <button type="button" class="btn-secondary" on:click={handleModalClose}>
        Cancel
      </button>
      <button type="button" class="btn-primary" disabled={isCreating} on:click={handleCreateOrg}>
        {#if isCreating}
          <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          Creating...
        {:else}
          Create Organization
        {/if}
      </button>
    </div>
  </Modal>
</div>
