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
	userService  usercommon.ServiceIf
	userChannels map[string]*chan *UserNotice
	channelLock  sync.RWMutex
	config       *usercommon.Config
}

// NewController creates a new Controller struct.
func NewController(
	service usercommon.ServiceIf,
	config *usercommon.Config) *Controller {
	channels := make(map[string]*chan *UserNotice)
	return &Controller{
		userService:  service,
		userChannels: channels,
		config:       config,
	}
}

// Close cleans up the server resources.
func (controller *Controller) Close() {
	controller.channelLock.Lock()
	for _, channel := range controller.userChannels {
		close(*channel)
	}
	controller.channelLock.Unlock()
}

// RegisterUser register a user in the system.
func (controller *Controller) RegisterUser(
	ctx context.Context,
	request *UserRegisterRequest) (*UserRegisterResponse, error) {
	userModel := usercommon.NewUser(request.Account, request.Name, request.Password)

	existingUser, err := controller.userService.FindFirstUser(ctx, &usercommon.User{
		Name:    request.Name,
		Account: request.Account,
	}, nil)

	if existingUser != nil {
		return nil, servererror.NewAlreadyExistsStatus(
			"User with same name or account already exists.").Err()
	}

	createdUser, err := controller.userService.CreateUser(ctx, userModel)

	if err != nil {
		return nil, servererror.GetGrpcStatus(ctx, err).Err()
	}

	controller.userStatusUpdater(userModel.Name, UserStatus_REGISTERED)

	response := UserRegisterResponse{
		Id:      createdUser.ID,
		Account: createdUser.Account,
		Name:    createdUser.Name,
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

	controller.userStatusUpdater(user.Name, UserStatus_UNREGISTERED)
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

// GetUserStatuses streams a list of all connected user and notifies when users
// connect, disconnect, register or unregister from the system.
func (controller *Controller) GetUserStatuses(
	empty *Empty,
	stream User_GetUserStatusesServer) error {
	userID := contexts.GetUserID(stream.Context())
	user, err := controller.userService.FindFirstUser(
		stream.Context(), &usercommon.User{ID: userID}, nil)

	if err != nil {
		return servererror.GetGrpcStatus(stream.Context(), err).Err()
	}

	// Add channel first so that if two user connect concurrently they will see
	// each others.
	userChannel := make(chan *UserNotice, 10)
	controller.addUserChannel(userChannel, user.ID, user.Name)

	userIDs := controller.getConnectedUsers()

	// Notify the newly connected user about already connected users.
	for _, curUserID := range userIDs {
		// Skip the new userusercommon.
		if curUserID == user.ID {
			continue
		}

		connectedUser, err := controller.userService.FindFirstUser(
			stream.Context(),
			&usercommon.User{
				ID: curUserID,
			}, nil)

		if err != nil {
			// TODO(tibo): add log here
			continue
		}

		err = controller.sendNotice(
			user,
			connectedUser,
			UserStatus_CONNECTED,
			stream)

		if err != nil {
			return err
		}
	}

	for notice := range userChannel {
		if notice.Name == user.Name {
			continue
		}
		err := stream.Send(notice)

		if err != nil {
			controller.removeUserChannel(user)
			return err
		}
	}

	return nil
}

func (controller *Controller) sendNotice(
	userToNotify *usercommon.User,
	updatedUser *usercommon.User,
	status UserStatus,
	stream User_GetUserStatusesServer) error {

	notice := UserNotice{
		Name:   updatedUser.Name,
		Status: status,
	}

	err := stream.Send(&notice)

	if err != nil {
		controller.removeUserChannel(userToNotify)
	}

	return err
}

func userModelToInfo(user *usercommon.User) *UserInfo {
	userInfo := UserInfo{
		Name: user.Name,
	}

	return &userInfo
}

func (controller *Controller) userStatusUpdater(
	name string, status UserStatus) {
	notice := UserNotice{
		Name:   name,
		Status: status,
	}
	controller.channelLock.RLock()
	defer controller.channelLock.RUnlock()
	for _, channel := range controller.userChannels {
		*channel <- &notice
	}
}

func (controller *Controller) addUserChannel(
	channel chan *UserNotice, id string, name string) {
	controller.channelLock.Lock()
	controller.userChannels[id] = &channel
	controller.channelLock.Unlock()
	controller.userStatusUpdater(name, UserStatus_CONNECTED)
}

func (controller *Controller) removeUserChannel(
	user *usercommon.User) {
	controller.channelLock.Lock()
	delete(controller.userChannels, user.ID)
	controller.channelLock.Unlock()
	controller.userStatusUpdater(user.Name, UserStatus_DISCONNECTED)
}

func (controller *Controller) getConnectedUsers() []string {
	controller.channelLock.RLock()
	defer controller.channelLock.RUnlock()
	nbUsers := len(controller.userChannels)
	userIds := make([]string, nbUsers)

	for userID := range controller.userChannels {
		nbUsers--
		userIds[nbUsers] = userID
	}

	return userIds
}
