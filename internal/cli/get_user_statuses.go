package cli

import (
	"context"
	"flag"
	"io"
	"log"

	"p2pderivatives-server/internal/user/usercontroller"

	"google.golang.org/grpc"
)

// GetUserStatusesCmd registers a user in the system.
type GetUserStatusesCmd struct {
	cmd     string
	flagSet *flag.FlagSet
	id      *string
}

// NewGetUserStatusesCmd returns a new GetUserStatusesCmd struct.
func NewGetUserStatusesCmd() *GetUserStatusesCmd {
	return &GetUserStatusesCmd{}
}

// Command returns the command name.
func (cmd *GetUserStatusesCmd) Command() string {
	return cmd.cmd
}

// Parse parses the command arguments.
func (cmd *GetUserStatusesCmd) Parse(args []string) {
	cmd.flagSet.Parse(args)
}

// Init initializes the command.
func (cmd *GetUserStatusesCmd) Init() {
	cmd.cmd = "getuserstatuses"
	cmd.flagSet = flag.NewFlagSet(cmd.cmd, flag.ExitOnError)
	cmd.id = cmd.flagSet.String("id", "", "The id of the user listening")
}

// GetFlagSet returns the flag set for this command.
func (cmd *GetUserStatusesCmd) GetFlagSet() *flag.FlagSet {
	return cmd.flagSet
}

// Do performs the command action.
func (cmd *GetUserStatusesCmd) Do(ctx context.Context, conn *grpc.ClientConn) {
	client := usercontroller.NewUserClient(conn)
	stream, err := client.GetUserStatuses(ctx, &usercontroller.Empty{})

	if err != nil {
		log.Fatalf("Could not get user list %v", err)
	}

	for {
		userNotice, err := stream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("%v.GetUserStatuses(_) = _, %v", client, err)
		}
		name := userNotice.Name
		status := userNotice.Status.String()
		log.Println("Name: ", name, " Status: ", status)
	}
}
