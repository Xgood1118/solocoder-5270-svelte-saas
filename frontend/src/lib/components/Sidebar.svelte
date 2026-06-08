<script lang="ts">
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { authStore, authUser } from '$stores/auth';
  import OrgSwitcher from './OrgSwitcher.svelte';
  import { cn } from '$helpers/cn';

  export let className = '';

  interface NavItem {
    href: string;
    label: string;
    icon: string;
  }

  const navItems: NavItem[] = [
    { href: '/dashboard', label: 'Dashboard', icon: 'dashboard' },
    { href: '/members', label: 'Members', icon: 'users' },
    { href: '/billing', label: 'Billing', icon: 'credit-card' },
    { href: '/audit', label: 'Audit Logs', icon: 'clock' },
    { href: '/settings', label: 'Settings', icon: 'settings' }
  ];

  const iconPaths: Record<string, string> = {
    dashboard: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6',
    users: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z',
    'credit-card': 'M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z',
    clock: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z',
    settings: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z'
  };

  async function handleLogout() {
    await authStore.logout();
    await goto('/login');
  }

  function isActive(href: string): boolean {
    return $page.url.pathname === href || $page.url.pathname.startsWith(href + '/');
  }
</script>

<aside class={cn('w-64 bg-white shadow-sm flex flex-col h-full', className)}>
  <div class="flex h-16 items-center justify-between px-6 border-b border-gray-200 flex-shrink-0">
    <h1 class="text-xl font-bold text-gray-900">SaaS</h1>
  </div>

  <div class="p-4 border-b border-gray-200 flex-shrink-0">
    <span class="text-xs font-medium text-gray-500 uppercase">Organization</span>
    <div class="mt-2">
      <OrgSwitcher />
    </div>
  </div>

  <nav class="flex-1 mt-6 px-3 overflow-y-auto">
    <ul class="space-y-1">
      {#each navItems as item (item.href)}
        <li>
          <a
            href={item.href}
            class={cn(
              'group flex items-center gap-3 px-3 py-2 text-sm font-medium rounded-md',
              isActive(item.href)
                ? 'bg-primary-50 text-primary-700'
                : 'text-gray-700 hover:bg-gray-50 hover:text-gray-900'
            )}
          >
            <svg
              class={cn(
                'h-5 w-5 flex-shrink-0',
                isActive(item.href) ? 'text-primary-600' : 'text-gray-400 group-hover:text-gray-500'
              )}
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path stroke-linecap="round" stroke-linejoin="round" d={iconPaths[item.icon]} />
            </svg>
            <span class="truncate">{item.label}</span>
          </a>
        </li>
      {/each}
    </ul>
  </nav>

  <div class="border-t border-gray-200 p-4 flex-shrink-0">
    <div class="flex items-center justify-between">
      <div class="flex items-center space-x-3 min-w-0">
        <div class="h-8 w-8 rounded-full bg-gray-300 flex items-center justify-center text-sm font-medium text-gray-700 flex-shrink-0">
          {$authUser?.name?.charAt(0) || 'U'}
        </div>
        <div class="text-sm min-w-0">
          <p class="font-medium text-gray-900 truncate">{$authUser?.name || 'User'}</p>
          <p class="text-gray-500 text-xs truncate">{$authUser?.email || ''}</p>
        </div>
      </div>
      <button
        type="button"
        on:click={handleLogout}
        class="text-gray-400 hover:text-gray-600 p-1 rounded-md hover:bg-gray-100 transition-colors"
        aria-label="Logout"
      >
        <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
        </svg>
      </button>
    </div>
  </div>
</aside>
