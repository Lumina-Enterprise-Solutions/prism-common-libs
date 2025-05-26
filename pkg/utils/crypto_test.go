package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomBytes(t *testing.T) {
	// Test different lengths
	lengths := []int{16, 32, 64}

	for _, length := range lengths {
		bytes := GenerateRandomBytes(length)
		assert.Len(t, bytes, length)

		// Test uniqueness by generating multiple times
		bytes2 := GenerateRandomBytes(length)
		assert.NotEqual(t, bytes, bytes2)
	}
}

func TestGenerateRandomString(t *testing.T) {
	// Test different lengths
	lengths := []int{16, 32, 64}

	for _, length := range lengths {
		str := GenerateRandomString(length)
		assert.Len(t, str, length)

		// Test uniqueness
		str2 := GenerateRandomString(length)
		assert.NotEqual(t, str, str2)

		// Test that it's valid hex
		for _, char := range str {
			assert.True(t, (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f'))
		}
	}
}
