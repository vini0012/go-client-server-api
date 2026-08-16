package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/vini0012/go-client-server-api/internal/domain"
)

func TestQuotationRepository_ShouldSaveQuotation(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "quotations.db")

	db, err := Open(databasePath)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	defer db.Close()

	repository := NewQuotationRepository(db)

	if err := repository.Migrate(context.Background()); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	fetchedAt := time.Date(
		2026,
		time.August,
		16,
		16,
		30,
		0,
		0,
		time.UTC,
	)

	quotation, err := domain.NewQuotation(
		"USD",
		"BRL",
		"5.1234",
		1754157600,
		fetchedAt,
	)
	if err != nil {
		t.Fatalf("failed to create quotation: %v", err)
	}

	if err := repository.Save(context.Background(), quotation); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var (
		code              string
		codeIn            string
		bid               string
		externalTimestamp int64
		storedFetchedAt   time.Time
	)

	err = db.QueryRow(`
		SELECT
			code,
			code_in,
			bid,
			external_timestamp,
			fetched_at
		FROM quotations
		LIMIT 1
	`).Scan(
		&code,
		&codeIn,
		&bid,
		&externalTimestamp,
		&storedFetchedAt,
	)
	if err != nil {
		t.Fatalf("failed to query saved quotation: %v", err)
	}

	if code != quotation.Code {
		t.Errorf("expected code %s, got %s", quotation.Code, code)
	}

	if codeIn != quotation.CodeIn {
		t.Errorf("expected codeIn %s, got %s", quotation.CodeIn, codeIn)
	}

	if bid != quotation.Bid {
		t.Errorf("expected bid %s, got %s", quotation.Bid, bid)
	}

	if externalTimestamp != quotation.ExternalTimestamp {
		t.Errorf(
			"expected timestamp %d, got %d",
			quotation.ExternalTimestamp,
			externalTimestamp,
		)
	}

	if !storedFetchedAt.Equal(quotation.FetchedAt) {
		t.Errorf(
			"expected fetchedAt %v, got %v",
			quotation.FetchedAt,
			storedFetchedAt,
		)
	}
}

func TestQuotationRepository_ShouldRespectCanceledContext(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "quotations.db")

	db, err := Open(databasePath)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	defer db.Close()

	repository := NewQuotationRepository(db)

	if err := repository.Migrate(context.Background()); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	quotation, err := domain.NewQuotation(
		"USD",
		"BRL",
		"5.1234",
		1754157600,
		time.Now(),
	)
	if err != nil {
		t.Fatalf("failed to create quotation: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = repository.Save(ctx, quotation)

	if !errors.Is(err, ErrSaveQuotation) {
		t.Fatalf("expected ErrSaveQuotation, got %v", err)
	}

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}
