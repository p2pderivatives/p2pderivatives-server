package usercontroller_test

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

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
		Account:  model.Account,
		Name:     model.Name,
		Password: model.Password,
	}
}

func createUser() *usercommon.User {
	userCount++
	return usercommon.NewUser(
		strings.Join([]string{"Account", strconv.Itoa(userCount)}, ""),
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

func TestGetUserStatuses_NotifiesUserRegistered(t *testing.T) {
	model := createUser()
	expectedNotice := usercontroller.UserNotice{
		Name:   model.Name,
		Status: usercontroller.UserStatus_REGISTERED,
	}

	setup := func(controller *usercontroller.Controller, ctx context.Context) {}

	action := func(
		controller *usercontroller.Controller,
		baseUser *usercommon.User,
		mockCtrl *gomock.Controller,
		otherMockStream *mock_usercontroller.MockUser_GetUserStatusesServer,
		ctx context.Context,
		wg *sync.WaitGroup) {
		// Let time to the other client to connect.
		time.Sleep(time.Millisecond * 5)
		controller.RegisterUser(ctx, createUserRegisterRequest(model))
	}

	userNoticeTestHelper(t, &expectedNotice, setup, action)
}

func TestGetUserStatuses_NotifiesUserConnected(t *testing.T) {
	model := createUser()
	expectedNotice := usercontroller.UserNotice{
		Name:   model.Name,
		Status: usercontroller.UserStatus_CONNECTED,
	}

	setup := func(controller *usercontroller.Controller, ctx context.Context) {
		response, _ := controller.RegisterUser(ctx, createUserRegisterRequest(model))
		model.ID = response.Id
	}

	action := func(
		controller *usercontroller.Controller,
		baseUser *usercommon.User,
		mockCtrl *gomock.Controller,
		otherMockStream *mock_usercontroller.MockUser_GetUserStatusesServer,
		ctx context.Context,
		wg *sync.WaitGroup) {
		otherExpectedNotice := usercontroller.UserNotice{
			Name:   baseUser.Name,
			Status: usercontroller.UserStatus_CONNECTED,
		}
		wg.Add(1)
		mockStream := mock_usercontroller.NewMockUser_GetUserStatusesServer(mockCtrl)
		ctx = contexts.SetUserID(ctx, model.ID)
		mockStream.EXPECT().Context().Return(ctx).AnyTimes()
		mockStream.EXPECT().Send(&otherExpectedNotice).Times(1).Return(nil).Do(
			func(notice *usercontroller.UserNotice) {
				wg.Done()
			},
		)
		go controller.GetUserStatuses(&usercontroller.Empty{}, mockStream)
	}

	userNoticeTestHelper(t, &expectedNotice, setup, action)
}

func TestGetUserStatuses_NotifiesUserDisconnected(t *testing.T) {

	model := createUser()
	expectedNotice := usercontroller.UserNotice{
		Name:   model.Name,
		Status: usercontroller.UserStatus_DISCONNECTED,
	}

	setup := func(controller *usercontroller.Controller, ctx context.Context) {
		response, _ := controller.RegisterUser(ctx, createUserRegisterRequest(model))
		model.ID = response.Id
	}

	action := func(
		controller *usercontroller.Controller,
		baseUser *usercommon.User,
		mockCtrl *gomock.Controller,
		otherMockStream *mock_usercontroller.MockUser_GetUserStatusesServer,
		ctx context.Context,
		wg *sync.WaitGroup) {
		var thisWg sync.WaitGroup
		thisWg.Add(1)
		otherExpectedNotice := usercontroller.UserNotice{
			Name:   baseUser.Name,
			Status: usercontroller.UserStatus_CONNECTED,
		}
		thisConnectionExpectedNotice := usercontroller.UserNotice{
			Name:   model.Name,
			Status: usercontroller.UserStatus_CONNECTED,
		}
		wg.Add(1)
		ctx = contexts.SetUserID(ctx, model.ID)
		otherMockStream.EXPECT().Send(&thisConnectionExpectedNotice).
			Times(1).Return(nil).Do(
			func(notice *usercontroller.UserNotice) {
				t.Log("Received this notice.")
				thisWg.Done()
			})
		mockStream := mock_usercontroller.NewMockUser_GetUserStatusesServer(mockCtrl)
		mockStream.EXPECT().Context().Return(ctx).AnyTimes()
		mockStream.EXPECT().Send(&otherExpectedNotice).
			Times(1).Return(errors.New("force disconnect")).Do(
			func(notice *usercontroller.UserNotice) {
				t.Log("Waiting for other notice.")
				thisWg.Wait()
				t.Log("Got other notice.")
				wg.Done()
			})
		go controller.GetUserStatuses(&usercontroller.Empty{}, mockStream)
	}

	userNoticeTestHelper(t, &expectedNotice, setup, action)
}

func TestGetUserStatuses_NotifiesUserUnregistered(t *testing.T) {

	model := createUser()
	expectedNotice := usercontroller.UserNotice{
		Name:   model.Name,
		Status: usercontroller.UserStatus_UNREGISTERED,
	}

	setup := func(controller *usercontroller.Controller, ctx context.Context) {
	}

	action := func(
		controller *usercontroller.Controller,
		baseUser *usercommon.User,
		mockCtrl *gomock.Controller,
		otherMockStream *mock_usercontroller.MockUser_GetUserStatusesServer,
		ctx context.Context,
		wg *sync.WaitGroup) {
		// Let time to the other client to connect.
		time.Sleep(time.Millisecond * 5)
		var thisWg sync.WaitGroup
		thisWg.Add(1)
		thisRegisterExpectedNotice := usercontroller.UserNotice{
			Name:   model.Name,
			Status: usercontroller.UserStatus_REGISTERED,
		}
		otherMockStream.EXPECT().Send(&thisRegisterExpectedNotice).Times(1).Do(
			func(notice *usercontroller.UserNotice) {
				thisWg.Done()
			},
		)
		response, _ := controller.RegisterUser(ctx, createUserRegisterRequest(model))
		model.ID = response.Id
		thisWg.Wait()
		request := usercontroller.UnregisterUserRequest{}
		ctx = contexts.SetUserID(ctx, model.ID)
		controller.UnregisterUser(ctx, &request)
	}

	userNoticeTestHelper(t, &expectedNotice, setup, action)
}

func userNoticeTestHelper(
	t *testing.T,
	expectedNotice *usercontroller.UserNotice,
	setup func(controller *usercontroller.Controller, ctx context.Context),
	action func(
		controller *usercontroller.Controller,
		baseUser *usercommon.User,
		mockCtrl *gomock.Controller,
		otherMockStream *mock_usercontroller.MockUser_GetUserStatusesServer,
		ctx context.Context,
		wg *sync.WaitGroup)) {

	// Arrange
	var wg sync.WaitGroup
	wg.Add(1)
	controller := createController()
	defer controller.Close()
	sharedContext := context.Background()
	setup(controller, sharedContext)
	model := createUser()
	mockCtrl := gomock.NewController(t)
	mockStream := mock_usercontroller.NewMockUser_GetUserStatusesServer(mockCtrl)
	mockStream.EXPECT().Send(expectedNotice).Return(nil).Times(1).Do(
		func(notice *usercontroller.UserNotice) {
			wg.Done()
		})
	response, _ := controller.RegisterUser(sharedContext, createUserRegisterRequest(model))
	id := response.Id
	ctx := contexts.SetUserID(sharedContext, id)
	mockStream.EXPECT().Context().Return(ctx).AnyTimes()

	// Act
	go controller.GetUserStatuses(&usercontroller.Empty{}, mockStream)
	action(controller, model, mockCtrl, mockStream, sharedContext, &wg)

	wg.Wait()

	// Assert
	mockCtrl.Finish()
}
