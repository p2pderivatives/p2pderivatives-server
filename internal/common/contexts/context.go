package contexts

import (
	"context"
	"errors"
)

//ContextKey Key parameter to set/retrieve from a context.
type ContextKey string

const (
	//UserID is a key to set and retrieve user IDs to/from contexts.
	UserID ContextKey = "user_id"
)

// GetUserID retrieves the ID of a user from the given context. If not founds,
// panics.
func GetUserID(ctx context.Context) string {
	val, ok := ctx.Value(UserID).(string)
	if ok {
		return val
	}
	panic(errors.New("unauthenticated request"))
}

//SetUserID sets the given user ID to the given context.
func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserID, userID)
}
