package billing

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

func (s *Service) InitQuotas(ctx context.Context, planID string) error {
	orgID, err := getOrgID(ctx)
	if err != nil {
		return err
	}

	quotas, err := s.GetPlanQuotas(ctx, planID)
	if err != nil {
		return err
	}

	now := time.Now()
	periodEnd := now.AddDate(0, 1, 0)

	for metric, limit := range quotas {
		id := uuid.New().String()
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO quotas (org_id, id, metric, "limit", used, period_start, period_end)
			VALUES (?, ?, ?, ?, 0, ?, ?)
		`, orgID, id, metric, limit, now, periodEnd)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) GetQuota(ctx context.Context, metric string) (*Quota, error) {
	orgID, err := getOrgID(ctx)
	if err != nil {
		return nil, err
	}

	row := s.db.QueryRowContext(ctx, `
		SELECT org_id, id, metric, "limit", used, period_start, period_end
		FROM quotas
		WHERE org_id = ? AND metric = ?
		ORDER BY period_start DESC
		LIMIT 1
	`, orgID, metric)

	quota := &Quota{}
	err = row.Scan(&quota.OrgID, &quota.ID, &quota.Metric, &quota.Limit, &quota.Used, &quota.PeriodStart, &quota.PeriodEnd)
	if err == sql.ErrNoRows {
		return nil, ErrQuotaNotFound
	}
	if err != nil {
		return nil, err
	}

	return quota, nil
}

func (s *Service) ListQuotas(ctx context.Context) ([]*Quota, error) {
	orgID, err := getOrgID(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT org_id, id, metric, "limit", used, period_start, period_end
		FROM quotas
		WHERE org_id = ?
		ORDER BY metric
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quotas []*Quota
	for rows.Next() {
		quota := &Quota{}
		err := rows.Scan(&quota.OrgID, &quota.ID, &quota.Metric, &quota.Limit, &quota.Used, &quota.PeriodStart, &quota.PeriodEnd)
		if err != nil {
			return nil, err
		}
		quotas = append(quotas, quota)
	}

	return quotas, nil
}

func (s *Service) CheckQuota(ctx context.Context, metric string, amount int) (bool, error) {
	quota, err := s.GetQuota(ctx, metric)
	if err != nil {
		if err == ErrQuotaNotFound {
			return true, nil
		}
		return false, err
	}

	if quota.Limit < 0 {
		return true, nil
	}

	return quota.Used+amount <= quota.Limit, nil
}

func (s *Service) ConsumeQuota(ctx context.Context, metric string, amount int) error {
	orgID, err := getOrgID(ctx)
	if err != nil {
		return err
	}

	quota, err := s.GetQuota(ctx, metric)
	if err != nil {
		if err == ErrQuotaNotFound {
			return nil
		}
		return err
	}

	if quota.Limit >= 0 && quota.Used+amount > quota.Limit {
		return ErrQuotaExceeded
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE quotas
		SET used = used + ?
		WHERE org_id = ? AND id = ?
	`, amount, orgID, quota.ID)
	if err != nil {
		return err
	}

	if err := s.RecordUsage(ctx, metric, amount); err != nil {
		return err
	}

	return nil
}

func (s *Service) CheckAndConsumeQuota(ctx context.Context, metric string, amount int) (bool, error) {
	ok, err := s.CheckQuota(ctx, metric, amount)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	err = s.ConsumeQuota(ctx, metric, amount)
	if err != nil {
		if err == ErrQuotaExceeded {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s *Service) ResetQuotas(ctx context.Context) error {
	orgID, err := getOrgID(ctx)
	if err != nil {
		return err
	}

	_, err = s.GetSubscription(ctx)
	if err != nil {
		return err
	}

	now := time.Now()
	newPeriodEnd := now.AddDate(0, 1, 0)

	quotas, err := s.ListQuotas(ctx)
	if err != nil {
		return err
	}

	for _, quota := range quotas {
		_, err := s.db.ExecContext(ctx, `
			UPDATE quotas
			SET used = 0, period_start = ?, period_end = ?
			WHERE org_id = ? AND id = ?
		`, now, newPeriodEnd, orgID, quota.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) RecalculateQuotas(ctx context.Context, newPlanID string) error {
	orgID, err := getOrgID(ctx)
	if err != nil {
		return err
	}

	newQuotas, err := s.GetPlanQuotas(ctx, newPlanID)
	if err != nil {
		return err
	}

	existingQuotas, err := s.ListQuotas(ctx)
	if err != nil && err != ErrQuotaNotFound {
		return err
	}

	existingMap := make(map[string]*Quota)
	for _, q := range existingQuotas {
		existingMap[q.Metric] = q
	}

	now := time.Now()
	periodEnd := now.AddDate(0, 1, 0)

	for metric, limit := range newQuotas {
		if existing, ok := existingMap[metric]; ok {
			_, err := s.db.ExecContext(ctx, `
				UPDATE quotas
				SET "limit" = ?
				WHERE org_id = ? AND id = ?
			`, limit, orgID, existing.ID)
			if err != nil {
				return err
			}
		} else {
			id := uuid.New().String()
			_, err := s.db.ExecContext(ctx, `
				INSERT INTO quotas (org_id, id, metric, "limit", used, period_start, period_end)
				VALUES (?, ?, ?, ?, 0, ?, ?)
			`, orgID, id, metric, limit, now, periodEnd)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
