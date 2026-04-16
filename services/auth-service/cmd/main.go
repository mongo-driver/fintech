package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/example/fintech-backend/services/auth-service/internal/app"
	"github.com/example/fintech-backend/services/auth-service/internal/repository"
	"github.com/example/fintech-backend/shared/config"
	"github.com/example/fintech-backend/shared/contracts/authpb"
	"github.com/example/fintech-backend/shared/db"
	"github.com/example/fintech-backend/shared/events"
	"github.com/example/fintech-backend/shared/logger"
	"github.com/example/fintech-backend/shared/metrics"
	"github.com/example/fintech-backend/shared/security"
)

func main() {
	cfg := config.Load("auth-service")
	log, err := logger.New(cfg.ServiceName)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPostgresPool(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatal("postgres init failed", zap.Error(err))
	}
	defer pool.Close()

	repo := repository.NewPostgresRepository(pool)
	if err = repo.Migrate(ctx); err != nil {
		log.Fatal("migration failed", zap.Error(err))
	}

	publisher := events.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopic)
	defer publisher.Close()

	jwtManager := security.NewJWTManager(cfg.JWTSecret, cfg.TokenTTL)
	authService := app.NewService(repo, jwtManager, publisher)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(metrics.GinMiddleware(cfg.ServiceName))
	metrics.RegisterEndpoint(r)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	app.NewHTTPHandler(authService).RegisterRoutes(r.Group("/api/v1/auth"))

	httpServer := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           r,
		ReadHeaderTimeout: 3 * time.Second,
	}

	grpcServer := grpc.NewServer()
	authpb.RegisterAuthServiceServer(grpcServer, app.NewGRPCServer(authService))
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal("grpc listen failed", zap.Error(err))
	}

	go func() {
		log.Info("auth grpc server started", zap.String("port", cfg.GRPCPort))
		if serveErr := grpcServer.Serve(lis); serveErr != nil {
			log.Fatal("grpc serve failed", zap.Error(serveErr))
		}
	}()

	go func() {
		log.Info("auth http server started", zap.String("port", cfg.HTTPPort))
		if serveErr := httpServer.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			log.Fatal("http serve failed", zap.Error(serveErr))
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	if err = httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("http shutdown failed", zap.Error(err))
	}
	log.Info("auth-service stopped")
}
