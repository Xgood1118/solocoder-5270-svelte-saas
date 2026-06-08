package billing

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

func (s *Service) CreateInvoice(ctx context.Context, subscriptionID string, amount float64) (*Invoice, error) {
	orgID, err := getOrgID(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	invoice := &Invoice{
		OrgID:          orgID,
		ID:             uuid.New().String(),
		SubscriptionID: subscriptionID,
		Amount:         amount,
		Status:         InvoiceStatusPending,
		CreatedAt:      now,
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO invoices (org_id, id, subscription_id, amount, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, invoice.OrgID, invoice.ID, invoice.SubscriptionID, invoice.Amount, invoice.Status, invoice.CreatedAt)
	if err != nil {
		return nil, err
	}

	return invoice, nil
}

func (s *Service) GetInvoice(ctx context.Context, invoiceID string) (*Invoice, error) {
	orgID, err := getOrgID(ctx)
	if err != nil {
		return nil, err
	}

	row := s.db.QueryRowContext(ctx, `
		SELECT org_id, id, subscription_id, amount, status, paid_at, created_at
		FROM invoices
		WHERE org_id = ? AND id = ?
	`, orgID, invoiceID)

	invoice := &Invoice{}
	var paidAt sql.NullTime
	err = row.Scan(&invoice.OrgID, &invoice.ID, &invoice.SubscriptionID, &invoice.Amount, &invoice.Status, &paidAt, &invoice.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrInvoiceNotFound
	}
	if err != nil {
		return nil, err
	}

	if paidAt.Valid {
		invoice.PaidAt = &paidAt.Time
	}

	return invoice, nil
}

func (s *Service) ListInvoices(ctx context.Context) ([]*Invoice, error) {
	orgID, err := getOrgID(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT org_id, id, subscription_id, amount, status, paid_at, created_at
		FROM invoices
		WHERE org_id = ?
		ORDER BY created_at DESC
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []*Invoice
	for rows.Next() {
		invoice := &Invoice{}
		var paidAt sql.NullTime
		err := rows.Scan(&invoice.OrgID, &invoice.ID, &invoice.SubscriptionID, &invoice.Amount, &invoice.Status, &paidAt, &invoice.CreatedAt)
		if err != nil {
			return nil, err
		}
		if paidAt.Valid {
			invoice.PaidAt = &paidAt.Time
		}
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

func (s *Service) PayInvoice(ctx context.Context, invoiceID string) error {
	orgID, err := getOrgID(ctx)
	if err != nil {
		return err
	}

	invoice, err := s.GetInvoice(ctx, invoiceID)
	if err != nil {
		return err
	}

	if invoice.Status == InvoiceStatusPaid {
		return nil
	}

	success, err := s.MockPayment(ctx, invoiceID)
	if err != nil {
		return err
	}

	if !success {
		_, err = s.db.ExecContext(ctx, `
			UPDATE invoices
			SET status = ?
			WHERE org_id = ? AND id = ?
		`, InvoiceStatusFailed, orgID, invoiceID)
		return ErrPaymentFailed
	}

	now := time.Now()
	_, err = s.db.ExecContext(ctx, `
		UPDATE invoices
		SET status = ?, paid_at = ?
		WHERE org_id = ? AND id = ?
	`, InvoiceStatusPaid, now, orgID, invoiceID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GenerateMonthlyInvoices(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT org_id, id, plan_id
		FROM subscriptions
		WHERE status = ? AND date(current_period_end) = date('now')
	`, SubscriptionStatusActive)
	if err != nil {
		return err
	}
	defer rows.Close()

	type subInfo struct {
		OrgID  string
		SubID  string
		PlanID string
	}

	var subs []subInfo
	for rows.Next() {
		var si subInfo
		if err := rows.Scan(&si.OrgID, &si.SubID, &si.PlanID); err != nil {
			return err
		}
		subs = append(subs, si)
	}

	for _, si := range subs {
		subCtx := WithOrgID(ctx, si.OrgID)
		plan, err := s.GetPlan(subCtx, si.PlanID)
		if err != nil {
			continue
		}

		_, err = s.CreateInvoice(subCtx, si.SubID, plan.PriceMonthly)
		if err != nil {
			continue
		}
	}

	return nil
}
