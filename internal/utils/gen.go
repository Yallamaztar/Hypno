package utils

import (
	"crypto/rand"
)

func GenerateCode() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}

	for i := range buf {
		buf[i] = chars[int(buf[i])%len(chars)]
	}

	return string(buf)
}
