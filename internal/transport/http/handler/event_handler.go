package http

import (
	"errors"
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
// @Description Accepts EventCreateInputDTO. Same project_id + message_id on retry returns the existing event (idempotent). ip and user_agent are optional fields on the JSON body.
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
		IP:         input.IP,
		UserAgent:  input.UserAgent,
		Name:       strings.TrimSpace(input.Name),
		Payload:    input.Payload,
		OccurredAt: input.OccurredAt,
	}

	out, err := h.svc.Create(c.Request.Context(), ev)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.EventCreateOutputDTO{Event: eventToDTO(*out)})
}

// CreateBatchV1
// @Summary Ingest multiple analytics events (atomic batch)
// @Description Accepts EventCreateBatchInputDTO. All rows are inserted in one transaction. Same rules as single create; duplicate project_id+message_id within the request body is rejected. Each item may include ip and user_agent on the JSON body.
// @Tags Events
// @Accept json
// @Produce json
// @Param input body dto.EventCreateBatchInputDTO true "Batch payload"
// @Success 200 {object} dto.EventCreateBatchOutputDTO
// @Router /v1/events/create-batch [post]
func (h *EventHandler) CreateBatchV1(c *gin.Context) {
	var input dto.EventCreateBatchInputDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(platformError.ErrInvalidRequestPayload)
		return
	}

	if err := h.validate.Struct(input); err != nil {
		_ = c.Error(err)
		return
	}

	events := make([]*domain.Event, 0, len(input.Events))
	for i := range input.Events {
		in := &input.Events[i]
		events = append(events, &domain.Event{
			ProjectID:  in.ProjectID,
			MessageID:  in.MessageID,
			DistinctID: in.DistinctID,
			UserID:     in.UserID,
			SessionID:  in.SessionID,
			IP:         in.IP,
			UserAgent:  in.UserAgent,
			Name:       strings.TrimSpace(in.Name),
			Payload:    in.Payload,
			OccurredAt: in.OccurredAt,
		})
	}

	out, err := h.svc.CreateBatch(c.Request.Context(), events)
	if err != nil {
		if errors.Is(err, service.ErrDuplicateBatchKeys) {
			_ = c.Error(platformError.ErrInvalidRequestPayload)
			return
		}
		_ = c.Error(err)
		return
	}

	dtos := make([]dto.EventDTO, 0, len(out))
	for i := range out {
		dtos = append(dtos, eventToDTO(out[i]))
	}
	c.JSON(http.StatusOK, dto.EventCreateBatchOutputDTO{Events: dtos})
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
// @Description Filters: all list fields use OR within the field (SQL IN, or ILIKE OR for names). names are substring terms; ips and user_agents are exact. Combined with AND across different fields.
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
	if len(input.ProjectIDs) > 0 {
		filters.ProjectIDs = input.ProjectIDs
	}
	if len(input.DistinctIDs) > 0 {
		filters.DistinctIDs = input.DistinctIDs
	}
	if len(input.Names) > 0 {
		filters.Names = input.Names
	}
	if len(input.MessageIDs) > 0 {
		filters.MessageIDs = input.MessageIDs
	}
	if len(input.UserIDs) > 0 {
		filters.UserIDs = input.UserIDs
	}
	if len(input.SessionIDs) > 0 {
		filters.SessionIDs = input.SessionIDs
	}
	if len(input.IPs) > 0 {
		filters.IPs = input.IPs
	}
	if len(input.UserAgents) > 0 {
		filters.UserAgents = input.UserAgents
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
