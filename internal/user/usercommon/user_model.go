package usercommon

import "github.com/google/uuid"

// User represents a user in the system.
type User struct {
	ID                    string `gorm:"primary_key; size:255"`
	Name                  string `gorm:"unique; not null; size:255"`
	Password              string `gorm:"not null; size:256"`
	RequireChangePassword bool   `gorm:"not null"`
	RefreshToken          string `gorm:"size:255"`
}

// Condition represents conditions when looking up users.
type Condition struct {
	ID             string
	Name           string
	Offset         int
	Limit          int
	SortConditions []string
	IDs            []string
}

//TokenInfo represents information about a user's JWT token.
type TokenInfo struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 //有効期限、○秒、例：1800秒
}

// NewUser creates a new User structure with the given parameters.
func NewUser(name string, password string) *User {
	user := User{
		ID:                    generateUserID(),
		Name:                  name,
		Password:              password,
		RequireChangePassword: false,
	}
	return &user
}

func generateUserID() string {
	return "user-" + GenerateUUID()
}

// GenerateUUID generates a new UUID
func GenerateUUID() string {
	return uuid.New().String()
}
