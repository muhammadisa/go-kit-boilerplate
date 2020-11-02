package delivery

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/muhammadisa/go-kit-boilerplate/services/user"
)

// Endpoints struct
type Endpoints struct {
	Register endpoint.Endpoint
	Login    endpoint.Endpoint
}

// MakeEndpoints initialize all registered endpoint
func MakeEndpoints(s user.Service) Endpoints {
	return Endpoints{
		Register: makeRegisterEndpoint(s),
		Login:    makeLoginEndpoint(s),
	}
}

// makeRegisterEndpoint using go kit endpoint
func makeRegisterEndpoint(s user.Service) endpoint.Endpoint {
	return func(
		ctx context.Context,
		request interface{},
	) (interface{}, error) {
		req := request.(CreateRegisterRequest)
		status, err := s.Register(ctx, req.Email, req.Passwords)
		return CreateRegisterResponse{Status: status}, err
	}
}

// makeLoginEndpoint using go kit endpoint
func makeLoginEndpoint(s user.Service) endpoint.Endpoint {
	return func(
		ctx context.Context,
		request interface{},
	) (interface{}, error) {
		req := request.(CreateLoginRequest)
		status, err := s.Login(ctx, req.Email, req.Passwords)
		return CreateLoginResponse{Status: status}, err
	}
}
