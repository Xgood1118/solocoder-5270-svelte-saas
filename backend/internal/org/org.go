package org

import (
	"errors"
	"time"

	"saas-system/internal/db"
)

var (
	ErrOrgNotFound       = errors.New("org not found")
	ErrOrgExists         = errors.New("org already exists")
	ErrMemberNotFound    = errors.New("member not found")
	ErrMemberExists      = errors.New("member already exists")
	ErrInvitationNotFound = errors.New("invitation not found")
	ErrInvitationExpired  = errors.New("invitation expired")
	ErrInvitationAccepted = errors.New("invitation already accepted")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrInvalidRole        = errors.New("invalid role")
)

const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
	RoleGuest  = "guest"
)

const (
	PlanFree = "free"
)

type Org struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Slug      string     `json:"slug"`
	Plan      string     `json:"plan"`
	OwnerID   string     `json:"owner_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type OrgMember struct {
	OrgID    string     `json:"org_id"`
	UserID   string     `json:"user_id"`
	Role     string     `json:"role"`
	JoinedAt time.Time  `json:"joined_at"`
	LeftAt   *time.Time `json:"left_at,omitempty"`
}

type Invitation struct {
	ID         string     `json:"id"`
	OrgID      string     `json:"org_id"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	Token      string     `json:"-"`
	ExpiresAt  time.Time  `json:"expires_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	InvitedBy  string     `json:"invited_by"`
	CreatedAt  time.Time  `json:"created_at"`
}

type Service struct {
	db *db.DB
}

func NewService(database *db.DB) *Service {
	return &Service{
		db: database,
	}
}

func IsValidRole(role string) bool {
	switch role {
	case RoleOwner, RoleAdmin, RoleMember, RoleGuest:
		return true
	default:
		return false
	}
}

func roleHierarchy(role string) int {
	switch role {
	case RoleOwner:
		return 4
	case RoleAdmin:
		return 3
	case RoleMember:
		return 2
	case RoleGuest:
		return 1
	default:
		return 0
	}
}

func CanModifyRole(actorRole, targetRole, newRole string) bool {
	actorLevel := roleHierarchy(actorRole)
	targetLevel := roleHierarchy(targetRole)
	newLevel := roleHierarchy(newRole)

	if actorLevel <= targetLevel {
		return false
	}

	if newLevel >= actorLevel {
		return false
	}

	return true
}
