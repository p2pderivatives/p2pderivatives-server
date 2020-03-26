package usercontroller

import (
	"context"
	"sync"

	"p2pderivatives-server/internal/common/contexts"
	"p2pderivatives-server/internal/common/servererror"
	"p2pderivatives-server/internal/user/usercommon"
)

var empty *Empty = &Empty{}

type void struct{}

var member void

const (
	ok    = iota
	notOk = iota
)

type dlcMessageWithAck struct {
	message *DlcMessage
	ackChan chan int
}

// Controller represents the grpc server serving the user services.
type Controller struct {
	userService  usercommon.ServiceIf
	userChannels map[string]map[chan *dlcMessageWithAck]void
	channelLock  sync.RWMutex
	config       *usercommon.Config
}

// NewController creates a new Controller struct.
func NewController(
	service usercommon.ServiceIf,
	config *usercommon.Config) *Controller {
	channels := make(map[string]map[chan *dlcMessageWithAck]void)
	return &Controller{
		userService:  service,
		userChannels: channels,
		config:       config,
	}
}

// Close cleans up the server resources.
func (controller *Controller) Close() {
	controller.channelLock.Lock()
	for _, channels := range controller.userChannels {
		for channel := range channels {
			close(channel)
		}
	}
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

// ReceiveDlcMessages enables receiving messages from other users pertaining
// to the DLC protocol.
func (controller *Controller) ReceiveDlcMessages(
	empty *Empty,
	stream User_ReceiveDlcMessagesServer) error {
	userID := contexts.GetUserID(stream.Context())
	user, err := controller.userService.FindFirstUser(
		stream.Context(), &usercommon.User{ID: userID}, nil)

	if err != nil {
		return servererror.GetGrpcStatus(stream.Context(), err).Err()
	}

	dlcChannel := make(chan *dlcMessageWithAck, 10)
	controller.addUserChannel(dlcChannel, user.Name)

	for messageWithAck := range dlcChannel {
		message := messageWithAck.message
		ackChannel := messageWithAck.ackChan
		if message.OrgName == user.Name {
			continue
		}

		err := stream.Send(message)

		if err != nil {
			ackChannel <- notOk
			controller.removeUserChannel(user, dlcChannel)
			return err
		}

		ackChannel <- ok
	}

	return nil

}

// SendDlcMessage enables sending a message to a user
func (controller *Controller) SendDlcMessage(
	ctx context.Context, message *DlcMessage) (*Empty, error) {

	userID := contexts.GetUserID(ctx)

	user, err := controller.userService.FindFirstUser(
		ctx, &usercommon.User{ID: userID}, nil)

	if err != nil {
		return nil, servererror.GetGrpcStatus(ctx, err).Err()
	}

	message.OrgName = user.Name

	destUserName := message.DestName

	channels, err := controller.getUserChannels(destUserName)

	if err != nil {
		return nil, err
	}

	nbChannels := 0
	ackChannel := make(chan int, len(channels))

	for channel := range channels {
		nbChannels++
		channel <- &dlcMessageWithAck{message: message, ackChan: ackChannel}
	}

	hasOk := false

	for nbChannels > 0 {
		ack := <-ackChannel
		hasOk = hasOk || ack == ok
		nbChannels--
	}

	if !hasOk {
		return nil, servererror.NewUnavailableStatus(
			"Peer connection returned error.").Err()
	}

	return empty, nil
}

// GetConnectedUsers returns a list of connected users.
func (controller *Controller) GetConnectedUsers(
	empty *Empty,
	stream User_GetConnectedUsersServer) error {
	userID := contexts.GetUserID(stream.Context())

	user, err := controller.userService.FindFirstUser(
		stream.Context(), &usercommon.User{ID: userID}, nil)

	if err != nil {
		return servererror.GetGrpcStatus(stream.Context(), err).Err()
	}

	userNames := controller.getConnectedUsers()

	for _, userName := range userNames {

		// Skip requesting user
		if userName == user.Name {
			continue
		}

		stream.Send(&UserInfo{Name: userName})
	}

	return nil
}

func userModelToInfo(user *usercommon.User) *UserInfo {
	userInfo := UserInfo{
		Name: user.Name,
	}

	return &userInfo
}

func (controller *Controller) getUserChannels(
	name string) (map[chan *dlcMessageWithAck]void, error) {
	controller.channelLock.RLock()
	defer controller.channelLock.RUnlock()
	if channels, ok := controller.userChannels[name]; ok {
		return channels, nil
	}

	return nil, servererror.NewNotFoundStatus("No such user").Err()
}

func (controller *Controller) addUserChannel(
	channel chan *dlcMessageWithAck, name string) {
	controller.channelLock.Lock()
	if controller.userChannels[name] == nil {
		controller.userChannels[name] = make(map[chan *dlcMessageWithAck]void)
	}
	controller.userChannels[name][channel] = member
	controller.channelLock.Unlock()
}

func (controller *Controller) removeUserChannel(
	user *usercommon.User, channel chan *dlcMessageWithAck) {
	controller.channelLock.Lock()
	delete(controller.userChannels[user.Name], channel)
	controller.channelLock.Unlock()
}

func (controller *Controller) getConnectedUsers() []string {
	controller.channelLock.RLock()
	defer controller.channelLock.RUnlock()
	nbUsers := len(controller.userChannels)
	userNames := make([]string, nbUsers)

	for userName := range controller.userChannels {
		nbUsers--
		userNames[nbUsers] = userName
	}

	return userNames
}
