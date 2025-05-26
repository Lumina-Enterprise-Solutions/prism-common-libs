package utils

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8,max=50"`
	Name     string `validate:"required"`
}

func TestFormatValidationErrors(t *testing.T) {
	validator := validator.New()

	// Test with validation errors
	testData := TestStruct{
		Email:    "invalid-email",
		Password: "short",
		Name:     "",
	}

	err := validator.Struct(testData)
	assert.Error(t, err)

	errors := FormatValidationErrors(err)

	assert.Contains(t, errors, "email")
	assert.Contains(t, errors, "password")
	assert.Contains(t, errors, "name")

	assert.Contains(t, errors["email"], "valid email")
	assert.Contains(t, errors["password"], "at least")
	assert.Contains(t, errors["name"], "required")
}

func TestFormatValidationErrors_NoErrors(t *testing.T) {
	// Test with non-validation error
	errors := FormatValidationErrors(assert.AnError)
	assert.Empty(t, errors)
}
