package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

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
