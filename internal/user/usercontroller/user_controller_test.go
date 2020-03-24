package usercontroller_test

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/assert"

	"p2pderivatives-server/internal/common/contexts"
	"p2pderivatives-server/internal/user/usercommon"
	"p2pderivatives-server/internal/user/usercontroller"
	"p2pderivatives-server/test/mocks/mock_usercontroller"
	"p2pderivatives-server/test/mocks/mock_userservice"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var userCount int

func createController() *usercontroller.Controller {
	userConfig := usercommon.DefaultUserConfiguration()
	service := mock_userservice.NewServiceMock()
	return usercontroller.NewController(service, userConfig)
}

func createUserRegisterRequest(model *usercommon.User) *usercontroller.UserRegisterRequest {
	return &usercontroller.UserRegisterRequest{
		Name:     model.Name,
		Password: model.Password,
	}
}

func createUser() *usercommon.User {
	userCount++
	return usercommon.NewUser(
		strings.Join([]string{"Name", strconv.Itoa(userCount)}, ""),
		strings.Join([]string{"Password", strconv.Itoa(userCount)}, ""))
}

func createInfos(count int) []usercontroller.UserInfo {
	users := make([]usercontroller.UserInfo, count)

	for i := 0; i < count; i++ {
		users[i] = usercontroller.UserInfo{Name: strconv.Itoa(i)}
	}

	return users
}

func TestRegisterUser_WithNewUser_IsRegistered(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	controller := createController()
	defer controller.Close()
	userRegisterRequest := createUserRegisterRequest(createUser())
	userInfo := usercontroller.UserInfo{
		Name: userRegisterRequest.Name,
	}
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	stream := mock_usercontroller.NewMockUser_GetUserListServer(mockCtrl)
	stream.EXPECT().Send(&userInfo).Times(1)
	stream.EXPECT().Context().Return(ctx).Times(1)

	// Act
	_, err := controller.RegisterUser(ctx, userRegisterRequest)
	err2 := controller.GetUserList(nil, stream)

	// Assert
	assert.NoError(err)
	assert.NoError(err2)
	mockCtrl.Finish()
}

func TestRegisterUser_WithExistingUser_ReturnsAlreadyExistsError(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	controller := createController()
	defer controller.Close()
	request := createUserRegisterRequest(createUser())
	ctx := context.Background()

	// Act
	controller.RegisterUser(ctx, request)
	_, err := controller.RegisterUser(ctx, request)
	st, ok := status.FromError(err)

	// Assert
	assert.Error(err)
	assert.True(ok)
	assert.Equal(codes.AlreadyExists, st.Code())
}

func TestUnregisterUser_WithRegisteredUser_RemovedFromList(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	controller := createController()
	defer controller.Close()
	request := createUserRegisterRequest(createUser())
	ctx := context.Background()
	response, _ := controller.RegisterUser(ctx, request)
	ctx = contexts.SetUserID(ctx, response.Id)
	mockCtrl := gomock.NewController(t)
	stream := mock_usercontroller.NewMockUser_GetUserListServer(mockCtrl)
	stream.EXPECT().Send(nil).Times(0)
	stream.EXPECT().Context().Return(ctx)
	unregisterRequest := usercontroller.UnregisterUserRequest{}

	// Act
	_, err := controller.UnregisterUser(ctx, &unregisterRequest)
	err2 := controller.GetUserList(nil, stream)

	// Assert
	assert.NoError(err)
	assert.NoError(err2)
	mockCtrl.Finish()
}

func TestUnregisterUser_WithNonRegisteredUser_ReturnsNotFoundError(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	controller := createController()
	defer controller.Close()
	model := createUser()
	unregisterRequest := usercontroller.UnregisterUserRequest{}
	ctx := context.Background()
	ctx = contexts.SetUserID(ctx, model.ID)

	// Act
	_, err := controller.UnregisterUser(ctx, &unregisterRequest)
	st, ok := status.FromError(err)
	// Assert
	assert.Error(err)
	assert.True(ok)
	assert.Equal(codes.NotFound, st.Code())
}
