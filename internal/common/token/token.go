package token

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

// IsTokenExpiredError returns true if error is TokenExpired error
func IsTokenExpiredError(err error) bool {
	return err == ErrTokenExpired
}

var conf *Config

//Init Token sets the global configuration to be used for token instances.
func Init(config *Config) {
	conf = config
}

//GenerateAccessToken creates a new jwt token using the provided id.
func GenerateAccessToken(id string) (string, int64, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Id:        id,
		ExpiresAt: time.Now().UTC().Add(conf.Exp).Unix(),
	})

	tokenStr, err := token.SignedString([]byte(conf.Secret))
	return tokenStr, int64(conf.Exp.Seconds()), err
}

//VerifyToken checks that the given token is valid.
func VerifyToken(tokenStr string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(conf.Secret), nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return "", ErrTokenExpired
			}
		}
		return "", errors.Errorf("%s is invalid", tokenStr)
	}

	if token == nil {
		return "", errors.Errorf("not found token in %s:", tokenStr)
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok {
		return "", errors.Errorf("not found claims in %s", tokenStr)
	}
	return claims.Id, nil
}

//GenerateRefreshToken creates a refresh token for the given id.
func GenerateRefreshToken(id string) (string, error) {
	exp := time.Now().UTC().Add(conf.RefreshExp).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Id:        id,
		ExpiresAt: exp,
	})

	tokenStr, err := token.SignedString([]byte(conf.Secret))
	return tokenStr, err
}
