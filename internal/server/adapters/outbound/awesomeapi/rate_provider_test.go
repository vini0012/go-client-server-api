package awesomeapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateProvider_ShouldFetchUSDBRL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != http.MethodGet {
			t.Errorf("expected method GET, got %s", request.Method)
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{
			"USDBRL": {
				"code": "USD",
				"codein": "BRL",
				"bid": "5.1234",
				"timestamp": "1754157600"
			}
		}`))
	}))
	defer server.Close()

	fetchedAt := time.Date(
		2026,
		time.August,
		9,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	provider := NewRateProvider(server.Client(), server.URL)
	provider.now = func() time.Time {
		return fetchedAt
	}

	quotation, err := provider.FetchUSDBRL(context.Background())
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
			"expected timestamp 1754157600, got %d",
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

func TestRateProvider_ShouldReturnErrorWhenStatusIsNotOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		http.Error(writer, "service unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	provider := NewRateProvider(server.Client(), server.URL)

	_, err := provider.FetchUSDBRL(context.Background())
	if !errors.Is(err, ErrUnexpectedStatus) {
		t.Fatalf("expected ErrUnexpectedStatus, got %v", err)
	}
}

func TestRateProvider_ShouldReturnErrorWhenResponseIsInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`invalid-json`))
	}))
	defer server.Close()

	provider := NewRateProvider(server.Client(), server.URL)

	_, err := provider.FetchUSDBRL(context.Background())
	if !errors.Is(err, ErrDecodeResponse) {
		t.Fatalf("expected ErrDecodeResponse, got %v", err)
	}
}

func TestRateProvider_ShouldReturnErrorWhenTimestampIsInvalid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{
			"USDBRL": {
				"code": "USD",
				"codein": "BRL",
				"bid": "5.1234",
				"timestamp": "invalid"
			}
		}`))
	}))
	defer server.Close()

	provider := NewRateProvider(server.Client(), server.URL)

	_, err := provider.FetchUSDBRL(context.Background())
	if !errors.Is(err, ErrInvalidTimestamp) {
		t.Fatalf("expected ErrInvalidTimestamp, got %v", err)
	}
}

func TestRateProvider_ShouldRespectContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		_ http.ResponseWriter,
		request *http.Request,
	) {
		<-request.Context().Done()
	}))
	defer server.Close()

	provider := NewRateProvider(server.Client(), server.URL)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		20*time.Millisecond,
	)
	defer cancel()

	_, err := provider.FetchUSDBRL(ctx)
	if !errors.Is(err, ErrRequestQuotation) {
		t.Fatalf("expected ErrRequestQuotation, got %v", err)
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded, got %v", err)
	}
}
