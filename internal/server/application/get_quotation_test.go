package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vini0012/go-client-server-api/internal/domain"
)

type rateProviderStub struct {
	quotation       domain.Quotation
	err             error
	receivedTimeout time.Duration
	hasDeadline     bool
}

func (stub *rateProviderStub) FetchUSDBRL(
	ctx context.Context,
) (domain.Quotation, error) {
	deadline, ok := ctx.Deadline()
	stub.hasDeadline = ok

	if ok {
		stub.receivedTimeout = time.Until(deadline)
	}

	return stub.quotation, stub.err
}

type quotationRepositorySpy struct {
	savedQuotation  domain.Quotation
	saveCalled      bool
	err             error
	receivedTimeout time.Duration
	hasDeadline     bool
}

func (spy *quotationRepositorySpy) Save(
	ctx context.Context,
	quotation domain.Quotation,
) error {
	deadline, ok := ctx.Deadline()
	spy.hasDeadline = ok

	if ok {
		spy.receivedTimeout = time.Until(deadline)
	}

	spy.saveCalled = true
	spy.savedQuotation = quotation

	return spy.err
}

func TestGetQuotation_ShouldApplyProviderTimeout(t *testing.T) {
	quotation := newValidQuotation(t)

	provider := &rateProviderStub{
		quotation: quotation,
	}

	repository := &quotationRepositorySpy{}

	useCase := NewGetQuotation(provider, repository)

	_, err := useCase.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !provider.hasDeadline {
		t.Fatal("expected provider context to have a deadline")
	}

	if provider.receivedTimeout <= 0 {
		t.Errorf(
			"expected positive provider timeout, got %v",
			provider.receivedTimeout,
		)
	}

	if provider.receivedTimeout > providerTimeout {
		t.Errorf(
			"expected provider timeout at most %v, got %v",
			providerTimeout,
			provider.receivedTimeout,
		)
	}
}

func TestGetQuotation_ShouldApplyRepositoryTimeout(t *testing.T) {
	quotation := newValidQuotation(t)

	provider := &rateProviderStub{
		quotation: quotation,
	}

	repository := &quotationRepositorySpy{}

	useCase := NewGetQuotation(provider, repository)

	_, err := useCase.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !repository.hasDeadline {
		t.Fatal("expected repository context to have a deadline")
	}

	if repository.receivedTimeout <= 0 {
		t.Errorf(
			"expected positive repository timeout, got %v",
			repository.receivedTimeout,
		)
	}

	if repository.receivedTimeout > repositoryTimeout {
		t.Errorf(
			"expected repository timeout at most %v, got %v",
			repositoryTimeout,
			repository.receivedTimeout,
		)
	}
}

func TestGetQuotation_ShouldFetchAndSaveQuotation(t *testing.T) {
	expectedQuotation := newValidQuotation(t)

	provider := &rateProviderStub{
		quotation: expectedQuotation,
	}

	repository := &quotationRepositorySpy{}

	useCase := NewGetQuotation(provider, repository)

	result, err := useCase.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expectedQuotation {
		t.Errorf(
			"expected quotation %+v, got %+v",
			expectedQuotation,
			result,
		)
	}

	if !repository.saveCalled {
		t.Fatal("expected repository Save to be called")
	}

	if repository.savedQuotation != expectedQuotation {
		t.Errorf(
			"expected saved quotation %+v, got %+v",
			expectedQuotation,
			repository.savedQuotation,
		)
	}
}

func TestGetQuotation_ShouldReturnErrorWhenProviderFails(t *testing.T) {
	providerError := errors.New("provider unavailable")

	provider := &rateProviderStub{
		err: providerError,
	}

	repository := &quotationRepositorySpy{}

	useCase := NewGetQuotation(provider, repository)

	_, err := useCase.Execute(context.Background())

	if !errors.Is(err, ErrFetchingQuotation) {
		t.Errorf(
			"expected ErrFetchingQuotation, got %v",
			err,
		)
	}

	if !errors.Is(err, providerError) {
		t.Errorf(
			"expected provider error to be preserved, got %v",
			err,
		)
	}

	if repository.saveCalled {
		t.Error("repository should not be called when provider fails")
	}
}

func TestGetQuotation_ShouldReturnErrorWhenRepositoryFails(t *testing.T) {
	expectedQuotation := newValidQuotation(t)
	repositoryError := errors.New("repository unavailable")

	provider := &rateProviderStub{
		quotation: expectedQuotation,
	}

	repository := &quotationRepositorySpy{
		err: repositoryError,
	}

	useCase := NewGetQuotation(provider, repository)

	_, err := useCase.Execute(context.Background())

	if !errors.Is(err, ErrSavingQuotation) {
		t.Errorf(
			"expected ErrSavingQuotation, got %v",
			err,
		)
	}

	if !errors.Is(err, repositoryError) {
		t.Errorf(
			"expected repository error to be preserved, got %v",
			err,
		)
	}
}

func newValidQuotation(t *testing.T) domain.Quotation {
	t.Helper()

	quotation, err := domain.NewQuotation(
		"USD",
		"BRL",
		"5.1234",
		1754157600,
		time.Now(),
	)
	if err != nil {
		t.Fatalf("failed to create test quotation: %v", err)
	}

	return quotation
}
