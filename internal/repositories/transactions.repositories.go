package repositories

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redha28/foomlet/internal/models"
)

var (
	ErrInsufficientBalance = errors.New("saldo tidak cukup")
	ErrWalletNotFound      = errors.New("wallet not found")
)

type TransactionRepoInterface interface {
	TopUp(ctx context.Context, userID string, amount float64) (*models.TopUpResponse, error)
	Payment(ctx context.Context, userID string, amount float64, remarks string) (*models.PaymentResponse, error)
	GetUserTransactions(ctx context.Context, userID string) ([]models.TransactionResponse, error)
	GetWalletByUserID(ctx context.Context, userID string) (string, float64, error)
	Transfer(ctx context.Context, senderID, recipientID string, amount float64, remarks string) (*models.TransferResponse, error)
	ProcessTransfer(ctx context.Context, transferID, senderID, recipientID string, amount float64, remarks string) error
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
}

type TransactionRepo struct {
	db *pgxpool.Pool
}

func NewTransactionRepo(db *pgxpool.Pool) *TransactionRepo {
	return &TransactionRepo{db: db}
}

func (t *TransactionRepo) GetWalletByUserID(ctx context.Context, userID string) (string, float64, error) {
	var walletID string
	var balance float64

	// Use numeric cast to properly convert MONEY to a numeric value Go can handle
	query := `SELECT id, balance::numeric FROM wallets WHERE user_id = $1`
	err := t.db.QueryRow(ctx, query, userID).Scan(&walletID, &balance)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", 0, ErrWalletNotFound
		}
		return "", 0, err
	}

	return walletID, balance, nil
}

func (t *TransactionRepo) TopUp(ctx context.Context, userID string, amount float64) (*models.TopUpResponse, error) {
	// Begin transaction
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Get wallet ID and current balance
	walletID, balance, err := t.GetWalletByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Calculate new balance
	balanceBefore := balance
	balanceAfter := balance + amount

	// Create transaction record with the top-up transaction type (ID 1)
	txID := models.NewTransaction().ID
	txQuery := `
		INSERT INTO transactions (id, wallet_id, transaction_type_id, amount, balance_before, balance_after)
		VALUES ($1, $2, $3, $4::money, $5::money, $6::money)`

	_, err = tx.Exec(ctx, txQuery, txID, walletID, models.TransactionTypeTopUp, amount, balanceBefore, balanceAfter)
	if err != nil {
		return nil, err
	}

	// Update wallet balance - cast amount to money type for proper calculation with PostgreSQL MONEY type
	updateQuery := `
		UPDATE wallets 
		SET balance = balance + $1::money, updated_at = NOW()
		WHERE id = $2`

	_, err = tx.Exec(ctx, updateQuery, amount, walletID)
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	// Return response
	response := &models.TopUpResponse{
		ID:            txID,
		Amount:        amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		CreatedAt:     models.NewTransaction().CreatedAt,
	}

	return response, nil
}

func (t *TransactionRepo) Payment(ctx context.Context, userID string, amount float64, remarks string) (*models.PaymentResponse, error) {
	// Begin transaction
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Get wallet ID and current balance
	walletID, balance, err := t.GetWalletByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check if balance is sufficient
	if balance < amount {
		return nil, ErrInsufficientBalance
	}

	// Calculate new balance
	balanceBefore := balance
	balanceAfter := balance - amount

	// Create transaction - cast to money type for PostgreSQL
	txID := models.NewTransaction().ID
	txQuery := `
		INSERT INTO transactions (id, wallet_id, transaction_type_id, amount, balance_before, balance_after)
		VALUES ($1, $2, $3, $4::money, $5::money, $6::money)`

	_, err = tx.Exec(ctx, txQuery, txID, walletID, models.TransactionTypePayment, amount, balanceBefore, balanceAfter)
	if err != nil {
		return nil, err
	}

	// Create payment record
	paymentQuery := `
		INSERT INTO payments (transaction_id, user_id, amount, remarks)
		VALUES ($1, $2, $3, $4)`

	// Use int value for amount in payments table if it's defined as INT in the schema
	_, err = tx.Exec(ctx, paymentQuery, txID, userID, int(amount), remarks)
	if err != nil {
		return nil, err
	}

	// Update wallet balance
	updateQuery := `
		UPDATE wallets 
		SET balance = balance - $1::money, updated_at = NOW()
		WHERE id = $2`

	_, err = tx.Exec(ctx, updateQuery, amount, walletID)
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	// Return response
	response := &models.PaymentResponse{
		ID:            txID,
		Amount:        amount,
		Remarks:       remarks,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		CreatedAt:     models.NewTransaction().CreatedAt,
	}

	return response, nil
}

