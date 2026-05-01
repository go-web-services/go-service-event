package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Lomank123/go-service-event/config"
	"github.com/Lomank123/go-service-event/docs"
	_ "github.com/Lomank123/go-service-event/docs"
	"github.com/Lomank123/go-service-event/internal/clients"
	"github.com/Lomank123/go-service-event/internal/repository"
	"github.com/Lomank123/go-service-event/internal/service"
	userHTTP "github.com/Lomank123/go-service-event/internal/transport/http"

	"github.com/gin-gonic/gin"
	platform "github.com/go-web-services/go-web-platform/entrypoint"
	"github.com/go-web-services/go-web-platform/logger"
	platformMiddleware "github.com/go-web-services/go-web-platform/middleware"
)

// @title           Event Service API
// @version         1.0
// @BasePath  		/api

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logg := logger.NewLogger(cfg.App.Env)

	logg.Info("Connecting to PostgreSQL...")
	pgPool, err := clients.NewPostgresClient(cfg.Postgres)
	if err != nil {
		logg.Fatal("PostgreSQL connection error: ", err)
	}
	defer pgPool.Close()
	logg.Info("Successfully connected to PostgreSQL!")

	eventRepo := repository.NewEventRepository(pgPool)
	eventSrv := service.NewEventService(eventRepo)

	router := gin.New()

	platform.SetupPlatform(
		router,
		logg,
		nil,
		platformMiddleware.DefaultLoggingConfig(),
		cfg.App.Env,
	)

	userHTTP.SetupRouter(router, logg, eventSrv)

	swaggerBasePath := "/api"
	if cfg.App.SwaggerBasePath != "" {
		swaggerBasePath = "/" + cfg.App.SwaggerBasePath + swaggerBasePath
	}
	docs.SwaggerInfo.BasePath = swaggerBasePath

	serverAddr := fmt.Sprintf(":%d", cfg.App.Port)
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}
	logg.Info("Starting server on port ", cfg.App.Port)
	go func() {
		if e := srv.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			logg.Fatal("Failed to start HTTP server: ", e)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logg.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if e := srv.Shutdown(ctx); e != nil {
		logg.Fatal("Server forced to shutdown: ", e)
	}

	logg.Info("Server stopped.")
}
