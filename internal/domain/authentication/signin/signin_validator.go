package signin

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidateSignupRequest validates the signup request
func ValidateSigninRequest(req *SigninRequest) []ValidationError {
	var errors []ValidationError

	if err := validate.Struct(req); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			ve := ValidationError{
				Field: err.Field(),
			}

			switch err.Tag() {
			case "required":
				ve.Message = fmt.Sprintf("%s is required", err.Field())
			case "email":
				ve.Message = fmt.Sprintf("%s must be a valid email address", err.Field())
			case "min":
				ve.Message = fmt.Sprintf("%s must be at least %s characters", err.Field(), err.Param())
			case "max":
				ve.Message = fmt.Sprintf("%s must not exceed %s characters", err.Field(), err.Param())
			default:
				ve.Message = fmt.Sprintf("%s is invalid", err.Field())
			}

			errors = append(errors, ve)
		}
	}

	return errors
}
