package cli

import (
	"context"
	"flag"
	"log"

	"p2pderivatives-server/internal/user/usercontroller"

	"google.golang.org/grpc"
)

// UnregisterUserCmd unregisters a user in the system.
type UnregisterUserCmd struct {
	cmd     string
	flagSet *flag.FlagSet
	id      *string
}

// NewUnregisterUserCmd returns a new UnregisterUserCmd struct.
func NewUnregisterUserCmd() *UnregisterUserCmd {
	return &UnregisterUserCmd{}
}

// Command returns the command name.
func (cmd *UnregisterUserCmd) Command() string {
	return cmd.cmd
}

// Init initializes the command.
func (cmd *UnregisterUserCmd) Init() {
	cmd.cmd = "unregisteruser"
	cmd.flagSet = flag.NewFlagSet(cmd.cmd, flag.ExitOnError)
	cmd.id = cmd.flagSet.String("id", "", "The id of the user to unregister")
}

// GetFlagSet returns the flag set for this command.
func (cmd *UnregisterUserCmd) GetFlagSet() *flag.FlagSet {
	return cmd.flagSet
}

// Do performs the command action.
func (cmd *UnregisterUserCmd) Do(ctx context.Context, conn *grpc.ClientConn) {

	client := usercontroller.NewUserClient(conn)

	if *cmd.id == "" {
		log.Fatal("Both id and name parameters are required")
	}

	request := usercontroller.UnregisterUserRequest{}

	_, err := client.UnregisterUser(ctx, &request)

	if err != nil {
		log.Printf("Error unregistering user %v", err)
	}
}
