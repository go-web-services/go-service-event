package types

type EventFilter struct {
	IDs         []string `json:"ids,omitempty"`
	ProjectIDs  []string `json:"project_ids,omitempty"`
	DistinctIDs []string `json:"distinct_ids,omitempty"`
	Names       []string `json:"names,omitempty"`
	MessageIDs  []string `json:"message_ids,omitempty"`
	UserIDs     []string `json:"user_ids,omitempty"`
	SessionIDs  []string `json:"session_ids,omitempty"`
	IPs         []string `json:"ips,omitempty"`
	UserAgents  []string `json:"user_agents,omitempty"`
}
