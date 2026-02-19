package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	minAmountDefault = int64(-9_000_000_000_000_000_000)
	maxAmountDefault = int64(9_000_000_000_000_000_000)
)

func parseOptionalInt64(value string, defaultValue int64) (int64, error) {
	if strings.TrimSpace(value) == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value: %w", err)
	}
	return parsed, nil
}

func parseOptionalTime(value string, defaultValue time.Time) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return defaultValue, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time value (use RFC3339): %w", err)
	}
	return parsed, nil
}

func parseSortOrder(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "desc", nil
	}
	switch strings.ToLower(value) {
	case "asc", "desc":
		return strings.ToLower(value), nil
	default:
		return "", fmt.Errorf("invalid sort value")
	}
}

func parseDirection(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "any", nil
	}
	switch strings.ToLower(value) {
	case "any", "in", "out":
		return strings.ToLower(value), nil
	default:
		return "", fmt.Errorf("invalid direction value")
	}
}
