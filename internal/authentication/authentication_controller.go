package authentication

import (
	"context"
	"p2pderivatives-server/internal/common/log"
	"p2pderivatives-server/internal/common/servererror"
	"p2pderivatives-server/internal/user/usercommon"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
)

// Controller handles authentication related functionalities.
type Controller struct {
	userService usercommon.ServiceIf
	userConfig  *usercommon.Config
}

// NewController returns a new Controller
func NewController(userService usercommon.ServiceIf, userConfig *usercommon.Config) *Controller {
	return &Controller{userService: userService, userConfig: userConfig}
}

// Login enables a user to login to the system.
func (s *Controller) Login(
	ctx context.Context,
	req *LoginRequest,
) (*LoginResponse, error) {
	ctx, log := log.Save(ctx, logrus.Fields{"name": req.Name})
	log.Info("Login Request")
	user, userToken, serr := s.userService.AuthenticateUser(ctx, req.Name, req.Password)
	if serr != nil {
		return nil, servererror.GetGrpcStatus(ctx, serr).Err()
	}

	log.Info("Login Success")
	return &LoginResponse{
		Name: user.Name,
		Token: &TokenInfo{
			AccessToken:  userToken.AccessToken,
			RefreshToken: userToken.RefreshToken,
			ExpiresIn:    userToken.ExpiresIn,
		},
		RequireChangePassword: user.RequireChangePassword,
	}, nil
}

// Refresh enables users to refresh their access token.
func (s *Controller) Refresh(
	ctx context.Context,
	req *RefreshRequest) (*RefreshResponse, error) {

	tokenInfo, sErr := s.userService.RefreshUserToken(ctx, req.RefreshToken)
	if sErr != nil {
		return nil, servererror.GetGrpcStatus(ctx, sErr).Err()
	}
	return &RefreshResponse{
		Token: &TokenInfo{
			AccessToken:  tokenInfo.AccessToken,
			RefreshToken: tokenInfo.RefreshToken,
			ExpiresIn:    tokenInfo.ExpiresIn,
		},
	}, nil
}

// Logout enables a user to log out the system.
func (s *Controller) Logout(
	ctx context.Context,
	req *LogoutRequest) (*Empty, error) {
	logger := ctxlogrus.Extract(ctx)
	logger.Info("Logout Request")

	if err := s.userService.RevokeRefreshToken(ctx, req.RefreshToken); err != nil {
		// Proceed even if token validation fails.
		logger.Infof("failed to revoke RefreshToken. err:%v", err)
	}

	logger.Info("Logout Success")
	return &Empty{}, nil
}
