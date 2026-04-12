package http

import (
	"net/http"
	"strings"

	"github.com/Lomank123/go-service-event/internal/domain"
	"github.com/Lomank123/go-service-event/internal/service"
	"github.com/Lomank123/go-service-event/internal/types"
	"github.com/Lomank123/go-service-event/pkg/client/dto"
	platformError "github.com/Lomank123/go-web-platform/error"
	"github.com/Lomank123/go-web-platform/logger"
	platformTypes "github.com/Lomank123/go-web-platform/types"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type EventHandler struct {
	log      logger.Logger
	svc      service.EventService
	validate *validator.Validate
}

func NewEventHandler(log logger.Logger, svc service.EventService) *EventHandler {
	return &EventHandler{
		log:      log,
		svc:      svc,
		validate: validator.New(),
	}
}

func eventToDTO(ev domain.Event) dto.EventDTO {
	return dto.EventDTO{
		ID:         ev.ID,
		ProjectID:  ev.ProjectID,
		MessageID:  ev.MessageID,
		DistinctID: ev.DistinctID,
		UserID:     ev.UserID,
		SessionID:  ev.SessionID,
		IP:         ev.IP,
		UserAgent:  ev.UserAgent,
		Name:       ev.Name,
		Payload:    ev.Payload,
		OccurredAt: ev.OccurredAt,
		ReceivedAt: ev.ReceivedAt,
		DeletedAt:  ev.DeletedAt,
	}
}

// CreateV1
// @Summary Ingest analytics event
// @Description Accepts EventCreateInputDTO. Same project_id + message_id on retry returns the existing event (idempotent). ip and user_agent are not sent in JSON; the server sets them from the HTTP request.
// @Tags Events
// @Accept json
// @Produce json
// @Param input body dto.EventCreateInputDTO true "Event payload (see model for per-field descriptions)"
// @Success 200 {object} dto.EventCreateOutputDTO
// @Router /v1/events/create [post]
func (h *EventHandler) CreateV1(c *gin.Context) {
	var input dto.EventCreateInputDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(platformError.ErrInvalidRequestPayload)
		return
	}

	if err := h.validate.Struct(input); err != nil {
		_ = c.Error(err)
		return
	}

	ev := &domain.Event{
		ProjectID:  input.ProjectID,
		MessageID:  input.MessageID,
		DistinctID: input.DistinctID,
		UserID:     input.UserID,
		SessionID:  input.SessionID,
		Name:       strings.TrimSpace(input.Name),
		Payload:    input.Payload,
		OccurredAt: input.OccurredAt,
	}
	if ip := c.ClientIP(); ip != "" {
		ev.IP = &ip
	}
	if ua := c.Request.UserAgent(); ua != "" {
		ev.UserAgent = &ua
	}

	out, err := h.svc.Create(c.Request.Context(), ev)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.EventCreateOutputDTO{Event: eventToDTO(*out)})
}

// UpdateV1
// @Summary Update event
// @Tags Events
// @Accept json
// @Produce json
// @Param input body dto.EventUpdateInputDTO true "Update Event Request"
// @Success 200 {object} dto.EventUpdateOutputDTO
// @Router /v1/events/update [post]
func (h *EventHandler) UpdateV1(c *gin.Context) {
	var input dto.EventUpdateInputDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(platformError.ErrInvalidRequestPayload)
		return
	}

	if err := h.validate.Struct(input); err != nil {
		_ = c.Error(err)
		return
	}

	ev, err := h.svc.Update(
		c.Request.Context(),
		input.ID,
		input.Name,
		input.Payload,
		input.UserID,
		input.SessionID,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.EventUpdateOutputDTO{Event: eventToDTO(*ev)})
}

// DeleteV1
// @Summary Soft delete event (sets deleted_at)
// @Tags Events
// @Accept json
// @Produce json
// @Param input body dto.EventDeleteInputDTO true "Delete Event Request"
// @Success 200 {object} dto.EventDeleteOutputDTO
// @Router /v1/events/delete [post]
func (h *EventHandler) DeleteV1(c *gin.Context) {
	var input dto.EventDeleteInputDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(platformError.ErrInvalidRequestPayload)
		return
	}

	if err := h.validate.Struct(input); err != nil {
		_ = c.Error(err)
		return
	}

	ev, err := h.svc.Delete(c.Request.Context(), input.ID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.EventDeleteOutputDTO{
		Message: "Event deleted successfully",
		Event:   eventToDTO(*ev),
	})
}

// DetailV1
// @Summary Get event detail
// @Tags Events
// @Accept json
// @Produce json
// @Param input body dto.EventDetailInputDTO true "Get Event Request"
// @Success 200 {object} dto.EventDetailOutputDTO
// @Router /v1/events/detail [post]
func (h *EventHandler) DetailV1(c *gin.Context) {
	var input dto.EventDetailInputDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(platformError.ErrInvalidRequestPayload)
		return
	}

	if err := h.validate.Struct(input); err != nil {
		_ = c.Error(err)
		return
	}

	ev, err := h.svc.Detail(c.Request.Context(), input.ID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.EventDetailOutputDTO{Event: eventToDTO(*ev)})
}

// QueryV1
// @Summary Returns list of events by query
// @Tags Events
// @Accept json
// @Produce json
// @Param input body dto.EventQueryInputDTO true "Query parameters"
// @Success 200 {object} dto.EventQueryOutputDTO
// @Router /v1/events/query [post]
func (h *EventHandler) QueryV1(c *gin.Context) {
	var input dto.EventQueryInputDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(platformError.ErrInvalidRequestPayload)
		return
	}

	if err := h.validate.Struct(input); err != nil {
		_ = c.Error(err)
		return
	}

	filters := types.EventFilter{}
	if len(input.IDs) > 0 {
		filters.IDs = input.IDs
	}
	if input.ProjectID != "" {
		filters.ProjectID = &input.ProjectID
	}
	if input.DistinctID != "" {
		filters.DistinctID = &input.DistinctID
	}
	if input.Name != "" {
		filters.Name = &input.Name
	}
	if input.MessageID != "" {
		filters.MessageID = &input.MessageID
	}
	if input.UserID != "" {
		filters.UserID = &input.UserID
	}
	if input.SessionID != "" {
		filters.SessionID = &input.SessionID
	}
	if input.IP != "" {
		filters.IP = &input.IP
	}
	if input.UserAgent != "" {
		filters.UserAgent = &input.UserAgent
	}

	sort := []string{"-occurred_at"}
	if input.Sort != "" {
		sort = strings.Split(input.Sort, ",")
	}

	limit := int64(20)
	page := int64(1)
	if input.Limit != 0 {
		limit = input.Limit
	}
	if input.Page != 0 {
		page = input.Page
	}

	paginationInput := platformTypes.PaginationInputParams{
		Page:  page,
		Limit: limit,
	}

	events, paginationOutput, err := h.svc.Query(c.Request.Context(), filters, sort, paginationInput)
	if err != nil {
		_ = c.Error(err)
		return
	}

	result := make([]dto.EventDTO, 0, len(events))
	for _, ev := range events {
		result = append(result, eventToDTO(ev))
	}

	c.JSON(http.StatusOK, dto.EventQueryOutputDTO{
		Events: result,
		Meta: dto.EventListMetaDTO{
			Pagination: dto.PaginationOutputParams{
				Page:       paginationOutput.Page,
				TotalPages: paginationOutput.TotalPages,
				PerPage:    paginationOutput.PerPage,
				Total:      paginationOutput.Total,
			},
		},
	})
}
