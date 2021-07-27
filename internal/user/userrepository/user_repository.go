package userrepository

import (
	context "context"
	"p2pderivatives-server/internal/database/interceptor"
	"p2pderivatives-server/internal/user/usercommon"

	"gorm.io/gorm"
)

// Repository represents a repository to store User related data.
type Repository struct {
}

// NewRepository creates a new UserRepository using the given DB object.
func NewRepository() *Repository {
	return &Repository{}
}

func (repo *Repository) extractTx(ctx context.Context) *gorm.DB {
	tx := interceptor.ExtractTx(ctx)
	if tx == nil {
		panic("Repository called without DB tx in context.")
	}

	return tx
}

// FindFirstUser returns first User matching with specified condition.
func (repo *Repository) FindFirstUser(
	ctx context.Context,
	condition interface{},
	orders []string,
) (*usercommon.User, error) {
	tx := repo.extractTx(ctx)
	query := tx.Where(condition)
	if orders == nil {
		orders = []string{}
	}

	for _, order := range orders {
		query = query.Order(order)
	}
	var result usercommon.User
	err := query.First(&result).Error
	return &result, err
}

// FindUsers returns Users matching with specified condition
func (repo *Repository) FindUsers(
	ctx context.Context,
	condition interface{},
	offset int,
	limit int,
	orders []string,
) (result []usercommon.User, err error) {
	tx := repo.extractTx(ctx)
	query := tx.Where(condition)
	if offset >= 0 {
		query = query.Offset(offset)
	}
	if limit >= 0 {
		query = query.Limit(limit)
	}

	if orders == nil {
		orders = []string{}
	}
	for _, order := range orders {
		query = query.Order(order)
	}

	err = query.Find(&result).Error
	return
}

// FindUserByCondition returns Users matching with specified condition
func (repo *Repository) FindUserByCondition(
	ctx context.Context,
	condition *usercommon.Condition,
) (result []usercommon.User, err error) {
	tx := repo.extractTx(ctx)
	filterCondition := &usercommon.User{
		ID:   condition.ID,
		Name: condition.Name,
	}
	query := tx.Where(filterCondition)
	if condition.Offset > 0 {
		query = query.Offset(condition.Offset)
	}
	if condition.Limit > 0 {
		query = query.Limit(condition.Limit)
	}
	for _, sortCondition := range condition.SortConditions {
		query = query.Order(sortCondition)
	}
	if len(condition.IDs) > 0 {
		query = query.Where("id in (?)", condition.IDs)
	}
	err = query.Find(&result).Error
	return
}

// GetAllUsers returns all the users registered in the system
func (repo *Repository) GetAllUsers(ctx context.Context) (users []usercommon.User, err error) {
	tx := repo.extractTx(ctx)
	err = tx.Find(&users).Error
	return
}

// CountUsers returns the number of Users matching specified condition
func (repo *Repository) CountUsers(ctx context.Context, condition interface{}) (
	count int64, err error) {
	tx := repo.extractTx(ctx)
	var users []usercommon.User
	err = tx.Where(condition).Find(&users).Count(&count).Error
	return
}

// CreateUser inserts new User record
func (repo *Repository) CreateUser(ctx context.Context, user *usercommon.User) error {
	tx := repo.extractTx(ctx)
	return repo.createUser(tx, user)
}

// UpdateUser updates User record
func (repo *Repository) UpdateUser(ctx context.Context, user *usercommon.User) error {
	tx := repo.extractTx(ctx)
	return repo.updateUser(tx, user)
}

// DeleteUser deletes User record
func (repo *Repository) DeleteUser(ctx context.Context, user *usercommon.User) error {
	if user.ID == "" {
		return nil // To avoid deleting all, return here.
	}
	tx := repo.extractTx(ctx)
	return repo.deleteUser(tx, user)
}

// CreateUsers inserts new User records.
func (repo *Repository) CreateUsers(
	ctx context.Context, users []*usercommon.User) (err error) {
	tx := repo.extractTx(ctx)
	for _, user := range users {
		err = repo.createUser(tx, user)
		if err != nil {
			return
		}
	}
	return
}

// UpdateUsers updates user records
func (repo *Repository) UpdateUsers(
	ctx context.Context, users []*usercommon.User) (err error) {
	tx := repo.extractTx(ctx)
	for _, user := range users {
		err = repo.updateUser(tx, user)
		if err != nil {
			return err
		}
	}
	return
}

// DeleteUsers deletes User records
func (repo *Repository) DeleteUsers(ctx context.Context, users []*usercommon.User) (err error) {
	tx := repo.extractTx(ctx)
	for _, user := range users {
		err = repo.deleteUser(tx, user)
		if err != nil {
			return
		}
	}
	return
}

func (repo *Repository) createUser(tx *gorm.DB, user *usercommon.User) error {
	return tx.Create(user).Error
}

func (repo *Repository) updateUser(tx *gorm.DB, user *usercommon.User) error {
	return tx.Save(user).Error
}

func (repo *Repository) deleteUser(tx *gorm.DB, user *usercommon.User) error {
	if user.ID == "" {
		return nil // To avoid deleting all, return here.
	}
	return tx.Delete(user).Error
}
