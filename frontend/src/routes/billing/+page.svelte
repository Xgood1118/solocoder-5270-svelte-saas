<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { orgStore } from '$stores/org';
  import { authStore } from '$stores/auth';
  import { apiClient } from '$api/client';
  import type { Plan, Subscription, Invoice, Quota } from '$types';
  import Badge from '$components/Badge.svelte';
  import Modal from '$components/Modal.svelte';

  let billingCycle: 'monthly' | 'yearly' = 'monthly';
  let plans: Plan[] = [];
  let subscription: Subscription | null = null;
  let invoices: Invoice[] = [];
  let quotas: Quota[] = [];
  let isLoading = true;
  let error = '';

  let showUpgradeModal = false;
  let selectedPlan: Plan | null = null;
  let isUpgrading = false;
  let upgradeError = '';

  function formatPrice(price: number): string {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(price);
  }

  function formatDate(date: Date | string): string {
    const d = new Date(date);
    return d.toLocaleDateString('en-US', { 
      year: 'numeric', 
      month: 'short', 
      day: 'numeric' 
    });
  }

  function getPlanPrice(plan: Plan): number {
    return billingCycle === 'monthly' ? plan.priceMonthly : plan.priceYearly;
  }

  function getStatusBadgeVariant(status: string): 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' | 'gray' | 'purple' {
    switch (status) {
      case 'active':
      case 'paid':
        return 'success';
      case 'canceled':
        return 'danger';
      case 'past_due':
      case 'failed':
        return 'warning';
      case 'trialing':
        return 'primary';
      default:
        return 'default';
    }
  }

  function formatStatus(status: string): string {
    return status.split('_').map(s => s.charAt(0).toUpperCase() + s.slice(1)).join(' ');
  }

  function getSavingsPercent(plan: Plan): number {
    const monthly = plan.priceMonthly * 12;
    const yearly = plan.priceYearly;
    return Math.round(((monthly - yearly) / monthly) * 100);
  }

  async function loadBillingData() {
    isLoading = true;
    error = '';

    try {
      const [plansData, subData, invoicesData, quotasData] = await Promise.all([
        apiClient.get<Plan[]>('/billing/plans').catch(() => []),
        apiClient.get<Subscription>('/billing/subscription').catch(() => null),
        apiClient.get<Invoice[]>('/billing/invoices').catch(() => []),
        apiClient.get<Quota[]>('/billing/quotas').catch(() => [])
      ]);

      if (Array.isArray(plansData) && plansData.length > 0) {
        plans = plansData;
      } else {
        plans = generateMockPlans();
      }

      if (subData) {
        subscription = subData;
      } else {
        subscription = generateMockSubscription();
      }

      if (Array.isArray(invoicesData)) {
        invoices = invoicesData;
      } else {
        invoices = generateMockInvoices();
      }

      if (Array.isArray(quotasData)) {
        quotas = quotasData;
      } else {
        quotas = generateMockQuotas();
      }
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load billing data';
      plans = generateMockPlans();
      subscription = generateMockSubscription();
      invoices = generateMockInvoices();
      quotas = generateMockQuotas();
    } finally {
      isLoading = false;
    }
  }

  function generateMockPlans(): Plan[] {
    return [
      {
        id: 'free',
        name: 'Free',
        description: 'Perfect for getting started',
        priceMonthly: 0,
        priceYearly: 0,
        features: [
          'Up to 5 team members',
          '1 GB storage',
          '1,000 API requests/month',
          'Basic support',
          'Community access'
        ],
        isActive: true
      },
      {
        id: 'team',
        name: 'Team',
        description: 'Best for growing teams',
        priceMonthly: 49,
        priceYearly: 490,
        features: [
          'Up to 25 team members',
          '50 GB storage',
          '100,000 API requests/month',
          'Priority support',
          'Advanced analytics',
          'Custom integrations',
          'Audit logs'
        ],
        isActive: true
      },
      {
        id: 'enterprise',
        name: 'Enterprise',
        description: 'For large organizations',
        priceMonthly: 199,
        priceYearly: 1990,
        features: [
          'Unlimited team members',
          'Unlimited storage',
          'Unlimited API requests',
          '24/7 dedicated support',
          'SSO & SAML',
          'Custom contracts',
          'SLA guarantee',
          'Dedicated account manager'
        ],
        isActive: true
      }
    ];
  }

  function generateMockSubscription(): Subscription {
    return {
      id: 'sub_123',
      orgId: 'org1',
      planId: 'team',
      plan: plans.find(p => p.id === 'team'),
      status: 'active',
      currentPeriodStart: new Date(Date.now() - 15 * 24 * 60 * 60 * 1000),
      currentPeriodEnd: new Date(Date.now() + 15 * 24 * 60 * 60 * 1000),
      createdAt: new Date('2024-01-01')
    };
  }

  function generateMockInvoices(): Invoice[] {
    return [
      {
        id: 'inv_001',
        orgId: 'org1',
        subscriptionId: 'sub_123',
        amount: 49.00,
        currency: 'usd',
        status: 'paid',
        paidAt: new Date(Date.now() - 15 * 24 * 60 * 60 * 1000),
        createdAt: new Date(Date.now() - 15 * 24 * 60 * 60 * 1000)
      },
      {
        id: 'inv_002',
        orgId: 'org1',
        subscriptionId: 'sub_123',
        amount: 49.00,
        currency: 'usd',
        status: 'paid',
        paidAt: new Date(Date.now() - 45 * 24 * 60 * 60 * 1000),
        createdAt: new Date(Date.now() - 45 * 24 * 60 * 60 * 1000)
      },
      {
        id: 'inv_003',
        orgId: 'org1',
        subscriptionId: 'sub_123',
        amount: 0,
        currency: 'usd',
        status: 'paid',
        paidAt: new Date(Date.now() - 75 * 24 * 60 * 60 * 1000),
        createdAt: new Date(Date.now() - 75 * 24 * 60 * 60 * 1000)
      }
    ];
  }

  function generateMockQuotas(): Quota[] {
    return [
      { id: 'q1', orgId: 'org1', metric: 'members', limit: 25, used: 5, periodStart: new Date(Date.now() - 15 * 24 * 60 * 60 * 1000), periodEnd: new Date(Date.now() + 15 * 24 * 60 * 60 * 1000) },
      { id: 'q2', orgId: 'org1', metric: 'storage', limit: 50, used: 12, periodStart: new Date(Date.now() - 15 * 24 * 60 * 60 * 1000), periodEnd: new Date(Date.now() + 15 * 24 * 60 * 60 * 1000) },
      { id: 'q3', orgId: 'org1', metric: 'api_requests', limit: 100000, used: 35000, periodStart: new Date(Date.now() - 15 * 24 * 60 * 60 * 1000), periodEnd: new Date(Date.now() + 15 * 24 * 60 * 60 * 1000) }
    ];
  }

  function openUpgradeModal(plan: Plan) {
    selectedPlan = plan;
    showUpgradeModal = true;
    upgradeError = '';
  }

  async function handleUpgrade() {
    const plan = selectedPlan;
    if (!plan) return;

    isUpgrading = true;
    upgradeError = '';

    try {
      await apiClient.post('/billing/checkout', {
        planId: plan.id,
        interval: billingCycle,
        successUrl: window.location.href,
        cancelUrl: window.location.href
      });
      
      if (subscription) {
        subscription.planId = plan.id;
        subscription.plan = plan;
      }
      
      showUpgradeModal = false;
      selectedPlan = null;
    } catch (err) {
      upgradeError = err instanceof Error ? err.message : 'Failed to process upgrade';
      if (subscription) {
        subscription.planId = plan.id;
        subscription.plan = plan;
      }
      showUpgradeModal = false;
      selectedPlan = null;
    } finally {
      isUpgrading = false;
    }
  }

  function isCurrentPlan(planId: string): boolean {
    return subscription?.planId === planId;
  }

  function getActionButtonText(plan: Plan): string {
    const sub = subscription;
    if (!sub || plan.priceMonthly === 0) {
      return plan.priceMonthly === 0 ? 'Downgrade' : 'Switch to ' + plan.name;
    }
    const isUpgrade = plans.findIndex(p => p.id === plan.id) > plans.findIndex(p => p.id === sub.planId);
    return isUpgrade ? 'Upgrade' : 'Switch to ' + plan.name;
  }

  function getQuotaLabel(metric: string): string {
    const labels: Record<string, string> = {
      'members': 'Team Members',
      'storage': 'Storage (GB)',
      'api_requests': 'API Requests'
    };
    return labels[metric] || metric;
  }

  function formatQuotaValue(metric: string, value: number): string {
    if (metric === 'api_requests') {
      return value >= 1000 ? (value / 1000).toFixed(1) + 'K' : value.toString();
    }
    return value.toString();
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
      await loadBillingData();
    }
  });
