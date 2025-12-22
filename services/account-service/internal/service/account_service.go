package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Caesarsage/bankflow/account-service/internal/kafka"
	"github.com/Caesarsage/bankflow/account-service/internal/models"
	"github.com/Caesarsage/bankflow/account-service/internal/repository"
	"github.com/Caesarsage/bankflow/account-service/pkg/account"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidAccountType = errors.New("Invalid account type")
	ErrAccountNotActive   = errors.New("account is not active")
	ErrCannotCloseAccount = errors.New("cannot close account with non-zero balance")
)

type AccountService struct {
	repo     *repository.AccountRepository
	producer *kafka.Producer
}

func NewAccountService(repository *repository.AccountRepository, producer *kafka.Producer) *AccountService {
	return &AccountService{
		repo:     repository,
		producer: producer,
	}
}

// CreateAccount creates a new account
func (s *AccountService) CreateAccount(ctx context.Context, req *models.CreateAccountRequest) (*models.Account, error) {
	// Validate account type
	if req.AccountType != models.AccountTypeChecking && req.AccountType != models.AccountTypeSavings {
		return nil, ErrInvalidAccountType
	}

	// Validate currency
	if req.Currency == "" {
		req.Currency = "NGN"
	}

	// Generate unique account number
	accountNumber, err := account.GenerateAccountNumber()

	if err != nil {
		fmt.Print("error need")
	}

	// Set default interest rate based on account type
	interestRate := decimal.NewFromFloat(0.0)
	if req.InterestRate != nil {
		interestRate = decimal.NewFromFloat(*req.InterestRate)
	} else if req.AccountType == models.AccountTypeSavings {
		interestRate = decimal.NewFromFloat(0.01) // 1% default for savings
	}

	now := time.Now()
	acc := &models.Account{
		ID:               uuid.New(),
		AccountNumber:    accountNumber,
		CustomerID:       req.CustomerID,
		AccountType:      req.AccountType,
		Currency:         req.Currency,
		Balance:          decimal.Zero,
		AvailableBalance: decimal.Zero,
		Status:           models.AccountStatusActive,
		InterestRate:     interestRate,
		OpenedAt:         now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Save to database
	if err := s.repo.CreateAccount(ctx, acc); err != nil {
		return nil, err
	}

	// Publish account created event
	s.publishAccountCreated(acc)

	return acc, nil
}

func (s *AccountService) GetAccountByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	return s.repo.GetAccountByID(ctx, id)
}

func (s *AccountService) GetAccountByNumber(ctx context.Context, accountNumber string) (*models.Account, error) {
	return s.repo.GetAccountByNumber(ctx, accountNumber)
}

// GetAccountsByCustomerID retrieves all accounts for a customer
func (s *AccountService) GetAccountsByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*models.Account, error) {
	return s.repo.GetAccountsByCustomerID(ctx, customerID)
}

// UpdateAccount updates account details
func (s *AccountService) UpdateAccount(ctx context.Context, id uuid.UUID, req *models.UpdateAccountRequest) (*models.Account, error) {
	// Get existing account
	acc, err := s.repo.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Status != nil {
		acc.Status = *req.Status
	}
	if req.InterestRate != nil {
		acc.InterestRate = decimal.NewFromFloat(*req.InterestRate)
	}

	acc.UpdatedAt = time.Now()

	// Save changes
	if err := s.repo.UpdateAccount(ctx, acc); err != nil {
		return nil, err
	}

	return acc, nil
}

// FreezeAccount freezes an account
func (s *AccountService) FreezeAccount(ctx context.Context, id uuid.UUID) error {
	acc, err := s.repo.GetAccountByID(ctx, id)
	if err != nil {
		return err
	}

	if acc.Status == models.AccountStatusFrozen {
		return errors.New("account is already frozen")
	}

	if acc.Status == models.AccountStatusClosed {
		return errors.New("cannot freeze a closed account")
	}

	acc.Status = models.AccountStatusFrozen
	acc.UpdatedAt = time.Now()

	if err := s.repo.UpdateAccount(ctx, acc); err != nil {
		return err
	}

	s.publishAccountFrozen(acc)
	return nil
}

// UnfreezeAccount unfreezes an account
func (s *AccountService) UnfreezeAccount(ctx context.Context, id uuid.UUID) error {
	acc, err := s.repo.GetAccountByID(ctx, id)
	if err != nil {
		return err
	}

	if acc.Status != models.AccountStatusFrozen {
		return errors.New("account is not frozen")
	}

	acc.Status = models.AccountStatusActive
	acc.UpdatedAt = time.Now()

	if err := s.repo.UpdateAccount(ctx, acc); err != nil {
		return err
	}

	return nil
}

// CloseAccount closes an account
func (s *AccountService) CloseAccount(ctx context.Context, id uuid.UUID) error {
	acc, err := s.repo.GetAccountByID(ctx, id)
	if err != nil {
		return err
	}

	// Cannot close account with non-zero balance
	if !acc.Balance.IsZero() {
		return ErrCannotCloseAccount
	}

	// Cannot close already closed account
	if acc.Status == models.AccountStatusClosed {
		return errors.New("account is already closed")
	}

	acc.Status = models.AccountStatusClosed
	now := time.Now()
	acc.ClosedAt = &now
	acc.UpdatedAt = now

	if err := s.repo.UpdateAccount(ctx, acc); err != nil {
		return err
	}

	s.publishAccountClosed(acc)
	return nil
}

