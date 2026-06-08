import { writable, derived } from 'svelte/store';
import type { Org, CreateOrgRequest, UpdateOrgRequest } from '$types';
import { apiClient } from '$api/client';

function createOrgStore() {
  const orgs = writable<Org[]>([]);
  const currentOrgId = writable<string | null>(null);
  const isLoading = writable(false);
  const error = writable<string | null>(null);

  const currentOrg = derived([orgs, currentOrgId], ([$orgs, $currentOrgId]) => {
    if (!$currentOrgId) return null;
    return $orgs.find((org) => org.id === $currentOrgId) || null;
  });

  const store = derived(
    [orgs, currentOrgId, currentOrg, isLoading, error],
    ([$orgs, $currentOrgId, $currentOrg, $isLoading, $error]) => ({
      orgs: $orgs,
      currentOrgId: $currentOrgId,
      currentOrg: $currentOrg,
      isLoading: $isLoading,
      error: $error
    })
  );

  function setCurrentOrgId(orgId: string | null) {
    currentOrgId.set(orgId);
    if (typeof window !== 'undefined' && orgId) {
      localStorage.setItem('current_org_id', orgId);
    } else if (typeof window !== 'undefined') {
      localStorage.removeItem('current_org_id');
    }
  }

  function initializeFromStorage() {
    if (typeof window !== 'undefined') {
      const storedOrgId = localStorage.getItem('current_org_id');
      if (storedOrgId) {
        currentOrgId.set(storedOrgId);
      }
    }
  }

  async function fetchOrgs(): Promise<Org[]> {
    isLoading.set(true);
    error.set(null);

    try {
      const response = await apiClient.get<Org[]>('/orgs', {
        skipOrgHeader: true
      });
      orgs.set(response);

      if (response.length > 0 && !getCurrentOrgIdSync()) {
        setCurrentOrgId(response[0].id);
      }

      return response;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch organizations';
      error.set(message);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  function getCurrentOrgIdSync(): string | null {
    let id: string | null = null;
    currentOrgId.subscribe(($id) => {
      id = $id;
    })();
    return id;
  }

  async function createOrg(data: CreateOrgRequest): Promise<Org> {
    isLoading.set(true);
    error.set(null);

    try {
      const response = await apiClient.post<Org>('/orgs', data, {
        skipOrgHeader: true
      });
      orgs.update(($orgs) => [...$orgs, response]);
      setCurrentOrgId(response.id);
      return response;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to create organization';
      error.set(message);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  async function updateOrg(orgId: string, data: UpdateOrgRequest): Promise<Org> {
    isLoading.set(true);
    error.set(null);

    try {
      const response = await apiClient.put<Org>(`/orgs/${orgId}`, data);
      orgs.update(($orgs) =>
        $orgs.map((org) => (org.id === orgId ? { ...org, ...response } : org))
      );
      return response;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to update organization';
      error.set(message);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  async function deleteOrg(orgId: string): Promise<void> {
    isLoading.set(true);
    error.set(null);

    try {
      await apiClient.delete(`/orgs/${orgId}`);
      orgs.update(($orgs) => $orgs.filter((org) => org.id !== orgId));

      const currentId = getCurrentOrgIdSync();
      if (currentId === orgId) {
        orgs.subscribe(($orgs) => {
          if ($orgs.length > 0) {
            setCurrentOrgId($orgs[0].id);
          } else {
            setCurrentOrgId(null);
          }
        })();
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to delete organization';
      error.set(message);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  function clearOrgs() {
    orgs.set([]);
    currentOrgId.set(null);
    if (typeof window !== 'undefined') {
      localStorage.removeItem('current_org_id');
    }
  }

  return {
    subscribe: store.subscribe,
    orgs,
    currentOrgId,
    currentOrg,
    isLoading,
    error,
    fetchOrgs,
    createOrg,
    updateOrg,
    deleteOrg,
    setCurrentOrgId,
    initializeFromStorage,
    clearOrgs
  };
}

export const orgStore = createOrgStore();
export const {
  orgs,
  currentOrgId,
  currentOrg,
  isLoading: orgLoading,
  error: orgError
} = orgStore;

export default orgStore;
