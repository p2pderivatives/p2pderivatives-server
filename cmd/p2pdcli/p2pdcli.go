package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"p2pderivatives-server/internal/common/token"

	"p2pderivatives-server/internal/cli"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	serverAddress       string
	authenticationToken string
)

// Command represent a command that can be invoked.
type Command interface {
	Command() string
	GetFlagSet() *flag.FlagSet
	Init()
	Do(context.Context, *grpc.ClientConn)
}

var commandMap map[string]Command

func init() {
	commandMap = make(map[string]Command)

	for _, cmd := range [...]Command{
		cli.NewRegisterUserCmd(),
		cli.NewUnregisterUserCmd(),
		cli.NewGetUserListCmd(),
		cli.NewLoginCmd(),
	} {
		cmd.Init()
		flagSet := cmd.GetFlagSet()
		flagSet.StringVar(&serverAddress, "server", "", "The address of the server.")
		flagSet.StringVar(&authenticationToken, "token", "", "The authentication token.")
		commandMap[cmd.Command()] = cmd
	}
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Need to specify a command. Available commands are:")

		for name := range commandMap {
			fmt.Println(name)
		}

		return
	}

	cmdName := os.Args[1]

	cmd, ok := commandMap[cmdName]

	if !ok {
		fmt.Println("Unknown command ", cmdName)
		return
	}

	if err := cmd.GetFlagSet().Parse(os.Args[2:]); err != nil {
		log.Fatalf("Error parsing flags %v", err)
	}

	conn, err := getConnection(&serverAddress)
	defer conn.Close()

	if err != nil {
		log.Fatalf("Error connecting to server %v", err)
	}

	ctx := context.Background()

	if authenticationToken != "" {
		ctx = metadata.AppendToOutgoingContext(
			ctx, token.MetaKeyAuthentication, authenticationToken)
	}

	cmd.Do(ctx, conn)
}

// getConnection returns a client for a GRPC server at the specified address.
func getConnection(serverAddress *string) (*grpc.ClientConn, error) {
	if *serverAddress == "" {
		return nil, errors.New("No server address provided")
	}

	var opts []grpc.DialOption

	opts = append(opts, grpc.WithInsecure())

	return grpc.Dial(*serverAddress, opts...)
}
