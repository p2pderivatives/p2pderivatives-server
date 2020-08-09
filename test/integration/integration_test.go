// +build integration

package integration

import (
	"context"
	"io"
	"log"
	"p2pderivatives-server/internal/authentication"
	"p2pderivatives-server/internal/common/token"
	"p2pderivatives-server/internal/user/usercommon"
	"p2pderivatives-server/internal/user/usercontroller"
	"path/filepath"
	"sync"
	"testing"
	"time"

	conf "github.com/cryptogarageinc/server-common-go/pkg/configuration"

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
	Name:     "Name1",
	Password: "P@ssw0rd1",
}

var user2 = &usercommon.User{
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

	_, authClient3, err := getClients(serverAddress)

	if err != nil {
		assert.FailNow("Could not get clients.")
	}

	assertUserRegistration(assert, userClient1, user1)
	assertUserRegistration(assert, userClient2, user2)

	accessToken1, _ := assertLogin(assert, authClient1, user1)
	accessToken2, _ := assertLogin(assert, authClient2, user2)
	assertFailedLogin(assert, authClient3)

	assertClientList(
		assert, userClient1, accessToken1, []string{user2.Name})

	assertMessaging(
		assert, userClient1, userClient2, user1, user2, accessToken1, accessToken2)

	assertUpdatePassword(assert, authClient1, accessToken1)

	assertUserUnregister(assert, userClient2, user2, accessToken2)

	assertClientList(
		assert, userClient1, accessToken1, []string{})

	assertGetConnectedUser(assert, serverAddress, userClient1, accessToken1)

	assertValidation(assert, authClient1)
}

func assertUserRegistration(
	assert *assert.Assertions, userClient usercontroller.UserClient, model *usercommon.User) {
	registerRequest := &usercontroller.UserRegisterRequest{
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
		Name:     model.Name,
		Password: model.Password,
	}

	loginResponse, err := authClient.Login(context.Background(), loginRequest)

	assert.NoError(err)
	accessToken = loginResponse.Token.AccessToken
	refreshToken = loginResponse.Token.RefreshToken
	return
}

func assertFailedLogin(
	assert *assert.Assertions,
	authClient authentication.AuthenticationClient) {
	loginRequest := &authentication.LoginRequest{
		Name:     "doesntexist",
		Password: "password",
	}
	loginResponse, err := authClient.Login(context.Background(), loginRequest)
	assert.Nil(loginResponse)
	statusErr, ok := status.FromError(err)
	assert.True(ok)
	if ok {
		assert.Equal(codes.Unauthenticated, statusErr.Code())
		assert.Equal("Fail to authenticate user.", statusErr.Message())
		assert.Empty(statusErr.Details())
	}
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

func assertMessaging(
	assert *assert.Assertions,
	userClient1 usercontroller.UserClient,
	userClient2 usercontroller.UserClient,
	user1 *usercommon.User,
	user2 *usercommon.User,
	accessToken1 string,
	accessToken2 string) {
	ctx1 := metadata.AppendToOutgoingContext(
		context.Background(), token.MetaKeyAuthentication, accessToken1)

	ctx2 := metadata.AppendToOutgoingContext(
		context.Background(), token.MetaKeyAuthentication, accessToken2)

	payload := []byte("Hello")
	message := &usercontroller.DlcMessage{Payload: payload, DestName: user1.Name}
	var receivedMessage *usercontroller.DlcMessage
	var err1, err2 error

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		var stream usercontroller.User_ReceiveDlcMessagesClient
		ctx1, cancel := context.WithCancel(ctx1)
		stream, err1 = userClient1.ReceiveDlcMessages(ctx1, &usercontroller.Empty{})

		if err1 != nil {
			wg.Done()
			cancel()
			return
		}

		wg.Done()

		receivedMessage, err1 = stream.Recv()
		cancel()
		wg.Done()
	}()

	time.Sleep(time.Millisecond * 5)
	wg.Wait()

	assert.NoError(err1)

	if err1 != nil {
		return
	}

	wg.Add(1)

	_, err2 = userClient2.SendDlcMessage(ctx2, message)

	wg.Wait()

	assert.NoError(err1)
	assert.NoError(err2)
	assert.Equal(user2.Name, receivedMessage.OrgName)
	assert.Equal(payload, receivedMessage.Payload)
	assert.Equal(user1.Name, receivedMessage.DestName)

	_, err2 = userClient2.SendDlcMessage(ctx2, message)

	assert.Error(err2)
}

