import type { Handle, HandleFetch } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
  const token = event.cookies.get('auth_token');
  const orgId = event.cookies.get('current_org_id');

  event.locals.authToken = token || null;
  event.locals.currentOrgId = orgId || null;

  const response = await resolve(event);
  return response;
};

export const handleFetch: HandleFetch = async ({ request, fetch, event }) => {
  if (event.locals.authToken) {
    request.headers.set('Authorization', `Bearer ${event.locals.authToken}`);
  }

  if (event.locals.currentOrgId) {
    request.headers.set('X-Org-Id', event.locals.currentOrgId);
  }

  return fetch(request);
};
