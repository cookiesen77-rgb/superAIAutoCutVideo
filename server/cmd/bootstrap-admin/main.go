package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	userDB := strings.TrimSpace(os.Getenv("USER_DATABASE_URL"))
	email := strings.TrimSpace(os.Getenv("ADMIN_EMAIL"))
	password := strings.TrimSpace(os.Getenv("ADMIN_PASSWORD"))
	if userDB == "" || email == "" || password == "" {
		log.Fatal("USER_DATABASE_URL, ADMIN_EMAIL, ADMIN_PASSWORD required")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("hash error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, userDB)
	if err != nil {
		log.Fatalf("connect error: %v", err)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, `
		INSERT INTO users (email, password_hash, is_admin, status)
		VALUES ($1, $2, TRUE, 'active')
		ON CONFLICT (email)
		DO UPDATE SET password_hash=EXCLUDED.password_hash, is_admin=TRUE, status='active'
	`, strings.ToLower(email), string(hash))
	if err != nil {
		log.Fatalf("upsert failed: %v", err)
	}
	log.Println("admin user ensured")
}
