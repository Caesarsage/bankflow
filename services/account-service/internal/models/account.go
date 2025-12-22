package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// AccountType represents the type of account
type AccountType string

const (
	AccountTypeChecking AccountType = "CURRENT"
	AccountTypeSavings  AccountType = "SAVINGS"
)

// AccountStatus represents the status of an account
type AccountStatus string

const (
	AccountStatusActive  AccountStatus = "ACTIVE"
	AccountStatusFrozen  AccountStatus = "FROZEN"
	AccountStatusClosed  AccountStatus = "CLOSED"
	AccountStatusPending AccountStatus = "PENDING"
)

// Account represents a bank account
type Account struct {
	ID               uuid.UUID       `json:"id" db:"id"`
	AccountNumber    string          `json:"account_number" db:"account_number"`
	CustomerID       uuid.UUID       `json:"customer_id" db:"customer_id"`
	AccountType      AccountType     `json:"account_type" db:"account_type"`
	Currency         string          `json:"currency" db:"currency"`
	Balance          decimal.Decimal `json:"balance" db:"balance"`
	AvailableBalance decimal.Decimal `json:"available_balance" db:"available_balance"`
	Status           AccountStatus   `json:"status" db:"status"`
	InterestRate     decimal.Decimal `json:"interest_rate" db:"interest_rate"`
	OpenedAt         time.Time       `json:"opened_at" db:"opened_at"`
	ClosedAt         *time.Time      `json:"closed_at,omitempty" db:"closed_at"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`
}

// AccountHold represents a hold on funds
type AccountHold struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	AccountID      uuid.UUID       `json:"account_id" db:"account_id"`
	Amount         decimal.Decimal `json:"amount" db:"amount"`
	Reason         string          `json:"reason" db:"reason"`
	TransactionRef *string         `json:"transaction_ref,omitempty" db:"transaction_ref"`
	ExpiresAt      *time.Time      `json:"expires_at,omitempty" db:"expires_at"`
	ReleasedAt     *time.Time      `json:"released_at,omitempty" db:"released_at"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
}

// CreateAccountRequest represents the request to create an account
type CreateAccountRequest struct {
	CustomerID   uuid.UUID   `json:"customer_id" binding:"required"`
	AccountType  AccountType `json:"account_type" binding:"required,oneof=CHECKING SAVINGS"`
	Currency     string      `json:"currency" binding:"required,len=3"`
	InterestRate *float64    `json:"interest_rate,omitempty"`
}

// UpdateAccountRequest represents the request to update an account
type UpdateAccountRequest struct {
	Status       *AccountStatus `json:"status,omitempty"`
	InterestRate *float64        `json:"interest_rate,omitempty"`
}


// AccountStatement represents account statement data
type AccountStatement struct {
	Account      *Account              `json:"account"`
	Transactions []StatementTransaction `json:"transactions"`
	StartDate    time.Time             `json:"start_date"`
	EndDate      time.Time             `json:"end_date"`
	OpeningBalance decimal.Decimal     `json:"opening_balance"`
	ClosingBalance decimal.Decimal     `json:"closing_balance"`
}

// StatementTransaction represents a transaction in statement
type StatementTransaction struct {
	Date        time.Time       `json:"date"`
	Description string          `json:"description"`
	Amount      decimal.Decimal `json:"amount"`
	Balance     decimal.Decimal `json:"balance"`
	Type        string          `json:"type"`
}

// HoldFundsRequest represents the request to hold funds
type HoldFundsRequest struct {
	Amount         decimal.Decimal `json:"amount" binding:"required,gt=0"`
	Reason         string          `json:"reason" binding:"required"`
	TransactionRef *string         `json:"transaction_ref,omitempty"`
	ExpiresAt      *time.Time      `json:"expires_at,omitempty"`
}

// ReleaseFundsRequest represents the request to release held funds
type ReleaseFundsRequest struct {
	HoldID uuid.UUID `json:"hold_id" binding:"required"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
