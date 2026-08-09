package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/vini0012/go-client-server-api/internal/domain"
	"github.com/vini0012/go-client-server-api/internal/server/ports"
)

const (
	providerTimeout   = 200 * time.Millisecond
	repositoryTimeout = 10 * time.Millisecond
)

var (
	ErrFetchingQuotation = errors.New("error fetching quotation")
	ErrSavingQuotation   = errors.New("error saving quotation")
)

type GetQuotation struct {
	rateProvider        ports.RateProvider
	quotationRepository ports.QuotationRepository
}

func NewGetQuotation(
	rateProvider ports.RateProvider,
	quotationRepository ports.QuotationRepository,
) *GetQuotation {
	return &GetQuotation{
		rateProvider:        rateProvider,
		quotationRepository: quotationRepository,
	}
}

func (useCase *GetQuotation) Execute(
	ctx context.Context,
) (domain.Quotation, error) {
	providerCtx, cancelProvider := context.WithTimeout(
		ctx,
		providerTimeout,
	)

	quotation, err := useCase.rateProvider.FetchUSDBRL(providerCtx)
	cancelProvider()

	if err != nil {
		return domain.Quotation{}, fmt.Errorf(
			"%w: %w",
			ErrFetchingQuotation,
			err,
		)
	}

	repositoryCtx, cancelRepository := context.WithTimeout(
		ctx,
		repositoryTimeout,
	)

	err = useCase.quotationRepository.Save(
		repositoryCtx,
		quotation,
	)
	cancelRepository()

	if err != nil {
		return domain.Quotation{}, fmt.Errorf(
			"%w: %w",
			ErrSavingQuotation,
			err,
		)
	}

	return quotation, nil
}
