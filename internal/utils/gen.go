package utils

import (
	"crypto/rand"
)

const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateCode generates a linking code for linking ingame account with discord account
func GenerateCode() string {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}

	for i := range buf {
		buf[i] = chars[int(buf[i])%len(chars)]
	}

	return string(buf)
}
