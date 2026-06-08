<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { orgStore } from '$stores/org';
  import { authStore } from '$stores/auth';
  import { apiClient } from '$api/client';
  import type { AuditLog, AuditLogAction } from '$types';
  import Badge from '$components/Badge.svelte';

  let auditLogs: AuditLog[] = [];
  let isLoading = true;
  let error = '';
  
  let currentPage = 1;
  let totalPages = 5;
  let perPage = 10;
  let total = 50;

  let selectedAction: AuditLogAction | '' = '';
  let startDate = '';
  let endDate = '';
  let selectedUserId = '';

  let expandedLogId: string | null = null;

  let visiblePages: number[] = [];

  $: {
    const pages: number[] = [];
    const maxVisible = 5;
    if (totalPages <= maxVisible) {
      for (let i = 1; i <= totalPages; i++) {
        pages.push(i);
      }
    } else if (currentPage <= 3) {
      for (let i = 1; i <= maxVisible; i++) {
        pages.push(i);
      }
    } else if (currentPage >= totalPages - 2) {
      for (let i = totalPages - maxVisible + 1; i <= totalPages; i++) {
        pages.push(i);
      }
    } else {
      for (let i = currentPage - 2; i <= currentPage + 2; i++) {
        pages.push(i);
      }
    }
    visiblePages = pages;
  }

  const actions: (AuditLogAction | 'all')[] = [
    'all',
    'user.login',
    'user.logout',
    'org.created',
    'org.updated',
    'org.deleted',
    'member.invited',
    'member.joined',
    'member.removed',
    'member.role_updated',
    'subscription.created',
    'subscription.updated',
    'subscription.canceled',
    'invoice.paid',
    'invoice.failed',
    'settings.updated'
  ];

  function formatDate(date: Date | string): string {
    const d = new Date(date);
    return d.toLocaleDateString('en-US', { 
      year: 'numeric', 
      month: 'short', 
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function getActionBadgeVariant(action: AuditLogAction): 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' | 'gray' | 'purple' {
    if (action.includes('create') || action.includes('join') || action.includes('paid')) {
      return 'success';
    }
    if (action.includes('delete') || action.includes('remove') || action.includes('failed') || action.includes('cancel')) {
      return 'danger';
    }
    if (action.includes('update') || action.includes('login') || action.includes('logout')) {
      return 'primary';
    }
    return 'default';
  }

  function formatAction(action: AuditLogAction): string {
    return action.split('.').map(s => s.charAt(0).toUpperCase() + s.slice(1)).join(' ');
  }

  function toggleExpand(logId: string) {
    expandedLogId = expandedLogId === logId ? null : logId;
  }

  async function loadAuditLogs() {
    isLoading = true;
    error = '';

    try {
      const params = new URLSearchParams({
        page: currentPage.toString(),
        perPage: perPage.toString()
      });
      if (selectedAction) params.append('action', selectedAction);
      if (startDate) params.append('startDate', startDate);
      if (endDate) params.append('endDate', endDate);
      if (selectedUserId) params.append('userId', selectedUserId);

      const data = await apiClient.get(`/audit?${params.toString()}`);
      
      if (data && Array.isArray((data as any).data)) {
        auditLogs = (data as any).data;
        total = (data as any).pagination?.total || 0;
        totalPages = (data as any).pagination?.totalPages || 1;
      } else if (Array.isArray(data)) {
        auditLogs = data as AuditLog[];
      } else {
        auditLogs = generateMockAuditLogs();
      }
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load audit logs';
      auditLogs = generateMockAuditLogs();
    } finally {
      isLoading = false;
    }
  }

  function generateMockAuditLogs(): AuditLog[] {
    const mockLogs: AuditLog[] = [];
    const actionsList: AuditLogAction[] = [
      'user.login',
      'member.joined',
      'member.invited',
      'member.role_updated',
      'member.removed',
      'org.updated',
      'subscription.updated',
      'invoice.paid',
      'settings.updated'
    ];

    for (let i = 0; i < 10; i++) {
      const action = actionsList[i % actionsList.length];
      mockLogs.push({
        id: `log-${i + 1}`,
        orgId: 'org1',
        userId: `user-${(i % 5) + 1}`,
        user: {
          id: `user-${(i % 5) + 1}`,
          email: `user${(i % 5) + 1}@example.com`,
          name: `User ${(i % 5) + 1}`,
          createdAt: new Date(),
          updatedAt: new Date()
        },
        action,
        targetType: action.split('.')[0],
        targetId: `target-${i}`,
        ipAddress: `192.168.1.${(i % 255) + 1}`,
        userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)',
        metadata: {
          before: { role: 'member' },
          after: { role: 'admin' }
        },
        createdAt: new Date(Date.now() - i * 3600000 * 2)
      });
    }
    return mockLogs;
  }

  function handleFilter() {
    currentPage = 1;
    loadAuditLogs();
  }

  function clearFilters() {
    selectedAction = '';
    startDate = '';
    endDate = '';
    selectedUserId = '';
    currentPage = 1;
    loadAuditLogs();
  }

  function goToPage(page: number) {
    if (page < 1 || page > totalPages) return;
    currentPage = page;
    loadAuditLogs();
  }

  onMount(async () => {
    if (!$authStore.isAuthenticated) {
      await goto('/login');
      return;
    }
    if ($orgStore.orgs.length === 0) {
      await orgStore.fetchOrgs();
    }
    if ($orgStore.currentOrgId) {
      await loadAuditLogs();
    }
  });
</script>

<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
  <div class="mb-8">
    <h1 class="text-2xl font-bold text-gray-900">Audit Log</h1>
    <p class="mt-1 text-sm text-gray-500">View all activity in your organization.</p>
  </div>

  <div class="card mb-6">
    <div class="p-6">
      <h3 class="text-sm font-medium text-gray-900 mb-4">Filters</h3>
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div>
          <label for="action" class="block text-xs font-medium text-gray-500 uppercase">Action</label>
          <select
            id="action"
            class="input-field mt-1"
            bind:value={selectedAction}
            on:change={handleFilter}
          >
            <option value="">All Actions</option>
            {#each actions as action}
              {#if action !== 'all'}
                <option value={action}>{formatAction(action)}</option>
              {/if}
            {/each}
          </select>
        </div>
        <div>
          <label for="startDate" class="block text-xs font-medium text-gray-500 uppercase">Start Date</label>
          <input
            id="startDate"
            type="date"
            class="input-field mt-1"
            bind:value={startDate}
            on:change={handleFilter}
          />
        </div>
        <div>
          <label for="endDate" class="block text-xs font-medium text-gray-500 uppercase">End Date</label>
          <input
            id="endDate"
            type="date"
            class="input-field mt-1"
            bind:value={endDate}
            on:change={handleFilter}
          />
        </div>
        <div class="flex items-end">
          <button
            type="button"
            class="btn-secondary w-full"
            on:click={clearFilters}
          >
            Clear Filters
          </button>
        </div>
      </div>
    </div>
  </div>

  <div class="card">
    {#if isLoading}
      <div class="py-16 text-center">
        <svg class="animate-spin h-8 w-8 text-primary-600 mx-auto" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <p class="mt-4 text-gray-500">Loading audit logs...</p>
      </div>
    {:else if auditLogs.length === 0}
      <div class="py-16 text-center">
        <div class="mx-auto h-12 w-12 rounded-full bg-gray-100 flex items-center justify-center">
          <svg class="h-6 w-6 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
        </div>
        <h3 class="mt-4 text-sm font-medium text-gray-900">No audit logs found</h3>
        <p class="mt-1 text-sm text-gray-500">Try adjusting your filters.</p>
      </div>
    {:else}
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Time
              </th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                User
              </th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Action
              </th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Entity
              </th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                IP Address
              </th>
              <th scope="col" class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                Details
              </th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            {#each auditLogs as log (log.id)}
              <tr class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {formatDate(log.createdAt)}
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="flex items-center">
                    <div class="flex-shrink-0 h-8 w-8">
                      <div class="h-8 w-8 rounded-full bg-gray-200 flex items-center justify-center">
                        <span class="text-gray-600 font-medium text-xs">
                          {log.user?.name?.charAt(0) || 'U'}
                        </span>
                      </div>
                    </div>
                    <div class="ml-3">
                      <div class="text-sm font-medium text-gray-900">
                        {log.user?.name || 'Unknown'}
                      </div>
                      <div class="text-xs text-gray-500">{log.user?.email || ''}</div>
                    </div>
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <Badge variant={getActionBadgeVariant(log.action)}>
                    {formatAction(log.action)}
                  </Badge>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-900">{log.targetType}</div>
                  <div class="text-xs text-gray-500 font-mono">{log.targetId}</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {log.ipAddress || '-'}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                  <button
                    class="text-primary-600 hover:text-primary-900"
                    on:click={() => toggleExpand(log.id)}
                  >
                    {expandedLogId === log.id ? 'Hide' : 'View'}
                  </button>
                </td>
              </tr>
              {#if expandedLogId === log.id}
                <tr class="bg-gray-50">
                  <td colspan="6" class="px-6 py-4">
                    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
                      <div>
                        <h4 class="text-xs font-medium text-gray-500 uppercase mb-2">Metadata</h4>
                        <pre class="text-xs bg-gray-100 p-3 rounded overflow-x-auto">
{JSON.stringify(log.metadata, null, 2)}
                        </pre>
                      </div>
                      <div>
                        <h4 class="text-xs font-medium text-gray-500 uppercase mb-2">User Agent</h4>
                        <p class="text-sm text-gray-600 break-words">{log.userAgent || '-'}</p>
                        <h4 class="text-xs font-medium text-gray-500 uppercase mt-4 mb-2">Log ID</h4>
                        <p class="text-sm text-gray-600 font-mono">{log.id}</p>
                      </div>
                    </div>
                  </td>
                </tr>
              {/if}
            {/each}
          </tbody>
        </table>
      </div>

      <div class="px-6 py-4 border-t border-gray-200 flex items-center justify-between">
        <div class="flex-1 flex justify-between sm:hidden">
          <button
            class="btn-secondary"
            on:click={() => goToPage(currentPage - 1)}
            disabled={currentPage === 1}
          >
            Previous
          </button>
          <button
            class="btn-secondary"
            on:click={() => goToPage(currentPage + 1)}
            disabled={currentPage === totalPages}
          >
            Next
          </button>
        </div>
        <div class="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
          <div>
            <p class="text-sm text-gray-700">
              Showing <span class="font-medium">{(currentPage - 1) * perPage + 1}</span> to{' '}
              <span class="font-medium">{Math.min(currentPage * perPage, total)}</span> of{' '}
              <span class="font-medium">{total}</span> results
            </p>
          </div>
          <div>
            <nav class="relative z-0 inline-flex rounded-md shadow-sm -space-x-px">
              <button
                class="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50"
                on:click={() => goToPage(currentPage - 1)}
                disabled={currentPage === 1}
              >
                <span class="sr-only">Previous</span>
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
                </svg>
              </button>
              {#each visiblePages as pageNum}
                <button
                  class="relative inline-flex items-center px-4 py-2 border text-sm font-medium {
                    currentPage === pageNum
                      ? 'z-10 bg-primary-50 border-primary-500 text-primary-600'
                      : 'bg-white border-gray-300 text-gray-500 hover:bg-gray-50'
                  }"
                  on:click={() => goToPage(pageNum)}
                >
                  {pageNum}
                </button>
              {/each}
              <button
                class="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50"
                on:click={() => goToPage(currentPage + 1)}
                disabled={currentPage === totalPages}
              >
                <span class="sr-only">Next</span>
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
</div>
