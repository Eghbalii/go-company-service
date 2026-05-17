package hash

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const cost = bcrypt.DefaultCost

// Password returns the bcrypt hash of plaintext.
func Password(plaintext string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plaintext), cost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(h), nil
}

// CheckPassword reports whether plaintext matches the stored hash.
func CheckPassword(hash, plaintext string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext)) == nil
}
