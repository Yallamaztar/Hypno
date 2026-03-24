package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

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
