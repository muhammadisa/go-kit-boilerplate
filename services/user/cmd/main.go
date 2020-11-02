package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/muhammadisa/go-kit-boilerplate/services/user/delivery/protobuf/user_grpc"
	"github.com/oklog/oklog/pkg/group"

	"github.com/muhammadisa/go-kit-boilerplate/services/user"
	"github.com/muhammadisa/go-kit-boilerplate/services/user/delivery"
	grpcdelivery "github.com/muhammadisa/go-kit-boilerplate/services/user/delivery/grpc"
	httpdelivery "github.com/muhammadisa/go-kit-boilerplate/services/user/delivery/http"
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
	ctx context.Context,
	logger log.Logger,
	endpoints delivery.Endpoints,
) {
	/*
		http addres flag and geting port from env
		Perpare HTTP Handler
	*/
	port := os.Getenv("HTTP_PORT")
	httpAddr := flag.String("http", port, "http listen address")
	handler := httpdelivery.NewHTTPServe(ctx, endpoints, logger)
	flag.Parse()

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()
	go func() {
		level.Info(logger).Log("transport", "HTTP", "addr", *httpAddr)
		server := &http.Server{
			Addr:    *httpAddr,
			Handler: handler,
		}
		errs <- server.ListenAndServe()
	}()
	level.Error(logger).Log("exit", <-errs)
}

func grpcMode(
	_ context.Context,
	logger log.Logger,
	endpoints delivery.Endpoints,
) {
	port := ":50051"
	var (
		accountService  = grpcdelivery.NewGRPCServer(endpoints, logger)
		grpcListener, _ = net.Listen("tcp", port)
		grpcServer      = grpc.NewServer()
		g               group.Group
	)

	g.Add(func() error {
		logger.Log("transport", "gRPC", "addr", port)
		user_grpc.RegisterUserServiceServer(grpcServer, accountService)
		return grpcServer.Serve(grpcListener)
	}, func(error) {
		grpcListener.Close()
	})

	cancelInterrupt := make(chan struct{})
	g.Add(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		select {
		case sig := <-c:
			return fmt.Errorf("received signal %s", sig)
		case <-cancelInterrupt:
			return nil
		}
	}, func(error) {
		close(cancelInterrupt)
	})
	level.Error(logger).Log("exit", g.Run())
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
	// Perpare endpoints
	endpoints := initEndpoints(service)

	// starting mode
	// restMode(ctx, logger, endpoints)
	grpcMode(ctx, logger, endpoints)
}
