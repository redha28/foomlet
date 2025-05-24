package seed

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redha28/foomlet/internal/models"
	"github.com/redha28/foomlet/pkg"
)

func SeedInitialData(ctx context.Context, db *pgxpool.Pool) error {
	log.Println("Starting initial data seeding...")

	// Create hash configuration
	hash := pkg.InitHashConfig()
	hash.UseDefaultConfig()

	// Hash PIN for both users (PIN: 123456)
	hashedPin, err := hash.GenHashedPassword("123456")
	if err != nil {
		return err
	}

	// Begin transaction for entire seeding process
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Fixed User IDs
	user1ID := "d73c798e-4363-4f8f-b57c-7a548c885bcb"
	user2ID := "faec4966-1ccd-43de-88ef-b22e9388665f"

	// 1. Create User 1 (08123456789)
	user1Query := `
		INSERT INTO users (id, firstname, lastname, phone, address, pin)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = tx.Exec(ctx, user1Query, user1ID, "John", "Doe", "08123456789", "Jakarta Selatan", hashedPin)
	if err != nil {
		log.Printf("Error creating user 1: %v", err)
		return err
	}
	log.Printf("Created user 1: %s (08123456789)", user1ID)

	// Create wallet for User 1
	wallet1ID := models.NewTransaction().ID
	wallet1Query := `INSERT INTO wallets (id, user_id, balance) VALUES ($1, $2, $3::money)`
	_, err = tx.Exec(ctx, wallet1Query, wallet1ID, user1ID, 0)
	if err != nil {
		log.Printf("Error creating wallet for user 1: %v", err)
		return err
	}
	log.Printf("Created wallet for user 1: %s", wallet1ID)

	// 2. Create User 2 (08987654321)
	user2Query := `
		INSERT INTO users (id, firstname, lastname, phone, address, pin)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = tx.Exec(ctx, user2Query, user2ID, "Jane", "Smith", "08987654321", "Jakarta Utara", hashedPin)
	if err != nil {
		log.Printf("Error creating user 2: %v", err)
		return err
	}
	log.Printf("Created user 2: %s (08987654321)", user2ID)

	// Create wallet for User 2
	wallet2ID := models.NewTransaction().ID
	wallet2Query := `INSERT INTO wallets (id, user_id, balance) VALUES ($1, $2, $3::money)`
	_, err = tx.Exec(ctx, wallet2Query, wallet2ID, user2ID, 0)
	if err != nil {
		log.Printf("Error creating wallet for user 2: %v", err)
		return err
	}
	log.Printf("Created wallet for user 2: %s", wallet2ID)

	// 3. User 1 Top-Up Transaction (500,000)
	topupAmount := 500000.0
	topupTxID := models.NewTransaction().ID
	topupQuery := `
		INSERT INTO transactions (id, wallet_id, transaction_type_id, amount, balance_before, balance_after)
		VALUES ($1, $2, $3, $4::money, $5::money, $6::money)`

	_, err = tx.Exec(ctx, topupQuery, topupTxID, wallet1ID, models.TransactionTypeTopUp,
		topupAmount, 0, topupAmount)
	if err != nil {
		log.Printf("Error creating top-up transaction: %v", err)
		return err
	}

	// Update User 1 wallet balance after top-up
	updateWallet1Query := `UPDATE wallets SET balance = $1::money, updated_at = NOW() WHERE id = $2`
	_, err = tx.Exec(ctx, updateWallet1Query, topupAmount, wallet1ID)
	if err != nil {
		log.Printf("Error updating wallet 1 balance: %v", err)
		return err
	}
	log.Printf("Created top-up transaction: %s (Amount: %.2f)", topupTxID, topupAmount)

	// 4. User 1 Payment Transaction (50,000)
	paymentAmount := 50000.0
	balanceAfterTopup := topupAmount
	balanceAfterPayment := balanceAfterTopup - paymentAmount

	paymentTxID := models.NewTransaction().ID
	paymentQuery := `
		INSERT INTO transactions (id, wallet_id, transaction_type_id, amount, balance_before, balance_after)
		VALUES ($1, $2, $3, $4::money, $5::money, $6::money)`

	_, err = tx.Exec(ctx, paymentQuery, paymentTxID, wallet1ID, models.TransactionTypePayment,
		paymentAmount, balanceAfterTopup, balanceAfterPayment)
	if err != nil {
		log.Printf("Error creating payment transaction: %v", err)
		return err
	}

	// Create payment record
	paymentRecordQuery := `
		INSERT INTO payments (transaction_id, user_id, amount, remarks)
		VALUES ($1, $2, $3, $4)`

	_, err = tx.Exec(ctx, paymentRecordQuery, paymentTxID, user1ID, int(paymentAmount), "Bayar listrik bulanan")
	if err != nil {
		log.Printf("Error creating payment record: %v", err)
		return err
	}

	// Update User 1 wallet balance after payment
	updateWallet1AfterPaymentQuery := `UPDATE wallets SET balance = $1::money, updated_at = NOW() WHERE id = $2`
	_, err = tx.Exec(ctx, updateWallet1AfterPaymentQuery, balanceAfterPayment, wallet1ID)
	if err != nil {
		log.Printf("Error updating wallet 1 balance after payment: %v", err)
		return err
	}
	log.Printf("Created payment transaction: %s (Amount: %.2f)", paymentTxID, paymentAmount)

	// 5. User 1 Transfer to User 2 (100,000)
	transferAmount := 100000.0
	balanceBeforeTransfer := balanceAfterPayment
	balanceAfterTransfer := balanceBeforeTransfer - transferAmount

	transferTxID := models.NewTransaction().ID
	transferQuery := `
		INSERT INTO transactions (id, wallet_id, transaction_type_id, amount, balance_before, balance_after)
		VALUES ($1, $2, $3, $4::money, $5::money, $6::money)`

	_, err = tx.Exec(ctx, transferQuery, transferTxID, wallet1ID, models.TransactionTypeTransfer,
		transferAmount, balanceBeforeTransfer, balanceAfterTransfer)
	if err != nil {
		log.Printf("Error creating transfer transaction: %v", err)
		return err
	}

	// Create transfer record
	transferRecordQuery := `
		INSERT INTO transfer (transaction_id, target_user, sender_user, remarks)
		VALUES ($1, $2, $3, $4)`

	_, err = tx.Exec(ctx, transferRecordQuery, transferTxID, user2ID, user1ID, "Hadiah Ultah")
	if err != nil {
		log.Printf("Error creating transfer record: %v", err)
		return err
	}

	// Update User 1 wallet balance after transfer (deduct)
	updateWallet1AfterTransferQuery := `UPDATE wallets SET balance = $1::money, updated_at = NOW() WHERE id = $2`
	_, err = tx.Exec(ctx, updateWallet1AfterTransferQuery, balanceAfterTransfer, wallet1ID)
	if err != nil {
		log.Printf("Error updating wallet 1 balance after transfer: %v", err)
		return err
	}

	// Create recipient transaction for User 2
	recipientTxID := models.NewTransaction().ID
	recipientTxQuery := `
		INSERT INTO transactions (id, wallet_id, transaction_type_id, amount, balance_before, balance_after)
		VALUES ($1, $2, $3, $4::money, $5::money, $6::money)`

	_, err = tx.Exec(ctx, recipientTxQuery, recipientTxID, wallet2ID, models.TransactionTypeTransfer,
		transferAmount, 0, transferAmount)
	if err != nil {
		log.Printf("Error creating recipient transaction: %v", err)
		return err
	}

	// Update User 2 wallet balance after receiving transfer
	updateWallet2Query := `UPDATE wallets SET balance = $1::money, updated_at = NOW() WHERE id = $2`
	_, err = tx.Exec(ctx, updateWallet2Query, transferAmount, wallet2ID)
	if err != nil {
		log.Printf("Error updating wallet 2 balance: %v", err)
		return err
	}

	log.Printf("Created transfer transaction: %s (Amount: %.2f)", transferTxID, transferAmount)

	// Commit all changes
	if err = tx.Commit(ctx); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	log.Println("Initial data seeding completed successfully!")
	log.Printf("User 1 ID: %s (08123456789) final balance: %.2f", user1ID, balanceAfterTransfer)
	log.Printf("User 2 ID: %s (08987654321) final balance: %.2f", user2ID, transferAmount)
	log.Println("PIN for both users: 123456")

	return nil
}
