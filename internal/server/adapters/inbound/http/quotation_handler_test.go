package httpadapter

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vini0012/go-client-server-api/internal/domain"
)

type getQuotationStub struct {
	quotation domain.Quotation
	err       error
	called    bool
}

func (stub *getQuotationStub) Execute(
	_ context.Context,
) (domain.Quotation, error) {
	stub.called = true
	return stub.quotation, stub.err
}

func TestQuotationHandler_ShouldReturnQuotation(t *testing.T) {
	quotation := newValidQuotation(t)
	useCase := &getQuotationStub{quotation: quotation}
	handler := NewQuotationHandler(useCase, testLogger())

	request := httptest.NewRequest(http.MethodGet, "/cotacao", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("expected application/json, got %s", contentType)
	}

	const expectedBody = "{\"bid\":\"5.1234\"}\n"
	if response.Body.String() != expectedBody {
		t.Errorf(
			"expected body %q, got %q",
			expectedBody,
			response.Body.String(),
		)
	}

	if !useCase.called {
		t.Error("expected use case to be called")
	}
}

func TestQuotationHandler_ShouldRejectUnsupportedMethod(t *testing.T) {
	useCase := &getQuotationStub{}
	handler := NewQuotationHandler(useCase, testLogger())

	request := httptest.NewRequest(http.MethodPost, "/cotacao", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", response.Code)
	}

	if allow := response.Header().Get("Allow"); allow != http.MethodGet {
		t.Errorf("expected Allow GET, got %s", allow)
	}

	if useCase.called {
		t.Error("use case should not be called for unsupported method")
	}
}

func TestQuotationHandler_ShouldReturnGatewayTimeout(t *testing.T) {
	useCase := &getQuotationStub{err: context.DeadlineExceeded}
	handler := NewQuotationHandler(useCase, testLogger())

	request := httptest.NewRequest(http.MethodGet, "/cotacao", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected status 504, got %d", response.Code)
	}
}

func TestQuotationHandler_ShouldReturnInternalServerError(t *testing.T) {
	useCase := &getQuotationStub{err: context.Canceled}
	handler := NewQuotationHandler(useCase, testLogger())

	request := httptest.NewRequest(http.MethodGet, "/cotacao", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", response.Code)
	}
}

func newValidQuotation(t *testing.T) domain.Quotation {
	t.Helper()

	quotation, err := domain.NewQuotation(
		"USD",
		"BRL",
		"5.1234",
		1754157600,
		time.Date(2026, time.August, 16, 16, 30, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("failed to create quotation: %v", err)
	}

	return quotation
}

func testLogger() *log.Logger {
	return log.New(io.Discard, "", 0)
}
