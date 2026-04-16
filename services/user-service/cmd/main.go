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

	"github.com/example/fintech-backend/services/user-service/internal/app"
	"github.com/example/fintech-backend/services/user-service/internal/repository"
	"github.com/example/fintech-backend/shared/cache"
	"github.com/example/fintech-backend/shared/config"
	"github.com/example/fintech-backend/shared/contracts/userpb"
	"github.com/example/fintech-backend/shared/db"
	"github.com/example/fintech-backend/shared/logger"
	"github.com/example/fintech-backend/shared/metrics"
)

func main() {
	cfg := config.Load("user-service")
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

	redisClient, err := cache.NewRedisClient(ctx, cfg.RedisAddr, cfg.RedisPass, cfg.RedisDB)
	if err != nil {
		log.Warn("redis unavailable, continuing without cache", zap.Error(err))
	}
	if redisClient != nil {
		defer redisClient.Close()
	}

	userService := app.NewService(repo, redisClient)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(metrics.GinMiddleware(cfg.ServiceName))
	metrics.RegisterEndpoint(r)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	app.NewHTTPHandler(userService).RegisterRoutes(r.Group("/api/v1/users"))

	httpServer := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           r,
		ReadHeaderTimeout: 3 * time.Second,
	}

	grpcServer := grpc.NewServer()
	userpb.RegisterUserServiceServer(grpcServer, app.NewGRPCServer(userService))
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal("grpc listen failed", zap.Error(err))
	}

	go func() {
		log.Info("user grpc server started", zap.String("port", cfg.GRPCPort))
		if serveErr := grpcServer.Serve(lis); serveErr != nil {
			log.Fatal("grpc serve failed", zap.Error(serveErr))
		}
	}()
	go func() {
		log.Info("user http server started", zap.String("port", cfg.HTTPPort))
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
	log.Info("user-service stopped")
}
