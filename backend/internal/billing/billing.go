package billing

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrPlanNotFound         = errors.New("plan not found")
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrSubscriptionExists   = errors.New("subscription already exists")
	ErrInvoiceNotFound      = errors.New("invoice not found")
	ErrQuotaNotFound        = errors.New("quota not found")
	ErrQuotaExceeded        = errors.New("QUOTA_EXCEEDED")
	ErrInvalidPlan          = errors.New("invalid plan")
	ErrPaymentFailed        = errors.New("payment failed")
	ErrOrgIDRequired        = errors.New("org_id is required")
)

type contextKey string

const orgIDKey contextKey = "org_id"

func WithOrgID(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, orgIDKey, orgID)
}

func OrgIDFromContext(ctx context.Context) (string, bool) {
	orgID, ok := ctx.Value(orgIDKey).(string)
	return orgID, ok
}

type Plan struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	PriceMonthly float64 `json:"price_monthly"`
	PriceYearly  float64 `json:"price_yearly"`
	FeaturesJSON string  `json:"features_json"`
}

type Subscription struct {
	OrgID              string     `json:"org_id"`
	ID                 string     `json:"id"`
	PlanID             string     `json:"plan_id"`
	Status             string     `json:"status"`
	CurrentPeriodStart time.Time  `json:"current_period_start"`
	CurrentPeriodEnd   time.Time  `json:"current_period_end"`
	CanceledAt         *time.Time `json:"canceled_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

type Invoice struct {
	OrgID          string     `json:"org_id"`
	ID             string     `json:"id"`
	SubscriptionID string     `json:"subscription_id"`
	Amount         float64    `json:"amount"`
	Status         string     `json:"status"`
	PaidAt         *time.Time `json:"paid_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type UsageRecord struct {
	OrgID       string    `json:"org_id"`
	ID          string    `json:"id"`
	Metric      string    `json:"metric"`
	Quantity    int       `json:"quantity"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	CreatedAt   time.Time `json:"created_at"`
}

type Quota struct {
	OrgID       string    `json:"org_id"`
	ID          string    `json:"id"`
	Metric      string    `json:"metric"`
	Limit       int       `json:"limit"`
	Used        int       `json:"used"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

const (
	SubscriptionStatusActive   = "active"
	SubscriptionStatusCanceled = "canceled"
	SubscriptionStatusPastDue  = "past_due"
)

const (
	InvoiceStatusPending  = "pending"
	InvoiceStatusPaid     = "paid"
	InvoiceStatusFailed   = "failed"
	InvoiceStatusRefunded = "refunded"
)

const (
	MetricMembers   = "members"
	MetricAPICalls  = "api_calls"
	MetricStorageGB = "storage_gb"
	MetricProjects  = "projects"
)

func getOrgID(ctx context.Context) (string, error) {
	orgID, ok := OrgIDFromContext(ctx)
	if !ok || orgID == "" {
		return "", ErrOrgIDRequired
	}
	return orgID, nil
}
