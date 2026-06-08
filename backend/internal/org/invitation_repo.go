package org

import (
	"context"
	"database/sql"
	"time"

	"saas-system/internal/db"
)

func (s *Service) createInvitation(ctx context.Context, invitation *Invitation) error {
	ctxOrgID, ok := db.OrgIDFromContext(ctx)
	if !ok || ctxOrgID != invitation.OrgID {
		return ErrPermissionDenied
	}

	now := time.Now()
	invitation.CreatedAt = now

	query := `
		INSERT INTO invitations (id, org_id, email, role, token, expires_at, invited_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		invitation.ID,
		invitation.OrgID,
		invitation.Email,
		invitation.Role,
		invitation.Token,
		invitation.ExpiresAt,
		invitation.InvitedBy,
		invitation.CreatedAt,
	)

	return err
}

func (s *Service) getInvitationByToken(ctx context.Context, token string) (*Invitation, error) {
	query := `
		SELECT id, org_id, email, role, token, expires_at, accepted_at, invited_by, created_at
		FROM invitations
		WHERE token = ?
	`

	invitation := &Invitation{}
	var acceptedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, query, token).Scan(
		&invitation.ID,
		&invitation.OrgID,
		&invitation.Email,
		&invitation.Role,
		&invitation.Token,
		&invitation.ExpiresAt,
		&acceptedAt,
		&invitation.InvitedBy,
		&invitation.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrInvitationNotFound
	}

	if err != nil {
		return nil, err
	}

	if acceptedAt.Valid {
		invitation.AcceptedAt = &acceptedAt.Time
	}

	if time.Now().After(invitation.ExpiresAt) {
		return nil, ErrInvitationExpired
	}

	return invitation, nil
}

func (s *Service) getInvitationByID(ctx context.Context, id string) (*Invitation, error) {
	orgID, ok := db.OrgIDFromContext(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}

	query := `
		SELECT id, org_id, email, role, token, expires_at, accepted_at, invited_by, created_at
		FROM invitations
		WHERE id = ? AND org_id = ?
	`

	invitation := &Invitation{}
	var acceptedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, query, id, orgID).Scan(
		&invitation.ID,
		&invitation.OrgID,
		&invitation.Email,
		&invitation.Role,
		&invitation.Token,
		&invitation.ExpiresAt,
		&acceptedAt,
		&invitation.InvitedBy,
		&invitation.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrInvitationNotFound
	}

	if err != nil {
		return nil, err
	}

	if acceptedAt.Valid {
		invitation.AcceptedAt = &acceptedAt.Time
	}

	return invitation, nil
}

func (s *Service) listInvitations(ctx context.Context, orgID string) ([]*Invitation, error) {
	ctxOrgID, ok := db.OrgIDFromContext(ctx)
	if !ok || ctxOrgID != orgID {
		return nil, ErrPermissionDenied
	}

	query := `
		SELECT id, org_id, email, role, token, expires_at, accepted_at, invited_by, created_at
		FROM invitations
		WHERE org_id = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []*Invitation
	for rows.Next() {
		invitation := &Invitation{}
		var acceptedAt sql.NullTime

		err := rows.Scan(
			&invitation.ID,
			&invitation.OrgID,
			&invitation.Email,
			&invitation.Role,
			&invitation.Token,
			&invitation.ExpiresAt,
			&acceptedAt,
			&invitation.InvitedBy,
			&invitation.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if acceptedAt.Valid {
			invitation.AcceptedAt = &acceptedAt.Time
		}

		invitations = append(invitations, invitation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return invitations, nil
}

func (s *Service) acceptInvitation(ctx context.Context, token, userID string) error {
	now := time.Now()

	query := `
		UPDATE invitations
		SET accepted_at = ?
		WHERE token = ? AND accepted_at IS NULL AND expires_at > ?
	`

	result, err := s.db.ExecContext(ctx, query, now, token, now)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrInvitationNotFound
	}

	return nil
}

func (s *Service) deleteInvitation(ctx context.Context, id string) error {
	orgID, ok := db.OrgIDFromContext(ctx)
	if !ok {
		return ErrPermissionDenied
	}

	query := `
		DELETE FROM invitations
		WHERE id = ? AND org_id = ?
	`

	result, err := s.db.ExecContext(ctx, query, id, orgID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrInvitationNotFound
	}

	return nil
}

func (s *Service) expireInvitations(ctx context.Context) error {
	query := `
		DELETE FROM invitations
		WHERE expires_at < ? AND accepted_at IS NULL
	`

	_, err := s.db.ExecContext(ctx, query, time.Now())
	return err
}
