package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
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
	bytes := GenerateRandomBytes(length / 2)
	return hex.EncodeToString(bytes)
}
