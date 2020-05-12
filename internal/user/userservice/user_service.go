package userservice

import (
	context "context"
	"p2pderivatives-server/internal/common/crypto"
	"p2pderivatives-server/internal/common/servererror"
	"p2pderivatives-server/internal/common/token"
	"p2pderivatives-server/internal/database/orm"
	"p2pderivatives-server/internal/user/usercommon"
	"unicode"

	"github.com/google/uuid"
)

// Service represent a structure
type Service struct {
	userConfig     *usercommon.Config
	userRepository usercommon.RepositoryIf
	*servererror.ServiceError
}

// NewService creates a new UserService instance.
func NewService(
	repository usercommon.RepositoryIf,
	config *usercommon.Config,
	serviceError *servererror.ServiceError) *Service {
	return &Service{
		userRepository: repository,
		userConfig:     config,
		ServiceError:   serviceError,
	}
}

// CreateUser creates a new user in the system.
func (s *Service) CreateUser(ctx context.Context, condition *usercommon.User) (*usercommon.User, error) {
	if !VerifyNewPassword(condition.Password) {
		return nil, s.CreateServiceError(ctx, servererror.InvalidArguments, "Failed to create user, password does not meet policy", nil)
	}

	hashedPasswordCondition, err := s.createHashedPasswordUser(condition)

	if err != nil {
		return nil, s.CreateServiceError(ctx, servererror.InternalError, "Failed to create User.", err)
	}

	if err := s.userRepository.CreateUser(ctx, hashedPasswordCondition); err != nil {
		return nil, s.CreateServiceError(ctx, servererror.DbError, "Failed to create User.", err)
	}
	return hashedPasswordCondition, nil
}

// FindFirstUser returns the first user matching the given condition.
func (s *Service) FindFirstUser(
	ctx context.Context, condition *usercommon.User, orders []string) (*usercommon.User, error) {
	findUsers, err := s.userRepository.FindFirstUser(ctx, condition, orders)
	if err != nil {
		if !orm.IsRecordNotFoundError(err) {
			return nil, s.CreateServiceError(ctx, servererror.DbError, "Failed to find User.", err)
		}
		return nil, s.CreateServiceError(ctx, servererror.NotFoundError, "User is not found.", err)
	}
	return findUsers, nil
}

// FindFirstUserByName returns the first user associated with the given
// name.
func (s *Service) FindFirstUserByName(
	ctx context.Context, name string) (*usercommon.User, error) {
	findUsers, err := s.userRepository.FindFirstUser(ctx, usercommon.User{Name: name}, nil)
	if err != nil {
		if !orm.IsRecordNotFoundError(err) {
			return nil, s.CreateServiceError(ctx, servererror.DbError, "Failed to find User.", err)
		}
		return nil, s.CreateServiceError(ctx, servererror.NotFoundError, "User is not found.", err)
	}
	return findUsers, nil
}

// FindUsers returns all the users matching the given condition.
func (s *Service) FindUsers(
	ctx context.Context, user *usercommon.User, offset int,
	limit int, orders []string) ([]usercommon.User, error) {
	condition := usercommon.User{
		ID:   user.ID,
		Name: user.Name,
	}
	findUser, err := s.userRepository.FindUsers(ctx, condition, offset, limit, orders)
	if err != nil {
		return nil, s.CreateServiceError(ctx, servererror.NotFoundError, "User is not found.", err)
	}
	return findUser, nil
}

// GetAllUsers returns all the users registered in the system.
func (s *Service) GetAllUsers(ctx context.Context) (users []usercommon.User, err error) {
	users, err = s.userRepository.GetAllUsers(ctx)
	return
}

// UpdateUser update a user information and returns the new user
func (s *Service) UpdateUser(ctx context.Context, condition *usercommon.User) (*usercommon.User, error) {
	targetUser, err := s.userRepository.FindFirstUser(ctx, &usercommon.User{ID: condition.ID}, nil)
	if err != nil {
		return nil, s.CreateServiceError(ctx, servererror.NotFoundError, "Failed to find user", err)
	}
	condition.Password = targetUser.Password
	if err := s.userRepository.UpdateUser(ctx, condition); err != nil {
		return nil, s.CreateServiceError(ctx, servererror.NotFoundError, "Failed to update User", err)
	}
	return condition, nil
}

