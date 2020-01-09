package cli

import (
	"context"
	"flag"
	"log"

	"p2pderivatives-server/internal/user/usercontroller"

	"google.golang.org/grpc"
)

// RegisterUserCmd registers a user in the system.
type RegisterUserCmd struct {
	cmd      string
	flagSet  *flag.FlagSet
	name     *string
	password *string
}

// NewRegisterUserCmd returns a new RegisterUserCmd struct.
func NewRegisterUserCmd() *RegisterUserCmd {
	return &RegisterUserCmd{}
}

// Command returns the command name.
func (cmd *RegisterUserCmd) Command() string {
	return cmd.cmd
}

// Init initializes the command.
func (cmd *RegisterUserCmd) Init() {
	cmd.cmd = "registeruser"
	cmd.flagSet = flag.NewFlagSet(cmd.cmd, flag.ExitOnError)
	cmd.name = cmd.flagSet.String("name", "", "The name of the user to register")
	cmd.password = cmd.flagSet.String("password", "", "The password of the user to register")
}

// GetFlagSet returns the flag set for this command.
func (cmd *RegisterUserCmd) GetFlagSet() *flag.FlagSet {
	return cmd.flagSet
}

// Do performs the command action.
func (cmd *RegisterUserCmd) Do(ctx context.Context, conn *grpc.ClientConn) {

	client := usercontroller.NewUserClient(conn)

	if *cmd.name == "" || *cmd.password == "" {
		log.Fatal("Account, name and password parameters are required")
	}

	request := usercontroller.UserRegisterRequest{
		Name:     *cmd.name,
		Password: *cmd.password,
	}

	resp, err := client.RegisterUser(ctx, &request)

	if err != nil {
		log.Fatalf("Error registering user %v", err)
	}

	log.Println("User ID: ", resp.Id)
}
