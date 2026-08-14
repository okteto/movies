package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	host     = "postgresql"
	port     = 5432
	user     = "okteto"
	password = "okteto"
	dbname   = "rentals"
)

var schema = []string{
	`CREATE TABLE IF NOT EXISTS users (
		email TEXT PRIMARY KEY,
		display_name TEXT NOT NULL DEFAULT '',
		banned BOOLEAN NOT NULL DEFAULT FALSE,
		ban_reason TEXT NOT NULL DEFAULT '',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`,
	`CREATE TABLE IF NOT EXISTS rentals (
		id SERIAL PRIMARY KEY,
		user_email TEXT NOT NULL,
		movie_id TEXT NOT NULL,
		price NUMERIC(10,2) NOT NULL DEFAULT 0,
		rented_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		returned_at TIMESTAMPTZ
	)`,
	`CREATE INDEX IF NOT EXISTS rentals_user_email_idx ON rentals (user_email)`,
	`CREATE INDEX IF NOT EXISTS rentals_movie_id_idx ON rentals (movie_id)`,
	`CREATE TABLE IF NOT EXISTS redemptions (
		id SERIAL PRIMARY KEY,
		user_email TEXT NOT NULL,
		good_deed TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`,
}

func Open() *sql.DB {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	for {
		db, err := sql.Open("postgres", psqlconn)
		if err == nil {
			return db
		}
	}
}

func Ping(db *sql.DB) {
	fmt.Println("Waiting for postgresql...")
	for {
		if err := db.Ping(); err == nil {
			fmt.Println("Postgresql connected!")
			return
		}
	}
}

// EnsureSchema creates the tables used by the movies app if they don't exist.
// Rentals are never dropped, so the rental history survives restarts.
func EnsureSchema(db *sql.DB) error {
	if err := migrateLegacyRentals(db); err != nil {
		return err
	}

	for _, stmt := range schema {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to apply %q: %w", stmt, err)
		}
	}

	return nil
}

// migrateLegacyRentals drops the single-user rentals table from previous
// versions of the app, which had no user column and one row per movie.
func migrateLegacyRentals(db *sql.DB) error {
	var legacy bool
	query := `SELECT EXISTS (
		SELECT 1 FROM information_schema.tables WHERE table_name = 'rentals'
	) AND NOT EXISTS (
		SELECT 1 FROM information_schema.columns WHERE table_name = 'rentals' AND column_name = 'user_email'
	)`
	if err := db.QueryRow(query).Scan(&legacy); err != nil {
		return err
	}

	if !legacy {
		return nil
	}

	fmt.Println("Dropping legacy single-user rentals table...")
	_, err := db.Exec(`DROP TABLE rentals`)
	return err
}
