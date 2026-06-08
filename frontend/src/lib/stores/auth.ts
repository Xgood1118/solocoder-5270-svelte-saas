import { writable, derived, type Readable, type Writable } from 'svelte/store';
import type { User, AuthResponse, LoginRequest, RegisterRequest } from '$types';
import { apiClient } from '$api/client';

function createAuthStore() {
  const user = writable<User | null>(null);
  const token = writable<string | null>(null);
  const isLoading = writable(false);
  const error = writable<string | null>(null);

  const isAuthenticated = derived(token, ($token) => !!$token);

  const store = derived(
    [user, token, isLoading, error, isAuthenticated],
    ([$user, $token, $isLoading, $error, $isAuthenticated]) => ({
      user: $user,
      token: $token,
      isLoading: $isLoading,
      error: $error,
      isAuthenticated: $isAuthenticated
    })
  );

  function setAuthData(data: AuthResponse) {
    user.set(data.user);
    token.set(data.token);
    if (typeof window !== 'undefined') {
      localStorage.setItem('auth_token', data.token);
      localStorage.setItem('auth_user', JSON.stringify(data.user));
    }
  }

  function clearAuth() {
    user.set(null);
    token.set(null);
    if (typeof window !== 'undefined') {
      localStorage.removeItem('auth_token');
      localStorage.removeItem('auth_user');
      localStorage.removeItem('current_org_id');
    }
  }

  function initializeFromStorage() {
    if (typeof window !== 'undefined') {
      const storedToken = localStorage.getItem('auth_token');
      const storedUser = localStorage.getItem('auth_user');

      if (storedToken) {
        token.set(storedToken);
      }

      if (storedUser) {
        try {
          user.set(JSON.parse(storedUser));
        } catch {
          clearAuth();
        }
      }
    }
  }

  async function login(credentials: LoginRequest): Promise<AuthResponse> {
    isLoading.set(true);
    error.set(null);

    try {
      const response = await apiClient.post<AuthResponse>('/auth/login', credentials, {
        skipAuth: true,
        skipOrgHeader: true
      });
      setAuthData(response);
      return response;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Login failed';
      error.set(message);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  async function register(data: RegisterRequest): Promise<AuthResponse> {
    isLoading.set(true);
    error.set(null);

    try {
      const response = await apiClient.post<AuthResponse>('/auth/register', data, {
        skipAuth: true,
        skipOrgHeader: true
      });
      setAuthData(response);
      return response;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Registration failed';
      error.set(message);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  async function logout(): Promise<void> {
    try {
      await apiClient.post('/auth/logout');
    } catch {
    } finally {
      clearAuth();
    }
  }

  async function fetchCurrentUser(): Promise<User> {
    isLoading.set(true);
    error.set(null);

    try {
      const response = await apiClient.get<User>('/auth/me', {
        skipOrgHeader: true
      });
      user.set(response);
      if (typeof window !== 'undefined') {
        localStorage.setItem('auth_user', JSON.stringify(response));
      }
      return response;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch user';
      error.set(message);
      clearAuth();
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  return {
    subscribe: store.subscribe,
    user,
    token,
    isLoading,
    error,
    isAuthenticated,
    login,
    register,
    logout,
    fetchCurrentUser,
    initializeFromStorage,
    clearAuth
  };
}

export const authStore = createAuthStore();
export const { user: authUser, token: authToken, isLoading: authLoading, error: authError, isAuthenticated } = authStore;

export default authStore;
