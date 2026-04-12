package types

type EventFilter struct {
	IDs        []string `json:"ids,omitempty"`
	ProjectID  *string  `json:"project_id,omitempty"`
	DistinctID *string  `json:"distinct_id,omitempty"`
	Name       *string  `json:"name,omitempty"`
	MessageID  *string  `json:"message_id,omitempty"`
	UserID     *string  `json:"user_id,omitempty"`
	SessionID  *string  `json:"session_id,omitempty"`
	IP         *string  `json:"ip,omitempty"`
	UserAgent  *string  `json:"user_agent,omitempty"`
}
