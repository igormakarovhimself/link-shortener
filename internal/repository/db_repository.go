package repository

import (
	"context"
	"database/sql"
)

type DBRepository struct {
	db *sql.DB
}

func NewDBRepository(db *sql.DB) (*DBRepository, error) {
	repo := &DBRepository{db: db}

	if err := repo.bootstrap(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *DBRepository) bootstrap() error {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		short_url VARCHAR(255) UNIQUE NOT NULL,
		original_url TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_urls_short_url ON urls(short_url);
	`

	_, err := r.db.ExecContext(context.Background(), query)
	return err
}

func (r *DBRepository) Save(shortURL, originalURL string) error {
	_, err := r.db.ExecContext(
		context.Background(),
		"INSERT INTO urls (short_url, original_url) VALUES ($1, $2)",
		shortURL, originalURL,
	)
	return err
}

func (r *DBRepository) Get(shortURL string) (string, error) {
	var originalURL string
	err := r.db.QueryRowContext(
		context.Background(),
		"SELECT original_url FROM urls WHERE short_url = $1",
		shortURL,
	).Scan(&originalURL)

	if err != nil {
		return "", err
	}
	return originalURL, nil
}