func (t *TransactionRepo) GetUserTransactions(ctx context.Context, userID string) ([]models.TransactionResponse, error) {
	// Get wallet ID
	walletID, _, err := t.GetWalletByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Query to get all transactions with their balance history
	// Use transaction type names directly from transaction_types table
	query := `
		SELECT 
			t.id,
			'SUCCESS' as status,
			$2 as user_id,
			CASE 
				WHEN t.transaction_type_id = 1 THEN 'Top-Up'
				WHEN t.transaction_type_id = 2 THEN 'Payment'
				WHEN t.transaction_type_id = 3 THEN 'Transfer'
				ELSE 'Unknown'
			END as transaction_type,
			t.amount::numeric,
			CASE 
				WHEN t.transaction_type_id = 2 THEN p.remarks
				WHEN t.transaction_type_id = 3 THEN tr.remarks
				ELSE ''
			END as remarks,
			t.balance_before::numeric,
			t.balance_after::numeric,
			t.created_at
		FROM 
			transactions t
			JOIN transaction_types tt ON t.transaction_type_id = tt.id
			LEFT JOIN payments p ON t.id = p.transaction_id
			LEFT JOIN transfer tr ON t.id = tr.transaction_id
		WHERE 
			t.wallet_id = $1
		ORDER BY 
			t.created_at DESC
	`

	rows, err := t.db.Query(ctx, query, walletID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.TransactionResponse
	for rows.Next() {
		var tx models.TransactionResponse
		if err := rows.Scan(
			&tx.ID,
			&tx.Status,
			&tx.Userid,
			&tx.TransactionType,
			&tx.Amount,
			&tx.Remarks,
			&tx.BalanceBefore,
			&tx.BalanceAfter,
			&tx.CreatedAt,
		); err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (t *TransactionRepo) Transfer(ctx context.Context, senderID, recipientID string, amount float64, remarks string) (*models.TransferResponse, error) {
	// Begin transaction
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Get sender wallet ID and current balance
	senderWalletID, senderBalance, err := t.GetWalletByUserID(ctx, senderID)
	if err != nil {
		return nil, err
	}

	// Check if balance is sufficient
	if senderBalance < amount {
		return nil, ErrInsufficientBalance
	}

	// Calculate new balance
	balanceBefore := senderBalance
	balanceAfter := senderBalance - amount

	// Check if recipient exists
	if _, err := t.GetUserByID(ctx, recipientID); err != nil {
		return nil, err
	}

	// Create transaction record with the transfer transaction type
	txID := models.NewTransaction().ID
	txQuery := `
		INSERT INTO transactions (id, wallet_id, transaction_type_id, amount, balance_before, balance_after)
		VALUES ($1, $2, $3, $4::money, $5::money, $6::money)`

	_, err = tx.Exec(ctx, txQuery, txID, senderWalletID, models.TransactionTypeTransfer, amount, balanceBefore, balanceAfter)
	if err != nil {
		return nil, err
	}

	// Create transfer record
	transferQuery := `
		INSERT INTO transfer (transaction_id, target_user, sender_user, remarks)
		VALUES ($1, $2, $3, $4)`

	_, err = tx.Exec(ctx, transferQuery, txID, recipientID, senderID, remarks)
	if err != nil {
		return nil, err
	}

	// Important: We don't update wallet balances here - that's done by ProcessTransfer
	// This function just creates the transaction records

	// Commit transaction to save the transaction record
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	// Return response
	response := &models.TransferResponse{
		ID:            txID,
		Amount:        amount,
		Remarks:       remarks,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		CreatedAt:     models.NewTransaction().CreatedAt,
	}

	return response, nil
}

func (t *TransactionRepo) ProcessTransfer(ctx context.Context, transferID, senderID, recipientID string, amount float64, remarks string) error {
	// Begin transaction
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	log.Printf("Processing transfer: ID=%s, Sender=%s, Recipient=%s, Amount=%.2f",
		transferID, senderID, recipientID, amount)

	// Update sender's wallet (deduct amount)
	updateSenderQuery := `
		UPDATE wallets 
		SET balance = balance - $1::money, updated_at = NOW()
		WHERE user_id = $2`

	senderResult, err := tx.Exec(ctx, updateSenderQuery, amount, senderID)
	if err != nil {
		log.Printf("Error updating sender wallet: %v", err)
		return err
	}

	// Check if sender's wallet was updated
	rowsAffected := senderResult.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("No sender wallet updated for user ID: %s", senderID)
		return errors.New("no sender wallet updated, possible invalid user ID")
	}

	log.Printf("Updated sender wallet for user ID: %s", senderID)

	// Get recipient wallet ID
	recipientWalletID, recipientBalance, err := t.GetWalletByUserID(ctx, recipientID)
	if err != nil {
		log.Printf("Error getting recipient wallet: %v", err)
		return err
	}

	log.Printf("Found recipient wallet: %s with balance %.2f", recipientWalletID, recipientBalance)

	// Create a credit transaction for the recipient
	recipientTxID := models.NewTransaction().ID
	recipientTxQuery := `
		INSERT INTO transactions (id, wallet_id, transaction_type_id, amount, balance_before, balance_after)
		VALUES ($1, $2, $3, $4::money, $5::money, $6::money)`

	_, err = tx.Exec(ctx, recipientTxQuery, recipientTxID, recipientWalletID, models.TransactionTypeTransfer,
		amount, recipientBalance, recipientBalance+amount)
	if err != nil {
		log.Printf("Error creating recipient transaction: %v", err)
		return err
	}

	// Update recipient's wallet (add amount)
	updateRecipientQuery := `
		UPDATE wallets 
		SET balance = balance + $1::money, updated_at = NOW()
		WHERE user_id = $2`

	recipientResult, err := tx.Exec(ctx, updateRecipientQuery, amount, recipientID)
	if err != nil {
		log.Printf("Error updating recipient wallet: %v", err)
		return err
	}

	// Check if recipient's wallet was updated
	rowsAffected = recipientResult.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("No recipient wallet updated for user ID: %s", recipientID)
		return errors.New("no recipient wallet updated, possible invalid user ID")
	}

	log.Printf("Updated recipient wallet for user ID: %s", recipientID)

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	log.Printf("Transfer completed successfully: %s", transferID)
	return nil
}

func (t *TransactionRepo) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	query := `SELECT id, firstname, lastname, phone, address, created_at, updated_at 
			  FROM users WHERE id = $1`

	err := t.db.QueryRow(ctx, query, userID).Scan(
		&user.ID, &user.Firstname, &user.Lastname, &user.Phone,
		&user.Address, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}
