package domain

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

var ErrInvalidQuotation = errors.New("invalid quotation")

type Quotation struct {
	Code              string
	CodeIn            string
	Bid               string
	ExternalTimestamp int64
	FetchedAt         time.Time
}

func NewQuotation(
	code string,
	codeIn string,
	bid string,
	externalTimestamp int64,
	fetchedAt time.Time,
) (Quotation, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	codeIn = strings.ToUpper(strings.TrimSpace(codeIn))
	bid = strings.TrimSpace(bid)

	if len(code) != 3 {
		return Quotation{}, fmt.Errorf(
			"%w: code must contain exactly 3 characters",
			ErrInvalidQuotation,
		)
	}

	if len(codeIn) != 3 {
		return Quotation{}, fmt.Errorf(
			"%w: codeIn must contain exactly 3 characters",
			ErrInvalidQuotation,
		)
	}

	bidValue, err := strconv.ParseFloat(bid, 64)
	if err != nil || bidValue <= 0 || math.IsNaN(bidValue) || math.IsInf(bidValue, 0) {
		return Quotation{}, fmt.Errorf(
			"%w: bid must be a positive number",
			ErrInvalidQuotation,
		)
	}

	if externalTimestamp <= 0 {
		return Quotation{}, fmt.Errorf(
			"%w: external timestamp must be positive",
			ErrInvalidQuotation,
		)
	}

	if fetchedAt.IsZero() {
		return Quotation{}, fmt.Errorf(
			"%w: fetchedAt is required",
			ErrInvalidQuotation,
		)
	}

	return Quotation{
		Code:              code,
		CodeIn:            codeIn,
		Bid:               bid,
		ExternalTimestamp: externalTimestamp,
		FetchedAt:         fetchedAt.UTC(),
	}, nil
}
