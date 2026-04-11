package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Lomank123/go-service-event/internal/domain"
	platformError "github.com/Lomank123/go-web-platform/error"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository interface {
	Create(ctx context.Context, event *domain.Event) error
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

func (r *eventRepository) Create(ctx context.Context, event *domain.Event) error {
	query := `INSERT INTO events (name, slug, description, event_type, payload, status, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	err := r.db.QueryRow(ctx, query,
		event.Name,
		event.Slug,
		event.Description,
		event.Type,
		event.Payload,
		event.Status,
		event.CreatedAt,
		event.UpdatedAt,
	).Scan(&event.ID)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

func (r *eventRepository) Update(ctx context.Context, event *domain.Event) error {
	query := `UPDATE events
			  SET name = $1, slug = $2, description = $3, event_type = $4, payload = $5, status = $6, updated_at = NOW()
			  WHERE id = $7 AND deleted_at IS NULL`

	cmdTag, err := r.db.Exec(ctx, query,
		event.Name,
		event.Slug,
		event.Description,
		event.Type,
		event.Payload,
		event.Status,
		event.ID,
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
	query := `SELECT id, name, slug, description, event_type, payload, status, created_at, updated_at, deleted_at
			  FROM events WHERE id = $1 AND deleted_at IS NULL`

	var ev domain.Event
	err := r.db.QueryRow(ctx, query, id).Scan(
		&ev.ID,
		&ev.Name,
		&ev.Slug,
		&ev.Description,
		&ev.Type,
		&ev.Payload,
		&ev.Status,
		&ev.CreatedAt,
		&ev.UpdatedAt,
		&ev.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, platformError.ErrEntityNotFound
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return &ev, nil
}

func (r *eventRepository) Delete(ctx context.Context, id string) (*domain.Event, error) {
	query := `UPDATE events SET deleted_at = NOW()
			  WHERE id = $1 AND deleted_at IS NULL
			  RETURNING id, name, slug, description, event_type, payload, status, created_at, updated_at, deleted_at`

	var ev domain.Event
	err := r.db.QueryRow(ctx, query, id).Scan(
		&ev.ID,
		&ev.Name,
		&ev.Slug,
		&ev.Description,
		&ev.Type,
		&ev.Payload,
		&ev.Status,
		&ev.CreatedAt,
		&ev.UpdatedAt,
		&ev.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, platformError.ErrEntityNotFound
		}
		return nil, fmt.Errorf("failed to soft delete event: %w", err)
	}

	return &ev, nil
}

func (r *eventRepository) Find(
	ctx context.Context,
	filters map[string]any,
	sort []string,
	skip int64,
	limit int64,
) ([]domain.Event, error) {
	var events []domain.Event
	query := "SELECT id, name, slug, description, event_type, payload, status, created_at, updated_at, deleted_at FROM events WHERE deleted_at IS NULL"

	whereClauses := []string{}
	args := []any{}
	argCount := 1

	for k, v := range filters {
		if k == "id" {
			if ids, ok := v.([]string); ok && len(ids) > 0 {
				placeholders := make([]string, len(ids))
				for i, id := range ids {
					placeholders[i] = fmt.Sprintf("$%d", argCount)
					args = append(args, id)
					argCount++
				}
				whereClauses = append(whereClauses, fmt.Sprintf("id IN (%s)", strings.Join(placeholders, ",")))
			}
		} else if k == "name" {
			if strV, ok := v.(string); ok && strings.Contains(strV, "%") {
				whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", argCount))
				args = append(args, v)
				argCount++
			} else {
				whereClauses = append(whereClauses, fmt.Sprintf("name = $%d", argCount))
				args = append(args, v)
				argCount++
			}
		} else {
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

	for rows.Next() {
		var ev domain.Event
		if err := rows.Scan(
			&ev.ID,
			&ev.Name,
			&ev.Slug,
			&ev.Description,
			&ev.Type,
			&ev.Payload,
			&ev.Status,
			&ev.CreatedAt,
			&ev.UpdatedAt,
			&ev.DeletedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, ev)
	}

	return events, nil
}

func (r *eventRepository) Count(ctx context.Context, filters map[string]any) (int64, error) {
	query := "SELECT COUNT(*) FROM events WHERE deleted_at IS NULL"

	whereClauses := []string{}
	args := []any{}
	argCount := 1

	for k, v := range filters {
		if k == "id" {
			if ids, ok := v.([]string); ok && len(ids) > 0 {
				placeholders := make([]string, len(ids))
				for i, id := range ids {
					placeholders[i] = fmt.Sprintf("$%d", argCount)
					args = append(args, id)
					argCount++
				}
				whereClauses = append(whereClauses, fmt.Sprintf("id IN (%s)", strings.Join(placeholders, ",")))
			}
		} else if k == "name" {
			if strV, ok := v.(string); ok && strings.Contains(strV, "%") {
				whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", argCount))
				args = append(args, v)
				argCount++
			} else {
				whereClauses = append(whereClauses, fmt.Sprintf("name = $%d", argCount))
				args = append(args, v)
				argCount++
			}
		} else {
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
