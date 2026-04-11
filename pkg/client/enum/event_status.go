package enum

type EventStatus string

const (
	EventStatusPending    EventStatus = "pending"
	EventStatusProcessed  EventStatus = "processed"
	EventStatusFailed     EventStatus = "failed"
)

var AllEventStatuses = []EventStatus{
	EventStatusPending,
	EventStatusProcessed,
	EventStatusFailed,
}
