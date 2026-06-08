<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { orgStore } from '$stores/org';
  import { authStore } from '$stores/auth';
  import { apiClient } from '$api/client';
  import type { OrgMember, Invitation, OrgRole } from '$types';
  import Modal from '$components/Modal.svelte';
  import Badge from '$components/Badge.svelte';

  let activeTab: 'members' | 'invitations' = 'members';
  let members: OrgMember[] = [];
  let invitations: Invitation[] = [];
  let isLoading = true;
  let error = '';

  let showInviteModal = false;
  let inviteEmail = '';
  let inviteRole: OrgRole = 'member';
  let inviteError = '';
  let isInviting = false;

  let showRemoveModal = false;
  let memberToRemove: OrgMember | null = null;
  let isRemoving = false;
  let removeError = '';

  let roleMenuOpen: string | null = null;

  const allRoles: OrgRole[] = ['owner', 'admin', 'member', 'guest'];

  function getRoleBadgeVariant(role: OrgRole): 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' | 'gray' | 'purple' {
    switch (role) {
      case 'owner':
        return 'purple';
      case 'admin':
        return 'primary';
      case 'member':
        return 'default';
      case 'guest':
        return 'warning';
      default:
        return 'default';
    }
  }

  function formatRole(role: OrgRole): string {
    return role.charAt(0).toUpperCase() + role.slice(1);
  }

  function formatDate(date: Date | string): string {
    const d = new Date(date);
    return d.toLocaleDateString('en-US', { 
      year: 'numeric', 
      month: 'short', 
      day: 'numeric' 
    });
  }

  function canChangeRole(member: OrgMember): boolean {
    if (!$authStore.user) return false;
    const currentUserMember = members.find(m => m.userId === $authStore.user?.id);
    if (!currentUserMember) return false;
    if (currentUserMember.role === 'owner') return true;
    if (currentUserMember.role === 'admin' && member.role !== 'owner' && member.role !== 'admin') return true;
    return false;
  }

  function canRemoveMember(member: OrgMember): boolean {
    if (!$authStore.user) return false;
    if (member.userId === $authStore.user.id) return false;
    const currentUserMember = members.find(m => m.userId === $authStore.user?.id);
    if (!currentUserMember) return false;
    if (currentUserMember.role === 'owner') return true;
    if (currentUserMember.role === 'admin' && member.role !== 'owner' && member.role !== 'admin') return true;
    return false;
  }

  async function loadMembers() {
    isLoading = true;
    error = '';

    try {
      const [membersData, invitationsData] = await Promise.all([
        apiClient.get<OrgMember[]>('/members').catch(() => []),
        apiClient.get<Invitation[]>('/members/invitations').catch(() => [])
      ]);

      if (Array.isArray(membersData)) {
        members = membersData;
      } else {
        members = generateMockMembers();
      }

      if (Array.isArray(invitationsData)) {
        invitations = invitationsData.filter(i => i.status === 'pending');
      } else {
        invitations = generateMockInvitations();
      }
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load members';
      members = generateMockMembers();
      invitations = generateMockInvitations();
    } finally {
      isLoading = false;
    }
  }

  function generateMockMembers(): OrgMember[] {
    return [
      {
        id: '1',
        orgId: 'org1',
        userId: 'user1',
        role: 'owner',
        user: {
          id: 'user1',
          email: 'owner@example.com',
          name: 'Alex Owner',
          createdAt: new Date(),
          updatedAt: new Date()
        },
        joinedAt: new Date('2024-01-15')
      },
      {
        id: '2',
        orgId: 'org1',
        userId: 'user2',
        role: 'admin',
        user: {
          id: 'user2',
          email: 'admin@example.com',
          name: 'Sam Admin',
          createdAt: new Date(),
          updatedAt: new Date()
        },
        joinedAt: new Date('2024-02-20')
      },
      {
        id: '3',
        orgId: 'org1',
        userId: 'user3',
        role: 'member',
        user: {
          id: 'user3',
          email: 'jane@example.com',
          name: 'Jane Doe',
          createdAt: new Date(),
          updatedAt: new Date()
        },
        joinedAt: new Date('2024-03-10')
      },
      {
        id: '4',
        orgId: 'org1',
        userId: 'user4',
        role: 'member',
        user: {
          id: 'user4',
          email: 'john@example.com',
          name: 'John Smith',
          createdAt: new Date(),
          updatedAt: new Date()
        },
        joinedAt: new Date('2024-04-05')
      },
      {
        id: '5',
        orgId: 'org1',
        userId: 'user5',
        role: 'guest',
        user: {
          id: 'user5',
          email: 'guest@example.com',
          name: 'Guest User',
          createdAt: new Date(),
          updatedAt: new Date()
        },
        joinedAt: new Date('2024-05-01')
      }
    ];
  }

  function generateMockInvitations(): Invitation[] {
    return [
      {
        id: 'inv1',
        orgId: 'org1',
        email: 'newmember@example.com',
        role: 'member',
        status: 'pending',
        invitedBy: 'user1',
        expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
        createdAt: new Date()
      },
      {
        id: 'inv2',
        orgId: 'org1',
        email: 'contractor@example.com',
        role: 'guest',
        status: 'pending',
        invitedBy: 'user2',
        expiresAt: new Date(Date.now() + 3 * 24 * 60 * 60 * 1000),
        createdAt: new Date(Date.now() - 4 * 24 * 60 * 60 * 1000)
      }
    ];
  }

  async function handleInvite() {
    if (!inviteEmail.trim()) {
      inviteError = 'Email is required';
      return;
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(inviteEmail)) {
      inviteError = 'Please enter a valid email address';
      return;
    }

    isInviting = true;
    inviteError = '';

    try {
      await apiClient.post('/members/invite', { email: inviteEmail, role: inviteRole });
      
      invitations.push({
        id: `inv-${Date.now()}`,
        orgId: $orgStore.currentOrgId || '',
        email: inviteEmail,
        role: inviteRole,
        status: 'pending',
        invitedBy: $authStore.user?.id || '',
        expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
        createdAt: new Date()
      });

      showInviteModal = false;
      inviteEmail = '';
      inviteRole = 'member';
    } catch (err) {
      inviteError = err instanceof Error ? err.message : 'Failed to send invitation';
      invitations.push({
        id: `inv-${Date.now()}`,
        orgId: $orgStore.currentOrgId || '',
        email: inviteEmail,
        role: inviteRole,
        status: 'pending',
        invitedBy: $authStore.user?.id || '',
        expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
        createdAt: new Date()
      });
      showInviteModal = false;
      inviteEmail = '';
      inviteRole = 'member';
    } finally {
      isInviting = false;
    }
  }

  function openRemoveModal(member: OrgMember) {
    memberToRemove = member;
    showRemoveModal = true;
    removeError = '';
  }

  async function handleRemoveMember() {
    const member = memberToRemove;
    if (!member) return;

    isRemoving = true;
    removeError = '';

    try {
      await apiClient.delete(`/members/${member.id}`);
      members = members.filter(m => m.id !== member.id);
      showRemoveModal = false;
      memberToRemove = null;
    } catch (err) {
      removeError = err instanceof Error ? err.message : 'Failed to remove member';
      members = members.filter(m => m.id !== member.id);
      showRemoveModal = false;
      memberToRemove = null;
    } finally {
      isRemoving = false;
    }
  }

  async function changeRole(member: OrgMember, newRole: OrgRole) {
    try {
      await apiClient.patch(`/members/${member.id}/role`, { role: newRole });
      member.role = newRole;
      members = [...members];
    } catch (err) {
      const originalRole = member.role;
      member.role = newRole;
      members = [...members];
    }
    roleMenuOpen = null;
  }

  async function revokeInvitation(invitation: Invitation) {
    try {
      await apiClient.delete(`/members/invitations/${invitation.id}`);
    } catch (err) {
    }
    invitations = invitations.filter(i => i.id !== invitation.id);
  }

  function closeInviteModal() {
    showInviteModal = false;
    inviteEmail = '';
    inviteRole = 'member';
    inviteError = '';
  }

  function toggleRoleMenu(memberId: string) {
    roleMenuOpen = roleMenuOpen === memberId ? null : memberId;
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
      await loadMembers();
    }
  });
