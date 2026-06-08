package billing

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

func (s *Service) CreateSubscription(ctx context.Context, orgID, planID string) (*Subscription, error) {
	_, err := s.GetPlan(ctx, planID)
	if err != nil {
		return nil, err
	}

	existing, err := s.GetSubscription(ctx)
	if err != nil && err != ErrSubscriptionNotFound {
		return nil, err
	}
	if existing != nil {
		return nil, ErrSubscriptionExists
	}

	now := time.Now()
	periodEnd := now.AddDate(0, 1, 0)

	sub := &Subscription{
		OrgID:              orgID,
		ID:                 uuid.New().String(),
		PlanID:             planID,
		Status:             SubscriptionStatusActive,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   periodEnd,
		CreatedAt:          now,
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO subscriptions (org_id, id, plan_id, status, current_period_start, current_period_end, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, sub.OrgID, sub.ID, sub.PlanID, sub.Status, sub.CurrentPeriodStart, sub.CurrentPeriodEnd, sub.CreatedAt)
	if err != nil {
		return nil, err
	}

	if err := s.InitQuotas(ctx, planID); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *Service) GetSubscription(ctx context.Context) (*Subscription, error) {
	orgID, err := getOrgID(ctx)
	if err != nil {
		return nil, err
	}

	row := s.db.QueryRowContext(ctx, `
		SELECT org_id, id, plan_id, status, current_period_start, current_period_end, canceled_at, created_at
		FROM subscriptions
		WHERE org_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, orgID)

	sub := &Subscription{}
	var canceledAt sql.NullTime
	err = row.Scan(&sub.OrgID, &sub.ID, &sub.PlanID, &sub.Status, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &canceledAt, &sub.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrSubscriptionNotFound
	}
	if err != nil {
		return nil, err
	}

	if canceledAt.Valid {
		sub.CanceledAt = &canceledAt.Time
	}

	return sub, nil
}

func (s *Service) UpdatePlan(ctx context.Context, newPlanID string) (*Subscription, error) {
	orgID, err := getOrgID(ctx)
	if err != nil {
		return nil, err
	}

	_, err = s.GetPlan(ctx, newPlanID)
	if err != nil {
		return nil, err
	}

	sub, err := s.GetSubscription(ctx)
	if err != nil {
		return nil, err
	}

	if sub.PlanID == newPlanID {
		return sub, nil
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE subscriptions
		SET plan_id = ?, status = ?
		WHERE org_id = ? AND id = ?
	`, newPlanID, SubscriptionStatusActive, orgID, sub.ID)
	if err != nil {
		return nil, err
	}

	sub.PlanID = newPlanID
	sub.Status = SubscriptionStatusActive

	if err := s.RecalculateQuotas(ctx, newPlanID); err != nil {
		return sub, err
	}

	return sub, nil
}

func (s *Service) CancelSubscription(ctx context.Context) error {
	orgID, err := getOrgID(ctx)
	if err != nil {
		return err
	}

	sub, err := s.GetSubscription(ctx)
	if err != nil {
		return err
	}

	now := time.Now()

	_, err = s.db.ExecContext(ctx, `
		UPDATE subscriptions
		SET status = ?, canceled_at = ?
		WHERE org_id = ? AND id = ?
	`, SubscriptionStatusCanceled, now, orgID, sub.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) RenewSubscription(ctx context.Context) error {
	orgID, err := getOrgID(ctx)
	if err != nil {
		return err
	}

	sub, err := s.GetSubscription(ctx)
	if err != nil {
		return err
	}

	newPeriodStart := sub.CurrentPeriodEnd
	newPeriodEnd := newPeriodStart.AddDate(0, 1, 0)

	_, err = s.db.ExecContext(ctx, `
		UPDATE subscriptions
		SET current_period_start = ?, current_period_end = ?, status = ?
		WHERE org_id = ? AND id = ?
	`, newPeriodStart, newPeriodEnd, SubscriptionStatusActive, orgID, sub.ID)
	if err != nil {
		return err
	}

	if err := s.ResetQuotas(ctx); err != nil {
		return err
	}

	plan, err := s.GetPlan(ctx, sub.PlanID)
	if err != nil {
		return err
	}

	_, err = s.CreateInvoice(ctx, sub.ID, plan.PriceMonthly)
	return err
}

func (s *Service) ListSubscriptions(ctx context.Context) ([]*Subscription, error) {
	orgID, err := getOrgID(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT org_id, id, plan_id, status, current_period_start, current_period_end, canceled_at, created_at
		FROM subscriptions
		WHERE org_id = ?
		ORDER BY created_at DESC
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*Subscription
	for rows.Next() {
		sub := &Subscription{}
		var canceledAt sql.NullTime
		err := rows.Scan(&sub.OrgID, &sub.ID, &sub.PlanID, &sub.Status, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &canceledAt, &sub.CreatedAt)
		if err != nil {
			return nil, err
		}
		if canceledAt.Valid {
			sub.CanceledAt = &canceledAt.Time
		}
		subs = append(subs, sub)
	}

	return subs, nil
}
