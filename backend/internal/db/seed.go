package db

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func Seed(db *sql.DB) error {
	exists, err := hasUsers(db)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now()

	userID := uuid.New().String()
	_, err = db.Exec(`
		INSERT INTO users (id, email, password_hash, name, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, userID, "demo@example.com", string(passwordHash), "Demo User", now, now)
	if err != nil {
		return err
	}

	user2ID := uuid.New().String()
	_, err = db.Exec(`
		INSERT INTO users (id, email, password_hash, name, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, user2ID, "admin@example.com", string(passwordHash), "Admin User", now, now)
	if err != nil {
		return err
	}

	orgID := uuid.New().String()
	_, err = db.Exec(`
		INSERT INTO orgs (id, name, slug, plan, owner_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, orgID, "Demo Org", "demo-org", "free", userID, now, now)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO org_members (org_id, user_id, role, joined_at)
		VALUES (?, ?, ?, ?)
	`, orgID, userID, "owner", now)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO org_members (org_id, user_id, role, joined_at)
		VALUES (?, ?, ?, ?)
	`, orgID, user2ID, "admin", now)
	if err != nil {
		return err
	}

	periodEnd := now.AddDate(0, 1, 0)

	quotas := []struct {
		metric string
		limit  int
	}{
		{"members", 5},
		{"api_calls", 1000},
		{"storage_gb", 1},
		{"projects", 3},
	}

	for _, q := range quotas {
		quotaID := uuid.New().String()
		_, err = db.Exec(`
			INSERT INTO quotas (org_id, id, metric, "limit", used, period_start, period_end)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, orgID, quotaID, q.metric, q.limit, 0, now, periodEnd)
		if err != nil {
			return err
		}
	}

	subID := uuid.New().String()
	_, err = db.Exec(`
		INSERT INTO subscriptions (org_id, id, plan_id, status, current_period_start, current_period_end, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, orgID, subID, "free", "active", now, periodEnd, now)
	if err != nil {
		return err
	}

	auditID := uuid.New().String()
	_, err = db.Exec(`
		INSERT INTO audit_logs (org_id, id, user_id, action, entity_type, entity_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, orgID, auditID, userID, "org.created", "org", orgID, now)
	if err != nil {
		return err
	}

	return nil
}

func hasUsers(db *sql.DB) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
