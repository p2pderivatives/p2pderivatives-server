package usercontroller_test

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"

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
	user := createUser()
	userRegisterRequest := createUserRegisterRequest(user)
	userInfo := usercontroller.UserInfo{
		Name: userRegisterRequest.Name,
	}
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	stream := mock_usercontroller.NewMockUser_GetUserListServer(mockCtrl)
	stream.EXPECT().Send(&userInfo).Times(1)
	ctxId := contexts.SetUserID(ctx, user.ID)
	stream.EXPECT().Context().Return(ctxId).Times(1)

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

func TestSendReceive_MessageIsReceived(t *testing.T) {
	controller := createController()
	var wg sync.WaitGroup
	wg.Add(1)
	defer controller.Close()
	modelUser1 := createUser()
	modelUser2 := createUser()
	ctx := context.Background()
	message := &usercontroller.DlcMessage{
		DestName: modelUser1.Name,
		Payload:  []byte("Hello"),
		OrgName:  modelUser2.Name,
	}
	mockCtrl := gomock.NewController(t)
	mockStream := mock_usercontroller.NewMockUser_ReceiveDlcMessagesServer(mockCtrl)
	mockStream.EXPECT().Send(message).Times(1).Do(
		func(message *usercontroller.DlcMessage) {
			wg.Done()
		})
	response, _ := controller.RegisterUser(
		ctx, createUserRegisterRequest(modelUser1))
	id1 := response.Id
	ctx1 := contexts.SetUserID(ctx, id1)
	mockStream.EXPECT().Context().Return(ctx1).AnyTimes()
	response, _ = controller.RegisterUser(
		ctx, createUserRegisterRequest(modelUser2))
	id2 := response.Id
	ctx2 := contexts.SetUserID(ctx, id2)

	// Act
	go controller.ReceiveDlcMessages(&usercontroller.Empty{}, mockStream)
	time.Sleep(time.Millisecond * 5)
	_, err := controller.SendDlcMessage(ctx2, message)

	wg.Wait()

	// Assert
	assert.NoError(t, err)
	mockCtrl.Finish()
}

func TestSendMessage_NoReceiver_Error(t *testing.T) {
	controller := createController()
	defer controller.Close()
	modelUser1 := createUser()
	modelUser2 := createUser()
	ctx := context.Background()
	message := &usercontroller.DlcMessage{
		DestName: modelUser1.Name,
		Payload:  []byte("Hello"),
		OrgName:  modelUser2.Name,
	}
	response, _ := controller.RegisterUser(
		ctx, createUserRegisterRequest(modelUser1))
	response, _ = controller.RegisterUser(
		ctx, createUserRegisterRequest(modelUser2))
	id2 := response.Id
	ctx2 := contexts.SetUserID(ctx, id2)

	// Act
	_, err := controller.SendDlcMessage(ctx2, message)

	// Assert
	assert.Error(t, err)
}

func TestSendReceive_MultipleReceiver_MessageIsReceived(t *testing.T) {
	controller := createController()
	var wg sync.WaitGroup
	wg.Add(2)
	defer controller.Close()
	modelUser1 := createUser()
	modelUser2 := createUser()
	ctx := context.Background()
	message := &usercontroller.DlcMessage{
		DestName: modelUser1.Name,
		Payload:  []byte("Hello"),
		OrgName:  modelUser2.Name,
	}
	mockCtrl := gomock.NewController(t)
	mockStream1 := mock_usercontroller.NewMockUser_ReceiveDlcMessagesServer(mockCtrl)
	mockStream1.EXPECT().Send(message).Times(1).Do(
		func(message *usercontroller.DlcMessage) {
			wg.Done()
		})
	mockStream2 := mock_usercontroller.NewMockUser_ReceiveDlcMessagesServer(mockCtrl)
	mockStream2.EXPECT().Send(message).Times(1).Do(
		func(message *usercontroller.DlcMessage) {
			wg.Done()
		})
	response, _ := controller.RegisterUser(
		ctx, createUserRegisterRequest(modelUser1))
	id1 := response.Id
	ctx1 := contexts.SetUserID(ctx, id1)
	mockStream1.EXPECT().Context().Return(ctx1).AnyTimes()
	mockStream2.EXPECT().Context().Return(ctx1).AnyTimes()
	response, _ = controller.RegisterUser(
		ctx, createUserRegisterRequest(modelUser2))
	id2 := response.Id
	ctx2 := contexts.SetUserID(ctx, id2)

	// Act
	go controller.ReceiveDlcMessages(&usercontroller.Empty{}, mockStream1)
	go controller.ReceiveDlcMessages(&usercontroller.Empty{}, mockStream2)
	time.Sleep(time.Millisecond * 5)
	_, err := controller.SendDlcMessage(ctx2, message)

	wg.Wait()

	// Assert
	assert.NoError(t, err)
	mockCtrl.Finish()
}

