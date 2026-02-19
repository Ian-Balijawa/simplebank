package gapi

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	minAmountDefault = int64(-9_000_000_000_000_000_000)
	maxAmountDefault = int64(9_000_000_000_000_000_000)
)

func optionalInt64(value *wrapperspb.Int64Value, defaultValue int64) int64 {
	if value == nil {
		return defaultValue
	}
	return value.Value
}

func optionalTime(value *timestamppb.Timestamp, defaultValue time.Time) (time.Time, error) {
	if value == nil {
		return defaultValue, nil
	}
	if err := value.CheckValid(); err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp: %w", err)
	}
	return value.AsTime().UTC(), nil
}
