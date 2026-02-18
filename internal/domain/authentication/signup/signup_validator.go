package signup

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// PasswordStrength represents password strength validation rules
type PasswordStrength struct {
	HasUppercase bool
	HasLowercase bool
	HasNumber    bool
	HasSpecial   bool
	IsValid      bool
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidatePasswordStrength checks if password contains uppercase, lowercase, numbers, and special characters
func ValidatePasswordStrength(password string) PasswordStrength {
	strength := PasswordStrength{
		HasUppercase: regexp.MustCompile(`[A-Z]`).MatchString(password),
		HasLowercase: regexp.MustCompile(`[a-z]`).MatchString(password),
		HasNumber:    regexp.MustCompile(`[0-9]`).MatchString(password),
		HasSpecial:   regexp.MustCompile(`[!@#$%^&*()_+=\[\]{};:'",.<>?/\\|-]`).MatchString(password),
	}

	strength.IsValid = strength.HasUppercase && strength.HasLowercase && strength.HasNumber && strength.HasSpecial
	return strength
}

// ValidateSignupRequest validates the signup request
func ValidateSignupRequest(req *SignupRequest) []ValidationError {
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

	// Validate password strength if no structural errors
	if len(errors) == 0 && req.Password != "" {
		strength := ValidatePasswordStrength(req.Password)
		if !strength.IsValid {
			ve := ValidationError{
				Field:   "Password",
				Message: "Password must contain uppercase letters, lowercase letters, numbers, and special characters",
			}
			errors = append(errors, ve)
		}
	}

	return errors
}
