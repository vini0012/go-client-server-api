package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewQuotation_ShouldCreateValidQuotation(t *testing.T) {
	fetchedAt := time.Date(
		2026,
		time.August,
		2,
		18,
		0,
		0,
		0,
		time.UTC,
	)

	quotation, err := NewQuotation(
		"usd",
		" brl ",
		"5.1234",
		1754157600,
		fetchedAt,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if quotation.Code != "USD" {
		t.Errorf("expected code USD, got %s", quotation.Code)
	}

	if quotation.CodeIn != "BRL" {
		t.Errorf("expected codeIn BRL, got %s", quotation.CodeIn)
	}

	if quotation.Bid != "5.1234" {
		t.Errorf("expected bid 5.1234, got %s", quotation.Bid)
	}

	if quotation.ExternalTimestamp != 1754157600 {
		t.Errorf(
			"expected external timestamp 1754157600, got %d",
			quotation.ExternalTimestamp,
		)
	}

	if !quotation.FetchedAt.Equal(fetchedAt) {
		t.Errorf(
			"expected fetchedAt %v, got %v",
			fetchedAt,
			quotation.FetchedAt,
		)
	}
}

func TestNewQuotation_ShouldRejectInvalidQuotation(t *testing.T) {
	validFetchedAt := time.Now()

	tests := []struct {
		name              string
		code              string
		codeIn            string
		bid               string
		externalTimestamp int64
		fetchedAt         time.Time
	}{
		{
			name:              "empty code",
			code:              "",
			codeIn:            "BRL",
			bid:               "5.1234",
			externalTimestamp: 1754157600,
			fetchedAt:         validFetchedAt,
		},
		{
			name:              "empty codeIn",
			code:              "USD",
			codeIn:            "",
			bid:               "5.1234",
			externalTimestamp: 1754157600,
			fetchedAt:         validFetchedAt,
		},
		{
			name:              "empty bid",
			code:              "USD",
			codeIn:            "BRL",
			bid:               "",
			externalTimestamp: 1754157600,
			fetchedAt:         validFetchedAt,
		},
		{
			name:              "non numeric bid",
			code:              "USD",
			codeIn:            "BRL",
			bid:               "invalid",
			externalTimestamp: 1754157600,
			fetchedAt:         validFetchedAt,
		},
		{
			name:              "zero bid",
			code:              "USD",
			codeIn:            "BRL",
			bid:               "0",
			externalTimestamp: 1754157600,
			fetchedAt:         validFetchedAt,
		},
		{
			name:              "invalid external timestamp",
			code:              "USD",
			codeIn:            "BRL",
			bid:               "5.1234",
			externalTimestamp: 0,
			fetchedAt:         validFetchedAt,
		},
		{
			name:              "empty fetchedAt",
			code:              "USD",
			codeIn:            "BRL",
			bid:               "5.1234",
			externalTimestamp: 1754157600,
			fetchedAt:         time.Time{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewQuotation(
				test.code,
				test.codeIn,
				test.bid,
				test.externalTimestamp,
				test.fetchedAt,
			)

			if !errors.Is(err, ErrInvalidQuotation) {
				t.Errorf(
					"expected ErrInvalidQuotation, got %v",
					err,
				)
			}
		})
	}
}
