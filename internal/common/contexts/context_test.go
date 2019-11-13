package contexts

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextsGetUserID_WithNoID_Panics(t *testing.T) {
	// Arrange
	assert := assert.New(t)

	// Act
	expectPanics := func() { GetUserID(context.Background()) }

	// Assert
	assert.Panics(expectPanics)
}

func TestContextsGetUserID_WithSetUserID_ReturnsCorrectValue(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	userID := "UserA"
	ctx := SetUserID(context.Background(), userID)

	// Act
	result := GetUserID(ctx)
	assert.Equal(userID, result)

}

func TestContextsGetUserID_WithValue_ReturnsCorrectValue(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	userID := "UserB"
	newCtx := context.WithValue(context.Background(), UserID, userID)

	// Act
	result := GetUserID(newCtx)

	// Assert
	assert.Equal(userID, result)
}
