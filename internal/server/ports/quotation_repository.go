package ports

import (
	"context"

	"github.com/vini0012/go-client-server-api/internal/domain"
)

type QuotationRepository interface {
	Save(ctx context.Context, quotation domain.Quotation) error
}