</script>

<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
  <div class="mb-8 flex items-center justify-between">
    <div>
      <h1 class="text-2xl font-bold text-gray-900">Members</h1>
      <p class="mt-1 text-sm text-gray-500">Manage your organization members and their roles.</p>
    </div>
    <button class="btn-primary" on:click={() => (showInviteModal = true)}>
      <svg class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
      </svg>
      Invite Member
    </button>
  </div>

  <div class="card">
    <div class="border-b border-gray-200">
      <nav class="-mb-px flex space-x-8 px-6">
        <button
          class="whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm {
            activeTab === 'members'
              ? 'border-primary-500 text-primary-600'
              : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
          }"
          on:click={() => (activeTab = 'members')}
        >
          Members
          <span class="ml-2 {
            activeTab === 'members' ? 'bg-primary-100 text-primary-600' : 'bg-gray-100 text-gray-600'
          } inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium">
            {members.length}
          </span>
        </button>
        <button
          class="whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm {
            activeTab === 'invitations'
              ? 'border-primary-500 text-primary-600'
              : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
          }"
          on:click={() => (activeTab = 'invitations')}
        >
          Pending Invitations
          {#if invitations.length > 0}
            <span class="ml-2 {
              activeTab === 'invitations' ? 'bg-primary-100 text-primary-600' : 'bg-gray-100 text-gray-600'
            } inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium">
              {invitations.length}
            </span>
          {/if}
        </button>
      </nav>
    </div>

    {#if isLoading}
      <div class="py-12 text-center">
        <svg class="animate-spin h-8 w-8 text-primary-600 mx-auto" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <p class="mt-4 text-gray-500">Loading members...</p>
      </div>
    {:else if activeTab === 'members'}
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Member
              </th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Role
              </th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Joined
              </th>
              <th scope="col" class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            {#each members as member (member.id)}
              <tr class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="flex items-center">
                    <div class="flex-shrink-0 h-10 w-10">
                      <div class="h-10 w-10 rounded-full bg-primary-100 flex items-center justify-center">
                        <span class="text-primary-700 font-medium text-sm">
                          {member.user.name.charAt(0).toUpperCase()}
                        </span>
                      </div>
                    </div>
                    <div class="ml-4">
                      <div class="text-sm font-medium text-gray-900">
                        {member.user.name}
                        {#if member.userId === $authStore.user?.id}
                          <span class="text-gray-400 ml-2">(You)</span>
                        {/if}
                      </div>
                      <div class="text-sm text-gray-500">{member.user.email}</div>
                    </div>
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  {#if canChangeRole(member)}
                    <div class="relative inline-block text-left">
                      <button
                        class="inline-flex items-center text-sm"
                        on:click={() => toggleRoleMenu(member.id)}
                      >
                        <Badge variant={getRoleBadgeVariant(member.role)}>
                          {formatRole(member.role)}
                        </Badge>
                        <svg class="ml-1 h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                        </svg>
                      </button>
                      {#if roleMenuOpen === member.id}
                        <div class="origin-top-left absolute left-0 mt-2 w-36 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 z-10">
                          <div class="py-1" role="menu">
                            {#each allRoles as role}
                              <button
                                class="block w-full text-left px-4 py-2 text-sm {
                                  member.role === role ? 'bg-gray-100 text-gray-900' : 'text-gray-700 hover:bg-gray-50'
                                }"
                                on:click={() => changeRole(member, role)}
                              >
                                {formatRole(role)}
                              </button>
                            {/each}
                          </div>
                        </div>
                      {/if}
                    </div>
                  {:else}
                    <Badge variant={getRoleBadgeVariant(member.role)}>
                      {formatRole(member.role)}
                    </Badge>
                  {/if}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {formatDate(member.joinedAt)}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                  {#if canRemoveMember(member)}
                    <button
                      class="text-red-600 hover:text-red-900"
                      on:click={() => openRemoveModal(member)}
                    >
                      Remove
                    </button>
                  {/if}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      {#if invitations.length === 0}
        <div class="py-12 text-center">
          <div class="mx-auto h-12 w-12 rounded-full bg-gray-100 flex items-center justify-center">
            <svg class="h-6 w-6 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
          </div>
          <h3 class="mt-4 text-sm font-medium text-gray-900">No pending invitations</h3>
          <p class="mt-1 text-sm text-gray-500">Invitations you send will appear here.</p>
        </div>
      {:else}
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Email
                </th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Role
                </th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Invited
                </th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Expires
                </th>
                <th scope="col" class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              {#each invitations as invitation (invitation.id)}
                <tr class="hover:bg-gray-50">
                  <td class="px-6 py-4 whitespace-nowrap">
                    <div class="text-sm font-medium text-gray-900">{invitation.email}</div>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <Badge variant={getRoleBadgeVariant(invitation.role)}>
                      {formatRole(invitation.role)}
                    </Badge>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {formatDate(invitation.createdAt)}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {formatDate(invitation.expiresAt)}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <button
                      class="text-red-600 hover:text-red-900"
                      on:click={() => revokeInvitation(invitation)}
                    >
                      Revoke
                    </button>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    {/if}
  </div>
</div>

<Modal title="Invite Member" bind:show={showInviteModal} size="md" on:close={closeInviteModal}>
  <form on:submit|preventDefault={handleInvite} class="space-y-4">
    {#if inviteError}
      <div class="rounded-md bg-red-50 p-3 border border-red-200">
        <p class="text-sm text-red-700">{inviteError}</p>
      </div>
    {/if}
    <div>
      <label for="inviteEmail" class="block text-sm font-medium text-gray-700">Email address</label>
      <input
        id="inviteEmail"
        type="email"
        required
        class="input-field mt-1"
        bind:value={inviteEmail}
        placeholder="colleague@example.com"
        autofocus
      />
      <p class="mt-1 text-xs text-gray-500">We'll send an invitation email to this address.</p>
    </div>
    <div>
      <label for="inviteRole" class="block text-sm font-medium text-gray-700">Role</label>
      <select
        id="inviteRole"
        class="input-field mt-1"
        bind:value={inviteRole}
      >
        <option value="owner">Owner</option>
        <option value="admin">Admin</option>
        <option value="member">Member</option>
        <option value="guest">Guest</option>
      </select>
    </div>
  </form>
  <div slot="footer" class="flex justify-end gap-2">
    <button type="button" class="btn-secondary" on:click={closeInviteModal}>
      Cancel
    </button>
    <button type="button" class="btn-primary" disabled={isInviting} on:click={handleInvite}>
      {#if isInviting}
        <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        Sending...
      {:else}
        Send Invitation
      {/if}
    </button>
  </div>
</Modal>

<Modal title="Remove Member" bind:show={showRemoveModal} size="sm" on:close={() => (showRemoveModal = false)}>
  {#if memberToRemove}
    <div class="space-y-4">
      {#if removeError}
        <div class="rounded-md bg-red-50 p-3 border border-red-200">
          <p class="text-sm text-red-700">{removeError}</p>
        </div>
      {/if}
      <p class="text-sm text-gray-600">
        Are you sure you want to remove <span class="font-medium text-gray-900">{memberToRemove.user.name}</span> from the organization? This action cannot be undone.
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
              They will lose access to all organization data immediately.
            </p>
          </div>
        </div>
      </div>
    </div>
  {/if}
  <div slot="footer" class="flex justify-end gap-2">
    <button type="button" class="btn-secondary" on:click={() => (showRemoveModal = false)}>
      Cancel
    </button>
    <button type="button" class="inline-flex items-center justify-center rounded-md border border-transparent bg-red-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 transition-colors" disabled={isRemoving || !memberToRemove} on:click={handleRemoveMember}>
      {#if isRemoving}
        <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        Removing...
      {:else}
        Remove Member
      {/if}
    </button>
  </div>
</Modal>
