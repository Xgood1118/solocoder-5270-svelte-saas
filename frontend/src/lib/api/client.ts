import { get } from 'svelte/store';
import { authToken } from '$stores/auth';
import { currentOrgId } from '$stores/org';
import type { ApiResponse } from '$types';

const API_BASE_URL = '/api';

interface RequestOptions extends Omit<RequestInit, 'body'> {
  body?: unknown;
  skipAuth?: boolean;
  skipOrgHeader?: boolean;
}

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  private getHeaders(skipAuth = false, skipOrgHeader = false): HeadersInit {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json'
    };

    if (!skipAuth) {
      const token = get(authToken);
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }
    }

    if (!skipOrgHeader) {
      const orgId = get(currentOrgId);
      if (orgId) {
        headers['X-Org-Id'] = orgId;
      }
    }

    return headers;
  }

  private async request<T>(
    endpoint: string,
    options: RequestOptions = {}
  ): Promise<T> {
    const { body, skipAuth, skipOrgHeader, ...init } = options;

    const url = `${this.baseUrl}${endpoint}`;
    const headers = this.getHeaders(skipAuth, skipOrgHeader);

    const config: RequestInit = {
      ...init,
      headers
    };

    if (body !== undefined) {
      config.body = JSON.stringify(body);
    }

    try {
      const response = await fetch(url, config);
      const data = (await response.json()) as ApiResponse<T>;

      if (!response.ok) {
        throw new Error(data.error || data.message || `HTTP error! status: ${response.status}`);
      }

      return data.data as T;
    } catch (error) {
      if (error instanceof Error) {
        throw error;
      }
      throw new Error('An unexpected error occurred');
    }
  }

  get<T>(endpoint: string, options: Omit<RequestOptions, 'body'> = {}): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'GET' });
  }

  post<T>(endpoint: string, body?: unknown, options: RequestOptions = {}): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'POST', body });
  }

  put<T>(endpoint: string, body?: unknown, options: RequestOptions = {}): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'PUT', body });
  }

  patch<T>(endpoint: string, body?: unknown, options: RequestOptions = {}): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'PATCH', body });
  }

  delete<T>(endpoint: string, options: Omit<RequestOptions, 'body'> = {}): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'DELETE' });
  }
}

export const apiClient = new ApiClient(API_BASE_URL);

export default apiClient;
