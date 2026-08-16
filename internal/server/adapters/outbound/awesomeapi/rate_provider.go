package awesomeapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/vini0012/go-client-server-api/internal/domain"
)

const DefaultEndpoint = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

var (
	ErrRequestQuotation = errors.New("error requesting quotation")
	ErrUnexpectedStatus = errors.New("unexpected provider status")
	ErrDecodeResponse   = errors.New("error decoding provider response")
	ErrInvalidTimestamp = errors.New("invalid provider timestamp")
)

type RateProvider struct {
	client   *http.Client
	endpoint string
	now      func() time.Time
}

func NewRateProvider(client *http.Client, endpoint string) *RateProvider {
	return &RateProvider{
		client:   client,
		endpoint: endpoint,
		now:      time.Now,
	}
}

func (provider *RateProvider) FetchUSDBRL(
	ctx context.Context,
) (domain.Quotation, error) {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		provider.endpoint,
		nil,
	)
	if err != nil {
		return domain.Quotation{}, fmt.Errorf(
			"%w: %w",
			ErrRequestQuotation,
			err,
		)
	}

	response, err := provider.client.Do(request)
	if err != nil {
		return domain.Quotation{}, fmt.Errorf(
			"%w: %w",
			ErrRequestQuotation,
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return domain.Quotation{}, fmt.Errorf(
			"%w: status code %d",
			ErrUnexpectedStatus,
			response.StatusCode,
		)
	}

	var payload awesomeAPIResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return domain.Quotation{}, fmt.Errorf(
			"%w: %w",
			ErrDecodeResponse,
			err,
		)
	}

	externalTimestamp, err := strconv.ParseInt(
		payload.USDBRL.Timestamp,
		10,
		64,
	)
	if err != nil {
		return domain.Quotation{}, fmt.Errorf(
			"%w: %w",
			ErrInvalidTimestamp,
			err,
		)
	}

	quotation, err := domain.NewQuotation(
		payload.USDBRL.Code,
		payload.USDBRL.CodeIn,
		payload.USDBRL.Bid,
		externalTimestamp,
		provider.now(),
	)
	if err != nil {
		return domain.Quotation{}, fmt.Errorf(
			"invalid quotation received from provider: %w",
			err,
		)
	}

	return quotation, nil
}

type awesomeAPIResponse struct {
	USDBRL awesomeAPIQuotation `json:"USDBRL"`
}

type awesomeAPIQuotation struct {
	Code      string `json:"code"`
	CodeIn    string `json:"codein"`
	Bid       string `json:"bid"`
	Timestamp string `json:"timestamp"`
}
