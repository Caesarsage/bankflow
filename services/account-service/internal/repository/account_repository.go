package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Caesarsage/bankflow/account-service/internal/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrAccountNotFound      = errors.New("account not found")
	ErrAccountAlreadyExists = errors.New("account already exists")
	ErrHoldNotFound         = errors.New("hold not found")
	ErrInsufficientFunds    = errors.New("insufficient funds")
)

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{
		db: db,
	}
}

func (r *AccountRepository) CreateAccount(ctx context.Context, account *models.Account) error {
	query := `
		INSERT INTO accounts (
			id, account_number, customer_id, account_type, currency, balance,
			available_balance, status, interest_rate, opened_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.ExecContext(ctx, query,
		account.ID,
		account.AccountNumber,
		account.CustomerID,
		account.AccountType,
		account.Currency,
		account.Balance,
		account.AvailableBalance,
		account.Status,
		account.InterestRate,
		account.OpenedAt,
		account.CreatedAt,
		account.UpdatedAt,
	)

	return err
}

func (r *AccountRepository) GetAccountByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	query := `
		SELECT id, account_number, customer_id, account_type, currency,
		       balance, available_balance, status, interest_rate,
		       opened_at, closed_at, created_at, updated_at
		FROM accounts
		WHERE id = $1
	`

	account := &models.Account{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&account.ID,
		&account.AccountNumber,
		&account.CustomerID,
		&account.AccountType,
		&account.Currency,
		&account.Balance,
		&account.AvailableBalance,
		&account.Status,
		&account.InterestRate,
		&account.OpenedAt,
		&account.ClosedAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrAccountNotFound
	}
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (r *AccountRepository) GetAccountByNumber(ctx context.Context, accountNumber string) (*models.Account, error) {
	query := `
		SELECT id, account_number, customer_id, account_type, currency,
		       balance, available_balance, status, interest_rate,
		       opened_at, closed_at, created_at, updated_at
		FROM accounts
		WHERE account_number = $1
	`

	account := &models.Account{}
	err := r.db.QueryRowContext(ctx, query, accountNumber).Scan(
		&account.ID,
		&account.AccountNumber,
		&account.CustomerID,
		&account.AccountType,
		&account.Currency,
		&account.Balance,
		&account.AvailableBalance,
		&account.Status,
		&account.InterestRate,
		&account.OpenedAt,
		&account.ClosedAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrAccountNotFound
	}
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (r *AccountRepository) GetAccountsByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*models.Account, error) {
	query := `
		SELECT id, account_number, customer_id, account_type, currency,
		       balance, available_balance, status, interest_rate,
		       opened_at, closed_at, created_at, updated_at
		FROM accounts
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []*models.Account{}
	for rows.Next() {
		account := &models.Account{}
		err := rows.Scan(
			&account.ID,
			&account.AccountNumber,
			&account.CustomerID,
			&account.AccountType,
			&account.Currency,
			&account.Balance,
			&account.AvailableBalance,
			&account.Status,
			&account.InterestRate,
			&account.OpenedAt,
			&account.ClosedAt,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// UpdateAccount updates account status and interest rate
func (r *AccountRepository) UpdateAccount(ctx context.Context, account *models.Account) error {
	query := `
		UPDATE accounts
		SET status = $1, interest_rate = $2, updated_at = $3
		WHERE id = $4
	`

	result, err := r.db.ExecContext(ctx, query,
		account.Status,
		account.InterestRate,
		time.Now(),
		account.ID,
	)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrAccountNotFound
	}

	return nil
}

// UpdateBalance updates account balance (called by transaction service)
func (r *AccountRepository) UpdateBalance(ctx context.Context, accountID uuid.UUID, amount decimal.Decimal) error {
	query := `
		UPDATE accounts
		SET balance = balance + $1,
		    available_balance = available_balance + $1,
		    updated_at = $2
		WHERE id = $3 AND balance + $1 >= 0
	`

	result, err := r.db.ExecContext(ctx, query, amount, time.Now(), accountID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		// Check if account exists
		var exists bool
		err = r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1)", accountID).Scan(&exists)
		if err != nil {
			return err
		}

		if !exists {
			return ErrAccountNotFound
		}

		// Account exists but balance would be negative
		return ErrInsufficientFunds
	}

	return nil
}

// DebitAccount debits an amount from the account (convenience method)
func (r *AccountRepository) DebitAccount(ctx context.Context, accountID uuid.UUID, amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("debit amount must be greater than zero")
	}
	return r.UpdateBalance(ctx, accountID, amount.Neg())
}

