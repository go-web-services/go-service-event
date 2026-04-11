package dto

import (
	"time"

	"github.com/Lomank123/go-service-event/pkg/client/enum"
)

type EventDTO struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Slug        string           `json:"slug"`
	Description string           `json:"description,omitempty"`
	Type        enum.EventType   `json:"type"`
	Payload     string           `json:"payload"`
	Status      enum.EventStatus `json:"status"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	DeletedAt   *time.Time       `json:"deleted_at,omitempty"`
}

type EventCreateInputDTO struct {
	Name        string           `json:"name" validate:"required,max=255"`
	Description string           `json:"description,omitempty" validate:"omitempty,max=1024"`
	Payload     string           `json:"payload" validate:"required"`
	Type        enum.EventType   `json:"type" validate:"required,event_type"`
	Status      enum.EventStatus `json:"status" validate:"required,event_status"`
}

type EventCreateOutputDTO struct {
	Event EventDTO `json:"event"`
}

type EventUpdateInputDTO struct {
	ID          string           `json:"id" validate:"required,uuid"`
	Name        string           `json:"name,omitempty" validate:"omitempty,max=255"`
	Description string           `json:"description,omitempty" validate:"omitempty,max=1024"`
	Payload     string           `json:"payload,omitempty" validate:"omitempty"`
	Type        enum.EventType   `json:"type,omitempty" validate:"omitempty,event_type"`
	Status      enum.EventStatus `json:"status,omitempty" validate:"omitempty,event_status"`
}

type EventUpdateOutputDTO struct {
	Event EventDTO `json:"event"`
}

type EventDeleteInputDTO struct {
	ID string `json:"id" validate:"required,uuid"`
}

type EventDeleteOutputDTO struct {
	Message string   `json:"message"`
	Event   EventDTO `json:"event"`
}

type EventDetailInputDTO struct {
	ID string `json:"id" validate:"required,uuid"`
}

type EventDetailOutputDTO struct {
	Event EventDTO `json:"event"`
}

type PaginationOutputParams struct {
	Page       int64 `json:"page"`
	TotalPages int64 `json:"total_pages"`
	PerPage    int64 `json:"per_page"`
	Total      int64 `json:"total"`
}

type EventListMetaDTO struct {
	Pagination PaginationOutputParams `json:"pagination"`
}

type EventQueryInputDTO struct {
	IDs     []string         `json:"ids" validate:"omitempty,dive,uuid"`
	Name    string           `json:"name" validate:"omitempty,max=255"`
	Type    enum.EventType   `json:"type" validate:"omitempty,event_type"`
	Status  enum.EventStatus `json:"status" validate:"omitempty,event_status"`
	Sort    string           `json:"sort" validate:"omitempty"`
	Page    int64            `json:"page" validate:"omitempty,min=1"`
	Limit   int64            `json:"limit" validate:"omitempty,min=1,max=100"`
}

type EventQueryOutputDTO struct {
	Events []EventDTO       `json:"events"`
	Meta   EventListMetaDTO `json:"meta"`
}
