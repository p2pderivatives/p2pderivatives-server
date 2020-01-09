package usercommon

import (
	"context"
)

// ServiceIf an interface representing a service to interact with user
// functionalities.
type ServiceIf interface {
	CreateUser(ctx context.Context, condition *User) (*User, error)
	FindFirstUser(ctx context.Context, condition *User, orders []string) (*User, error)
	FindFirstUserByName(ctx context.Context, name string) (*User, error)
	FindUsers(ctx context.Context, condition *User, offset int, limit int, orders []string) ([]User, error)
	GetAllUsers(ctx context.Context) ([]User, error)
	UpdateUser(ctx context.Context, condition *User) (*User, error)
	DeleteUser(ctx context.Context, condition *User) error
	AuthenticateUser(ctx context.Context, account, password string) (*User, *TokenInfo, error)
	FindUserByCondition(ctx context.Context, condition *Condition) ([]User, error)
	ChangeUserPassword(ctx context.Context, account, oldPassword, newPassword string) (*User, error)
	RefreshUserToken(ctx context.Context, refreshToken string) (*TokenInfo, error)
	RevokeRefreshToken(ctx context.Context, refreshToken string) error
}

// RepositoryIf is used to interact with a storage layer for User data.
type RepositoryIf interface {
	FindFirstUser(ctx context.Context, condition interface{}, orders []string) (*User, error)
	FindUsers(ctx context.Context, condition interface{}, offset int, limit int, orders []string) (result []User, err error)
	FindUserByCondition(ctx context.Context, condition *Condition) (result []User, err error)
	GetAllUsers(ctx context.Context) ([]User, error)
	CountUsers(ctx context.Context, condition interface{}) (count int, err error)
	CreateUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, user *User) error
	CreateUsers(ctx context.Context, users []*User) error
	DeleteUsers(ctx context.Context, users []*User) error
	UpdateUsers(ctx context.Context, users []*User) error
}
