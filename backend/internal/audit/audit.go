package audit

import (
	"time"
)

type AuditLog struct {
	OrgID      string     `json:"org_id"`
	ID         string     `json:"id"`
	UserID     *string    `json:"user_id,omitempty"`
	Action     string     `json:"action"`
	EntityType string     `json:"entity_type"`
	EntityID   string     `json:"entity_id"`
	BeforeData *string    `json:"before_data,omitempty"`
	AfterData  *string    `json:"after_data,omitempty"`
	IPAddress  *string    `json:"ip_address,omitempty"`
	UserAgent  *string    `json:"user_agent,omitempty"`
	Archived   bool       `json:"archived"`
	CreatedAt  time.Time  `json:"created_at"`
}

type LogFilter struct {
	Action      string
	EntityType  string
	EntityID    string
	UserID      string
	Archived    *bool
	Limit       int
	Offset      int
}

const (
	ActionUserLogin           = "user.login"
	ActionUserLogout          = "user.logout"
	ActionOrgCreate           = "org.create"
	ActionOrgUpdate           = "org.update"
	ActionOrgDelete           = "org.delete"
	ActionOrgRestore          = "org.restore"
	ActionMemberInvite        = "member.invite"
	ActionMemberJoin          = "member.join"
	ActionMemberLeave         = "member.leave"
	ActionMemberRoleChange    = "member.role_change"
	ActionMemberRemove        = "member.remove"
	ActionPlanUpgrade         = "plan.upgrade"
	ActionPlanDowngrade       = "plan.downgrade"
	ActionSubscriptionCancel  = "subscription.cancel"
	ActionSubscriptionRenew   = "subscription.renew"
	ActionInvoicePaid         = "invoice.paid"
	ActionInvoiceRefund       = "invoice.refund"
	ActionQuotaExceeded       = "quota.exceeded"
	ActionWebhookSend         = "webhook.send"
)
