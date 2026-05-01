package http

import (
	"github.com/gin-gonic/gin"
	"github.com/go-web-services/go-service-event/internal/service"
	httpHandler "github.com/go-web-services/go-service-event/internal/transport/http/handler"
	"github.com/go-web-services/go-web-platform/logger"
)

func SetupRouter(
	router *gin.Engine,
	log logger.Logger,
	eventService service.EventService,
) *gin.Engine {
	v1 := router.Group("/api/v1")
	{
		newEventRouterGroup(v1, log, eventService)
	}

	return router
}

func newEventRouterGroup(
	parentGroup *gin.RouterGroup,
	log logger.Logger,
	eventService service.EventService,
) *gin.RouterGroup {
	h := httpHandler.NewEventHandler(log, eventService)
	g := parentGroup.Group("/events")
	{
		g.POST("/create", h.CreateV1)
		g.POST("/create-batch", h.CreateBatchV1)
		g.POST("/update", h.UpdateV1)
		g.POST("/delete", h.DeleteV1)
		g.POST("/detail", h.DetailV1)
		g.POST("/query", h.QueryV1)
	}

	return g
}
