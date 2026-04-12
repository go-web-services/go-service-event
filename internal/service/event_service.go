package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Lomank123/go-service-event/internal/domain"
	"github.com/Lomank123/go-service-event/internal/repository"
	"github.com/Lomank123/go-service-event/internal/types"
	platformTypes "github.com/Lomank123/go-web-platform/types"
)

// ErrDuplicateBatchKeys is returned when two items in a batch share the same project_id + message_id.
var ErrDuplicateBatchKeys = errors.New("duplicate project_id+message_id in batch")

type EventService interface {
	Detail(ctx context.Context, id string) (*domain.Event, error)
	Create(ctx context.Context, ev *domain.Event) (*domain.Event, error)
	CreateBatch(ctx context.Context, events []*domain.Event) ([]domain.Event, error)
	Update(
		ctx context.Context,
		id string,
		name *string,
		payload *map[string]any,
		userID *string,
		sessionID *string,
	) (*domain.Event, error)
	Delete(ctx context.Context, id string) (*domain.Event, error)
	Query(ctx context.Context, filter types.EventFilter, sort []string, pagination platformTypes.PaginationInputParams) ([]domain.Event, platformTypes.PaginationOutputParams, error)
}

type eventService struct {
	repo repository.EventRepository
}

func NewEventService(repo repository.EventRepository) EventService {
	return &eventService{repo: repo}
}

func (s *eventService) Detail(ctx context.Context, id string) (*domain.Event, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *eventService) Create(ctx context.Context, ev *domain.Event) (*domain.Event, error) {
	now := time.Now().UTC()
	ev.ReceivedAt = now
	if ev.Payload == nil {
		ev.Payload = map[string]any{}
	}

	if err := s.repo.Create(ctx, ev); err != nil {
		return nil, err
	}

	return ev, nil
}

func duplicateKeysInBatch(events []*domain.Event) error {
	seen := make(map[string]struct{}, len(events))
	for _, ev := range events {
		k := ev.ProjectID + "\x00" + ev.MessageID
		if _, ok := seen[k]; ok {
			return ErrDuplicateBatchKeys
		}
		seen[k] = struct{}{}
	}
	return nil
}

func (s *eventService) CreateBatch(ctx context.Context, events []*domain.Event) ([]domain.Event, error) {
	if len(events) == 0 {
		return []domain.Event{}, nil
	}
	if err := duplicateKeysInBatch(events); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	for _, ev := range events {
		ev.ReceivedAt = now
		if ev.Payload == nil {
			ev.Payload = map[string]any{}
		}
	}

	if err := s.repo.CreateBatch(ctx, events); err != nil {
		return nil, err
	}

	out := make([]domain.Event, len(events))
	for i, ev := range events {
		out[i] = *ev
	}
	return out, nil
}

func (s *eventService) Update(
	ctx context.Context,
	id string,
	name *string,
	payload *map[string]any,
	userID *string,
	sessionID *string,
) (*domain.Event, error) {
	ev, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != nil {
		ev.Name = strings.TrimSpace(*name)
	}
	if payload != nil {
		ev.Payload = *payload
	}
	if userID != nil {
		ev.UserID = userID
	}
	if sessionID != nil {
		ev.SessionID = sessionID
	}

	if err := s.repo.Update(ctx, ev); err != nil {
		return nil, err
	}

	return ev, nil
}

func (s *eventService) Delete(ctx context.Context, id string) (*domain.Event, error) {
	return s.repo.Delete(ctx, id)
}

func (s *eventService) Query(
	ctx context.Context,
	filter types.EventFilter,
	sort []string,
	pagination platformTypes.PaginationInputParams,
) ([]domain.Event, platformTypes.PaginationOutputParams, error) {
	filters := map[string]any{}

	if len(filter.IDs) > 0 {
		filters["id"] = filter.IDs
	}
	if filter.ProjectID != nil {
		filters["project_id"] = *filter.ProjectID
	}
	if filter.DistinctID != nil {
		filters["distinct_id"] = *filter.DistinctID
	}
	if filter.Name != nil {
		filters["name"] = fmt.Sprintf("%%%s%%", *filter.Name)
	}
	if filter.MessageID != nil {
		filters["message_id"] = *filter.MessageID
	}
	if filter.UserID != nil {
		filters["user_id"] = *filter.UserID
	}
	if filter.SessionID != nil {
		filters["session_id"] = *filter.SessionID
	}
	if filter.IP != nil {
		filters["ip"] = *filter.IP
	}
	if filter.UserAgent != nil {
		filters["user_agent"] = *filter.UserAgent
	}

	count, err := s.repo.Count(ctx, filters)
	if err != nil {
		return nil, platformTypes.PaginationOutputParams{}, err
	}

	totalPages := int64(math.Ceil(float64(count) / float64(pagination.Limit)))

	if pagination.Page < 1 {
		pagination.Page = 1
	}

	skip := (pagination.Page - 1) * pagination.Limit

	events, err := s.repo.Find(ctx, filters, sort, skip, pagination.Limit)
	if err != nil {
		return nil, platformTypes.PaginationOutputParams{}, err
	}

	return events, platformTypes.PaginationOutputParams{
		Page:       pagination.Page,
		PerPage:    pagination.Limit,
		Total:      count,
		TotalPages: totalPages,
	}, nil
}
