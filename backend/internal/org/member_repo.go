package org

import (
	"context"
	"database/sql"
	"time"

	"saas-system/internal/db"
)

func (s *Service) addMember(ctx context.Context, orgID, userID, role string) error {
	ctxOrgID, ok := db.OrgIDFromContext(ctx)
	if !ok || ctxOrgID != orgID {
		return ErrPermissionDenied
	}

	now := time.Now()

	query := `
		INSERT INTO org_members (org_id, user_id, role, joined_at)
		VALUES (?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query, orgID, userID, role, now)
	return err
}

func (s *Service) removeMember(ctx context.Context, orgID, userID string) error {
	ctxOrgID, ok := db.OrgIDFromContext(ctx)
	if !ok || ctxOrgID != orgID {
		return ErrPermissionDenied
	}

	now := time.Now()

	query := `
		UPDATE org_members
		SET left_at = ?
		WHERE org_id = ? AND user_id = ? AND left_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, now, orgID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrMemberNotFound
	}

	return nil
}

func (s *Service) updateMemberRole(ctx context.Context, orgID, userID, newRole string) error {
	ctxOrgID, ok := db.OrgIDFromContext(ctx)
	if !ok || ctxOrgID != orgID {
		return ErrPermissionDenied
	}

	query := `
		UPDATE org_members
		SET role = ?
		WHERE org_id = ? AND user_id = ? AND left_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, newRole, orgID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrMemberNotFound
	}

	return nil
}

func (s *Service) getMember(ctx context.Context, orgID, userID string) (*OrgMember, error) {
	ctxOrgID, ok := db.OrgIDFromContext(ctx)
	if !ok || ctxOrgID != orgID {
		return nil, ErrPermissionDenied
	}

	query := `
		SELECT org_id, user_id, role, joined_at, left_at
		FROM org_members
		WHERE org_id = ? AND user_id = ? AND left_at IS NULL
	`

	member := &OrgMember{}
	var leftAt sql.NullTime

	err := s.db.QueryRowContext(ctx, query, orgID, userID).Scan(
		&member.OrgID,
		&member.UserID,
		&member.Role,
		&member.JoinedAt,
		&leftAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrMemberNotFound
	}

	if err != nil {
		return nil, err
	}

	if leftAt.Valid {
		member.LeftAt = &leftAt.Time
	}

	return member, nil
}

func (s *Service) listMembers(ctx context.Context, orgID string) ([]*OrgMember, error) {
	ctxOrgID, ok := db.OrgIDFromContext(ctx)
	if !ok || ctxOrgID != orgID {
		return nil, ErrPermissionDenied
	}

	query := `
		SELECT org_id, user_id, role, joined_at, left_at
		FROM org_members
		WHERE org_id = ? AND left_at IS NULL
		ORDER BY joined_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*OrgMember
	for rows.Next() {
		member := &OrgMember{}
		var leftAt sql.NullTime

		err := rows.Scan(
			&member.OrgID,
			&member.UserID,
			&member.Role,
			&member.JoinedAt,
			&leftAt,
		)
		if err != nil {
			return nil, err
		}

		if leftAt.Valid {
			member.LeftAt = &leftAt.Time
		}

		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

func (s *Service) getMemberRole(ctx context.Context, orgID, userID string) (string, error) {
	ctxOrgID, ok := db.OrgIDFromContext(ctx)
	if !ok || ctxOrgID != orgID {
		return "", ErrPermissionDenied
	}

	query := `
		SELECT role
		FROM org_members
		WHERE org_id = ? AND user_id = ? AND left_at IS NULL
	`

	var role string
	err := s.db.QueryRowContext(ctx, query, orgID, userID).Scan(&role)

	if err == sql.ErrNoRows {
		return "", ErrMemberNotFound
	}

	if err != nil {
		return "", err
	}

	return role, nil
}

func (s *Service) countMembers(ctx context.Context, orgID string) (int, error) {
	ctxOrgID, ok := db.OrgIDFromContext(ctx)
	if !ok || ctxOrgID != orgID {
		return 0, ErrPermissionDenied
	}

	query := `
		SELECT COUNT(*)
		FROM org_members
		WHERE org_id = ? AND left_at IS NULL
	`

	var count int
	err := s.db.QueryRowContext(ctx, query, orgID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
