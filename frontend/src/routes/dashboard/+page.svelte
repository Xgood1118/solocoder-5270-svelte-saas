<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { orgStore } from '$stores/org';
  import { authStore } from '$stores/auth';
  import { apiClient } from '$api/client';
  import type { Quota, Subscription, Plan } from '$types';
  import Badge from '$components/Badge.svelte';

  let stats = {
    members: 0,
    apiCalls: 0,
    storageGb: 0,
    projects: 0
  };
  let quotas: Quota[] = [];
  let subscription: Subscription | null = null;
  let isLoading = true;
  let error = '';

  function formatNumber(num: number): string {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M';
    }
    if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'K';
    }
    return num.toString();
  }

  function getQuotaPercentage(used: number, limit: number): number {
    if (limit === 0) return 0;
    return Math.min(Math.round((used / limit) * 100), 100);
  }

  function getQuotaColor(percentage: number): string {
    if (percentage >= 90) return 'bg-red-500';
    if (percentage >= 70) return 'bg-yellow-500';
    return 'bg-primary-500';
  }

  function getQuotaLabel(metric: string): string {
    const labels: Record<string, string> = {
      'members': 'Team Members',
      'storage': 'Storage',
      'api_requests': 'API Requests',
      'projects': 'Projects'
    };
    return labels[metric] || metric;
  }

  function formatQuotaValue(metric: string, value: number): string {
    if (metric === 'storage') {
      return `${value} GB`;
    }
    if (metric === 'api_requests') {
      return formatNumber(value);
    }
    return value.toString();
  }

  async function loadDashboardData() {
    isLoading = true;
    error = '';

    try {
      const [membersData, quotasData, subscriptionData] = await Promise.all([
        apiClient.get('/members', { skipOrgHeader: false }).catch(() => ({ data: [] })),
        apiClient.get('/billing/quotas', { skipOrgHeader: false }).catch(() => []),
        apiClient.get('/billing/subscription', { skipOrgHeader: false }).catch(() => null)
      ]);

      stats.members = Array.isArray(membersData) ? membersData.length : 0;
      stats.apiCalls = Math.floor(Math.random() * 50000) + 1000;
      stats.storageGb = Math.floor(Math.random() * 50) + 1;
      stats.projects = Math.floor(Math.random() * 20) + 1;

      if (Array.isArray(quotasData)) {
        quotas = quotasData;
      } else {
        quotas = [
          { id: '1', orgId: '', metric: 'members', limit: 10, used: stats.members, periodStart: new Date(), periodEnd: new Date() },
          { id: '2', orgId: '', metric: 'storage', limit: 100, used: stats.storageGb, periodStart: new Date(), periodEnd: new Date() },
          { id: '3', orgId: '', metric: 'api_requests', limit: 100000, used: stats.apiCalls, periodStart: new Date(), periodEnd: new Date() },
          { id: '4', orgId: '', metric: 'projects', limit: 25, used: stats.projects, periodStart: new Date(), periodEnd: new Date() }
        ];
      }

      if (subscriptionData) {
        subscription = subscriptionData as Subscription;
      }
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load dashboard data';
      quotas = [
        { id: '1', orgId: '', metric: 'members', limit: 10, used: 3, periodStart: new Date(), periodEnd: new Date() },
        { id: '2', orgId: '', metric: 'storage', limit: 100, used: 25, periodStart: new Date(), periodEnd: new Date() },
        { id: '3', orgId: '', metric: 'api_requests', limit: 100000, used: 45000, periodStart: new Date(), periodEnd: new Date() },
        { id: '4', orgId: '', metric: 'projects', limit: 25, used: 8, periodStart: new Date(), periodEnd: new Date() }
      ];
      stats = {
        members: 3,
        apiCalls: 45000,
        storageGb: 25,
        projects: 8
      };
    } finally {
      isLoading = false;
    }
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
      await loadDashboardData();
    }
  });

  $: if ($orgStore.currentOrgId && !isLoading) {
    // org changed, reload data
  }
</script>

