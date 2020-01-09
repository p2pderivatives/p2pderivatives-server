package userservice_test

import (
	"context"
	"testing"

	"p2pderivatives-server/internal/common/servererror"
	"p2pderivatives-server/internal/common/token"
	"p2pderivatives-server/internal/user/usercommon"
	"p2pderivatives-server/internal/user/userservice"
	"p2pderivatives-server/test"
	"p2pderivatives-server/test/mocks/mock_userrepository"

	"github.com/stretchr/testify/assert"
)

func createRepoAndService() (repo *mock_userrepository.RepositoryMock, service *userservice.Service) {
	repo = mock_userrepository.NewRepositoryMock()
	config := usercommon.DefaultUserConfiguration()
	service = userservice.NewService(repo, config, &servererror.ServiceError{})
	return
}

func initToken() {
	tokenConfig := token.Config{}
	conf := test.GetTestConfig()
	conf.InitializeComponentConfig(&tokenConfig)
	token.Init(&tokenConfig)
}

func TestService_FindFirstUser(t *testing.T) {
	repo, service := createRepoAndService()
	ctx := context.Background()

	model := &usercommon.User{
		ID:       "id1",
		Name:     "hoge_taro1",
		Password: "Pass1",
	}

	condition := &usercommon.User{Name: "hoge_taro1"}
	_, err := service.FindFirstUser(ctx, condition, []string{"id"})

	assert.NotNil(t, err)

	insertTarget := append(make([]*usercommon.User, 0), model)

	mock_userrepository.InsertTestUserData(repo, insertTarget)

	result2, _ := service.FindFirstUser(ctx, condition, []string{"id"})

	assert.Equal(t, result2.ID, "id1")
	assert.Equal(t, result2.Name, "hoge_taro1")
	assert.Equal(t, result2.Password, "Pass1")
}

func TestService_FindUsers(t *testing.T) {
	repo, service := createRepoAndService()
	ctx := context.Background()

	user1 := &usercommon.User{
		ID:       "id1",
		Name:     "hoge_taro1",
		Password: "Pass1",
	}
	user2 := &usercommon.User{
		ID:       "id2",
		Name:     "hoge_taro2",
		Password: "Pass2",
	}
	user3 := &usercommon.User{
		ID:       "id3",
		Name:     "hoge_taro3",
		Password: "Pass3",
	}
	user4 := &usercommon.User{
		ID:       "id4",
		Name:     "hoge_jiro",
		Password: "Pass4",
	}
	user5 := &usercommon.User{
		ID:       "id5",
		Name:     "hoge_taro5",
		Password: "Pass5",
	}

	insertTarget := append(
		make([]*usercommon.User, 0),
		user1,
		user2,
		user3,
		user4,
		user5,
	)

	mock_userrepository.InsertTestUserData(repo, insertTarget)

	condition := &usercommon.User{}

	results, _ := service.FindUsers(ctx, condition, 0, 100, []string{"id"})

	assert.Equal(t, len(results), 5)
	assert.Equal(t, results[0].ID, "id1")
	assert.Equal(t, results[0].Name, "hoge_taro1")
	assert.Equal(t, results[0].Password, "Pass1")

	assert.Equal(t, results[1].ID, "id2")
	assert.Equal(t, results[1].Name, "hoge_taro2")
	assert.Equal(t, results[1].Password, "Pass2")

	assert.Equal(t, results[2].ID, "id3")
	assert.Equal(t, results[2].Name, "hoge_taro3")
	assert.Equal(t, results[2].Password, "Pass3")

	assert.Equal(t, results[4].ID, "id5")
	assert.Equal(t, results[4].Name, "hoge_taro5")
	assert.Equal(t, results[4].Password, "Pass5")

	// test with offset
	results2, _ := service.FindUsers(ctx, condition, 2, 100, []string{"id"})

	assert.Equal(t, len(results2), 3)
	assert.Equal(t, results2[0].ID, "id3")
	assert.Equal(t, results2[0].Name, "hoge_taro3")
	assert.Equal(t, results2[0].Password, "Pass3")

	assert.Equal(t, results2[2].ID, "id5")
	assert.Equal(t, results2[2].Name, "hoge_taro5")
	assert.Equal(t, results2[2].Password, "Pass5")

	// test with limit
	results3, _ := service.FindUsers(ctx, condition, 0, 2, []string{"id"})

	assert.Equal(t, results3[0].ID, "id1")
	assert.Equal(t, results3[0].Name, "hoge_taro1")
	assert.Equal(t, results3[0].Password, "Pass1")

	assert.Equal(t, results3[1].ID, "id2")
	assert.Equal(t, results3[1].Name, "hoge_taro2")
	assert.Equal(t, results3[1].Password, "Pass2")
}

func TestService_CreateUser(t *testing.T) {
	_, service := createRepoAndService()

	user := &usercommon.User{
		ID:       "id1",
		Name:     "hoge_taro1",
		Password: "Pass1",
	}

	createdUser, _ := service.CreateUser(context.Background(), user)

	assert.Equal(t, createdUser.ID, "id1")
	assert.Equal(t, createdUser.Name, "hoge_taro1")
}

