package http

import (
	"net/http"
	"strings"

	"github.com/Lomank123/go-service-event/internal/service"
	"github.com/Lomank123/go-service-event/internal/types"
	"github.com/Lomank123/go-service-event/pkg/client/dto"
	clientValidator "github.com/Lomank123/go-service-event/pkg/client/validator"
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
	v := validator.New()
	_ = v.RegisterValidation("event_type", clientValidator.ValidateEventType)
	_ = v.RegisterValidation("event_status", clientValidator.ValidateEventStatus)

	return &EventHandler{
		log:      log,
		svc:      svc,
		validate: v,
	}
}

// CreateV1
// @Summary Create new event
// @Tags Events
// @Accept json
// @Produce json
// @Param input body dto.EventCreateInputDTO true "Create Event Request"
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

	ev, err := h.svc.Create(
		c.Request.Context(),
		input.Name,
		input.Description,
		input.Payload,
		input.Type,
		input.Status,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.EventCreateOutputDTO{
		Event: dto.EventDTO{
			ID:          ev.ID,
			Name:        ev.Name,
			Slug:        ev.Slug,
			Description: ev.Description,
			Type:        ev.Type,
			Payload:     ev.Payload,
			Status:      ev.Status,
			CreatedAt:   ev.CreatedAt,
			UpdatedAt:   ev.UpdatedAt,
			DeletedAt:   ev.DeletedAt,
		},
	})
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
		input.Description,
		input.Payload,
		input.Type,
		input.Status,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.EventUpdateOutputDTO{
		Event: dto.EventDTO{
			ID:          ev.ID,
			Name:        ev.Name,
			Slug:        ev.Slug,
			Description: ev.Description,
			Type:        ev.Type,
			Payload:     ev.Payload,
			Status:      ev.Status,
			CreatedAt:   ev.CreatedAt,
			UpdatedAt:   ev.UpdatedAt,
			DeletedAt:   ev.DeletedAt,
		},
	})
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
		Event: dto.EventDTO{
			ID:          ev.ID,
			Name:        ev.Name,
			Slug:        ev.Slug,
			Description: ev.Description,
			Type:        ev.Type,
			Payload:     ev.Payload,
			Status:      ev.Status,
			CreatedAt:   ev.CreatedAt,
			UpdatedAt:   ev.UpdatedAt,
			DeletedAt:   ev.DeletedAt,
		},
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

	c.JSON(http.StatusOK, dto.EventDetailOutputDTO{
		Event: dto.EventDTO{
			ID:          ev.ID,
			Name:        ev.Name,
			Slug:        ev.Slug,
			Description: ev.Description,
			Type:        ev.Type,
			Payload:     ev.Payload,
			Status:      ev.Status,
			CreatedAt:   ev.CreatedAt,
			UpdatedAt:   ev.UpdatedAt,
			DeletedAt:   ev.DeletedAt,
		},
	})
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
	if input.Name != "" {
		filters.Name = &input.Name
	}
	if input.Type != "" {
		filters.Type = &input.Type
	}
	if input.Status != "" {
		filters.Status = &input.Status
	}

	var sort []string
	if input.Sort != "" {
		sort = strings.Split(input.Sort, ",")
	}

	var limit int64 = 20
	var page int64 = 1

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

	var result []dto.EventDTO
	for _, ev := range events {
		result = append(result, dto.EventDTO{
			ID:          ev.ID,
			Name:        ev.Name,
			Slug:        ev.Slug,
			Description: ev.Description,
			Type:        ev.Type,
			Payload:     ev.Payload,
			Status:      ev.Status,
			CreatedAt:   ev.CreatedAt,
			UpdatedAt:   ev.UpdatedAt,
			DeletedAt:   ev.DeletedAt,
		})
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