// GetBalance gets account balance
func (s *AccountService) GetBalance(ctx context.Context, id uuid.UUID) (decimal.Decimal, decimal.Decimal, error) {
	acc, err := s.repo.GetAccountByID(ctx, id)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	return acc.Balance, acc.AvailableBalance, nil
}

// CreateHold creates a hold on funds
func (s *AccountService) CreateHold(ctx context.Context, accountID uuid.UUID, amount decimal.Decimal, reason string, transactionRef *string) (*models.AccountHold, error) {
	// Validate amount
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("hold amount must be greater than zero")
	}

	// Check if account is active
	acc, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	if acc.Status != models.AccountStatusActive {
		return nil, ErrAccountNotActive
	}

	// Create hold
	hold := &models.AccountHold{
		ID:             uuid.New(),
		AccountID:      accountID,
		Amount:         amount,
		Reason:         reason,
		TransactionRef: transactionRef,
		CreatedAt:      time.Now(),
	}

	if err := s.repo.CreateHold(ctx, hold); err != nil {
		return nil, err
	}

	return hold, nil
}

// ReleaseHold releases a hold
func (s *AccountService) ReleaseHold(ctx context.Context, holdID uuid.UUID) error {
	return s.repo.ReleaseHold(ctx, holdID)
}

// UpdateBalance updates account balance (called by transaction service)
func (s *AccountService) UpdateBalance(ctx context.Context, accountID uuid.UUID, amount decimal.Decimal) error {
	// Get account first to check status
	acc, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return err
	}

	if acc.Status != models.AccountStatusActive {
		return ErrAccountNotActive
	}

	// Update balance
	if err := s.repo.UpdateBalance(ctx, accountID, amount); err != nil {
		return err
	}

	// Get updated account
	updatedAcc, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return err
	}

	// Publish balance updated event
	s.publishBalanceUpdated(updatedAcc)
	return nil
}

// DebitAccount debits amount from account (used for transfers)
func (s *AccountService) DebitAccount(ctx context.Context, accountID uuid.UUID, amount decimal.Decimal) error {
	// Validate amount
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("debit amount must be greater than zero")
	}

	// Debit is a negative amount
	return s.UpdateBalance(ctx, accountID, amount.Neg())
}

// CreditAccount credits amount to account (used for transfers)
func (s *AccountService) CreditAccount(ctx context.Context, accountID uuid.UUID, amount decimal.Decimal) error {
	// Validate amount
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("credit amount must be greater than zero")
	}

	// Credit is a positive amount
	return s.UpdateBalance(ctx, accountID, amount)
}

// GetAccountStatement generates account statement
func (s *AccountService) GetAccountStatement(ctx context.Context, accountID uuid.UUID, startDate, endDate time.Time) (*models.AccountStatement, error) {
	acc, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// TOOD:, this would fetch transactions from transaction service
	// For now, return basic statement structure
	statement := &models.AccountStatement{
		Account:        acc,
		Transactions:   []models.StatementTransaction{},
		StartDate:      startDate,
		EndDate:        endDate,
		OpeningBalance: acc.Balance,
		ClosingBalance: acc.Balance,
	}

	return statement, nil
}

// ValidateAccountOwnership validates that a customer owns an account
func (s *AccountService) ValidateAccountOwnership(ctx context.Context, accountID, customerID uuid.UUID) (bool, error) {
	acc, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return false, err
	}

	return acc.CustomerID == customerID, nil
}

// Event publishing methods
func (s *AccountService) publishAccountCreated(acc *models.Account) {

	event := &kafka.Event{
		EventID:   uuid.NewString(),
		EventType: "account.created",
		AccountID: acc.ID,
		Data: map[string]interface{}{
			"customer_id":    acc.CustomerID.String(),
			"account_number": acc.AccountNumber,
			"account_type":   acc.AccountType,
			"currency":       acc.Currency,
		},
		Timestamp: time.Now(),
	}

	go func() {
		if err := s.producer.PublishEvent(context.Background(), event); err != nil {
			// Log error but don't fail the operation
			// In production, use proper logging
			println("Failed to publish account.created event:", err.Error())
		}

	}()
}

func (s *AccountService) publishBalanceUpdated(acc *models.Account) {
	event := &kafka.Event{
		EventID:   uuid.NewString(),
		EventType: "account.balance.updated",
		AccountID: acc.ID,
		Data: map[string]interface{}{
			"balance":           acc.Balance.String(),
			"available_balance": acc.AvailableBalance.String(),
		},
		Timestamp: time.Now(),
	}

	go func() {
		if err := s.producer.PublishEvent(context.Background(), event); err != nil {
			println("Failed to publish account.balance.updated event:", err.Error())
		}
	}()
}

func (s *AccountService) publishAccountFrozen(acc *models.Account) {
	event := &kafka.Event{
		EventType: "account.frozen",
		AccountID: acc.ID,
		Timestamp: time.Now(),
	}
	go func() {
		if err := s.producer.PublishEvent(context.Background(), event); err != nil {
			println("Failed to publish account.frozen event:", err.Error())
		}
	}()
}

func (s *AccountService) publishAccountClosed(acc *models.Account) {
	event := &kafka.Event{
		EventID:   uuid.NewString(),
		AccountID: acc.ID,
		Timestamp: time.Now(),
	}
	go func() {
		if err := s.producer.PublishEvent(context.Background(), event); err != nil {
			println("Failed to publish account.closed event:", err.Error())
		}
	}()
}
