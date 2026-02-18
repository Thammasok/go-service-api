package hashpassword

import (
	"fmt"

	"golang.org/x/crypto/argon2"
)

// HashPassword hashes a password using Argon2
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	// Argon2 parameters
	// time=3, memory=64MB, parallelism=4, tag length=32
	hash := argon2.IDKey(
		[]byte(password),
		[]byte("salt"),
		3,       // time cost: number of iterations
		64*1024, // memory cost: 64MB
		4,       // parallelism
		32,      // tag length
	)

	// Convert hash to string format (base64-like representation)
	hashStr := fmt.Sprintf("%x", hash)
	return hashStr, nil
}

// CheckPassword checks if a given password matches a hashed password
func CheckPassword(password, hashedPassword string) bool {
	hash := argon2.IDKey(
		[]byte(password),
		[]byte("salt"),
		3,       // time cost: number of iterations
		64*1024, // memory cost: 64MB
		4,       // parallelism
		32,      // tag length
	)
	hashStr := fmt.Sprintf("%x", hash)
	return hashStr == hashedPassword
}
