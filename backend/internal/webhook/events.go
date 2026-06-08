package webhook

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func (s *Service) DispatchEvent(ctx context.Context, orgID, eventType string, payload interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	id := uuid.New().String()

	query := `INSERT INTO webhook_events (id, org_id, event_type, payload, status, attempts) VALUES (?, ?, ?, ?, ?, 0)`

	_, err = s.db.ExecContext(ctx, query, id, orgID, eventType, string(payloadJSON), StatusPending)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ProcessEvents(ctx context.Context) error {
	query := `SELECT id, org_id, event_type, payload, status, attempts, last_attempt_at, created_at FROM webhook_events WHERE status = ? ORDER BY created_at ASC LIMIT 100`

	rows, err := s.db.QueryContext(ctx, query, StatusPending)
	if err != nil {
		return err
	}
	defer rows.Close()

	events := make([]*WebhookEvent, 0)
	for rows.Next() {
		event := &WebhookEvent{}
		var lastAttemptAt sql.NullTime
		err := rows.Scan(&event.ID, &event.OrgID, &event.EventType, &event.Payload, &event.Status, &event.Attempts, &lastAttemptAt, &event.CreatedAt)
		if err != nil {
			return err
		}
		if lastAttemptAt.Valid {
			event.LastAttemptAt = &lastAttemptAt.Time
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	for _, event := range events {
		go s.processEvent(ctx, event)
	}

	return nil
}

func (s *Service) processEvent(ctx context.Context, event *WebhookEvent) {
	endpoints, err := s.ListActiveEndpoints(ctx, event.OrgID)
	if err != nil {
		return
	}

	for _, endpoint := range endpoints {
		go func(ep *WebhookEndpoint) {
			err := s.SendEvent(event, ep)
			now := time.Now()

			if err != nil {
				newAttempts := event.Attempts + 1
				if newAttempts >= MaxRetries {
					s.updateEventStatus(ctx, event.ID, StatusDead, newAttempts, now)
				} else {
					s.updateEventStatus(ctx, event.ID, StatusFailed, newAttempts, now)
				}
			} else {
				s.updateEventStatus(ctx, event.ID, StatusSuccess, event.Attempts+1, now)
			}
		}(endpoint)
	}
}

func (s *Service) SendEvent(event *WebhookEvent, endpoint *WebhookEndpoint) error {
	signature := GenerateSignature(event.Payload, endpoint.Secret)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	req, err := http.NewRequest("POST", endpoint.URL, bytes.NewBufferString(event.Payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("X-Webhook-Event", event.EventType)
	req.Header.Set("X-Webhook-Timestamp", timestamp)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook delivery failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (s *Service) RetryFailedEvents(ctx context.Context) error {
	query := `SELECT id, org_id, event_type, payload, status, attempts, last_attempt_at, created_at FROM webhook_events WHERE status = ? AND attempts < ? ORDER BY last_attempt_at ASC LIMIT 100`

	rows, err := s.db.QueryContext(ctx, query, StatusFailed, MaxRetries)
	if err != nil {
		return err
	}
	defer rows.Close()

	events := make([]*WebhookEvent, 0)
	for rows.Next() {
		event := &WebhookEvent{}
		var lastAttemptAt sql.NullTime
		err := rows.Scan(&event.ID, &event.OrgID, &event.EventType, &event.Payload, &event.Status, &event.Attempts, &lastAttemptAt, &event.CreatedAt)
		if err != nil {
			return err
		}
		if lastAttemptAt.Valid {
			event.LastAttemptAt = &lastAttemptAt.Time
		}

		if event.LastAttemptAt != nil {
			delay := s.getRetryDelay(event.Attempts)
			if time.Since(*event.LastAttemptAt) < delay {
				continue
			}
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	for _, event := range events {
		go s.retryEvent(ctx, event)
	}

	return nil
}

func (s *Service) retryEvent(ctx context.Context, event *WebhookEvent) {
	endpoints, err := s.ListActiveEndpoints(ctx, event.OrgID)
	if err != nil {
		return
	}

	for _, endpoint := range endpoints {
		go func(ep *WebhookEndpoint) {
			err := s.SendEvent(event, ep)
			now := time.Now()

			if err != nil {
				newAttempts := event.Attempts + 1
				if newAttempts >= MaxRetries {
					s.updateEventStatus(ctx, event.ID, StatusDead, newAttempts, now)
				} else {
					s.updateEventStatus(ctx, event.ID, StatusFailed, newAttempts, now)
				}
			} else {
				s.updateEventStatus(ctx, event.ID, StatusSuccess, event.Attempts+1, now)
			}
		}(endpoint)
	}
}

func (s *Service) getRetryDelay(attempts int) time.Duration {
	switch attempts {
	case 1:
		return RetryDelay1
	case 2:
		return RetryDelay2
	default:
		return RetryDelay4
	}
}

func (s *Service) updateEventStatus(ctx context.Context, eventID, status string, attempts int, lastAttemptAt time.Time) {
	query := `UPDATE webhook_events SET status = ?, attempts = ?, last_attempt_at = ? WHERE id = ?`
	s.db.ExecContext(ctx, query, status, attempts, lastAttemptAt, eventID)
}

func (s *Service) GetEvent(ctx context.Context, id string) (*WebhookEvent, error) {
	query := `SELECT id, org_id, event_type, payload, status, attempts, last_attempt_at, created_at FROM webhook_events WHERE id = ?`

	event := &WebhookEvent{}
	var lastAttemptAt sql.NullTime

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID, &event.OrgID, &event.EventType, &event.Payload, &event.Status, &event.Attempts, &lastAttemptAt, &event.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("webhook event not found")
		}
		return nil, err
	}

	if lastAttemptAt.Valid {
		event.LastAttemptAt = &lastAttemptAt.Time
	}

	return event, nil
}

func (s *Service) ListEvents(ctx context.Context, orgID string, limit, offset int) ([]*WebhookEvent, int, error) {
	if limit <= 0 {
		limit = 20
	}

	countQuery := `SELECT COUNT(*) FROM webhook_events WHERE org_id = ?`
	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, org_id, event_type, payload, status, attempts, last_attempt_at, created_at FROM webhook_events WHERE org_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := s.db.QueryContext(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	events := make([]*WebhookEvent, 0)
	for rows.Next() {
		event := &WebhookEvent{}
		var lastAttemptAt sql.NullTime
		err := rows.Scan(&event.ID, &event.OrgID, &event.EventType, &event.Payload, &event.Status, &event.Attempts, &lastAttemptAt, &event.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		if lastAttemptAt.Valid {
			event.LastAttemptAt = &lastAttemptAt.Time
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return events, total, nil
}