// ChangeUserPassword changes the user password and returns the new user
func (s *Service) ChangeUserPassword(
	ctx context.Context,
	userID string,
	newPassword string,
	oldPassword string,
) (*usercommon.User, error) {
	targetUser, err := s.FindFirstUser(ctx, &usercommon.User{ID: userID}, nil)
	if err != nil {
		// Use the same message when returning different type of errors on
		// password change.
		return nil, s.CreateServiceError(ctx, servererror.InvalidArguments, "Failed to update user password", err)
	}
	if !s.isPasswordValid(oldPassword, targetUser.Password) {
		// Use the same message when returning different type of errors on
		// password change.
		return nil, s.CreateServiceError(ctx, servererror.InvalidArguments, "Failed to update user password", nil)
	}
	if !VerifyNewPassword(newPassword) {
		// Use the same message when returning different type of errors on
		// password change.
		return nil, s.CreateServiceError(ctx, servererror.InvalidArguments, "Failed to update user password", nil)
	}

	hashedPasswordUser, err := s.createHashedPasswordUser(&usercommon.User{
		ID:                    targetUser.ID,
		Name:                  targetUser.Name,
		Password:              newPassword,
		RequireChangePassword: false,
		RefreshToken:          targetUser.RefreshToken,
	})

	if err != nil {
		return nil, s.CreateServiceError(ctx, servererror.InternalError, "Failed to update user password", err)
	}

	if err := s.userRepository.UpdateUser(ctx, hashedPasswordUser); err != nil {
		return nil, s.CreateServiceError(ctx, servererror.NotFoundError, "Failed to update User", err)
	}
	return hashedPasswordUser, nil
}

// ResetUserPassword resets the password of the user associated with the given
// account and return the reset user.
func (s *Service) ResetUserPassword(
	ctx context.Context,
	name string,
	newPassword string,
) (*usercommon.User, error) {
	targetUser, err := s.FindFirstUserByName(ctx, name)
	if err != nil {
		return nil, s.CreateServiceError(ctx, servererror.NotFoundError, "Failed to find user", err)
	}
	if !verifyNewPassword(newPassword) {
		return nil, s.CreateServiceError(ctx, servererror.PreconditionError, "Failed to verify new password", nil)
	}
	hashedPasswordUser, err := s.createHashedPasswordUser(&usercommon.User{
		ID:                    targetUser.ID,
		Name:                  targetUser.Name,
		Password:              newPassword,
		RequireChangePassword: true,
	})

	if err != nil {
		return nil, s.CreateServiceError(ctx, servererror.InternalError, "Failed to reset password", err)
	}

	if err := s.userRepository.UpdateUser(ctx, hashedPasswordUser); err != nil {
		return nil, s.CreateServiceError(ctx, servererror.NotFoundError, "Failed to update User", err)
	}
	return hashedPasswordUser, nil
}

// DeleteUser deletes the user associated with the given condition.
func (s *Service) DeleteUser(ctx context.Context, condition *usercommon.User) error {
	if err := s.userRepository.DeleteUser(ctx, condition); err != nil {
		return s.CreateServiceError(ctx, servererror.NotFoundError, "Failed to delete User.", err)
	}

	return nil
}

func (s *Service) createHashedPasswordUser(
	user *usercommon.User) (*usercommon.User, error) {
	if user.Password == "" {
		return user, nil
	}

	protectedForm, err := s.getPasswordProtectedForm(user.Password)

	if err != nil {
		return nil, err
	}

	updatedUser := &usercommon.User{
		ID:                    user.ID,
		Name:                  user.Name,
		Password:              protectedForm,
		RequireChangePassword: user.RequireChangePassword,
		RefreshToken:          user.RefreshToken,
	}

	return updatedUser, nil
}

// AuthenticateUser checks that the password provided matches the one of the
// associated user. If it does, returns the matching user and the token info to
// be used as authentication, otherwise returns an error.
func (s *Service) AuthenticateUser(
	ctx context.Context, name, password string) (*usercommon.User, *usercommon.TokenInfo, error) {
	condition := usercommon.User{
		Name: name,
	}
	userInfo, err := s.userRepository.FindFirstUser(ctx, condition, []string{})
	if err != nil {
		return nil, nil, s.CreateServiceError(
			ctx, servererror.UnauthenticatedError, "Fail to authenticate user.", err,
		)
	}
	isPasswordValid := s.isPasswordValid(password, userInfo.Password)
	if !isPasswordValid {
		return nil, nil, s.CreateServiceError(
			ctx, servererror.UnauthenticatedError, "Fail to authenticate user.", err,
		)
	}
	tokenInfo, err := s.generateUserToken(ctx, userInfo)
	if err != nil {
		return nil, nil, err
	}
	return userInfo, tokenInfo, nil
}

