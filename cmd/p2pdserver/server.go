package main

import (
	"flag"
	stdlog "log"
	"net"
	"os"

	"p2pderivatives-server/internal/authentication"
	"p2pderivatives-server/internal/common/conf"
	"p2pderivatives-server/internal/common/grpc/methods"
	"p2pderivatives-server/internal/common/log"
	"p2pderivatives-server/internal/common/servererror"
	"p2pderivatives-server/internal/common/token"
	"p2pderivatives-server/internal/database/orm"
	"p2pderivatives-server/internal/user/usercommon"
	"p2pderivatives-server/internal/user/usercontroller"
	"p2pderivatives-server/internal/user/userrepository"
	"p2pderivatives-server/internal/user/userservice"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	configPath = flag.String("config", "", "Path to the configuration file to use.")
	appName    = flag.String("appname", "", "The name of the application. Will be use as a prefix for environment variables.")
	envname    = flag.String("e", "", "environment (ex., \"development\"). Should match with the name of the configuration file.")
	migrate    = flag.Bool("migrate", false, "If set performs a db migration before starting.")
)

// Config contains the configuration parameters for the server.
type Config struct {
	Address  string `configkey:"server.address" validate:"required"`
	TLS      bool   `configkey:"server.tls"`
	CertFile string `configkey:"server.certfile" validate:"required_with=TLS"`
	KeyFile  string `configkey:"server.keyfile" validate:"required_with=TLS"`
}

func newInitializedLog(config *conf.Configuration) *log.Log {
	logConfig := &log.Config{}
	config.InitializeComponentConfig(logConfig)
	logger := log.NewLog(logConfig)
	logger.Initialize()
	return logger
}

func newInitializedOrm(config *conf.Configuration, log *log.Log) *orm.ORM {
	ormConfig := &orm.Config{}
	config.InitializeComponentConfig(ormConfig)
	ormInstance := orm.NewORM(ormConfig, log)
	err := ormInstance.Initialize()

	if err != nil {
		panic("Could not initialize database.")
	}

	return ormInstance
}

func newUserService(config *conf.Configuration) (
	*userservice.Service, *usercommon.Config) {
	userConfig := &usercommon.Config{}
	repo := userrepository.NewRepository()
	config.InitializeComponentConfig(userConfig)
	return userservice.NewService(repo, userConfig, &servererror.ServiceError{}), userConfig
}

func main() {
	flag.Parse()

	if *configPath == "" {
		stdlog.Fatal("No configuration path specified")
	}

	if *appName == "" {
		stdlog.Fatal("No configuration name specified")
	}

	if *envname != "" {
		os.Setenv("P2PD_ENV", *envname)
	}

	config := conf.NewConfiguration(*appName, *envname, []string{*configPath})
	err := config.Initialize()

	if err != nil {
		stdlog.Fatalf("Could not read configuration %v.", err)
	}

	serverConfig := &Config{}

	config.InitializeComponentConfig(serverConfig)

	lis, err := net.Listen("tcp", serverConfig.Address)
	if err != nil {
		stdlog.Fatalf("failed to listen: %v", err)
	}

	opts := make([]grpc.ServerOption, 0)
	if serverConfig.TLS {
		certFile := serverConfig.CertFile
		keyFile := serverConfig.KeyFile
		if certFile == "" {
			stdlog.Fatal("Need to provide the path to the certificate file")
		}
		if keyFile == "" {
			stdlog.Fatal("Need to provide the path to the key file")
		}
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			stdlog.Fatalf("Failed to generate credentials %v", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	logInstance := newInitializedLog(config)
	ormInstance := newInitializedOrm(config, logInstance)
	tokenConfig := &token.Config{}
	config.InitializeComponentConfig(tokenConfig)
	token.Init(tokenConfig)

	if *migrate {
		err := doMigration(logInstance, ormInstance)

		if err != nil {
			stdlog.Fatalf("Failed to apply migration %v", err)
		}
	}

	opts = append(opts, grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		token.UnaryInterceptor(),
		orm.TransactionUnaryServerInterceptor(
			logInstance.NewEntry(),
			methods.TxOption,
			ormInstance),
		grpc_validator.UnaryServerInterceptor(),
	)), grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
		token.StreamInterceptor(),
		orm.TransactionStreamServerInterceptor(
			logInstance.NewEntry(),
			methods.TxOption,
			ormInstance),
		grpc_validator.StreamServerInterceptor(),
	)))

	userService, userConfig := newUserService(config)
	userController := usercontroller.NewController(userService, userConfig)
	authenticationController := authentication.NewController(userService, userConfig)

	grpcServer := grpc.NewServer(opts...)
	usercontroller.RegisterUserServer(grpcServer, userController)
	authentication.RegisterAuthenticationServer(
		grpcServer, authenticationController)
	stdlog.Printf("Ready to listen on %v", serverConfig.Address)
	methods.Init(grpcServer)
	grpcServer.Serve(lis)
}

func doMigration(l *log.Log, o *orm.ORM) error {
	migrator := orm.NewMigrator(
		o,
		&usercommon.User{},
	)

	return migrator.Initialize()
}
