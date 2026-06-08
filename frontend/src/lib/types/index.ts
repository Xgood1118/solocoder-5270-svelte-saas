export interface User {
  id: string;
  email: string;
  name: string;
  avatarUrl?: string;
  createdAt: Date;
  updatedAt: Date;
}

export type OrgRole = 'owner' | 'admin' | 'member' | 'guest';

export interface Org {
  id: string;
  name: string;
  slug: string;
  logoUrl?: string;
  plan: string;
  ownerId: string;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date;
}

export interface OrgMember {
  id: string;
  orgId: string;
  userId: string;
  role: OrgRole;
  user: User;
  joinedAt: Date;
  leftAt?: Date;
}

export type InvitationStatus = 'pending' | 'accepted' | 'declined' | 'expired';

export interface Invitation {
  id: string;
  orgId: string;
  email: string;
  role: OrgRole;
  status: InvitationStatus;
  invitedBy: string;
  expiresAt: Date;
  acceptedAt?: Date;
  createdAt: Date;
}

export interface Plan {
  id: string;
  name: string;
  description: string;
  priceMonthly: number;
  priceYearly: number;
  features: string[];
  isActive: boolean;
}

export type SubscriptionStatus = 'active' | 'canceled' | 'past_due' | 'trialing' | 'inactive';

export interface Subscription {
  id: string;
  orgId: string;
  planId: string;
  plan?: Plan;
  status: SubscriptionStatus;
  currentPeriodStart: Date;
  currentPeriodEnd: Date;
  canceledAt?: Date;
  trialEndsAt?: Date;
  createdAt: Date;
}

export type InvoiceStatus = 'paid' | 'pending' | 'failed' | 'refunded';

export interface Invoice {
  id: string;
  orgId: string;
  subscriptionId: string;
  amount: number;
  currency: string;
  status: InvoiceStatus;
  paidAt?: Date;
  createdAt: Date;
  pdfUrl?: string;
}

export interface Quota {
  id: string;
  orgId: string;
  metric: string;
  limit: number;
  used: number;
  periodStart: Date;
  periodEnd: Date;
}

export type AuditLogAction =
  | 'user.login'
  | 'user.logout'
  | 'org.created'
  | 'org.updated'
  | 'org.deleted'
  | 'member.invited'
  | 'member.joined'
  | 'member.removed'
  | 'member.role_updated'
  | 'subscription.created'
  | 'subscription.updated'
  | 'subscription.canceled'
  | 'invoice.paid'
  | 'invoice.failed'
  | 'settings.updated';

export interface AuditLog {
  id: string;
  orgId: string;
  userId?: string;
  user?: User;
  action: AuditLogAction;
  targetType?: string;
  targetId?: string;
  metadata?: Record<string, unknown>;
  ipAddress?: string;
  userAgent?: string;
  createdAt: Date;
}

export interface AuditLogPagination {
  page: number;
  perPage: number;
  total: number;
  totalPages: number;
}

export interface AuditLogFilterParams {
  action?: AuditLogAction;
  userId?: string;
  startDate?: Date;
  endDate?: Date;
}

export type WebhookEventType =
  | 'org.created'
  | 'member.invited'
  | 'member.joined'
  | 'plan.upgraded'
  | 'plan.downgraded'
  | 'subscription.canceled'
  | 'invoice.paid'
  | 'quota.exceeded';

export type WebhookEventStatus = 'pending' | 'success' | 'failed' | 'dead';

export interface WebhookEvent {
  id: string;
  orgId: string;
  eventType: WebhookEventType;
  payload: string;
  status: WebhookEventStatus;
  attempts: number;
  lastAttemptAt?: Date;
  createdAt: Date;
}

export interface WebhookEndpoint {
  id: string;
  orgId: string;
  url: string;
  secret?: string;
  active: boolean;
  createdAt: Date;
}

export interface CreateWebhookEndpointRequest {
  url: string;
  events: WebhookEventType[];
}

export interface UpdateWebhookEndpointRequest {
  url?: string;
  active?: boolean;
  events?: WebhookEventType[];
}

export interface ApiResponse<T = unknown> {
  data?: T;
  error?: string;
  message?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    page: number;
    perPage: number;
    total: number;
    totalPages: number;
  };
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
  orgName?: string;
}

export interface AuthResponse {
  user: User;
  token: string;
  orgs: Org[];
}

export interface CreateOrgRequest {
  name: string;
  slug?: string;
}

export interface UpdateOrgRequest {
  name?: string;
  slug?: string;
  logoUrl?: string;
}

export interface InviteMemberRequest {
  email: string;
  role: OrgRole;
}

export interface UpdateMemberRoleRequest {
  role: OrgRole;
}

export interface CreateCheckoutSessionRequest {
  planId: string;
  interval: 'month' | 'year';
  successUrl: string;
  cancelUrl: string;
}

export interface AcceptInvitationRequest {
  token: string;
}
