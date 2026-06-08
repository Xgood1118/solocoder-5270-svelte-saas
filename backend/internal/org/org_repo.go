package org

import (
	"context"
	"database/sql"
	"time"

	"saas-system/internal/db"
)

func (s *Service) createOrg(ctx context.Context, org *Org) error {
	now := time.Now()
	org.CreatedAt = now
	org.UpdatedAt = now

	query := `
		INSERT INTO orgs (id, name, slug, plan, owner_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		org.ID,
		org.Name,
		org.Slug,
		org.Plan,
		org.OwnerID,
		org.CreatedAt,
		org.UpdatedAt,
	)

	return err
}

func (s *Service) getOrgByID(ctx context.Context, id string) (*Org, error) {
	orgID, ok := db.OrgIDFromContext(ctx)
	if !ok || orgID != id {
		return nil, ErrPermissionDenied
	}

	query := `
		SELECT id, name, slug, plan, owner_id, created_at, updated_at, deleted_at
		FROM orgs
		WHERE id = ? AND deleted_at IS NULL
	`

	org := &Org{}
	var deletedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.Plan,
		&org.OwnerID,
		&org.CreatedAt,
		&org.UpdatedAt,
		&deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrOrgNotFound
	}

	if err != nil {
		return nil, err
	}

	if deletedAt.Valid {
		org.DeletedAt = &deletedAt.Time
	}

	return org, nil
}

func (s *Service) getOrgBySlug(ctx context.Context, slug string) (*Org, error) {
	orgID, ok := db.OrgIDFromContext(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}

	query := `
		SELECT id, name, slug, plan, owner_id, created_at, updated_at, deleted_at
		FROM orgs
		WHERE slug = ? AND deleted_at IS NULL
	`

	org := &Org{}
	var deletedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, query, slug).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.Plan,
		&org.OwnerID,
		&org.CreatedAt,
		&org.UpdatedAt,
		&deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrOrgNotFound
	}

	if err != nil {
		return nil, err
	}

	if org.ID != orgID {
		return nil, ErrPermissionDenied
	}

	if deletedAt.Valid {
		org.DeletedAt = &deletedAt.Time
	}

	return org, nil
}

func (s *Service) listOrgsByUser(ctx context.Context, userID string) ([]*Org, error) {
	query := `
		SELECT o.id, o.name, o.slug, o.plan, o.owner_id, o.created_at, o.updated_at
		FROM orgs o
		INNER JOIN org_members om ON o.id = om.org_id
		WHERE om.user_id = ? AND o.deleted_at IS NULL AND om.left_at IS NULL
		ORDER BY o.created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*Org
	for rows.Next() {
		org := &Org{}
		err := rows.Scan(
			&org.ID,
			&org.Name,
			&org.Slug,
			&org.Plan,
			&org.OwnerID,
			&org.CreatedAt,
			&org.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orgs, nil
}

func (s *Service) updateOrg(ctx context.Context, org *Org) error {
	orgID, ok := db.OrgIDFromContext(ctx)
	if !ok || orgID != org.ID {
		return ErrPermissionDenied
	}

	org.UpdatedAt = time.Now()

	query := `
		UPDATE orgs
		SET name = ?, slug = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query,
		org.Name,
		org.Slug,
		org.UpdatedAt,
		org.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrOrgNotFound
	}

	return nil
}

func (s *Service) softDeleteOrg(ctx context.Context, id string) error {
	orgID, ok := db.OrgIDFromContext(ctx)
	if !ok || orgID != id {
		return ErrPermissionDenied
	}

	now := time.Now()

	query := `
		UPDATE orgs
		SET deleted_at = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrOrgNotFound
	}

	return nil
}

func (s *Service) restoreOrg(ctx context.Context, id string) error {
	orgID, ok := db.OrgIDFromContext(ctx)
	if !ok || orgID != id {
		return ErrPermissionDenied
	}

	now := time.Now()

	query := `
		UPDATE orgs
		SET deleted_at = NULL, updated_at = ?
		WHERE id = ? AND deleted_at IS NOT NULL
	`

	result, err := s.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrOrgNotFound
	}

	return nil
}

func (s *Service) hardDeleteOrg(ctx context.Context, id string) error {
	orgID, ok := db.OrgIDFromContext(ctx)
	if !ok || orgID != id {
		return ErrPermissionDenied
	}

	query := `
		DELETE FROM orgs
		WHERE id = ? AND deleted_at IS NOT NULL
		AND deleted_at <= datetime('now', '-30 days')
	`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrOrgNotFound
	}

	return nil
}
