package domain

import (
	"time"

	"github.com/Lomank123/go-service-event/pkg/client/enum"
)

type Event struct {
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
