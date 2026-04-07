package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func ParseDurationMultiplier(duration string) (int, error) {
	duration = strings.ToLower(strings.TrimSpace(duration))
	if duration == "" {
		return 0, fmt.Errorf("invalid duration")
	}

	unit := duration[len(duration)-1]
	valueStr := duration[:len(duration)-1]

	base := SafeInt(valueStr, 0)
	if base <= 0 {
		return 0, fmt.Errorf("invalid duration value")
	}

	var minutes int64

	switch unit {
	case 'm':
		if base > 60 {
			return 0, fmt.Errorf("minutes cannot exceed 60")
		}
		minutes = base

	case 'h':
		if base > 24 {
			return 0, fmt.Errorf("hours cannot exceed 24")
		}
		minutes = base * 60

	case 'd':
		if base > 30 {
			return 0, fmt.Errorf("days cannot exceed 30")
		}
		minutes = base * 1440

	default:
		return 0, fmt.Errorf("invalid duration unit")
	}

	// Tuned scaling (~30d ≈ 1500x)
	multiplier := max(int(float64(minutes)*0.035), 1)

	return multiplier, nil
}

// SafeInt safely converts a string to an int64, returning a default value if the conversion fails
func SafeInt(value string, defaultVal int64) int64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultVal
	}
	clean := strings.ReplaceAll(value, ",", "")
	if n, err := strconv.ParseInt(clean, 10, 64); err == nil {
		return n
	}
	return defaultVal
}

// ParseAmountArg parses a string argument representing an amount and returns the corresponding int value
func ParseAmountArg(arg string, balance int) (int, error) {
	switch strings.ToLower(arg) {
	case "all", "a":
		return int(balance), nil
	case "half", "h":
		return int(balance / 2), nil
	default:
		amount := ParseAmount(arg)
		if amount <= 0 {
			return 0, fmt.Errorf("invalid amount")
		}
		return int(amount), nil
	}
}

// ParseAmount parses a string representing an amount and returns the corresponding int64 value
func ParseAmount(amount string) int64 {
	amount = strings.ToLower(strings.TrimSpace(amount))
	if amount == "" {
		return 0
	}

	var mult int64
	switch amount[len(amount)-1] {
	case 'k':
		mult = 1_000
	case 'm':
		mult = 1_000_000
	case 'b':
		mult = 1_000_000_000
	case 't':
		mult = 1_000_000_000_000
	case 'q':
		mult = 1_000_000_000_000_000
	case 'z':
		return math.MaxInt64
	default:
		return SafeInt(amount, 0)
	}

	base := SafeInt(amount[:len(amount)-1], 0)
	return safeMulClamp(base, mult)
}
