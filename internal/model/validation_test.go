package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationError_Error(t *testing.T) {
	t.Run("Returns message", func(t *testing.T) {
		err := ValidationError{
			Field:   "username",
			Message: "username is required",
		}

		assert.Equal(t, "username is required", err.Error())
	})

	t.Run("Empty message", func(t *testing.T) {
		err := ValidationError{
			Field:   "password",
			Message: "",
		}

		assert.Equal(t, "", err.Error())
	})
}

func TestNewValidationError(t *testing.T) {
	t.Run("Creates validation error with field and message", func(t *testing.T) {
		err := NewValidationError("email", "email format is invalid")

		assert.Equal(t, "email", err.Field)
		assert.Equal(t, "email format is invalid", err.Message)
	})

	t.Run("Creates validation error with empty field", func(t *testing.T) {
		err := NewValidationError("", "general error")

		assert.Equal(t, "", err.Field)
		assert.Equal(t, "general error", err.Message)
	})

	t.Run("Creates validation error with empty message", func(t *testing.T) {
		err := NewValidationError("title", "")

		assert.Equal(t, "title", err.Field)
		assert.Equal(t, "", err.Message)
	})

	t.Run("Validation error is error interface", func(t *testing.T) {
		err := NewValidationError("url", "URL is invalid")

		var errorInterface error = err
		assert.NotNil(t, errorInterface)
		assert.Equal(t, "URL is invalid", errorInterface.Error())
	})

	t.Run("Multiple validation errors are independent", func(t *testing.T) {
		err1 := NewValidationError("field1", "error1")
		err2 := NewValidationError("field2", "error2")

		assert.Equal(t, "field1", err1.Field)
		assert.Equal(t, "error1", err1.Message)
		assert.Equal(t, "field2", err2.Field)
		assert.Equal(t, "error2", err2.Message)
	})

	t.Run("Validation error struct equality", func(t *testing.T) {
		err1 := ValidationError{
			Field:   "test_field",
			Message: "test message",
		}
		err2 := ValidationError{
			Field:   "test_field",
			Message: "test message",
		}

		assert.Equal(t, err1.Field, err2.Field)
		assert.Equal(t, err1.Message, err2.Message)
	})
}
