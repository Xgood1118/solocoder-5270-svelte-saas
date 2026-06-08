package audit

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) Log(ctx context.Context, orgID, userID, action, entityType, entityID, beforeData, afterData, ip, userAgent string) error {
	id := uuid.New().String()

	query := `INSERT INTO audit_logs (org_id, id, user_id, action, entity_type, entity_id, before_data, after_data, ip_address, user_agent) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}

	var beforeDataPtr *string
	if beforeData != "" {
		beforeDataPtr = &beforeData
	}

	var afterDataPtr *string
	if afterData != "" {
		afterDataPtr = &afterData
	}

	var ipPtr *string
	if ip != "" {
		ipPtr = &ip
	}

	var userAgentPtr *string
	if userAgent != "" {
		userAgentPtr = &userAgent
	}

	_, err := s.db.ExecContext(ctx, query, orgID, id, userIDPtr, action, entityType, entityID, beforeDataPtr, afterDataPtr, ipPtr, userAgentPtr)
	return err
}

func (s *Service) ListLogs(ctx context.Context, orgID string, filter LogFilter) ([]*AuditLog, int, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "org_id = ?")
	args = append(args, orgID)

	if filter.Action != "" {
		conditions = append(conditions, "action = ?")
		args = append(args, filter.Action)
	}

	if filter.EntityType != "" {
		conditions = append(conditions, "entity_type = ?")
		args = append(args, filter.EntityType)
	}

	if filter.EntityID != "" {
		conditions = append(conditions, "entity_id = ?")
		args = append(args, filter.EntityID)
	}

	if filter.UserID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, filter.UserID)
	}

	if filter.Archived != nil {
		conditions = append(conditions, "archived = ?")
		args = append(args, *filter.Archived)
	}

	whereClause := strings.Join(conditions, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs WHERE %s", whereClause)
	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset

	query := fmt.Sprintf(`SELECT org_id, id, user_id, action, entity_type, entity_id, before_data, after_data, ip_address, user_agent, archived, created_at FROM audit_logs WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?`, whereClause)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	logs := make([]*AuditLog, 0)
	for rows.Next() {
		log := &AuditLog{}
		var userID, beforeData, afterData, ipAddress, userAgent sql.NullString
		err := rows.Scan(&log.OrgID, &log.ID, &userID, &log.Action, &log.EntityType, &log.EntityID, &beforeData, &afterData, &ipAddress, &userAgent, &log.Archived, &log.CreatedAt)
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

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (s *Service) GetLog(ctx context.Context, orgID, id string) (*AuditLog, error) {
	query := `SELECT org_id, id, user_id, action, entity_type, entity_id, before_data, after_data, ip_address, user_agent, archived, created_at FROM audit_logs WHERE org_id = ? AND id = ?`

	log := &AuditLog{}
	var userID, beforeData, afterData, ipAddress, userAgent sql.NullString

	err := s.db.QueryRowContext(ctx, query, orgID, id).Scan(
		&log.OrgID, &log.ID, &userID, &log.Action, &log.EntityType, &log.EntityID,
		&beforeData, &afterData, &ipAddress, &userAgent, &log.Archived, &log.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("audit log not found")
		}
		return nil, err
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

	return log, nil
}

func (s *Service) LogWithTx(tx *sql.Tx, ctx context.Context, orgID, userID, action, entityType, entityID, beforeData, afterData, ip, userAgent string) error {
	id := uuid.New().String()

	query := `INSERT INTO audit_logs (org_id, id, user_id, action, entity_type, entity_id, before_data, after_data, ip_address, user_agent) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}

	var beforeDataPtr *string
	if beforeData != "" {
		beforeDataPtr = &beforeData
	}

	var afterDataPtr *string
	if afterData != "" {
		afterDataPtr = &afterData
	}

	var ipPtr *string
	if ip != "" {
		ipPtr = &ip
	}

	var userAgentPtr *string
	if userAgent != "" {
		userAgentPtr = &userAgent
	}

	_, err := tx.ExecContext(ctx, query, orgID, id, userIDPtr, action, entityType, entityID, beforeDataPtr, afterDataPtr, ipPtr, userAgentPtr)
	return err
}
