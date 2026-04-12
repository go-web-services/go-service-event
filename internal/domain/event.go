package domain

import (
	"time"
)

// Event is a single ingested analytics occurrence.
type Event struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`

	// Payload is the arbitrary JSON payload for segmentation; set by the emitter with event-specific keys.
	Payload map[string]any `json:"payload"`

	// ProjectID scopes the event to one product/tenant. Usually derived from the API key or token at ingest,
	// or sent by a calling backend; not typically a raw user-typed value from the browser.
	ProjectID string `json:"project_id"`
	// MessageID uniquely identifies this delivery for idempotency. Set by the client (FE or service) per send
	// (e.g. UUID) so retries do not create duplicate rows.
	MessageID string `json:"message_id"`
	// DistinctID is the stable analytics identity for a person/device (e.g. anonymous UUID in localStorage).
	// Set by the client on first visit and reused until identify/login changes how you merge identity.
	DistinctID string `json:"distinct_id"`
	// UserID is your authenticated account id when the user is logged in; set by the client from auth state
	// or omitted until you have authentication. Used with distinct_id for identity after signup.
	UserID *string `json:"user_id,omitempty"`
	// SessionID groups events within one visit or tab session. Typically set by the FE (new id per session);
	// optional for server-only or non-session analytics.
	SessionID *string `json:"session_id,omitempty"`

	// IP is the client address for this event (e.g. from your edge, gateway, or client-reported); optional.
	IP *string `json:"ip,omitempty"`
	// UserAgent is the raw user-agent string for this event; optional.
	UserAgent *string `json:"user_agent,omitempty"`

	// OccurredAt is when the action happened on the client or source system; set by the sender (may skew vs server time).
	OccurredAt time.Time `json:"occurred_at"`
	// ReceivedAt is when the ingest service accepted and stored the event; set by the server at write time.
	ReceivedAt time.Time `json:"received_at"`
}
