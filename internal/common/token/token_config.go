package token

import "time"

// Config contains configuration data for JWT tokens.
type Config struct {
	Secret     string        `configkey:"app.token.secret" validate:"required"`
	Exp        time.Duration `configkey:"app.token.exp,duration" validate:"required"`
	RefreshExp time.Duration `configkey:"app.token.refresh_exp,duration" validate:"required"`
}
