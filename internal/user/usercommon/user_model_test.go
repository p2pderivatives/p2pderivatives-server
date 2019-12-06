package usercommon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_NewUser_ContainsCorrectInfo(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	account := "account"
	name := "name"
	password := "password"

	// Act
	user := NewUser(account, name, password)

	// Assert
	assert.NotEmpty(user.ID)
	assert.Equal(account, user.Account)
	assert.Equal(name, user.Name)
	assert.Equal(password, user.Password)
}
