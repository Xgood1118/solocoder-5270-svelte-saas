package webhook

import (
	"time"
)

type WebhookEvent struct {
	ID            string     `json:"id"`
	OrgID         string     `json:"org_id"`
	EventType     string     `json:"event_type"`
	Payload       string     `json:"payload"`
	Status        string     `json:"status"`
	Attempts      int        `json:"attempts"`
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type WebhookEndpoint struct {
	OrgID     string    `json:"org_id"`
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Secret    string    `json:"-"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

const (
	EventOrgCreated           = "org.created"
	EventMemberInvited        = "member.invited"
	EventMemberJoined         = "member.joined"
	EventPlanUpgraded         = "plan.upgraded"
	EventPlanDowngraded       = "plan.downgraded"
	EventSubscriptionCanceled = "subscription.canceled"
	EventInvoicePaid          = "invoice.paid"
	EventQuotaExceeded        = "quota.exceeded"
)

const (
	StatusPending = "pending"
	StatusSuccess = "success"
	StatusFailed  = "failed"
	StatusDead    = "dead"
)

const (
	MaxRetries  = 3
	RetryDelay1 = 1 * time.Second
	RetryDelay2 = 2 * time.Second
	RetryDelay4 = 4 * time.Second
)
