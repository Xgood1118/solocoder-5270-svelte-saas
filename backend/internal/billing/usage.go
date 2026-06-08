package billing

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (s *Service) RecordUsage(ctx context.Context, metric string, quantity int) error {
	orgID, err := getOrgID(ctx)
	if err != nil {
		return err
	}

	now := time.Now()
	periodStart := now.AddDate(0, 0, -now.Day()+1)
	periodEnd := periodStart.AddDate(0, 1, 0)

	record := &UsageRecord{
		OrgID:       orgID,
		ID:          uuid.New().String(),
		Metric:      metric,
		Quantity:    quantity,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		CreatedAt:   now,
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO usage_records (org_id, id, metric, quantity, period_start, period_end, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, record.OrgID, record.ID, record.Metric, record.Quantity, record.PeriodStart, record.PeriodEnd, record.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetUsageForPeriod(ctx context.Context, metric string, periodStart, periodEnd time.Time) (int, error) {
	orgID, err := getOrgID(ctx)
	if err != nil {
		return 0, err
	}

	row := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(quantity), 0)
		FROM usage_records
		WHERE org_id = ? AND metric = ? AND created_at >= ? AND created_at < ?
	`, orgID, metric, periodStart, periodEnd)

	var total int
	err = row.Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (s *Service) ListUsageRecords(ctx context.Context, metric string, limit int) ([]*UsageRecord, error) {
	orgID, err := getOrgID(ctx)
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 100
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT org_id, id, metric, quantity, period_start, period_end, created_at
		FROM usage_records
		WHERE org_id = ? AND metric = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, orgID, metric, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*UsageRecord
	for rows.Next() {
		record := &UsageRecord{}
		err := rows.Scan(&record.OrgID, &record.ID, &record.Metric, &record.Quantity, &record.PeriodStart, &record.PeriodEnd, &record.CreatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

func (s *Service) GetCurrentUsage(ctx context.Context, metric string) (int, error) {
	quota, err := s.GetQuota(ctx, metric)
	if err != nil {
		if err == ErrQuotaNotFound {
			return 0, nil
		}
		return 0, err
	}

	return quota.Used, nil
}
