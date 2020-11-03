package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	grpcdelivery "github.com/muhammadisa/go-kit-boilerplate/services/user/delivery/grpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/muhammadisa/go-kit-boilerplate/protobuf/user_grpc"

	"github.com/muhammadisa/go-kit-boilerplate/services/user"
	"github.com/muhammadisa/go-kit-boilerplate/services/user/delivery"
	"github.com/muhammadisa/go-kit-boilerplate/services/user/implementation"
	"github.com/muhammadisa/go-kit-boilerplate/services/user/repository"
	"google.golang.org/grpc"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gocraft/dbr/dialect"
	"github.com/gocraft/dbr/v2"
	"github.com/joho/godotenv"
	"github.com/muhammadisa/godbconn"
)

func restMode(
	_ context.Context,
	logger log.Logger,
	userServiceHttp http.Handler,
) {
	/*
		http address flag and getting port from env
		Prepare HTTP Handler
	*/
	port := os.Getenv("HTTP_PORT")
	httpAddr := flag.String("http", port, "http listen address")
	flag.Parse()

	errs := make(chan error, 1)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()
	go func() {
		level.Info(logger).Log("transport", "HTTP", "addr", *httpAddr)
		server := &http.Server{
			Addr:    *httpAddr,
			Handler: userServiceHttp,
		}
		errs <- server.ListenAndServe()
	}()
	level.Error(logger).Log("exit", <-errs)
}

func grpcMode(
	_ context.Context,
	logger log.Logger,
	userServiceGrpc user_grpc.UserServiceServer,
) {
	port := ":50051"
	grpcListener, _ := net.Listen("tcp", port)
	grpcServer := grpc.NewServer()

	errs := make(chan error, 1)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()
	go func() {
		logger.Log("transport", "gRPC", "addr", port)
		user_grpc.RegisterUserServiceServer(grpcServer, userServiceGrpc)
		errs <- grpcServer.Serve(grpcListener)
	}()
	level.Error(logger).Log("exit", <-errs)
}

func grpcGatewayMode(
	_ context.Context,
	logger log.Logger,
	userServiceGrpc user_grpc.UserServiceServer,
) {
	mux := runtime.NewServeMux()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := user_grpc.RegisterUserServiceHandlerServer(ctx, mux, userServiceGrpc)
	if err != nil {
		level.Error(logger).Log("exit", err)
		os.Exit(-1)
	}
	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		level.Error(logger).Log("exit", err)
		os.Exit(-1)
	}
	http.Serve(listen, mux)
}

func createLogger() log.Logger {
	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.NewSyncLogger(logger)
	logger = level.NewFilter(logger, level.AllowDebug())
	logger = log.With(
		logger,
		"service", "user",
		"time: ", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)
	level.Info(logger).Log("msg", "service started")
	defer level.Info(logger).Log("msg", "service ended")
	return logger
}

func loadEnvironment(logger log.Logger) {
	// Load environment
	err := godotenv.Load()
	if err != nil {
		level.Error(logger).Log("exit", err)
		os.Exit(-1)
	}
}

func createDBRSession(logger log.Logger) *dbr.Session {
	// Load database credential env and use it
	db, err := godbconn.DBCred{
		DBDriver:   "mysql",
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
	}.Connect()
	if err != nil {
		level.Error(logger).Log("exit", err)
		os.Exit(-1)
	}
	conn := &dbr.Connection{
		DB:            db,
		EventReceiver: &dbr.NullEventReceiver{},
		Dialect:       dialect.MySQL,
	}
	conn.SetMaxOpenConns(10)
	session := conn.NewSession(nil)
	session.Begin()
	return session
}

func initService(session *dbr.Session, logger log.Logger) user.Service {
	repository := repository.NewUserRepository(session, logger)
	return implementation.NewService(repository, logger)
}

func initEndpoints(service user.Service) delivery.Endpoints {
	endpoints := delivery.MakeEndpoints(service)
	endpoints = delivery.Endpoints{
		Register: endpoints.Register,
		Login:    endpoints.Login,
	}
	return endpoints
}

func main() {
	// Initialize logger
	logger := createLogger()
	// Load environment
	loadEnvironment(logger)
	// Create dbr session
	session := createDBRSession(logger)
	// Init context and parse flags
	ctx := context.Background()
	// Prepare service
	service := initService(session, logger)
	// Prepare endpoints
	endpoints := initEndpoints(service)

	// Rest Http
	//userServiceHttp := httpdelivery.NewHTTPServe(ctx, endpoints, logger)
	//restMode(ctx, logger, userServiceHttp)

	// Grpc Http2
	userServiceGrpc := grpcdelivery.NewGRPCServer(endpoints, logger)
	grpcGatewayMode(ctx, logger, userServiceGrpc)
	//grpcMode(ctx, logger, userServiceGrpc)
}
