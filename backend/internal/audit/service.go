package audit

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

type ListAuditLogsParams struct {
	OrgID      string
	Action     string
	EntityType string
	EntityID   string
	UserID     string
	Limit      int
	Offset     int
}

func (s *Service) ListAuditLogs(ctx context.Context, params ListAuditLogsParams) ([]AuditLog, int, error) {
	query := `SELECT org_id, id, user_id, action, entity_type, entity_id, before_data, after_data, ip_address, user_agent, created_at 
	          FROM audit_logs WHERE org_id = ?`
	countQuery := `SELECT COUNT(*) FROM audit_logs WHERE org_id = ?`

	args := []interface{}{params.OrgID}
	countArgs := []interface{}{params.OrgID}

	if params.Action != "" {
		query += " AND action = ?"
		countQuery += " AND action = ?"
		args = append(args, params.Action)
		countArgs = append(countArgs, params.Action)
	}

	if params.EntityType != "" {
		query += " AND entity_type = ?"
		countQuery += " AND entity_type = ?"
		args = append(args, params.EntityType)
		countArgs = append(countArgs, params.EntityType)
	}

	if params.EntityID != "" {
		query += " AND entity_id = ?"
		countQuery += " AND entity_id = ?"
		args = append(args, params.EntityID)
		countArgs = append(countArgs, params.EntityID)
	}

	if params.UserID != "" {
		query += " AND user_id = ?"
		countQuery += " AND user_id = ?"
		args = append(args, params.UserID)
		countArgs = append(countArgs, params.UserID)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	args = append(args, limit, params.Offset)

	var total int
	err := s.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		var userID, beforeData, afterData, ipAddress, userAgent sql.NullString
		err := rows.Scan(
			&log.OrgID, &log.ID, &userID, &log.Action, &log.EntityType, &log.EntityID,
			&beforeData, &afterData, &ipAddress, &userAgent, &log.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		if userID.Valid {
			log.UserID = &userID.String
		}
		if beforeData.Valid {
			log.BeforeData = &beforeData.String
		}
		if afterData.Valid {
			log.AfterData = &afterData.String
		}
		if ipAddress.Valid {
			log.IPAddress = &ipAddress.String
		}
		if userAgent.Valid {
			log.UserAgent = &userAgent.String
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}

func (s *Service) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	log.ID = uuid.New().String()
	log.CreatedAt = time.Now()

	var userID, beforeData, afterData, ipAddress, userAgent interface{}
	if log.UserID != nil {
		userID = *log.UserID
	}
	if log.BeforeData != nil {
		beforeData = *log.BeforeData
	}
	if log.AfterData != nil {
		afterData = *log.AfterData
	}
	if log.IPAddress != nil {
		ipAddress = *log.IPAddress
	}
	if log.UserAgent != nil {
		userAgent = *log.UserAgent
	}

	_, err := s.db.ExecContext(ctx,
		"INSERT INTO audit_logs (org_id, id, user_id, action, entity_type, entity_id, before_data, after_data, ip_address, user_agent, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		log.OrgID, log.ID, userID, log.Action, log.EntityType, log.EntityID,
		beforeData, afterData, ipAddress, userAgent, log.CreatedAt,
	)
	return err
}

func (s *Service) ArchiveLogs(ctx context.Context, orgID string, before time.Time) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM audit_logs WHERE org_id = ? AND created_at < ?",
		orgID, before,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
