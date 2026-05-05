package repository

import (
	"database/sql"
	"errors"
	"fmt"

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
			return ErrDublicateID
		}

		return fmt.Errorf("failed to save url: %w", err)
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
