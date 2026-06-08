package billing

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
)

type planQuota struct {
	Members   int `json:"members"`
	APICalls  int `json:"api_calls"`
	StorageGB int `json:"storage_gb"`
	Projects  int `json:"projects"`
}

var builtInPlans = map[string]*Plan{
	"free": {
		ID:           "free",
		Name:         "Free",
		PriceMonthly: 0,
		PriceYearly:  0,
	},
	"team": {
		ID:           "team",
		Name:         "Team",
		PriceMonthly: 29,
		PriceYearly:  290,
	},
	"enterprise": {
		ID:           "enterprise",
		Name:         "Enterprise",
		PriceMonthly: 99,
		PriceYearly:  990,
	},
}

var planQuotas = map[string]planQuota{
	"free": {
		Members:   5,
		APICalls:  1000,
		StorageGB: 1,
		Projects:  3,
	},
	"team": {
		Members:   20,
		APICalls:  10000,
		StorageGB: 50,
		Projects:  -1,
	},
	"enterprise": {
		Members:   -1,
		APICalls:  100000,
		StorageGB: 1024,
		Projects:  -1,
	},
}

func init() {
	for id, plan := range builtInPlans {
		quotas := planQuotas[id]
		featuresJSON, _ := json.Marshal(quotas)
		plan.FeaturesJSON = string(featuresJSON)
	}
}

func (s *Service) GetPlan(ctx context.Context, planID string) (*Plan, error) {
	if plan, ok := builtInPlans[planID]; ok {
		return plan, nil
	}

	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, price_monthly, price_yearly, features_json
		FROM plans
		WHERE id = ?
	`, planID)

	plan := &Plan{}
	err := row.Scan(&plan.ID, &plan.Name, &plan.PriceMonthly, &plan.PriceYearly, &plan.FeaturesJSON)
	if err == sql.ErrNoRows {
		return nil, ErrPlanNotFound
	}
	if err != nil {
		return nil, err
	}

	return plan, nil
}

func (s *Service) ListPlans(ctx context.Context) ([]*Plan, error) {
	plans := make([]*Plan, 0, len(builtInPlans))
	for _, plan := range builtInPlans {
		plans = append(plans, plan)
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, price_monthly, price_yearly, features_json
		FROM plans
	`)
	if err != nil {
		return plans, nil
	}
	defer rows.Close()

	seen := make(map[string]bool)
	for _, p := range plans {
		seen[p.ID] = true
	}

	for rows.Next() {
		plan := &Plan{}
		err := rows.Scan(&plan.ID, &plan.Name, &plan.PriceMonthly, &plan.PriceYearly, &plan.FeaturesJSON)
		if err != nil {
			return plans, err
		}
		if !seen[plan.ID] {
			plans = append(plans, plan)
		}
	}

	return plans, nil
}

func (s *Service) GetPlanQuotas(ctx context.Context, planID string) (map[string]int, error) {
	quotas, ok := planQuotas[planID]
	if ok {
		return map[string]int{
			MetricMembers:   quotas.Members,
			MetricAPICalls:  quotas.APICalls,
			MetricStorageGB: quotas.StorageGB,
			MetricProjects:  quotas.Projects,
		}, nil
	}

	plan, err := s.GetPlan(ctx, planID)
	if err != nil {
		return nil, err
	}

	var q planQuota
	if err := json.Unmarshal([]byte(plan.FeaturesJSON), &q); err != nil {
		return nil, err
	}

	return map[string]int{
		MetricMembers:   q.Members,
		MetricAPICalls:  q.APICalls,
		MetricStorageGB: q.StorageGB,
		MetricProjects:  q.Projects,
	}, nil
}

func (s *Service) CreatePlan(ctx context.Context, plan *Plan) (*Plan, error) {
	if plan.ID == "" {
		plan.ID = uuid.New().String()
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO plans (id, name, price_monthly, price_yearly, features_json)
		VALUES (?, ?, ?, ?, ?)
	`, plan.ID, plan.Name, plan.PriceMonthly, plan.PriceYearly, plan.FeaturesJSON)
	if err != nil {
		return nil, err
	}

	return plan, nil
}
