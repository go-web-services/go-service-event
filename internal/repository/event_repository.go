package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Lomank123/go-service-event/internal/domain"
	platformError "github.com/go-web-services/go-web-platform/error"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository interface {
	Create(ctx context.Context, event *domain.Event) error
	CreateBatch(ctx context.Context, events []*domain.Event) error
	Update(ctx context.Context, event *domain.Event) error
	GetByID(ctx context.Context, id string) (*domain.Event, error)
	Delete(ctx context.Context, id string) (*domain.Event, error)
	Find(ctx context.Context, filters map[string]any, sort []string, skip int64, limit int64) ([]domain.Event, error)
	Count(ctx context.Context, filters map[string]any) (int64, error)
}

type eventRepository struct {
	db *pgxpool.Pool
}

func NewEventRepository(db *pgxpool.Pool) EventRepository {
	return &eventRepository{db: db}
}

func marshalPayload(p map[string]any) ([]byte, error) {
	if p == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(p)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanEvent(row rowScanner) (*domain.Event, error) {
	var ev domain.Event
	var payloadBytes []byte
	err := row.Scan(
		&ev.ID,
		&ev.ProjectID,
		&ev.MessageID,
		&ev.DistinctID,
		&ev.UserID,
		&ev.SessionID,
		&ev.IP,
		&ev.UserAgent,
		&ev.Name,
		&payloadBytes,
		&ev.OccurredAt,
		&ev.ReceivedAt,
		&ev.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	if len(payloadBytes) > 0 {
		if err := json.Unmarshal(payloadBytes, &ev.Payload); err != nil {
			return nil, fmt.Errorf("unmarshal payload: %w", err)
		}
	}
	if ev.Payload == nil {
		ev.Payload = map[string]any{}
	}
	return &ev, nil
}

// pgxQuerier matches *pgxpool.Pool and pgx.Tx for single-row operations used in Create / CreateBatch.
type pgxQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (r *eventRepository) createOne(ctx context.Context, q pgxQuerier, event *domain.Event) error {
	payloadBytes, err := marshalPayload(event.Payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	const insert = `
INSERT INTO events (
  project_id, message_id, distinct_id, user_id, session_id,
  ip, user_agent, name, payload, occurred_at, received_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10, $11)
ON CONFLICT (project_id, message_id) DO NOTHING
RETURNING id`

	err = q.QueryRow(ctx, insert,
		event.ProjectID,
		event.MessageID,
		event.DistinctID,
		event.UserID,
		event.SessionID,
		event.IP,
		event.UserAgent,
		event.Name,
		payloadBytes,
		event.OccurredAt,
		event.ReceivedAt,
	).Scan(&event.ID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("failed to create event: %w", err)
	}

	const sel = `
SELECT id, project_id, message_id, distinct_id, user_id, session_id, ip, user_agent, name, payload, occurred_at, received_at, deleted_at
FROM events WHERE project_id = $1 AND message_id = $2 AND deleted_at IS NULL`

	ev, err := scanEvent(q.QueryRow(ctx, sel, event.ProjectID, event.MessageID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("idempotent insert conflict but row not found: %w", err)
		}
		return fmt.Errorf("failed to load event after conflict: %w", err)
	}
	*event = *ev
	return nil
}

func (r *eventRepository) Create(ctx context.Context, event *domain.Event) error {
	return r.createOne(ctx, r.db, event)
}

func (r *eventRepository) CreateBatch(ctx context.Context, events []*domain.Event) error {
	if len(events) == 0 {
		return nil
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin batch tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, ev := range events {
		if ev == nil {
			return fmt.Errorf("nil event in batch")
		}
		if err := r.createOne(ctx, tx, ev); err != nil {
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit batch tx: %w", err)
	}
	return nil
}

func (r *eventRepository) Update(ctx context.Context, event *domain.Event) error {
	payloadBytes, err := marshalPayload(event.Payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	const q = `UPDATE events
SET name = $2, payload = $3::jsonb, user_id = $4, session_id = $5
WHERE id = $1 AND deleted_at IS NULL`

	cmdTag, err := r.db.Exec(ctx, q,
		event.ID,
		event.Name,
		payloadBytes,
		event.UserID,
		event.SessionID,
	)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return platformError.ErrEntityNotFound
	}

	return nil
}

func (r *eventRepository) GetByID(ctx context.Context, id string) (*domain.Event, error) {
	const q = `SELECT id, project_id, message_id, distinct_id, user_id, session_id, ip, user_agent, name, payload, occurred_at, received_at, deleted_at
FROM events WHERE id = $1 AND deleted_at IS NULL`

	ev, err := scanEvent(r.db.QueryRow(ctx, q, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, platformError.ErrEntityNotFound
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return ev, nil
}

func (r *eventRepository) Delete(ctx context.Context, id string) (*domain.Event, error) {
	const q = `UPDATE events SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, project_id, message_id, distinct_id, user_id, session_id, ip, user_agent, name, payload, occurred_at, received_at, deleted_at`

	ev, err := scanEvent(r.db.QueryRow(ctx, q, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, platformError.ErrEntityNotFound
		}
		return nil, fmt.Errorf("failed to soft delete event: %w", err)
	}

	return ev, nil
}

func (r *eventRepository) Find(
	ctx context.Context,
	filters map[string]any,
	sort []string,
	skip int64,
	limit int64,
) ([]domain.Event, error) {
	query := `SELECT id, project_id, message_id, distinct_id, user_id, session_id, ip, user_agent, name, payload, occurred_at, received_at, deleted_at
FROM events WHERE deleted_at IS NULL`

	whereClauses := []string{}
	args := []any{}
	argCount := 1

	for k, v := range filters {
		switch k {
		case "id", "project_id", "distinct_id", "message_id", "user_id", "session_id", "ip", "user_agent":
			if ids, ok := v.([]string); ok && len(ids) > 0 {
				placeholders := make([]string, len(ids))
				for i, id := range ids {
					placeholders[i] = fmt.Sprintf("$%d", argCount)
					args = append(args, id)
					argCount++
				}
				whereClauses = append(whereClauses, fmt.Sprintf("%s IN (%s)", k, strings.Join(placeholders, ",")))
			}
		case "name":
			if names, ok := v.([]string); ok && len(names) > 0 {
				ors := make([]string, 0, len(names))
				for _, n := range names {
					if strings.Contains(n, "%") {
						ors = append(ors, fmt.Sprintf("name ILIKE $%d", argCount))
					} else {
						ors = append(ors, fmt.Sprintf("name = $%d", argCount))
					}
					args = append(args, n)
					argCount++
				}
				whereClauses = append(whereClauses, "("+strings.Join(ors, " OR ")+")")
			}
		default:
			whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", k, argCount))
			args = append(args, v)
			argCount++
		}
	}

	if len(whereClauses) > 0 {
		query += " AND " + strings.Join(whereClauses, " AND ")
	}

	if len(sort) > 0 {
		orderByClauses := []string{}
		for _, s := range sort {
			direction := "ASC"
			field := s
			if strings.HasPrefix(s, "-") {
				direction = "DESC"
				field = strings.TrimPrefix(s, "-")
			}
			orderByClauses = append(orderByClauses, fmt.Sprintf("%s %s", field, direction))
		}
		query += " ORDER BY " + strings.Join(orderByClauses, ", ")
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, skip)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		ev, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, *ev)
	}

	return events, nil
}

func (r *eventRepository) Count(ctx context.Context, filters map[string]any) (int64, error) {
	query := "SELECT COUNT(*) FROM events WHERE deleted_at IS NULL"

	whereClauses := []string{}
	args := []any{}
	argCount := 1

	for k, v := range filters {
		switch k {
		case "id", "project_id", "distinct_id", "message_id", "user_id", "session_id", "ip", "user_agent":
			if ids, ok := v.([]string); ok && len(ids) > 0 {
				placeholders := make([]string, len(ids))
				for i, id := range ids {
					placeholders[i] = fmt.Sprintf("$%d", argCount)
					args = append(args, id)
					argCount++
				}
				whereClauses = append(whereClauses, fmt.Sprintf("%s IN (%s)", k, strings.Join(placeholders, ",")))
			}
		case "name":
			if names, ok := v.([]string); ok && len(names) > 0 {
				ors := make([]string, 0, len(names))
				for _, n := range names {
					if strings.Contains(n, "%") {
						ors = append(ors, fmt.Sprintf("name ILIKE $%d", argCount))
					} else {
						ors = append(ors, fmt.Sprintf("name = $%d", argCount))
					}
					args = append(args, n)
					argCount++
				}
				whereClauses = append(whereClauses, "("+strings.Join(ors, " OR ")+")")
			}
		default:
			whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", k, argCount))
			args = append(args, v)
			argCount++
		}
	}

	if len(whereClauses) > 0 {
		query += " AND " + strings.Join(whereClauses, " AND ")
	}

	var count int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