// FindUserByCondition returns the set of users matching the given condition.
func (s *Service) FindUserByCondition(
	ctx context.Context, condition *usercommon.Condition) ([]usercommon.User, error) {
	return s.userRepository.FindUserByCondition(ctx, condition)
}

//RevokeRefreshToken revokes the given refresh token.
func (s *Service) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	refreshTokenID, err := token.VerifyToken(refreshToken)
	if err != nil {
		return s.CreateServiceError(ctx, servererror.InvalidArguments, "Failed to verify refresh token", err)
	}
	condition := usercommon.User{
		RefreshToken: refreshTokenID,
	}
	user, err := s.userRepository.FindFirstUser(ctx, condition, []string{})
	if err != nil {
		return s.CreateServiceError(ctx, servererror.NotFoundError, "user with specific RefreshToken not found", err)
	}
	user.RefreshToken = ""
	if err = s.userRepository.UpdateUser(ctx, user); err != nil {
		return s.CreateServiceError(ctx, servererror.DbError, "failed to update user info", err)
	}
	return nil
}

//RefreshUserToken refreshes the given token and returns the new token info.
func (s *Service) RefreshUserToken(ctx context.Context, refreshToken string) (*usercommon.TokenInfo, error) {
	refreshTokenID, err := token.VerifyToken(refreshToken)
	if err != nil {
		return nil, s.CreateServiceError(ctx, servererror.InvalidArguments, "Failed to verify refresh token", err)
	}
	condition := usercommon.User{
		RefreshToken: refreshTokenID,
	}
	user, err := s.userRepository.FindFirstUser(ctx, condition, []string{})
	if err != nil {
		return nil, s.CreateServiceError(ctx, servererror.NotFoundError, "user with specific RefreshToken not found", err)
	}
	return s.generateUserToken(ctx, user)
}

//generateUserToken generates a JWT token for the given user.
func (s *Service) generateUserToken(ctx context.Context, userInfo *usercommon.User) (*usercommon.TokenInfo, error) {
	//Generate JWT Token
	accessToken, expiresIn, err := token.GenerateAccessToken(userInfo.ID)
	if err != nil {
		return nil, s.CreateServiceError(ctx, servererror.InternalError, "Failed to generate access token.", err)
	}
	refreshTokenID := uuid.New().String()
	refreshToken, err := token.GenerateRefreshToken(refreshTokenID)
	if err != nil {
		return nil, s.CreateServiceError(ctx, servererror.InternalError, "Failed to generate refresh token.", err)
	}
	userInfo.RefreshToken = refreshTokenID
	if err = s.userRepository.UpdateUser(ctx, userInfo); err != nil {
		return nil, err
	}
	return &usercommon.TokenInfo{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

func (s *Service) isPasswordValid(password, protectedForm string) bool {
	return crypto.IsPasswordValid(
		password,
		protectedForm,
		s.userConfig.SaltLen,
		s.userConfig.PasswordTime,
		s.userConfig.PasswordMemory,
		s.userConfig.PasswordThreads,
		s.userConfig.KeyLen)
}

func (s *Service) getPasswordProtectedForm(password string) (string, error) {
	salt, err := crypto.GenerateSalt(s.userConfig.SaltLen)

	if err != nil {
		return "", err
	}

	protectedForm := crypto.GetPasswordProtectedForm(
		password,
		salt,
		s.userConfig.PasswordTime,
		s.userConfig.PasswordMemory,
		s.userConfig.PasswordThreads,
		s.userConfig.KeyLen)
	return protectedForm, nil
}

func verifyNewPassword(password string) bool {
	var number, upper, lower, special bool
	letters := 0
	for _, c := range password {
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsUpper(c):
			upper = true
		case unicode.IsLower(c):
			lower = true
		case isSpecialCharacter(c):
			special = true
		case unicode.IsLetter(c):
		}
		letters++
	}
	validLength := (letters >= 8 && letters <= 32)

	//Only containing all cases is allowed
	return (number && upper && lower && special && validLength)
}
