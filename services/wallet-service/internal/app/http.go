package app

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/example/fintech-backend/services/wallet-service/internal/repository"
	"github.com/example/fintech-backend/shared/httpx"
)

type HTTPHandler struct {
	svc *Service
}

func NewHTTPHandler(svc *Service) *HTTPHandler {
	return &HTTPHandler{svc: svc}
}

func (h *HTTPHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/", h.createWallet)
	r.GET("/:user_id", h.getWallet)
	r.POST("/:user_id/deposit", h.deposit)
	r.POST("/:user_id/withdraw", h.withdraw)
	r.POST("/transfer", h.transfer)
	r.GET("/:user_id/transactions", h.listTransactions)
}

func (h *HTTPHandler) createWallet(c *gin.Context) {
	var req CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	wallet, err := h.svc.CreateWallet(c.Request.Context(), req)
	if err != nil {
		httpx.BadRequest(c, err)
		return
	}
	c.JSON(http.StatusCreated, walletDTO(wallet))
}

func (h *HTTPHandler) getWallet(c *gin.Context) {
	wallet, err := h.svc.GetWallet(c.Request.Context(), c.Param("user_id"))
	if err != nil {
		if err == repository.ErrWalletNotFound {
			httpx.NotFound(c, err.Error())
			return
		}
		httpx.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, walletDTO(wallet))
}

func (h *HTTPHandler) deposit(c *gin.Context) {
	var req MovementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	wallet, err := h.svc.Deposit(c.Request.Context(), c.Param("user_id"), req)
	if err != nil {
		if err == ErrInvalidAmount {
			httpx.BadRequest(c, err)
			return
		}
		httpx.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, walletDTO(wallet))
}

func (h *HTTPHandler) withdraw(c *gin.Context) {
	var req MovementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	wallet, err := h.svc.Withdraw(c.Request.Context(), c.Param("user_id"), req)
	if err != nil {
		if err == repository.ErrInsufficientFunds {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		httpx.BadRequest(c, err)
		return
	}
	c.JSON(http.StatusOK, walletDTO(wallet))
}

func (h *HTTPHandler) transfer(c *gin.Context) {
	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	fromWallet, toWallet, err := h.svc.Transfer(c.Request.Context(), req)
	if err != nil {
		if err == repository.ErrInsufficientFunds {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		httpx.BadRequest(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"from_wallet": walletDTO(fromWallet),
		"to_wallet":   walletDTO(toWallet),
	})
}

func (h *HTTPHandler) listTransactions(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	txs, err := h.svc.ListTransactions(c.Request.Context(), c.Param("user_id"), limit, offset)
	if err != nil {
		httpx.Internal(c, err)
		return
	}
	data := make([]gin.H, 0, len(txs))
	for _, tx := range txs {
		data = append(data, gin.H{
			"id":         tx.ID,
			"user_id":    tx.UserID,
			"type":       tx.Type,
			"amount":     FormatCents(tx.AmountCents),
			"reference":  tx.Reference,
			"created_at": tx.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func walletDTO(wallet repository.Wallet) gin.H {
	return gin.H{
		"id":         wallet.ID,
		"user_id":    wallet.UserID,
		"currency":   wallet.Currency,
		"balance":    FormatCents(wallet.BalanceCents),
		"created_at": wallet.CreatedAt,
		"updated_at": wallet.UpdatedAt,
	}
}
