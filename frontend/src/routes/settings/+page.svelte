<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { orgStore } from '$stores/org';
  import { authStore } from '$stores/auth';
  import { apiClient } from '$api/client';
  import type { WebhookEndpoint, WebhookEventType } from '$types';
  import Modal from '$components/Modal.svelte';
  import Badge from '$components/Badge.svelte';

  let orgName = '';
  let orgSlug = '';
  let isSaving = false;
  let saveError = '';
  let saveSuccess = false;

  let showDeleteModal = false;
  let deleteConfirmation = '';
  let isDeleting = false;
  let deleteError = '';

  let webhooks: WebhookEndpoint[] = [];
  let isLoadingWebhooks = true;

  let showWebhookModal = false;
  let isEditingWebhook = false;
  let editingWebhookId = '';
  let webhookUrl = '';
  let webhookActive = true;
  let webhookError = '';
  let isSavingWebhook = false;

  let selectedEvents: WebhookEventType[] = [];

  let showSecret = false;

  const allEvents: WebhookEventType[] = [
    'org.created',
    'member.invited',
    'member.joined',
    'plan.upgraded',
    'plan.downgraded',
    'subscription.canceled',
    'invoice.paid',
    'quota.exceeded'
  ];

  function formatEventName(event: WebhookEventType): string {
    return event.split('.').map(s => s.charAt(0).toUpperCase() + s.slice(1)).join(' ');
  }

  function isEventSelected(event: WebhookEventType): boolean {
    return selectedEvents.includes(event);
  }

  function toggleEvent(event: WebhookEventType) {
    if (selectedEvents.includes(event)) {
      selectedEvents = selectedEvents.filter(e => e !== event);
    } else {
      selectedEvents = [...selectedEvents, event];
    }
  }

  async function loadWebhooks() {
    isLoadingWebhooks = true;

    try {
      const data = await apiClient.get<WebhookEndpoint[]>('/webhooks');
      if (Array.isArray(data)) {
        webhooks = data;
      } else {
        webhooks = generateMockWebhooks();
      }
    } catch (err) {
      webhooks = generateMockWebhooks();
    } finally {
      isLoadingWebhooks = false;
    }
  }

  function generateMockWebhooks(): WebhookEndpoint[] {
    return [
      {
        id: 'wh_1',
        orgId: 'org1',
        url: 'https://example.com/webhooks/org',
        secret: 'whsec_abc123def456',
        active: true,
        createdAt: new Date('2024-01-15')
      },
      {
        id: 'wh_2',
        orgId: 'org1',
        url: 'https://api.example.com/billing',
        secret: 'whsec_xyz789uvw012',
        active: false,
        createdAt: new Date('2024-02-20')
      }
    ];
  }

  function initForm() {
    if ($orgStore.currentOrg) {
      orgName = $orgStore.currentOrg.name;
      orgSlug = $orgStore.currentOrg.slug;
    }
  }

  async function handleSaveOrg() {
    if (!orgName.trim()) {
      saveError = 'Organization name is required';
      return;
    }

    isSaving = true;
    saveError = '';
    saveSuccess = false;

    try {
      const updated = await orgStore.updateOrg($orgStore.currentOrgId!, {
        name: orgName.trim(),
        slug: orgSlug.trim() || undefined
      });
      orgName = updated.name;
      orgSlug = updated.slug;
      saveSuccess = true;
      setTimeout(() => (saveSuccess = false), 3000);
    } catch (err) {
      saveError = err instanceof Error ? err.message : 'Failed to save changes';
    } finally {
      isSaving = false;
    }
  }

  async function handleDeleteOrg() {
    if (deleteConfirmation !== orgName) {
      deleteError = 'Please type the organization name to confirm';
      return;
    }

    isDeleting = true;
    deleteError = '';

    try {
      await orgStore.deleteOrg($orgStore.currentOrgId!);
      showDeleteModal = false;
      await goto('/orgs');
    } catch (err) {
      deleteError = err instanceof Error ? err.message : 'Failed to delete organization';
    } finally {
      isDeleting = false;
    }
  }

  function openAddWebhook() {
    isEditingWebhook = false;
    editingWebhookId = '';
    webhookUrl = '';
    webhookActive = true;
    webhookError = '';
    selectedEvents = [];
    showWebhookModal = true;
  }

  function openEditWebhook(webhook: WebhookEndpoint) {
    isEditingWebhook = true;
    editingWebhookId = webhook.id;
    webhookUrl = webhook.url;
    webhookActive = webhook.active;
    webhookError = '';
    selectedEvents = [...allEvents];
    showWebhookModal = true;
  }

  async function handleSaveWebhook() {
    if (!webhookUrl.trim()) {
      webhookError = 'Webhook URL is required';
      return;
    }

    try {
      const urlRegex = /^https?:\/\/.+/;
      if (!urlRegex.test(webhookUrl)) {
        webhookError = 'Please enter a valid URL starting with http:// or https://';
        return;
      }
    } catch (e) {
    }

    isSavingWebhook = true;
    webhookError = '';

    try {
      if (isEditingWebhook) {
        await apiClient.patch(`/webhooks/${editingWebhookId}`, {
          url: webhookUrl,
          active: webhookActive
        });
        const index = webhooks.findIndex(w => w.id === editingWebhookId);
        if (index !== -1) {
          webhooks[index] = { ...webhooks[index], url: webhookUrl, active: webhookActive };
          webhooks = [...webhooks];
        }
      } else {
        const newWebhook: WebhookEndpoint = {
          id: `wh_${Date.now()}`,
          orgId: $orgStore.currentOrgId || '',
          url: webhookUrl,
          secret: `whsec_${Math.random().toString(36).substring(2, 15)}`,
          active: webhookActive,
          createdAt: new Date()
        };
        await apiClient.post('/webhooks', { url: webhookUrl, events: selectedEvents });
        webhooks = [...webhooks, newWebhook];
      }
      showWebhookModal = false;
    } catch (err) {
      webhookError = err instanceof Error ? err.message : 'Failed to save webhook';
      const newWebhook: WebhookEndpoint = {
        id: `wh_${Date.now()}`,
        orgId: $orgStore.currentOrgId || '',
        url: webhookUrl,
        secret: `whsec_${Math.random().toString(36).substring(2, 15)}`,
        active: webhookActive,
        createdAt: new Date()
      };
      if (!isEditingWebhook) {
        webhooks = [...webhooks, newWebhook];
      } else {
        const index = webhooks.findIndex(w => w.id === editingWebhookId);
        if (index !== -1) {
          webhooks[index] = { ...webhooks[index], url: webhookUrl, active: webhookActive };
          webhooks = [...webhooks];
        }
      }
      showWebhookModal = false;
    } finally {
      isSavingWebhook = false;
    }
  }

  async function handleDeleteWebhook(webhookId: string) {
    try {
      await apiClient.delete(`/webhooks/${webhookId}`);
    } catch (err) {
    }
    webhooks = webhooks.filter(w => w.id !== webhookId);
  }

  async function toggleWebhookActive(webhook: WebhookEndpoint) {
    try {
      await apiClient.patch(`/webhooks/${webhook.id}`, { active: !webhook.active });
    } catch (err) {
    }
    webhook.active = !webhook.active;
    webhooks = [...webhooks];
  }

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text);
  }

  function formatDate(date: Date | string): string {
    const d = new Date(date);
    return d.toLocaleDateString('en-US', { 
      year: 'numeric', 
      month: 'short', 
      day: 'numeric' 
    });
  }

  onMount(async () => {
    if (!$authStore.isAuthenticated) {
      await goto('/login');
      return;
    }
    if ($orgStore.orgs.length === 0) {
      await orgStore.fetchOrgs();
    }
    if ($orgStore.currentOrg) {
      initForm();
      await loadWebhooks();
    }
  });
