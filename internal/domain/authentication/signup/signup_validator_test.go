package signup

import (
	"testing"
)

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		valid       bool
		wantUpper   bool
		wantLower   bool
		wantNum     bool
		wantSpecial bool
	}{
		{
			name:        "valid password with all requirements",
			password:    "SecurePass123!",
			valid:       true,
			wantUpper:   true,
			wantLower:   true,
			wantNum:     true,
			wantSpecial: true,
		},
		{
			name:        "missing uppercase",
			password:    "securepass123!",
			valid:       false,
			wantUpper:   false,
			wantLower:   true,
			wantNum:     true,
			wantSpecial: true,
		},
		{
			name:        "missing lowercase",
			password:    "SECUREPASS123!",
			valid:       false,
			wantUpper:   true,
			wantLower:   false,
			wantNum:     true,
			wantSpecial: true,
		},
		{
			name:        "missing numbers",
			password:    "SecurePassWord!",
			valid:       false,
			wantUpper:   true,
			wantLower:   true,
			wantNum:     false,
			wantSpecial: true,
		},
		{
			name:        "missing special character",
			password:    "SecurePass123",
			valid:       false,
			wantUpper:   true,
			wantLower:   true,
			wantNum:     true,
			wantSpecial: false,
		},
		{
			name:        "password with various special characters",
			password:    "Pass@2024#Word",
			valid:       true,
			wantUpper:   true,
			wantLower:   true,
			wantNum:     true,
			wantSpecial: true,
		},
		{
			name:        "password with underscore special char",
			password:    "MyPass_123word",
			valid:       true,
			wantUpper:   true,
			wantLower:   true,
			wantNum:     true,
			wantSpecial: true,
		},
		{
			name:        "empty password",
			password:    "",
			valid:       false,
			wantUpper:   false,
			wantLower:   false,
			wantNum:     false,
			wantSpecial: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strength := ValidatePasswordStrength(tt.password)

			if strength.IsValid != tt.valid {
				t.Errorf("ValidatePasswordStrength() IsValid = %v, want %v", strength.IsValid, tt.valid)
			}

			if strength.HasUppercase != tt.wantUpper {
				t.Errorf("ValidatePasswordStrength() HasUppercase = %v, want %v", strength.HasUppercase, tt.wantUpper)
			}

			if strength.HasLowercase != tt.wantLower {
				t.Errorf("ValidatePasswordStrength() HasLowercase = %v, want %v", strength.HasLowercase, tt.wantLower)
			}

			if strength.HasNumber != tt.wantNum {
				t.Errorf("ValidatePasswordStrength() HasNumber = %v, want %v", strength.HasNumber, tt.wantNum)
			}

			if strength.HasSpecial != tt.wantSpecial {
				t.Errorf("ValidatePasswordStrength() HasSpecial = %v, want %v", strength.HasSpecial, tt.wantSpecial)
			}
		})
	}
}

func TestValidateSignupRequest_PasswordStrength(t *testing.T) {
	tests := []struct {
		name             string
		request          *SignupRequest
		shouldValidate   bool
		hasPasswordError bool
	}{
		{
			name: "valid signup request with strong password",
			request: &SignupRequest{
				Email:    "user@example.com",
				Password: "SecurePass123!",
				Username: "john_doe",
				FullName: "John Doe",
			},
			shouldValidate:   true,
			hasPasswordError: false,
		},
		{
			name: "weak password missing special character",
			request: &SignupRequest{
				Email:    "user@example.com",
				Password: "SecurePass123",
				Username: "john_doe",
				FullName: "John Doe",
			},
			shouldValidate:   true,
			hasPasswordError: true,
		},
		{
			name: "weak password missing numbers",
			request: &SignupRequest{
				Email:    "user@example.com",
				Password: "SecurePassword!",
				Username: "john_doe",
				FullName: "John Doe",
			},
			shouldValidate:   true,
			hasPasswordError: true,
		},
		{
			name: "weak password missing uppercase",
			request: &SignupRequest{
				Email:    "user@example.com",
				Password: "securepass123!",
				Username: "john_doe",
				FullName: "John Doe",
			},
			shouldValidate:   true,
			hasPasswordError: true,
		},
		{
			name: "weak password missing lowercase",
			request: &SignupRequest{
				Email:    "user@example.com",
				Password: "SECUREPASS123!",
				Username: "john_doe",
				FullName: "John Doe",
			},
			shouldValidate:   true,
			hasPasswordError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateSignupRequest(tt.request)

			if !tt.shouldValidate && len(errors) == 0 {
				t.Errorf("ValidateSignupRequest() expected errors but got none")
			}

			hasPasswordError := false
			for _, err := range errors {
				if err.Field == "Password" && len(err.Message) > 0 {
					hasPasswordError = true
					break
				}
			}

			if hasPasswordError != tt.hasPasswordError {
				t.Errorf("ValidateSignupRequest() hasPasswordError = %v, want %v", hasPasswordError, tt.hasPasswordError)
				if hasPasswordError {
					for _, err := range errors {
						if err.Field == "Password" {
							t.Logf("Password error: %s", err.Message)
						}
					}
				}
			}
		})
	}
}
