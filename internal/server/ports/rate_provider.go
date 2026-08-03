package ports

import (
	"context"

	"github.com/vini0012/go-client-server-api/internal/domain"
)

type RateProvider interface {
	FetchUSDBRL(ctx context.Context) (domain.Quotation, error)
}
