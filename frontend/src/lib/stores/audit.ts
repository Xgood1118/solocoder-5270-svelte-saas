import { writable, derived } from 'svelte/store';
import type { AuditLog, AuditLogFilterParams, PaginatedResponse } from '$types';
import { apiClient } from '$api/client';

interface AuditPagination {
  page: number;
  perPage: number;
  total: number;
  totalPages: number;
}

function createAuditStore() {
  const logs = writable<AuditLog[]>([]);
  const pagination = writable<AuditPagination>({
    page: 1,
    perPage: 20,
    total: 0,
    totalPages: 0
  });
  const isLoading = writable(false);
  const error = writable<string | null>(null);

  const store = derived(
    [logs, pagination, isLoading, error],
    ([$logs, $pagination, $isLoading, $error]) => ({
      logs: $logs,
      pagination: $pagination,
      isLoading: $isLoading,
      error: $error
    })
  );

  async function fetchLogs(params?: AuditLogFilterParams & { page?: number; perPage?: number }): Promise<PaginatedResponse<AuditLog>> {
    isLoading.set(true);
    error.set(null);

    try {
      const searchParams = new URLSearchParams();

      if (params?.page) {
        searchParams.append('page', params.page.toString());
      }
      if (params?.perPage) {
        searchParams.append('per_page', params.perPage.toString());
      }
      if (params?.action) {
        searchParams.append('action', params.action);
      }
      if (params?.userId) {
        searchParams.append('user_id', params.userId);
      }
      if (params?.startDate) {
        searchParams.append('start_date', params.startDate.toISOString());
      }
      if (params?.endDate) {
        searchParams.append('end_date', params.endDate.toISOString());
      }

      const queryString = searchParams.toString();
      const endpoint = queryString ? `/audit-logs?${queryString}` : '/audit-logs';

      const response = await apiClient.get<PaginatedResponse<AuditLog>>(endpoint);

      logs.set(response.data);
      pagination.set(response.pagination);

      return response;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch audit logs';
      error.set(message);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  async function archiveLogs(beforeDate: Date): Promise<{ archived: number }> {
    isLoading.set(true);
    error.set(null);

    try {
      const response = await apiClient.post<{ archived: number }>('/audit-logs/archive', {
        beforeDate: beforeDate.toISOString()
      });

      await fetchLogs();

      return response;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to archive audit logs';
      error.set(message);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  function clearAudit() {
    logs.set([]);
    pagination.set({
      page: 1,
      perPage: 20,
      total: 0,
      totalPages: 0
    });
    error.set(null);
  }

  return {
    subscribe: store.subscribe,
    logs,
    pagination,
    isLoading,
    error,
    fetchLogs,
    archiveLogs,
    clearAudit
  };
}

export const auditStore = createAuditStore();
export const {
  logs: auditLogs,
  pagination: auditPagination,
  isLoading: auditLoading,
  error: auditError
} = auditStore;

export default auditStore;
