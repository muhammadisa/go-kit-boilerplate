package grpc

import (
	"context"

	"github.com/go-kit/kit/log"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/muhammadisa/go-kit-boilerplate/services/user/delivery"
	"github.com/muhammadisa/go-kit-boilerplate/services/user/delivery/protobuf/user_grpc"
	oldcontext "golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	register kitgrpc.Handler
	login    kitgrpc.Handler
	logger   log.Logger
}

func NewGRPCServer(
	svcEndpoints delivery.Endpoints,
	logger log.Logger,
) user_grpc.UserServiceServer {
	var options []kitgrpc.ServerOption
	errorLogger := kitgrpc.ServerErrorLogger(logger)
	options = append(options, errorLogger)

	return &grpcServer{
		register: kitgrpc.NewServer(
			svcEndpoints.Register, decodeRegisterRequest, encodeRegisterResponse, options...,
		),
		login: kitgrpc.NewServer(
			svcEndpoints.Login, decodeLoginRequest, encodeLoginResponse, options...,
		),
		logger: logger,
	}
}

func (s *grpcServer) Register(
	ctx oldcontext.Context, req *user_grpc.RegisterRequest,
) (*user_grpc.RegisterResponse, error) {
	_, rep, err := s.register.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*user_grpc.RegisterResponse), nil
}

func (s *grpcServer) Login(
	ctx oldcontext.Context, req *user_grpc.LoginRequest,
) (*user_grpc.LoginResponse, error) {
	_, rep, err := s.login.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*user_grpc.LoginResponse), nil
}

// decodeRegisterRequest to json
func decodeRegisterRequest(
	_ context.Context,
	request interface{},
) (interface{}, error) {
	req := request.(user_grpc.RegisterRequest)
	return delivery.CreateRegisterRequest{
		Email:     req.Email,
		Passwords: req.Passwords,
	}, nil
}

// decodeLoginRequest to json
func decodeLoginRequest(
	_ context.Context,
	request interface{},
) (interface{}, error) {
	req := request.(user_grpc.LoginRequest)
	return delivery.CreateLoginRequest{
		Email:     req.Email,
		Passwords: req.Passwords,
	}, nil
}

// encodeRegisterResponse to json
func encodeRegisterResponse(
	_ context.Context,
	response interface{},
) (interface{}, error) {
	res := response.(delivery.CreateRegisterResponse)
	return &user_grpc.RegisterResponse{Status: res.Status}, nil
}

// encodeLoginResponse to json
func encodeLoginResponse(
	_ context.Context,
	response interface{},
) (interface{}, error) {
	res := response.(delivery.CreateLoginResponse)
	return &user_grpc.LoginResponse{Status: res.Status}, nil
}

func getError(err error) error {
	switch err {
	case nil:
		return nil
	default:
		return status.Error(codes.Unknown, err.Error())
	}
}
