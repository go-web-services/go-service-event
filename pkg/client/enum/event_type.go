package enum

type EventType string

const (
	EventTypeUser    EventType = "user"
	EventTypeSystem  EventType = "system"
	EventTypeWebhook EventType = "webhook"
)

var AllEventTypes = []EventType{
	EventTypeUser,
	EventTypeSystem,
	EventTypeWebhook,
}
