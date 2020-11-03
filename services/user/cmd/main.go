package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/muhammadisa/go-kit-boilerplate/middleware"
	grpcdelivery "github.com/muhammadisa/go-kit-boilerplate/services/user/delivery/grpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
)

func restMode(
	_ context.Context,
	logger log.Logger,
	userServiceHttp http.Handler,
) {
	port := os.Getenv("HTTP_PORT")
	httpAddr := flag.String("http", port, "http listen address")
	flag.Parse()

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()
	go func() {
		_ = level.Info(logger).Log("transport", "HTTP", "addr", *httpAddr)
		server := &http.Server{
			Addr:    *httpAddr,
			Handler: userServiceHttp,
		}
		errs <- server.ListenAndServe()
	}()
	_ = level.Error(logger).Log("exit", <-errs)
}

func grpcMode(
	_ context.Context,
	logger log.Logger,
	userServiceGrpc user_grpc.UserServiceServer,
) {
	port := os.Getenv("GRPC_PORT")
	grpcListener, _ := net.Listen("tcp", port)
	grpcServer := grpc.NewServer()

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()
	go func() {
		_ = logger.Log("transport", "gRPC", "addr", port)
		user_grpc.RegisterUserServiceServer(grpcServer, userServiceGrpc)
		errs <- grpcServer.Serve(grpcListener)
	}()
	_ = level.Error(logger).Log("exit", <-errs)
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
		_ = level.Error(logger).Log("exit", err)
		os.Exit(-1)
	}
	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		_ = level.Error(logger).Log("exit", err)
		os.Exit(-1)
	}

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()
	go func() {
		_ = logger.Log("transport", "gRPC and Restful", "addr", ":8080")
		errs <- http.Serve(listen, mux)
	}()
	_ = level.Error(logger).Log("exit", <-errs)
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
	_ = level.Info(logger).Log("msg", "service started")
	defer level.Info(logger).Log("msg", "service ended")
	return logger
}

func loadEnvironment(logger log.Logger) {
	// Load environment
	err := godotenv.Load()
	if err != nil {
		_ = level.Error(logger).Log("exit", err)
		os.Exit(-1)
	}
}

func createDBRSession(logger log.Logger) *dbr.Session {
	driverAndStrConn := fmt.Sprintf("mysql~%s", fmt.Sprintf(
		"%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	))
	connectionStr := strings.Split(driverAndStrConn, "~")
	db, err := sql.Open(connectionStr[0], connectionStr[1])
	if err != nil {
		_ = level.Error(logger).Log("exit", err)
		os.Exit(-1)
	}
	conn := &dbr.Connection{
		DB:            db,
		EventReceiver: &dbr.NullEventReceiver{},
		Dialect:       dialect.MySQL,
	}
	conn.SetMaxOpenConns(10)
	session := conn.NewSession(nil)
	_, err = session.Begin()
	if err != nil {
		_ = level.Error(logger).Log("exit", err)
		os.Exit(-1)
	}
	return session
}

func initService(session *dbr.Session) user.Service {
	userRepository := repository.NewUserRepository(session)
	return implementation.NewService(userRepository)
}

func initEndpoints(service user.Service, logger log.Logger) delivery.Endpoints {
	endpoints := delivery.MakeEndpoints(service)
	endpoints.Login = middleware.LoggingMiddleware(logger)(endpoints.Login)
	endpoints.Register = middleware.LoggingMiddleware(logger)(endpoints.Register)
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
	service := initService(session)
	// Prepare endpoints
	endpoints := initEndpoints(service, logger)

	// Rest Http
	//userServiceHttp := httpdelivery.NewHTTPServe(ctx, endpoints, logger)
	//restMode(ctx, logger, userServiceHttp)

	// Grpc Http2
	userServiceGrpc := grpcdelivery.NewGRPCServer(endpoints, logger)
	grpcGatewayMode(ctx, logger, userServiceGrpc)
	//grpcMode(ctx, logger, userServiceGrpc)

	defer ctx.Done()
}
