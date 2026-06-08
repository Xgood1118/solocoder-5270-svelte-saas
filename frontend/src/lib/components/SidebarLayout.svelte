<script lang="ts">
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { authStore } from '$stores/auth';
  import { orgStore } from '$stores/org';

  const navItems = [
    { href: '/dashboard', label: 'Dashboard', icon: 'dashboard' },
    { href: '/members', label: 'Members', icon: 'users' },
    { href: '/billing', label: 'Billing', icon: 'credit-card' },
    { href: '/audit', label: 'Audit Log', icon: 'clock' },
    { href: '/settings', label: 'Settings', icon: 'settings' }
  ];

  async function handleLogout() {
    await authStore.logout();
    await goto('/login');
  }

  function switchOrg(event: Event) {
    const target = event.target as HTMLSelectElement;
    orgStore.setCurrentOrgId(target.value);
  }
</script>

<div class="flex h-screen bg-gray-100">
  <aside class="w-64 bg-white shadow-sm">
    <div class="flex h-16 items-center justify-between px-6 border-b border-gray-200">
      <h1 class="text-xl font-bold text-gray-900">SaaS</h1>
    </div>

    <div class="p-4 border-b border-gray-200">
      <label class="text-xs font-medium text-gray-500 uppercase">Current Org</label>
      <div class="mt-2">
        <select
          class="w-full rounded-md border-gray-300 text-sm focus:border-primary-500 focus:ring-primary-500"
          on:change={switchOrg}
          value={$orgStore.currentOrgId || ''}
        >
          {#each $orgStore.orgs as org (org.id)}
            <option value={org.id}>{org.name}</option>
          {/each}
        </select>
      </div>
    </div>

    <nav class="mt-6 px-3">
      {#each navItems as item (item.href)}
        <a
          href={item.href}
          class="group flex items-center px-3 py-2 text-sm font-medium rounded-md {
            $page.url.pathname === item.href
              ? 'bg-primary-50 text-primary-700'
              : 'text-gray-700 hover:bg-gray-50 hover:text-gray-900'
          }"
        >
          <span class="truncate">{item.label}</span>
        </a>
      {/each}
    </nav>

    <div class="absolute bottom-0 w-64 border-t border-gray-200 p-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center space-x-3">
          <div class="h-8 w-8 rounded-full bg-gray-300 flex items-center justify-center text-sm font-medium text-gray-700">
            {$authStore.user?.name?.charAt(0) || 'U'}
          </div>
          <div class="text-sm">
            <p class="font-medium text-gray-900">{$authStore.user?.name || 'User'}</p>
            <p class="text-gray-500 text-xs">{$authStore.user?.email || ''}</p>
          </div>
        </div>
        <button on:click={handleLogout} class="text-gray-400 hover:text-gray-600">
          <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
          </svg>
        </button>
      </div>
    </div>
  </aside>

  <main class="flex-1 overflow-y-auto">
    <div class="py-8">
      <slot />
    </div>
  </main>
</div>
