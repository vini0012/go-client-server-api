package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	"github.com/vini0012/go-client-server-api/internal/domain"
)

var (
	ErrOpenDatabase       = errors.New("error opening sqlite database")
	ErrInitializeDatabase = errors.New("error initializing sqlite database")
	ErrSaveQuotation      = errors.New("error saving quotation")
)

type QuotationRepository struct {
	db *sql.DB
}

func Open(databasePath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOpenDatabase, err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("%w: %w", ErrOpenDatabase, err)
	}

	return db, nil
}

func NewQuotationRepository(db *sql.DB) *QuotationRepository {
	return &QuotationRepository{db: db}
}

func (repository *QuotationRepository) Migrate(ctx context.Context) error {
	const statement = `
		CREATE TABLE IF NOT EXISTS quotations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT NOT NULL,
			code_in TEXT NOT NULL,
			bid TEXT NOT NULL,
			external_timestamp INTEGER NOT NULL,
			fetched_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`

	if _, err := repository.db.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf("%w: %w", ErrInitializeDatabase, err)
	}

	return nil
}

func (repository *QuotationRepository) Save(
	ctx context.Context,
	quotation domain.Quotation,
) error {
	const statement = `
		INSERT INTO quotations (
			code,
			code_in,
			bid,
			external_timestamp,
			fetched_at
		) VALUES (?, ?, ?, ?, ?)
	`

	_, err := repository.db.ExecContext(
		ctx,
		statement,
		quotation.Code,
		quotation.CodeIn,
		quotation.Bid,
		quotation.ExternalTimestamp,
		quotation.FetchedAt,
	)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSaveQuotation, err)
	}

	return nil
}
