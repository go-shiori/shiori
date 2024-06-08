package model

// ValidationError represents a validation error.
// This errors are used in the domain layer to indicate an error that is caused generally
// by the user and has to be sent back via the API or appropriate channel.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (v ValidationError) Error() string {
	return v.Message
}

func NewValidationError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}
