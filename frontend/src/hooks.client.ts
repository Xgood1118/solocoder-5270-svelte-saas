import { authStore } from '$stores/auth';
import { orgStore } from '$stores/org';

export function handleClientLoad() {
  authStore.initializeFromStorage();
  orgStore.initializeFromStorage();
}

if (typeof window !== 'undefined') {
  handleClientLoad();
}
