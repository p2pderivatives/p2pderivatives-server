// +build integration_test

package integration

import (
	"context"
	"io"
	"log"
	"p2pderivatives-server/internal/authentication"
	"p2pderivatives-server/internal/common/conf"
	"p2pderivatives-server/internal/common/token"
	"p2pderivatives-server/internal/user/usercommon"
	"p2pderivatives-server/internal/user/usercontroller"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	appName = "p2pd"
	envName = "integration"
)

var user1 = &usercommon.User{
	Account:  "Account1",
	Name:     "Name1",
	Password: "P@ssw0rd1",
}

var user2 = &usercommon.User{
	Account:  "Account2",
	Name:     "Name2",
	Password: "P@ssw2rd2",
}

func TestIntegration(t *testing.T) {
	assert := assert.New(t)
	config := conf.NewConfiguration(
		appName,
		envName,
		[]string{filepath.Join("..", "config")})
	config.Initialize()
	serverAddress := config.GetString("server.address")

	userClient1, authClient1, err := getClients(serverAddress)

	if err != nil {
		assert.FailNow("Could not get clients.")
	}

	userClient2, authClient2, err := getClients(serverAddress)

	if err != nil {
		assert.FailNow("Could not get clients.")
	}

	assertUserRegistration(assert, userClient1, user1)
	assertUserRegistration(assert, userClient2, user2)

	accessToken1, _ := assertLogin(assert, authClient1, user1)
	accessToken2, _ := assertLogin(assert, authClient2, user2)

	assertClientList(
		assert, userClient1, accessToken1, []string{user1.Name, user2.Name})

	assertUserStatuses(assert, userClient1, userClient2, user1, user2, accessToken1, accessToken2)

	assertUserUnregister(assert, userClient2, user2, accessToken2)

	assertClientList(
		assert, userClient1, accessToken1, []string{user1.Name})

}

func assertUserRegistration(
	assert *assert.Assertions, userClient usercontroller.UserClient, model *usercommon.User) {
	registerRequest := &usercontroller.UserRegisterRequest{
		Account:  model.Account,
		Name:     model.Name,
		Password: model.Password,
	}

	_, err := userClient.RegisterUser(context.Background(), registerRequest)

	assert.NoError(err)
}

func assertUserUnregister(
	assert *assert.Assertions,
	userClient usercontroller.UserClient,
	model *usercommon.User,
	accessToken string) {
	unregisterRequest := &usercontroller.UnregisterUserRequest{}

	ctx := metadata.AppendToOutgoingContext(
		context.Background(), token.MetaKeyAuthentication, accessToken)
	_, err := userClient.UnregisterUser(ctx, unregisterRequest)

	assert.NoError(err)
}

func assertLogin(
	assert *assert.Assertions,
	authClient authentication.AuthenticationClient,
	model *usercommon.User) (accessToken string, refreshToken string) {

	loginRequest := &authentication.LoginRequest{
		Account:  model.Account,
		Password: model.Password,
	}

	loginResponse, err := authClient.Login(context.Background(), loginRequest)

	assert.NoError(err)
	accessToken = loginResponse.Token.AccessToken
	refreshToken = loginResponse.Token.RefreshToken
	return
}

func assertClientList(
	assert *assert.Assertions, userClient usercontroller.UserClient, accessToken string, expectedList []string) {
	ctx := metadata.AppendToOutgoingContext(
		context.Background(), token.MetaKeyAuthentication, accessToken)

	stream, err := userClient.GetUserList(ctx, &usercontroller.Empty{})

	if err != nil {
		assert.Fail("Could not get client list.")
		return
	}

	userNames := make([]string, 0)

	for {
		userInfo, err := stream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("%v.GetUserList(_) = _, %v", userClient, err)
		}

		userNames = append(userNames, userInfo.Name)
	}

	assert.Len(userNames, len(expectedList))

	for _, name := range userNames {
		assert.Contains(expectedList, name)
	}
}

func assertUserStatuses(
	assert *assert.Assertions,
	userClient1 usercontroller.UserClient,
	userClient2 usercontroller.UserClient,
	user1 *usercommon.User,
	user2 *usercommon.User,
	accessToken1 string,
	accessToken2 string) {
	ctx, cancel := context.WithCancel(context.Background())

	var userNotices1 []*usercontroller.UserNotice
	var userNotices2 []*usercontroller.UserNotice
	var err1 error
	var err2 error
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		userNotices1, err1 = getUserStatuses(ctx, userClient1, accessToken1, &wg)
	}()
	go func() {
		userNotices2, err2 = getUserStatuses(ctx, userClient2, accessToken2, &wg)
	}()

	time.Sleep(5 * time.Second)
	cancel()
	wg.Wait()

	grpcStatus1, ok1 := status.FromError(err1)
	grpcStatus2, ok2 := status.FromError(err2)

	assert.True(ok1)
	assert.True(ok2)
	assert.Equal(codes.Canceled, grpcStatus1.Code())
	assert.Equal(codes.Canceled, grpcStatus2.Code())

	assert.Contains(userNotices1, &usercontroller.UserNotice{
		Name:   user2.Name,
		Status: usercontroller.UserStatus_CONNECTED,
	})

	assert.Contains(userNotices2, &usercontroller.UserNotice{
		Name:   user1.Name,
		Status: usercontroller.UserStatus_CONNECTED,
	})
}

func getUserStatuses(
	ctx context.Context,
	userClient usercontroller.UserClient,
	accessToken string,
	wg *sync.WaitGroup) ([]*usercontroller.UserNotice, error) {
	ctx = metadata.AppendToOutgoingContext(
		ctx, token.MetaKeyAuthentication, accessToken)
	stream, err := userClient.GetUserStatuses(ctx, &usercontroller.Empty{})
	if err != nil {
		wg.Done()
		return nil, err
	}

	notices := make([]*usercontroller.UserNotice, 0)

	for {
		notice, err := stream.Recv()

		if err != nil {
			wg.Done()
			return notices, err
		}

		notices = append(notices, notice)
	}
}

func getConnection(serverAddress string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	opts = append(opts, grpc.WithInsecure())

	return grpc.Dial(serverAddress, opts...)
}

func getClients(serverAddress string) (usercontroller.UserClient, authentication.AuthenticationClient, error) {
	conn, err := getConnection(serverAddress)

	if err != nil {
		return nil, nil, err
	}

	userClient := usercontroller.NewUserClient(conn)
	authClient := authentication.NewAuthenticationClient(conn)

	return userClient, authClient, nil
}
