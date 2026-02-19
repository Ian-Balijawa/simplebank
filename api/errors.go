package api

import "errors"

var (
	errInvalidAmountRange = errors.New("invalid amount range")
	errInvalidDateRange   = errors.New("invalid date range")
)
