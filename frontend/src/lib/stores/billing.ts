import { writable, derived } from 'svelte/store';
import type {
  Plan,
  Subscription,
  Invoice,
  Quota
} from '$types';
import { apiClient } from '$api/client';

function createBillingStore() {
  const plans = writable<Plan[]>([]);
  const subscription = writable<Subscription | null>(null);
  const invoices = writable<Invoice[]>([]);
  const quotas = writable<Quota[]>([]);
  const isLoading = writable(false);
  const error = writable<string | null>(null);

  const isSubscribed = derived(subscription, ($subscription) => {
    return $subscription?.status === 'active' || $subscription?.status === 'trialing';
  });

  const currentPlan = derived([plans, subscription], ([$plans, $subscription]) => {
    if (!$subscription) return null;
    return $plans.find((plan) => plan.id === $subscription.planId) || null;
  });

  const store = derived(
    [plans, subscription, invoices, quotas, isLoading, error, isSubscribed, currentPlan],
    ([$plans, $subscription, $invoices, $quotas, $isLoading, $error, $isSubscribed, $currentPlan]) => ({
      plans: $plans,
      subscription: $subscription,
      invoices: $invoices,
      quotas: $quotas,
      isLoading: $isLoading,
      error: $error,
      isSubscribed: $isSubscribed,
      currentPlan: $currentPlan
    })
  );

  async function fetchPlans(): Promise<Plan[]> {
    isLoading.set(true);
    error.set(null);

    try {
      const response = await apiClient.get<Plan[]>('/plans', { skipOrgHeader: true });
      plans.set(response.filter((p) => p.isActive));
      return response;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch plans';
      error.set(message);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  async function fetchSubscription(): Promise<Subscription> {
    isLoading.set(true);
    error.set(null);

    try {
      const response = await apiClient.get<Subscription>('/billing/subscription');
      subscription.set(response);
      return response;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch subscription';
      error.set(message);
      subscription.set(null);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  async function upgradePlan(planId: string): Promise<Subscription> {
    isLoading.set(true);
    error.set(null);

    try {
      const response = await apiClient.post<Subscription>('/billing/subscription/upgrade', {
        planId
      });
      subscription.set(response);
      return response;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to upgrade plan';
      error.set(message);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  async function downgradePlan(planId: string): Promise<Subscription> {
    isLoading.set(true);
    error.set(null);

    try {
      const response = await apiClient.post<Subscription>('/billing/subscription/downgrade', {
        planId
      });
      subscription.set(response);
      return response;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to downgrade plan';
      error.set(message);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  async function cancelSubscription(): Promise<Subscription> {
    isLoading.set(true);
    error.set(null);

    try {
      const response = await apiClient.post<Subscription>('/billing/subscription/cancel');
      subscription.set(response);
      return response;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to cancel subscription';
      error.set(message);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  async function fetchInvoices(): Promise<Invoice[]> {
    isLoading.set(true);
    error.set(null);

    try {
      const response = await apiClient.get<Invoice[]>('/billing/invoices');
      invoices.set(response);
      return response;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch invoices';
      error.set(message);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  async function payInvoice(invoiceId: string): Promise<Invoice> {
    isLoading.set(true);
    error.set(null);

    try {
      const response = await apiClient.post<Invoice>(`/billing/invoices/${invoiceId}/pay`);
      invoices.update(($invoices) =>
        $invoices.map((inv) => (inv.id === invoiceId ? response : inv))
      );
      return response;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to pay invoice';
      error.set(message);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  async function fetchQuotas(): Promise<Quota[]> {
    isLoading.set(true);
    error.set(null);

    try {
      const response = await apiClient.get<Quota[]>('/billing/quotas');
      quotas.set(response);
      return response;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch quotas';
      error.set(message);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  function clearBilling() {
    plans.set([]);
    subscription.set(null);
    invoices.set([]);
    quotas.set([]);
    error.set(null);
  }

  return {
    subscribe: store.subscribe,
    plans,
    subscription,
    invoices,
    quotas,
    isLoading,
    error,
    isSubscribed,
    currentPlan,
    fetchPlans,
    fetchSubscription,
    upgradePlan,
    downgradePlan,
    cancelSubscription,
    fetchInvoices,
    payInvoice,
    fetchQuotas,
    clearBilling
  };
}

export const billingStore = createBillingStore();
export const {
  plans,
  subscription,
  invoices,
  quotas,
  isLoading: billingLoading,
  error: billingError,
  isSubscribed,
  currentPlan
} = billingStore;

export default billingStore;
