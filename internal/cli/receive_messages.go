package cli

import (
	"context"
	"flag"
	"io"
	"log"

	"p2pderivatives-server/internal/user/usercontroller"

	"google.golang.org/grpc"
)

// ReceiveDlcMsg registers a user in the system.
type ReceiveDlcMsg struct {
	cmd     string
	flagSet *flag.FlagSet
}

// NewReceiveDlcMsg returns a new GetUserStatusesCmd struct.
func NewReceiveDlcMsg() *ReceiveDlcMsg {
	return &ReceiveDlcMsg{}
}

// Command returns the command name.
func (cmd *ReceiveDlcMsg) Command() string {
	return cmd.cmd
}

// Parse parses the command arguments.
func (cmd *ReceiveDlcMsg) Parse(args []string) {
	cmd.flagSet.Parse(args)
}

// Init initializes the command.
func (cmd *ReceiveDlcMsg) Init() {
	cmd.cmd = "receivedlcmsg"
	cmd.flagSet = flag.NewFlagSet(cmd.cmd, flag.ExitOnError)
}

// GetFlagSet returns the flag set for this command.
func (cmd *ReceiveDlcMsg) GetFlagSet() *flag.FlagSet {
	return cmd.flagSet
}

// Do performs the command action.
func (cmd *ReceiveDlcMsg) Do(ctx context.Context, conn *grpc.ClientConn) {
	client := usercontroller.NewUserClient(conn)
	stream, err := client.ReceiveDlcMessages(ctx, &usercontroller.Empty{})

	if err != nil {
		log.Fatalf("Could not receive dlc messages %v", err)
	}

	for {
		message, err := stream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("%v.ReceiveDlcMsg(_) = _, %v", client, err)
		}
		name := message.OrgName
		payload := string(message.Payload)
		log.Println("Sender: ", name, " Message: ", payload)
	}
}
