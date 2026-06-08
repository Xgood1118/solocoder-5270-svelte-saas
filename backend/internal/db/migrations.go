package db

import (
	"database/sql"
)

func Migrate(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			name TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS orgs (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			plan TEXT NOT NULL DEFAULT 'free',
			owner_id TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			FOREIGN KEY (owner_id) REFERENCES users(id)
		)`,

		`CREATE TABLE IF NOT EXISTS org_members (
			org_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			role TEXT NOT NULL CHECK(role IN ('owner', 'admin', 'member', 'guest')),
			joined_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			left_at DATETIME,
			PRIMARY KEY (org_id, user_id),
			FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,

		`CREATE TABLE IF NOT EXISTS invitations (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL,
			email TEXT NOT NULL,
			role TEXT NOT NULL CHECK(role IN ('admin', 'member', 'guest')),
			token TEXT NOT NULL UNIQUE,
			expires_at DATETIME NOT NULL,
			accepted_at DATETIME,
			invited_by TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE CASCADE,
			FOREIGN KEY (invited_by) REFERENCES users(id)
		)`,

		`CREATE TABLE IF NOT EXISTS plans (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			price_monthly REAL NOT NULL DEFAULT 0,
			price_yearly REAL NOT NULL DEFAULT 0,
			features_json TEXT NOT NULL DEFAULT '{}'
		)`,

		`CREATE TABLE IF NOT EXISTS subscriptions (
			org_id TEXT NOT NULL,
			id TEXT NOT NULL,
			plan_id TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			current_period_start DATETIME NOT NULL,
			current_period_end DATETIME NOT NULL,
			canceled_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (org_id, id),
			FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE CASCADE,
			FOREIGN KEY (plan_id) REFERENCES plans(id)
		)`,

		`CREATE TABLE IF NOT EXISTS invoices (
			org_id TEXT NOT NULL,
			id TEXT NOT NULL,
			subscription_id TEXT NOT NULL,
			amount REAL NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			paid_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (org_id, id),
			FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS usage_records (
			org_id TEXT NOT NULL,
			id TEXT NOT NULL,
			metric TEXT NOT NULL,
			quantity INTEGER NOT NULL DEFAULT 0,
			period_start DATETIME NOT NULL,
			period_end DATETIME NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (org_id, id),
			FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS quotas (
			org_id TEXT NOT NULL,
			id TEXT NOT NULL,
			metric TEXT NOT NULL,
			"limit" INTEGER NOT NULL DEFAULT 0,
			used INTEGER NOT NULL DEFAULT 0,
			period_start DATETIME NOT NULL,
			period_end DATETIME NOT NULL,
			PRIMARY KEY (org_id, id),
			FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS audit_logs (
			org_id TEXT NOT NULL,
			id TEXT NOT NULL,
			user_id TEXT,
			action TEXT NOT NULL,
			entity_type TEXT NOT NULL,
			entity_id TEXT NOT NULL,
			before_data TEXT,
			after_data TEXT,
			ip_address TEXT,
			user_agent TEXT,
			archived INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (org_id, id),
			FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,

		`CREATE TABLE IF NOT EXISTS webhook_events (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL,
			event_type TEXT NOT NULL,
			payload TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			attempts INTEGER NOT NULL DEFAULT 0,
			last_attempt_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS webhook_endpoints (
			org_id TEXT NOT NULL,
			id TEXT NOT NULL,
			url TEXT NOT NULL,
			secret TEXT NOT NULL,
			active INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (org_id, id),
			FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE CASCADE
		)`,

		`CREATE INDEX IF NOT EXISTS idx_org_members_user_id ON org_members(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_invitations_org_id ON invitations(org_id)`,
		`CREATE INDEX IF NOT EXISTS idx_invitations_token ON invitations(token)`,
		`CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(org_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(org_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_records_metric ON usage_records(org_id, metric)`,
		`CREATE INDEX IF NOT EXISTS idx_quotas_metric ON quotas(org_id, metric)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(org_id, action)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(org_id, entity_type, entity_id)`,
		`CREATE INDEX IF NOT EXISTS idx_webhook_events_status ON webhook_events(status)`,
		`CREATE INDEX IF NOT EXISTS idx_webhook_endpoints_active ON webhook_endpoints(org_id, active)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	if err := migrateAddArchivedColumn(db); err != nil {
		return err
	}

	return nil
}

func migrateAddArchivedColumn(db *sql.DB) error {
	rows, err := db.Query("PRAGMA table_info(audit_logs)")
	if err != nil {
		return err
	}
	defer rows.Close()

	hasArchived := false
	for rows.Next() {
		var cid int
		var name string
		var dtype string
		var notnull int
		var dfltValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &dtype, &notnull, &dfltValue, &pk); err != nil {
			return err
		}
		if name == "archived" {
			hasArchived = true
			break
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if !hasArchived {
		_, err := db.Exec("ALTER TABLE audit_logs ADD COLUMN archived INTEGER NOT NULL DEFAULT 0")
		if err != nil {
			return err
		}
	}

	return nil
}
