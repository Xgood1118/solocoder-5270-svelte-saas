<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import SidebarLayout from '$components/SidebarLayout.svelte';
  import { authStore } from '$stores/auth';
  import { orgStore } from '$stores/org';

  onMount(async () => {
    if (!$authStore.isAuthenticated) {
      await goto('/login');
      return;
    }
    if ($orgStore.orgs.length === 0) {
      await orgStore.fetchOrgs();
    }
    if (!$orgStore.currentOrgId && $orgStore.orgs.length > 0) {
      orgStore.setCurrentOrgId($orgStore.orgs[0].id);
    }
  });
</script>

{#if $authStore.isAuthenticated && $orgStore.currentOrgId}
  <SidebarLayout>
    <slot />
  </SidebarLayout>
{/if}
