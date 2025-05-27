package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

const (
	HexDivisor = 2 // For converting length to bytes when generating hex strings
)

// GenerateRandomBytes generates random bytes of specified length
func GenerateRandomBytes(length int) []byte {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(fmt.Sprintf("Failed to generate random bytes: %v", err))
	}
	return bytes
}

// GenerateRandomString generates a random hex string
func GenerateRandomString(length int) string {
	bytes := GenerateRandomBytes(length / HexDivisor)
	return hex.EncodeToString(bytes)
}
