package usercontroller

import (
	"context"
	"sync"

	"p2pderivatives-server/internal/common/contexts"
	"p2pderivatives-server/internal/common/servererror"
	"p2pderivatives-server/internal/user/usercommon"
)

var empty *Empty = &Empty{}

// Controller represents the grpc server serving the user services.
type Controller struct {
	userService usercommon.ServiceIf
	channelLock sync.RWMutex
	config      *usercommon.Config
}

// NewController creates a new Controller struct.
func NewController(
	service usercommon.ServiceIf,
	config *usercommon.Config) *Controller {
	return &Controller{
		userService: service,
		config:      config,
	}
}

// Close cleans up the server resources.
func (controller *Controller) Close() {
	controller.channelLock.Lock()
	controller.channelLock.Unlock()
}

// RegisterUser register a user in the system.
func (controller *Controller) RegisterUser(
	ctx context.Context,
	request *UserRegisterRequest) (*UserRegisterResponse, error) {
	userModel := usercommon.NewUser(request.Name, request.Password)

	existingUser, err := controller.userService.FindFirstUser(ctx, &usercommon.User{
		Name: request.Name,
	}, nil)

	if existingUser != nil {
		return nil, servererror.NewAlreadyExistsStatus(
			"User with same name or account already exists.").Err()
	}

	createdUser, err := controller.userService.CreateUser(ctx, userModel)

	if err != nil {
		return nil, servererror.GetGrpcStatus(ctx, err).Err()
	}

	response := UserRegisterResponse{
		Id:   createdUser.ID,
		Name: createdUser.Name,
	}

	return &response, nil
}

// UnregisterUser unregisters a user from the system.
func (controller *Controller) UnregisterUser(
	ctx context.Context,
	request *UnregisterUserRequest) (*Empty, error) {

	userID := contexts.GetUserID(ctx)
	user, err := controller.userService.FindFirstUser(ctx, &usercommon.User{ID: userID}, nil)
	if err != nil {
		return nil, servererror.GetGrpcStatus(ctx, err).Err()
	}

	err = controller.userService.DeleteUser(ctx, user)
	if err != nil {
		return nil, servererror.GetGrpcStatus(ctx, err).Err()
	}

	return empty, nil
}

// GetUserList returns a list of all registered users in the system.
func (controller *Controller) GetUserList(
	empty *Empty,
	stream User_GetUserListServer) error {
	users, err := controller.userService.GetAllUsers(stream.Context())

	if err != nil {
		return servererror.GetGrpcStatus(stream.Context(), err).Err()
	}

	for _, user := range users {
		stream.Send(userModelToInfo(&user))
	}

	return err
}

func userModelToInfo(user *usercommon.User) *UserInfo {
	userInfo := UserInfo{
		Name: user.Name,
	}

	return &userInfo
}