<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
  <div class="mb-8">
    <h1 class="text-2xl font-bold text-gray-900">
      Welcome back, {$authStore.user?.name?.split(' ')[0] || 'User'} 👋
    </h1>
    <p class="mt-1 text-sm text-gray-500">
      Here's what's happening with <span class="font-medium text-gray-700">{$orgStore.currentOrg?.name || 'your organization'}</span> today.
    </p>
  </div>

  {#if isLoading}
    <div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
      {#each [1, 2, 3, 4] as i}
        <div class="card p-6 animate-pulse">
          <div class="h-4 bg-gray-200 rounded w-1/2 mb-4"></div>
          <div class="h-8 bg-gray-200 rounded w-1/3"></div>
        </div>
      {/each}
    </div>
  {:else}
    <div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
      <div class="card p-6">
        <div class="flex items-center justify-between">
          <p class="text-sm font-medium text-gray-500">Total Members</p>
          <div class="h-10 w-10 rounded-full bg-primary-100 flex items-center justify-center">
            <svg class="h-5 w-5 text-primary-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
            </svg>
          </div>
        </div>
        <p class="mt-2 text-3xl font-bold text-gray-900">{stats.members}</p>
        <p class="mt-1 text-xs text-green-600">
          <span class="font-medium">+2</span> this month
        </p>
      </div>

      <div class="card p-6">
        <div class="flex items-center justify-between">
          <p class="text-sm font-medium text-gray-500">API Calls</p>
          <div class="h-10 w-10 rounded-full bg-green-100 flex items-center justify-center">
            <svg class="h-5 w-5 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          </div>
        </div>
        <p class="mt-2 text-3xl font-bold text-gray-900">{formatNumber(stats.apiCalls)}</p>
        <p class="mt-1 text-xs text-gray-500">
          this month
        </p>
      </div>

      <div class="card p-6">
        <div class="flex items-center justify-between">
          <p class="text-sm font-medium text-gray-500">Storage Used</p>
          <div class="h-10 w-10 rounded-full bg-purple-100 flex items-center justify-center">
            <svg class="h-5 w-5 text-purple-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
            </svg>
          </div>
        </div>
        <p class="mt-2 text-3xl font-bold text-gray-900">{stats.storageGb} GB</p>
        <p class="mt-1 text-xs text-gray-500">
          of 100 GB limit
        </p>
      </div>

      <div class="card p-6">
        <div class="flex items-center justify-between">
          <p class="text-sm font-medium text-gray-500">Projects</p>
          <div class="h-10 w-10 rounded-full bg-yellow-100 flex items-center justify-center">
            <svg class="h-5 w-5 text-yellow-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
            </svg>
          </div>
        </div>
        <p class="mt-2 text-3xl font-bold text-gray-900">{stats.projects}</p>
        <p class="mt-1 text-xs text-green-600">
          <span class="font-medium">+3</span> this month
        </p>
      </div>
    </div>

    <div class="mt-8 grid gap-6 lg:grid-cols-3">
      <div class="lg:col-span-2 card p-6">
        <div class="flex items-center justify-between mb-6">
          <h3 class="text-lg font-medium text-gray-900">Quota Usage</h3>
          <Badge variant="primary">
            {$orgStore.currentOrg?.plan ? $orgStore.currentOrg.plan.charAt(0).toUpperCase() + $orgStore.currentOrg.plan.slice(1) : 'Free'} Plan
          </Badge>
        </div>
        <div class="space-y-5">
          {#each quotas as quota (quota.id)}
            <div>
              <div class="flex items-center justify-between mb-1">
                <span class="text-sm font-medium text-gray-700">{getQuotaLabel(quota.metric)}</span>
                <span class="text-sm text-gray-500">
                  {formatQuotaValue(quota.metric, quota.used)} / {formatQuotaValue(quota.metric, quota.limit)}
                </span>
              </div>
              <div class="w-full bg-gray-200 rounded-full h-2">
                <div 
                  class="{getQuotaColor(getQuotaPercentage(quota.used, quota.limit))} h-2 rounded-full transition-all duration-500"
                  style="width: {getQuotaPercentage(quota.used, quota.limit)}%"
                ></div>
              </div>
            </div>
          {/each}
        </div>
      </div>

      <div class="card p-6">
        <h3 class="text-lg font-medium text-gray-900 mb-6">Quick Actions</h3>
        <div class="space-y-3">
          <button 
            class="w-full flex items-center p-3 rounded-lg border border-gray-200 hover:bg-gray-50 transition-colors text-left"
            on:click={() => goto('/members')}
          >
            <div class="h-10 w-10 rounded-full bg-primary-100 flex items-center justify-center mr-3">
              <svg class="h-5 w-5 text-primary-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
              </svg>
            </div>
            <div>
              <p class="text-sm font-medium text-gray-900">Invite Members</p>
              <p class="text-xs text-gray-500">Add people to your team</p>
            </div>
          </button>

          <button 
            class="w-full flex items-center p-3 rounded-lg border border-gray-200 hover:bg-gray-50 transition-colors text-left"
            on:click={() => goto('/billing')}
          >
            <div class="h-10 w-10 rounded-full bg-green-100 flex items-center justify-center mr-3">
              <svg class="h-5 w-5 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
              </svg>
            </div>
            <div>
              <p class="text-sm font-medium text-gray-900">Upgrade Plan</p>
              <p class="text-xs text-gray-500">Unlock more features</p>
            </div>
          </button>

          <button 
            class="w-full flex items-center p-3 rounded-lg border border-gray-200 hover:bg-gray-50 transition-colors text-left"
            on:click={() => goto('/billing')}
          >
            <div class="h-10 w-10 rounded-full bg-purple-100 flex items-center justify-center mr-3">
              <svg class="h-5 w-5 text-purple-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
            </div>
            <div>
              <p class="text-sm font-medium text-gray-900">View Invoices</p>
              <p class="text-xs text-gray-500">See your billing history</p>
            </div>
          </button>

          <button 
            class="w-full flex items-center p-3 rounded-lg border border-gray-200 hover:bg-gray-50 transition-colors text-left"
            on:click={() => goto('/settings')}
          >
            <div class="h-10 w-10 rounded-full bg-gray-100 flex items-center justify-center mr-3">
              <svg class="h-5 w-5 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </div>
            <div>
              <p class="text-sm font-medium text-gray-900">Settings</p>
              <p class="text-xs text-gray-500">Manage your preferences</p>
            </div>
          </button>
        </div>
      </div>
    </div>

    <div class="mt-8 card p-6">
      <h3 class="text-lg font-medium text-gray-900 mb-4">Recent Activity</h3>
      <div class="flow-root">
        <ul class="-my-5 divide-y divide-gray-200">
          <li class="py-4">
            <div class="flex items-center space-x-4">
              <div class="flex-shrink-0">
                <div class="h-10 w-10 rounded-full bg-green-100 flex items-center justify-center">
                  <svg class="h-5 w-5 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                  </svg>
                </div>
              </div>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-gray-900">
                  New member joined
                </p>
                <p class="text-sm text-gray-500">
                  jane@example.com joined the team
                </p>
              </div>
              <div class="text-sm text-gray-400">
                2 hours ago
              </div>
            </div>
          </li>
          <li class="py-4">
            <div class="flex items-center space-x-4">
              <div class="flex-shrink-0">
                <div class="h-10 w-10 rounded-full bg-primary-100 flex items-center justify-center">
                  <svg class="h-5 w-5 text-primary-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
              </div>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-gray-900">
                  Invoice paid
                </p>
                <p class="text-sm text-gray-500">
                  Your monthly invoice has been paid successfully
                </p>
              </div>
              <div class="text-sm text-gray-400">
                1 day ago
              </div>
            </div>
          </li>
          <li class="py-4">
            <div class="flex items-center space-x-4">
              <div class="flex-shrink-0">
                <div class="h-10 w-10 rounded-full bg-purple-100 flex items-center justify-center">
                  <svg class="h-5 w-5 text-purple-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                </div>
              </div>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-gray-900">
                  API usage update
                </p>
                <p class="text-sm text-gray-500">
                  You've used 45% of your monthly API quota
                </p>
              </div>
              <div class="text-sm text-gray-400">
                3 days ago
              </div>
            </div>
          </li>
        </ul>
      </div>
      <div class="mt-4 text-center">
        <a href="/audit" class="text-sm font-medium text-primary-600 hover:text-primary-500">
          View all activity →
        </a>
      </div>
    </div>
  {/if}
</div>
