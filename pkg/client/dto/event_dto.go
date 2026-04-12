package dto

import (
	"time"
)

type EventDTO struct {
	// ID is the server-generated UUID for this stored row.
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// ProjectID scopes the event to one product or tenant (from your API key, env, or calling service).
	ProjectID string `json:"project_id" example:"proj_live_01"`
	// MessageID uniquely identifies this delivery; use a new UUID per HTTP request for idempotent retries.
	MessageID string `json:"message_id" example:"7c9e6679-7425-40de-944b-e07fc1f90ae7"`
	// DistinctID is the stable analytics identity (e.g. anonymous id in localStorage) for funnels and retention.
	DistinctID string `json:"distinct_id" example:"anon_6ba7b810-9dad-11d1-80b4-00c04fd430c8"`
	// UserID is your logged-in account id when available; omitted before authentication.
	UserID *string `json:"user_id,omitempty" example:"usr_42"`
	// SessionID groups events within one visit or tab session; optional for server-only events.
	SessionID *string `json:"session_id,omitempty" example:"sess_01j8xyz"`
	// IP is the client address stored with the event (caller-supplied, e.g. from edge or client).
	IP *string `json:"ip,omitempty" example:"203.0.113.10"`
	// UserAgent is the raw user-agent string stored with the event (caller-supplied).
	UserAgent *string `json:"user_agent,omitempty" example:"Mozilla/5.0 ..."`
	// Name is the event name / type key (e.g. page_viewed, button_clicked).
	Name string `json:"name" example:"page_viewed"`
	// Payload holds arbitrary JSON properties for segmentation and reporting.
	Payload map[string]any `json:"payload"`
	// OccurredAt is when the action happened on the client (RFC3339); may differ from server time.
	OccurredAt time.Time `json:"occurred_at" example:"2026-04-12T10:30:00Z"`
	// ReceivedAt is when this service persisted the event (set by the server).
	ReceivedAt time.Time `json:"received_at" example:"2026-04-12T10:30:00.456Z"`
	// DeletedAt is set when the row is soft-deleted.
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// EventCreateInputDTO is the JSON body for POST /api/v1/events/create.
type EventCreateInputDTO struct {
	// ProjectID identifies which product or tenant this event belongs to (from API key, config, or backend caller).
	ProjectID string `json:"project_id" validate:"required" example:"proj_live_01"`
	// MessageID is a new UUID per request so retries do not duplicate rows (idempotency key).
	MessageID string `json:"message_id" validate:"required,uuid" example:"7c9e6679-7425-40de-944b-e07fc1f90ae7"`
	// DistinctID is the stable person/device id for analytics (e.g. UUID in localStorage until login).
	DistinctID string `json:"distinct_id" validate:"required" example:"anon_6ba7b810-9dad-11d1-80b4-00c04fd430c8"`
	// UserID is your application’s user id when the visitor is authenticated; omit if anonymous.
	UserID *string `json:"user_id,omitempty" example:"usr_42"`
	// SessionID ties events to one browser session or visit; omit if not tracking sessions.
	SessionID *string `json:"session_id,omitempty" example:"sess_01j8xyz"`
	// IP is the client IP for this event (e.g. from your API gateway or client); omit if unknown.
	IP *string `json:"ip,omitempty" example:"203.0.113.10"`
	// UserAgent is the raw user-agent for this event; omit if unknown.
	UserAgent *string `json:"user_agent,omitempty" example:"Mozilla/5.0 (compatible; ...)"`
	// Name is the event key (snake_case recommended), e.g. signup_started, purchase_completed.
	Name string `json:"name" validate:"required,max=255" example:"page_viewed"`
	// Payload is a JSON object of custom properties (strings, numbers, booleans) for this event.
	Payload map[string]any `json:"payload" validate:"required"`
	// OccurredAt is client-side event time in RFC3339 (used with server received_at for clock skew).
	OccurredAt time.Time `json:"occurred_at" validate:"required" example:"2026-04-12T10:30:00.123Z"`
}

// EventCreateOutputDTO is returned after a successful ingest; see EventDTO for field meanings.
type EventCreateOutputDTO struct {
	// Event is the stored row including server id, received_at, and request-provided ip / user_agent when present.
	Event EventDTO `json:"event"`
}

// EventCreateBatchInputDTO is the JSON body for POST /api/v1/events/create-batch.
// Each item follows EventCreateInputDTO rules; message_id values must be unique within the batch (and across idempotent retries).
type EventCreateBatchInputDTO struct {
	// Events is the list of events to ingest in one atomic transaction (max 100).
	Events []EventCreateInputDTO `json:"events" validate:"required,min=1,max=100,dive"`
}

// EventCreateBatchOutputDTO lists persisted events in the same order as the request.
type EventCreateBatchOutputDTO struct {
	Events []EventDTO `json:"events"`
}

type EventUpdateInputDTO struct {
	ID        string          `json:"id" validate:"required,uuid"`
	Name      *string         `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Payload   *map[string]any `json:"payload,omitempty"`
	UserID    *string         `json:"user_id,omitempty"`
	SessionID *string         `json:"session_id,omitempty"`
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

// EventQueryInputDTO filters events; omit a field to not filter by it. List fields use OR within the same field (SQL IN, or ILIKE OR for names); combined with AND across different fields.
type EventQueryInputDTO struct {
	// IDs filters by event row UUIDs.
	IDs []string `json:"ids" validate:"omitempty,dive,uuid"`
	// ProjectIDs matches any listed project_id.
	ProjectIDs []string `json:"project_ids" validate:"omitempty,dive"`
	// DistinctIDs matches any listed distinct_id.
	DistinctIDs []string `json:"distinct_ids" validate:"omitempty,dive"`
	// Names matches if event name matches any term (substring ILIKE per term).
	Names []string `json:"names" validate:"omitempty,dive,max=255"`
	// MessageIDs matches any listed message_id (UUID each).
	MessageIDs []string `json:"message_ids" validate:"omitempty,dive,uuid"`
	// UserIDs matches any listed user_id.
	UserIDs []string `json:"user_ids" validate:"omitempty,dive"`
	// SessionIDs matches any listed session_id.
	SessionIDs []string `json:"session_ids" validate:"omitempty,dive"`
	// IPs matches any listed ip (exact).
	IPs []string `json:"ips" validate:"omitempty,dive"`
	// UserAgents matches any listed user_agent (exact).
	UserAgents []string `json:"user_agents" validate:"omitempty,dive"`
	Sort       string   `json:"sort" validate:"omitempty"`
	Page       int64    `json:"page" validate:"omitempty,min=1"`
	Limit      int64    `json:"limit" validate:"omitempty,min=1,max=100"`
}

type EventQueryOutputDTO struct {
	Events []EventDTO       `json:"events"`
	Meta   EventListMetaDTO `json:"meta"`
}
