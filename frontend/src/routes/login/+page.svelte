<script lang="ts">
  import { goto } from '$app/navigation';
  import { authStore } from '$stores/auth';
  import { orgStore } from '$stores/org';

  let email = '';
  let password = '';
  let error = '';
  let isLoading = false;
  let emailError = '';
  let passwordError = '';

  function validateEmail(): boolean {
    if (!email) {
      emailError = 'Email is required';
      return false;
    }
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      emailError = 'Please enter a valid email address';
      return false;
    }
    emailError = '';
    return true;
  }

  function validatePassword(): boolean {
    if (!password) {
      passwordError = 'Password is required';
      return false;
    }
    if (password.length < 6) {
      passwordError = 'Password must be at least 6 characters';
      return false;
    }
    passwordError = '';
    return true;
  }

  async function handleLogin() {
    error = '';
    
    const isEmailValid = validateEmail();
    const isPasswordValid = validatePassword();
    
    if (!isEmailValid || !isPasswordValid) {
      return;
    }

    isLoading = true;

    try {
      const response = await authStore.login({ email, password });
      if (response.orgs && response.orgs.length > 0) {
        orgStore.setCurrentOrgId(response.orgs[0].id);
        await goto('/dashboard');
      } else {
        await goto('/orgs');
      }
    } catch (err) {
      error = err instanceof Error ? err.message : 'Login failed';
    } finally {
      isLoading = false;
    }
  }
</script>

<div class="flex min-h-screen items-center justify-center bg-gradient-to-br from-primary-50 via-white to-blue-50 py-12 px-4 sm:px-6 lg:px-8">
  <div class="w-full max-w-md">
    <div class="text-center">
      <div class="mx-auto h-12 w-12 rounded-lg bg-primary-600 flex items-center justify-center">
        <svg class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
      </div>
      <h2 class="mt-6 text-3xl font-bold tracking-tight text-gray-900">
        Welcome back
      </h2>
      <p class="mt-2 text-sm text-gray-600">
        Sign in to your account to continue
      </p>
    </div>

    <form class="mt-8 space-y-6" on:submit|preventDefault={handleLogin}>
      {#if error}
        <div class="rounded-lg bg-red-50 p-4 border border-red-200">
          <div class="flex">
            <div class="flex-shrink-0">
              <svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
              </svg>
            </div>
            <div class="ml-3">
              <p class="text-sm font-medium text-red-800">{error}</p>
            </div>
          </div>
        </div>
      {/if}

      <div class="space-y-5">
        <div>
          <label for="email" class="block text-sm font-medium text-gray-700">Email address</label>
          <div class="mt-1">
            <input
              id="email"
              name="email"
              type="email"
              autocomplete="email"
              required
              class="input-field {emailError ? 'border-red-300 focus:ring-red-500 focus:border-red-500' : ''} mt-1"
              bind:value={email}
              on:blur={validateEmail}
              placeholder="you@example.com"
            />
            {#if emailError}
              <p class="mt-1 text-sm text-red-600">{emailError}</p>
            {/if}
          </div>
        </div>

        <div>
          <label for="password" class="block text-sm font-medium text-gray-700">Password</label>
          <div class="mt-1">
            <input
              id="password"
              name="password"
              type="password"
              autocomplete="current-password"
              required
              class="input-field {passwordError ? 'border-red-300 focus:ring-red-500 focus:border-red-500' : ''} mt-1"
              bind:value={password}
              on:blur={validatePassword}
              placeholder="••••••••"
            />
            {#if passwordError}
              <p class="mt-1 text-sm text-red-600">{passwordError}</p>
            {/if}
          </div>
        </div>
      </div>

      <div class="flex items-center justify-between">
        <div class="flex items-center">
          <input id="remember-me" name="remember-me" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
          <label for="remember-me" class="ml-2 block text-sm text-gray-700">Remember me</label>
        </div>

        <div class="text-sm">
          <a href="#" class="font-medium text-primary-600 hover:text-primary-500">
            Forgot your password?
          </a>
        </div>
      </div>

      <div>
        <button type="submit" class="btn-primary w-full" disabled={isLoading}>
          {#if isLoading}
            <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Signing in...
          {:else}
            Sign in
          {/if}
        </button>
      </div>

      <div class="text-center text-sm">
        <span class="text-gray-600">Don't have an account?</span>
        <a href="/register" class="font-medium text-primary-600 hover:text-primary-500 ml-1">
          Sign up for free
        </a>
      </div>
    </form>
  </div>
</div>
