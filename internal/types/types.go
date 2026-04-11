package types

import (
	"github.com/Lomank123/go-service-event/pkg/client/enum"
)

type EventFilter struct {
	IDs    []string          `json:"ids,omitempty"`
	Name   *string           `json:"name,omitempty"`
	Type   *enum.EventType   `json:"type,omitempty"`
	Status *enum.EventStatus `json:"status,omitempty"`
}
