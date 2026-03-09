package repository

import (
	"context"
	"database/sql"
	"errors"

	"link-shortener/internal/model"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// ConflictError возвращается при попытке сохранить URL, который уже существует.
// ShortURL содержит уже имеющийся короткий URL для этого оригинала.
type ConflictError struct {
	ShortURL string
}

func (e *ConflictError) Error() string {
	return "url already exists"
}

// ErrURLDeleted возвращается при обращении к URL, который был удален.
var ErrURLDeleted = errors.New("url has been deleted")

// DBRepository реализует репозиторий с подключением к БД.
type DBRepository struct {
	db *sql.DB
}

func NewDBRepository(db *sql.DB) (*DBRepository, error) {
	repo := &DBRepository{db: db}

	if err := repo.bootstrap(context.Background()); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *DBRepository) bootstrap(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		short_url VARCHAR(255) UNIQUE NOT NULL,
		original_url TEXT UNIQUE NOT NULL,
		user_id TEXT NOT NULL DEFAULT '',
		is_deleted BOOLEAN NOT NULL DEFAULT FALSE
	);

	CREATE INDEX IF NOT EXISTS idx_urls_short_url ON urls(short_url);
	CREATE INDEX IF NOT EXISTS idx_urls_user_id ON urls(user_id);
	`

	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *DBRepository) Save(ctx context.Context, shortURL, originalURL, userID string) error {
	_, err := r.db.ExecContext(
		ctx,
		"INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)",
		shortURL, originalURL, userID,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			existingShortURL, getErr := r.GetByOriginalURL(ctx, originalURL)
			if getErr != nil {
				return getErr
			}
			return &ConflictError{ShortURL: existingShortURL}
		}
		return err
	}
	return nil
}

func (r *DBRepository) Get(ctx context.Context, shortURL string) (string, error) {
	var originalURL string
	var isDeleted bool
	err := r.db.QueryRowContext(
		ctx,
		"SELECT original_url, is_deleted FROM urls WHERE short_url = $1",
		shortURL,
	).Scan(&originalURL, &isDeleted)

	if err != nil {
		return "", err
	}

	if isDeleted {
		return "", ErrURLDeleted
	}

	return originalURL, nil
}

func (r *DBRepository) GetByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	var shortURL string
	err := r.db.QueryRowContext(
		ctx,
		"SELECT short_url FROM urls WHERE original_url = $1",
		originalURL,
	).Scan(&shortURL)

	if err != nil {
		return "", err
	}
	return shortURL, nil
}

func (r *DBRepository) SaveBatch(ctx context.Context, shortURLs, originalURLs []string, userID string) error {
	if len(shortURLs) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)")
	if err != nil {
		return err
	}
	defer func() { _ = stmt.Close() }()

	for i := range shortURLs {
		_, err = stmt.ExecContext(ctx, shortURLs[i], originalURLs[i], userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *DBRepository) GetURLsByUserID(ctx context.Context, userID string) ([]model.URLPair, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT short_url, original_url FROM urls WHERE user_id = $1",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var urls []model.URLPair
	for rows.Next() {
		var pair model.URLPair
		if err := rows.Scan(&pair.ShortURL, &pair.OriginalURL); err != nil {
			return nil, err
		}
		urls = append(urls, pair)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func (r *DBRepository) DeleteURLs(ctx context.Context, shortURLs []string, userID string) error {
	if len(shortURLs) == 0 {
		return nil
	}

	query := `UPDATE urls SET is_deleted = TRUE WHERE short_url = ANY($1) AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, shortURLs, userID)
	return err
}

func (r *DBRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
