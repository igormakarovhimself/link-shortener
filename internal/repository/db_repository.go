package repository

import (
	"context"
	"database/sql"
)

type DBRepository struct {
	db *sql.DB
}

func NewDBRepository(db *sql.DB) *DBRepository {
	return &DBRepository{db: db}
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
