package authentication

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"p2pderivatives-server/internal/common/contexts"
	"p2pderivatives-server/internal/common/token"
	"p2pderivatives-server/internal/user/usercommon"
	"p2pderivatives-server/test"
	"p2pderivatives-server/test/mocks/mock_usercommon"
	"p2pderivatives-server/test/mocks/mock_userservice"
)

const (
	name        = "test"
	password    = "p@ssw0rd"
	badPassword = "p@sw0rd"
)

var userID string

func initService() (context.Context, *usercommon.Config, usercommon.ServiceIf) {
	userConfig := usercommon.DefaultUserConfiguration()
	ctx := context.Background()
	userService := mock_userservice.NewServiceMock()
	user := usercommon.NewUser(name, password)
	userService.CreateUser(ctx, user)
	userID = user.ID
	tokenConfig := &token.Config{}
	test.GetTestConfig().InitializeComponentConfig(tokenConfig)
	token.Init(tokenConfig)
	return ctx, userConfig, userService
}

func TestAuthenticationLogin_WithCorrectParameters_Succeeds(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	ctx, config, service := initService()
	controller := NewController(service, config)
	request := &LoginRequest{
		Name:     name,
		Password: password,
	}
	// Act
	response, err := controller.Login(ctx, request)

	// Assert
	assert.NoError(err)
	assert.NotNil(response.Token)
}

func TestAuthenticationLogin_WithIncorrectParameters_Fails(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	ctx, config, service := initService()
	controller := NewController(service, config)
	request := &LoginRequest{
		Name:     name,
		Password: badPassword,
	}
	// Act
	response, err := controller.Login(ctx, request)

	// Assert
	assert.Error(err)
	assert.Nil(response)
}

func TestAuthenticationRefresh_WithCorrectToken_Succeeds(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	service := mock_usercommon.NewMockServiceIf(ctrl)
	ctx, config, _ := initService()
	userInstance := usercommon.NewUser(name, "")
	tokenInfo := &usercommon.TokenInfo{
		AccessToken:  "",
		RefreshToken: "",
		ExpiresIn:    1,
	}
	controller := NewController(service, config)
	loginRequest := &LoginRequest{
		Name:     name,
		Password: password,
	}
	service.EXPECT().AuthenticateUser(gomock.Any(), name, password).Return(userInstance, tokenInfo, nil)
	service.EXPECT().RefreshUserToken(gomock.Any(), gomock.Any()).Return(tokenInfo, nil)
	loginResponse, err := controller.Login(ctx, loginRequest)
	request := &RefreshRequest{
		RefreshToken: loginResponse.Token.RefreshToken,
	}

	// Act
	response, err := controller.Refresh(ctx, request)

	// Assert
	assert.NoError(err)
	assert.NotNil(response)
}

func TestAuthenticationRefresh_WithIncorrectToken_Fails(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	service := mock_usercommon.NewMockServiceIf(ctrl)
	ctx, config, _ := initService()
	controller := NewController(service, config)
	request := &RefreshRequest{
		RefreshToken: "thisIsNotAToken",
	}

	service.EXPECT().RefreshUserToken(gomock.Any(), request.RefreshToken).Return(nil, errors.New(""))

	// Act
	response, err := controller.Refresh(ctx, request)

	// Assert
	assert.Error(err)
	assert.Nil(response)
}

func TestAuthenticationLogout_WithCorrectToken_NoError(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	service := mock_usercommon.NewMockServiceIf(ctrl)
	ctx, config, _ := initService()
	controller := NewController(service, config)
	loginRequest := &LoginRequest{
		Name:     name,
		Password: password,
	}
	userInstance := usercommon.NewUser(name, "")
	service.EXPECT().AuthenticateUser(gomock.Any(), name, password).
		Return(userInstance, &usercommon.TokenInfo{}, nil)
	service.EXPECT().RevokeRefreshToken(gomock.Any(), "").Return(nil)
	loginResponse, err := controller.Login(ctx, loginRequest)
	request := &LogoutRequest{
		RefreshToken: loginResponse.Token.RefreshToken,
	}

	// Act
	response, err := controller.Logout(ctx, request)

	// Assert
	assert.NoError(err)
	assert.NotNil(response)
}

func TestAuthenticationLogout_WithIncorrectToken_NoError(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	service := mock_usercommon.NewMockServiceIf(ctrl)
	ctx, config, _ := initService()
	controller := NewController(service, config)
	request := &LogoutRequest{
		RefreshToken: "thisIsNotAToken",
	}
	service.EXPECT().RevokeRefreshToken(gomock.Any(), request.RefreshToken).Return(errors.New(""))

	// Act
	response, err := controller.Logout(ctx, request)

	// Assert
	assert.NoError(err)
	assert.NotNil(response)
}

func TestChangeUserPassword_WithValidPassword_Success(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	ctx, config, service := initService()
	controller := NewController(service, config)
	ctx = contexts.SetUserID(ctx, userID)
	newPassword := "P@ssw0rd"
	updatePasswordRequest := UpdatePasswordRequest{
		NewPassword: newPassword,
		OldPassword: password,
	}
	loginRequest := LoginRequest{Name: name, Password: password}

	// Act
	_, err1 := controller.UpdatePassword(ctx, &updatePasswordRequest)
	_, err2 := controller.Login(context.Background(), &loginRequest)
	loginRequest.Password = newPassword
	_, err3 := controller.Login(context.Background(), &loginRequest)

	// Assert
	assert.NoError(err1)
	assert.Error(err2)
	assert.NoError(err3)
}

func TestChangeUserPassword_WithInvalidPassword_Fails(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	ctx, config, service := initService()
	controller := NewController(service, config)
	ctx = contexts.SetUserID(ctx, userID)
	newPassword := "Test"
	updatePasswordRequest := UpdatePasswordRequest{
		NewPassword: newPassword,
	}

	// Act
	_, err := controller.UpdatePassword(ctx, &updatePasswordRequest)

	// Assert
	assert.Error(err)
}
