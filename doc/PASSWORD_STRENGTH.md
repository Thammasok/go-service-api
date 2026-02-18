# Password Strength Requirements

## Overview
Passwords must meet strong security requirements to protect user accounts from unauthorized access.

## Requirements

A valid password must contain ALL of the following:

1. **Uppercase Letters** (A-Z)
   - At least one uppercase letter is required
   - Example: `S` in `SecurePass123!`

2. **Lowercase Letters** (a-z)
   - At least one lowercase letter is required  
   - Example: `e` in `SecurePass123!`

3. **Numbers** (0-9)
   - At least one numeric digit is required
   - Example: `1`, `2`, `3` in `SecurePass123!`

4. **Special Characters**
   - At least one special character is required
   - Allowed special characters: `!@#$%^&*()_+=[]{}";:<>?/\|-`
   - Example: `!` in `SecurePass123!`

## Length Requirement
- Minimum: **8 characters**
- Maximum: **255 characters**

## Examples

### ✅ Valid Passwords
- `SecurePass123!` - Has uppercase, lowercase, number, and special char
- `MyPassword@2024` - Strong combination of all requirements
- `Tr0pic@lBreeze` - Uses letter 'O' as zero substitute with special char
- `Summer#Heat99` - Multiple numbers and special character
- `Welcome_L0gin` - Underscore as special character

### ❌ Invalid Passwords
- `securepass123!` - Missing uppercase letter
- `SECUREPASS123!` - Missing lowercase letter
- `SecurePassword!` - Missing number
- `SecurePass123` - Missing special character
- `Short1!` - Too short (only 7 characters)

## Security Best Practices

1. **Use Unique Passwords**: Don't reuse passwords across different services
2. **Avoid Dictionary Words**: Don't use common words that can be easily guessed
3. **Mix Character Types**: Use uppercase, lowercase, numbers, and special characters strategically
4. **Avoid Personal Information**: Don't use birthdays, names, or other personal details
5. **Change Regularly**: Update passwords periodically for enhanced security

## Testing Password Strength

The application includes `ValidatePasswordStrength()` function to check password requirements:

```go
strength := ValidatePasswordStrength("YourPassword123!")

// Check individual requirements
if !strength.HasUppercase {
    // Password is missing uppercase letters
}
if !strength.HasLowercase {
    // Password is missing lowercase letters
}
if !strength.HasNumber {
    // Password is missing numbers
}
if !strength.HasSpecial {
    // Password is missing special characters
}

// Check overall validity
if !strength.IsValid {
    // Password does not meet all requirements
}
```

## API Response Examples

### Error Response (Weak Password)
```json
{
  "error": "Validation failed",
  "errors": [
    {
      "field": "Password",
      "message": "Password must contain uppercase letters, lowercase letters, numbers, and special characters (!@#$%^&*()_+=[]{};:'\",.<>?/\\|-)"
    }
  ]
}
```

### Success Response (Strong Password)
```json
{
  "message": "User registered successfully",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "username": "john_doe",
    "fullName": "John Doe",
    "isActive": true,
    "createdAt": "2026-02-18T10:30:00Z"
  }
}
```

## Implementation Details

### Validation Layers

1. **Handler Layer** (`signup_handler.go`)
   - Receives and binds request
   - Calls `ValidateSignupRequest()` for field validation
   - Returns 400 Bad Request with errors if validation fails

2. **Validator Layer** (`signup_validator.go`)
   - Structural validation using `go-playground/validator`
   - Password strength validation using regex patterns
   - Returns detailed error messages for each failed validation

3. **Service Layer** (`signup_service.go`)
   - Double-checks password strength before hashing
   - Prevents weak passwords from being saved
   - Hashes validated passwords using Argon2-ID

### Regex Patterns Used

- **Uppercase**: `[A-Z]` - Matches any uppercase letter
- **Lowercase**: `[a-z]` - Matches any lowercase letter
- **Numbers**: `[0-9]` - Matches any digit
- **Special**: `[!@#$%^&*()_+=\[\]{};:\'",.<>?/\\|-]` - Matches special characters

## Testing

Run the password strength validation tests:

```bash
go test ./internal/domain/authentication/signup -v -run TestValidatePasswordStrength
go test ./internal/domain/authentication/signup -v -run TestValidateSignupRequest_PasswordStrength
```

Expected test coverage:
- Valid password with all requirements ✓
- Missing uppercase ✓
- Missing lowercase ✓
- Missing numbers ✓
- Missing special characters ✓
- Various special character combinations ✓
- Empty password ✓
