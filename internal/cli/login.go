package cli

import (
	"context"
	"flag"
	"log"

	"p2pderivatives-server/internal/authentication"

	"google.golang.org/grpc"
)

// LoginCmd registers a user in the system.
type LoginCmd struct {
	cmd      string
	flagSet  *flag.FlagSet
	name     *string
	password *string
}

// NewLoginCmd returns a new GetUserListCmd struct.
func NewLoginCmd() *LoginCmd {
	return &LoginCmd{}
}

// Command returns the command name.
func (cmd *LoginCmd) Command() string {
	return cmd.cmd
}

// Parse parses the command arguments.
func (cmd *LoginCmd) Parse(args []string) {
	cmd.flagSet.Parse(args)
}

// Init initializes the command.
func (cmd *LoginCmd) Init() {
	cmd.cmd = "login"
	cmd.flagSet = flag.NewFlagSet(cmd.cmd, flag.ExitOnError)
	cmd.name = cmd.flagSet.String("name", "", "The name of the user to login")
	cmd.password = cmd.flagSet.String("password", "", "The password of the user to login")
}

// GetFlagSet returns the flag set for this command.
func (cmd *LoginCmd) GetFlagSet() *flag.FlagSet {
	return cmd.flagSet
}

// Do performs the command action.
func (cmd *LoginCmd) Do(ctx context.Context, conn *grpc.ClientConn) {
	client := authentication.NewAuthenticationClient(conn)

	request := &authentication.LoginRequest{
		Name: *cmd.name, Password: *cmd.password,
	}

	response, err := client.Login(ctx, request)

	if err != nil {
		log.Fatalf("Could not login %v", err)
	}

	log.Println("Logged in. Token: ", response.Token)
}
