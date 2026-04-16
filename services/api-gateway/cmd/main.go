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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/example/fintech-backend/services/api-gateway/internal/app"
	"github.com/example/fintech-backend/shared/config"
	"github.com/example/fintech-backend/shared/contracts/authpb"
	"github.com/example/fintech-backend/shared/contracts/userpb"
	"github.com/example/fintech-backend/shared/contracts/walletpb"
	"github.com/example/fintech-backend/shared/logger"
	"github.com/example/fintech-backend/shared/metrics"
	"github.com/example/fintech-backend/shared/middleware"
	"github.com/example/fintech-backend/shared/security"
)

func main() {
	cfg := config.Load("api-gateway")
	log, err := logger.New(cfg.ServiceName)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	authConn, err := grpc.DialContext(ctx, cfg.AuthGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("failed to dial auth service", zap.Error(err))
	}
	defer authConn.Close()
	userConn, err := grpc.DialContext(ctx, cfg.UserGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("failed to dial user service", zap.Error(err))
	}
	defer userConn.Close()
	walletConn, err := grpc.DialContext(ctx, cfg.WalletGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("failed to dial wallet service", zap.Error(err))
	}
	defer walletConn.Close()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(metrics.GinMiddleware(cfg.ServiceName))
	r.Use(middleware.RateLimit(cfg.RateLimitRPS, cfg.RateLimitBurst))
	metrics.RegisterEndpoint(r)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	s := app.NewServer(
		authpb.NewAuthServiceClient(authConn),
		userpb.NewUserServiceClient(userConn),
		walletpb.NewWalletServiceClient(walletConn),
		security.NewJWTManager(cfg.JWTSecret, cfg.TokenTTL),
	)
	s.RegisterRoutes(r)

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           r,
		ReadHeaderTimeout: 3 * time.Second,
	}

	go func() {
		log.Info("api gateway started", zap.String("port", cfg.HTTPPort))
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			log.Fatal("gateway serve failed", zap.Error(serveErr))
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = server.Shutdown(shutdownCtx); err != nil {
		log.Error("gateway shutdown failed", zap.Error(err))
	}
	log.Info("api-gateway stopped")
}
