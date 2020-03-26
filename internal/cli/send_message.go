package cli

import (
	"context"
	"flag"
	"log"

	"p2pderivatives-server/internal/user/usercontroller"

	"google.golang.org/grpc"
)

// SendMsgCmd registers a user in the system.
type SendMsgCmd struct {
	cmd      string
	flagSet  *flag.FlagSet
	destName *string
	message  *string
}

// NewSendMsgCmd returns a new RegisterUserCmd struct.
func NewSendMsgCmd() *SendMsgCmd {
	return &SendMsgCmd{}
}

// Command returns the command name.
func (cmd *SendMsgCmd) Command() string {
	return cmd.cmd
}

// Init initializes the command.
func (cmd *SendMsgCmd) Init() {
	cmd.cmd = "sendmsg"
	cmd.flagSet = flag.NewFlagSet(cmd.cmd, flag.ExitOnError)
	cmd.destName = cmd.flagSet.String("destname", "", "The name of the recipient for the message")
	cmd.message = cmd.flagSet.String("message", "", "The message to send")
}

// GetFlagSet returns the flag set for this command.
func (cmd *SendMsgCmd) GetFlagSet() *flag.FlagSet {
	return cmd.flagSet
}

// Do performs the command action.
func (cmd *SendMsgCmd) Do(ctx context.Context, conn *grpc.ClientConn) {

	client := usercontroller.NewUserClient(conn)

	if *cmd.destName == "" || *cmd.message == "" {
		log.Fatal("Recipient name and message parameters are required")
	}

	message := usercontroller.DlcMessage{
		DestName: *cmd.destName,
		Payload:  []byte(*cmd.message),
	}

	_, err := client.SendDlcMessage(ctx, &message)

	if err != nil {
		log.Fatalf("Error sending message %v", err)
	}
}
