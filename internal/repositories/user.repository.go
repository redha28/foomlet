package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redha28/foomlet/internal/models"
)

// Common errors
var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)

type UserRepoInterface interface {
	UseRegister(ctx context.Context, user models.UserRegist, hashedPin string) (*models.UserResponse, error)
	GetUserByPhone(ctx context.Context, phone string) (*models.User, error)
	UpdateUserProfile(ctx context.Context, userID string, profile models.UpdateProfileRequest) (*models.UpdateProfileResponse, error)
}

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (u *UserRepo) UseRegister(ctx context.Context, user models.UserRegist, hashedPin string) (*models.UserResponse, error) {
	// Begin transaction
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	// Defer rollback in case of error
	defer tx.Rollback(ctx)

	// Check if user already exists
	var existingPhone string
	queryCheckPhone := `SELECT phone FROM users WHERE phone = $1`
	err = tx.QueryRow(ctx, queryCheckPhone, user.Phone).Scan(&existingPhone)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}
	if existingPhone == user.Phone {
		return nil, ErrUserAlreadyExists
	}

	// Create new user
	var result models.UserResponse
	query := `
		INSERT INTO users (firstname, lastname, phone, address, pin)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, firstname, lastname, phone, address, created_at`

	err = tx.QueryRow(ctx, query,
		user.Firstname,
		user.Lastname,
		user.Phone,
		user.Address,
		hashedPin).Scan(
		&result.ID,
		&result.Firstname,
		&result.Lastname,
		&result.Phone,
		&result.Address,
		&result.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Create wallet for user
	queryWallet := `INSERT INTO wallets (user_id) VALUES ($1);`
	_, err = tx.Exec(ctx, queryWallet, result.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &result, nil
}

func (u *UserRepo) GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	var user models.User
	query := `SELECT id, firstname, lastname, phone, address, pin, created_at, updated_at 
			  FROM users WHERE phone = $1`

	err := u.db.QueryRow(ctx, query, phone).Scan(
		&user.ID, &user.Firstname, &user.Lastname, &user.Phone,
		&user.Address, &user.Pin, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (u *UserRepo) UpdateUserProfile(ctx context.Context, userID string, profile models.UpdateProfileRequest) (*models.UpdateProfileResponse, error) {
	query := `
		UPDATE users 
		SET firstname = $1, lastname = $2, address = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING id, updated_at`

	var updatedResponse models.UpdateProfileResponse
	err := u.db.QueryRow(ctx, query, profile.Firstname, profile.Lastname, profile.Address, userID).Scan(
		&updatedResponse.ID, &updatedResponse.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	updatedResponse.Firstname = profile.Firstname
	updatedResponse.Lastname = profile.Lastname
	updatedResponse.Address = profile.Address

	return &updatedResponse, nil
}
