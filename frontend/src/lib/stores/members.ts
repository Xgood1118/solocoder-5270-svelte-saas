import { writable, derived } from 'svelte/store';
import type {
  OrgMember,
  Invitation,
  InviteMemberRequest,
  UpdateMemberRoleRequest,
  AcceptInvitationRequest
} from '$types';
import { apiClient } from '$api/client';

function createMembersStore() {
  const members = writable<OrgMember[]>([]);
  const invitations = writable<Invitation[]>([]);
  const isLoading = writable(false);
  const error = writable<string | null>(null);

  const memberCount = derived(members, ($members) => $members.length);
  const pendingInvitations = derived(invitations, ($invitations) =>
    $invitations.filter((inv) => inv.status === 'pending')
  );

  const store = derived(
    [members, invitations, isLoading, error, memberCount, pendingInvitations],
    ([$members, $invitations, $isLoading, $error, $memberCount, $pendingInvitations]) => ({
      members: $members,
      invitations: $invitations,
      isLoading: $isLoading,
      error: $error,
      memberCount: $memberCount,
      pendingInvitations: $pendingInvitations
    })
  );

  async function fetchMembers(): Promise<OrgMember[]> {
    isLoading.set(true);
    error.set(null);

    try {
      const response = await apiClient.get<OrgMember[]>('/members');
      members.set(response);
      return response;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch members';
      error.set(message);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  async function inviteMember(email: string, role: string): Promise<Invitation> {
    isLoading.set(true);
    error.set(null);

    try {
      const request: InviteMemberRequest = { email, role: role as OrgMember['role'] };
      const response = await apiClient.post<Invitation>('/invitations', request);
      invitations.update(($invitations) => [...$invitations, response]);
      return response;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to invite member';
      error.set(message);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  async function updateMemberRole(userId: string, newRole: string): Promise<OrgMember> {
    isLoading.set(true);
    error.set(null);

    try {
      const request: UpdateMemberRoleRequest = { role: newRole as OrgMember['role'] };
      const response = await apiClient.patch<OrgMember>(`/members/${userId}/role`, request);
      members.update(($members) =>
        $members.map((member) =>
          member.userId === userId ? { ...member, role: response.role } : member
        )
      );
      return response;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to update member role';
      error.set(message);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  async function removeMember(userId: string): Promise<void> {
    isLoading.set(true);
    error.set(null);

    try {
      await apiClient.delete(`/members/${userId}`);
      members.update(($members) =>
        $members.filter((member) => member.userId !== userId)
      );
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to remove member';
      error.set(message);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  async function fetchInvitations(): Promise<Invitation[]> {
    isLoading.set(true);
    error.set(null);

    try {
      const response = await apiClient.get<Invitation[]>('/invitations');
      invitations.set(response);
      return response;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch invitations';
      error.set(message);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  async function cancelInvitation(invitationId: string): Promise<void> {
    isLoading.set(true);
    error.set(null);

    try {
      await apiClient.delete(`/invitations/${invitationId}`);
      invitations.update(($invitations) =>
        $invitations.filter((inv) => inv.id !== invitationId)
      );
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to cancel invitation';
      error.set(message);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  async function acceptInvitation(token: string): Promise<OrgMember> {
    isLoading.set(true);
    error.set(null);

    try {
      const request: AcceptInvitationRequest = { token };
      const response = await apiClient.post<OrgMember>(
        `/invitations/${token}/accept`,
        request,
        { skipOrgHeader: true, skipAuth: true }
      );
      return response;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to accept invitation';
      error.set(message);
      throw err;
    } finally {
      isLoading.set(false);
    }
  }

  function clearMembers() {
    members.set([]);
    invitations.set([]);
    error.set(null);
  }

  return {
    subscribe: store.subscribe,
    members,
    invitations,
    isLoading,
    error,
    memberCount,
    pendingInvitations,
    fetchMembers,
    inviteMember,
    updateMemberRole,
    removeMember,
    fetchInvitations,
    cancelInvitation,
    acceptInvitation,
    clearMembers
  };
}

export const membersStore = createMembersStore();
export const {
  members,
  invitations,
  isLoading: membersLoading,
  error: membersError,
  memberCount,
  pendingInvitations
} = membersStore;

export default membersStore;
