package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/example/fintech-backend/services/notification-service/internal/app"
	"github.com/example/fintech-backend/shared/config"
	"github.com/example/fintech-backend/shared/events"
	"github.com/example/fintech-backend/shared/logger"
	"github.com/example/fintech-backend/shared/metrics"
)

func main() {
	cfg := config.Load("notification-service")
	log, err := logger.New(cfg.ServiceName)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	svc := app.NewService(app.NewLogSender(log), log)
	consumer := events.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, "notification-service-group")
	defer consumer.Close()

	go func() {
		if err = svc.Consume(ctx, consumer); err != nil {
			log.Error("notification consumer stopped with error", zap.Error(err))
		}
	}()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(metrics.GinMiddleware(cfg.ServiceName))
	metrics.RegisterEndpoint(r)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	app.NewHTTPHandler(svc).RegisterRoutes(r.Group("/api/v1/notifications"))

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           r,
		ReadHeaderTimeout: 3 * time.Second,
	}

	go func() {
		log.Info("notification http server started", zap.String("port", cfg.HTTPPort))
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			log.Fatal("http serve failed", zap.Error(serveErr))
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = server.Shutdown(shutdownCtx); err != nil {
		log.Error("http shutdown failed", zap.Error(err))
	}
	log.Info("notification-service stopped")
}
