package crypto

import (
	"crypto/rand"
	"encoding/hex"

	"golang.org/x/crypto/argon2"
)

// GenerateSalt returns securely generated random bytes as a string.
// An error is returned if the system's secure random number generator fails
// to function correctly, in which case the caller should not continue.
func GenerateSalt(saltSize int) (string, error) {
	saltBytes := make([]byte, saltSize)
	_, err := rand.Read(saltBytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(saltBytes), nil
}

// GetPasswordProtectedForm returns the protected form of the password using the
// salt and other provided parameters.
// The returned value is as follow:
// salt + hashedPassword
// Recommended parameters are time = 3 and memory = 32 * 1024
func GetPasswordProtectedForm(
	password, salt string,
	time, memory uint32,
	threads uint8, keyLen uint32) string {
	hashedPassword := hex.EncodeToString(
		argon2.Key([]byte(password), []byte(salt), time, memory, threads, keyLen))
	return salt + hashedPassword
}

// IsPasswordValid return true if the given password matches the given protected
// form, false otherwise.
func IsPasswordValid(
	password, protectedForm string,
	saltLen int,
	time, memory uint32,
	threads uint8, keyLen uint32) bool {
	saltStringLen := saltLen * 2
	salt := protectedForm[:saltStringLen]
	hashedPassword := hex.EncodeToString(
		argon2.Key([]byte(password), []byte(salt), time, memory, threads, keyLen))
	return hashedPassword == protectedForm[saltStringLen:]
}
