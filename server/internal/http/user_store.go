package httpapi

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRecord struct {
	ID           string
	Email        string
	PasswordHash string
	IsAdmin      bool
	Status       string
	CreatedAt    time.Time
}

func createUser(ctx context.Context, db *pgxpool.Pool, email, passwordHash string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var id string
	err := db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, status, is_admin)
		VALUES ($1, $2, 'active', FALSE)
		RETURNING id
	`, email, passwordHash).Scan(&id)
	return id, err
}

func getUserByEmail(ctx context.Context, db *pgxpool.Pool, email string) (userRecord, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var user userRecord
	err := db.QueryRow(ctx, `
		SELECT id, email, password_hash, is_admin, status, created_at
		FROM users
		WHERE email=$1
	`, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.IsAdmin, &user.Status, &user.CreatedAt)
	return user, err
}

func getUserByID(ctx context.Context, db *pgxpool.Pool, userID string) (userRecord, error) {
	var user userRecord
	err := db.QueryRow(ctx, `
		SELECT id, email, password_hash, is_admin, status, created_at
		FROM users
		WHERE id=$1
	`, userID).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.IsAdmin, &user.Status, &user.CreatedAt)
	return user, err
}

func listUsers(ctx context.Context, db *pgxpool.Pool, limit, offset int) ([]userRecord, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := db.Query(ctx, `
		SELECT id, email, password_hash, is_admin, status, created_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]userRecord, 0)
	for rows.Next() {
		var user userRecord
		if err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.IsAdmin, &user.Status, &user.CreatedAt); err == nil {
			users = append(users, user)
		}
	}
	return users, nil
}

func updateUser(ctx context.Context, db *pgxpool.Pool, userID string, email *string, isAdmin *bool, status *string) (userRecord, error) {
	if email != nil {
		normalized := strings.ToLower(strings.TrimSpace(*email))
		if normalized == "" {
			return userRecord{}, errors.New("email empty")
		}
		email = &normalized
	}
	if status != nil {
		normalized := strings.ToLower(strings.TrimSpace(*status))
		if normalized == "" {
			return userRecord{}, errors.New("status empty")
		}
		status = &normalized
	}
	_, err := db.Exec(ctx, `
		UPDATE users
		SET email=COALESCE($2, email),
		    is_admin=COALESCE($3, is_admin),
		    status=COALESCE($4, status)
		WHERE id=$1
	`, userID, email, isAdmin, status)
	if err != nil {
		return userRecord{}, err
	}
	return getUserByID(ctx, db, userID)
}

func updateUserPassword(ctx context.Context, db *pgxpool.Pool, userID string, passwordHash string) error {
	_, err := db.Exec(ctx, `
		UPDATE users SET password_hash=$2 WHERE id=$1
	`, userID, passwordHash)
	return err
}

func ensureUserExists(ctx context.Context, db *pgxpool.Pool, userID string) error {
	_, err := getUserByID(ctx, db, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return errors.New("user not found")
	}
	return err
}
