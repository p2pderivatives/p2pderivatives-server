package userrepository

import (
	context "context"
	"fmt"
	"testing"

	"p2pderivatives-server/internal/database/orm"
	"p2pderivatives-server/internal/user/usercommon"
	"p2pderivatives-server/test"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

// createTxAndUserRepo creates a new DB transaction and user repository.
func createContextRepoAndTx() (ctx context.Context, repo *Repository, tx *gorm.DB) {
	ormInstance := test.InitializeORM(&usercommon.User{})
	tx = ormInstance.GetDB().Begin()
	ctx = orm.SaveTx(context.Background(), tx)
	repo = NewRepository()
	return
}

func insertTestDataToTx(tx *gorm.DB, models []*usercommon.User) (err error) {
	for _, model := range models {
		err = tx.Create(model).Error
		if err != nil {
			return
		}
	}
	return
}

func TestRepository_FindFirstUser(t *testing.T) {
	ctx, repo, tx := createContextRepoAndTx()

	user := &usercommon.User{
		ID:       "id1",
		Account:  "account1",
		Name:     "hoge_taro1",
		Password: "Pass1",
	}

	insertTarget := append(make([]*usercommon.User, 0), user)
	insertTestDataToTx(tx, insertTarget)

	condition := usercommon.User{Account: "account1"}

	result, _ := repo.FindFirstUser(ctx, condition, []string{"id"})

	assert.Equal(t, result.ID, "id1")
	assert.Equal(t, result.Account, "account1")
	assert.Equal(t, result.Name, "hoge_taro1")
	assert.Equal(t, result.Password, "Pass1")
}

func TestAddressRepository_CountUsers(t *testing.T) {
	ctx, repo, tx := createContextRepoAndTx()
	defer tx.Rollback()

	user1 := usercommon.NewUser("account1", "hoge_taro", "password1")
	user2 := usercommon.NewUser("account2", "hoge_jiro", "password2")
	user3 := usercommon.NewUser("account3", "hoge_taro2", "password3")

	insertTarget := append(
		make([]*usercommon.User, 0),
		user1,
		user2,
		user3,
	)

	_ = insertTestDataToTx(tx, insertTarget)

	condition := usercommon.User{}
	result, _ := repo.CountUsers(ctx, condition)

	assert.Equal(t, result, 3)
}

func TestRepository_CreateUser(t *testing.T) {
	ctx, repo, tx := createContextRepoAndTx()
	defer tx.Rollback()

	condition := usercommon.User{Name: "hoge_taro"}
	notfoundResults, _ := repo.FindUsers(ctx, condition, 0, 100, []string{"id"})

	assert.Equal(t, len(notfoundResults), 0)

	user := usercommon.NewUser("account1", "hoge_taro", "password1")

	_ = repo.CreateUser(ctx, user)
	foundResults, _ := repo.FindUsers(ctx, condition, 0, 100, []string{"id"})

	assert.Equal(t, len(foundResults), 1)
	assert.NotNil(t, foundResults[0].ID)
	assert.Equal(t, foundResults[0].Account, "account1")
	assert.Equal(t, foundResults[0].Name, "hoge_taro")
	assert.NotNil(t, foundResults[0].Password)
}

func TestRepository_UpdateUser(t *testing.T) {
	ctx, repo, tx := createContextRepoAndTx()
	defer tx.Rollback()

	orgUser := usercommon.NewUser("account1", "hoge_taro", "password1")

	insertTarget := append(make([]*usercommon.User, 0), orgUser)
	_ = insertTestDataToTx(tx, insertTarget)

	condition := usercommon.User{Name: "hoge_taro"}
	orgResults, _ := repo.FindUsers(ctx, condition, 0, 100, []string{"id"})

	assert.Equal(t, len(orgResults), 1)
	assert.NotNil(t, orgResults[0].ID)
	assert.Equal(t, orgResults[0].Account, "account1")
	assert.Equal(t, orgResults[0].Name, "hoge_taro")
	assert.NotNil(t, orgResults[0].Password)

	updateUser := orgResults[0]
	updateUser.Name = "piyo_taro"

	_ = repo.UpdateUser(ctx, &updateUser)

	updatedResults, _ := repo.FindUsers(
		ctx,
		usercommon.User{ID: updateUser.ID},
		0,
		100,
		[]string{"id"},
	)

	assert.Equal(t, len(updatedResults), 1)
	assert.NotNil(t, updatedResults[0].ID)
	assert.Equal(t, updatedResults[0].Account, "account1")
	assert.Equal(t, updatedResults[0].Name, "piyo_taro")
	assert.NotNil(t, updatedResults[0].Password)
}

func TestRepository_DeleteUser(t *testing.T) {
	ctx, repo, tx := createContextRepoAndTx()
	defer tx.Rollback()

	orgUser := usercommon.NewUser("account1", "hoge_taro", "password1")

	insertTarget := append(make([]*usercommon.User, 0), orgUser)
	_ = insertTestDataToTx(tx, insertTarget)

	condition := usercommon.User{Name: "hoge_taro"}
	orgResults, _ := repo.FindUsers(ctx, condition, 0, 100, []string{"id"})

	assert.Equal(t, len(orgResults), 1)
	assert.NotNil(t, orgResults[0].ID)
	assert.Equal(t, orgResults[0].Account, "account1")
	assert.Equal(t, orgResults[0].Name, "hoge_taro")
	assert.NotNil(t, orgResults[0].Password)

	// If no key is specified, no deletion occurs.
	_ = repo.DeleteUser(ctx, &usercommon.User{ID: ""})

	orgResults2, _ := repo.FindUsers(ctx, condition, 0, 100, []string{"id"})

	assert.Equal(t, len(orgResults2), 1)
	assert.NotNil(t, orgResults2[0].ID)
	assert.Equal(t, orgResults2[0].Account, "account1")
	assert.Equal(t, orgResults2[0].Name, "hoge_taro")
	assert.NotNil(t, orgResults2[0].Password)

	// If key is specified, the deletion is performed.
	_ = repo.DeleteUser(ctx, &usercommon.User{ID: orgResults2[0].ID})

	deletedResults, _ := repo.FindUsers(ctx, condition, 0, 100, []string{"id"})

	assert.Equal(t, len(deletedResults), 0)
}

func TestRepository_CreateUsers(t *testing.T) {
	ctx, repo, tx := createContextRepoAndTx()
	defer tx.Rollback()

	condition := usercommon.User{}
	notfoundResults, _ := repo.FindUsers(ctx, condition, 0, 100, []string{"id"})

	assert.Equal(t, len(notfoundResults), 0)

	exchange1 := usercommon.NewUser("account1", "hoge_taro", "password")
	exchange2 := usercommon.NewUser("account2", "hoge_jiro", "password")

	insertTargets := append(make([]*usercommon.User, 0), exchange1, exchange2)

	_ = repo.CreateUsers(ctx, insertTargets)
	foundResults, _ := repo.FindUsers(ctx, condition, 0, 100, []string{"account"})

	assert.Equal(t, len(foundResults), 2)
	assert.NotNil(t, foundResults[0].ID)
	assert.Equal(t, foundResults[0].Account, "account1")
	assert.Equal(t, foundResults[0].Name, "hoge_taro")
	assert.NotNil(t, foundResults[0].Password)

	assert.NotNil(t, foundResults[1].ID)
	assert.Equal(t, foundResults[1].Account, "account2")
	assert.Equal(t, foundResults[1].Name, "hoge_jiro")
	assert.NotNil(t, foundResults[1].Password)
}

func TestRepository_UpdateUsers(t *testing.T) {
	ctx, repo, tx := createContextRepoAndTx()
	defer tx.Rollback()

	user1 := usercommon.NewUser("account1", "hoge_taro", "password")
	user2 := usercommon.NewUser("account2", "hoge_jiro", "password")

	insertTargets := append(make([]*usercommon.User, 0), user1, user2)

	_ = repo.CreateUsers(ctx, insertTargets)

	condition := usercommon.User{}
	orgResults, _ := repo.FindUsers(ctx, condition, 0, 100, []string{"account"})

	assert.Equal(t, len(orgResults), 2)
	assert.NotNil(t, orgResults[0].ID)
	assert.Equal(t, orgResults[0].Account, "account1")
	assert.Equal(t, orgResults[0].Name, "hoge_taro")
	assert.NotNil(t, orgResults[0].Password)

	assert.NotNil(t, orgResults[1].ID)
	assert.Equal(t, orgResults[1].Account, "account2")
	assert.Equal(t, orgResults[1].Name, "hoge_jiro")
	assert.NotNil(t, orgResults[1].Password)

	user1.Name = "piyo_taro"
	user2.Name = "piyo_jiro"

	updateTargets := append(make([]*usercommon.User, 0), user1, user2)

	_ = repo.UpdateUsers(ctx, updateTargets)

	updateResults1, _ := repo.FindUsers(ctx, usercommon.User{Name: "hoge_taro"}, 0, 100, []string{"id"})

	assert.Equal(t, len(updateResults1), 0)

	updateResults2, _ := repo.FindUsers(ctx, usercommon.User{Name: "hoge_jiro"}, 0, 100, []string{"id"})

	assert.Equal(t, len(updateResults2), 0)

	updateResults3, _ := repo.FindUsers(ctx, condition, 0, 100, []string{"account"})

	assert.Equal(t, len(updateResults3), 2)
	assert.NotNil(t, updateResults3[0].ID)
	assert.Equal(t, updateResults3[0].Account, "account1")
	assert.Equal(t, updateResults3[0].Name, "piyo_taro")
	assert.NotNil(t, updateResults3[0].Password)

	assert.NotNil(t, updateResults3[1].ID)
	assert.Equal(t, updateResults3[1].Account, "account2")
	assert.Equal(t, updateResults3[1].Name, "piyo_jiro")
	assert.NotNil(t, updateResults3[1].Password)
}

func TestRepository_DeleteUsers(t *testing.T) {
	ctx, repo, tx := createContextRepoAndTx()
	defer tx.Rollback()

	user1 := usercommon.NewUser("account1", "hoge_taro", "password")
	user2 := usercommon.NewUser("account2", "hoge_jiro", "password")

	insertTargets := append(make([]*usercommon.User, 0), user1, user2)

	_ = repo.CreateUsers(ctx, insertTargets)

	condition := usercommon.User{}
	orgResults, _ := repo.FindUsers(ctx, condition, 0, 100, []string{"account"})

	assert.Equal(t, len(orgResults), 2)
	assert.NotNil(t, orgResults[0].ID)
	assert.Equal(t, orgResults[0].Account, "account1")
	assert.Equal(t, orgResults[0].Name, "hoge_taro")
	assert.NotNil(t, orgResults[0].Password)

	assert.NotNil(t, orgResults[1].ID)
	assert.Equal(t, orgResults[1].Account, "account2")
	assert.Equal(t, orgResults[1].Name, "hoge_jiro")
	assert.NotNil(t, orgResults[1].Password)

	_ = repo.DeleteUsers(ctx, insertTargets)

	results, _ := repo.FindUsers(ctx, condition, 0, 100, []string{"id"})

	assert.Equal(t, len(results), 0)
}

// createTestUserData creates an array of users.
func createTestUserData(count int) (result []*usercommon.User) {
	result = make([]*usercommon.User, 0)
	for i := 0; i < count; i++ {
		data := usercommon.User{ID: "id" + fmt.Sprintf("%02d", i),
			Account:  "account" + fmt.Sprintf("%02d", i),
			Name:     "name" + fmt.Sprintf("%02d", i),
			Password: "password" + fmt.Sprintf("%02d", i)}
		result = append(result, &data)
	}
	return
}

func TestRepository_FindUserByCondition(t *testing.T) {
	ctx, repo, tx := createContextRepoAndTx()
	defer tx.Rollback()

	users := createTestUserData(5)
	err := insertTestDataToTx(tx, users)
	assert.NoError(t, err)

	condition1 := &usercommon.Condition{
		IDs:            []string{"id01", "id02"},
		SortConditions: []string{"id"},
	}

	results, _ := repo.FindUserByCondition(ctx, condition1)

	assert.Equal(t, 2, len(results))
	assert.Equal(t, "id01", results[0].ID)
	assert.Equal(t, "id02", results[1].ID)

	condition2 := &usercommon.Condition{
		SortConditions: []string{"id"},
		Offset:         3,
		Limit:          10, // SQLite requires setting a limit when using offset
	}

	results2, _ := repo.FindUserByCondition(ctx, condition2)

	assert.Equal(t, 2, len(results2))
	assert.Equal(t, "id03", results2[0].ID)
	assert.Equal(t, "id04", results2[1].ID)

	condition3 := &usercommon.Condition{
		SortConditions: []string{"id"},
		Limit:          2,
	}

	results3, _ := repo.FindUserByCondition(ctx, condition3)

	assert.Equal(t, 2, len(results3))
	assert.Equal(t, "id00", results3[0].ID)
	assert.Equal(t, "id01", results3[1].ID)
}