func TestService_UpdateUser(t *testing.T) {
	repo, service := createRepoAndService()
	ctx := context.Background()

	orgUser := usercommon.NewUser("hoge_taro", "password1")

	insertTarget := append(make([]*usercommon.User, 0), orgUser)
	mock_userrepository.InsertTestUserData(repo, insertTarget)

	condition := &usercommon.User{Name: "hoge_taro"}
	orgResults, _ := service.FindUsers(ctx, condition, 0, 100, []string{"id"})

	assert.Equal(t, len(orgResults), 1)
	assert.NotNil(t, orgResults[0].ID)
	assert.Equal(t, orgResults[0].Name, "hoge_taro")
	assert.NotNil(t, orgResults[0].Password)

	expected := &usercommon.User{
		ID:                    orgResults[0].ID,
		Name:                  "piyo_taro",
		Password:              orgResults[0].Password,
		RequireChangePassword: orgResults[0].RequireChangePassword,
	}

	actual, _ := service.UpdateUser(ctx, &usercommon.User{
		ID:                    expected.ID,
		Name:                  expected.Name,
		RequireChangePassword: expected.RequireChangePassword,
	})

	assert.Equal(t, expected, actual)

	updatedResults, _ := service.FindUsers(
		ctx,
		&usercommon.User{ID: actual.ID},
		0,
		100,
		[]string{"id"},
	)

	assert.Equal(t, len(updatedResults), 1)
	assert.NotNil(t, updatedResults[0].ID)
	assert.Equal(t, updatedResults[0].Name, "piyo_taro")
	assert.NotNil(t, updatedResults[0].Password)
}

func TestService_DeleteUser(t *testing.T) {
	repo, service := createRepoAndService()
	ctx := context.Background()

	orgUser := usercommon.NewUser("hoge_taro", "password1")

	insertTarget := append(make([]*usercommon.User, 0), orgUser)
	mock_userrepository.InsertTestUserData(repo, insertTarget)

	condition := &usercommon.User{Name: "hoge_taro"}
	orgResults, _ := service.FindUsers(ctx, condition, 0, 100, []string{"id"})

	assert.Equal(t, len(orgResults), 1)
	assert.NotNil(t, orgResults[0].ID)
	assert.Equal(t, orgResults[0].Name, "hoge_taro")
	assert.NotNil(t, orgResults[0].Password)

	err1 := service.DeleteUser(ctx, &usercommon.User{ID: ""})

	assert.Nil(t, err1)

	orgResults2, _ := service.FindUsers(ctx, condition, 0, 100, []string{"id"})

	assert.Equal(t, len(orgResults2), 1)
	assert.NotNil(t, orgResults2[0].ID)
	assert.Equal(t, orgResults2[0].Name, "hoge_taro")
	assert.NotNil(t, orgResults2[0].Password)

	err2 := service.DeleteUser(ctx, &usercommon.User{ID: orgResults2[0].ID})

	assert.Nil(t, err2)

	deletedFindResults, _ := service.FindUsers(ctx, condition, 0, 100, []string{"id"})

	assert.Equal(t, len(deletedFindResults), 0)
}

func TestService_ChangeUserPassword(t *testing.T) {
	assert := assert.New(t)
	_, service := createRepoAndService()
	ctx := context.Background()

	const (
		name        = "name1"
		oldPassword = "oldP@ssw0rd"
		newPassword = "newP@ssw0rd"
	)
	orgUser := usercommon.NewUser(name, oldPassword)
	orgUser, _ = service.CreateUser(ctx, orgUser)

	newPasswordUser, err := service.ChangeUserPassword(
		ctx,
		name,
		oldPassword,
		newPassword,
	)

	assert.NoError(err)
	assert.NotEqual(orgUser, newPasswordUser)
}

func TestService_ResetUserPassword(t *testing.T) {
	_, service := createRepoAndService()
	ctx := context.Background()

	const (
		name        = "name1"
		oldPassword = "oldP@ssw0rd"
		newPassword = "newP@ssw0rd"
	)
	orgUser := usercommon.NewUser(name, oldPassword)
	service.CreateUser(ctx, orgUser)

	newPasswordUser, err := service.ResetUserPassword(
		ctx,
		name,
		newPassword,
	)

	assert.NoError(t, err)
	assert.True(t, newPasswordUser.RequireChangePassword)
}

func TestServiceAuthenticateUser_WithValidPassword_Succeeds(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	service, ctx := initTestHelper()

	// Act
	actual, tokenInfo, err := service.AuthenticateUser(ctx, name, password)

	// Assert
	assert.NoError(err)
	assert.NotNil(tokenInfo)
	assert.NotNil(actual)
}

func TestServiceAuthenticateUser_WithInvalidPassword_Error(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	service, ctx := initTestHelper()

	// Act
	actual, tokenInfo, err := service.AuthenticateUser(ctx, name, badPassword)

	// Assert
	assert.Error(err)
	assert.Nil(tokenInfo)
	assert.Nil(actual)
}

func TestRevokeRefreshToken_WithCorrectToken_IsRevoked(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	service, ctx := initTestHelper()
	_, tokenInfo, _ := service.AuthenticateUser(ctx, name, password)

	// Act
	err := service.RevokeRefreshToken(ctx, tokenInfo.RefreshToken)
	user, _ := service.FindFirstUserByName(ctx, name)

	// Assert
	assert.NoError(err)
	assert.Empty(user.RefreshToken)
}

func TestRefreshUserToken_WithCorrectRefreshToken_IsRefreshed(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	service, ctx := initTestHelper()
	orgUser, tokenInfo, _ := service.AuthenticateUser(ctx, name, password)

	// Act
	refreshedTokenInfo, err := service.RefreshUserToken(ctx, tokenInfo.RefreshToken)
	refreshedUser, _ := service.FindFirstUserByName(ctx, name)

	// Assert
	assert.NoError(err)
	assert.NotEqual(tokenInfo, refreshedTokenInfo)
	assert.NotEqual(orgUser.RefreshToken, refreshedUser.RefreshToken)
}

const (
	name        = "test"
	password    = "p@assw0rd"
	badPassword = "p@ssword"
)

func initTestHelper() (*userservice.Service, context.Context) {
	_, service := createRepoAndService()
	model := usercommon.NewUser(name, password)
	ctx := context.Background()
	service.CreateUser(ctx, model)
	initToken()
	return service, ctx
}
