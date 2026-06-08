package billing

import (
	"context"
	"math/rand"
	"time"
)

func (s *Service) UpgradePlan(ctx context.Context, newPlanID string) (*Subscription, error) {
	sub, err := s.GetSubscription(ctx)
	if err != nil {
		return nil, err
	}

	currentPlan, err := s.GetPlan(ctx, sub.PlanID)
	if err != nil {
		return nil, err
	}

	newPlan, err := s.GetPlan(ctx, newPlanID)
	if err != nil {
		return nil, err
	}

	if newPlan.PriceMonthly <= currentPlan.PriceMonthly {
		return s.DowngradePlan(ctx, newPlanID)
	}

	proratedAmount, err := s.CalculateProratedAmount(ctx, sub.PlanID, newPlanID, sub.CurrentPeriodStart, sub.CurrentPeriodEnd)
	if err != nil {
		return nil, err
	}

	if proratedAmount > 0 {
		_, err = s.CreateInvoice(ctx, sub.ID, proratedAmount)
		if err != nil {
			return nil, err
		}
	}

	updatedSub, err := s.UpdatePlan(ctx, newPlanID)
	if err != nil {
		return nil, err
	}

	return updatedSub, nil
}

func (s *Service) DowngradePlan(ctx context.Context, newPlanID string) (*Subscription, error) {
	sub, err := s.GetSubscription(ctx)
	if err != nil {
		return nil, err
	}

	currentPlan, err := s.GetPlan(ctx, sub.PlanID)
	if err != nil {
		return nil, err
	}

	newPlan, err := s.GetPlan(ctx, newPlanID)
	if err != nil {
		return nil, err
	}

	if newPlan.PriceMonthly >= currentPlan.PriceMonthly {
		return s.UpgradePlan(ctx, newPlanID)
	}

	proratedAmount, err := s.CalculateProratedAmount(ctx, sub.PlanID, newPlanID, sub.CurrentPeriodStart, sub.CurrentPeriodEnd)
	if err != nil {
		return nil, err
	}

	if proratedAmount < 0 {
		refundAmount := -proratedAmount
		creditInvoice, err := s.CreateInvoice(ctx, sub.ID, -refundAmount)
		if err != nil {
			return nil, err
		}

		if err := s.RefundInvoice(ctx, creditInvoice.ID, refundAmount); err != nil {
			return nil, err
		}
	}

	updatedSub, err := s.UpdatePlan(ctx, newPlanID)
	if err != nil {
		return nil, err
	}

	return updatedSub, nil
}

func (s *Service) CalculateProratedAmount(ctx context.Context, currentPlanID, newPlanID string, currentPeriodStart, currentPeriodEnd time.Time) (float64, error) {
	currentPlan, err := s.GetPlan(ctx, currentPlanID)
	if err != nil {
		return 0, err
	}

	newPlan, err := s.GetPlan(ctx, newPlanID)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	totalDays := currentPeriodEnd.Sub(currentPeriodStart).Hours() / 24
	daysRemaining := currentPeriodEnd.Sub(now).Hours() / 24

	if totalDays <= 0 {
		return 0, nil
	}

	priceDiff := newPlan.PriceMonthly - currentPlan.PriceMonthly
	proratedAmount := priceDiff * (daysRemaining / totalDays)

	return proratedAmount, nil
}

func (s *Service) ProcessSubscriptionRenewal(ctx context.Context) error {
	sub, err := s.GetSubscription(ctx)
	if err != nil {
		return err
	}

	if sub.Status != SubscriptionStatusActive {
		return nil
	}

	now := time.Now()
	if now.Before(sub.CurrentPeriodEnd) {
		return nil
	}

	plan, err := s.GetPlan(ctx, sub.PlanID)
	if err != nil {
		return err
	}

	invoice, err := s.CreateInvoice(ctx, sub.ID, plan.PriceMonthly)
	if err != nil {
		return err
	}

	if err := s.PayInvoice(ctx, invoice.ID); err != nil {
		_, updateErr := s.db.ExecContext(ctx, `
			UPDATE subscriptions
			SET status = ?
			WHERE org_id = ? AND id = ?
		`, SubscriptionStatusPastDue, sub.OrgID, sub.ID)
		if updateErr != nil {
			return updateErr
		}
		return err
	}

	return s.RenewSubscription(ctx)
}

func (s *Service) MockPayment(ctx context.Context, invoiceID string) (bool, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	successRate := 0.95

	return r.Float64() < successRate, nil
}

func (s *Service) RefundInvoice(ctx context.Context, invoiceID string, amount float64) error {
	orgID, err := getOrgID(ctx)
	if err != nil {
		return err
	}

	invoice, err := s.GetInvoice(ctx, invoiceID)
	if err != nil {
		return err
	}

	if invoice.Status == InvoiceStatusRefunded {
		return nil
	}

	now := time.Now()
	_, err = s.db.ExecContext(ctx, `
		UPDATE invoices
		SET status = ?, paid_at = ?
		WHERE org_id = ? AND id = ?
	`, InvoiceStatusRefunded, now, orgID, invoiceID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ProcessRenewals(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT org_id
		FROM subscriptions
		WHERE status = ? AND date(current_period_end) <= date('now')
	`, SubscriptionStatusActive)
	if err != nil {
		return err
	}
	defer rows.Close()

	var orgIDs []string
	for rows.Next() {
		var orgID string
		if err := rows.Scan(&orgID); err != nil {
			return err
		}
		orgIDs = append(orgIDs, orgID)
	}

	for _, orgID := range orgIDs {
		subCtx := WithOrgID(ctx, orgID)
		_ = s.ProcessSubscriptionRenewal(subCtx)
	}

	return nil
}

func (s *Service) RetryFailedPayments(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `
		SELECT org_id, id
		FROM invoices
		WHERE status = ?
	`, InvoiceStatusFailed)
	if err != nil {
		return err
	}
	defer rows.Close()

	type invoiceInfo struct {
		OrgID     string
		InvoiceID string
	}

	var invoices []invoiceInfo
	for rows.Next() {
		var ii invoiceInfo
		if err := rows.Scan(&ii.OrgID, &ii.InvoiceID); err != nil {
			return err
		}
		invoices = append(invoices, ii)
	}

	for _, ii := range invoices {
		subCtx := WithOrgID(ctx, ii.OrgID)
		_ = s.PayInvoice(subCtx, ii.InvoiceID)
	}

	return nil
}
