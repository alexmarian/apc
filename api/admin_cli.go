package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/alexmarian/apc/api/internal/auth"
	"github.com/alexmarian/apc/api/internal/database"
	"github.com/joho/godotenv"
)

// runCreateAdmin bootstraps an admin user directly in the DB, so there is
// someone able to log in and issue registration tokens via /v1/api/admin/tokens.
//
// Usage: apc create-admin -login <login> [-qr <file.png>] [-reset]
// The password is read from APC_ADMIN_PASSWORD or, if unset, from stdin.
func runCreateAdmin(args []string) error {
	fset := flag.NewFlagSet("create-admin", flag.ExitOnError)
	login := fset.String("login", "", "admin login (required)")
	qrPath := fset.String("qr", "", "where to write the TOTP QR code PNG (default <login>-totp.png)")
	reset := fset.Bool("reset", false, "if the login exists, reset its password/TOTP and make it admin")
	fset.Parse(args)

	if *login == "" {
		fset.Usage()
		return errors.New("-login is required")
	}
	if *qrPath == "" {
		*qrPath = *login + "-totp.png"
	}

	_ = godotenv.Load(".env")
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		return errors.New("DB_PATH environment variable is not set")
	}

	password := os.Getenv("APC_ADMIN_PASSWORD")
	if password == "" {
		fmt.Fprint(os.Stderr, "Password: ")
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil && line == "" {
			return fmt.Errorf("read password: %w", err)
		}
		password = strings.TrimRight(line, "\r\n")
	}
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	migFS, err := fs.Sub(schemaFS, "sql/schema")
	if err != nil {
		return fmt.Errorf("migrations fs: %w", err)
	}
	if err := database.RunMigrations(db, migFS); err != nil {
		return err
	}
	q := database.New(db)
	ctx := context.Background()

	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	secret, qrCode, err := auth.GenerateQRCode(*login)
	if err != nil {
		return err
	}

	_, err = q.GetUserByLogin(ctx, *login)
	switch {
	case err == nil && !*reset:
		return fmt.Errorf("user %q already exists (use -reset to reset it and make it admin)", *login)
	case err == nil:
		var user database.User
		user, err = q.UpdateUserEmailAndPassword(ctx, database.UpdateUserEmailAndPasswordParams{
			PasswordHash: hash,
			ToptSecret:   secret,
			IsAdmin:      true,
			Login:        *login,
		})
		// The operator scans the QR file directly, so there's no onboarding
		// step: count TOTP as confirmed and make sure the account is enabled.
		if err == nil {
			err = q.ConfirmUserTotp(ctx, user.ID)
		}
		if err == nil {
			err = q.SetUserDisabled(ctx, database.SetUserDisabledParams{ID: user.ID})
		}
	case errors.Is(err, sql.ErrNoRows):
		_, err = q.CreateUser(ctx, database.CreateUserParams{
			Login:           *login,
			PasswordHash:    hash,
			ToptSecret:      secret,
			IsAdmin:         true,
			TotpConfirmedAt: sql.NullTime{Time: time.Now().UTC(), Valid: true},
		})
	}
	if err != nil {
		return fmt.Errorf("save user: %w", err)
	}

	png, err := base64.StdEncoding.DecodeString(qrCode)
	if err != nil {
		return err
	}
	if err := os.WriteFile(*qrPath, png, 0o600); err != nil {
		return fmt.Errorf("write QR code: %w", err)
	}

	fmt.Printf("Admin user %q saved.\n", *login)
	fmt.Printf("TOTP QR code written to %s (scan it, then delete the file).\n", *qrPath)
	fmt.Printf("Manual TOTP secret: %s\n", secret)
	return nil
}
