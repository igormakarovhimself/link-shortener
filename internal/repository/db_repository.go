package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type ConflictError struct {
	ShortURL string
}

func (e *ConflictError) Error() string {
	return "url already exists"
}

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

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return &ConflictError{ShortURL: shortURL}
		}
		return err
	}
	return nil
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

func (r *DBRepository) GetByOriginalURL(originalURL string) (string, error) {
	var shortURL string
	err := r.db.QueryRowContext(
		context.Background(),
		"SELECT short_url FROM urls WHERE original_url = $1",
		originalURL,
	).Scan(&shortURL)

	if err != nil {
		return "", err
	}
	return shortURL, nil
}

func (r *DBRepository) SaveBatch(shortURLs, originalURLs []string) error {
	if len(shortURLs) == 0 {
		return nil
	}

	ctx := context.Background()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO urls (short_url, original_url) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range shortURLs {
		_, err = stmt.ExecContext(ctx, shortURLs[i], originalURLs[i])
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
