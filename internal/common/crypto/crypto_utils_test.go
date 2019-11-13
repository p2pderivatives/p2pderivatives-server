package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	saltLen       = 32
	saltStringLen = 64
	keyLen        = 32
	keyStringLen  = 64
)

func TestGenerateSalt_With32_Returns64CharString(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	// Act
	salt, err := GenerateSalt(saltLen)
	// Assert
	assert.NoError(err)
	assert.NotEmpty(salt)
	assert.Len(salt, saltStringLen)
}

func TestGetPasswordProtectedForm_With32BytesSaltAndKey_Returns128CharString(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	password := "1SuperStrongPassword%"
	salt, _ := GenerateSalt(saltLen)

	// Act
	protectedForm := GetPasswordProtectedForm(password, salt, 3, 32*1024, 4, keyLen)

	// Assert
	assert.Len(protectedForm, 128)
}

func TestIsPasswordValid_WithValidPassword_ReturnsTrue(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	password := "1SuperStrongPassword%"
	salt, _ := GenerateSalt(saltLen)
	protectedForm := GetPasswordProtectedForm(password, salt, 3, 32*1024, 4, keyLen)

	// Act
	isValid := IsPasswordValid(password, protectedForm, saltLen, 3, 32*1024, 4, keyLen)

	// Assert
	assert.True(isValid)
}

func TestIsPasswordValid_WithInvalidPassword_ReturnsFalse(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	password := "1SuperStrongPassword%"
	badPassword := "2SuperStrongPassword%"
	salt, _ := GenerateSalt(32)
	protectedForm := GetPasswordProtectedForm(password, salt, 3, 32*1024, 4, keyLen)

	// Act
	isValid := IsPasswordValid(badPassword, protectedForm, saltLen, 3, 32*1024, 4, keyLen)

	// Assert
	assert.False(isValid)
}