func TestSendMessage_ToClosedStream_Error(t *testing.T) {
	controller := createController()
	var wg sync.WaitGroup
	wg.Add(1)
	defer controller.Close()
	modelUser1 := createUser()
	modelUser2 := createUser()
	ctx := context.Background()
	message := &usercontroller.DlcMessage{
		DestName: modelUser1.Name,
		Payload:  []byte("Hello"),
		OrgName:  modelUser2.Name,
	}
	mockCtrl := gomock.NewController(t)
	mockStream := mock_usercontroller.NewMockUser_ReceiveDlcMessagesServer(mockCtrl)
	mockStream.EXPECT().Send(message).Times(1).DoAndReturn(
		func(message *usercontroller.DlcMessage) error {
			wg.Done()
			return errors.New("Error")
		})
	response, _ := controller.RegisterUser(
		ctx, createUserRegisterRequest(modelUser1))
	id1 := response.Id
	ctx1 := contexts.SetUserID(ctx, id1)
	mockStream.EXPECT().Context().Return(ctx1).AnyTimes()
	response, _ = controller.RegisterUser(
		ctx, createUserRegisterRequest(modelUser2))
	id2 := response.Id
	ctx2 := contexts.SetUserID(ctx, id2)

	// Act
	go controller.ReceiveDlcMessages(&usercontroller.Empty{}, mockStream)
	time.Sleep(time.Millisecond * 5)
	_, err := controller.SendDlcMessage(ctx2, message)

	wg.Wait()

	// Assert
	assert.Error(t, err)
	mockCtrl.Finish()
}

func TestGetConnectedUsers_ReturnsConnectedUsers(t *testing.T) {
	controller := createController()
	var wg sync.WaitGroup
	wg.Add(2)
	defer controller.Close()
	modelUser1 := createUser()
	modelUser2 := createUser()
	modelUser3 := createUser()
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	mockStream1 := mock_usercontroller.NewMockUser_ReceiveDlcMessagesServer(mockCtrl)
	mockStream2 := mock_usercontroller.NewMockUser_ReceiveDlcMessagesServer(mockCtrl)
	mockStream3 := mock_usercontroller.NewMockUser_GetConnectedUsersServer(mockCtrl)
	response, _ := controller.RegisterUser(
		ctx, createUserRegisterRequest(modelUser1))
	id1 := response.Id
	ctx1 := contexts.SetUserID(ctx, id1)
	mockStream1.EXPECT().Context().Return(ctx1).AnyTimes()
	response, _ = controller.RegisterUser(
		ctx, createUserRegisterRequest(modelUser2))
	id2 := response.Id
	ctx2 := contexts.SetUserID(ctx, id2)
	mockStream2.EXPECT().Context().Return(ctx2).AnyTimes()
	response, _ = controller.RegisterUser(
		ctx, createUserRegisterRequest(modelUser3))
	id3 := response.Id
	ctx3 := contexts.SetUserID(ctx, id3)
	mockStream3.EXPECT().Context().Return(ctx3).AnyTimes()

	// mock pings
	mockStream1.EXPECT().Send(&usercontroller.DlcMessage{DestName: modelUser1.Name}).Times(1)
	mockStream2.EXPECT().Send(&usercontroller.DlcMessage{DestName: modelUser2.Name}).Times(1)

	userInfo1 := &usercontroller.UserInfo{Name: modelUser1.Name}
	userInfo2 := &usercontroller.UserInfo{Name: modelUser2.Name}

	mockStream3.EXPECT().Send(userInfo1).Times(1)
	mockStream3.EXPECT().Send(userInfo2).Times(1)

	// Act
	go controller.ReceiveDlcMessages(&usercontroller.Empty{}, mockStream1)
	go controller.ReceiveDlcMessages(&usercontroller.Empty{}, mockStream2)
	time.Sleep(time.Millisecond * 100)
	err := controller.GetConnectedUsers(&usercontroller.Empty{}, mockStream3)
	time.Sleep(time.Millisecond * 100)

	// Assert
	assert.NoError(t, err)
	mockCtrl.Finish()
}
