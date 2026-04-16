package app

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/example/fintech-backend/shared/contracts/authpb"
	"github.com/example/fintech-backend/shared/contracts/userpb"
	"github.com/example/fintech-backend/shared/contracts/walletpb"
	"github.com/example/fintech-backend/shared/grpcx"
	"github.com/example/fintech-backend/shared/httpx"
	"github.com/example/fintech-backend/shared/middleware"
	"github.com/example/fintech-backend/shared/security"
)

type Server struct {
	authClient   authpb.AuthServiceClient
	userClient   userpb.UserServiceClient
	walletClient walletpb.WalletServiceClient
	jwt          *security.JWTManager
}

func NewServer(
	authClient authpb.AuthServiceClient,
	userClient userpb.UserServiceClient,
	walletClient walletpb.WalletServiceClient,
	jwt *security.JWTManager,
) *Server {
	return &Server{
		authClient:   authClient,
		userClient:   userClient,
		walletClient: walletClient,
		jwt:          jwt,
	}
}

func (s *Server) RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	auth := v1.Group("/auth")
	{
		auth.POST("/register", s.register)
		auth.POST("/login", s.login)
	}

	protected := v1.Group("/")
	protected.Use(middleware.JWTAuth(s.jwt))
	{
		users := protected.Group("/users")
		users.POST("/", s.createUser)
		users.GET("/", s.listUsers)
		users.GET("/:id", s.getUser)
		users.PUT("/:id", s.updateUser)
		users.DELETE("/:id", s.deleteUser)

		wallets := protected.Group("/wallets")
		wallets.POST("/", s.createWallet)
		wallets.GET("/:user_id", s.getWallet)
		wallets.POST("/:user_id/deposit", s.deposit)
		wallets.POST("/:user_id/withdraw", s.withdraw)
		wallets.POST("/transfer", s.transfer)
		wallets.GET("/:user_id/transactions", s.listTransactions)
	}
}

func (s *Server) register(c *gin.Context) {
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	in, err := grpcx.ToStruct(req)
	if err != nil {
		httpx.BadRequest(c, err)
		return
	}
	out, err := s.authClient.Register(c.Request.Context(), in)
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusCreated, grpcx.ToMap(out))
}

func (s *Server) login(c *gin.Context) {
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	in, err := grpcx.ToStruct(req)
	if err != nil {
		httpx.BadRequest(c, err)
		return
	}
	out, err := s.authClient.Login(c.Request.Context(), in)
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, grpcx.ToMap(out))
}

func (s *Server) createUser(c *gin.Context) {
	s.proxyUser(c, s.userClient.CreateUser)
}

func (s *Server) listUsers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	in := grpcx.MustStruct(map[string]any{"limit": limit, "offset": offset})
	out, err := s.userClient.ListUsers(c.Request.Context(), in)
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, grpcx.ToMap(out))
}

func (s *Server) getUser(c *gin.Context) {
	in := grpcx.MustStruct(map[string]any{"id": c.Param("id")})
	out, err := s.userClient.GetUser(c.Request.Context(), in)
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, grpcx.ToMap(out))
}

func (s *Server) updateUser(c *gin.Context) {
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	req["id"] = c.Param("id")
	in, err := grpcx.ToStruct(req)
	if err != nil {
		httpx.BadRequest(c, err)
		return
	}
	out, err := s.userClient.UpdateUser(c.Request.Context(), in)
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, grpcx.ToMap(out))
}

func (s *Server) deleteUser(c *gin.Context) {
	in := grpcx.MustStruct(map[string]any{"id": c.Param("id")})
	_, err := s.userClient.DeleteUser(c.Request.Context(), in)
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) createWallet(c *gin.Context) {
	s.proxyWallet(c, s.walletClient.CreateWallet, http.StatusCreated)
}

func (s *Server) getWallet(c *gin.Context) {
	in := grpcx.MustStruct(map[string]any{"user_id": c.Param("user_id")})
	out, err := s.walletClient.GetWallet(c.Request.Context(), in)
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, grpcx.ToMap(out))
}

func (s *Server) deposit(c *gin.Context) {
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	req["user_id"] = c.Param("user_id")
	in, err := grpcx.ToStruct(req)
	if err != nil {
		httpx.BadRequest(c, err)
		return
	}
	out, err := s.walletClient.Deposit(c.Request.Context(), in)
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, grpcx.ToMap(out))
}

func (s *Server) withdraw(c *gin.Context) {
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	req["user_id"] = c.Param("user_id")
	in, err := grpcx.ToStruct(req)
	if err != nil {
		httpx.BadRequest(c, err)
		return
	}
	out, err := s.walletClient.Withdraw(c.Request.Context(), in)
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, grpcx.ToMap(out))
}

func (s *Server) transfer(c *gin.Context) {
	s.proxyWallet(c, s.walletClient.Transfer, http.StatusOK)
}

func (s *Server) listTransactions(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	in := grpcx.MustStruct(map[string]any{
		"user_id": c.Param("user_id"),
		"limit":   limit,
		"offset":  offset,
	})
	out, err := s.walletClient.ListTransactions(c.Request.Context(), in)
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, grpcx.ToMap(out))
}

func (s *Server) proxyUser(c *gin.Context, fn func(context.Context, *structpb.Struct, ...grpc.CallOption) (*structpb.Struct, error)) {
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	in, err := grpcx.ToStruct(req)
	if err != nil {
		httpx.BadRequest(c, err)
		return
	}
	out, err := fn(c.Request.Context(), in)
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusCreated, grpcx.ToMap(out))
}

func (s *Server) proxyWallet(c *gin.Context, fn func(context.Context, *structpb.Struct, ...grpc.CallOption) (*structpb.Struct, error), statusCode int) {
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	in, err := grpcx.ToStruct(req)
	if err != nil {
		httpx.BadRequest(c, err)
		return
	}
	out, err := fn(c.Request.Context(), in)
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(statusCode, grpcx.ToMap(out))
}

func writeGRPCError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if ok {
		switch st.Code() {
		case codes.InvalidArgument:
			c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
		case codes.Unauthenticated:
			c.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
		case codes.NotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
		case codes.FailedPrecondition:
			c.JSON(http.StatusConflict, gin.H{"error": st.Message()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		}
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
