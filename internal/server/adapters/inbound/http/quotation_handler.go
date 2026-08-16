package httpadapter

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/vini0012/go-client-server-api/internal/domain"
)

type GetQuotation interface {
	Execute(ctx context.Context) (domain.Quotation, error)
}

type QuotationHandler struct {
	getQuotation GetQuotation
	logger       *log.Logger
}

func NewQuotationHandler(
	getQuotation GetQuotation,
	logger *log.Logger,
) *QuotationHandler {
	return &QuotationHandler{
		getQuotation: getQuotation,
		logger:       logger,
	}
}

func (handler *QuotationHandler) ServeHTTP(
	writer http.ResponseWriter,
	request *http.Request,
) {
	if request.Method != http.MethodGet {
		writer.Header().Set("Allow", http.MethodGet)
		http.Error(
			writer,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	quotation, err := handler.getQuotation.Execute(request.Context())
	if err != nil {
		handler.logger.Printf("failed to get quotation: %v", err)

		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(
				writer,
				"quotation processing timed out",
				http.StatusGatewayTimeout,
			)
			return
		}

		http.Error(
			writer,
			"failed to get quotation",
			http.StatusInternalServerError,
		)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	response := quotationResponse{
		Bid: quotation.Bid,
	}

	if err := json.NewEncoder(writer).Encode(response); err != nil {
		handler.logger.Printf("failed to encode quotation response: %v", err)
	}
}

type quotationResponse struct {
	Bid string `json:"bid"`
}
