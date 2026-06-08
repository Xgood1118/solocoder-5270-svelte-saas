package org

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/google/uuid"

	"saas-system/internal/db"
)

const (
	invitationTokenLength = 32
	invitationExpiryHours = 72
)

func (s *Service) CreateOrg(userID, name, slug string) (*Org, error) {
	org := &Org{
		ID:      uuid.New().String(),
		Name:    name,
		Slug:    slug,
		Plan:    PlanFree,
		OwnerID: userID,
	}

	ctx := context.Background()

	err := s.db.WithTx(ctx, func(tx *sql.Tx) error {
		now := time.Now()
		org.CreatedAt = now
		org.UpdatedAt = now

		orgQuery := `
			INSERT INTO orgs (id, name, slug, plan, owner_id, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`
		_, err := tx.ExecContext(ctx, orgQuery,
			org.ID,
			org.Name,
			org.Slug,
			org.Plan,
			org.OwnerID,
			org.CreatedAt,
			org.UpdatedAt,
		)
		if err != nil {
			return err
		}

		memberQuery := `
			INSERT INTO org_members (org_id, user_id, role, joined_at)
			VALUES (?, ?, ?, ?)
		`
		_, err = tx.ExecContext(ctx, memberQuery,
			org.ID,
			userID,
			RoleOwner,
			now,
		)
		if err != nil {
			return err
		}

		quotaQuery := `
			INSERT INTO quotas (org_id, id, metric, "limit", used, period_start, period_end)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`
		now = time.Now()
		periodEnd := now.AddDate(0, 1, 0)

		quotas := []struct {
			metric string
			limit  int
		}{
			{"members", 5},
			{"projects", 3},
			{"storage_gb", 1},
		}

		for _, q := range quotas {
			quotaID := uuid.New().String()
			_, err = tx.ExecContext(ctx, quotaQuery,
				org.ID,
				quotaID,
				q.metric,
				q.limit,
				0,
				now,
				periodEnd,
			)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return org, nil
}

func (s *Service) ListUserOrgs(userID string) ([]*Org, error) {
	ctx := context.Background()
	return s.listOrgsByUser(ctx, userID)
}

func (s *Service) GetOrg(orgID string) (*Org, error) {
	ctx := db.WithOrgID(context.Background(), orgID)
	return s.getOrgByID(ctx, orgID)
}

func (s *Service) UpdateOrg(orgID, name string) (*Org, error) {
	ctx := db.WithOrgID(context.Background(), orgID)

	org, err := s.getOrgByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	org.Name = name
	org.UpdatedAt = time.Now()

	err = s.updateOrg(ctx, org)
	if err != nil {
		return nil, err
	}

	return org, nil
}

func (s *Service) DeleteOrg(orgID string) error {
	ctx := db.WithOrgID(context.Background(), orgID)
	return s.softDeleteOrg(ctx, orgID)
}

func (s *Service) InviteMember(orgID, inviterID, email, role string) (*Invitation, error) {
	if !IsValidRole(role) {
		return nil, ErrInvalidRole
	}

	if role == RoleOwner {
		return nil, ErrPermissionDenied
	}

	ctx := db.WithOrgID(context.Background(), orgID)

	inviterRole, err := s.getMemberRole(ctx, orgID, inviterID)
	if err != nil {
		return nil, err
	}

	if !CanModifyRole(inviterRole, RoleGuest, role) {
		return nil, ErrPermissionDenied
	}

	token, err := generateInvitationToken()
	if err != nil {
		return nil, err
	}

	invitation := &Invitation{
		ID:        uuid.New().String(),
		OrgID:     orgID,
		Email:     email,
		Role:      role,
		Token:     token,
		ExpiresAt: time.Now().Add(invitationExpiryHours * time.Hour),
		InvitedBy: inviterID,
	}

	err = s.createInvitation(ctx, invitation)
	if err != nil {
		return nil, err
	}

	return invitation, nil
}

func (s *Service) AcceptInvitation(token, userID string) (*OrgMember, error) {
	ctx := context.Background()

	invitation, err := s.getInvitationByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if invitation.AcceptedAt != nil {
		return nil, ErrInvitationAccepted
	}

	if time.Now().After(invitation.ExpiresAt) {
		return nil, ErrInvitationExpired
	}

	orgCtx := db.WithOrgID(ctx, invitation.OrgID)

	err = s.db.WithTx(orgCtx, func(tx *sql.Tx) error {
		now := time.Now()

		acceptQuery := `
			UPDATE invitations
			SET accepted_at = ?
			WHERE id = ? AND accepted_at IS NULL
		`
		_, err := tx.ExecContext(orgCtx, acceptQuery, now, invitation.ID)
		if err != nil {
			return err
		}

		memberQuery := `
			INSERT INTO org_members (org_id, user_id, role, joined_at)
			VALUES (?, ?, ?, ?)
		`
		_, err = tx.ExecContext(orgCtx, memberQuery,
			invitation.OrgID,
			userID,
			invitation.Role,
			now,
		)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	member := &OrgMember{
		OrgID:    invitation.OrgID,
		UserID:   userID,
		Role:     invitation.Role,
		JoinedAt: time.Now(),
	}

	return member, nil
}

func (s *Service) ListMembers(orgID string) ([]*OrgMember, error) {
	ctx := db.WithOrgID(context.Background(), orgID)
	return s.listMembers(ctx, orgID)
}

func (s *Service) UpdateMemberRole(orgID, userID, newRole, actorID string) error {
	if !IsValidRole(newRole) {
		return ErrInvalidRole
	}

	ctx := db.WithOrgID(context.Background(), orgID)

	actorRole, err := s.getMemberRole(ctx, orgID, actorID)
	if err != nil {
		return err
	}

	targetMember, err := s.getMember(ctx, orgID, userID)
	if err != nil {
		return err
	}

	if !CanModifyRole(actorRole, targetMember.Role, newRole) {
		return ErrPermissionDenied
	}

	beforeRole := targetMember.Role

	err = s.updateMemberRole(ctx, orgID, userID, newRole)
	if err != nil {
		return err
	}

	s.recordRoleChangeAudit(ctx, orgID, userID, beforeRole, newRole, actorID)

	return nil
}

func (s *Service) RemoveMember(orgID, userID, actorID string) error {
	ctx := db.WithOrgID(context.Background(), orgID)

	actorRole, err := s.getMemberRole(ctx, orgID, actorID)
	if err != nil {
		return err
	}

	targetMember, err := s.getMember(ctx, orgID, userID)
	if err != nil {
		return err
	}

	actorLevel := roleHierarchy(actorRole)
	targetLevel := roleHierarchy(targetMember.Role)

	if actorLevel <= targetLevel {
		return ErrPermissionDenied
	}

	err = s.removeMember(ctx, orgID, userID)
	if err != nil {
		return err
	}

	s.recordMemberRemoveAudit(ctx, orgID, userID, targetMember.Role, actorID)

	return nil
}

func (s *Service) ListInvitations(orgID string) ([]*Invitation, error) {
	ctx := db.WithOrgID(context.Background(), orgID)
	return s.listInvitations(ctx, orgID)
}

func (s *Service) CancelInvitation(orgID, invitationID, actorID string) error {
	ctx := db.WithOrgID(context.Background(), orgID)

	actorRole, err := s.getMemberRole(ctx, orgID, actorID)
	if err != nil {
		return err
	}

	actorLevel := roleHierarchy(actorRole)
	if actorLevel < roleHierarchy(RoleAdmin) {
		return ErrPermissionDenied
	}

	return s.deleteInvitation(ctx, invitationID)
}

func (s *Service) GetMemberRole(orgID, userID string) (string, error) {
	ctx := db.WithOrgID(context.Background(), orgID)
	return s.getMemberRole(ctx, orgID, userID)
}

func generateInvitationToken() (string, error) {
	bytes := make([]byte, invitationTokenLength)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *Service) recordRoleChangeAudit(ctx context.Context, orgID, userID, beforeRole, afterRole, actorID string) {
	auditID := uuid.New().String()
	beforeData := beforeRole
	afterData := afterRole

	query := `
		INSERT INTO audit_logs (org_id, id, user_id, action, entity_type, entity_id, before_data, after_data, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	s.db.ExecContext(ctx, query,
		orgID,
		auditID,
		actorID,
		"member.role_updated",
		"member",
		userID,
		&beforeData,
		&afterData,
		now,
	)
}

func (s *Service) recordMemberRemoveAudit(ctx context.Context, orgID, userID, role, actorID string) {
	auditID := uuid.New().String()
	beforeData := role

	query := `
		INSERT INTO audit_logs (org_id, id, user_id, action, entity_type, entity_id, before_data, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	s.db.ExecContext(ctx, query,
		orgID,
		auditID,
		actorID,
		"member.removed",
		"member",
		userID,
		&beforeData,
		now,
	)
}