// CreditAccount credits an amount to the account (convenience method)
func (r *AccountRepository) CreditAccount(ctx context.Context, accountID uuid.UUID, amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("credit amount must be greater than zero")
	}
	return r.UpdateBalance(ctx, accountID, amount)
}

// CreateHold creates a hold on funds
func (r *AccountRepository) CreateHold(ctx context.Context, hold *models.AccountHold) error {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check available balance
	var availableBalance decimal.Decimal
	err = tx.QueryRowContext(ctx,
		"SELECT available_balance FROM accounts WHERE id = $1 FOR UPDATE",
		hold.AccountID,
	).Scan(&availableBalance)

	if err != nil {
		return err
	}

	if availableBalance.LessThan(hold.Amount) {
		return ErrInsufficientFunds
	}

	// Create hold
	query := `
		INSERT INTO account_holds (id, account_id, amount, reason, transaction_ref, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = tx.ExecContext(ctx, query,
		hold.ID,
		hold.AccountID,
		hold.Amount,
		hold.Reason,
		hold.TransactionRef,
		hold.ExpiresAt,
		hold.CreatedAt,
	)

	if err != nil {
		return err
	}

	// Update available balance
	_, err = tx.ExecContext(ctx,
		"UPDATE accounts SET available_balance = available_balance - $1, updated_at = $2 WHERE id = $3",
		hold.Amount, time.Now(), hold.AccountID,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

// ReleaseHold releases a hold on funds
func (r *AccountRepository) ReleaseHold(ctx context.Context, holdID uuid.UUID) error {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get hold
	var hold models.AccountHold
	err = tx.QueryRowContext(ctx,
		"SELECT id, account_id, amount FROM account_holds WHERE id = $1 AND released_at IS NULL FOR UPDATE",
		holdID,
	).Scan(&hold.ID, &hold.AccountID, &hold.Amount)

	if err == sql.ErrNoRows {
		return ErrHoldNotFound
	}
	if err != nil {
		return err
	}

	// Release hold
	_, err = tx.ExecContext(ctx,
		"UPDATE account_holds SET released_at = $1 WHERE id = $2",
		time.Now(), holdID,
	)

	if err != nil {
		return err
	}

	// Update available balance
	_, err = tx.ExecContext(ctx,
		"UPDATE accounts SET available_balance = available_balance + $1, updated_at = $2 WHERE id = $3",
		hold.Amount, time.Now(), hold.AccountID,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}


// GetActiveHolds retrieves all active holds for an account
func (r *AccountRepository) GetActiveHolds(ctx context.Context, accountID uuid.UUID) ([]*models.AccountHold, error) {
	query := `
		SELECT id, account_id, amount, reason, transaction_ref, expires_at, released_at, created_at
		FROM account_holds
		WHERE account_id = $1 AND released_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	holds := []*models.AccountHold{}
	for rows.Next() {
		hold := &models.AccountHold{}
		err := rows.Scan(
			&hold.ID,
			&hold.AccountID,
			&hold.Amount,
			&hold.Reason,
			&hold.TransactionRef,
			&hold.ExpiresAt,
			&hold.ReleasedAt,
			&hold.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		holds = append(holds, hold)
	}

	return holds, nil
}

// GetHoldByID retrieves a hold by ID
func (r *AccountRepository) GetHoldByID(ctx context.Context, holdID uuid.UUID) (*models.AccountHold, error) {
	query := `
		SELECT id, account_id, amount, reason, transaction_ref, expires_at, released_at, created_at
		FROM account_holds
		WHERE id = $1
	`

	hold := &models.AccountHold{}
	err := r.db.QueryRowContext(ctx, query, holdID).Scan(
		&hold.ID,
		&hold.AccountID,
		&hold.Amount,
		&hold.Reason,
		&hold.TransactionRef,
		&hold.ExpiresAt,
		&hold.ReleasedAt,
		&hold.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrHoldNotFound
	}
	if err != nil {
		return nil, err
	}

	return hold, nil
}
