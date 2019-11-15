package cli

import (
	"context"
	"flag"
	"io"
	"log"

	"p2pderivatives-server/internal/user/usercontroller"

	"google.golang.org/grpc"
)

// GetUserListCmd registers a user in the system.
type GetUserListCmd struct {
	cmd     string
	flagSet *flag.FlagSet
}

// NewGetUserListCmd returns a new GetUserListCmd struct.
func NewGetUserListCmd() *GetUserListCmd {
	return &GetUserListCmd{}
}

// Command returns the command name.
func (cmd *GetUserListCmd) Command() string {
	return cmd.cmd
}

// Parse parses the command arguments.
func (cmd *GetUserListCmd) Parse(args []string) {
	cmd.flagSet.Parse(args)
}

// Init initializes the command.
func (cmd *GetUserListCmd) Init() {
	cmd.cmd = "getuserlist"
	cmd.flagSet = flag.NewFlagSet(cmd.cmd, flag.ExitOnError)
}

// GetFlagSet returns the flag set for this command.
func (cmd *GetUserListCmd) GetFlagSet() *flag.FlagSet {
	return cmd.flagSet
}

// Do performs the command action.
func (cmd *GetUserListCmd) Do(ctx context.Context, conn *grpc.ClientConn) {
	client := usercontroller.NewUserClient(conn)
	stream, err := client.GetUserList(ctx, &usercontroller.Empty{})

	if err != nil {
		log.Fatalf("Could not get user list %v", err)
	}

	for {
		userInfo, err := stream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("%v.GetUserList(_) = _, %v", client, err)
		}
		name := userInfo.Name
		log.Println("Name: ", name)
	}
}
