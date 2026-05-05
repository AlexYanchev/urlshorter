package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/AlexYanchev/urlshorter/internal/model"
	"github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgres(dsn string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	repo := &PostgresRepository{db: db}

	if err := repo.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	return repo, nil
}

func (r *PostgresRepository) Save(id, value string) error {
	query := `INSERT INTO short_urls (short_id, original_url) VALUES ($1, $2)`

	_, err := r.db.Exec(query, id, value)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			if pqErr.Constraint == "short_urls_original_url_idx" {
				return ErrDuplicateOriginalURL
			}

			return ErrDublicateID
		}

		return fmt.Errorf("failed to save url: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetByOriginalURL(originalURL string) (string, bool) {
	query := `SELECT short_id FROM short_urls WHERE original_url = $1`

	var shortID string
	err := r.db.QueryRow(query, originalURL).Scan(&shortID)
	if err != nil {
		return "", false
	}

	return shortID, true
}

func (r *PostgresRepository) SaveBatch(items []model.BatchURLItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO short_urls (short_id, original_url) VALUES ($1, $2)`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch insert: %w", err)
	}
	defer stmt.Close()

	for _, item := range items {
		_, err := stmt.Exec(item.ShortURL, item.OriginalURL)
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" {
				return ErrDublicateID
			}

			return fmt.Errorf("failed to save batch urls: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *PostgresRepository) Get(id string) (string, bool) {
	query := `SELECT original_url FROM short_urls WHERE short_id = $1`

	var originalURL string
	err := r.db.QueryRow(query, id).Scan(&originalURL)
	if err != nil {
		return "", false
	}

	return originalURL, true
}

func (r *PostgresRepository) Ping() error {
	return r.db.Ping()
}

func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

func (r *PostgresRepository) DB() *sql.DB {
	return r.db
}
