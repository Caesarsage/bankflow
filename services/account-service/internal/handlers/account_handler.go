package handlers

import (
	"net/http"

	"github.com/Caesarsage/bankflow/account-service/internal/models"
	"github.com/Caesarsage/bankflow/account-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)


type AccountHandler struct {
	service *service.AccountService
}

func NewAccountHandler(service *service.AccountService) *AccountHandler {
	return &AccountHandler{service: service}
}

// CreateAccount handles POST /api/v1/accounts
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := h.service.CreateAccount(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, account)
}

// GetAccount handles GET /api/v1/accounts/:id
func (h *AccountHandler) GetAccount(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	account, err := h.service.GetAccountByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	c.JSON(http.StatusOK, account)
}

// GetAccountByNumber handles GET /api/v1/accounts/number/:number
func (h *AccountHandler) GetAccountByNumber(c *gin.Context) {
	accountNumber := c.Param("number")

	account, err := h.service.GetAccountByNumber(c.Request.Context(), accountNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	c.JSON(http.StatusOK, account)
}

// GetCustomerAccounts handles GET /api/v1/accounts/customer/:customerId
func (h *AccountHandler) GetCustomerAccounts(c *gin.Context) {
	customerID, err := uuid.Parse(c.Param("customerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer ID"})
		return
	}

	accounts, err := h.service.GetAccountsByCustomerID(c.Request.Context(), customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, accounts)
}

// UpdateAccount handles PUT /api/v1/accounts/:id
func (h *AccountHandler) UpdateAccount(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	var req models.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := h.service.UpdateAccount(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, account)
}

// GetBalance handles GET /api/v1/accounts/:id/balance
func (h *AccountHandler) GetBalance(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	balance, availableBalance, err := h.service.GetBalance(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance":           balance,
		"available_balance": availableBalance,
	})
}

// FreezeAccount handles POST /api/v1/accounts/:id/freeze
func (h *AccountHandler) FreezeAccount(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	if err := h.service.FreezeAccount(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account frozen successfully"})
}

// UnfreezeAccount handles POST /api/v1/accounts/:id/unfreeze
func (h *AccountHandler) UnfreezeAccount(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	if err := h.service.UnfreezeAccount(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account unfrozen successfully"})
}

// CloseAccount handles DELETE /api/v1/accounts/:id
func (h *AccountHandler) CloseAccount(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	if err := h.service.CloseAccount(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account closed successfully"})
}

// CreateHold handles POST /api/v1/accounts/:id/holds
func (h *AccountHandler) CreateHold(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	var req struct {
		Amount         float64 `json:"amount" binding:"required,gt=0"`
		Reason         string  `json:"reason" binding:"required"`
		TransactionRef *string `json:"transaction_ref,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	amount := decimal.NewFromFloat(req.Amount)
	hold, err := h.service.CreateHold(c.Request.Context(), id, amount, req.Reason, req.TransactionRef)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, hold)
}

// ReleaseHold handles POST /api/v1/accounts/holds/:holdId/release
func (h *AccountHandler) ReleaseHold(c *gin.Context) {
	holdID, err := uuid.Parse(c.Param("holdId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hold ID"})
		return
	}

	if err := h.service.ReleaseHold(c.Request.Context(), holdID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "hold released successfully"})
}

// RegisterRoutes registers all account routes
func (h *AccountHandler) RegisterRoutes(router *gin.RouterGroup) {
	accounts := router.Group("/accounts")
	{
		accounts.POST("", h.CreateAccount)
		accounts.GET("/:id", h.GetAccount)
		accounts.GET("/number/:number", h.GetAccountByNumber)
		accounts.GET("/customer/:customerId", h.GetCustomerAccounts)
		accounts.PUT("/:id", h.UpdateAccount)
		accounts.GET("/:id/balance", h.GetBalance)
		accounts.POST("/:id/freeze", h.FreezeAccount)
		accounts.POST("/:id/unfreeze", h.UnfreezeAccount)
		accounts.DELETE("/:id", h.CloseAccount)
		accounts.POST("/:id/holds", h.CreateHold)
		accounts.POST("/holds/:holdId/release", h.ReleaseHold)
	}
}