</script>

<div class="mx-auto max-w-4xl px-4 sm:px-6 lg:px-8">
  <div class="mb-8">
    <h1 class="text-2xl font-bold text-gray-900">Settings</h1>
    <p class="mt-1 text-sm text-gray-500">Manage your organization settings and preferences.</p>
  </div>

  <div class="space-y-6">
    <div class="card">
      <div class="px-6 py-5 border-b border-gray-200">
        <h3 class="text-lg font-medium text-gray-900">Organization Information</h3>
        <p class="mt-1 text-sm text-gray-500">Update your organization name and slug.</p>
      </div>
      <div class="px-6 py-5">
        {#if saveSuccess}
          <div class="mb-4 rounded-md bg-green-50 p-3 border border-green-200">
            <p class="text-sm text-green-700">Changes saved successfully!</p>
          </div>
        {/if}
        {#if saveError}
          <div class="mb-4 rounded-md bg-red-50 p-3 border border-red-200">
            <p class="text-sm text-red-700">{saveError}</p>
          </div>
        {/if}

        <form on:submit|preventDefault={handleSaveOrg} class="space-y-4">
          <div>
            <label for="orgName" class="block text-sm font-medium text-gray-700">Organization Name</label>
            <input
              id="orgName"
              type="text"
              class="input-field mt-1"
              bind:value={orgName}
              placeholder="Your organization name"
            />
          </div>
          <div>
            <label for="orgSlug" class="block text-sm font-medium text-gray-700">Slug</label>
            <div class="mt-1 flex rounded-md shadow-sm">
              <span class="inline-flex items-center px-3 rounded-l-md border border-r-0 border-gray-300 bg-gray-50 text-gray-500 text-sm">
                app.example.com/
              </span>
              <input
                id="orgSlug"
                type="text"
                class="flex-1 min-w-0 block w-full rounded-none rounded-r-md border-gray-300 focus:ring-primary-500 focus:border-primary-500 sm:text-sm"
                bind:value={orgSlug}
                placeholder="your-org"
              />
            </div>
            <p class="mt-1 text-xs text-gray-500">The unique identifier for your organization.</p>
          </div>
          <div class="pt-4">
            <button type="submit" class="btn-primary" disabled={isSaving}>
              {#if isSaving}
                <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Saving...
              {:else}
                Save Changes
              {/if}
            </button>
          </div>
        </form>
      </div>
    </div>

    <div class="card">
      <div class="px-6 py-5 border-b border-gray-200">
        <div class="flex items-center justify-between">
          <div>
            <h3 class="text-lg font-medium text-gray-900">Webhook Endpoints</h3>
            <p class="mt-1 text-sm text-gray-500">Manage webhooks to receive event notifications.</p>
          </div>
          <button class="btn-primary" on:click={openAddWebhook}>
            <svg class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            Add Webhook
          </button>
        </div>
      </div>
      <div class="divide-y divide-gray-200">
        {#if isLoadingWebhooks}
          <div class="py-12 text-center">
            <svg class="animate-spin h-6 w-6 text-primary-600 mx-auto" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            <p class="mt-4 text-gray-500 text-sm">Loading webhooks...</p>
          </div>
        {:else if webhooks.length === 0}
          <div class="py-12 text-center">
            <div class="mx-auto h-12 w-12 rounded-full bg-gray-100 flex items-center justify-center">
              <svg class="h-6 w-6 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
              </svg>
            </div>
            <h3 class="mt-4 text-sm font-medium text-gray-900">No webhooks yet</h3>
            <p class="mt-1 text-sm text-gray-500">Add a webhook endpoint to receive event notifications.</p>
          </div>
        {:else}
          {#each webhooks as webhook (webhook.id)}
            <div class="px-6 py-4 hover:bg-gray-50">
              <div class="flex items-center justify-between">
                <div class="flex-1 min-w-0">
                  <div class="flex items-center space-x-3">
                    <p class="text-sm font-medium text-gray-900 truncate">{webhook.url}</p>
                    <Badge variant={webhook.active ? 'success' : 'default'}>
                      {webhook.active ? 'Active' : 'Inactive'}
                    </Badge>
                  </div>
                  <div class="mt-2 flex items-center space-x-4">
                    <div class="flex items-center">
                      <span class="text-xs text-gray-500 mr-2">Signing secret:</span>
                      <code class="text-xs bg-gray-100 px-2 py-1 rounded font-mono">
                        {showSecret ? webhook.secret : '••••••••••••••••'}
                      </code>
                      <button
                        class="ml-2 text-gray-400 hover:text-gray-600"
                        on:click={() => copyToClipboard(webhook.secret || '')}
                        title="Copy to clipboard"
                      >
                        <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
                        </svg>
                      </button>
                    </div>
                  </div>
                </div>
                <div class="ml-4 flex items-center space-x-2">
                  <button
                    class="text-gray-400 hover:text-gray-600"
                    on:click={() => toggleWebhookActive(webhook)}
                    title={webhook.active ? 'Disable' : 'Enable'}
                  >
                    {#if webhook.active}
                      <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    {:else}
                      <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    {/if}
                  </button>
                  <button
                    class="text-gray-400 hover:text-primary-600"
                    on:click={() => openEditWebhook(webhook)}
                    title="Edit"
                  >
                    <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                    </svg>
                  </button>
                  <button
                    class="text-gray-400 hover:text-red-600"
                    on:click={() => handleDeleteWebhook(webhook.id)}
                    title="Delete"
                  >
                    <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          {/each}
        {/if}
      </div>
      <div class="px-6 py-4 bg-gray-50 border-t border-gray-200">
        <button
          class="text-sm text-gray-500 hover:text-gray-700"
          on:click={() => (showSecret = !showSecret)}
        >
          {showSecret ? 'Hide' : 'Show'} signing secrets
        </button>
      </div>
    </div>

    <div class="card border-red-200">
      <div class="px-6 py-5 border-b border-red-200 bg-red-50">
        <h3 class="text-lg font-medium text-red-900">Danger Zone</h3>
        <p class="mt-1 text-sm text-red-700">
          Deleting your organization is permanent and cannot be undone.
        </p>
      </div>
      <div class="px-6 py-5">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-gray-900">Delete Organization</p>
            <p class="text-sm text-gray-500 mt-1">
              Permanently delete this organization and all its data.
            </p>
          </div>
          <button
            class="inline-flex items-center justify-center rounded-md border border-red-300 bg-white px-4 py-2 text-sm font-medium text-red-700 shadow-sm hover:bg-red-50 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 transition-colors"
            on:click={() => {
              showDeleteModal = true;
              deleteConfirmation = '';
              deleteError = '';
            }}
          >
            Delete Organization
          </button>
        </div>
      </div>
    </div>
  </div>
</div>

<Modal title={isEditingWebhook ? 'Edit Webhook' : 'Add Webhook'} bind:show={showWebhookModal} size="lg">
  <form on:submit|preventDefault={handleSaveWebhook} class="space-y-4">
    {#if webhookError}
      <div class="rounded-md bg-red-50 p-3 border border-red-200">
        <p class="text-sm text-red-700">{webhookError}</p>
      </div>
    {/if}
    
    <div>
      <label for="webhookUrl" class="block text-sm font-medium text-gray-700">Webhook URL</label>
      <input
        id="webhookUrl"
        type="url"
        required
        class="input-field mt-1"
        bind:value={webhookUrl}
        placeholder="https://example.com/webhooks"
        autofocus
      />
      <p class="mt-1 text-xs text-gray-500">We'll send event POST requests to this URL.</p>
    </div>

    <div class="flex items-center">
      <input
        id="webhookActive"
        type="checkbox"
        class="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
        bind:checked={webhookActive}
      />
      <label for="webhookActive" class="ml-2 block text-sm text-gray-700">
        Active
      </label>
    </div>

    <div>
      <label class="block text-sm font-medium text-gray-700 mb-2">Events to send</label>
      <div class="space-y-2 max-h-48 overflow-y-auto border border-gray-200 rounded-md p-3">
        {#each allEvents as event}
          <div class="flex items-center">
            <input
              id={event}
              type="checkbox"
              class="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
              checked={isEventSelected(event)}
              on:change={() => toggleEvent(event)}
            />
            <label for={event} class="ml-2 block text-sm text-gray-700">
              {formatEventName(event)}
            </label>
          </div>
        {/each}
      </div>
    </div>
  </form>
  <div slot="footer" class="flex justify-end gap-2">
    <button type="button" class="btn-secondary" on:click={() => (showWebhookModal = false)}>
      Cancel
    </button>
    <button type="button" class="btn-primary" disabled={isSavingWebhook} on:click={handleSaveWebhook}>
      {#if isSavingWebhook}
        <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        Saving...
      {:else}
        {isEditingWebhook ? 'Save Changes' : 'Add Webhook'}
      {/if}
    </button>
  </div>
</Modal>

<Modal title="Delete Organization" bind:show={showDeleteModal} size="md">
  <form on:submit|preventDefault={handleDeleteOrg} class="space-y-4">
    {#if deleteError}
      <div class="rounded-md bg-red-50 p-3 border border-red-200">
        <p class="text-sm text-red-700">{deleteError}</p>
      </div>
    {/if}
    
    <p class="text-sm text-gray-600">
      Are you sure you want to delete <span class="font-medium text-gray-900">{orgName}</span>? This action cannot be undone.
    </p>
    
    <div class="rounded-md bg-red-50 p-4 border border-red-200">
      <div class="flex">
        <div class="flex-shrink-0">
          <svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
          </svg>
        </div>
        <div class="ml-3">
          <h4 class="text-sm font-medium text-red-800">This will permanently delete:</h4>
          <ul class="mt-2 text-sm text-red-700 space-y-1">
            <li>• All organization data and settings</li>
            <li>• All members and their access</li>
            <li>• All projects and associated data</li>
            <li>• Billing information and subscription</li>
          </ul>
        </div>
      </div>
    </div>

    <div>
      <label for="deleteConfirm" class="block text-sm font-medium text-gray-700">
        To confirm, type <span class="font-bold">{orgName}</span> below
      </label>
      <input
        id="deleteConfirm"
        type="text"
        required
        class="input-field mt-1"
        bind:value={deleteConfirmation}
        placeholder={orgName}
      />
    </div>
  </form>
  <div slot="footer" class="flex justify-end gap-2">
    <button type="button" class="btn-secondary" on:click={() => (showDeleteModal = false)}>
      Cancel
    </button>
    <button 
      type="button" 
      class="inline-flex items-center justify-center rounded-md border border-transparent bg-red-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 transition-colors"
      disabled={isDeleting || deleteConfirmation !== orgName}
      on:click={handleDeleteOrg}
    >
      {#if isDeleting}
        <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        Deleting...
      {:else}
        Delete Organization
      {/if}
    </button>
  </div>
</Modal>
