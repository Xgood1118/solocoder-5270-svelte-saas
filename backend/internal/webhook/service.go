package webhook

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEndpointNotFound = errors.New("webhook endpoint not found")
)

type Service struct {
	db     *sql.DB
	secret string
}

func NewService(db *sql.DB, secret string) *Service {
	return &Service{
		db:     db,
		secret: secret,
	}
}

func (s *Service) ListEndpoints(ctx context.Context, orgID string) ([]WebhookEndpoint, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT org_id, id, url, secret, active, created_at FROM webhook_endpoints WHERE org_id = ?",
		orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var endpoints []WebhookEndpoint
	for rows.Next() {
		var ep WebhookEndpoint
		var active int
		err := rows.Scan(&ep.OrgID, &ep.ID, &ep.URL, &ep.Secret, &active, &ep.CreatedAt)
		if err != nil {
			return nil, err
		}
		ep.Active = active == 1
		endpoints = append(endpoints, ep)
	}
	return endpoints, nil
}

func (s *Service) GetEndpointByID(ctx context.Context, orgID, endpointID string) (*WebhookEndpoint, error) {
	var ep WebhookEndpoint
	var active int
	err := s.db.QueryRowContext(ctx,
		"SELECT org_id, id, url, secret, active, created_at FROM webhook_endpoints WHERE org_id = ? AND id = ?",
		orgID, endpointID,
	).Scan(&ep.OrgID, &ep.ID, &ep.URL, &ep.Secret, &active, &ep.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrEndpointNotFound
	}
	if err != nil {
		return nil, err
	}
	ep.Active = active == 1
	return &ep, nil
}

func (s *Service) CreateEndpoint(ctx context.Context, orgID, url string, active bool) (*WebhookEndpoint, error) {
	ep := &WebhookEndpoint{
		OrgID:     orgID,
		ID:        uuid.New().String(),
		URL:       url,
		Secret:    uuid.New().String(),
		Active:    active,
		CreatedAt: time.Now(),
	}

	activeInt := 0
	if active {
		activeInt = 1
	}

	_, err := s.db.ExecContext(ctx,
		"INSERT INTO webhook_endpoints (org_id, id, url, secret, active, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		ep.OrgID, ep.ID, ep.URL, ep.Secret, activeInt, ep.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return ep, nil
}

func (s *Service) UpdateEndpoint(ctx context.Context, orgID, endpointID, url string, active bool) (*WebhookEndpoint, error) {
	_, err := s.GetEndpointByID(ctx, orgID, endpointID)
	if err != nil {
		return nil, err
	}

	activeInt := 0
	if active {
		activeInt = 1
	}

	_, err = s.db.ExecContext(ctx,
		"UPDATE webhook_endpoints SET url = ?, active = ? WHERE org_id = ? AND id = ?",
		url, activeInt, orgID, endpointID,
	)
	if err != nil {
		return nil, err
	}

	return s.GetEndpointByID(ctx, orgID, endpointID)
}

func (s *Service) DeleteEndpoint(ctx context.Context, orgID, endpointID string) error {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM webhook_endpoints WHERE org_id = ? AND id = ?",
		orgID, endpointID,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrEndpointNotFound
	}
	return nil
}

func (s *Service) ListActiveEndpoints(ctx context.Context, orgID string) ([]*WebhookEndpoint, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT org_id, id, url, secret, active, created_at FROM webhook_endpoints WHERE org_id = ? AND active = 1 ORDER BY created_at DESC",
		orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var endpoints []*WebhookEndpoint
	for rows.Next() {
		ep := &WebhookEndpoint{}
		var active int
		err := rows.Scan(&ep.OrgID, &ep.ID, &ep.URL, &ep.Secret, &active, &ep.CreatedAt)
		if err != nil {
			return nil, err
		}
		ep.Active = active == 1
		endpoints = append(endpoints, ep)
	}
	return endpoints, nil
}

func (s *Service) CreateEvent(ctx context.Context, orgID, eventType, payload string) (*WebhookEvent, error) {
	event := &WebhookEvent{
		ID:        uuid.New().String(),
		OrgID:     orgID,
		EventType: eventType,
		Payload:   payload,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	_, err := s.db.ExecContext(ctx,
		"INSERT INTO webhook_events (id, org_id, event_type, payload, status, attempts, created_at) VALUES (?, ?, ?, ?, ?, 0, ?)",
		event.ID, event.OrgID, event.EventType, event.Payload, event.Status, event.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return event, nil
}
