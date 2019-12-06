package mock_userrepository

import (
	context "context"
	"errors"
	"p2pderivatives-server/internal/common/servererror"
	"p2pderivatives-server/internal/user/usercommon"
	"sort"
)

// RepositoryMock is a mock for the usercommon.RepositoryIf interface.
type RepositoryMock struct {
	storage map[string]*usercommon.User
}

// NewRepositoryMock creates a new RepositoryMock instance.
func NewRepositoryMock() *RepositoryMock {
	return &RepositoryMock{storage: make(map[string]*usercommon.User)}
}

// CountUsers return the number of user matching the given condition.
func (repo *RepositoryMock) CountUsers(ctx context.Context, condition interface{}) (int, error) {
	return len(repo.storage), nil
}

// CreateUser creates a new usercommon.
func (repo *RepositoryMock) CreateUser(ctx context.Context, user *usercommon.User) error {
	repo.storage[user.ID] = makeUserCopy(user)
	return nil
}

// CreateUsers insert the given users.
func (repo *RepositoryMock) CreateUsers(ctx context.Context, users []*usercommon.User) error {
	for _, user := range users {
		repo.CreateUser(ctx, user)
	}
	return nil
}

// DeleteUser deletes a usercommon.
func (repo *RepositoryMock) DeleteUser(ctx context.Context, user *usercommon.User) error {
	delete(repo.storage, user.ID)
	return nil
}

// DeleteUsers deletes users.
func (repo *RepositoryMock) DeleteUsers(ctx context.Context, users []*usercommon.User) error {
	for _, user := range users {
		repo.DeleteUser(ctx, user)
	}
	return nil
}

// FindFirstUser returns the first user matching the given condition.
func (repo *RepositoryMock) FindFirstUser(
	ctx context.Context, condition interface{}, orders []string) (*usercommon.User, error) {
	model, ok := condition.(usercommon.User)
	if !ok {
		model = *condition.(*usercommon.User)
	}

	if model.ID == "" {
		if model.Account != "" {
			return repo.FindFirstUserByAccount(ctx, condition)
		} else if model.RefreshToken != "" {
			return repo.FindFirstUserByRefreshToken(ctx, condition)
		} else {
			panic("No implemented.")
		}
	}

	result, ok := repo.storage[model.ID]

	if !ok {
		return nil, servererror.NewError(servererror.NotFoundError, "Not Found", nil)
	}

	return result, nil
}

// FindFirstUserByAccount return the user matching the given condition.
func (repo *RepositoryMock) FindFirstUserByAccount(
	ctx context.Context, condition interface{}) (*usercommon.User, error) {

	query, ok := condition.(usercommon.User)
	if !ok {
		query = *condition.(*usercommon.User)
	}

	if query.Account == "" {
		panic("No account in query.")
	}

	for _, user := range repo.storage {
		if user.Account == query.Account {
			return makeUserCopy(user), nil
		}
	}

	return nil, errors.New("Not found")
}

// FindFirstUserByRefreshToken return the user matching the given condition.
func (repo *RepositoryMock) FindFirstUserByRefreshToken(
	ctx context.Context, condition interface{}) (*usercommon.User, error) {

	query, ok := condition.(usercommon.User)
	if !ok {
		query = *condition.(*usercommon.User)
	}

	if query.RefreshToken == "" {
		panic("No account in query.")
	}

	for _, user := range repo.storage {
		if user.RefreshToken == query.RefreshToken {
			return makeUserCopy(user), nil
		}
	}

	return nil, errors.New("Not found")
}

// FindUserByCondition is not implemented.
func (repo *RepositoryMock) FindUserByCondition(ctx context.Context, condition *usercommon.Condition) (result []usercommon.User, err error) {
	panic("Not implemented")
}

// ByName struct to order users by name.
type ByName struct {
	users []usercommon.User
}

func (s ByName) Less(i, j int) bool { return s.users[i].ID < s.users[j].ID }
func (s ByName) Len() int           { return len(s.users) }
func (s ByName) Swap(i, j int)      { s.users[i], s.users[j] = s.users[j], s.users[i] }

// FindUsers find users using the given parameters.
func (repo *RepositoryMock) FindUsers(
	ctx context.Context,
	condition interface{},
	offset int,
	limit int,
	orders []string) ([]usercommon.User, error) {
	query := condition.(usercommon.User)
	if query.ID == "" {
		if query.Account == "" {
			result, err := repo.GetAllUsers(ctx)
			sort.Sort(ByName{users: result})
			return result[offset:], err
		}
		result := make([]usercommon.User, 0)
		user, _ := repo.FindFirstUserByAccount(ctx, query)
		result = append(result, *user)
		return result, nil
	}

	if model, ok := repo.storage[query.ID]; ok {
		result := make([]usercommon.User, 0)
		return append(result, *model), nil

	}

	return nil, errors.New("Not Found")
}

// GetAllUsers return all the users.
func (repo *RepositoryMock) GetAllUsers(ctx context.Context) ([]usercommon.User, error) {
	users := make([]usercommon.User, 0)

	for _, user := range repo.storage {
		users = append(users, *makeUserCopy(user))
	}

	return users, nil
}

// UpdateUser updates user data.
func (repo *RepositoryMock) UpdateUser(ctx context.Context, user *usercommon.User) error {
	repo.storage[user.ID] = makeUserCopy(user)
	return nil
}

// UpdateUsers updates multiple users.
func (repo *RepositoryMock) UpdateUsers(ctx context.Context, users []*usercommon.User) error {
	for _, user := range users {
		repo.UpdateUser(ctx, user)
	}

	return nil
}

// InsertTestUserData inserts users directly in the repository.
func InsertTestUserData(repo *RepositoryMock, models []*usercommon.User) {
	for _, model := range models {
		repo.storage[model.ID] = makeUserCopy(model)
	}
}

func makeUserCopy(model *usercommon.User) *usercommon.User {
	return &usercommon.User{
		ID:                    model.ID,
		Account:               model.Account,
		Name:                  model.Name,
		Password:              model.Password,
		RequireChangePassword: model.RequireChangePassword,
		RefreshToken:          model.RefreshToken,
	}
}
