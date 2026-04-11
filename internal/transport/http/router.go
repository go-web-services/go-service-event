package http

import (
	"github.com/Lomank123/go-service-event/internal/service"
	httpHandler "github.com/Lomank123/go-service-event/internal/transport/http/handler"
	"github.com/Lomank123/go-web-platform/logger"
	"github.com/gin-gonic/gin"
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
		g.POST("/update", h.UpdateV1)
		g.POST("/delete", h.DeleteV1)
		g.POST("/detail", h.DetailV1)
		g.POST("/query", h.QueryV1)
	}

	return g
}