func assertUpdatePassword(
	assert *assert.Assertions, authClient authentication.AuthenticationClient, accessToken string) {

	ctx := metadata.AppendToOutgoingContext(
		context.Background(), token.MetaKeyAuthentication, accessToken)
	newPassword := "P@ssw0rd"
	request := &authentication.UpdatePasswordRequest{
		NewPassword: newPassword,
		OldPassword: user1.Password,
	}
	_, err := authClient.UpdatePassword(ctx, request)

	assert.NoError(err)

	loginRequest := &authentication.LoginRequest{Name: user1.Name, Password: newPassword}

	_, err = authClient.Login(context.Background(), loginRequest)

	assert.NoError(err)

	loginRequest.Password = user1.Password

	_, err = authClient.Login(context.Background(), loginRequest)

	assert.Error(err)
}

func assertGetConnectedUser(assert *assert.Assertions, serverAddress string, client usercontroller.UserClient, accessToken string) {
	// arrange
	ctx := metadata.AppendToOutgoingContext(
		context.Background(), token.MetaKeyAuthentication, accessToken)
	connToShutdown, err := getConnection(serverAddress)
	assert.NoError(err)

	shutdownUserClient := usercontroller.NewUserClient(connToShutdown)
	shutdownAuthClient := authentication.NewAuthenticationClient(connToShutdown)
	shutdownUser := &usercommon.User{
		Name:     "NameShutdown",
		Password: "P@ssword0",
	}
	user3 := &usercommon.User{
		Name:     "Client3",
		Password: "P@ssword0",
	}
	client3, auth3, err := getClients(serverAddress)
	assert.NoError(err)
	assertUserRegistration(assert, shutdownUserClient, shutdownUser)
	assertUserRegistration(assert, client3, user3)

	shutdownAccessToken, _ := assertLogin(assert, shutdownAuthClient, shutdownUser)
	user3AccessToken, _ := assertLogin(assert, auth3, user3)
	assertClientList(
		assert, client, accessToken, []string{shutdownUser.Name, user3.Name})
	ctxShutdown := metadata.AppendToOutgoingContext(
		context.Background(), token.MetaKeyAuthentication, shutdownAccessToken)
	ctx3 := metadata.AppendToOutgoingContext(
		context.Background(), token.MetaKeyAuthentication, user3AccessToken)

	_, err = shutdownUserClient.ReceiveDlcMessages(ctxShutdown, &usercontroller.Empty{})
	assert.NoError(err)
	_, err = client3.ReceiveDlcMessages(ctx3, &usercontroller.Empty{})
	assert.NoError(err)

	time.Sleep(10 * time.Millisecond)
	userNames := getConnectedUser(assert, client, ctx)
	assert.Contains(userNames, shutdownUser.Name)
	assert.Contains(userNames, user3.Name)

	// act
	err = connToShutdown.Close()

	// assert
	assert.NoError(err)
	userNames = getConnectedUser(assert, client, ctx)
	assert.NotContains(userNames, shutdownUser.Name)
	assert.Contains(userNames, user3.Name)
}

func assertValidation(
	assert *assert.Assertions, authClient authentication.AuthenticationClient) {
	loginRequest := &authentication.LoginRequest{
		Name:     "",
		Password: "",
	}

	_, err := authClient.Login(context.Background(), loginRequest)

	assert.Error(err)
	errStatus, _ := status.FromError(err)
	assert.Equal(codes.InvalidArgument, errStatus.Code())
	assert.Equal("invalid field Name: value '' must not be an empty string", errStatus.Message())
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

func getConnectedUser(assert *assert.Assertions, client usercontroller.UserClient, ctx context.Context) []string {
	stream, err := client.GetConnectedUsers(ctx, &usercontroller.Empty{})
	assert.NoError(err)
	userNames := make([]string, 0)

	for {
		userInfo, err := stream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			assert.Fail("Error while receiving getConnectedUserStream %v", err)
		}

		userNames = append(userNames, userInfo.Name)
	}
	return userNames
}
