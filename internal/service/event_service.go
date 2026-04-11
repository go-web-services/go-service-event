package service

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/Lomank123/go-service-event/internal/domain"
	"github.com/Lomank123/go-service-event/internal/repository"
	"github.com/Lomank123/go-service-event/internal/types"
	"github.com/Lomank123/go-service-event/pkg/client/enum"
	platformTypes "github.com/Lomank123/go-web-platform/types"
)

type EventService interface {
	Detail(ctx context.Context, id string) (*domain.Event, error)
	Create(
		ctx context.Context,
		name, description, payload string,
		eventType enum.EventType,
		status enum.EventStatus,
	) (*domain.Event, error)
	Update(
		ctx context.Context,
		id, name, description, payload string,
		eventType enum.EventType,
		status enum.EventStatus,
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

func generateSlug(s string) string {
	s = strings.ToLower(s)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func (s *eventService) Create(
	ctx context.Context,
	name, description, payload string,
	eventType enum.EventType,
	status enum.EventStatus,
) (*domain.Event, error) {
	now := time.Now()
	slug := generateSlug(name)

	ev := &domain.Event{
		Name:        name,
		Slug:        slug,
		Description: description,
		Type:        eventType,
		Payload:     payload,
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, ev); err != nil {
		return nil, err
	}

	return ev, nil
}

func (s *eventService) Update(
	ctx context.Context,
	id, name, description, payload string,
	eventType enum.EventType,
	status enum.EventStatus,
) (*domain.Event, error) {
	ev, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		ev.Name = name
		ev.Slug = generateSlug(name)
	}
	if description != "" {
		ev.Description = description
	}
	if payload != "" {
		ev.Payload = payload
	}
	if eventType != "" {
		ev.Type = eventType
	}
	if status != "" {
		ev.Status = status
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
	if filter.Name != nil {
		filters["name"] = fmt.Sprintf("%%%s%%", *filter.Name)
	}
	if filter.Type != nil {
		filters["event_type"] = *filter.Type
	}
	if filter.Status != nil {
		filters["status"] = *filter.Status
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