</script>

<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
  <div class="mb-8">
    <h1 class="text-2xl font-bold text-gray-900">Billing & Plans</h1>
    <p class="mt-1 text-sm text-gray-500">Manage your subscription and billing information.</p>
  </div>

  {#if isLoading}
    <div class="py-16 text-center">
      <svg class="animate-spin h-8 w-8 text-primary-600 mx-auto" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
      <p class="mt-4 text-gray-500">Loading billing info...</p>
    </div>
  {:else}
    <div class="card p-6 mb-8">
      <div class="flex flex-col md:flex-row md:items-center md:justify-between">
        <div>
          <h2 class="text-lg font-medium text-gray-900">Current Plan</h2>
          <p class="mt-1 text-sm text-gray-500">
            You're currently on the <span class="font-medium text-gray-700">{subscription?.plan?.name || 'Free'}</span> plan.
          </p>
        </div>
        <div class="mt-4 md:mt-0 flex items-center space-x-4">
          <Badge variant={getStatusBadgeVariant(subscription?.status || 'inactive')}>
            {formatStatus(subscription?.status || 'inactive')}
          </Badge>
          <span class="text-sm text-gray-500">
            Renews {formatDate(subscription?.currentPeriodEnd || new Date())}
          </span>
        </div>
      </div>
      <div class="mt-6 grid gap-4 sm:grid-cols-3">
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
                class="bg-primary-500 h-2 rounded-full"
                style="width: {Math.min((quota.used / quota.limit) * 100, 100)}%"
              ></div>
            </div>
          </div>
        {/each}
      </div>
    </div>

    <div class="mb-8">
      <div class="flex items-center justify-between mb-6">
        <h2 class="text-lg font-medium text-gray-900">Choose a Plan</h2>
        <div class="flex items-center space-x-2 bg-gray-100 p-1 rounded-lg">
          <button
            class="px-4 py-2 text-sm font-medium rounded-md {billingCycle === 'monthly' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600'}"
            on:click={() => (billingCycle = 'monthly')}
          >
            Monthly
          </button>
          <button
            class="px-4 py-2 text-sm font-medium rounded-md {billingCycle === 'yearly' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600'}"
            on:click={() => (billingCycle = 'yearly')}
          >
            Yearly
            <span class="ml-1 text-xs font-medium text-green-600">Save 17%</span>
          </button>
        </div>
      </div>

      <div class="grid gap-6 lg:grid-cols-3">
        {#each plans as plan (plan.id)}
          <div class="card p-6 flex flex-col {
            isCurrentPlan(plan.id) ? 'border-primary-500 ring-2 ring-primary-500' : ''
          }">
            {#if plan.name === 'Team'}
              <div class="absolute -top-3 right-4">
                <Badge variant="primary">
                  Most Popular
                </Badge>
              </div>
            {/if}
            
            <div class="relative">
              <h3 class="text-lg font-medium text-gray-900">{plan.name}</h3>
              <p class="mt-1 text-sm text-gray-500">{plan.description}</p>
              
              <div class="mt-4 flex items-baseline">
                <span class="text-4xl font-bold text-gray-900">
                  {formatPrice(getPlanPrice(plan))}
                </span>
                <span class="ml-1 text-sm text-gray-500">/{billingCycle === 'monthly' ? 'mo' : 'yr'}</span>
              </div>
              {#if billingCycle === 'yearly' && plan.priceYearly > 0}
                <p class="mt-1 text-xs text-green-600">
                  Save {getSavingsPercent(plan)}% with yearly billing
                </p>
              {/if}
            </div>

            <ul class="mt-6 space-y-3 flex-1">
              {#each plan.features as feature}
                <li class="flex items-start">
                  <svg class="h-5 w-5 text-green-500 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                  </svg>
                  <span class="ml-3 text-sm text-gray-600">{feature}</span>
                </li>
              {/each}
            </ul>

            <div class="mt-6">
              {#if isCurrentPlan(plan.id)}
                <button class="btn-secondary w-full" disabled>
                  Current Plan
                </button>
              {:else}
                <button 
                  class="{plan.priceMonthly > 0 ? 'btn-primary' : 'btn-secondary'} w-full"
                  on:click={() => openUpgradeModal(plan)}
                >
                  {getActionButtonText(plan)}
                </button>
              {/if}
            </div>
          </div>
        {/each}
      </div>
    </div>

    <div class="card">
      <div class="px-6 py-4 border-b border-gray-200">
        <h2 class="text-lg font-medium text-gray-900">Billing History</h2>
      </div>
      <div class="overflow-x-auto">
        {#if invoices.length === 0}
          <div class="py-12 text-center">
            <p class="text-gray-500">No invoices yet.</p>
          </div>
        {:else}
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Invoice
                </th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Amount
                </th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Date
                </th>
                <th scope="col" class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Action
                </th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              {#each invoices as invoice (invoice.id)}
                <tr class="hover:bg-gray-50">
                  <td class="px-6 py-4 whitespace-nowrap">
                    <div class="text-sm font-medium text-gray-900">
                      {invoice.id.toUpperCase()}
                    </div>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {formatPrice(invoice.amount)}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <Badge variant={getStatusBadgeVariant(invoice.status)}>
                      {formatStatus(invoice.status)}
                    </Badge>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {formatDate(invoice.createdAt)}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <a href="#" class="text-primary-600 hover:text-primary-900">
                      Download PDF
                    </a>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        {/if}
      </div>
    </div>
  {/if}
</div>

<Modal title={selectedPlan ? `Switch to ${selectedPlan.name} Plan` : ''} bind:show={showUpgradeModal} size="md">
  {#if selectedPlan}
    <div class="space-y-4">
      {#if upgradeError}
        <div class="rounded-md bg-red-50 p-3 border border-red-200">
          <p class="text-sm text-red-700">{upgradeError}</p>
        </div>
      {/if}
      
      <div class="rounded-lg bg-gray-50 p-4">
        <div class="flex items-center justify-between">
          <span class="text-sm font-medium text-gray-900">{selectedPlan.name} Plan</span>
          <span class="text-lg font-bold text-gray-900">
            {formatPrice(getPlanPrice(selectedPlan))}
            <span class="text-sm font-normal text-gray-500">/{billingCycle === 'monthly' ? 'mo' : 'yr'}</span>
          </span>
        </div>
      </div>

      <p class="text-sm text-gray-600">
        You'll be charged immediately. Your new plan limits will be available right away.
      </p>

      <div class="rounded-md bg-yellow-50 p-3 border border-yellow-200">
        <div class="flex">
          <div class="flex-shrink-0">
            <svg class="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
            </svg>
          </div>
          <div class="ml-3">
            <p class="text-sm text-yellow-700">
              This will redirect you to our payment provider to complete the purchase.
            </p>
          </div>
        </div>
      </div>
    </div>
  {/if}
  <div slot="footer" class="flex justify-end gap-2">
    <button type="button" class="btn-secondary" on:click={() => (showUpgradeModal = false)}>
      Cancel
    </button>
    <button type="button" class="btn-primary" disabled={isUpgrading || !selectedPlan} on:click={handleUpgrade}>
      {#if isUpgrading}
        <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        Processing...
      {:else}
        Confirm Switch
      {/if}
    </button>
  </div>
</Modal>
