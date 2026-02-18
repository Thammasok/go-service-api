package hashpassword

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantError bool
	}{
		{
			name:      "valid password",
			password:  "testPassword123!",
			wantError: false,
		},
		{
			name:      "empty password",
			password:  "",
			wantError: true,
		},
		{
			name:      "short password",
			password:  "abc",
			wantError: false,
		},
		{
			name:      "long password",
			password:  "this_is_a_very_long_password_with_many_characters_that_should_still_work_fine_for_hashing",
			wantError: false,
		},
		{
			name:      "password with special characters",
			password:  "P@ssw0rd!#$%^&*()",
			wantError: false,
		},
		{
			name:      "password with unicode",
			password:  "password123",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)

			if (err != nil) != tt.wantError {
				t.Errorf("HashPassword() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && hash == "" {
				t.Errorf("HashPassword() returned empty hash for valid password")
			}
		})
	}
}

func TestHashPasswordConsistency(t *testing.T) {
	password := "consistencyTestPassword123"

	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)

	if hash1 != hash2 {
		t.Errorf("HashPassword() is not consistent: got %s, then %s", hash1, hash2)
	}
}

func TestCheckPassword(t *testing.T) {
	tests := []struct {
		name       string
		password   string
		checkPass  string
		shouldPass bool
	}{
		{
			name:       "correct password",
			password:   "correctPassword123",
			checkPass:  "correctPassword123",
			shouldPass: true,
		},
		{
			name:       "wrong password",
			password:   "password123",
			checkPass:  "wrongPassword123",
			shouldPass: false,
		},
		{
			name:       "case sensitive",
			password:   "Password123",
			checkPass:  "password123",
			shouldPass: false,
		},
		{
			name:       "special characters match",
			password:   "P@ss!word#123",
			checkPass:  "P@ss!word#123",
			shouldPass: true,
		},
		{
			name:       "special characters no match",
			password:   "P@ss!word#123",
			checkPass:  "P@ss!word$123",
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPassword, err := HashPassword(tt.password)
			if err != nil {
				t.Fatalf("HashPassword() failed: %v", err)
			}

			matches := CheckPassword(tt.checkPass, hashedPassword)

			if matches != tt.shouldPass {
				t.Errorf("CheckPassword() = %v, want %v", matches, tt.shouldPass)
			}
		})
	}
}

func BenchmarkHashPassword(b *testing.B) {
	password := "benchmarkPassword123"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		HashPassword(password)
	}
}

func BenchmarkCheckPassword(b *testing.B) {
	password := "benchmarkPassword123"
	hashedPassword, _ := HashPassword(password)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		CheckPassword(password, hashedPassword)
	}
}
